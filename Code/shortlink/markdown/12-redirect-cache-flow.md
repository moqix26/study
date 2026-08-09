# 12 · 跳转与缓存全链路精读

> 主线请跟 [`../study.md`](../study.md)；本文是精读加餐。

> 跨文件：`handler` → `service` → `urlx` / `shortcode` / `repo` / `cache`  
> 目标：从一次 POST 到两次 GET，把 Cache Aside、302、异步点击、验收命令和坑串成一条故事。

---

## 0. 先记住三张图

### 0.1 组件关系

```text
cmd/server/main.go → app.Run()
  ├── config.Load()
  ├── repo.LinkRepo (MySQL :3307)
  ├── cache.LinkCache (Redis :6379)
  ├── service.LinkService
  ├── handler.Handler
  └── gin 路由 + middleware.Logger
```

### 0.2 写路径（创建）

```text
POST /api/links
  → handler.CreateLink
  → service.Create
      → urlx.Normalize
      → shortcode.Random (可能重试)
      → repo.Create → MySQL INSERT
  → 201 JSON（不写 Redis）
```

### 0.3 读路径（解析 + 跳转）

```text
GET /api/links/:code  或  GET /:code
  → handler.GetLinkJSON / Redirect
  → service.Resolve
      → cache.Get (Redis)
      → [miss 或 error 降级] repo.FindByCode (MySQL)
      → cache.Set (回填，失败只 log)
  → X-Cache: HIT|MISS
  → [仅 Redirect] IncrClickAsync → go repo.IncrClick
  → JSON 200 或 302 Location
```

**真源**：MySQL `links` 表。Redis 是加速层，可丢、可过期。

---

## 1. 启动与依赖（app.Run）

```go
// internal/app/app.go（节选）
cfg := config.Load()
db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), ...)
linkRepo := repo.NewLinkRepo(db)
linkRepo.AutoMigrate()

rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
rdb.Ping(context.Background())

linkCache := cache.NewLinkCache(rdb, cfg.CacheTTL)
svc := service.NewLinkService(cfg, linkRepo, linkCache)
h := handler.New(svc)

r.GET("/health", h.Health)
r.POST("/api/links", h.CreateLink)
r.GET("/api/links/:code", h.GetLinkJSON)
r.GET("/:code", h.Redirect)   // 最后
r.Use(middleware.Logger())
```

| 默认配置 | 值 | 环境变量 |
| --- | --- | --- |
| HTTP | `:8080` | `SHORTLINK_HTTP_ADDR` |
| MySQL | `127.0.0.1:3307` / `study` | `SHORTLINK_MYSQL_DSN` |
| Redis | `127.0.0.1:6379` | `SHORTLINK_REDIS_ADDR` |
| 短码长 | 6 | `SHORTLINK_CODE_LEN` |
| 缓存 TTL | 1h | `SHORTLINK_CACHE_TTL` |

---

## 2. 链路 A：创建短链（POST）

### 2.1 HTTP 入口

```http
POST /api/links HTTP/1.1
Content-Type: application/json

{"url":"https://www.bilibili.com"}
```

`handler.CreateLink`：

1. `ShouldBindJSON` → `createReq{URL}`
2. `h.svc.Create(req.URL)`
3. 成功 → `201` + `CreateResult`

### 2.2 service.Create 逐步

```go
longURL, err := urlx.Normalize(rawURL)
// TrimSpace → 非空 → Parse → scheme/host → 仅 http/https

for i := 0; i < MaxRetries; i++ {
    code, _ := shortcode.Random(CodeLength)  // crypto/rand, base62
    link := &model.Link{Code: code, LongURL: longURL}
    err = repo.Create(link)                  // INSERT
    if err == nil { return CreateResult{...}, nil }
    if !repo.IsDuplicate(err) { return nil, err }
}
return nil, errors.New("failed to allocate code")
```

| 层 | 文件 | 动作 |
| --- | --- | --- |
| 校验 | `urlx/urlx.go` | 拒绝空/非法/非 http(s) |
| 生成 | `shortcode/shortcode.go` | 6 位随机码 |
| 持久化 | `repo/link.go` | `db.Create`；`code` UNIQUE |
| 碰撞 | `repo.IsDuplicate` | duplicate/unique/1062 → 重试 |

### 2.3 创建后系统状态


| 存储 | 状态 |
| --- | --- |
| MySQL `links` | 新行：`code`, `long_url`, `click_count=0` |
| Redis `link:{code}` | **不存在**（故意不写） |
| 日志 | `[IN] POST /api/links` → `[OUT] ... 201 X-Cache=` |

