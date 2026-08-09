# internal/config/config.go 逐块精讲

> 主线请跟 [`../study.md`](../study.md)；本文是精读加餐。

> 对应源码：`internal/config/config.go`、`configs/config.example.env`  
> 目标：搞清每个配置项干什么、环境变量怎么覆盖、DSN 每个字段什么意思。

---

## 0. 配置层在整体里的位置

```text
app.Run()
  └─ cfg := config.Load()     ← 本文
       ├─ HTTPAddr / BaseURL   → gin 监听、拼 short_url
       ├─ MySQLDSN             → gorm.Open
       ├─ RedisAddr            → redis.NewClient
       ├─ CacheTTL             → LinkCache.Set 的过期时间
       ├─ CodeLength           → 短码长度、Resolve 校验
       └─ MaxRetries           → 创建时撞唯一索引重试次数
```

**设计原则：** 本地练习用**合理默认值**直接能跑；上环境用**环境变量**覆盖，不把生产密码写进仓库。

---

## 1. 完整源码

```go
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
```

---

## 2. `package config` 与 `import`

```go
package config

import (
	"os"
	"strconv"
	"time"
)
```

| 包 | 用途 |
|----|------|
| `os` | `os.Getenv` 读环境变量 |
| `strconv` | `Atoi` 把字符串转成 `int` |
| `time` | `Duration` 类型、`ParseDuration` 解析 `1h`、`30m` 等 |

本包**不**依赖 Gin、GORM、Redis——纯配置，方便单测和复用。

---

## 3. `Config` 结构体

```go
type Config struct {
	HTTPAddr    string
	BaseURL     string
	MySQLDSN    string
	RedisAddr   string
	CacheTTL    time.Duration
	CodeLength  int
	MaxRetries  int
}
```

| 字段 | 类型 | 默认值（环境变量） | 谁消费 | 干什么 |
|------|------|-------------------|--------|--------|
| `HTTPAddr` | `string` | `:8080`（`SHORTLINK_HTTP_ADDR`） | `app.Run` → `r.Run(cfg.HTTPAddr)` | HTTP 监听地址；`:8080` = 所有网卡 8080 |
| `BaseURL` | `string` | `http://localhost:8080`（`SHORTLINK_BASE_URL`） | `LinkService.Create` / `ShortURL` | 拼返回 JSON 里的 `short_url` |
| `MySQLDSN` | `string` | 见下文（`SHORTLINK_MYSQL_DSN`） | `gorm.Open(mysql.Open(...))` | MySQL 连接串 |
| `RedisAddr` | `string` | `127.0.0.1:6379`（`SHORTLINK_REDIS_ADDR`） | `redis.Options{Addr: ...}` | Redis 地址:端口 |
| `CacheTTL` | `time.Duration` | `1h`（`SHORTLINK_CACHE_TTL`） | `cache.NewLinkCache(..., ttl)` | Redis 缓存过期时间 |
| `CodeLength` | `int` | `6`（`SHORTLINK_CODE_LEN`） | `shortcode.Random`、 `Resolve` 长度校验 | 短码字符个数 |
| `MaxRetries` | `int` | `8`（`SHORTLINK_MAX_RETRIES`） | `LinkService.Create` 循环上限 | 短码碰撞时最多重试几次 |

**为什么用结构体而不是全局变量？**

- `Load()` 一次得到完整快照，可传给 `NewLinkService(cfg, ...)`，依赖清晰。
- 测试时可构造假 `Config`，不必污染进程环境变量。

**为什么注释写「勿把生产密钥提交仓库」？**

- 默认值里的 `root:root123` 仅适合本机 Docker 练习。
- 生产密码应走环境变量 / 密钥管理系统，**不要**写进 `config.go` 或提交 `.env`。

---

## 4. `Load()` 函数

```go
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
```

| 符号 | 含义 |
|------|------|
| `Load()` | 无参数；读环境变量 + 默认值，返回**值类型** `Config`（拷贝一份） |
| `getenv` / `intEnv` / `durationEnv` | 三个小助手，统一「有环境变量用环境变量，否则默认」 |

**为什么返回 `Config` 而不是 `*Config`？**

- 结构体不大，值拷贝成本低。
- 避免调用方误改共享指针；需要修改时显式传参。

**环境变量命名约定：** 统一前缀 `SHORTLINK_`，避免和系统里别的 `HTTP_ADDR` 冲突。

---

## 5. MySQL DSN 逐字段拆解

默认 DSN：

```text
root:root123@tcp(127.0.0.1:3307)/study?charset=utf8mb4&parseTime=True&loc=Local
```

Go MySQL 驱动（经 GORM）常见格式：

