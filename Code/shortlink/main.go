package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	codeLen      = 6
	codeAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	baseURL      = "http://localhost:8080"
	cacheTTL     = time.Hour
	maxRetries   = 8
)

type Link struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Code      string    `json:"code" gorm:"size:6;uniqueIndex;not null"`
	LongURL   string    `json:"long_url" gorm:"size:2048;not null"`
	CreatedAt time.Time `json:"created_at"`
}

type createLinkRequest struct {
	URL string `json:"url"`
}

var (
	db  *gorm.DB
	rdb *redis.Client
	ctx = context.Background()
)

func main() {
	var err error
	dsn := "root:root123@tcp(127.0.0.1:3307)/study?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("MySQL: " + err.Error())
	}
	if err := db.AutoMigrate(&Link{}); err != nil {
		panic(err)
	}
	fmt.Println("MySQL ok")

	rdb = redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic("Redis: " + err.Error())
	}
	fmt.Println("Redis ok")

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(Logger())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.POST("/api/links", createLink)
	r.GET("/api/links/:code", getLinkJSON)
	r.GET("/:code", redirectLink)

	fmt.Println(":8080 is on")
	r.Run(":8080")
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path
		fmt.Println("[IN ]", method, path)
		c.Next()
		fmt.Println("[OUT]", method, path, "->", c.Writer.Status())
	}
}

func linkKey(code string) string {
	return "link:" + code
}

func randomCode(n int) (string, error) {
	var b strings.Builder
	b.Grow(n)
	max := big.NewInt(int64(len(codeAlphabet)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b.WriteByte(codeAlphabet[idx.Int64()])
	}
	return b.String(), nil
}

func normalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("url required")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", errors.New("invalid url")
	}
	return u.String(), nil
}

func createLink(c *gin.Context) {
	var req createLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}
	longURL, err := normalizeURL(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var link Link
	for i := 0; i < maxRetries; i++ {
		code, err := randomCode(codeLen)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		link = Link{Code: code, LongURL: longURL}
		err = db.Create(&link).Error
		if err == nil {
			c.JSON(http.StatusCreated, gin.H{
				"code":      link.Code,
				"short_url": baseURL + "/" + link.Code,
				"long_url":  link.LongURL,
			})
			return
		}
		if !errors.Is(err, gorm.ErrDuplicatedKey) && !isDuplicate(err) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to allocated code"})
}

func isDuplicate(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}

func loadLongURL(code string) (string, bool, error) {
	if len(code) != codeLen {
		return "", false, nil
	}
	key := linkKey(code)

	val, err := rdb.Get(ctx, key).Result()
	if err == nil && val != "" {
		return val, true, nil
	}
	if err != nil && !errors.Is(err, redis.Nil) {
		fmt.Println("redis get error:", err)
	}

	var link Link
	if err := db.Where("code = ?", code).First(&link).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", false, nil
		}
		return "", false, err
	}

	if err := rdb.Set(ctx, key, link.LongURL, cacheTTL).Err(); err != nil {
		fmt.Println("redis set error:", err)
	}
	return link.LongURL, false, nil
}

func getLinkJSON(c *gin.Context) {
	code := c.Param("code")
	longURL, hit, err := loadLongURL(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if longURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if hit {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}
	c.JSON(http.StatusOK, gin.H{
		"code":      code,
		"long_url":  longURL,
		"short_url": baseURL + "/" + code,
	})
}

func redirectLink(c *gin.Context) {
	code := c.Param("code")
	if code == "health" || code == "api" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	longURL, hit, err := loadLongURL(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if longURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if hit {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}
	c.Redirect(http.StatusFound, longURL)
}
