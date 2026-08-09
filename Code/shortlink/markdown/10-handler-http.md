# 10 · handler/http.go HTTP 层精读

> 主线请跟 [`../study.md`](../study.md)；本文是精读加餐。

> 对应源码：`internal/handler/http.go`（全文）  
> 目标：把路由 handler 如何翻译 HTTP 状态码、头、JSON 与 302 讲透。

---

## 0. 这个文件在分层里的位置

```text
客户端 HTTP
    ↓
gin 路由 + middleware.Logger
    ↓
handler（本文）— 绑参、状态码、头
    ↓
service.LinkService
```

`internal/app/app.go` 注册：

```go
r.GET("/health", h.Health)
r.POST("/api/links", h.CreateLink)
r.GET("/api/links/:code", h.GetLinkJSON)
r.GET("/:code", h.Redirect)   // 必须最后
```

---

## 1. 完整源码

```go
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
```

---

## 2. 类型与构造

### 2.1 `Handler`

```go
type Handler struct {
	svc *service.LinkService
}

func New(svc *service.LinkService) *Handler {
	return &Handler{svc: svc}
}
```

| 字段 | 含义 |
| --- | --- |
| `svc` | 业务入口；所有 handler 方法共享 |

方法 receiver `(h *Handler)`：不拷贝大结构，且可扩展字段。

### 2.2 `createReq`

```go
type createReq struct {
	URL string `json:"url"`
}
```

| 字段 | JSON 键 | 说明 |
| --- | --- | --- |
| `URL` | `url` | 仅接收长链，不让客户端指定 `code` |

---

## 3. `Health` — 探活

```go
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
```

| 项 | 值 |
| --- | --- |
| 路由 | `GET /health` |
| 状态码 | `200` |
| Body | `{"status":"ok"}` |

**用途**：Docker/K8s/手动检查进程是否活着；**不查** MySQL/Redis（启动时 `app.Run` 已 Ping 过）。

---

## 4. `CreateLink` — POST /api/links

### 4.1 读 JSON

```go
var req createReq
if err := c.ShouldBindJSON(&req); err != nil {
	c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
	return
}
```

| 情况 | 状态码 | `error` |
| --- | --- | --- |
| body 不是合法 JSON、类型不对 | **400** | `bad json` |
| 合法 JSON | 继续 | — |

`&req` 必须指针，Gin 才能写入字段。

### 4.2 调 service 与状态码分支

```go
res, err := h.svc.Create(req.URL)
if err != nil {
	msg := err.Error()
	if msg == "url required" || msg == "invalid url" || msg == "url must be http or https" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
	return
}
c.JSON(http.StatusCreated, res)
```

| 场景 | 状态码 | 响应 |
| --- | --- | --- |
| URL 校验失败（三种固定文案） | **400** | `{"error":"<msg>"}` |
| DB/随机码/`failed to allocate code` | **500** | `{"error":"<msg>"}` |
| 成功 | **201 Created** | `CreateResult`：`code`, `short_url`, `long_url` |

**为什么 201 不是 200？**  
REST 惯例：`201` 表示资源已创建，且 body 含新资源表示。

**上游**：`service.Create` → `urlx` + `shortcode` + `repo`。  
**不写缓存**：创建成功响应里**没有** `X-Cache`（尚未走读路径）。

### 4.3 验收示例

```powershell
Invoke-RestMethod -Uri http://localhost:8080/api/links -Method POST `
  -ContentType "application/json" -Body '{"url":"https://www.bilibili.com"}'