```text
用户名:密码@tcp(主机:端口)/数据库名?参数1=值1&参数2=值2
```

| 片段 | 含义 | 为什么这样设 |
|------|------|------------|
| `root` | MySQL 用户名 | Docker 练习容器常用 root |
| `root123` | 密码 | **仅本地**；勿提交真实生产密码 |
| `@` | 分隔认证信息与网络地址 | 驱动解析约定 |
| `tcp(...)` | 走 TCP 协议连接 | 本机连 Docker 映射端口 |
| `127.0.0.1` | 主机 | 本机回环 |
| `3307` | 端口 | 见下节「为何 3307」 |
| `/study` | 数据库名 | **前面的 `/` 不能少** |
| `charset=utf8mb4` | 字符集 | 支持 emoji、完整 Unicode |
| `parseTime=True` | 解析时间列 | 让 GORM 把 `DATETIME` 映射成 `time.Time` |
| `loc=Local` | 时区 | 与本地时间一致 |

**常见 DSN 错误：**

- 漏写 `/study` → 连不上或连错库
- 密码含特殊字符未 URL 编码 → 认证失败
- `parseTime=False` → `CreatedAt` 等时间字段行为异常

---

## 6. 为何 MySQL 用 3307、Redis 用 6379？

| 服务 | 默认端口 | 本项目配置 | 原因 |
|------|----------|------------|------|
| MySQL | 容器内 3306 | 宿主机 **3307** | 你本机若已装 MySQL 占 3306，Docker 把容器 3306 **映射**到宿主机 3307，避免端口冲突 |
| Redis | 6379 | **6379** | Redis 官方默认端口；练习环境通常直接映射 6379:6379，本机无冲突就不改 |

```text
你的程序 ──tcp──► 127.0.0.1:3307 ──Docker 映射──► 容器内 MySQL:3306
你的程序 ──tcp──► 127.0.0.1:6379 ──Docker 映射──► 容器内 Redis:6379
```

若你改了 Docker `-p` 映射，必须同步改 `SHORTLINK_MYSQL_DSN` / `SHORTLINK_REDIS_ADDR`。

---

## 7. 三个 env 助手函数

### 7.1 `getenv`

```go
func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
```

| 参数/变量 | 含义 |
|-----------|------|
| `k` | 环境变量名，如 `SHORTLINK_HTTP_ADDR` |
| `def` | 未设置或为空字符串时的默认值 |
| `os.Getenv(k)` | 读环境变量；不存在返回 `""` |
| `v != ""` | 空字符串视为「未配置」，仍用默认值 |

**为何空字符串回退默认？** 避免误设 `SHORTLINK_HTTP_ADDR=` 导致监听地址非法；宁可显式不配。

### 7.2 `intEnv`

```go
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
```

| 符号 | 含义 |
|------|------|
| `strconv.Atoi` | ASCII to int；`"6"` → `6` |
| `err != nil` | 如设成 `SHORTLINK_CODE_LEN=abc`，**静默回退默认**，不 panic |

**权衡：** 学习项目选择「坏值用默认」；生产可加日志或启动失败。

### 7.3 `durationEnv`

```go
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
```

| 合法示例 | 含义 |
|----------|------|
| `1h` | 1 小时（代码里 `time.Hour` 的默认） |
| `30m` | 30 分钟 |
| `3600s` | 3600 秒 |

`time.ParseDuration` 支持 `ns`、`us`、`µs`、`ms`、`s`、`m`、`h` 组合。

---

## 8. `configs/config.example.env`

仓库里的示例（全部注释掉，防止误加载）：

```env
# SHORTLINK_HTTP_ADDR=:8080
# SHORTLINK_BASE_URL=http://localhost:8080
# SHORTLINK_MYSQL_DSN=root:root123@tcp(127.0.0.1:3307)/study?charset=utf8mb4&parseTime=True&loc=Local
# SHORTLINK_REDIS_ADDR=127.0.0.1:6379
# SHORTLINK_CACHE_TTL=1h
# SHORTLINK_CODE_LEN=6
# SHORTLINK_MAX_RETRIES=8
```

| 要点 | 说明 |
|------|------|
| 文件名 `*.example.env` | 模板进 Git；真实 `.env` 应进 `.gitignore` |
| 行首 `#` | 只是文本文件注释；**Go 不会自动读这个文件** |
| 如何使用 | 复制为 `.env` 后，用工具 `source` / `dotenv` 注入环境变量，或在 PowerShell 里 `$env:SHORTLINK_HTTP_ADDR=":9090"` |

**本项目的 `config.Load()` 只认 `os.Getenv`，不认 `.env` 文件本身。** 你需要自己把变量 export 到进程环境。

