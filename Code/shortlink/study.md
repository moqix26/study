# 短链项目 · 完整学习路线（study.md）

> **怎么用这份文件**  
> 1. 你会员到期后，**只靠这份 + `markdown/`** 也能自学跟完。  
> 2. 每一关都按：目标 → 你要懂什么 → 读哪篇 md → 怎么跑 → 验收 → 口述检查。  
> 3. 代码在本目录规范分层项目里；入口：`go run .`（在 `F:\study\Code\shortlink`）。  
> 4. 相对链接：例如 [环境](./markdown/01-environment.md)。

---

# 第零部分 · 你是谁、学到哪了

## 0.1 背景备忘（给未来的你）

- 方向：Go 后端；目标约 2026 冬实习  
- 已会：竞赛算法底；Go 语法过一遍；**net/http → Gin → GORM/MySQL → Redis Cache Aside**  
- 习惯：完整代码自己敲；Windows + PowerShell；`curl.exe` / `Invoke-RestMethod`  
- 练习库：Docker MySQL `3307` / Redis `6379`（示例密码仅本地）

## 0.2 技术栈已通的证据

你仓库里曾经亲手跑通过（路径可能在 `Code/gin*`）：

| 阶段 | 能力 |
|------|------|
| Go 05 | `/health`、POST/GET JSON、map+锁 |
| Go 06 | Gin CRUD、中间件 `c.Next` |
| Go 07 | GORM、列表 Find、事务 batch 回滚 |
| Go 08 | Redis HIT/MISS、写后删缓存 |

短链 = 把这些接到**一个能写进简历的业务**上。

## 0.3 本项目最终目标（V1）

- [ ] `POST /api/links` 创建短链  
- [ ] `GET /:code` 302 跳转  
- [ ] Redis 缓存加速读  
- [ ] 异步点击计数  
- [ ] 能画架构图、能口述 3 分钟  

**不做进 V1**：完整 JWT 用户系统、网关限流、K8s、分布式 ID（见文末「以后」）。

---

# 第一部分 · 开跑之前

## 关卡 P0 · 打开项目

**目标**：电脑能编译运行 shortlink。

**读**：

- [README.md](./README.md)
- [markdown/00-overview.md](./markdown/00-overview.md)
- [markdown/01-environment.md](./markdown/01-environment.md)

**做**：

```powershell
docker start study-mysql
docker start study-redis
cd F:\study\Code\shortlink
go mod tidy
go run .
```

**验收**：终端出现 `mysql ok` / `redis ok` / `:8080 is on`。

**口述**：

1. 为什么 shortlink 有自己的 `go.mod`？  
2. 本机 MySQL 为什么常用 3307？

**卡住**：看 [09-faq-pitfalls.md](./markdown/09-faq-pitfalls.md) 第 1～2、6 条。

---

## 关卡 P1 · 分层是什么（不必先背代码）

**目标**：能指着目录说出每层干什么。

**读**：[markdown/02-architecture.md](./markdown/02-architecture.md)

**做**：打开这些文件只看函数名（先不抠实现）：

- `internal/handler/http.go`
- `internal/service/link.go`
- `internal/repo/link.go`
- `internal/cache/redis.go`
- `internal/app/app.go`

**口述**：一次「创建短链」请求经过哪些包？

---

# 第二部分 · 数据与短码

## 关卡 D1 · 表结构

**读**：[markdown/03-model-and-db.md](./markdown/03-model-and-db.md)  
**代码**：`internal/model/link.go`

**做**：

```powershell
docker exec study-mysql mysql -uroot -proot123 -e "DESCRIBE study.links;"
```

**口述**：`Code` 为什么要 uniqueIndex？`ClickCount` 为什么可以异步加？

---

## 关卡 D2 · 短码怎么来

**读**：[markdown/04-shortcode.md](./markdown/04-shortcode.md)  
**代码**：`internal/pkg/shortcode/shortcode.go`、`service.Create` 里的重试循环

**想清楚**：

- `crypto/rand` vs `math/rand`  
- 碰撞时谁说了算？（DB 唯一约束）

**小实验（可选）**：把 `SHORTLINK_CODE_LEN` 临时改成 1，疯狂创建，观察是否更多重试（别在生产干）。

---

