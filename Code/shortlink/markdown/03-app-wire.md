# internal/app/app.go 依赖组装与启动 逐块精讲

> 主线请跟 [`../study.md`](../study.md)；本文是精读加餐。

> 对应源码：`internal/app/app.go`  
> 目标：搞清 `Run()` 里**依赖组装顺序**、每层职责、**路由为什么必须这个顺序**。

---

## 0. `app` 包在整体里的位置

```text
cmd/server/main.go
  └─ app.Run()                    ← 本文（「接线员」）
       ├─ config.Load()
       ├─ gorm + repo + migrate
       ├─ redis + cache
       ├─ service + handler
       └─ gin 引擎 + 路由 + r.Run
```

入口只有 3 行有效代码；**真正启动流程全在这里**。读懂 `Run()` = 读懂服务怎么拼起来。

---

## 1. 完整源码

```go
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
	cfg := config.Load()

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
```

---

## 2. `import` 一览

| 导入 | 角色 |
|------|------|
| `context` | `rdb.Ping(context.Background())` |
| `fmt` | 错误包装 `fmt.Errorf`、启动日志 `Println` |
| `shortlink/internal/config` | `config.Load()` |
| `shortlink/internal/cache` | Redis 封装 `LinkCache` |
| `shortlink/internal/repo` | MySQL 封装 `LinkRepo` |
| `shortlink/internal/service` | 业务逻辑 `LinkService` |
| `shortlink/internal/handler` | HTTP 处理 `Handler` |
| `shortlink/internal/middleware` | 请求日志 |
| `github.com/gin-gonic/gin` | HTTP 框架 |
| `github.com/redis/go-redis/v9` | Redis 客户端 |
| `gorm.io/driver/mysql` | GORM MySQL 驱动 |
| `gorm.io/gorm` | ORM |

依赖方向：**app → handler → service → repo/cache → model**，app 不反向被 internal 引用（除 `main` 调 `Run`）。

---

## 3. `Run()` 逐步拆解

### 3.1 加载配置

```go
cfg := config.Load()
```

| 符号 | 含义 |
|------|------|
| `cfg` | 本次进程的配置快照（见 [`02-config.md`](./02-config.md)） |

**必须先 Load：** 后面连 DB、Redis、监听端口都依赖 `cfg`。

---

### 3.2 打开 MySQL

```go
db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
if err != nil {
	return fmt.Errorf("mysql: %w", err)
}
```

| 符号 | 含义 |
|------|------|
| `gorm.Open` | 创建 `*gorm.DB`（连接池句柄，非单次连接） |
| `mysql.Open(cfg.MySQLDSN)` | 用配置里的 DSN 建 MySQL 方言 |
| `&gorm.Config{}` | GORM 选项；空配置用默认 |
| `fmt.Errorf("mysql: %w", err)` | 包装错误，上层 `log.Fatal` 能看到前缀 |

**失败即返回：** 没有 MySQL 无法持久化短链，不必继续连 Redis。

---

### 3.3 创建 Repo 并迁移表结构

```go
linkRepo := repo.NewLinkRepo(db)
if err := linkRepo.AutoMigrate(); err != nil {
	return fmt.Errorf("migrate: %w", err)
}
fmt.Println("mysql ok")
```

| 符号 | 含义 |
|------|------|
| `NewLinkRepo(db)` | 把 `*gorm.DB` 注入仓储层 |
| `AutoMigrate()` | 按 `model.Link` 建/对齐 `links` 表 |
| `mysql ok` | 给人看的启动检查点（验收用） |

**顺序原因：** 必须先有 `db` 才能 `NewLinkRepo`；迁移应在接受流量前完成。

详见 [`05-repo-link.md`](./05-repo-link.md)。

---

### 3.4 连接 Redis 并 Ping

```go
rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
if err := rdb.Ping(context.Background()).Err(); err != nil {
	return fmt.Errorf("redis: %w", err)
}
fmt.Println("redis ok")
```

| 符号 | 含义 |
|------|------|
| `redis.NewClient` | 创建 Redis 客户端（连接池） |
| `&redis.Options{Addr: cfg.RedisAddr}` | 只配地址，其余默认 |
| `Ping(ctx)` | 发 `PING`，验证连通 |
| `context.Background()` | 无超时根上下文；启动探活够用 |

**为何启动时 Ping 失败就退出？**