---

## 9. 与上下游怎么接

### 9.1 上游

| 谁 | 怎么影响配置 |
|----|--------------|
| 操作系统 / Shell | `export` / `$env:...` 设置环境变量 |
| Docker Compose | `environment:` 段注入 |
| K8s | `ConfigMap` / `Secret` 挂成 env |
| 你 | 不设变量 → 用 `Load()` 内置默认值 |

### 9.2 下游

```text
config.Load()
  ├─► app.Run
  │     ├─ cfg.MySQLDSN  → gorm.Open
  │     ├─ cfg.RedisAddr → redis.NewClient
  │     ├─ cfg.CacheTTL  → cache.NewLinkCache
  │     └─ cfg.HTTPAddr  → r.Run
  └─► service.NewLinkService(cfg, ...)
        ├─ cfg.BaseURL / CodeLength / MaxRetries
        └─ Resolve 用 cfg.CodeLength 校验路径参数长度
```

`BaseURL` 与 `HTTPAddr` **可以不一致**（例如内网监听 `:8080`，对外 `BaseURL` 是 `https://s.example.com`）。本地练习两者通常都是 `localhost:8080`。

---

## 10. 常见坑

| 坑 | 现象 | 修法 |
|----|------|------|
| 改了 `config.example.env` 但没 export | 服务仍用旧默认 | 环境变量要注入**进程**；改 example 文件 alone 无效 |
| DSN 漏 `/study` | `Error 1049: Unknown database` | 检查 `/数据库名` |
| 3306 vs 3307 搞混 | `connection refused` | 看 `docker ps` 端口映射 |
| `BaseURL` 末尾多 `/` | `short_url` 变成 `http://x//abc` | `BaseURL` 不要尾斜杠；拼接在 service 里用 `BaseURL + "/" + code` |
| `SHORTLINK_CODE_LEN` 与表字段不一致 | 6 位码能存，但以后改 8 位可能截断 | 模型 `gorm:"size:16"` 留余量；改长度要同步 service、校验逻辑 |
| 把真实密码 commit | 泄露风险 | 只用 example；密码走 env / Secret |
| `ParseDuration` 写成 `1 hour` | 解析失败，静默用默认 1h | 必须 `1h` 这种格式 |

---

## 11. 本地怎么验证

### 11.1 默认配置启动

```powershell
cd F:\study\Code\shortlink
docker start study-mysql study-redis
go run .
# 应看到 :8080 is on
```

### 11.2 覆盖 HTTP 端口

```powershell
$env:SHORTLINK_HTTP_ADDR=":9090"
$env:SHORTLINK_BASE_URL="http://localhost:9090"
go run .
# 终端应打印 :9090 is on

curl.exe http://localhost:9090/health
```

### 11.3 验证 DSN（故意错端口）

```powershell
$env:SHORTLINK_MYSQL_DSN="root:root123@tcp(127.0.0.1:3306)/study?charset=utf8mb4&parseTime=True&loc=Local"
go run .
# 若本机 3306 没 MySQL：应 mysql: ... 错误退出
```

改回 3307 或清掉变量：

```powershell
Remove-Item Env:SHORTLINK_MYSQL_DSN -ErrorAction SilentlyContinue
```

### 11.4 验证 CacheTTL（需配合 Redis CLI）

```powershell
$env:SHORTLINK_CACHE_TTL="30s"
go run .
# 创建短链并访问一次后：
docker exec -it study-redis redis-cli TTL link:<你的短码>
# 应接近 30（秒），不是 3600
```

### 11.5 读配置的小技巧（临时 debug）

在 `app.Run()` 里临时加一行（学完删掉）：

```go
fmt.Printf("%+v\n", cfg)
```

确认 env 是否生效。**不要**在生产日志里打印完整 DSN（含密码）。

---

## 12. 口述检查（2～3 题）

1. **`Load()` 没有读 `config.example.env`，那 example 文件有什么用？本地要怎么让 `SHORTLINK_HTTP_ADDR` 生效？**  
   （期望：模板/documentation；PowerShell `$env:...` 或 dotenv 注入进程环境变量。）

2. **把 DSN `root:root123@tcp(127.0.0.1:3307)/study?...` 拆开说：用户名、密码、主机、端口、库名、两个 query 参数各干什么？**  
   （期望：能说出 3307 是宿主机映射、parseTime/loc/charset 的作用。）

3. **`BaseURL` 和 `HTTPAddr` 有什么区别？如果服务跑在反向代理后面，你会改哪个？**  
   （期望：前者拼对外短链 URL，后者是进程 bind 地址；对外域名改 BaseURL，监听可能仍是内网端口。）
