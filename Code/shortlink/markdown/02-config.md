# 配置文件 `configs/config.env`

> 主线请跟 [`../study.md`](../study.md)；本文解释当前项目唯一的配置来源。
>
> 对应源码：`configs/config.env`、`internal/config/config.go`。

---

## 1. 配置从哪里来

项目启动时只读取一个固定路径：

```text
configs/config.env
```

当前项目不读取 PowerShell、Docker 或系统环境变量，也没有代码默认值。配置文件缺失、缺字段或格式错误时，程序会在连接 MySQL 和 Redis 前直接退出。

```text
cmd/server/main.go
  -> app.Run()
  -> config.Load()
  -> 读取 configs/config.env
  -> 生成有类型的 Config
  -> 连接 MySQL、Redis，启动 Gin
```

## 2. 配置文件内容

[`../configs/config.env`](../configs/config.env) 使用简单的 `KEY=VALUE` 格式：

```env
SHORTLINK_HTTP_ADDR=:8080
SHORTLINK_BASE_URL=http://localhost:8080
SHORTLINK_MYSQL_DSN=root:root123@tcp(127.0.0.1:3307)/study?charset=utf8mb4&parseTime=True&loc=Local
SHORTLINK_REDIS_ADDR=127.0.0.1:6379
SHORTLINK_CACHE_TTL=1h
SHORTLINK_CODE_LEN=6
SHORTLINK_MAX_RETRIES=8
```

规则：

| 写法 | 结果 |
|------|------|
| `KEY=VALUE` | 读取为一项配置 |
| 空行 | 忽略 |
| `# 注释` | 忽略 |
| 少了 `=` | 启动失败，报告文件行号 |
| 必填项没有值 | 启动失败，报告变量名 |
| 数字或时长格式错误 | 启动失败，报告变量名 |

## 3. 每一项的含义

| 配置项 | 类型 | 作用 |
|--------|------|------|
| `SHORTLINK_HTTP_ADDR` | 字符串 | Gin 监听地址，例如 `:8080` |
| `SHORTLINK_BASE_URL` | 字符串 | 创建短链时拼接 `short_url` 的前缀 |
| `SHORTLINK_MYSQL_DSN` | 字符串 | MySQL 连接串 |
| `SHORTLINK_REDIS_ADDR` | 字符串 | Redis 地址，例如 `127.0.0.1:6379` |
| `SHORTLINK_CACHE_TTL` | `time.Duration` | Redis 缓存有效期，例如 `1h`、`30m` |
| `SHORTLINK_CODE_LEN` | 正整数 | 随机短码长度 |
| `SHORTLINK_MAX_RETRIES` | 正整数 | 短码碰撞时的最大重试次数 |

`SHORTLINK_HTTP_ADDR` 和 `SHORTLINK_BASE_URL` 要一起改。比如改到 9090：

```env
SHORTLINK_HTTP_ADDR=:9090
SHORTLINK_BASE_URL=http://localhost:9090
```

改完后重启：

```powershell
go run ./cmd/server
```

## 4. 为什么还需要 `Config` 结构体

配置文件读出的都是字符串，但下游代码需要明确类型：

```text
"1h" -> time.Duration
"6"  -> int
```

因此 `config.Load()` 会把文件内容校验、转换后组装为：

```go
type Config struct {
	HTTPAddr   string
	BaseURL    string
	MySQLDSN   string
	RedisAddr  string
	CacheTTL   time.Duration
	CodeLength int
	MaxRetries int
}
```

`app.Run()`、缓存和业务服务只使用这个结构体，不需要关心文件解析细节。

## 5. 启动失败示例

如果把：

```env
SHORTLINK_CODE_LEN=abc
```

写进配置文件，启动会失败：

```text
config: config SHORTLINK_CODE_LEN must be a positive integer
```

如果删掉：

```env
SHORTLINK_REDIS_ADDR=127.0.0.1:6379
```

启动会失败：

```text
config: config SHORTLINK_REDIS_ADDR is required
```

这比静默回退到某个默认值更容易发现配置错误。

## 6. 本地验收

```powershell
cd F:\study\Code\shortlink
docker start study-mysql study-redis
go run ./cmd/server
```

期望输出：

```text
mysql ok
redis ok
:8080 is on
```

再开一个终端：

```powershell
Invoke-RestMethod http://localhost:8080/health
```

期望：

```json
{"status":"ok"}
```

## 7. 口述检查

1. 当前项目启动时从哪里读取配置？
2. `SHORTLINK_CACHE_TTL=1h` 为什么要转换成 `time.Duration`？
3. 为什么配置文件缺字段时要拒绝启动？
4. 改 HTTP 端口时，为什么还要一起改 `SHORTLINK_BASE_URL`？
