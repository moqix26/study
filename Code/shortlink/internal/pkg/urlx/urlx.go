package urlx

import (
	"errors"
	"net/url"
	"strings"
)

// Normalize 校验并规范化用户提交的 URL。
func Normalize(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("url required")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", errors.New("invalid url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("url must be http or https")
	}
	return u.String(), nil
}