- 本项目读路径强依赖 Cache Aside；Redis 挂了虽可降级查 MySQL，但练习环境默认 Redis 必须在线，早失败早排查。
- 生产可改为：Ping 失败只告警、继续启动（纯降级模式）。

---

### 3.5 组装 Cache → Service → Handler

```go
linkCache := cache.NewLinkCache(rdb, cfg.CacheTTL)
svc := service.NewLinkService(cfg, linkRepo, linkCache)
h := handler.New(svc)
```

| 变量 | 依赖 | 职责 |
|------|------|------|
| `linkCache` | `rdb`, `CacheTTL` | `Get/Set/Del` 长链缓存 |
| `svc` | `cfg`, `linkRepo`, `linkCache` | 创建短链、解析、异步点击 |
| `h` | `svc` | HTTP 入参出参、状态码 |

**依赖组装顺序（不能乱）：**

```text
db → repo → (migrate)
rdb → cache
repo + cache + cfg → service → handler
```

- `service` 需要 `repo` 写库/查库、`cache` 读加速、`cfg` 拼 URL 和短码参数。
- `handler` 只调 `service`，**不**直接碰 DB/Redis——分层边界。

这是典型的**构造函数注入**（Constructor Injection）：依赖从外往里传，便于测试时换 mock。

---

### 3.6 创建 Gin 引擎与中间件

```go
r := gin.New()
r.Use(gin.Recovery())
r.Use(middleware.Logger())
```

| 调用 | 含义 |
|------|------|
| `gin.New()` | 空引擎，**不带** Gin 默认 Logger |
| `gin.Recovery()` | handler panic 时恢复，返回 500，避免进程崩溃 |
| `middleware.Logger()` | 自定义：请求前后打印 method、path、状态码、`X-Cache` |

**中间件执行顺序：**

```text
请求进入 → Recovery → Logger(前半) → 路由 handler → Logger(后半) → 响应
```

`Recovery` 放最外，能兜住包括 Logger 在内层的 panic。

---

### 3.7 注册路由（顺序极其重要）

```go
r.GET("/health", h.Health)
r.POST("/api/links", h.CreateLink)
r.GET("/api/links/:code", h.GetLinkJSON)
r.GET("/:code", h.Redirect)
```

| 路由 | 方法 | Handler | 作用 |
|------|------|---------|------|
| `/health` | GET | `h.Health` | 探活 |
| `/api/links` | POST | `h.CreateLink` | 创建短链 |
| `/api/links/:code` | GET | `h.GetLinkJSON` | JSON 查映射（带 `X-Cache`） |
| `/:code` | GET | `h.Redirect` | 302 跳转 |

#### 为什么 `GET /:code` 必须放最后？

Gin 按注册顺序匹配。`/:code` 是**单段通配**：任意一层路径都可能命中。

| 错误顺序 | 后果 |
|----------|------|
| `/:code` 写在 `/health` 前 | 访问 `/health` 时 `code=health`，走跳转逻辑 |
| `/:code` 写在 `/api/links/:code` 前 | `/api/links/xxx` 可能先被 `/:code` 吃掉（`code=api` 等） |

**正确策略：** 先注册**字面量多、更具体**的路由；把「兜底式」通配放最后。

`Redirect` handler 里还对 `health`、`api`、`favicon.ico` 做了额外 404 防护，但**不能替代**正确的路由顺序——顺序是第一道防线。

---

### 3.8 启动 HTTP 服务

```go
fmt.Println(cfg.HTTPAddr, "is on")
return r.Run(cfg.HTTPAddr)
```

| 符号 | 含义 |
|------|------|
| `r.Run(addr)` | 等价 `http.ListenAndServe(addr, r)`；**阻塞**直到进程退出 |
| 返回值 `error` | 端口占用等监听错误会返回；正常跑时一直阻塞，通常只有出错才返回到 `main` |

`main` 里：

```go
if err := app.Run(); err != nil {
	log.Fatal(err)
}
```

监听失败 → 打印并退出。

---

## 4. 启动流程总图

```text
config.Load()
    │
    ▼
gorm.Open(MySQL) ──失败──► return mysql: ...
    │
    ▼
NewLinkRepo + AutoMigrate ──失败──► return migrate: ...
    │
    ▼
redis.NewClient + Ping ──失败──► return redis: ...
    │
    ▼
NewLinkCache → NewLinkService → New Handler
    │
    ▼
gin.New + Recovery + Logger
    │
    ▼
注册路由（具体 → 通配）
    │
    ▼
r.Run(HTTPAddr) 阻塞
```

