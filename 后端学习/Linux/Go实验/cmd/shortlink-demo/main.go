package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

type config struct {
	Addr              string
	Environment       string
	PublicBaseURL     string
	ShutdownTimeout   time.Duration
	TrustedProxyCIDRs []string
}

type link struct {
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	ShortURL  string    `json:"short_url"`
	CreatedAt time.Time `json:"created_at"`
}

type linkStore struct {
	mu     sync.RWMutex
	nextID uint64
	links  map[string]link
}

func newLinkStore() *linkStore {
	return &linkStore{
		nextID: 1,
		links:  make(map[string]link),
	}
}

func (s *linkStore) create(rawURL, publicBaseURL string) link {
	s.mu.Lock()
	defer s.mu.Unlock()

	code := encodeBase62(s.nextID)
	s.nextID++

	item := link{
		Code:      code,
		URL:       rawURL,
		ShortURL:  strings.TrimRight(publicBaseURL, "/") + "/r/" + code,
		CreatedAt: time.Now().UTC(),
	}
	s.links[code] = item
	return item
}

func (s *linkStore) get(code string) (link, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.links[code]
	return item, ok
}

func encodeBase62(value uint64) string {
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if value == 0 {
		return "0"
	}

	var buf [11]byte
	position := len(buf)
	for value > 0 {
		position--
		buf[position] = alphabet[value%62]
		value /= 62
	}
	return string(buf[position:])
}

func main() {
	healthCheck := flag.Bool("health-check", false, "check an already running service")
	healthURL := flag.String("health-url", "http://127.0.0.1:8080/healthz", "health endpoint used by -health-check")
	flag.Parse()

	if *healthCheck {
		if err := checkHealth(*healthURL); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := run(logger); err != nil {
		logger.Error("service_stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	store := newLinkStore()
	var ready atomic.Bool
	ready.Store(true)

	router := gin.New()
	if err := router.SetTrustedProxies(cfg.TrustedProxyCIDRs); err != nil {
		return fmt.Errorf("configure trusted proxies: %w", err)
	}
	router.Use(requestIDMiddleware())
	router.Use(accessLogMiddleware(logger))
	router.Use(recoveryMiddleware(logger))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/readyz", func(c *gin.Context) {
		if !ready.Load() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "shutting_down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":    version,
			"commit":     commit,
			"build_time": buildTime,
		})
	})

	router.POST("/api/links", func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		var request struct {
			URL string `json:"url" binding:"required"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_json_or_missing_url"})
			return
		}
		if err := validateHTTPURL(request.URL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_url", "detail": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, store.create(request.URL, cfg.PublicBaseURL))
	})

	router.GET("/api/links/:code", func(c *gin.Context) {
		item, ok := store.get(c.Param("code"))
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "link_not_found"})
			return
		}
		c.JSON(http.StatusOK, item)
	})

	router.GET("/r/:code", func(c *gin.Context) {
		item, ok := store.get(c.Param("code"))
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "link_not_found"})
			return
		}
		c.Redirect(http.StatusFound, item.URL)
	})

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "route_not_found"})
	})

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server_started", "addr", cfg.Addr, "version", version, "environment", cfg.Environment)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown_signal_received", "signal", ctx.Err())
	case serveErr := <-errCh:
		if !errors.Is(serveErr, http.ErrServerClosed) {
			return fmt.Errorf("listen: %w", serveErr)
		}
		return nil
	}

	ready.Store(false)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	logger.Info("server_stopped_cleanly")
	return nil
}

func loadConfig() (config, error) {
	timeoutText := envOrDefault("SHUTDOWN_TIMEOUT", "15s")
	timeout, err := time.ParseDuration(timeoutText)
	if err != nil || timeout <= 0 {
		return config{}, fmt.Errorf("invalid SHUTDOWN_TIMEOUT %q", timeoutText)
	}

	proxyText := strings.TrimSpace(envOrDefault("TRUSTED_PROXIES", "127.0.0.1"))
	var proxies []string
	if proxyText != "" {
		for _, value := range strings.Split(proxyText, ",") {
			value = strings.TrimSpace(value)
			if value != "" {
				proxies = append(proxies, value)
			}
		}
	}

	return config{
		Addr:              envOrDefault("APP_ADDR", "127.0.0.1:8080"),
		Environment:       envOrDefault("APP_ENV", "development"),
		PublicBaseURL:     envOrDefault("PUBLIC_BASE_URL", "http://localhost"),
		ShutdownTimeout:   timeout,
		TrustedProxyCIDRs: proxies,
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func validateHTTPURL(raw string) error {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil {
		return errors.New("URL syntax is invalid")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("only http and https are allowed")
	}
	if parsed.Host == "" {
		return errors.New("host is required")
	}
	return nil
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if requestID == "" || len(requestID) > 128 {
			requestID = newRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func accessLogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		logger.Info("http_request",
			"request_id", c.GetString("request_id"),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"bytes", c.Writer.Size(),
			"latency_ms", time.Since(started).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

func recoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("panic_recovered", "request_id", c.GetString("request_id"), "panic", recovered)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal_server_error"})
	})
}

func newRequestID() string {
	var value [12]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(value[:])
}

func checkHealth(endpoint string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(endpoint)
	if err != nil {
		return fmt.Errorf("health request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned %s", response.Status)
	}
	return nil
}
