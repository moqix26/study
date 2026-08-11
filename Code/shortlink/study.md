# 短链项目 · 自学主教材（study.md）

> **怎么用这份文件**
>
> 1. **只跟本文件推进**。像和 AI 聊天一样：问一点 → 推一点 → 验收 → 再下一步。
> 2. **`markdown/` 是精读加餐**。study 说「精读打开 xxx」时再点开，不要两套主线并行通读。
> 3. 工作目录固定：`F:\study\Code\shortlink`。入口命令：`go run ./cmd/server`
> 4. 相对链接示例：[01-cmd-server](./markdown/01-cmd-server.md)

---

## 学法说明

| 习惯 | 做法 |
|------|------|
| 推进方式 | 一次只做一个「我说」小节；过了【过关】再往下 |
| 卡住时 | 把**完整报错**贴给 AI，附上你刚执行的命令 |
| 精读时机 | 【看哪里】里标了 markdown 的，打开对应篇逐块看 |
| 验收 | 每条【怎么验收】都要亲手跑；口述题要能脱稿说 30 秒 |
| 工具 | Windows PowerShell；HTTP 用 `Invoke-RestMethod` 或 `curl.exe` |

---

## 背景（你已具备的前提）

- **会**：Go 语法、Gin 路由/中间件、GORM + MySQL、Redis Cache Aside 基本读写
- **环境**：Docker 容器 `study-mysql`（宿主机 **3307**）、`study-redis`（**6379**）；库名 `study`；示例密码 `root:root123` **仅本地练习**
- **目标**：把已会的技术接到一个能写进简历的短链 V1 上

---

## 目录

