# （历史）单文件 main.go 逐块精讲

> **注意**：当前规范代码已分层到 `internal/` + `cmd/server`。  
> 本文保留帮助你对照「从单文件到分层」。正式学习请跟 [`../study.md`](../study.md) 与 [00-index.md](./00-index.md)。  
> 旧单文件逻辑与现网功能基本等价（现网额外有 click_count、配置化）。

> 对应历史：曾存在根目录超大 `main.go`  
> 目标：把「每一块在干什么」讲清楚。


---

## 0. 这个程序整体在干什么

这是一个**短链 V1** 服务：

1. `POST /api/links`：你提交一个长网址 → 生成 6 位短码 → 存进 MySQL → 返回短链
2. `GET /:code`：浏览器访问短码 → 查出长链 → **302 跳转**
3. `GET /api/links/:code`：不跳转，用 JSON 查看映射，并带 `X-Cache` 看是否命中 Redis

数据流：

```text
创建：HTTP JSON → 校验 URL → 随机短码 → MySQL INSERT
跳转：HTTP GET → Redis？→ MySQL？→ 回填 Redis → 302 Location
```

---

## 1. `package main` 与 `import`

```go
package main
```

- 可执行程序的入口包，必须有 `func main()`。

### 标准库


| 包             | 本文件里干什么                                          |
| ------------- | ------------------------------------------------ |
| `context`     | 给 Redis 命令传「上下文」`ctx`（取消/超时用；这里用最简单的 Background） |
| `crypto/rand` | **密码学安全**的随机数，用来抽短码字符（比 `math/rand` 更适合当唯一码）     |
| `errors`      | 构造错误、`errors.Is` 判断错误类型                          |
| `fmt`         | 打印日志                                             |
| `math/big`    | 和大整数配合 `rand.Int`，在 0～61 之间均匀随机                  |
| `net/http`    | 状态码常量，如 `StatusOK`、`StatusFound`(302)            |
| `net/url`     | 解析、校验用户提交的 URL                                   |
| `strings`     | 剪空格、拼字符串、`Builder` 高效拼短码、小写错误信息                  |
| `time`        | `cacheTTL = time.Hour`、`Link.CreatedAt`          |




### 第三方库


| 包                              | 干什么                     |
| ------------------------------ | ----------------------- |
| `github.com/gin-gonic/gin`     | HTTP 路由、中间件、JSON、重定向    |
| `github.com/redis/go-redis/v9` | Redis 客户端（GET/SET/PING） |
| `gorm.io/gorm`                 | ORM：用结构体操作 MySQL        |
| `gorm.io/driver/mysql`         | GORM 的 MySQL 驱动适配       |


---



## 2. 常量 `const (...)`

```go
const (
	codeLen      = 6
	codeAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	baseURL      = "http://localhost:8080"
	cacheTTL     = time.Hour
	maxRetries   = 8
)
```


| 名字             | 值的含义              | 为什么存在                                 |
| -------------- | ----------------- | ------------------------------------- |
| `codeLen`      | 短码长度固定 6          | 和表字段 `size:6`、校验 `len(code)!=6` 一致    |
| `codeAlphabet` | 62 个字符（数字+大小写字母）  | 每位有 62 种可能，6 位空间约 62^6，碰撞概率低          |
| `baseURL`      | 拼短链用的网站前缀         | 返回 `short_url = baseURL + "/" + code` |
| `cacheTTL`     | Redis 里一条缓存活 1 小时 | 过期后自动没，减少永久脏数据                        |
| `maxRetries`   | 短码撞唯一索引时最多再试 8 次  | 防止死循环；极端情况仍可能失败                       |


常量编译期固定，改业务规则时只改这里即可。

---



## 3. 结构体



### 3.1 `Link`（数据库一行 + JSON 外形）

```go
type Link struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Code      string    `json:"code" gorm:"size:6;uniqueIndex;not null"`
	LongURL   string    `json:"long_url" gorm:"size:2048;not null"`
	CreatedAt time.Time `json:"created_at"`
}
```


| 字段          | 类型          | json / gorm  | 含义                            |
| ----------- | ----------- | ------------ | ----------------------------- |
| `ID`        | `uint`      | 主键           | 数据库自增内部 id（业务上主要用 Code）       |
| `Code`      | `string`    | 唯一索引、非空、最长 6 | 短码，如 `BaLrEf`                 |
| `LongURL`   | `string`    | 非空、最长 2048   | 原始长链接                         |
| `CreatedAt` | `time.Time` | GORM 约定字段名   | 创建时间，AutoMigrate/Create 常会自动填 |


