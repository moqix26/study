# 学习状态板 · 实时更新

> **最后更新**：2026-08-07 20:40  
> **用法**：和 AI 聊天时以本文件为准；每聊完一段，AI 会更新「当前学什么」「进度」「今日记录」。  
> **AI 交接**：见 `F:\study\.memory\`（readme.md 为协议，摘要按时间倒序读）  
> **主路线文档**：[`go-backend-learning-plan.md`](go-backend-learning-plan.md) · [`后端学习/Go/00-学习路线图与说明.md`](后端学习/Go/00-学习路线图与说明.md)

---

## 1. 基本信息

| 项 | 内容 |
|----|------|
| 背景 | 双非 · 大一升大二 · CCPC 省金 + ICPC 省银 |
| 方向 | **Go 后端**（Java 路线 MySQL/Redis 八股稍后补） |
| 实习目标 | **2026 冬或更晚**（不急于今夏，但需持续积累） |
| 暑假状态 | 已开始，时间较自由 |
| 最近 4 天学时 | 5h · 7h · 2h · 4h（7.10～7.13 左右） |
| 学习偏好 | 主要看仓库 md；**看不懂时问 AI**；计网 md 吃力补视频；**AI 在聊天里给完整代码，你自己敲，不直接改 `Code/`** |
| 娱乐 | 刷抖音、偶尔游戏；其余时间主要在学 |

---

## 2. 当前进度快照（2026-08-07）

### 2.1 各模块完成度（自评 / 10）

| 模块 | 分数 | 说明 |
|------|:----:|------|
| Go 基础语法 | 6.5 | 变量/函数/结构体 OK；**接口、error、包结构**仍虚 |
| Go 并发 | 5.5 | goroutine/channel **大概理解**，不熟练 |
| **Go Web（net/http）** | **6.5** | health + POST + GET 完成；第 4 步内存 map |
| **Go 05 进度** | **已完成** | ①～④ 含 map+锁 ✅ |
| 计网 | 1.5 | 01～03 md 看过但**看不懂**；计划 B 站短课 + curl |
| MySQL / Redis | 6 | Go 07 GORM + Go 08 Cache Aside 已过关（本地 Docker） |
| 项目（短链） | 5 | V1 规范工程已落地；按 `study.md` 验收/跟学中 |
| 算法 | 7 | **Go 刷题**，函数名有时不熟；C++ 竞赛底仍在 |
| Git / 环境 | 6 | Go 1.26 + VS Code + go mod OK；Git 基础够用 |

**综合**：约 **6 / 10** — Web 栈（Gin/MySQL/Redis）已通；主线转入短链项目跟学与口述验收。

---

### 2.2 仓库章节对照

| 路线 | 状态 | 备注 |
|------|------|------|
| Tour of Go | ☑ 看过 | 细节已模糊，不必重看全文 |
| Go 00～04 | ☑ 过一遍 | 接口、并发需**小题巩固**，不重修全文 |
| **Go 05 net/http** | ☑ | 内存用户 API + RWMutex 已过关 |
| Go 补洞视频 | △ | 赵珊珊课语法到闭包；**并发/网络后半基本没看**（不强制补完） |
| Go 06 Gin | ☑ | CRUD + 自写 Logger 中间件过关（`Code/gin`） |
| Go 07 MySQL | ☑ | CRUD + 列表 Find + 事务 batch 回滚验收通过 |
| Go 08 Redis | ☑ | Cache Aside：MISS→HIT；PUT 后 DEL 再 MISS（`Code/gin-redis`） |
| Go 07～11 短链 | ▶ | **当前主线**：短链 V1 规范工程（`Code/shortlink`） |
| 计网 01～03 | △ 看过 md | 未建立直觉，**暂停通读** |
| 计网 04 HTTP | ☐ 优先 | 配合 Go 05 + 短视频 |
| 计网 02 TCP | ☐ 优先 | curl 连不上时分层排查 |
| Java 06/07 | ☐ | 项目用到 MySQL/Redis 再读 |

---

### 2.3 已验证会做的事

- [x] `go mod init`、`go run`、子目录多个 `main.go`
- [x] 起 **8080** 静态站（`ListenAndServe` + `HandleFunc` + `FileServer` + `ServeFile`）
- [x] 基础语法、并发**能读能改**简单代码
- [x] 独立写出 `/health` JSON **不查 AI**（2026-07-13 晚）
- [x] `curl` 测通 `/health` 200 + JSON（2026-07-13）
- [x] 用 **`curl.exe -v`** 看懂 `>` / `<` 原始报文（2026-07-13）
- [x] GET `/api/users/1` 路径参数（2026-07-13，修复 HandleFunc 注册）
- [x] 内存 map 存 POST 用户，GET 按 id 查，404；含 RUnlock（2026-07-13～14）

---

### 2.4 当前卡点（按优先级）

1. **net/http 函数名陌生** — 正常，靠写 + [pkg.go.dev/net/http](https://pkg.go.dev/net/http)，不背
2. **HTTP / 计网无直觉** — md 硬啃失败；改 **视频 30min + curl 跟做**
3. **接口不熟练** — 安排 **1 天 3 道小题**，不重读 Go 03 全文
4. **无后端 API 项目** — Go 05 验收 → Gin 内存 CRUD → 再上 MySQL

---

## 3. 现在学什么（AI 指定 · 只看这一块）

> **更新于 2026-08-07 20:40**  
> **短链 V1 规范项目已落地** ✅ → 跟 `Code/shortlink/study.md` 学

### 当前：**短链 V1 · 按 study.md 自学/验收**

**路径**：`F:\study\Code\shortlink`

```powershell
cd F:\study\Code\shortlink
docker start study-mysql study-redis
go run .
```

- 学习路线：[`study.md`](Code/shortlink/study.md)  
- 模块详解：[`markdown/`](Code/shortlink/markdown/)  
- AI 记忆：[`memory/`](Code/shortlink/memory/)  

会员到期后：只打开 `study.md` 从 P0 关卡勾选即可。

### Go 05 已完成备忘

| 步 | 内容 | 状态 |
|----|------|------|
| ① | GET `/health` | ✅ |
| ② | POST `/api/users` 读 JSON | ✅ |
| ③ | GET `/api/users/:id` 路径参数 | ✅ |
| ④ | 内存 map + mutex，POST 存、GET 查、404 | ✅ |

**Windows 测 API**（PowerShell）：
```powershell
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod -Uri http://localhost:8080/api/users -Method POST -ContentType "application/json" -Body '{"name":"test"}'
Invoke-RestMethod http://localhost:8080/api/users/1
```

---

## 4. 每日时间模板（4h 版 · 暑假）

| 块 | 时长 | 内容 |
|----|------|------|
| 动手 | 2h | Go 05 / 06 代码 + curl |
| 视频 | 1h | 只补**当前卡住**的点（HTTP / 接口 / Gin） |
| 算法 | 45min | 0～2 题，Go 写，不会查标准库 |
| 复盘 | 15min | 在本文件 §6 记 3 行 |

---

## 5. 学习规则（和 AI 协作）

1. **md 当地图**，看不懂就 **问 AI → 必须跟敲 15min**  
2. **函数名不背**，写三次自然熟  
3. **计网**：视频建立直觉 → md 只查 FAQ / 状态码表  
4. **项目只做短链**，不做商城/教程并行项目  
5. **学习阶段**：AI **在聊天里给完整代码**，你在 `Code/` **自己敲**；AI **不直接改**你的练习文件  
6. 聊天时可说：**「今天学完了，更新 learn.md」** 或 **「我卡在第 X 步」**

---

## 6. 学习日志

### 2026-08-07（20:40）

- **学了啥**：短链 V1 规范工程落地（分层 + Redis + 计数 + study.md/markdown/memory）
- **现在干啥**：按 `Code/shortlink/study.md` 验收；会员到期后靠该目录自学
- **备注**：在 shortlink 目录下 `go run .`（独立 go.mod）

### 2026-08-07（16:51）

- **学了啥**：短链 V1 完整参考（后已升级为分层工程）
- **现在干啥**：（已完成）见上条
- **备注**：以 `study.md` + `internal/` 为准，勿再当单文件作业

### 2026-08-07（16:11）

- **学了啥**：确认 gin-redis main.go 逻辑已基本明白
- **现在干啥**：进入 **短链 V1**（创建 + 302）；8 月主线正式开项目
- **备注**：基础栈够用，不必再扩 users CRUD

### 2026-07-16（21:44）

- **学了啥**：Go 08 验收通过 — Redis 6379 + ctx；MISS→HIT；PUT 后缓存失效再 MISS
- **现在干啥**：Go 08 可收；下一关短链或讲穿透
- **备注**：中途踩过 nil ctx、8379 写错端口

### 2026-07-16（19:56）

- **学了啥**：开 Go 08 Redis — Cache Aside 读用户
- **现在干啥**：Docker 起 Redis → 敲 `gin-redis` 完整代码 → 验收 HIT/MISS
- **备注**：聊天给完整代码，自敲；密码/端口仅本地练习

### 2026-07-16（18:10）

- **学了啥**：`gin-mysql` 列表 + 事务 batch 验收全绿；`["g",""]` 回滚正确
- **现在干啥**：Go 07 可收；下一关 Redis
- **备注**：`main.go` 第 45 行 batch 路由建议写成 `"/api/users/batch"`（少 `/` 目前也能跑）

### 2026-07-16（05:02）

- **学了啥**：决定先巩固 GORM 列表 + 事务，再开 Redis
- **现在干啥**：按聊天讲解扩展 `gin-mysql`（Find + Transaction）
- **备注**：完整代码在聊天，自敲

### 2026-07-16（04:55）

- **学了啥**：Go 07 验收通过 — Docker MySQL:3307 + GORM；POST 后重启仍能 GET；中间件日志正常
- **现在干啥**：建议下一关 Go 08 Redis；可选补 PUT/DELETE/列表
- **备注**：3306 被本机 MySQL 占用时正确改用 3307；终端中文显示 `??` 多半是控制台编码

### 2026-07-15（21:50）

- **学了啥**：开 Go 07；推荐 Docker MySQL + GORM 重写 users API
- **现在干啥**：先 `docker run` 起库 → 建表 → 敲 `gin-mysql` 完整代码
- **备注**：练习密码仅本地用 `root123`

### 2026-07-15（21:36）

- **学了啥**：中间件已改对 — `r.Use(Logger())` + `c.Request.URL.Path`；**Go 06 收工**
- **现在干啥**：下一关 **Go 07 MySQL**
- **备注**：仓库代码已核对正确；旧 log.txt 仍是改前内置 Logger 输出

### 2026-07-15（20:56）

- **学了啥**：`Code/gin` 验收：health + CRUD 全绿；用了内置 `gin.Logger()`，自写 Logger 未挂上且 Path 写错
- **现在干啥**：推荐 5 分钟补自写中间件 → 然后 **Go 07 MySQL**
- **备注**：DELETE 404 在 PowerShell 报红是正常现象

### 2026-07-15（15:10）

- **学了啥**：Gin 第一版代码能写、逻辑能懂；API 名（Context/ShouldBindJSON 等）还不熟；已对照讲解过
- **现在干啥**：Go 06 第 2 步 — 中间件 + PUT/DELETE（仍内存）；要完整代码再说
- **备注**：先修 json tag 空格问题再测

### 2026-07-15（01:25）

- **学了啥**：赵珊珊并发/网络后半基本没看 → 决策直接开 Gin
- **现在干啥**：（已推进）Gin users 已写出，进入巩固关
- **备注**：并发靠已有 mutex 够用

### 2026-07-14（15:30）

- **学了啥**：写好 Codex 交接文档；后改为 `.memory` 协议交接
- **现在干啥**：（已更新）见上条 — 开 Gin
- **备注**：Go 05 第 4 步（含解锁）此前已验收通过

### 2026-07-14（05:00）

- **学了啥**：马士兵/赵珊珊 Go 课，约 2h 倍速跳看到 **第 66 集闭包**；比之前透一点
- **现在干啥**：再看 `defer`（约 67）后建议停语法 → 跳到协程/锁，或回来开 Go 06 Gin
- **备注**：长课只补洞，不作主线

### 2026-07-13（21:45）

- **学了啥**：Go 05 ①～③ **全部 OK**；搞懂 HandleFunc 必须注册对应 handler
- **现在干啥**：Go 05 第 4 步 — 按聊天里的参考代码**自己敲** map 版

### 2026-07-13（21:40）

- **学了啥**：GET 405 原因——`/api/users/` 错绑 POST handler；已改绑 `userByIDHandler`
- **现在干啥**：重启 `go run .` 后三条 Invoke-RestMethod 全绿 → Go 05 前三步完成

### 2026-07-13（晚）

- **学了啥**：项目规划（短链 V2 接口/表/channel）；明确主线仍是 Go 05
- **卡在哪**：net/http 函数名；`/health` **尚未自己写完**
- **现在干啥**：只做 Go 05 第 1 步 ① `/health` + ② `curl -v`
- **备注**：短链规划先记着，**8 月再写代码**

### 2026-07-13（午）

- **学了啥**：20 题进度摸底；用 AI 写了 8080 静态 FileServer（`C:/Users/honor/Desktop/AF/复习/Web`）
- **卡在哪**：net/http 函数名不认识；HTTP/计网理论懵；接口朦朦胧胧
- **明天干啥**：加 `/health` + `curl -v` + 1 条 HTTP 短视频
- **备注**：学习状态不错，除娱乐外主要在啃 md；需改为 **动手优先**

---

## 7. 资源备忘

| 用途 | 资源 |
|------|------|
| AI 交接协议 | `F:\study\.memory\readme.md`（摘要按文件名倒序读最新） |
| Go 官方 API | https://pkg.go.dev/net/http |
| Go 路线 | `后端学习/Go/05-Go标准库与HTTP基础.md` |
| 计网速成 | `前端学习/计算机网络/04-HTTP协议深入.md` §0；02 TCP |
| 视频（试） | B 站 IT营 BV1Rm421N7Jy **按关键词跳**（HTTP/Gin/接口），不从头跟 |
| Go 补洞长课 | 马士兵赵珊珊 BV1nKWAzJEav（已看到约 P66 闭包；跳着看，别当主线） |
| 练习目录 | `F:/study/code/go-daily/`（建议把静态站也迁到此处统一管理） |

---

## 8. 里程碑（倒推）

| 时间 | 目标 |
|------|------|
| **7 月底** | Go 05～06 熟练；curl + 简单 Gin API；计网 02+04 有直觉 |
| **8 月底** | GORM + MySQL + Redis 入门；短链骨架 |
| **10～12 月** | 短链可 demo + 能讲；MySQL/Redis 八股；算法维持 |
| **2026 冬** | 投实习 / 日常实习 |

---

*下次和 AI 聊天：直接说「按 learn.md 今天学什么」或贴代码/ curl 输出。*