## 关卡 D3 · URL 校验

**代码**：`internal/pkg/urlx/urlx.go`

**验收思维**：

| 输入 | 期望 |
|------|------|
| `https://a.com` | 成功 |
| `ftp://a.com` | 400 |
| `not a url` | 400 |
| 空 | 400 |

用 POST 试（见下一关）。

---

# 第三部分 · 创建短链

## 关卡 C1 · POST /api/links

**读**：[markdown/05-create-api.md](./markdown/05-create-api.md)

**做**：

```powershell
Invoke-RestMethod -Uri http://localhost:8080/api/links -Method POST -ContentType "application/json" -Body '{"url":"https://www.bilibili.com"}'
```

**验收**：返回 `code`、`short_url`、`long_url`。

**对照阅读顺序**（带着问题看代码）：

1. `handler.CreateLink` 怎么拿 JSON？  
2. `service.Create` 何时返回 400 vs 500？  
3. `repo.Create` 失败且是 duplicate 时上层干什么？

**口述**：创建成功时 Redis 里一定有 key 吗？（V1：**不一定**，第一次跳转才回填。）

---

# 第四部分 · 跳转与缓存（核心）

## 关卡 R1 · 先建立 Cache Aside 直觉

你在 gin-redis 已经做过用户缓存。短链版：

```text
GET /{code}
  Redis GET link:{code}
    HIT  → 302
    MISS → MySQL → SET Redis → 302
```

**读**：[markdown/06-redirect-cache.md](./markdown/06-redirect-cache.md)

**复习对照**：打开旧项目 `Code/gin-redis/main.go` 的 `getUser`，和 `service.Resolve` 并排看。

---

## 关卡 R2 · 302 与 curl

**关键**：观察跳转用 `curl.exe -i`，不要用会自动跟随、藏掉头的工具当唯一手段。

```powershell
# 先拿到 $code
$r = Invoke-RestMethod -Uri http://localhost:8080/api/links -Method POST -ContentType "application/json" -Body '{"url":"https://www.bilibili.com"}'
$code = $r.code

curl.exe -i "http://localhost:8080/$code"
curl.exe -i "http://localhost:8080/$code"
```

**验收**：

1. 状态码 302  
2. `Location: https://www.bilibili.com`  
3. 第一次 `X-Cache: MISS`，第二次 `HIT`

**Redis**：

```powershell
docker exec study-redis redis-cli GET link:$code
```

---

## 关卡 R3 · JSON 查询接口

```powershell
curl.exe -i "http://localhost:8080/api/links/$code"
```

不跳转，适合盯 `X-Cache`。实现与 Redirect 共用 `Resolve`。

---

## 关卡 R4 · 点击计数

跳转几次后：

```powershell
docker exec study-mysql mysql -uroot -proot123 -e "SELECT code,click_count FROM study.links WHERE code='$code';"
```

**口述**：为什么用 goroutine？计数失败用户能不能跳？（能。）

---

## 关卡 R5 · 缓存与一致性（加餐口述）

V1 策略：读时回填；**没有**「改长链」接口，所以简单。

若以后支持「修改目标 URL」：必须 **先写 DB，再 DEL 缓存**（和你学过的用户 PUT 一样），不要只 SET 新值到 Redis（并发下易脏）。

**穿透**（加餐）：疯狂请求不存在的 code → 打爆 DB。对策方向：空值短 TTL、布隆过滤器（面试认识即可，V1 未实现）。

---

# 第五部分 · 配置、日志、工程习惯

## 关卡 E1 · 配置

**读**：[markdown/07-config.md](./markdown/07-config.md)

试着改端口：

```powershell
$env:SHORTLINK_HTTP_ADDR=":8081"
$env:SHORTLINK_BASE_URL="http://localhost:8081"
go run .
```

## 关卡 E2 · 日志中间件

看 `internal/middleware/logger.go`：`c.Next()` 前后打印。  
对应你 Go 06 自己写 Logger 的关卡。

## 关卡 E3 · 验收清单一次性过

**读并执行**：[markdown/08-acceptance.md](./markdown/08-acceptance.md)

全部勾完，V1 算跑通。

---

# 第六部分 · 面试怎么讲（3 分钟稿）

可以背结构，不要背逐字稿：

