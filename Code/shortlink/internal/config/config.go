package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const configPath = "configs/config.env"

// Config is loaded only from configs/config.env.
type Config struct {
	HTTPAddr   string
	BaseURL    string
	MySQLDSN   string
	RedisAddr  string
	CacheTTL   time.Duration
	CodeLength int
	MaxRetries int
}

func Load() (Config, error) {
	values, err := loadConfigFile(configPath)
	if err != nil {
		return Config{}, err
	}

	httpAddr, err := requiredString(values, "HTTP_ADDR")
	if err != nil {
		return Config{}, err
	}
	baseURL, err := requiredString(values, "BASE_URL")
	if err != nil {
		return Config{}, err
	}
	mysqlDSN, err := requiredString(values, "MYSQL_DSN")
	if err != nil {
		return Config{}, err
	}
	redisAddr, err := requiredString(values, "REDIS_ADDR")
	if err != nil {
		return Config{}, err
	}
	cacheTTL, err := requiredDuration(values, "CACHE_TTL")
	if err != nil {
		return Config{}, err
	}
	codeLength, err := requiredPositiveInt(values, "CODE_LEN")
	if err != nil {
		return Config{}, err
	}
	maxRetries, err := requiredPositiveInt(values, "MAX_RETRIES")
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTPAddr:   httpAddr,
		BaseURL:    baseURL,
		MySQLDSN:   mysqlDSN,
		RedisAddr:  redisAddr,
		CacheTTL:   cacheTTL,
		CodeLength: codeLength,
		MaxRetries: maxRetries,
	}, nil
}

func loadConfigFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(f)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, fmt.Errorf("parse %s line %d: expected KEY=VALUE", path, lineNumber)
		}
		values[key] = strings.TrimSpace(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return values, nil
}

func requiredString(values map[string]string, key string) (string, error) {
	value := values[key]
	if value == "" {
		return "", fmt.Errorf("config %s is required", key)
	}
	return value, nil
}

func requiredPositiveInt(values map[string]string, key string) (int, error) {
	value, err := requiredString(values, key)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("config %s must be a positive integer", key)
	}
	return n, nil
}

func requiredDuration(values map[string]string, key string) (time.Duration, error) {
	value, err := requiredString(values, key)
	if err != nil {
		return 0, err
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("config %s must be a positive duration", key)
	}
	return duration, nil
}
