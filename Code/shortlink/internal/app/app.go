package app

import (
	"context"
	"fmt"

	"shortlink/internal/cache"
	"shortlink/internal/config"
	"shortlink/internal/handler"
	"shortlink/internal/middleware"
	"shortlink/internal/repo"
	"shortlink/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Run 组装依赖并启动 HTTP 服务。
func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("mysql: %w", err)
	}
	linkRepo := repo.NewLinkRepo(db)
	if err := linkRepo.AutoMigrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	fmt.Println("mysql ok")

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	fmt.Println("redis ok")

	linkCache := cache.NewLinkCache(rdb, cfg.CacheTTL)
	svc := service.NewLinkService(cfg, linkRepo, linkCache)
	h := handler.New(svc)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())

	r.GET("/health", h.Health)
	r.POST("/api/links", h.CreateLink)
	r.GET("/api/links/:code", h.GetLinkJSON)
	r.GET("/:code", h.Redirect)

	fmt.Println(cfg.HTTPAddr, "is on")
	return r.Run(cfg.HTTPAddr)
}