### 2.4 响应示例

```json
{
  "code": "xY3kPq",
  "short_url": "http://localhost:8080/xY3kPq",
  "long_url": "https://www.bilibili.com"
}
```

---

## 3. 链路 B：首次读（Cache MISS）

以 `GET /api/links/xY3kPq` 为例（跳转路径逻辑相同，只是最后返回 302）。

### 3.1 handler

```go
code := c.Param("code")
longURL, hit, err := h.svc.Resolve(c.Request.Context(), code)
// longURL 非空 → setCacheHeader(c, false) → X-Cache: MISS
// GetLinkJSON → 200 JSON
// Redirect → IncrClickAsync + 302
```

### 3.2 service.Resolve

```go
if len(code) != CodeLength { return "", false, nil }

// 1) Redis
v, ok, err := cache.Get(ctx, code)
// redis.Nil → miss
// err != nil → log，继续 MySQL（降级）

// 2) MySQL
link, err := repo.FindByCode(code)
// ErrRecordNotFound → "", false, nil

// 3) 回填
cache.Set(ctx, code, link.LongURL)  // 失败只 log
return link.LongURL, false, nil     // cacheHit=false → MISS
```

### 3.3 cache 层 key

```go
// internal/cache/redis.go
func key(code string) string { return "link:" + code }
// GET link:xY3kPq
// SET link:xY3kPq <longURL> EX 3600
```

### 3.4 首次读时序

```text
Client          Handler           Service           Redis           MySQL
  | GET /api/...  |                  |                 |                |
  |-------------->| Resolve          |                 |                |
  |               |----------------->| GET link:code   |                |
  |               |                  |--------------->| miss           |
  |               |                  |                | FindByCode     |
  |               |                  |-------------------------------->| row
  |               |                  | SET link:code  |                |
  |               |                  |--------------->| OK             |
  |               | X-Cache: MISS    |                 |                |
  |<--------------| 200 JSON         |                 |                |
```

| 观察点 | 首次 |
| --- | --- |
| `X-Cache` | `MISS` |
| MySQL 查询 | 有 |
| Redis 读后 | 有 key（SET 成功时） |
| middleware OUT | `200 X-Cache= MISS` |

---

## 4. 链路 C：二次读（Cache HIT）

同一 `GET /api/links/xY3kPq` 或 `GET /xY3kPq`：

```go
v, ok, err := cache.Get(ctx, code)
if ok {
    return v, true, nil  // 直接返回，不访问 MySQL
}
```

```text
Client          Service           Redis           MySQL
  | Resolve       |                 |                |
  |-------------->| GET link:code   |                |
  |               |<---------------| hit (longURL)  |
  |               | (跳过 MySQL)    |                |
  | X-Cache: HIT  |                 |                |
```

| 观察点 | 二次 |
| --- | --- |
| `X-Cache` | `HIT` |
| MySQL | 正常情况**不打** |
| TTL | 仍按首次 SET 的 1h 倒计时 |

---

## 5. 链路 D：302 跳转 + 异步点击

`GET /xY3kPq`（浏览器或 curl）：

```go
// handler.Redirect
if code == "health" || code == "api" || code == "favicon.ico" {
    c.JSON(404, ...)
}
longURL, hit, err := h.svc.Resolve(ctx, code)
setCacheHeader(c, hit)
h.svc.IncrClickAsync(code)
c.Redirect(http.StatusFound, longURL)  // 302
```

### 5.1 响应头（curl -i）

```http
HTTP/1.1 302 Found
Location: https://www.bilibili.com
X-Cache: HIT
```

浏览器跟 `Location` 再发请求到 bilibili；**那是出站流量，不再经过本服务**。

### 5.2 IncrClickAsync

```go
go func() {
    repo.IncrClick(code)  // UPDATE click_count = click_count + 1
}()
```

| 特性 | 说明 |
| --- | --- |
| 时机 | 确认长链存在后、`Redirect` 前启动 goroutine |
| 阻塞 | 不阻塞 302 |
| 失败 | 只 log；用户无感 |
| JSON 接口 | **不计数** |

两次跳转 → `click_count` 通常 +2（略延迟）。

---

## 6. 「找不到」全链路

短码不存在或 `len(code)!=6`：

```go
// service.Resolve
return "", false, nil   // 不是 error

// handler
if longURL == "" {
    c.JSON(404, gin.H{"error": "not found"})
}
```

