package config

import (
	"os"
	"strconv"
	"time"
)

// Config 本地练习配置；可用环境变量覆盖，勿把生产密钥提交仓库。
type Config struct {
	HTTPAddr    string
	BaseURL     string
	MySQLDSN    string
	RedisAddr   string
	CacheTTL    time.Duration
	CodeLength  int
	MaxRetries  int
}

func Load() Config {
	return Config{
		HTTPAddr:   getenv("SHORTLINK_HTTP_ADDR", ":8080"),
		BaseURL:    getenv("SHORTLINK_BASE_URL", "http://localhost:8080"),
		MySQLDSN:   getenv("SHORTLINK_MYSQL_DSN", "root:root123@tcp(127.0.0.1:3307)/study?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr:  getenv("SHORTLINK_REDIS_ADDR", "127.0.0.1:6379"),
		CacheTTL:   durationEnv("SHORTLINK_CACHE_TTL", time.Hour),
		CodeLength: intEnv("SHORTLINK_CODE_LEN", 6),
		MaxRetries: intEnv("SHORTLINK_MAX_RETRIES", 8),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func intEnv(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func durationEnv(k string, def time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