**tag 原理简述：**

- `json:"long_url"`：变成 JSON 时字段名是 `long_url`（蛇形），不是 `LongURL`  
- `gorm:"uniqueIndex"`：数据库对 `code` 建唯一索引 → 两个相同 code 无法同时插入

`AutoMigrate(&Link{})` 会按这个结构去建/对齐表 `links`。

### 3.2 `createLinkRequest`（只用于读 POST body）

```go
type createLinkRequest struct {
	URL string `json:"url"`
}
```


| 字段    | 含义                                |
| ----- | --------------------------------- |
| `URL` | 客户端 JSON 里的 `"url":"https://..."` |


不直接用 `Link` 绑 body，是因为创建时不应该让客户端指定 `id`/`code`。

---



## 4. 全局变量 `var (...)`

```go
var (
	db  *gorm.DB
	rdb *redis.Client
	ctx = context.Background()
)
```


| 名字    | 类型                | 含义                                          |
| ----- | ----------------- | ------------------------------------------- |
| `db`  | `*gorm.DB`        | 全局数据库句柄（连接池）。所有 `Create`/`First` 都用它        |
| `rdb` | `*redis.Client`   | 全局 Redis 客户端                                |
| `ctx` | `context.Context` | 传给 Redis 的上下文；`Background()` = 没有超时/取消的根上下文 |


为什么是指针 / 全局：

- 启动时在 `main` 里初始化一次，各个 handler 共享  
- 注意：赋值必须用 `db, err = ...`（`=`），不要 `:=`，否则会做出局部变量，全局仍是 `nil`

---



## 5. `main()`：启动流程

按执行顺序：

### 5.1 连 MySQL

```go
dsn := "root:root123@tcp(127.0.0.1:3307)/study?charset=utf8mb4&parseTime=True&loc=Local"
db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
```


| 片段                    | 含义                         |
| --------------------- | -------------------------- |
| `dsn`                 | Data Source Name，连接串       |
| `root`                | 用户名                        |
| `root123`             | 密码（仅本地练习）                  |
| `tcp(127.0.0.1:3307)` | 本机 3307（Docker 映射到容器 3306） |
| `/study`              | 数据库名（**斜杠不能少**）            |
| `charset=utf8mb4`     | 字符集                        |
| `parseTime=True`      | 把 DATETIME 解成 `time.Time`  |
| `gorm.Open`           | 打开连接池，得到 `*gorm.DB`        |


失败 → `panic`，进程直接退出（启动期失败比带着坏连接跑更安全）。

### 5.2 `AutoMigrate`

```go
db.AutoMigrate(&Link{})
```

按 `Link` 结构确保表存在、列大致对齐。学习项目够用；生产常用正式迁移工具。

### 5.3 连 Redis

```go
rdb = redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
rdb.Ping(ctx)
```


| 名字     | 含义              |
| ------ | --------------- |
| `Addr` | Redis 地址端口      |
| `Ping` | 发 `PING`，通了说明连上 |




### 5.4 挂路由

```go
r := gin.New()
r.Use(gin.Recovery())
r.Use(Logger())
```


| 调用           | 含义                           |
| ------------ | ---------------------------- |
| `gin.New()`  | 空引擎（不自带默认 Logger）            |
| `Recovery()` | handler panic 时恢复，尽量不把整个进程打死 |
| `Logger()`   | 你自定义的请求日志中间件                 |


```go
r.GET("/health", ...)           // 探活
r.POST("/api/links", createLink)
r.GET("/api/links/:code", getLinkJSON)
r.GET("/:code", redirectLink)   // 必须放后面，避免乱匹配
r.Run(":8080")
```

**路由顺序很重要：**  
若把 `/:code` 写在最前，可能把 `api`、`health` 当成短码。你现在是具体路径在前、通配在后，正确。

`r.Run(":8080")`：监听 8080，阻塞运行。

---



## 6. `Logger()` 中间件

```go
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method   // GET / POST ...
		path := c.Request.URL.Path   // /BaLrEf
		fmt.Println("[IN ]", method, path)
		c.Next()                     // 放行到后面的 handler
		fmt.Println("[OUT]", method, path, "->", c.Writer.Status())
	}
}
```