| 步骤 | 内容 | 跳转 |
|------|------|------|
| S0 | 开跑：Docker → `go run ./cmd/server` → health | [S0](#s0-开跑) |
| S1 | 业务长什么样：三条 API + 请求路径图 | [S1](#s1-业务长什么样) |
| S2 | 创建短链：handler → service → urlx → shortcode → repo | [S2](#s2-创建短链) |
| S3 | 跳转 + 缓存：Cache Aside、302、X-Cache、异步点击 | [S3](#s3-跳转--缓存) |
| S4 | 分层与配置：目录职责、依赖方向、配置文件 | [S4](#s4-分层与配置) |
| S5 | 验收 + 口述 + V1 不做 | [S5](#s5-验收--口述--以后不做) |

精读索引：[markdown/00-index.md](./markdown/00-index.md)

---

# S0 · 开跑

> 本步目标：容器起来、服务跑起来、`/health` 返回 ok。  
> 精读：[01-cmd-server](./markdown/01-cmd-server.md) · [02-config](./markdown/02-config.md)

---

## S0.1 启动 Docker 依赖

### 【我说】

先把 MySQL 和 Redis 拉起来。程序连的是宿主机 **3307** 和 **6379**，别记成 3306 或 8379。

### 【先懂】

- 程序不自带数据库，必须先有 MySQL（存短链）和 Redis（读缓存）。
- Docker 把容器内 MySQL 3306 映射到你本机 **3307**，避免和你本机已装的 MySQL 抢端口。
- Redis 默认映射 **6379:6379**，一般不改。

### 【看哪里】

- 配置默认值：`internal/config/config.go`（DSN 里的 `3307`、`6379`）
- 精读：[02-config](./markdown/02-config.md) 第 5～6 节（DSN 拆解、为何 3307）

### 【你做什么】

```powershell
docker start study-mysql
docker start study-redis
docker ps --filter "name=study-"
```

观察 `PORTS` 列：MySQL 应是 `0.0.0.0:3307->3306/tcp`，Redis 应是 `0.0.0.0:6379->6379/tcp`。

### 【怎么验收】

```text
study-mysql   Up   ...   0.0.0.0:3307->3306/tcp
study-redis   Up   ...   0.0.0.0:6379->6379/tcp
```

两个容器都是 `Up`。

### 【卡了】

| 报错 | 修法 |
|------|------|
| `No such container: study-mysql` | 容器还没创建，先按你环境文档 `docker run` 建好 |
| `Cannot connect to the Docker daemon` | 打开 Docker Desktop，等托盘图标就绪 |
| 端口被占用 | `docker ps` 看谁占了 3307/6379；停冲突进程或改映射后同步改 `configs/config.env` |

### 【过关】

能说出口令：**「MySQL 走 3307，Redis 走 6379，库名 study。」** —— 过了再进 S0.2。

---

## S0.2 编译并启动服务

### 【我说】

进项目目录，`go run ./cmd/server`，看到三行字就算服务活了。

### 【先懂】

- `cmd/server/main.go` 是唯一进程入口，只调 `app.Run()`。
- `app.Run()` 依次：读配置 → 连 MySQL → AutoMigrate → 连 Redis → 组装 Gin → 监听端口。
- 启动失败会以 `error` 返回，入口 `log.Fatal` 退出（非 0 退出码）。

### 【看哪里】

- 入口：`cmd/server/main.go`
- 启动全貌：`internal/app/app.go` 的 `Run()`
- 精读：[01-cmd-server](./markdown/01-cmd-server.md)

### 【你做什么】

```powershell
cd F:\study\Code\shortlink
go mod tidy
go run ./cmd/server
```

终端保持运行，不要关。

### 【怎么验收】

```text
mysql ok
redis ok
:8080 is on
```

三行都出现，且进程没有退出。

### 【卡了】

| 报错 | 修法 |
|------|------|
| `go: go.mod file not found` | 没 `cd` 到含 `go.mod` 的 `shortlink` 目录 |
| `mysql: ... connection refused` | MySQL 没起、或 DSN 端口写成 3306 了 |
| `redis: ... connection refused` | Redis 没起、或地址写成 8379 了 |
| `migrate: ...` | 库 `study` 不存在；进容器建库或检查 DSN |

### 【过关】

口令：**「cmd/server 只调 app.Run，业务在 internal。」** —— 过了再进 S0.3。

---

## S0.3 探活 health

### 【我说】

另开一个 PowerShell 窗口，打一下 `/health`，确认 HTTP 层也通了。

### 【先懂】

- `/health` 不查库、不查 Redis，只证明 Gin 在监听。
- 路由注册在 `app.go`：`r.GET("/health", h.Health)`。
- handler 返回 `{"status":"ok"}`。

### 【看哪里】

- `internal/handler/http.go` → `Health`
- `internal/app/app.go` 路由注册段
- 精读：[01-cmd-server](./markdown/01-cmd-server.md) 第 8 节验收命令

### 【你做什么】

```powershell
Invoke-RestMethod http://localhost:8080/health
```

或：

```powershell
curl.exe http://localhost:8080/health
```

### 【怎么验收】

```json
{"status":"ok"}
```

PowerShell 里 `Invoke-RestMethod` 可能直接显示 `status : ok`，等价。

### 【卡了】

| 报错 | 修法 |
|------|------|
| `无法连接到远程服务器` | `go run ./cmd/server` 那窗口没跑、或端口不是 8080 |
| `404` | 路径打错；应是 `/health` 不是 `/Health` |
| PowerShell `curl` 行为怪异 | 用 `curl.exe`（真 curl）或 `Invoke-RestMethod` |

### 【过关】

口令：**「health 只验 HTTP 监听，不验业务。」** —— S0 全部完成，进 S1。

---

## S0 小结

| 验收项 | 状态 |
|--------|------|
| Docker MySQL 3307、Redis 6379 运行中 | ☐ |
| `go run ./cmd/server` 三行输出 | ☐ |
| `/health` 返回 ok | ☐ |
| 能解释为何 MySQL 用 3307 | ☐ |

---

# S1 · 业务长什么样

> 本步目标：知道 V1 只做三件事，能画请求路径文字图。  
> 精读：[12-redirect-cache-flow](./markdown/12-redirect-cache-flow.md) 前半 · [03-app-wire](./markdown/03-app-wire.md) 路由段

---

## S1.1 三条 API 是什么

### 【我说】

短链 V1 就三个对外能力：创建、跳转、查 JSON。别的都是辅助。

### 【先懂】

| 方法 | 路径 | 作用 | 成功响应 |
|------|------|------|----------|
| POST | `/api/links` | 提交长链，生成短码 | `201` + `code` / `short_url` / `long_url` |
| GET | `/:code` | 浏览器访问短链，302 跳到长链 | `302` + `Location` 头 |
| GET | `/api/links/:code` | API 查长链（不跳转） | `200` JSON + `X-Cache: HIT\|MISS` |

额外：`GET /health` 探活（S0 已验）。

### 【看哪里】

- 路由表：`internal/app/app.go` 第 48～51 行
- 精读：[03-app-wire](./markdown/03-app-wire.md) 路由注册段
- 精读：[12-redirect-cache-flow](./markdown/12-redirect-cache-flow.md) §0 三张图

### 【你做什么】

打开 `internal/app/app.go`，对照下面文字图，把四行路由注册念一遍：

```text
r.GET("/health", ...)           // 探活
r.POST("/api/links", ...)       // 创建
r.GET("/api/links/:code", ...)  // JSON 查询（带缓存头）
r.GET("/:code", ...)            // 302 跳转（必须最后注册）
```

### 【怎么验收】

你能不看代码说出：

1. 创建用 POST，路径带 `/api/`
2. 跳转是根路径 `/:code`，返回 302
3. JSON 查询路径是 `/api/links/:code`，响应带 `X-Cache`

### 【卡了】

| 困惑 | 澄清 |
|------|------|
| 为什么 `/:code` 要最后注册？ | Gin 按注册顺序匹配；放前面会把 `/api/links/xxx` 的 `api` 当 code |
| `/:code` 和 `/api/links/:code` 区别？ | 前者浏览器跳转 302；后者返回 JSON 不跳 |

### 【过关】

口令：**「POST 创建，GET 根路径跳转，GET api 路径查 JSON。」** —— 过了再进 S1.2。

---

## S1.2 请求路径文字图

### 【我说】

把一次「用户点短链」从浏览器到数据库的路径画在脑子里（或纸上）。

### 【先懂】

**写路径（创建）**—— 只碰 MySQL，不写 Redis：

```text
客户端
  │  POST /api/links  {"url":"https://example.com"}
  ▼
Gin 路由 → middleware.Logger
  ▼
handler.CreateLink          （解析 JSON）
  ▼
service.Create              （校验 URL → 生成短码 → 入库）
  ▼
repo.Create → MySQL INSERT
  ▼
201 {"code","short_url","long_url"}
```

**读路径（跳转）**—— Cache Aside：先 Redis，miss 再 MySQL，再回填：

```text
客户端
  │  GET /Ab3xYz
  ▼
Gin → handler.Redirect
  ▼
service.Resolve
  ├─ cache.Get(Redis)     → 命中：直接拿 longURL，X-Cache: HIT
  └─ miss → repo.FindByCode(MySQL) → cache.Set 回填 → X-Cache: MISS
  ▼
handler: IncrClickAsync（goroutine 加点击，不阻塞跳转）
  ▼
302 Location: 长链 URL
```

**读路径（JSON 查询）**—— Resolve 逻辑相同，最后返回 JSON 而非 302。

### 【看哪里】

- 精读：[12-redirect-cache-flow](./markdown/12-redirect-cache-flow.md) §0.2 写路径、§0.3 读路径
- 组件关系图：同文件 §0.1

### 【你做什么】

在纸上或笔记里各画一遍「写」和「读」路径。标出：

- 哪一步碰 MySQL
- 哪一步碰 Redis
- 创建时**故意不写** Redis

### 【怎么验收】

闭眼能复述：

> 「创建只写库。读时先 Redis，没有再去 MySQL，查到回填 Redis。跳转额外异步加点击数。」

### 【卡了】

| 困惑 | 澄清 |
|------|------|
| 创建后立刻访问短链，Redis 有吗？ | 第一次没有；第一次读会 MISS，从 MySQL 取并回填 |
| 真源是哪个？ | MySQL `links` 表；Redis 可丢、可过期 |

### 【过关】

口令：**「写只 MySQL，读先 Redis 后 MySQL 再回填。」** —— S1 完成，进 S2。

---

## S1 小结

| 验收项 | 状态 |
|--------|------|
| 能说出三条 API 的方法、路径、响应 | ☐ |
| 能画写路径和读路径文字图 | ☐ |
| 知道 `/:code` 路由要最后注册 | ☐ |

---

# S2 · 创建短链

> 本步目标：跟完创建链路，能 POST 成功、无效 URL 得 400。  
> 精读：[08-urlx](./markdown/08-urlx.md) · [07-shortcode](./markdown/07-shortcode.md) · [04-model-link](./markdown/04-model-link.md) · [05-repo-link](./markdown/05-repo-link.md) · [09-service-link](./markdown/09-service-link.md) Create 段 · [10-handler-http](./markdown/10-handler-http.md) CreateLink 段

---

## S2.1 HTTP 入口 CreateLink

### 【我说】

从用户 POST JSON 开始，看 handler 怎么接请求、怎么分 400 和 500。

### 【先懂】

- 请求体：`{"url": "https://..."}`，字段名 `url`。
- `ShouldBindJSON` 失败 → `400 bad json`。
- 业务错误：`url required` / `invalid url` / `url must be http or https` → `400`；其它 → `500`。
- 成功 → `201` + `CreateResult`。

### 【看哪里】

- `internal/handler/http.go` → `CreateLink`、`createReq`
- 精读：[10-handler-http](./markdown/10-handler-http.md) CreateLink 段

### 【你做什么】

读 `CreateLink` 函数，标出三条返回路径：400 JSON、400 校验、201 成功。

### 【怎么验收】

能回答：handler 里**没有**生成短码、**没有**访问数据库——它只调 `h.svc.Create`。

### 【卡了】

| 困惑 | 澄清 |
|------|------|
| 为什么校验错误在 handler 判字符串？ | V1 简单做法；以后可换自定义 error 类型 |
| 201 和 200 区别？ | REST 惯例：创建资源用 201 Created |

### 【过关】

口令：**「handler 只绑 JSON、调 service、映射状态码。」** —— 过了再进 S2.2。

---

## S2.2 URL 校验 urlx.Normalize

### 【我说】

长链进库前必须先洗干净：去空格、有 scheme、只要 http/https。

### 【先懂】

`urlx.Normalize` 步骤：

1. `TrimSpace`
2. 空 → `url required`
3. `url.Parse`；无 scheme 或无 host → `invalid url`
4. scheme 不是 http/https → `url must be http or https`
5. 返回 `u.String()` 规范化结果

### 【看哪里】

- `internal/pkg/urlx/urlx.go`
- 精读：[08-urlx](./markdown/08-urlx.md)

### 【你做什么】

服务保持运行，测三条边界：

```powershell
# 合法
Invoke-RestMethod -Method POST -Uri http://localhost:8080/api/links `
  -ContentType "application/json" -Body '{"url":"https://www.bilibili.com"}'

# 空 URL
Invoke-RestMethod -Method POST -Uri http://localhost:8080/api/links `
  -ContentType "application/json" -Body '{"url":""}'

# 非 http(s)
Invoke-RestMethod -Method POST -Uri http://localhost:8080/api/links `
  -ContentType "application/json" -Body '{"url":"ftp://x.com"}'
```

后两条若 PowerShell 抛错，用：

```powershell
try {
  Invoke-RestMethod -Method POST -Uri http://localhost:8080/api/links `
    -ContentType "application/json" -Body '{"url":""}'
} catch {
  $_.ErrorDetails.Message
}
```

### 【怎么验收】

| 请求 | 期望 |
|------|------|
| 合法 bilibili | `201`，含 `code`、`short_url`、`long_url` |
| 空 url | `400`，`error: url required` |
| ftp | `400`，`error: url must be http or https` |

### 【卡了】

| 报错 | 修法 |
|------|------|
| `400 bad json` | Body 不是合法 JSON；检查引号、Content-Type |
| 合法 URL 也 400 | 是否少了 `https://` 前缀 |

### 【过关】

口令：**「Normalize 四步：trim、非空、parse、仅 http(s)。」** —— 过了再进 S2.3。

---

## S2.3 短码生成 shortcode.Random

### 【我说】

合法 URL 之后，要生成一个 6 位随机码（默认长度，可配置）。

### 【先懂】

- 字母表：base62（`0-9a-zA-Z`），共 62 个字符。
- 用 `crypto/rand`（密码学安全），不是 `math/rand`。
- 默认长度 `CodeLength = 6`（配置文件 `CODE_LEN`）。
- 模型字段 `Code` 的 gorm tag 是 `size:16`，留余量。

### 【看哪里】

- `internal/pkg/shortcode/shortcode.go`
- `internal/model/link.go` → `Code` 字段
- `internal/config/config.go` → `CodeLength`
- 精读：[07-shortcode](./markdown/07-shortcode.md) · [04-model-link](./markdown/04-model-link.md)

### 【你做什么】

创建一条短链，记下返回的 `code`：

```powershell
$r = Invoke-RestMethod -Method POST -Uri http://localhost:8080/api/links `
  -ContentType "application/json" -Body '{"url":"https://example.com/test"}'
$r.code
$r | ConvertTo-Json
```

检查：`code` 长度是否为 6；字符是否都在 base62 内。

### 【怎么验收】

```json
{
  "code": "xY3kPq",
  "short_url": "http://localhost:8080/xY3kPq",
  "long_url": "https://example.com/test"
}
```

- `short_url` = `BaseURL` + `/` + `code`
- `code` 长度 6（默认值）

### 【卡了】

| 困惑 | 澄清 |
|------|------|
| 6 位够吗？ | 62^6 ≈ 568 亿；练习够用；碰撞靠重试 |
| 为什么不用自增 ID？ | V1 故意练随机码 + 唯一索引；分布式 ID 以后再说 |

### 【过关】

口令：**「crypto/rand + base62，默认 6 位，表字段 size 16。」** —— 过了再进 S2.4。

---

## S2.4 入库 repo.Create 与碰撞重试

### 【我说】

短码写入 MySQL；撞了唯一索引就换一个再试，最多 `MaxRetries` 次。

### 【先懂】

- `repo.Create` → `db.Create(link)`，INSERT 一行。
- `code` 列有 `uniqueIndex`；重复 INSERT → duplicate error。
- `repo.IsDuplicate` 判断：GORM `ErrDuplicatedKey`、或错误信息含 duplicate/unique/1062。
- `service.Create` 循环 `0..MaxRetries-1`（默认 8 次）；全失败 → `failed to allocate code`。
- **创建不写 Redis**——第一次读才回填。

### 【看哪里】

- `internal/repo/link.go` → `Create`、`IsDuplicate`
- `internal/service/link.go` → `Create` 循环
- `internal/model/link.go` 表结构
- 精读：[05-repo-link](./markdown/05-repo-link.md) · [09-service-link](./markdown/09-service-link.md) Create 段

### 【你做什么】

1. 再 POST 同一条长链两次，应得到**两个不同** `code`（各一行记录）。
2. 进 MySQL 确认：

```powershell
docker exec -it study-mysql mysql -uroot -proot123 study -e "SELECT code, long_url, click_count FROM links ORDER BY id DESC LIMIT 5;"
```

### 【怎么验收】

- 两次创建均 `201`，`code` 不同
- MySQL 能看到对应行，`click_count` 为 0
- Redis 里**还没有** `link:{code}`（可选验证）：

```powershell
docker exec -it study-redis redis-cli KEYS "link:*"
# 创建后、访问前：应为空或没有刚建的 code
```

### 【卡了】

| 报错 | 修法 |
|------|------|
| `500 failed to allocate code` | 极罕见；检查 MaxRetries、是否测试数据塞满 |
| `500` 非 duplicate | 看终端日志；可能是 MySQL 连接断开 |
| `1062 Duplicate entry` 但没重试 | 检查 `IsDuplicate` 是否覆盖你的 MySQL 驱动错误格式 |

### 【过关】

口令：**「撞唯一索引就重试，默认最多 8 次，创建不写缓存。」** —— 过了再进 S2.5。

---

## S2.5 创建链路串联验收

### 【我说】

把 S2.1～S2.4 收成一条线，做一次完整创建验收。

### 【先懂】

完整调用链：

```text
POST /api/links
  → handler.CreateLink
  → service.Create
      → urlx.Normalize
      → for loop:
          → shortcode.Random(CodeLength)
          → repo.Create → MySQL
          → duplicate? retry : return CreateResult
  → 201 JSON
```

### 【看哪里】

- 精读：[12-redirect-cache-flow](./markdown/12-redirect-cache-flow.md) §2 链路 A
- 对照源码走一遍上述调用链

### 【你做什么】

```powershell
# 1. 正常创建
$link = Invoke-RestMethod -Method POST -Uri http://localhost:8080/api/links `
  -ContentType "application/json" -Body '{"url":"https://github.com/golang/go"}'
$code = $link.code
Write-Host "code=$code short=$($link.short_url)"

# 2. 无效 URL
curl.exe -s -w "`nHTTP %{http_code}`n" -X POST http://localhost:8080/api/links `
  -H "Content-Type: application/json" -d "{\"url\":\"not-a-url\"}"

# 3. 日志：服务终端应出现 [IN] POST /api/links → [OUT] ... 201
```

### 【怎么验收】

| 项 | 期望 |
|----|------|
| 正常创建 | HTTP 201，三个字段齐全 |
| 无效 URL | HTTP 400 |
| 服务日志 | 有 POST 入站、201 出站记录 |
| 口述 | 能按顺序说 5 个包名：handler → service → urlx → shortcode → repo |

### 【卡了】

| 报错 | 修法 |
|------|------|
| PowerShell curl 转义问题 | 用 `Invoke-RestMethod` 或 `curl.exe` 注意 JSON 引号 |
| 201 但 short_url 不对 | 检查 `configs/config.env` 中的 `BASE_URL` |

### 【过关】

口令：**「创建五层：handler → service → urlx → shortcode → repo，201 返回三字段。」** —— S2 完成，进 S3。

---

## S2 小结

| 验收项 | 状态 |
|--------|------|
| POST 合法 URL 得 201 + code/short_url/long_url | ☐ |
| 空 URL、ftp 得 400 | ☐ |
| 能解释碰撞重试与 MaxRetries | ☐ |
| 知道创建不写 Redis | ☐ |
| MySQL 有记录、click_count=0 | ☐ |

---

# S3 · 跳转 + 缓存

> 本步目标：Cache Aside 读路径、X-Cache HIT/MISS、302 跳转、异步点击、黑名单路由。  
> 精读：[06-cache-redis](./markdown/06-cache-redis.md) · [09-service-link](./markdown/09-service-link.md) Resolve 段 · [10-handler-http](./markdown/10-handler-http.md) Redirect/GetLinkJSON 段 · [12-redirect-cache-flow](./markdown/12-redirect-cache-flow.md) · [11-middleware-logger](./markdown/11-middleware-logger.md)

---

## S3.1 Redis 缓存层 LinkCache

### 【我说】

先搞清 Redis 里 key 长什么样、Get/Set 返回什么。

### 【先懂】

- Key 格式：`link:{code}`（函数 `key(code)` 拼接）。
- `Get`：`redis.Nil` → 未命中 `( "", false, nil )`；有值 → `( longURL, true, nil )`；真错误 → 返回 error。
- `Set`：`SET link:{code} {longURL} EX TTL`（默认 TTL 1h）。
- `Del` 存在但 V1 创建/读路径不用（以后更新长链可用）。

### 【看哪里】

- `internal/cache/redis.go`
- 精读：[06-cache-redis](./markdown/06-cache-redis.md)

### 【你做什么】

读 `Get` 和 `Set` 源码，回答：

1. `redis.Nil` 时算命中还是未命中？
2. TTL 从哪传入？

### 【怎么验收】

- `redis.Nil` → **未命中**（`ok == false`）
- TTL 来自 `config.CacheTTL`，在 `app.Run` 里 `cache.NewLinkCache(rdb, cfg.CacheTTL)`

### 【卡了】

| 困惑 | 澄清 |
|------|------|
| 空字符串值 | `Get` 里 `val == ""` 也当未命中 |
| 为什么不用 Hash | V1 只存 longURL 字符串，String 足够 |

### 【过关】

口令：**「key 是 link:code，Nil 是 miss，Set 带 TTL。」** —— 过了再进 S3.2。

---

## S3.2 service.Resolve：Cache Aside

### 【我说】

读链路核心：先缓存、再库、再回填；找不到返回空串而不是 error。

### 【先懂】

`Resolve(ctx, code)` 逻辑：

1. `len(code) != CodeLength` → 直接 `("", false, nil)`（当作不存在）
2. `cache.Get` → 命中返回 `(longURL, true, nil)`
3. Redis 真错误 → 打日志，**降级**继续查 MySQL
4. `repo.FindByCode` → `gorm.ErrRecordNotFound` → `("", false, nil)`
5. 其它 DB 错误 → 返回 error（handler 变 500）
6. 查到 → `cache.Set` 回填（失败只 log）→ `(longURL, false, nil)`

**关键**：`not found` 是**空串**，不是 error。handler 看到 `longURL == ""` 返回 **404**，不是 500。

### 【看哪里】

- `internal/service/link.go` → `Resolve`
- 精读：[09-service-link](./markdown/09-service-link.md) Resolve 段

### 【你做什么】

用 S2 创建的 `$code`，连续两次 JSON 查询：

```powershell
# 第一次 — 期望 MISS
curl.exe -i http://localhost:8080/api/links/$code

# 第二次 — 期望 HIT
curl.exe -i http://localhost:8080/api/links/$code
```

把 `$code` 换成你真实的 6 位短码。

### 【怎么验收】

| 次序 | X-Cache | HTTP |
|------|---------|------|
| 第 1 次 | `MISS` | 200，body 含 `long_url` |
| 第 2 次 | `HIT` | 200，同样 `long_url` |

可选 Redis 验证：

```powershell
docker exec -it study-redis redis-cli GET "link:你的短码"
# 第一次访问后应有长链 URL
```

### 【卡了】

| 现象 | 修法 |
|------|------|
| 两次都 MISS | Redis 没连上、Set 失败看日志；或 TTL 已过 |
| 两次都 HIT | 正常若第一次已回填；若第一次就该 MISS 却 HIT，查是否之前测过 |
| 404 not found | code 错、长度不是 6、或库里没有 |

### 【过关】

口令：**「Resolve：先 Redis，miss 查 MySQL，回填；找不到返空串非 error。」** —— 过了再进 S3.3。

---

## S3.3 handler：X-Cache 头与 404/500 分界

### 【我说】

handler 怎么把 Resolve 结果翻译成 HTTP；重点分清 404 和 500。

### 【先懂】

`GetLinkJSON` / `Redirect` 共同模式：

```go
longURL, hit, err := h.svc.Resolve(...)
if err != nil        → 500
if longURL == ""     → 404 not found
setCacheHeader(c, hit)  → X-Cache: HIT 或 MISS
// 然后 JSON 200 或 302
```

`setCacheHeader`：hit=true → `HIT`，否则 `MISS`。

### 【看哪里】

- `internal/handler/http.go` → `GetLinkJSON`、`Redirect`、`setCacheHeader`
- 精读：[10-handler-http](./markdown/10-handler-http.md)

### 【你做什么】

```powershell
# 存在的 code
curl.exe -i http://localhost:8080/api/links/$code

# 不存在的 code（6 位但库里没有）
curl.exe -i http://localhost:8080/api/links/ZZZZZZ

# 错误长度
curl.exe -i http://localhost:8080/api/links/abc
```

### 【怎么验收】

| 请求 | HTTP | X-Cache | body |
|------|------|---------|------|
| 真实 code | 200 | HIT 或 MISS | 含 long_url |
| ZZZZZZ | 404 | 无或 MISS | `not found` |
| abc（长度≠6） | 404 | — | `not found` |

**不应出现**：不存在短码却返回 500。

### 【卡了】

| 现象 | 原因 |
|------|------|
| 不存在却 500 | handler 没判 `longURL == ""`；或 service 把 not found 当 error 返回了 |
| 没有 X-Cache 头 | 404 分支在 setHeader 之前就 return 了——正常 |

### 【过关】

口令：**「err→500，空串→404，有 URL 才设 X-Cache。」** —— 过了再进 S3.4。

---

## S3.4 302 跳转 Redirect

### 【我说】

浏览器访问短链根路径，应 302 到长链，并异步加点击。

### 【先懂】

`Redirect` 在 Resolve 成功后：

1. `setCacheHeader`
2. `h.svc.IncrClickAsync(code)` — 开 goroutine 调 `repo.IncrClick`，失败只打日志
3. `c.Redirect(http.StatusFound, longURL)` — **302**

跳转**不等待**点击计数完成。

### 【看哪里】

- `internal/handler/http.go` → `Redirect`
- `internal/service/link.go` → `IncrClickAsync`
- `internal/repo/link.go` → `IncrClick`
- 精读：[10-handler-http](./markdown/10-handler-http.md) Redirect 段 · [12](./markdown/12-redirect-cache-flow.md) 链路 C

### 【你做什么】

```powershell
curl.exe -i http://localhost:8080/$code
```

看响应头：

```text
HTTP/1.1 302 Found
Location: https://...
X-Cache: HIT 或 MISS
```

等 1～2 秒后查点击数：

```powershell
docker exec -it study-mysql mysql -uroot -proot123 study `
  -e "SELECT code, click_count FROM links WHERE code='$code';"
```

多跳几次，`click_count` 应递增。

### 【怎么验收】

- HTTP 302
- `Location` 等于创建时的 `long_url`
- 有 `X-Cache` 头
- `click_count` > 0（访问过至少一次）

### 【卡了】

| 现象 | 修法 |
|------|------|
| 200 而不是 302 | 路径打成了 `/api/links/xxx`；应用根路径 `/$code` |
| click_count 不变 | 看服务日志 `incr click error`；MySQL 是否可写 |
| 浏览器直接打开长链 | 正常；curl `-i` 更能看清 302 |

### 【过关】

口令：**「Redirect：Resolve → X-Cache → 异步 IncrClick → 302。」** —— 过了再进 S3.5。

---

## S3.5 黑名单与路由陷阱

### 【我说】

`/:code` 会吃掉所有单段路径；health、api、favicon 要特别挡掉。

### 【先懂】

`Redirect` 开头：

```go
if code == "health" || code == "api" || code == "favicon.ico" {
    c.JSON(404, ...)
    return
}
```

原因：Gin 把 `/health` 也匹配到 `/:code`（若 health 路由没单独注册或顺序错了）。本项目 `/health` 已单独注册；黑名单是双保险，避免 `/api` 被当成短码。

### 【看哪里】

- `internal/handler/http.go` → `Redirect` 黑名单
- `internal/app/app.go` 路由顺序
- 精读：[12-redirect-cache-flow](./markdown/12-redirect-cache-flow.md) 路由与坑段

### 【你做什么】

```powershell
curl.exe -i http://localhost:8080/health
curl.exe -i http://localhost:8080/api
curl.exe -i http://localhost:8080/favicon.ico
```

### 【怎么验收】

| 路径 | 期望 |
|------|------|
| `/health` | 200 `{"status":"ok"}`（走 Health handler） |
| `/api` | 404 JSON `not found`（黑名单） |
| `/favicon.ico` | 404 JSON `not found`（黑名单） |

### 【卡了】

| 现象 | 原因 |
|------|------|
| `/health` 变 302 | 路由顺序错；`/:code` 在 `/health` 前面了 |
| `/api/links` 404 | 正常；应 POST `/api/links` 不是 GET 根 `/api` |

### 【过关】

口令：**「黑名单 health/api/favicon；具体路由先于 /:code。」** —— 过了再进 S3.6。

---

## S3.6 读路径完整验收

### 【我说】

用一条短链走完「创建 → 首次 MISS → 再次 HIT → 302 跳转 → 点击+1」。

### 【先懂】

Cache Aside 在本项目的取舍：

| 场景 | 行为 |
|------|------|
| 创建 | 只写 MySQL |
| 首次读 | Redis miss → MySQL → 回填 Redis |
| 再次读 | Redis hit |
| Redis 宕机 | Get 报错打日志，降级 MySQL，仍可返回（慢） |
| 不存在短码 | 空串 → 404 |

V1 **不做**：空值缓存（防穿透）、更新/删除时主动删缓存。

### 【看哪里】

- 精读：[12-redirect-cache-flow](./markdown/12-redirect-cache-flow.md) 全文
- 可选：[11-middleware-logger](./markdown/11-middleware-logger.md)（日志 `[IN]`/`[OUT]` 格式）

### 【你做什么】

```powershell
# 0. 新建（换 URL 避免旧缓存干扰）
$new = Invoke-RestMethod -Method POST -Uri http://localhost:8080/api/links `
  -ContentType "application/json" -Body '{"url":"https://go.dev/doc/"}'
$c = $new.code

# 1. JSON 第一次
curl.exe -i http://localhost:8080/api/links/$c
# → X-Cache: MISS

# 2. JSON 第二次
curl.exe -i http://localhost:8080/api/links/$c
# → X-Cache: HIT

# 3. 302 跳转
curl.exe -i http://localhost:8080/$c
# → 302, Location, X-Cache

# 4. 不存在
curl.exe -i http://localhost:8080/api/links/AAAAAA
# → 404

# 5. 点击数
docker exec -it study-mysql mysql -uroot -proot123 study `
  -e "SELECT click_count FROM links WHERE code='$c';"
```

### 【怎么验收】

| 步骤 | 期望 |
|------|------|
| 创建 | 201 |
| JSON ×2 | 先 MISS 后 HIT |
| 302 | Location 正确 |
| 假码 | 404，非 500 |
| click_count | ≥ 1 |

服务日志大致类似：

```text
[IN]  POST /api/links
[OUT] POST /api/links 201 ...
[IN]  GET /api/links/xxxxxx
[OUT] GET /api/links/xxxxxx 200 ... X-Cache=MISS
[IN]  GET /api/links/xxxxxx
[OUT] GET /api/links/xxxxxx 200 ... X-Cache=HIT
[IN]  GET /xxxxxx
[OUT] GET /xxxxxx 302 ... X-Cache=...
```

### 【卡了】

| 现象 | 修法 |
|------|------|
| 日志没有 X-Cache | 看 [11-middleware-logger](./markdown/11-middleware-logger.md)；404 请求可能无缓存头 |
| HIT 但 Redis 无 key | 可能刚被 TTL 过期；等 TTL 或查配置文件中的 `CACHE_TTL` |

### 【过关】

口令：**「首次 MISS、再次 HIT、跳转 302、假码 404、点击异步涨。」** —— S3 完成，进 S4。

---

## S3 小结

| 验收项 | 状态 |
|--------|------|
| X-Cache 头：MISS → HIT | ☐ |
| 302 Location 正确 | ☐ |
| 不存在短码返回 404（非 500） | ☐ |
| click_count 异步递增 | ☐ |
| 能口述 Cache Aside 四步 | ☐ |

---

# S4 · 分层与配置

> 本步目标：指着目录说职责，说清依赖方向与配置文件。
> 精读：[03-app-wire](./markdown/03-app-wire.md) · [02-config](./markdown/02-config.md)

---

## S4.1 目录职责一张表

### 【我说】

分层不是为了炫，是为了改一层别牵动全身。

### 【先懂】

```text
shortlink/
├── cmd/server/main.go      # 唯一进程入口，调 app.Run
├── configs/
│   └── config.env          # 启动时读取的本地配置
├── internal/
│   ├── app/app.go          # 组装：配置、DB、Redis、路由
│   ├── config/config.go    # Load() 与配置文件解析
│   ├── handler/http.go     # HTTP 入出站、状态码
│   ├── service/link.go     # 业务编排：Create、Resolve
│   ├── repo/link.go        # MySQL CRUD
│   ├── cache/redis.go      # Redis Get/Set
│   ├── model/link.go       # 表结构 struct
│   ├── middleware/logger.go# 请求日志
│   └── pkg/
│       ├── urlx/urlx.go      # URL 校验
│       └── shortcode/        # 短码生成
├── go.mod                  # 模块名 shortlink
└── markdown/               # 精读（非运行时）
```

### 【看哪里】

- 精读：[03-app-wire](./markdown/03-app-wire.md)
- 浏览 `internal/` 各包第一行 `package` 声明

### 【你做什么】

闭卷填写：

| 包 | 职责 | 不应做什么 |
|----|------|------------|
| handler | 解析 HTTP、映射状态码 | 直接操作 SQL |
| service | 业务逻辑、编排 repo/cache | 绑定 Gin Context |
| repo | MySQL 读写 | 知道 HTTP 状态码 |
| cache | Redis 读写 | 业务判断 |
| config | 读取和校验配置文件 | 连数据库 |

### 【怎么验收】

能回答：**创建短链时，handler 调了谁？service 调了谁？**  
→ handler → service → urlx + shortcode + repo（不经过 cache）。

### 【卡了】

| 困惑 | 澄清 |
|------|------|
| pkg 和 internal 区别 | `internal` 只能本模块 import；`pkg` 可被外部引（本项目暂无外部引用） |
| 为什么 model 单独 | 表结构被 repo/service 共用，避免循环依赖 |

### 【过关】

口令：**「handler 薄、service 厚、repo/cache 只存取。」** —— 过了再进 S4.2。

---

## S4.2 依赖方向

### 【我说】

依赖只能从上往下指，不能 repo 反过来调 handler。

### 【先懂】

```text
        main / cmd
            │
            ▼
          app ─── config
            │
    ┌───────┼───────┐
    ▼       ▼       ▼
 handler service middleware
            │
      ┌─────┴─────┐
      ▼           ▼
    repo        cache
      │           │
      ▼           ▼
    model       (redis)
      │
   urlx / shortcode（被 service 调）
```

规则：

- **handler** 只依赖 **service**
- **service** 依赖 **repo**、**cache**、**config**、**pkg**
- **repo** 依赖 **model**、GORM
- **cache** 依赖 Redis 客户端
- **app** 负责 `New` 并注入，类似简易 DI

### 【看哪里】

- `internal/app/app.go` 组装顺序
- 精读：[03-app-wire](./markdown/03-app-wire.md) 依赖注入段

### 【你做什么】

打开 `internal/app/app.go`，按启动顺序列出 10 步（从 `config.Load` 到 `r.Run`）。

### 【怎么验收】

能画上述依赖图；能说出：**测试 service 时可以 mock repo/cache，不必起 HTTP**。

### 【卡了】

| 困惑 | 澄清 |
|------|------|
| handler 能否调 repo？ | V1 不行；会破坏分层，以后改不动 |
| 入口要维护业务逻辑吗？ | 不用；`cmd/server` 只负责调用 `app.Run` |

### 【过关】

口令：**「依赖向下：handler→service→repo/cache。」** —— 过了再进 S4.3。

---

## S4.3 为何独立 go.mod

### 【我说】

shortlink 是单独模块，不是整个 study 仓库的一个子文件夹凑合跑。

### 【先懂】

- `go.mod` 第一行 `module shortlink` → import 路径前缀 `shortlink/internal/...`
- 独立模块：自己的依赖版本、自己的 `go.sum`、可单独 `go build` 部署
- 与 `F:\study` 根目录无父子 module 关系；在 `shortlink` 目录内执行 go 命令

### 【看哪里】

- `go.mod`、`go.sum`
- 任意文件的 import：`shortlink/internal/app`

### 【你做什么】

```powershell
cd F:\study\Code\shortlink
go list -m
head -5 go.mod   # 或 Get-Content go.mod -Head 5
```

### 【怎么验收】

- `go list -m` 输出 `shortlink`
- 能解释：为什么写 `import "shortlink/internal/app"` 而不是 `../internal/app`

### 【卡了】

| 困惑 | 澄清 |
|------|------|
| 能否放在 GOPATH 里不用 module？ | Go 1.17+ 默认 module mode；练习项目用 module 是正道 |
| 根 study 仓库要 go.mod 吗？ | 不必；shortlink 自给自足 |

### 【过关】

口令：**「module 名即 import 前缀，与磁盘相对路径无关。」** —— 过了再进 S4.4。

---

## S4.4 配置文件

### 【我说】

启动时只读取 `configs/config.env`。所有配置项都必须写在文件里。

### 【先懂】

| 配置项 | 当前值 | 作用 |
|----------|--------|------|
| `HTTP_ADDR` | `:8080` | 监听地址 |
| `BASE_URL` | `http://localhost:8080` | 拼 short_url |
| `MYSQL_DSN` | `root:root123@tcp(127.0.0.1:3307)/study?...` | MySQL |
| `REDIS_ADDR` | `127.0.0.1:6379` | Redis |
| `CACHE_TTL` | `1h` | 缓存过期 |
| `CODE_LEN` | `6` | 短码长度 |
| `MAX_RETRIES` | `8` | 碰撞重试上限 |

`Load()` 解析 `config.env` 后把字符串转换为配置结构体。缺字段、空值、非法整数或非法时长都会使启动失败。

### 【看哪里】

- `internal/config/config.go`
- `configs/config.env`
- 精读：[02-config](./markdown/02-config.md)

### 【你做什么】

换端口实验（需改文件并重启服务）：

```powershell
# 先在 configs/config.env 中改成：
# HTTP_ADDR=:9090
# BASE_URL=http://localhost:9090
go run ./cmd/server

# 终端 B
Invoke-RestMethod http://localhost:9090/health

# 实验结束后，把 config.env 中两个值改回 8080
```

### 【怎么验收】

- 服务打印 `:9090 is on`
- `/health` 在 9090 可访问
- 能说出：**改 config.env 后重启会生效；缺字段或格式错误会拒绝启动**

### 【卡了】

| 现象 | 修法 |
|------|------|
| 改了 config.env 仍 8080 | 没重启进程；或配置项名拼错 |
| short_url 端口不对 | 只改了 HTTP_ADDR 没改 BASE_URL |
| `ParseDuration` 失败 | 应写 `1h` 不是 `1 hour` |

### 【过关】

口令：**「Load 只读 config.env；配置错误就拒绝启动。」** —— S4 完成，进 S5。

---

## S4 小结

| 验收项 | 状态 |
|--------|------|
| 能说出各 internal 包职责 | ☐ |
| 能画 handler→service→repo/cache | ☐ |
| 能解释独立 go.mod | ☐ |
| 会修改 config.env 中的端口 | ☐ |

---

# S5 · 验收 + 口述 + 以后不做

> 本步目标：完整跑通验收清单、能 3 分钟口述、知道 V1 边界。  
> 精读：[12-redirect-cache-flow](./markdown/12-redirect-cache-flow.md) 验收段 · 可选 [H-main-singlefile](./markdown/H-main-singlefile.md)

---

## S5.1 完整验收清单

### 【我说】

像面试官盯着你一样，从头跑一遍。全部打勾才算 V1 毕业。

### 【先懂】

验收分四块：环境、创建、读取、非功能。

### 【看哪里】

- 精读：[12-redirect-cache-flow](./markdown/12-redirect-cache-flow.md) 验收章节
- 本清单

### 【你做什么】

**A. 环境**

```powershell
docker start study-mysql study-redis
cd F:\study\Code\shortlink
go run ./cmd/server
# mysql ok / redis ok / :8080 is on
Invoke-RestMethod http://localhost:8080/health
```

**B. 创建**

```powershell
$ok = Invoke-RestMethod -Method POST -Uri http://localhost:8080/api/links `
  -ContentType "application/json" -Body '{"url":"https://www.bilibili.com"}'
$ok.code.Length -eq 6
$ok.short_url -match "http://localhost:8080/"

# 400
curl.exe -s -o NUL -w "%{http_code}" -X POST http://localhost:8080/api/links `
  -H "Content-Type: application/json" -d "{\"url\":\"\"}"
# 期望 400
```

**C. 读取与缓存**

```powershell
$c = $ok.code
curl.exe -i http://localhost:8080/api/links/$c | Select-String "X-Cache: MISS"
curl.exe -i http://localhost:8080/api/links/$c | Select-String "X-Cache: HIT"
curl.exe -i http://localhost:8080/$c | Select-String "302"
curl.exe -i http://localhost:8080/api/links/NOPE12 | Select-String "404"
```

**D. 数据一致性**

```powershell
docker exec -it study-mysql mysql -uroot -proot123 study `
  -e "SELECT code, long_url, click_count FROM links WHERE code='$c';"
docker exec -it study-redis redis-cli GET "link:$c"
```

### 【怎么验收】

| # | 项 | ☐ |
|---|-----|---|
| 1 | Docker MySQL **3307**、Redis **6379** 运行 | |
| 2 | `go run ./cmd/server` 三行成功输出 | |
| 3 | GET `/health` → ok | |
| 4 | POST 合法 URL → **201** + 三字段 | |
| 5 | POST 空/非法 URL → **400** | |
| 6 | GET `/api/links/:code` 第一次 **X-Cache: MISS** | |
| 7 | 第二次 **X-Cache: HIT** | |
| 8 | GET `/:code` → **302** + Location | |
| 9 | 不存在短码 → **404**（非 500） | |
| 10 | `click_count` 跳转后递增 | |
| 11 | MySQL 有记录；Redis 首次访问后有 `link:code` | |
| 12 | 服务日志有 `[IN]`/`[OUT]` | |

### 【卡了】

| 失败项 | 回到 |
|--------|------|
| 2～3 | S0 |
| 4～5 | S2 |
| 6～10 | S3 |
| 配置文件/端口 | S4 |

### 【过关】

12 项全勾。—— 过了再进 S5.2。

---

## S5.2 三分钟口述提纲

### 【我说】

验收会跑不够，要能对着面试官讲清楚项目。按下面提纲练到 3 分钟内说完。

### 【先懂】

口述结构：**背景 → 架构 → 写链路 → 读链路 → 技术取舍 → 可扩展**。

### 【看哪里】

- 本提纲
- 可选对照：[H-main-singlefile](./markdown/H-main-singlefile.md)（历史单文件版，帮助回忆演进）

### 【你做什么】

对着镜子或录音，按提纲说一遍，计时。

### 【口述提纲】（建议 2.5～3 分钟）

**1. 项目是什么（15 秒）**  
短链服务 V1：用户 POST 长链得短链；访问短链 302 跳转；提供 JSON 查询接口。Go + Gin + GORM + MySQL + Redis。

**2. 分层架构（30 秒）**  
`handler` 处理 HTTP；`service` 写业务；`repo` 访问 MySQL；`cache` 访问 Redis。`app` 组装依赖。独立 `go.mod`，模块名 `shortlink`。

**3. 创建链路（40 秒）**  
POST `/api/links` → 校验 JSON → `urlx.Normalize` 只允许 http(s) → `shortcode.Random` 生成 6 位 base62 → `repo.Create` 写 MySQL。`code` 唯一索引，碰撞则重试最多 8 次。返回 201。**故意不写缓存**，避免脏数据。

**4. 读链路 + Cache Aside（50 秒）**  
GET 解析 `service.Resolve`：先 Redis `link:code`；miss 则查 MySQL；查到回填 Redis，TTL 默认 1 小时。响应头 `X-Cache: HIT/MISS`。  
跳转接口再 `IncrClickAsync` 异步更新 `click_count`，不阻塞 302。  
找不到记录：`Resolve` 返回空串，handler 回 **404**，不是 500。

**5. 配置与部署（20 秒）**  
`config.Load` 只读取 `configs/config.env`。MySQL 宿主机 **3307**（Docker 映射），Redis **6379**。

**6. V1 不做什么 + 以后（25 秒）**  
未做：JWT 登录、限流、缓存穿透空值、K8s。以后可加：用户体系、布隆过滤器、监控、分库分表。

### 【怎么验收】

- 计时 3 分钟内讲完
- 能白板画出写/读两条箭头图
- 面试官追问「缓存一致性」能答：V1 只缓存读、创建不写；更新/删除未实现故未做失效

### 【卡了】

| 追问 | 参考答 |
|------|--------|
| 为什么不用布隆过滤器？ | V1 量小；404 直接打 MySQL 可接受；以后加 |
| Redis 挂了怎么办？ | Get 报错降级查 MySQL，可用但慢 |
| 短码会撞吗？ | 会；唯一索引 + 重试 |

### 【过关】

口令：**「能 3 分钟讲完写读缓存四条线。」** —— 过了再进 S5.3。

---

## S5.3 V1 刻意不做的事

### 【我说】

知道没做什么，和知道做了什么一样重要——面试常问边界。

### 【先懂】

| 不做 | 原因 | 以后方向 |
|------|------|----------|
| JWT / 用户系统 | V1 聚焦短链核心链路 | 加 user_id 字段、鉴权中间件 |
| 网关限流 | 练习项目无流量压力 | middleware 令牌桶 / Nginx |
| 缓存穿透空值 | 短码随机，恶意查不存在码成本可控 | 缓存空占位、布隆过滤器 |
| K8s 部署 | 本地 Docker 足够 | Deployment + Service + Ingress |
| 分布式 ID | 用随机码 + 唯一索引已能练碰撞 | Snowflake、号段 |
| 更新/删除短链 | V1 只有创建和读 | 加 API + 缓存失效 Del |
| 自定义短码 | 降低复杂度 | 校验保留字 + 唯一索引 |

### 【看哪里】

- 本表
- 可选：[H-main-singlefile](./markdown/H-main-singlefile.md) 对比分层前后差异

### 【你做什么】

每条用一句话解释「为什么 V1 可以不做」。

### 【怎么验收】

被问「你的项目有什么不足」能诚实列 2～3 条，并说出改进思路（不是只说「没时间」）。

### 【卡了】

| 困惑 | 澄清 |
|------|------|
| 不做 JWT 能写简历吗？ | 可以；短链 + 缓存 + 分层已是完整后端项目 |
| 要做空值缓存吗？ | V1 规范里明确不做；别 scope creep |

### 【过关】

口令：**「V1 边界：无登录、无限流、无穿透防护、无 K8s。」** —— 全课完成。

---

## S5.4 可选：对照历史单文件版

### 【我说】

如果你是从一整个 `main.go` 拆过来的，对照一下会更有成就感。

### 【先懂】

- [H-main-singlefile](./markdown/H-main-singlefile.md) 是分层前的逐行精讲
- 逻辑等价：现在功能 = 旧单文件 + 配置化 + `click_count` + 更清晰分层

### 【看哪里】

- [H-main-singlefile](./markdown/H-main-singlefile.md)

### 【你做什么】

翻一眼单文件里的路由注册和 `Resolve`，在分层代码里找到对应函数，标 3 处「搬去哪了」。

### 【怎么验收】

能说出至少 3 个函数从单文件搬到了哪个 `internal/` 包。

### 【卡了】

| 困惑 | 澄清 |
|------|------|
| 单文件还能跑吗？ | 仓库现行是分层版；单文件仅作对照阅读 |
| 该学哪个？ | **以分层版为准**；单文件帮助理解演进 |

### 【过关】

口令：**「单文件是历史；现行逻辑在 internal。」** —— 选修完成。

---

## S5 毕业标准

| 维度 | 标准 |
|------|------|
| 动手 | S5.1 清单 12 项全过 |
| 口述 | S5.2 提纲 3 分钟内流畅 |
| 边界 | S5.3 能列 V1 不做的 4 项 |
| 精读 | 按 [00-index](./markdown/00-index.md) 读过对应篇章 |

---

# 附录 A · 命令速查（Windows PowerShell）

```powershell
# 环境
docker start study-mysql study-redis
cd F:\study\Code\shortlink
go run ./cmd/server

# 健康检查
Invoke-RestMethod http://localhost:8080/health

# 创建
Invoke-RestMethod -Method POST -Uri http://localhost:8080/api/links `
  -ContentType "application/json" -Body '{"url":"https://example.com"}'

# 查 JSON（看响应头用 curl.exe）
curl.exe -i http://localhost:8080/api/links/你的短码

# 跳转
curl.exe -i http://localhost:8080/你的短码

# MySQL
docker exec -it study-mysql mysql -uroot -proot123 study -e "SELECT * FROM links LIMIT 5;"

# Redis
docker exec -it study-redis redis-cli KEYS "link:*"
docker exec -it study-redis redis-cli GET "link:你的短码"
```

---

# 附录 B · 技术事实速查（勿记错）

| 项 | 正确值 |
|----|--------|
| MySQL 宿主机端口 | **3307**（不是 3306） |
| Redis 端口 | **6379**（不是 8379） |
| 数据库名 | `study` |
| 默认短码长度 | 6（`Code` 字段 gorm `size:16`） |
| 缓存响应头 | `X-Cache: HIT` 或 `MISS` |
| 不存在短码 | `Resolve` 返回空串 → handler **404** |
| 创建响应 | **201** + `code` / `short_url` / `long_url` |
| 跳转 | **302** Found |
| 创建写缓存吗 | **不写** |
| 配置文件 | `configs/config.env` |

---

# 附录 C · 精读文档地图

| 源码 | 精读 |
|------|------|
| `cmd/server/main.go` | [01-cmd-server](./markdown/01-cmd-server.md) |
| `internal/config` | [02-config](./markdown/02-config.md) |
| `internal/app` | [03-app-wire](./markdown/03-app-wire.md) |
| `internal/model` | [04-model-link](./markdown/04-model-link.md) |
| `internal/repo` | [05-repo-link](./markdown/05-repo-link.md) |
| `internal/cache` | [06-cache-redis](./markdown/06-cache-redis.md) |
| `internal/pkg/shortcode` | [07-shortcode](./markdown/07-shortcode.md) |
| `internal/pkg/urlx` | [08-urlx](./markdown/08-urlx.md) |
| `internal/service` | [09-service-link](./markdown/09-service-link.md) |
| `internal/handler` | [10-handler-http](./markdown/10-handler-http.md) |
| `internal/middleware` | [11-middleware-logger](./markdown/11-middleware-logger.md) |
| 全链路 | [12-redirect-cache-flow](./markdown/12-redirect-cache-flow.md) |
| 历史单文件 | [H-main-singlefile](./markdown/H-main-singlefile.md) |

---

*文档版本：与 shortlink 分层代码同步。卡住就带报错回到对应 S 步，或打开上表精读篇。*