| 层 | 行为 |
| --- | --- |
| Redis | miss |
| MySQL | `ErrRecordNotFound` |
| HTTP | **404**，不是 500 |
| X-Cache | 通常 **MISS**（没读到有效缓存） |

---

## 7. Redis 故障降级

```go
if v, ok, err := s.cache.Get(ctx, code); err != nil {
    log.Println("redis get error:", err)
} else if ok {
    return v, true, nil
}
// 继续 MySQL
```

| 场景 | 行为 |
| --- | --- |
| Redis 宕机 | GET 报错 → 日志 → 仍查 MySQL → 可能 200/302 |
| SET 失败 | 仍返回 DB 结果；下次仍 MISS |
| 启动时 Redis 不通 | `app.Run` Ping 失败，**进程起不来** |

读路径可降级；写路径（创建）只依赖 MySQL。

---

## 8. middleware 贯穿每次请求

```go
fmt.Println("[IN ]", method, path)
c.Next()   // handler 执行完
fmt.Println("[OUT]", method, path, "->", Status, "X-Cache=", Header.Get("X-Cache"))
```

用 OUT 行快速确认 MISS→HIT 切换，无需每次解析 curl。

---

## 9. PowerShell 完整验收脚本

### 9.1 前置：启动依赖与服务

```powershell
cd F:\study\Code\shortlink
docker start study-mysql
docker start study-redis
go run .
# 另开终端继续，或后台运行
```

确认：

```powershell
Invoke-RestMethod http://localhost:8080/health
# status: ok
```

### 9.2 创建（201）

```powershell
$body = '{"url":"https://www.bilibili.com"}'
$r = Invoke-RestMethod -Uri http://localhost:8080/api/links -Method POST `
  -ContentType "application/json" -Body $body
$r
$code = $r.code
Write-Host "code=$code short=$($r.short_url)"
```

### 9.3 创建后 Redis 应无 key

```powershell
docker exec study-redis redis-cli GET "link:$code"
# (nil)  — 符合「创建不写缓存」
```

### 9.4 首次 JSON 读 — MISS

```powershell
curl.exe -i "http://localhost:8080/api/links/$code"
```

期望片段：

```text
HTTP/1.1 200 OK
X-Cache: MISS
{"code":"...","long_url":"https://www.bilibili.com","short_url":"..."}
```

### 9.5 二次 JSON 读 — HIT

```powershell
curl.exe -i "http://localhost:8080/api/links/$code"
```

期望：`X-Cache: HIT`

### 9.6 跳转两次 — 302 + 点击

```powershell
curl.exe -i "http://localhost:8080/$code"
curl.exe -i "http://localhost:8080/$code"
# 第一次可能 MISS，第二次 HIT（若上一步 JSON 已回填则两次都 HIT）
```

期望：

```text
HTTP/1.1 302 Found
Location: https://www.bilibili.com
X-Cache: HIT
```

（不要用 `curl -L`，否则看不到 302。）

### 9.7 Redis / MySQL 抽查

```powershell
docker exec study-redis redis-cli GET "link:$code"
# 应返回长 URL 字符串

Start-Sleep -Seconds 1
docker exec study-mysql mysql -uroot -proot123 -e `
  "SELECT code, long_url, click_count FROM study.links WHERE code='$code';"
# click_count >= 2（若执行了两次 Redirect）
```

### 9.8 负例

```powershell
# 假短码 404
curl.exe -i "http://localhost:8080/xxxxxx"
curl.exe -i "http://localhost:8080/api/links/xxxxxx"

# 非法 URL 400
curl.exe -i -X POST http://localhost:8080/api/links `
  -H "Content-Type: application/json" -d "{\"url\":\"not-a-url\"}"
