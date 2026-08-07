package handler

import (
	"net/http"

	"shortlink/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.LinkService
}

func New(svc *service.LinkService) *Handler {
	return &Handler{svc: svc}
}

type createReq struct {
	URL string `json:"url"`
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) CreateLink(c *gin.Context) {
	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}
	res, err := h.svc.Create(req.URL)
	if err != nil {
		// 校验类错误用 400
		msg := err.Error()
		if msg == "url required" || msg == "invalid url" || msg == "url must be http or https" {
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *Handler) GetLinkJSON(c *gin.Context) {
	code := c.Param("code")
	longURL, hit, err := h.svc.Resolve(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if longURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	setCacheHeader(c, hit)
	c.JSON(http.StatusOK, gin.H{
		"code":      code,
		"long_url":  longURL,
		"short_url": h.svc.ShortURL(code),
	})
}

func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("code")
	if code == "health" || code == "api" || code == "favicon.ico" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	longURL, hit, err := h.svc.Resolve(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if longURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	setCacheHeader(c, hit)
	h.svc.IncrClickAsync(code)
	c.Redirect(http.StatusFound, longURL)
}

func setCacheHeader(c *gin.Context, hit bool) {
	if hit {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}
}