1. **做什么**：短链生成与跳转服务。  
2. **技术**：Gin + MySQL(GORM) + Redis；Cache Aside。  
3. **创建**：校验 URL → 随机 base62 短码 → DB 唯一约束 + 重试。  
4. **跳转**：先 Redis 后 MySQL，回填 TTL；302。  
5. **计数**：异步 incr，失败不影响跳转。  
6. **取舍**：V1 无登录/限流；已知穿透风险与改进方向。  
7. **踩过的坑**（真实加分）：nil context、Redis 端口写错、PowerShell curl 别名、路由写错 404。

---

# 第七部分 · 从零复习路径（若你忘掉 Gin）

若隔很久回来，按这个最短路径热身（每关 0.5～2h）：

| 顺序 | 内容 | 旧练习 |
|------|------|--------|
| 1 | HTTP JSON + 状态码 | `Code/http` |
| 2 | Gin 路由 + 中间件 | `Code/gin` / gin-users |
| 3 | GORM CRUD + 事务印象 | `Code/gin-mysql` |
| 4 | Redis HIT/MISS | `Code/gin-redis` |
| 5 | 回到本 shortlink study 关卡 P0 |

仓库学习板：`F:\study\learn.md`（总进度）；本文件专注短链。

---

# 第八部分 · 代码导读（按文件）

下面按「打开文件时该想什么」列清单。细节仍以 `markdown/` 为准。

## `cmd/server/main.go` / 根 `main.go`

只调用 `app.Run()`。进程入口保持极瘦。

## `internal/app/app.go`

- Load config  
- Open MySQL → AutoMigrate  
- Ping Redis  
- New repo/cache/service/handler  
- 注册路由：`/health`、`/api/links`、`/api/links/:code`、`/:code`  
- **注意**：`/:code` 放在后面

## `internal/handler/http.go`

HTTP 边界。问自己：这里有没有 SQL？不应有。

## `internal/service/link.go`

业务心脏：`Create` / `Resolve` / `IncrClickAsync`。

## `internal/repo/link.go`

GORM only。`IsDuplicate` 统一判断唯一冲突。

## `internal/cache/redis.go`

key 格式 `link:{code}`，Get/Set/Del。

## `internal/config/config.go`

12-factor 风格环境变量，带默认值。

## `internal/pkg/*`

无框架依赖的纯函数，最好测、最好懂。

---

# 第九部分 · 建议日程（暑假可执行）

假设每天 3～4h：

| 天 | 内容 |
|----|------|
| Day1 | P0～P1 跑起来 + 画目录图 |
| Day2 | D1～D3 + C1 创建接口打通 |
| Day3 | R1～R4 缓存与跳转 + 计数 |
| Day4 | E1～E3 验收 + 写一页自己的 README 笔记 |
| Day5 | 面试稿口述 3 遍；对照 FAQ 复盘坑 |
| Day6+ | 可选 V2（见下）或补计网 HTTP 视频 |

---

# 第十部分 · V2 / 以后可以做什么（先别急着写）

按优先级：

1. **空值缓存** 防穿透  
2. **限流**（令牌桶 / Redis + Lua）  
3. **JWT** 登录后才能创建  
4. **统计看板**：按天 PV  
5. **物理删除/过期时间** 字段  
6. 配置文件 YAML + 优雅退出  

每做一项：先改 `memory/DECISIONS.md` 写清「为什么」，再改代码与 markdown。

---

# 附录 A · 命令速查

```powershell
cd F:\study\Code\shortlink
docker start study-mysql; docker start study-redis
go run .

Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod -Uri http://localhost:8080/api/links -Method POST -ContentType "application/json" -Body '{"url":"https://example.com"}'
curl.exe -i http://localhost:8080/<code>
```

---

# 附录 B · markdown 索引

[00-index.md](./markdown/00-index.md)

---

# 附录 C · 给 AI 续写的人

记忆目录：[memory/README.md](./memory/README.md)  
进度：[memory/PROGRESS.md](./memory/PROGRESS.md)  
决策：[memory/DECISIONS.md](./memory/DECISIONS.md)

---

# 附录 D · 关卡总表（打印用）