| 变量/调用                  | 含义                         |
| ---------------------- | -------------------------- |
| 返回类型 `gin.HandlerFunc` | `func(*gin.Context)`       |
| `c`                    | 这一次请求的上下文（工作台）             |
| `c.Request`            | 底层 `*http.Request`         |
| `c.Next()`             | 执行后续中间件和最终业务函数；返回后才跑「后置」日志 |
| `c.Writer.Status()`    | 最终 HTTP 状态码（要在 Next 之后读）   |


---



## 7. `linkKey(code)`

```go
func linkKey(code string) string {
	return "link:" + code
}
```


| 参数/返回  | 含义                          |
| ------ | --------------------------- |
| `code` | 短码，如 `BaLrEf`               |
| 返回值    | Redis 的 key，如 `link:BaLrEf` |


加前缀是为了避免和别的业务 key 撞名。

---



## 8. `randomCode(n)`：生成 n 位随机短码

```go
func randomCode(n int) (string, error) {
	var b strings.Builder
	b.Grow(n)
	max := big.NewInt(int64(len(codeAlphabet)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, max)
		...
		b.WriteByte(codeAlphabet[idx.Int64()])
	}
	return b.String(), nil
}
```


| 名字                   | 含义                                |
| -------------------- | --------------------------------- |
| `n`                  | 要生成几位（这里传入 `codeLen`即 6）          |
| `b`                  | `strings.Builder`，高效拼接字符          |
| `b.Grow(n)`          | 预留容量，减少扩容                         |
| `max`                | 大整数 `62`，表示随机范围上界（不包含 62，即 0..61） |
| `rand.Reader`        | 系统安全随机源                           |
| `rand.Int(..., max)` | 得到 `[0, max)` 的随机下标               |
| `idx.Int64()`        | 把 `*big.Int` 转成普通整数下标             |
| `codeAlphabet[下标]`   | 取出对应字符写入 builder                  |


返回：`(短码字符串, error)`。随机源出错时把 error 往上抛。

---



## 9. `normalizeURL(raw)`：校验长链

```go
func normalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" { return "", errors.New("url required") }
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", errors.New("invalid url")
	}
	return u.String(), nil
}
```


| 名字               | 含义                  |
| ---------------- | ------------------- |
| `raw`            | 用户传入的原始字符串          |
| `TrimSpace`      | 去掉首尾空白              |
| `u`              | 解析后的 `*url.URL`     |
| `u.Scheme`       | 协议，如 `https`（空则不合法） |
| `u.Host`         | 主机名（空则不合法）          |
| 返回的 `u.String()` | 规范化后的 URL 字符串       |


作用：拒绝空串、明显不是 URL 的垃圾输入，再拿去入库。

---



## 10. `createLink`：创建短链（POST /api/links）



### 步骤拆解

1. **读 JSON**

```go
var req createLinkRequest
c.ShouldBindJSON(&req)
```


| 名字               | 含义                    |
| ---------------- | --------------------- |
| `req`            | 存放 `{"url":"..."}`    |
| `ShouldBindJSON` | body → 结构体；失败你自己回 400 |
| `&req`           | 必须指针，才能写入             |


1. **规范化 URL** → `longURL`
  失败 → 400 + error 文案。
2. **循环尝试插入（处理短码碰撞）**

```go
var link Link
for i := 0; i < maxRetries; i++ {
	code, err := randomCode(codeLen)
	link = Link{Code: code, LongURL: longURL}
	err = db.Create(&link).Error
	if err == nil {
		// 成功，返回 201
		return
	}
	if !撞库类错误 {
		// 别的 DB 错误，直接 500
		return
	}
	// 唯一冲突：进入下一轮 for，换个 code
}
```


| 名字                 | 含义                                |
| ------------------ | --------------------------------- |
| `i`                | 第几次尝试                             |
| `code`             | 本轮生成的短码                           |
| `link`             | 本轮要插入的行                           |
| `db.Create(&link)` | `INSERT`；成功后 GORM 会回填 `link.ID` 等 |
| `err == nil`       | 插入成功                              |


成功响应字段：


| JSON 字段     | 来源                          |
| ----------- | --------------------------- |
| `code`      | `link.Code`                 |
| `short_url` | `baseURL + "/" + link.Code` |
| `long_url`  | `link.LongURL`              |