```

---

## 5. `GetLinkJSON` — GET /api/links/:code

```go
code := c.Param("code")
longURL, hit, err := h.svc.Resolve(c.Request.Context(), code)
```

| 步骤 | 说明 |
| --- | --- |
| `c.Param("code")` | 路径参数，如 `BaLrEf` |
| `c.Request.Context()` | 传给 Redis，**不要传 nil** |

### 5.1 错误与 404

```go
if err != nil {
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	return
}
if longURL == "" {
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	return
}
```

| 条件 | 状态码 |
| --- | --- |
| `err != nil`（MySQL 等） | **500** |
| `longURL == ""`（不存在或长度不对） | **404** |

**强调**：service 把「找不到」放在空串，handler 翻译成 404。

### 5.2 成功响应

```go
setCacheHeader(c, hit)
c.JSON(http.StatusOK, gin.H{
	"code":      code,
	"long_url":  longURL,
	"short_url": h.svc.ShortURL(code),
})
```

| 字段 | 含义 |
| --- | --- |
| `code` | 路径里的短码 |
| `long_url` | 解析出的长链 |
| `short_url` | `BaseURL/code` |
| 响应头 `X-Cache` | `HIT` 或 `MISS` |

**不跳转**：方便 `curl -i` 看缓存，不给浏览器 302。

---

## 6. `Redirect` — GET /:code

### 6.1 避开 health / api / favicon

```go
if code == "health" || code == "api" || code == "favicon.ico" {
	c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	return
}
```

| 原因 | 说明 |
| --- | --- |
| 路由顺序 | `/:code` 在最后，但浏览器可能请求 `/favicon.ico` 被当成 code |
| `health` / `api` | 防止误把保留路径当短码（若有人访问 `/health` 走 Redirect 路由的边界情况） |

**注意**：真正的 `GET /health` 由更具体的路由处理，不会进 `Redirect`。黑名单是**兜底**。

### 6.2 Resolve + 404/500

与 `GetLinkJSON` 相同：`Resolve` → 判 `err` / 空 `longURL`。

### 6.3 点击与 302

```go
setCacheHeader(c, hit)
h.svc.IncrClickAsync(code)
c.Redirect(http.StatusFound, longURL)
```

| 调用 | 含义 |
| --- | --- |
| `setCacheHeader` | 跳转响应也带 `X-Cache`，便于 curl 验收 |
| `IncrClickAsync` | 异步 `click_count++`，在 Redirect **之前**调度即可 |
| `c.Redirect(http.StatusFound, longURL)` | **302** + `Location: longURL` |

`http.StatusFound` 常量值 = **302**。浏览器自动跟 `Location`；`curl.exe -i` 默认不跟随，能看到 302 头。

| 对比 | GetLinkJSON | Redirect |
| --- | --- | --- |
| 状态码 | 200 + JSON | 302 |
| IncrClick | 否 | 是 |
| X-Cache | 有 | 有 |

---

## 7. `setCacheHeader`

```go
func setCacheHeader(c *gin.Context, hit bool) {
	if hit {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}
}
```

| `hit`（来自 service） | 响应头 |
| --- | --- |
| `true` | `X-Cache: HIT` |
| `false` | `X-Cache: MISS` |

**语义**：本次请求是否从 Redis 读到长链。首次 MySQL 命中后 SET 成功，仍标 **MISS**（本次未命中缓存）。

**与 middleware**：`Logger` 在 `c.Next()` 之后读 `c.Writer.Header().Get("X-Cache")` 打 OUT 日志。

HTTP 头名大小写不敏感；统一写 `X-Cache`。

---

## 8. 状态码总表（本文件）


| 码 | Handler | 条件 |
| --- | --- | --- |
| 200 | Health, GetLinkJSON | 成功 |
| 201 | CreateLink | 创建成功 |
| 302 | Redirect | 找到长链 |
| 400 | CreateLink | bad json / URL 校验 |
| 404 | GetLinkJSON, Redirect | not found / 黑名单 code |
| 500 | CreateLink, GetLinkJSON, Redirect | service 基础设施错误 |

---

## 9. 常见坑


| 坑 | 说明 |
| --- | --- |
| `/:code` 注册在前 | 会吃掉 `/api/links/...`；必须具体路由在前 |
| 用 `longURL==""` 当 500 | 应是 404 |
| 新增 urlx 错误未加入 400 分支 | 校验错误变 500 |
| `curl` 跟随重定向 | 看不到 302/`X-Cache`；用 `curl.exe -i` 且不 `-L` |
| 期望 Create 返回 X-Cache | 创建不走 Resolve，无此头 |
| Redirect 黑名单漏 `robots.txt` 等 | 可能 404 JSON 而非静态文件；V1 仅列三个 |

---

## 10. 验收

```powershell
# Health
Invoke-RestMethod http://localhost:8080/health

# 创建 201
curl.exe -i -X POST http://localhost:8080/api/links -H "Content-Type: application/json" -d "{\"url\":\"https://example.com\"}"

# 坏 JSON 400
curl.exe -i -X POST http://localhost:8080/api/links -H "Content-Type: application/json" -d "not-json"

# JSON 查询 + X-Cache
$code = "你的短码"
curl.exe -i "http://localhost:8080/api/links/$code"

# 跳转 302
curl.exe -i "http://localhost:8080/$code"

# 不存在 404
curl.exe -i "http://localhost:8080/xxxxxx"
curl.exe -i "http://localhost:8080/api/links/xxxxxx"
```

| 检查 | 期望 |
| --- | --- |
| POST 合法 | 201 + 三字段 |
| POST 非法 URL | 400 |
| GET JSON 首次 | 200 + `X-Cache: MISS` |
| GET JSON 二次 | `X-Cache: HIT` |
| GET /:code | 302 + `Location` |
| 假 code | 404 |

---

## 口述题

1. `CreateLink` 如何把 service 错误分成 400 和 500？列举各一种例子。
2. `GetLinkJSON` 和 `Redirect` 都调 `Resolve`，响应有何不同？为什么只有 Redirect 调 `IncrClickAsync`？
3. `Redirect` 里为什么要黑名单 `health`、`api`、`favicon.ico`？
4. 第一次从 MySQL 读到数据并 SET Redis 后，响应头 `X-Cache` 是 HIT 还是 MISS？为什么？
5. 路由注册顺序为什么要求 `GET /:code` 放在最后？