---

## 5. 与上下游怎么接

### 5.1 上游

| 调用方 | 调用 |
|--------|------|
| `cmd/server/main.go` | `app.Run()` |

### 5.2 下游（一次 HTTP 请求）

**创建短链：**

```text
POST /api/links
  → middleware.Logger
  → handler.CreateLink
  → service.Create
  → repo.Create (MySQL)
```

**跳转：**

```text
GET /BaLrEf
  → handler.Redirect
  → service.Resolve
       → cache.Get (Redis)
       → repo.FindByCode (MySQL miss 时)
       → cache.Set (回填)
  → service.IncrClickAsync → repo.IncrClick
  → 302 Location
```

`app` 包只负责**把线接好**；单次请求细节在 handler/service（其他 md 文档）。

---

## 6. 常见坑

| 坑 | 现象 | 原因 / 修法 |
|----|------|-------------|
| 路由顺序错 | `/health` 变跳转或 404 怪异 | `/:code` 移到最后 |
| 跳过 AutoMigrate | 表不存在，首请求 500 | 保持 migrate 在启动阶段 |
| MySQL 成功、Redis 失败仍继续 | 本项目直接 return | 若你改成弱依赖 Redis，要显式改 Ping 逻辑 |
| `gin.Default()` vs `gin.New()` | 多一套 Gin 自带日志 | 本项目要自定义 Logger，用 `New` + 手动 `Use` |
| 在 `Run()` 里用 `panic` | 绕过 error 返回约定 | 启动错误用 `return fmt.Errorf` |
| 重复 `r.Run` 或漏 `return` | 编译错误或双监听 | `return r.Run(...)` 一行即可 |
| handler 里直接 `gorm.Open` | 分层破坏、难测试 | 保持只在 `app` 接线 |

---

## 7. 本地怎么验证

### 7.1 正常启动

```powershell
cd F:\study\Code\shortlink
docker start study-mysql study-redis
go run ./cmd/server
```

期望输出（顺序大致如此）：

```text
mysql ok
redis ok
:8080 is on
```

### 7.2 验证路由与中间件

```powershell
# 探活
curl.exe -i http://localhost:8080/health

# 创建
curl.exe -X POST http://localhost:8080/api/links `
  -H "Content-Type: application/json" `
  -d '{"url":"https://www.example.com"}'

# JSON 查询（记下返回的 code）
curl.exe -i http://localhost:8080/api/links/<code>
# 第一次 X-Cache: MISS；再请求一次 HIT

# 跳转
curl.exe -i http://localhost:8080/<code>
# 302 Location: https://www.example.com
```

终端应看到 `[IN ]` / `[OUT]` 日志（`middleware.Logger`）。

### 7.3 验证路由顺序（反面教材）

**不要真改生产代码**，理解即可：若把 `r.GET("/:code", ...)` 移到第一行，`curl /health` 会进 `Redirect`，`code=health`。

### 7.4 验证启动失败快速退出

```powershell
docker stop study-redis
go run ./cmd/server
# 应无 redis ok，直接 redis: ... 错误
docker start study-redis
```

---

## 8. 和旧版单文件 main 的对照

| 旧单文件 | 现 `app.Run` |
|----------|--------------|
| `main()` 里全局 `var db, rdb` | 局部变量注入到各层 struct |
| 路由与 handler 同文件 | handler 在 `internal/handler` |
| `panic` 连不上 DB | `return error` 给 `main` |

行为对齐，结构更清晰。

---

## 9. 口述检查（2～3 题）

1. **`Run()` 里为什么先连 MySQL 再连 Redis？`NewLinkService` 为什么在 `NewLinkCache` 之后？**  
   （期望：持久化优先；service 依赖 repo 和 cache，构造顺序由内层依赖决定。）

2. **解释为什么 `GET /:code` 必须注册在 `/api/links/:code` 后面。若顺序反了会发生什么？**  
   （期望：Gin 先匹配先注册；通配会抢具体路径。）

3. **`gin.New()` + `Recovery` + `middleware.Logger` 和 `gin.Default()` 差在哪？`Recovery` 为什么建议放第一个 `Use`？**  
   （期望：Default 自带 logger/recovery；Recovery 在最外层兜 panic。）