状态码 `201 Created`：表示「新建成功」。

若 8 次都撞唯一索引 → `"failed to allocated code"`（文案里 allocated 少了个 e，不影响运行）。

---



## 11. `isDuplicate(err)`

```go
func isDuplicate(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}
```


| 名字    | 含义              |
| ----- | --------------- |
| `err` | `Create` 失败时的错误 |
| `msg` | 错误信息小写，方便包含判断   |


因为不同 MySQL/驱动报错文案略有差别，除了 `gorm.ErrDuplicatedKey` 再兜底用字符串判断「是不是唯一约束冲突」。

---



## 12. `loadLongURL(code)`：读长链（Cache Aside 核心）

签名：

```go
func loadLongURL(code string) (string, bool, error)
```


| 返回值            | 含义                                   |
| -------------- | ------------------------------------ |
| 第 1 个 `string` | 长链；没有则为 `""`                         |
| 第 2 个 `bool`   | **是否 Redis HIT**（`true`=命中缓存）        |
| 第 3 个 `error`  | 基础设施错误（DB 挂了等）；「找不到」用空字符串表示，不是 error |




### 流程

```text
1. len(code) != 6 → 直接当没有（"", false, nil）
2. key = link:xxx
3. Redis GET
   - 成功且非空 → (长链, true, nil)   // HIT
   - redis.Nil → 正常 miss，继续
   - 其它错误 → 打印，降级去 MySQL
4. MySQL：Where code = ? First
   - ErrRecordNotFound → ("", false, nil)
   - 其它错 → ("", false, err)
5. SET Redis（TTL=cacheTTL），忽略 SET 失败（_ = ...）
6. 返回 (longURL, false, nil)   // 有数据但是 MISS
```


| 局部变量   | 含义               |
| ------ | ---------------- |
| `key`  | Redis 键          |
| `val`  | Redis 读出的字符串（长链） |
| `link` | MySQL 查到的整行      |


`_ = rdb.Set(...).Err()`：`_` 表示故意忽略 error（缓存写失败仍可返回 DB 结果）。

---



## 13. `getLinkJSON`：调试用 JSON 接口

```go
GET /api/links/:code
```


| 变量                  | 含义                                 |
| ------------------- | ---------------------------------- |
| `code`              | `c.Param("code")` 路径参数             |
| `longURL, hit, err` | `loadLongURL` 的三个返回值               |
| `cache`             | `"HIT"` 或 `"MISS"`，写入响应头 `X-Cache` |


给 curl 看缓存是否生效；**不负责跳转**。

---



## 14. `redirectLink`：短链跳转

```go
GET /:code
```


| 步骤                                      | 含义                                                             |
| --------------------------------------- | -------------------------------------------------------------- |
| `code := c.Param("code")`               | 取出短码                                                           |
| `loadLongURL`                           | 同缓存逻辑                                                          |
| `c.Header("X-cache", ...)`              | 标记 HIT/MISS（注意：这里头名字是 `X-cache`，JSON 口是 `X-Cache`，大小写不敏感但建议统一） |
| `c.Redirect(http.StatusFound, longURL)` | **302**，并设置 `Location: longURL`                                |


浏览器收到 302 后会自动再请求长链。  
`StatusFound` 常量值就是 302。

---



## 15. 两张「请求生命周期」对照



### 创建（你测过的 BaLrEf）

```text
POST /api/links  {"url":"https://www.example.com"}
  → ShouldBindJSON
  → normalizeURL
  → randomCode × 可能多次
  → INSERT links
  → 201 + code/short_url/long_url
  （此时故意不写 Redis）
```



### 第一次查询 / 跳转

```text
GET .../BaLrEf
  → Redis 无 → MySQL 有
  → SET link:BaLrEf
  → MISS + 200 JSON 或 302
```



### 第二次

```text
  → Redis 有 → HIT，不再打 MySQL（正常情况）
```

---



## 16. 状态码在本项目中的用法


| 码   | 哪里                 | 含义       |
| --- | ------------------ | -------- |
| 200 | health、getLinkJSON | 成功       |
| 201 | createLink         | 新建成功     |
| 302 | redirectLink       | 临时重定向到长链 |
| 400 | JSON/URL 不合法       | 客户端错     |
| 404 | 短码不存在              | 没有这个资源   |
| 500 | DB/随机码失败等          | 服务器错     |


---