- [ ] P0 编译运行  
- [ ] P1 说清分层  
- [ ] D1 看懂表  
- [ ] D2 短码+碰撞  
- [ ] D3 URL 校验  
- [ ] C1 创建成功  
- [ ] R2 302 MISS/HIT  
- [ ] R3 JSON 查询  
- [ ] R4 click_count  
- [ ] E3 验收清单全绿  
- [ ] 面试 3 分钟口述  

全部勾完：短链 V1 毕业。

---

*文风说明：本文件刻意写成「逐步教学」而不是论文。若某关太短，以对应 markdown 详解为准；若 markdown 与代码冲突，以代码 + 验收为准并改文档。*

---

# 第十一部分 · 深度加餐（像聊天里那样讲透）

> 下面这些内容不要求一天读完。卡在哪一块，就精读哪一节。  
> 目标：把「能跑」变成「能讲清楚」。

---

## 11.1 HTTP 在短链里到底发生了什么

### 创建

```text
客户端
  POST /api/links HTTP/1.1
  Content-Type: application/json

  {"url":"https://www.bilibili.com"}
```

服务器：

1. Gin 匹配到 `POST /api/links`  
2. Body 读进结构体  
3. 业务成功 → `201 Created` + JSON  

### 跳转

```text
客户端
  GET /aB3xY9 HTTP/1.1
```

服务器：

1. 找到长链  
2. 响应：

```text
HTTP/1.1 302 Found
Location: https://www.bilibili.com
X-Cache: HIT
```

浏览器（或 curl 跟随）再去请求 Location。  
**短链服务的核心价值**：把「记不住的长 URL」变成「短码 + 一次 302」。

### 为什么验收坚持 `curl.exe -i`

| 工具 | 行为 | 适合 |
|------|------|------|
| `curl.exe -i` | 显示响应头，默认不跟 302（除非 -L） | 看 Location / X-Cache |
| `Invoke-RestMethod` | 常自动处理，跳转细节不明显 | 快速看 JSON 创建结果 |
| 浏览器 | 直接跳走 | 体感 demo |

---

## 11.2 再讲一遍 Cache Aside（对照你的 gin-redis）

### 读路径

```text
应用 → Redis GET
         │
         ├─ 有值 → 用缓存
         └─ 无值 → MySQL SELECT → Redis SET(带 TTL) → 返回
```

### 写路径（若以后有「修改长链」）

```text
应用 → MySQL UPDATE → Redis DEL
```

**为什么删而不是直接 SET 新值？**  
并发下：请求 A 读到旧值、请求 B 写入新值并 SET、请求 A 又把旧值 SET 回去 → 脏缓存。  
DEL 后下次读强制回源，更安全（经典 Cache Aside）。

### 短链 V1 为何创建时不写 Redis？

懒加载：没人访问的短链不占缓存。第一次跳转再灌。也可以改成创建成功立刻 SET——两种都对，选一种能说清即可。

### Redis 挂了怎么办？

本项目：`Resolve` 里 Get 出错只打日志，继续查 MySQL。  
口述：**缓存是加速器，不是真源**。

---

## 11.3 GORM 你真正用到的子集

你不需要背全 API。短链里高频就这些：

| 调用 | 含义 |
|------|------|
| `AutoMigrate(&Link{})` | 按结构体建/改表（学习用） |
| `Create(&link)` | INSERT |
| `Where("code = ?", code).First(&link)` | 条件查一条 |
| `UpdateColumn + Expr` | 字段原子更新 |
| `ErrRecordNotFound` | 没这行 |
| 唯一冲突 | 当碰撞处理 |

事务：你在 gin-mysql 的 batch 练过。短链 V1 创建是单条 INSERT，不必事务；以后「建链 + 写用户配额」再用。

---

## 11.4 Gin 路由顺序为什么重要

```go
r.GET("/health", ...)
r.POST("/api/links", ...)
r.GET("/api/links/:code", ...)
r.GET("/:code", ...)   // 万能匹配，放最后
```

若 `/:code` 在最前，可能把 `health`、`api` 当成短码。  
handler 里仍防御：`code == "api" || code == "health"`。

---

## 11.5 短码空间直觉（面试够用）

- 字母表 62 个字符，长度 6 → \(62^6 \approx 568 亿\)  
- 练习项目随便抽，碰撞概率极低，但仍用 DB 唯一约束兜底  
- 不要用可预测的自增当公开短码（可被扫库）

