package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"shortlink/internal/cache"
	"shortlink/internal/config"
	"shortlink/internal/model"
	"shortlink/internal/pkg/shortcode"
	"shortlink/internal/pkg/urlx"
	"shortlink/internal/repo"

	"gorm.io/gorm"
)

type LinkService struct {
	cfg   config.Config
	repo  *repo.LinkRepo
	cache *cache.LinkCache
}

func NewLinkService(cfg config.Config, r *repo.LinkRepo, c *cache.LinkCache) *LinkService {
	return &LinkService{cfg: cfg, repo: r, cache: c}
}

type CreateResult struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
}

func (s *LinkService) Create(rawURL string) (*CreateResult, error) {
	longURL, err := urlx.Normalize(rawURL)
	if err != nil {
		return nil, err
	}

	for i := 0; i < s.cfg.MaxRetries; i++ {
		code, err := shortcode.Random(s.cfg.CodeLength)
		if err != nil {
			return nil, err
		}
		link := &model.Link{Code: code, LongURL: longURL}
		err = s.repo.Create(link)
		if err == nil {
			return &CreateResult{
				Code:     link.Code,
				ShortURL: s.cfg.BaseURL + "/" + link.Code,
				LongURL:  link.LongURL,
			}, nil
		}
		if !repo.IsDuplicate(err) {
			return nil, err
		}
	}
	return nil, errors.New("failed to allocate code")
}

// Resolve 返回长链与是否缓存命中。
func (s *LinkService) Resolve(ctx context.Context, code string) (longURL string, cacheHit bool, err error) {
	if len(code) != s.cfg.CodeLength {
		return "", false, nil
	}

	if v, ok, err := s.cache.Get(ctx, code); err != nil {
		log.Println("redis get error:", err)
	} else if ok {
		return v, true, nil
	}

	link, err := s.repo.FindByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", false, nil
		}
		return "", false, err
	}

	if err := s.cache.Set(ctx, code, link.LongURL); err != nil {
		log.Println("redis set error:", err)
	}
	return link.LongURL, false, nil
}

// IncrClickAsync 异步增加点击（失败只打日志，不影响跳转）。
func (s *LinkService) IncrClickAsync(code string) {
	go func() {
		if err := s.repo.IncrClick(code); err != nil {
			log.Println("incr click error:", err)
		}
	}()
}

func (s *LinkService) ShortURL(code string) string {
	return fmt.Sprintf("%s/%s", s.cfg.BaseURL, code)
}