```

### 9.9 验收总表


| # | 操作 | 期望 |
| --- | --- | --- |
| 1 | `GET /health` | 200 |
| 2 | `POST /api/links` 合法 URL | 201，含 code/short_url/long_url |
| 3 | 创建后立即 `redis-cli GET link:code` | nil |
| 4 | 首次 `GET /api/links/code` | 200，`X-Cache: MISS` |
| 5 | 二次同 URL | `X-Cache: HIT` |
| 6 | `curl -i GET /code` | 302，`Location` 正确 |
| 7 | Redis GET | 有长链 |
| 8 | MySQL click_count | 跳转后增加 |
| 9 | 不存在 code | 404 |
| 10 | 终端 Logger | IN/OUT，OUT 含状态与 X-Cache |

---

## 10. 常见坑（务必过一遍）

### 10.1 端口 3307 vs 3306

| 现象 | 原因 |
| --- | --- |
| `mysql: connection refused` | DSN 写错端口 |
| 本仓库默认 | Docker `study-mysql` 映射 **宿主机 3307 → 容器 3306** |

```text
SHORTLINK_MYSQL_DSN=root:root123@tcp(127.0.0.1:3307)/study?...
```

若你连 3306 而映射是 3307，启动即失败。

### 10.2 Redis 6379

| 现象 | 原因 |
| --- | --- |
| `redis: connection refused` | 容器未启或端口占用 |
| 读路径 Redis 挂 | 若进程已启动后 Redis 才挂，GET 失败会**降级 MySQL**，不必然 500 |

### 10.3 `context` 传 nil

```go
h.svc.Resolve(c.Request.Context(), code)  // 正确
```

| 错误写法 | 后果 |
| --- | --- |
| `Resolve(nil, code)` | redis 库可能异常或行为未定义 |
| 不用 Request.Context | 客户端断开时无法取消下游 |

### 10.4 路由 `/:code` 顺序

```go
r.GET("/api/links/:code", h.GetLinkJSON)
r.GET("/:code", h.Redirect)  // 必须最后
```

若 `/:code` 在前，`/api/links/xxx` 可能被错误 handler 处理。

`Redirect` 内黑名单：`health`、`api`、`favicon.ico` — 防止边角请求被当短码。

### 10.5 以为创建后缓存必有

验收步骤 3 专门查 `(nil)`。Cache Aside：**写只 MySQL，读才回填**。

### 10.6 X-Cache 语义

| 误解 | 正解 |
| --- | --- |
| SET 成功就是 HIT | 本次请求仍是 MISS；下次才 HIT |
| Create 返回 X-Cache | 创建不走 Resolve，无此头 |
| HIT = 数据一定最新 | TTL 内可能旧；更新/删除需失效策略（V1 无更新 API） |

### 10.7 curl 跟随重定向

`curl -L` 只看到最终站点的 200，看不到 302 和 `X-Cache`。验收用 `curl.exe -i` 且不加 `-L`。

### 10.8 PowerShell Invoke-RestMethod 与 400

校验失败返回 400 时，IRM 可能抛异常。用 `try/catch` 或改 `curl.exe` 看 body。

### 10.9 click_count 延迟

异步 goroutine，查 DB 太快可能仍是 0；`Start-Sleep 1` 再查。

### 10.9 短码长度与假码

`xxxxxx` 长度 6 但不存在 → 404。`health` 走 `GET /health` 路由，不是短码。

---

## 11. 与单文件 main.go 的对照


| 概念 | 单文件 | 分层后 |
| --- | --- | --- |
| 随机码 | `randomCode` | `shortcode.Random` |
| URL 校验 | `normalizeURL` | `urlx.Normalize` + https 白名单 |
| 读缓存 | `loadLongURL` | `service.Resolve` + `cache` 包 |
| 路由 | 同顺序 | `app.go` 注册 |
| 日志 | `Logger()` | `middleware.Logger` + X-Cache |

行为主线一致；分层便于测与扩展（如 `click_count`）。

---

## 12. 扩展思考（非 V1 实现）

| 话题 | V1 | 常见演进 |
| --- | --- | --- |
| 缓存失效 | 无更新 API | 删链时 `cache.Del` |
| 点击统计 | 异步 UPDATE | Kafka + 聚合 |
| 缓存穿透 | 不存在也打 DB | 布隆 / 空值缓存 |
| 热 key | 无 | 本地 LRU + Redis |

---

## 口述题

1. 画出从 `POST /api/links` 到 MySQL 一行数据落库的路径，标出哪几步**不**访问 Redis。
2. 第一次 `GET /api/links/:code` 时，Redis、MySQL、`X-Cache` 分别发生什么？
3. 第二次同样请求呢？什么情况下会再次 MISS？
4. `GET /:code` 与 `GET /api/links/:code` 在 service 层有何相同与不同？
5. Redis GET 报错时，用户请求一定会 500 吗？为什么？
6. 本机 MySQL 为什么用 3307？连错端口会有什么现象？
7. 为什么路由要把 `GET /:code` 放在最后？`favicon.ico` 不设黑名单会怎样？
8. 创建成功后立刻 `redis-cli GET link:{code}` 期望什么？说明设计理由。
