# 学习状态板 · 实时更新

> **最后更新**：2026-07-13 晚  
> **用法**：和 AI 聊天时以本文件为准；每聊完一段，AI 会更新「当前学什么」「进度」「今日记录」。  
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

## 2. 当前进度快照（2026-07-13）

### 2.1 各模块完成度（自评 / 10）

| 模块 | 分数 | 说明 |
|------|:----:|------|
| Go 基础语法 | 6.5 | 变量/函数/结构体 OK；**接口、error、包结构**仍虚 |
| Go 并发 | 5.5 | goroutine/channel **大概理解**，不熟练 |
| **Go Web（net/http）** | **6.5** | health + POST + GET 完成；第 4 步内存 map |
| **Go 05 进度** | **第 4 步** | ①health ②POST ③GET ✅ |
| 计网 | 1.5 | 01～03 md 看过但**看不懂**；计划 B 站短课 + curl |
| MySQL / Redis | 0 | 未开始（正常，Go 07～08 再开） |
| 项目（短链） | 1 | 仅 hello / health / 静态站级别 |
| 算法 | 7 | **Go 刷题**，函数名有时不熟；C++ 竞赛底仍在 |
| Git / 环境 | 6 | Go 1.26 + VS Code + go mod OK；Git 基础够用 |

**综合**：约 **4.5 / 10** — 语法阶段可收工，**卡在 Web + 计网**，需换学法（多动手 + 视频，少通读 md）。

---

### 2.2 仓库章节对照

| 路线 | 状态 | 备注 |
|------|------|------|
| Tour of Go | ☑ 看过 | 细节已模糊，不必重看全文 |
| Go 00～04 | ☑ 过一遍 | 接口、并发需**小题巩固**，不重修全文 |
| **Go 05 net/http** | **▶ 第 4 步** | ①～③ 完成；内存 map + mutex |
| Go 06 Gin | ☐ | Go 05 验收后再开 |
| Go 07～11 短链 | ☐ | 12 月前目标：能 demo + 能讲 |
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
- [ ] 内存 map 存 POST 用户，GET 按 id 查，404 查无此人

---

### 2.4 当前卡点（按优先级）

1. **net/http 函数名陌生** — 正常，靠写 + [pkg.go.dev/net/http](https://pkg.go.dev/net/http)，不背
2. **HTTP / 计网无直觉** — md 硬啃失败；改 **视频 30min + curl 跟做**
3. **接口不熟练** — 安排 **1 天 3 道小题**，不重读 Go 03 全文
4. **无后端 API 项目** — Go 05 验收 → Gin 内存 CRUD → 再上 MySQL

---

> **最后更新**：2026-07-13 21:45  

---

## 3. 现在学什么（AI 指定 · 只看这一块）

> **更新于 2026-07-13 21:45**  
> **Go 05 第 1～3 步已完成** ✅（health + POST + GET）

### 当前：**Go 05 · 第 4 步 — 内存 map 存用户**

**文件**：`F:\study\Code\http\http.go`（**你自己改**，AI 不直接动你的代码）

**本关新增**：
- `map[int]User` 当内存数据库
- `sync.RWMutex`：POST 写锁、GET 读锁（HTTP 每请求一个 goroutine，必须防并发写坏 map）
- POST 自动分配 `id`；GET 查不到返回 **404**
- `strconv.Atoi` 把路径里的 `"1"` 转成数字

**验收（按顺序）**：
```powershell
# 1. 重启服务后 POST 两次
Invoke-RestMethod -Uri http://localhost:8080/api/users -Method POST -ContentType "application/json" -Body '{"name":"张三"}'
Invoke-RestMethod -Uri http://localhost:8080/api/users -Method POST -ContentType "application/json" -Body '{"name":"李四"}'

# 2. GET 应返回创建时存的数据
Invoke-RestMethod http://localhost:8080/api/users/1
Invoke-RestMethod http://localhost:8080/api/users/2

# 3. 不存在的 id → 404
Invoke-RestMethod http://localhost:8080/api/users/999
```

**注意**：重启 `go run .` 后 map 会清空，id 从 1 重新开始。

**第 4 步 OK 后**：Go 05 基础闭环完成 → 下一章 **Go 06 Gin**（同样 API 用 Gin 重写一遍）。

---

### 已完成步骤备忘

| 步 | 内容 | 状态 |
|----|------|------|
| ① | GET `/health` | ✅ |
| ② | POST `/api/users` 读 JSON | ✅ |
| ③ | GET `/api/users/:id` 路径参数 | ✅ |
| ④ | 内存 map + mutex，POST 存、GET 查、404 | ▶ 当前 |

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
| Go 官方 API | https://pkg.go.dev/net/http |
| Go 路线 | `后端学习/Go/05-Go标准库与HTTP基础.md` |
| 计网速成 | `前端学习/计算机网络/04-HTTP协议深入.md` §0；02 TCP |
| 视频（试） | B 站 IT营 BV1Rm421N7Jy **按关键词跳**（HTTP/Gin/接口），不从头跟 |
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