---

## 11.6 异步计数的一致性模型

```text
用户跳转成功 ← 主路径（必须快、必须对）
click_count+1 ← 尽力而为（允许偶尔丢）
```

这叫**最终一致**的一种弱形式：统计可以差一点，跳转不能错。  
若计数是计费核心，就不能这么随意（要换成可靠消息/同事务等）——面试说出边界就加分。

---

## 11.7 和「单文件 main.go」对照学习法

1. 打开旧的逻辑（或 [markdown/main.go.md](./markdown/main.go.md)）  
2. 找到 `createLink` / `loadLongURL` / `redirectLink`  
3. 在新项目里找同名职责：`service.Create` / `service.Resolve` / `handler.Redirect`  
4. 问自己：这段代码搬到哪一层？有没有多做不该做的事？

这是从「会写 demo」到「会做项目」的关键练习。

---

## 11.8 完整故障排查清单

| 现象 | 先查 |
|------|------|
| 编译失败 import | 是否在 `shortlink` 目录？`go.mod` module 名是否 `shortlink`？ |
| mysql panic | Docker 是否 start？DSN 端口 3307？密码？ |
| redis panic nil ctx | context 是否 Background？ |
| redis refused | 端口是否 6379？ |
| 创建 400 | URL 是否带 http(s)://？ |
| 跳转 404 | code 长度是否等于配置？库里是否有行？ |
| 总是 MISS | Redis 是否 Set 失败？看日志 |
| 看不到 X-Cache | 是否用了 curl.exe -i？ |

---

## 11.9 建议你手写的「一页纸」笔记

自己在本子或文件写（不要复制粘贴）：

1. 架构图（方框：Client / Gin / Service / MySQL / Redis）  
2. 创建序列（3～5 步）  
3. 跳转序列（含 HIT/MISS 两岔）  
4. 三个自己踩过的坑  

面试前只看这一页。

---

## 11.10 V1 完成后的自我测验（闭卷）

1. Cache Aside 读写各怎么做？  
2. 为什么短码要唯一索引？  
3. 302 和 301 对短链通常选哪个？为什么？（提示：302 更灵活，便于改指向）  
4. Redis 挂了跳转还能用吗？本项目呢？  
5. `/:code` 为什么放路由最后？  
6. 点击计数为什么异步？  
7. 如何用 curl 证明第二次 HIT？  
8. 配置里 BaseURL 干什么用？  
9. 穿透是什么？V1 防了吗？  
10. 把项目 3 分钟讲给同学听。

答不上的：回对应关卡 + markdown。

---

## 11.11 从 Go 05 到短链的能力地图

```text
ListenAndServe / HandleFunc     → 知道 HTTP 服务器
JSON encode/decode              → 创建接口 body
Gin Router + Context            → 路由与输入输出
Middleware + Next               → 日志
GORM Model/Create/First         → 持久化
Transaction（gin-mysql）        → 以后扩展
Redis Get/Set/Del + TTL         → 缓存
分层 internal/*                 → 项目形态
短链业务                         → 简历故事
```

你已经走完大半；短链是「收束成作品」。

---

## 11.12 每日最小行动（怕拖延时用）

每天只做一件：

- A：跑通 `go run .`  
- B：创建一条短链  
- C：curl 两次看 HIT  
- D：看懂 `Resolve` 30 行  
- E：口述架构 1 分钟  

连续 5 天比一天熬夜有效。

---

## 11.13 文档维护约定（给你自己 / 未来 AI）

- 改 API → 同步 README + markdown/05～08 + study 关卡  
- 改默认端口 → 同步 01-environment + 07-config + 验收  
- 大决策 → 写 `memory/DECISIONS.md`  
- 进度 → `memory/PROGRESS.md`

---

## 11.14 结语

会员没了也不等于学习断了：

1. 打开 [`study.md`](./study.md)（本文件）从 P0 勾  
2. 细节进 [`markdown/`](./markdown/)  
3. 代码在 `internal/`，跑起来是最好的老师  

短链 V1 的意义，不只是 302，而是证明你能把 **HTTP + DB + Cache** 做成一个完整小系统。
