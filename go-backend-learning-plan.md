# Go 后端学习路线 · 详细计划

> **适用背景**：双非计算机专业 · 大一升大二 · CCPC 省金 + ICPC 省银 · 主攻 Go 后端 · 目标大二下/大二暑假拿到对口实习 · 冲字节/腾讯/阿里等大厂  
> **计划周期**：暑假 8 周（全职） + 大二学年延续  
> **建议每日有效学习**：6–8 小时（含练手，不含摸鱼）  
> **详细教程（仓库内）**：[`后端学习/Go/00-学习路线图与说明.md`](后端学习/Go/00-学习路线图与说明.md) — **每章含代码、FAQ、自测，跟这个学**

---

## 跟学顺序（先看哪份文档）

| 用途 | 文档 |
|------|------|
| **每天学什么章节** | 本文件 §17～§18（日程表） |
| **每章详细怎么学** | [Go/00～15](后端学习/Go/00-学习路线图与说明.md) |
| MySQL/Redis 原理 | [Java/06](后端学习/Java/06-MySQL基础索引与事务.md)、[Java/07](后端学习/Java/07-Redis核心原理与缓存实战.md) |
| 短链架构设计 | [系统设计/08](后端学习/系统设计/08-短链服务设计.md) |
| Go 刷题模板 | [数据结构/13](后端学习/数据结构/13-Go手撕模板与LeetCode刷题.md) |
| Git / Linux / 计网 | [Git/00](前端学习/Git/00-学习路线图与说明.md)、[Linux/00](后端学习/Linux/00-学习路线图与说明.md)、[计网/00](前端学习/计算机网络/00-学习路线图与说明.md) |

---

## 目录

1. [学习目标与里程碑](#1-学习目标与里程碑)
2. [本地目录结构建议](#2-本地目录结构建议)
3. [每日时间分配模板](#3-每日时间分配模板)
4. [技能模块总览](#4-技能模块总览)
5. [模块详解：Git](#5-模块详解git)
6. [模块详解：Linux](#6-模块详解linux)
7. [模块详解：Go 语言](#7-模块详解go-语言)
8. [模块详解：数据结构与算法](#8-模块详解数据结构与算法)
9. [模块详解：计算机网络](#9-模块详解计算机网络)
10. [模块详解：MySQL / SQL](#10-模块详解mysql--sql)
11. [模块详解：Redis](#11-模块详解redis)
12. [模块详解：Go Web 工程](#12-模块详解go-web-工程)
13. [模块详解：操作系统](#13-模块详解操作系统)
14. [模块详解：Docker 与部署](#14-模块详解docker-与部署)
15. [模块详解：分布式入门](#15-模块详解分布式入门)
16. [简历项目：短链服务（主项目）](#16-简历项目短链服务主项目)
17. [暑假 8 周逐周计划](#17-暑假-8-周逐周计划)
18. [暑假 56 天日程表（Day by Day）](#18-暑假-56-天日程表day-by-day)
19. [面试八股清单（Go 后端向）](#19-面试八股清单go-后端向)
20. [推荐资源](#20-推荐资源)
21. [大二学年延续计划](#21-大二学年延续计划)
22. [自检清单](#22-自检清单)

---

## 1. 学习目标与里程碑

### 1.1 暑假结束时应达到的状态

| 能力 | 验收标准 |
|------|----------|
| Go | 熟练写并发程序，能独立用 Gin + GORM 开发 REST API |
| Git | 会 branch/merge/rebase，能协作开发，会写规范 commit |
| Linux | 能在 Linux 上部署项目，会 vim 基本操作、进程/端口排查 |
| SQL/MySQL | 会建表、写复杂查询、懂索引与事务，能优化简单慢查询 |
| Redis | 会用五种基本结构，会缓存设计，能讲穿透/击穿/雪崩 |
| 计网 | 能讲清 TCP/HTTP/HTTPS/DNS，能答「从 URL 到页面」 |
| 项目 | 1 个可演示的 Go 后端完整项目（短链服务），Docker 可跑 |
| 算法 | LeetCode 用 Go 累计 80+ 题（维持竞赛手感，不冲 ACM 强度） |

### 1.2 关键里程碑日期（假设 7 月初放暑假）

| 时间 | 里程碑 |
|------|--------|
| 第 1 周末 | Git + Linux 入门完成，Go 环境跑通 |
| 第 2 周末 | Go 语法 + 并发基础完成，LeetCode Go 20 题 |
| 第 3 周末 | MySQL 入门 + SQL 练习 50 题，计网 TCP/HTTP 学完 |
| 第 4 周末 | Gin Hello World → 用户登录 API 跑通 |
| 第 6 周末 | 短链项目核心功能完成（创建/跳转/统计） |
| 第 8 周末 | 项目 + Redis + Docker + README，开始投实习 |
| 大二上期中 | 第二项目或开源 PR，八股第一轮过完 |
| 大二寒假 | 第一份实习 offer 或进入终面 |

---

## 2. 本地目录结构建议

在 `F:\study` 下建议这样组织：

```
F:\study\
├── go-backend-learning-plan.md      # 本文件
├── notes\                           # 笔记（按模块分子目录）
│   ├── git\
│   ├── linux\
│   ├── go\
│   ├── mysql\
│   ├── redis\
│   ├── network\
│   ├── os\
│   └── interview\                   # 八股卡片
├── code\
│   ├── go-daily\                    # 每日 Go 练习小文件
│   ├── leetcode-go\                 # 算法题 Go 解法
│   ├── sql-practice\                # SQL 练习脚本
│   └── projects\
│       └── shorturl\                # 主项目：短链服务
├── resources\                       # 电子书、 cheatsheet
└── logs\
    └── daily-log.md                 # 每日学习记录（强烈建议）
```

**每日记录模板**（写在 `logs/daily-log.md`）：

```markdown
## 2026-07-08
- 计划：Go slice + map
- 完成：slice 底层笔记、3 道 LeetCode
- 卡点：channel 关闭时机
- 明日：select + worker pool 练习
```

---

## 3. 每日时间分配模板

**暑假全职（6–8h）**

| 时段 | 时长 | 内容 |
|------|------|------|
| 块 1 | 2h | 理论：视频/书/笔记（当天主模块） |
| 块 2 | 2h | 编码：Go 练习 / SQL / 项目 |
| 块 3 | 1.5h | 算法：LeetCode 2–3 题（Go 写） |
| 块 4 | 1h | 复习：八股卡片 / 昨日笔记 |
| 块 5 | 0.5h | 整理：daily-log + Git commit |

**大二学期（课多，3–4h）**

| 块 1 | 1.5h | 项目 or 八股 |
| 块 2 | 1h | 算法 1–2 题 |
| 块 3 | 0.5h | 复习笔记 |

---

## 4. 技能模块总览

| 模块 | 暑假周次 | 优先级 | 与实习关系 |
|------|----------|--------|------------|
| Git | W1 | P0 | 必须，第一天就要会 |
| Linux | W1–W2 | P0 | 部署、面试必问 |
| Go 语言 | W1–W3 | P0 | 主武器 |
| 数据结构/算法 | W1–W8 | P0 | 竞赛底 + 笔试 |
| MySQL/SQL | W2–W4 | P0 | 后端核心 |
| 计算机网络 | W2–W3 | P0 | 面试重灾区 |
| Go Web 工程 | W3–W6 | P0 | 项目 |
| Redis | W4–W5 | P0 | 缓存必考 |
| 操作系统 | W5–W6 | P1 | 大二课 + 面试 |
| Docker | W7 | P1 | 简历加分 |
| 分布式入门 | W7–W8 | P2 | 面试进阶 |
| 消息队列 | W8 / 大二上 | P2 | 项目亮点 |

---

## 5. 模块详解：Git

### 5.1 学习目标

- 理解版本控制概念
- 能独立管理个人项目
- 会用分支协作，能处理常见冲突
- commit message 规范

### 5.2 知识点清单

```
[ ] 安装 Git，配置 user.name / user.email
[ ] 工作区 / 暂存区 / 版本库
[ ] git init / clone / add / commit / status / diff / log
[ ] .gitignore 编写
[ ] branch / checkout / switch / merge
[ ] 冲突产生与解决
[ ] rebase（了解，慎用 force push）
[ ] remote / push / pull / fetch
[ ] stash
[ ] tag
[ ] GitHub / Gitee 创建远程仓库
[ ] Pull Request 流程（fork → branch → PR）
[ ] 常见规范：Conventional Commits（feat/fix/docs）
```

### 5.3 练习任务

1. 在 `code/go-daily` 初始化仓库，每天至少 1 次 commit  
2. 故意制造 merge conflict，练习解决  
3. 把 `shorturl` 项目推到 GitHub，README 写清楚  
4. 给任意开源项目提 1 个文档 typo 修复 PR（大二上完成也行）

### 5.4 推荐学习时长

**3 天集中 + 全程习惯使用**（第 1 周前 3 天）

### 5.5 验收

- [ ] 不看教程完成：clone → 改代码 → branch → merge → push  
- [ ] 能解释 HEAD、commit、branch 是什么  
- [ ] 远程仓库有完整提交历史，无「一大坨 initial commit」

---

## 6. 模块详解：Linux

### 6.1 学习目标

- 能在 Linux 环境下开发和部署 Go 项目
- 会常用命令排查问题
- 面试能答基础 Linux 题

### 6.2 环境建议

- **WSL2（Ubuntu 22.04）** 或虚拟机 VMware/VirtualBox  
- 日常 Go 开发可在 Windows，**部署练习必须在 Linux**

### 6.3 知识点清单

```
文件与目录
[ ] ls / cd / pwd / mkdir / rm / cp / mv / touch / cat / less / head / tail
[ ] 权限：chmod / chown，rwx 数字表示（755、644）
[ ] 软链接 ln -s

文本与编辑
[ ] vim 基础：i、a、:wq、:q!、/ 搜索、dd 删行
[ ] grep / find / wc / sort / uniq
[ ] 管道 | 与重定向 > >>

进程与系统
[ ] ps / top / htop / kill / kill -9
[ ] 前台/后台：&、jobs、fg、bg
[ ] 环境变量：export、echo $PATH、~/.bashrc

网络
[ ] ip addr / ping / curl / wget
[ ] netstat / ss - 查端口（ss -tlnp）
[ ] lsof -i :8080

包管理与服务
[ ] apt update / apt install
[ ] systemctl start/stop/status（了解）

其他
[ ] ssh / scp
[ ] tar / zip
[ ] nohup 后台运行
[ ] 定时任务 crontab（了解）
```

### 6.4 Go 相关 Linux 练习

```bash
# 在 Linux 上编译 Go 项目
GOOS=linux GOARCH=amd64 go build -o shorturl .

# 后台运行
nohup ./shorturl > app.log 2>&1 &

# 查端口占用
ss -tlnp | grep 8080

# 看日志
tail -f app.log
```

### 6.5 推荐学习时长

**W1 后 4 天 + W2 每天 0.5h 巩固**

### 6.6 验收

- [ ] 在 WSL 里从零部署短链项目并 curl 访问  
- [ ] 能查 8080 端口被谁占用并 kill  
- [ ] 能口述 10 个最常用命令及场景

---

## 7. 模块详解：Go 语言

### 7.1 学习路线（分 4 层）

#### 第 1 层：基础语法（W1 末 – W2 初，约 5 天）

```
[ ] 环境：go install、GOROOT、GOPATH、go mod
[ ] 基础类型：bool、string、int、float、byte、rune
[ ] 变量：var、:=、常量 const、iota
[ ] 流程控制：if、for（只有 for）、switch
[ ] 函数：多返回值、命名返回值、defer（调用栈理解）
[ ] 数组 vs 切片 slice：底层、扩容、append、copy
[ ] map：创建、遍历、不存在时的 zero value、非线程安全
[ ] 指针：* 和 &，与 C++ 对比（无指针运算）
[ ] struct：定义、嵌入、标签 tag
[ ] 方法：值接收者 vs 指针接收者
[ ] interface：隐式实现、空接口 any、类型断言
[ ] error 处理：if err != nil 惯用法、errors.Is / errors.As
[ ] package 与可见性（大写导出）
[ ] go fmt、go vet、golangci-lint（可选）
```

**每日编码练习建议**

| 天 | 练习 |
|----|------|
| D1 | Hello World + 命令行参数 os.Args |
| D2 | 文件读写：统计 wc-l 单词数 |
| D3 | slice/map 操作：词频统计 |
| D4 | struct + 方法：简易银行账户 |
| D5 | interface：多种 Shape 求面积 |

#### 第 2 层：并发（W2，约 5 天）⭐ 面试重点

```
[ ] goroutine：go func()、与线程区别
[ ] channel：无缓冲/有缓冲、关闭、range
[ ] select：多路复用、default 非阻塞
[ ] sync.Mutex / RWMutex
[ ] sync.WaitGroup
[ ] sync.Once / sync.Pool（了解）
[ ] context：WithCancel、WithTimeout、WithValue
[ ] 常见模式：worker pool、pipeline、fan-in/fan-out
[ ] 竞态：go run -race
[ ]  goroutine 泄漏场景与避免
```

**并发练习项目**

1. **并发爬虫**：10 个 URL，限制最大 3 并发，channel 收结果  
2. **Worker Pool**：任务队列 + N 个 worker 消费  
3. **带超时的 HTTP 请求**：context.WithTimeout + http.Get  

#### 第 3 层：标准库（W2 末 – W3 初）

```
[ ] net/http：http.Server、Handler、HandlerFunc、中间件写法
[ ] encoding/json：Marshal、Unmarshal、结构体 tag
[ ] io / ioutil（或 io 新 API）
[ ] bufio：带缓冲读写
[ ] strconv、strings、time
[ ] os / path/filepath
[ ] testing：单元测试 Table-Driven Tests
[ ] benchmark：性能测试
[ ] flag：命令行 flag 包
```

#### 第 4 层：进阶（W3 及以后，穿插项目）

```
[ ] go mod 依赖管理、replace、vendor（了解）
[ ] 反射 reflect（了解即可，框架会用）
[ ] 泛型（Go 1.18+）：基本会用
[ ] pprof 性能分析（项目优化时用）
[ ] zap 结构化日志
[ ] viper 配置
```

### 7.2 Go 面试必背（提前收录到 notes/go/interview.md）

|  topic | 要点 |
|--------|------|
| slice 底层 | ptr + len + cap，扩容规则 |
| map 底层 | hmap、bucket、扩容、线程不安全 |
| GMP | G goroutine、M 系统线程、P 处理器，M:N 调度 |
| channel | 有缓冲阻塞条件、关闭后读写行为 |
| defer | LIFO，return 执行顺序 |
| interface | 动态派发、eface vs iface |
| GC | 三色标记、写屏障（了解） |

### 7.3 推荐学习时长

**W1 末 2 天 + W2 整周 + W3 前 3 天**

---

## 8. 模块详解：数据结构与算法

### 8.1 定位（你有 ACM 底）

- **不必**再刷 ACM 难度题  
- **必须**用 Go 熟练写 LeetCode 中等题（面试手撕）  
- 重点：**链表、二叉树、哈希、双指针、二分、栈队列、堆、并查集、简单 DP、图 BFS/DFS**

### 8.2 题单规划（暑假 80 题，Go 写）

#### 基础热身（20 题，W1–W2）

| 题号 | 题目 | 考点 |
|------|------|------|
| 1 | 两数之和 | 哈希 |
| 20 | 有效的括号 | 栈 |
| 21 | 合并两个有序链表 | 链表 |
| 53 | 最大子数组和 | DP/贪心 |
| 70 | 爬楼梯 | DP |
| 121 | 买卖股票 | 贪心 |
| 141 | 环形链表 | 快慢指针 |
| 226 | 翻转二叉树 | 树 |
| 704 | 二分查找 | 二分 |
| 347 | 前 K 个高频元素 | 堆 |
| … | 按 LeetCode 热题 100 刷 | |

**完整热题 100 中优先：**

```
数组/哈希：1, 49, 128, 238, 560
链表：206, 92, 142, 160, 234
栈/队列：155, 739, 84
二叉树：94, 104, 226, 543, 236, 124
二分：33, 34, 153
DP：62, 64, 72, 139, 152, 300, 322
回溯：46, 78, 17
图：200, 994, 207（拓扑）
```

### 8.3 每日算法安排

| 阶段 | 题量/天 | 说明 |
|------|---------|------|
| W1 | 1 题 | 熟悉 Go 语法 |
| W2–W8 | 2–3 题 | 中等为主 |
| 大二上 | 1–2 题/天 | 维持手感 |

### 8.4 竞赛底如何利用

- 笔试：大部分题可在 30 分钟内 AC  
- 面试：讲清楚**时间/空间复杂度**和**为什么这样写**  
- 简历：CCPC 省金 + ICPC 省银 写进简历，**不要丢**

### 8.5 验收

- [ ] LeetCode Go 提交通过 80+  
- [ ] 热题 100 完成 60+  
- [ ] 能在 25 分钟内独立 AC 一道中等题（纸笔/IDE 均可）

---

## 9. 模块详解：计算机网络

### 9.1 学习目标

- 应付后端面试 80% 计网题  
- 理解 HTTP 服务编程背后的原理

### 9.2 知识点清单

```
应用层
[ ] HTTP 1.1 / 2 / 3 区别（了解 2 多路复用）
[ ] HTTP 方法：GET/POST/PUT/DELETE
[ ] 状态码：200/301/302/400/401/403/404/500/502/503
[ ] Header 常见字段：Host、Cookie、Authorization、Content-Type
[ ] HTTPS = HTTP + TLS
[ ] TLS 握手过程（简化版：证书、对称密钥）
[ ] Cookie / Session / Token（JWT）区别

传输层
[ ] TCP vs UDP
[ ] TCP 三次握手、四次挥手（为什么不是两次/三次关闭）
[ ] TCP 可靠传输：序号、确认、重传、滑动窗口
[ ] 流量控制 vs 拥塞控制（概念）
[ ] 粘包拆包（Go net 编程相关）
[ ] 端口概念

网络层 / 其他
[ ] IP 地址、子网掩码（基础）
[ ] DNS 解析流程
[ ] ARP（了解）
[ ] 从浏览器输入 URL 到页面展示（⭐ 超级高频）
[ ] GET 和 POST 区别（语义 + 实践）

实战关联
[ ] 用 curl -v 看 HTTP 请求响应头
[ ] Go http.Client 超时设置
[ ] Keep-Alive
```

### 9.3 学习安排

- **W2**：TCP/UDP + 三次握手四次挥手 + 抓包（Wireshark 可选）  
- **W3**：HTTP/HTTPS + DNS + 「URL 到页面」串联  

### 9.4 验收

- [ ] 能白板画 TCP 三次握手  
- [ ] 能 5 分钟讲「URL 到页面」  
- [ ] 能解释 JWT 放在 Header 里怎么走

---

## 10. 模块详解：MySQL / SQL

### 10.1 学习目标

- 会写生产级 SQL  
- 懂表设计、索引、事务  
- 能在 Go 里用 GORM / sqlx 操作数据库

### 10.2 SQL 基础（W2，3 天）

```sql
-- 必须熟练
SELECT ... FROM ... WHERE ... GROUP BY ... HAVING ... ORDER BY ... LIMIT
JOIN：INNER / LEFT / RIGHT
子查询
聚合：COUNT / SUM / AVG / MAX / MIN
INSERT / UPDATE / DELETE
CREATE TABLE / ALTER TABLE
主键、外键、唯一约束、非空、默认值
```

**练习平台**：LeetCode SQL 50 题 或 SQLBolt + 自建表练习

**LeetCode SQL 推荐题**：175, 176, 181, 182, 196, 197, 511, 512, 584, 595, 627, 1068, 1148

### 10.3 MySQL 进阶（W3–W4）

```
[ ] InnoDB vs MyISAM（事务、行锁）
[ ] 索引：B+ 树原理（为什么用 B+ 不用 B）
[ ] 聚簇索引 vs 非聚簇索引
[ ] 最左前缀原则
[ ] 覆盖索引、索引下推（了解）
[ ] EXPLAIN：type、key、rows、Extra
[ ] 事务 ACID
[ ] 隔离级别：读未提交、读已提交、可重复读、串行化
[ ] 脏读、不可重复读、幻读
[ ] MVCC 概念（可重复读如何解决幻读）
[ ] 行锁、表锁、间隙锁（概念）
[ ] 慢查询优化思路
[ ] 分库分表（了解，大二后再深学）
```

### 10.4 表设计练习：短链项目 ER

```sql
-- users
CREATE TABLE users (
    id         BIGINT PRIMARY KEY AUTO_INCREMENT,
    username   VARCHAR(64) NOT NULL UNIQUE,
    password   VARCHAR(128) NOT NULL,  -- bcrypt 哈希
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- urls
CREATE TABLE urls (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id     BIGINT NOT NULL,
    short_code  VARCHAR(16) NOT NULL UNIQUE,
    long_url    VARCHAR(2048) NOT NULL,
    click_count BIGINT NOT NULL DEFAULT 0,
    expire_at   DATETIME NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_short_code (short_code)
);

-- click_logs（可选，统计用）
CREATE TABLE click_logs (
    id         BIGINT PRIMARY KEY AUTO_INCREMENT,
    url_id     BIGINT NOT NULL,
    ip         VARCHAR(45),
    user_agent VARCHAR(512),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_url_id (url_id)
);
```

### 10.5 Go 操作 MySQL

```go
// GORM 入门路径
// 1. 连接 DSN
// 2. AutoMigrate
// 3. Create / First / Where / Updates / Delete
// 4. 预加载 Preload（如有需要）
// 5. 事务 db.Transaction
```

### 10.6 验收

- [ ] 手写 10 道多表 JOIN SQL  
- [ ] 能解释为什么 `WHERE a != 1` 可能不走索引  
- [ ] 短链项目表结构自己设计并实现

---

## 11. 模块详解：Redis

### 11.1 学习目标

- 缓存设计融入短链项目  
- 能答 Redis 面试八股

### 11.2 知识点清单

```
基础
[ ] 五种类型：String / Hash / List / Set / ZSet 及使用场景
[ ] 过期策略：TTL、惰性删除 + 定期删除
[ ] 持久化：RDB vs AOF（概念）
[ ] 单线程模型（6.0+ 多 IO 线程了解）
[ ] 为什么快：内存、IO 多路复用、单线程无锁

缓存
[ ] 缓存穿透：原因 + 布隆过滤器 / 空值缓存
[ ] 缓存击穿：热点 key + 互斥锁
[ ] 缓存雪崩：过期时间打散 + 集群
[ ] 缓存与数据库一致性：先更 DB 再删缓存、延迟双删（了解）

分布式（了解）
[ ] 主从复制
[ ] 哨兵
[ ] Cluster 分片

Go 客户端
[ ] go-redis/redis/v9 基本用法
[ ] 连接池配置
```

### 11.3 在短链项目中的应用

| 场景 | Key 设计 | 说明 |
|------|----------|------|
| 短码 → 长链 | `shorturl:{code}` | 读多写少，TTL 可选 |
| 防重复长链 | `longhash:{md5}` | 同一 URL 复用短码 |
| 限流 | `ratelimit:{ip}` | 滑动窗口 / 简单 INCR |
| 登录 Token 黑名单 | `blacklist:{jti}` | 可选 |

### 11.4 学习安排

**W4 后 3 天 + W5 前 2 天**，与项目并行

### 11.5 验收

- [ ] Docker 跑 Redis，Go 程序连上  
- [ ] 能讲清穿透/击穿/雪崩及解决方案  
- [ ] 短链跳转走 Redis 缓存，miss 回源 MySQL

---

## 12. 模块详解：Go Web 工程

### 12.1 技术栈

| 组件 | 选型 |
|------|------|
| 框架 | Gin |
| ORM | GORM |
| 配置 | Viper 或环境变量 |
| 日志 | zap |
| 鉴权 | JWT（golang-jwt/jwt） |
| 校验 | validator（gin binding） |
| API 文档 | swaggo/swag（可选） |
| Redis | go-redis/v9 |
| MySQL | GORM + mysql driver |

### 12.2 项目分层结构

```
shorturl/
├── cmd/
│   └── server/
│       └── main.go          # 入口
├── internal/
│   ├── config/              # 配置
│   ├── handler/             # HTTP 处理器
│   ├── service/             # 业务逻辑
│   ├── repository/          # 数据访问
│   ├── model/               # 结构体 / GORM model
│   └── middleware/          # JWT、日志、Recovery、CORS
├── pkg/                     # 可复用工具（hash、短码生成）
├── migrations/              # SQL 迁移（可选）
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

### 12.3 必须掌握的 Web 概念

```
[ ] RESTful 设计
[ ] 中间件链式调用
[ ] 统一响应格式 {code, msg, data}
[ ] 错误码设计
[ ] JWT 签发与校验
[ ] 密码 bcrypt 存储，绝不明文
[ ] 参数校验
[ ] 分页：page + page_size
[ ] 跨域 CORS
[ ] 优雅关机 graceful shutdown
```

### 12.4 学习顺序

1. Gin 路由 + JSON 绑定  
2. 中间件：Logger、Recovery、JWT  
3. GORM CRUD  
4. 整合 Redis  
5. 统一错误处理  
6. 配置文件  
7. 单元测试 handler/service  

---

## 13. 模块详解：操作系统

### 13.1 定位

- 大二上会学 OS 课，暑假先打基础  
- 面试常问：进程线程、死锁、内存、IO

### 13.2 知识点清单

```
[ ] 进程 vs 线程 vs 协程（Go goroutine 对比）
[ ] 进程状态：就绪、运行、阻塞
[ ] 进程间通信：管道、消息队列、共享内存、信号
[ ] 线程同步：互斥锁、信号量
[ ] 死锁：四个必要条件、如何避免
[ ] 虚拟内存、分页、页表
[ ] 用户态 vs 内核态
[ ] 系统调用
[ ] IO 多路复用：select / poll / epoll（Go netpoller 关联）
[ ] 零拷贝（sendfile，了解）
```

### 13.3 学习安排

**W5–W6**，每天 1h 理论，与八股合并

---

## 14. 模块详解：Docker 与部署

### 14.1 学习目标

- 项目 Docker 化，docker-compose 一键起 MySQL + Redis + App

### 14.2 知识点

```
[ ] 镜像 vs 容器
[ ] Dockerfile 编写：多阶段构建 Go 项目
[ ] docker build / run / ps / logs / exec
[ ] docker-compose.yml
[ ] 数据卷 volume（MySQL 数据持久化）
[ ] 端口映射
[ ] .dockerignore
```

### 14.3 Go 多阶段 Dockerfile 模板

```dockerfile
# 构建阶段
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o shorturl ./cmd/server

# 运行阶段
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/shorturl .
EXPOSE 8080
CMD ["./shorturl"]
```

### 14.4 学习安排

**W7 集中 3 天**

---

## 15. 模块详解：分布式入门

### 15.1 W8 概念级（大二再深入）

```
[ ] CAP、BASE（了解）
[ ] 分布式 ID：雪花算法、UUID
[ ] 分布式锁：Redis SET NX EX
[ ] 消息队列：为什么需要、解耦/async
[ ] Kafka / RocketMQ 基本概念（Producer/Consumer/Topic）
[ ] RPC vs REST
[ ] gRPC 是什么（会用 proto 定义即可）
[ ] 限流：令牌桶
[ ] 熔断降级（概念）
```

### 15.2 短链项目可选亮点

- Redis 分布式锁防止短码碰撞  
- 本地消息表 / MQ 异步写 click_log（进阶，可选）

---

## 16. 简历项目：短链服务（主项目）

### 16.1 项目描述（简历用）

> **ShortURL - 高并发短链服务** | Go + Gin + GORM + MySQL + Redis  
> - 实现短链生成（Base62 + 发号器）、302 跳转、访问统计  
> - Redis 缓存热点短码，缓存命中率 XX%，跳转 P99 延迟 XXms  
> - JWT 用户体系，RESTful API，Docker Compose 一键部署  
> - GitHub：https://github.com/xxx/shorturl  

### 16.2 功能清单（分 MVP / 进阶）

#### MVP（必须完成）

- [ ] 用户注册 / 登录（JWT）
- [ ] 创建短链（长链 → 短码）
- [ ] 302 跳转
- [ ] 点击计数
- [ ] 用户查看自己的短链列表
- [ ] Redis 缓存
- [ ] README + 架构图

#### 进阶（加分）

- [ ] 短链过期时间
- [ ] 访问日志（IP、UA）
- [ ] 简单 QPS 限流
- [ ] Swagger API 文档
- [ ] 单元测试覆盖率 > 40%
- [ ] GitHub Actions CI

### 16.3 API 设计参考

```
POST   /api/v1/register
POST   /api/v1/login
POST   /api/v1/urls          创建短链
GET    /api/v1/urls          列表（分页）
GET    /api/v1/urls/:code    详情
DELETE /api/v1/urls/:code    删除
GET    /:code                302 跳转（可放根路径）
```

### 16.4 短码生成方案

1. **自增 ID + Base62**（推荐 MVP）  
2. MurmurHash 长链 + 冲突检测  
3. 雪花算法（进阶）

---

## 17. 暑假 8 周逐周计划

### 第 1 周：工具链 + Go 入门

| 天 | 主题 | 产出 |
|----|------|------|
| 1 | Git 安装配置 + 基础命令 | notes/git + 初始化 code 仓库 |
| 2 | Git branch/merge + GitHub | 远程仓库就绪 |
| 3 | Linux/WSL 安装 + 基础命令 | 能在 WSL 里操作文件 |
| 4 | Linux 进程/网络命令 + vim 基础 | 部署 checklist 笔记 |
| 5 | Go 环境 + 基础语法 | go-daily D1–D2 |
| 6 | slice/map/struct | go-daily D3–D4 |
| 7 | 复习 + LeetCode 5 题 Go | W1 复盘 |

### 第 2 周：Go 并发 + 算法加强

| 天 | 主题 | 产出 |
|----|------|------|
| 8 | goroutine + channel | 并发爬虫 demo |
| 9 | select + sync 包 | worker pool |
| 10 | context + net/http 基础 | 带超时 HTTP 客户端 |
| 11 | testing + benchmark | 1 个 Table-Driven Test |
| 12 | MySQL 安装 + SQL 基础 | 10 道 SQL 题 |
| 13 | SQL JOIN + 表设计 | 短链 ER 图 |
| 14 | 计网 TCP/UDP + 复习 | TCP 笔记 |

### 第 3 周：计网 + MySQL 进阶 + Gin 入门

| 天 | 主题 | 产出 |
|----|------|------|
| 15 | HTTP/HTTPS/DNS | URL→页面 笔记 |
| 16 | MySQL 索引 + EXPLAIN | 索引笔记 5 页 |
| 17 | 事务 + 隔离级别 | 八股卡片 |
| 18 | Gin 路由 + JSON | Hello API |
| 19 | Gin 中间件 | Logger + Recovery |
| 20 | JWT 原理 + 实现 | 登录 demo |
| 21 | 周复盘 + 算法 6 题 | — |

### 第 4 周：项目启动 + Redis

| 天 | 主题 | 产出 |
|----|------|------|
| 22 | 项目脚手架 + GORM 连接 | shorturl 骨架 |
| 23 | 用户注册登录 | auth 完成 |
| 24 | 创建短链 API | create 完成 |
| 25 | Redis 安装 + go-redis | 连 Redis 成功 |
| 26 | 跳转 + Redis 缓存 | 302 跑通 |
| 27 | 点击统计 | count 完成 |
| 28 | 周复盘 + 算法 6 题 | — |

### 第 5 周：项目完善 + OS

| 天 | 主题 | 产出 |
|----|------|------|
| 29 | 短链列表分页 | list API |
| 30 | 统一错误码 + 参数校验 | 代码整理 |
| 31 | zap 日志 | 结构化日志 |
| 32 | OS 进程线程 | OS 笔记 |
| 33 | 死锁 + 内存 | OS 笔记 |
| 34 | epoll + Go netpoller | 关联笔记 |
| 35 | 算法 + 项目 bugfix | — |

### 第 6 周：项目进阶 + 压测

| 天 | 主题 | 产出 |
|----|------|------|
| 36 | 限流中间件 | 简单 rate limit |
| 37 | 访问日志表 | click_logs |
| 38 | 单元测试 | service 测试 |
| 39 | wrk/ab 压测 | 压测数据写 README |
| 40 | 代码 review 自己 | refactor |
| 41 | README + 架构图 | 文档 |
| 42 | 周复盘 | — |

### 第 7 周：Docker + 八股

| 天 | 主题 | 产出 |
|----|------|------|
| 43 | Dockerfile | 镜像构建成功 |
| 44 | docker-compose | 一键启动 |
| 45 | Linux 部署演练 | WSL 全流程 |
| 46 | Go 八股第一轮 | 20 题 |
| 47 | MySQL 八股 | 20 题 |
| 48 | Redis + 计网八股 | 20 题 |
| 49 | 模拟面试自问 | 录音复盘 |

### 第 8 周：分布式入门 + 投递准备

| 天 | 主题 | 产出 |
|----|------|------|
| 50 | 分布式 ID + 分布式锁 | 笔记 + 可选代码 |
| 51 | MQ 概念 + gRPC 了解 | 笔记 |
| 52 | 简历第一版 | PDF |
| 53 | GitHub 整理 | 项目可公开 |
| 54 | 牛客/BOSS 实习 JD 调研 | 目标公司列表 |
| 55 | 模拟笔试 2 题 | — |
| 56 | 总复盘 + 大二计划 | 调整 notes |

---

## 18. 暑假 56 天日程表（Day by Day）

> 下面每天列出 **主任务**；算法块默认 **2 题/天**（从 D8 起），D1–D7 为 **1 题/天**。

| Day | 日期参考 | 主任务（≈4–5h） | 副任务（≈1–2h） |
|-----|----------|-----------------|-----------------|
| 1 | 7/07 | Git 安装、config、init/add/commit/log | 环境搭建清单 |
| 2 | 7/08 | branch、merge、.gitignore、远程仓库 | 1 题 LC |
| 3 | 7/09 | WSL Ubuntu、ls/cd/vim 基础 | 1 题 LC |
| 4 | 7/10 | 权限、grep/find、进程命令 | 1 题 LC |
| 5 | 7/11 | Go 安装、go mod、基础类型、函数 | 1 题 LC |
| 6 | 7/12 | slice、map、struct | 1 题 LC |
| 7 | 7/13 | interface、error、周复盘 | 1 题 LC |
| 8 | 7/14 | goroutine、channel | 2 题 LC |
| 9 | 7/15 | select、Mutex、WaitGroup | 2 题 LC |
| 10 | 7/16 | context、worker pool | 2 题 LC |
| 11 | 7/17 | net/http、json、testing | 2 题 LC |
| 12 | 7/18 | MySQL 安装、SELECT/WHERE/JOIN | 2 题 LC |
| 13 | 7/19 | SQL 练习、短链表设计 | 2 题 LC |
| 14 | 7/20 | TCP/UDP、三次握手 | 2 题 LC |
| 15 | 7/21 | HTTP、HTTPS、DNS | 2 题 LC |
| 16 | 7/22 | 索引、B+树、EXPLAIN | 2 题 LC |
| 17 | 7/23 | 事务、隔离级别 | 2 题 LC |
| 18 | 7/24 | Gin 路由、绑定、响应 | 2 题 LC |
| 19 | 7/25 | 中间件、CORS | 2 题 LC |
| 20 | 7/26 | JWT 登录注册 demo | 2 题 LC |
| 21 | 7/27 | 周复盘 | 2 题 LC |
| 22 | 7/28 | shorturl 项目初始化、GORM | 2 题 LC |
| 23 | 7/29 | 用户模块 | 2 题 LC |
| 24 | 7/30 | 创建短链 | 2 题 LC |
| 25 | 7/31 | Redis 基础、go-redis | 2 题 LC |
| 26 | 8/01 | 跳转 + 缓存 | 2 题 LC |
| 27 | 8/02 | 统计 | 2 题 LC |
| 28 | 8/03 | 周复盘 | 2 题 LC |
| 29 | 8/04 | 列表分页 | 2 题 LC |
| 30 | 8/05 | 错误处理、validator | 2 题 LC |
| 31 | 8/06 | zap 日志 | 2 题 LC |
| 32 | 8/07 | OS 进程线程 | 2 题 LC |
| 33 | 8/08 | 死锁、内存 | 2 题 LC |
| 34 | 8/09 | epoll | 2 题 LC |
| 35 | 8/10 | 项目修 bug | 2 题 LC |
| 36 | 8/11 | 限流 | 2 题 LC |
| 37 | 8/12 | click_log | 2 题 LC |
| 38 | 8/13 | 单元测试 | 2 题 LC |
| 39 | 8/14 | 压测 wrk | 2 题 LC |
| 40 | 8/15 | refactor | 2 题 LC |
| 41 | 8/16 | README 架构图 | 2 题 LC |
| 42 | 8/17 | 周复盘 | 2 题 LC |
| 43 | 8/18 | Dockerfile | 2 题 LC |
| 44 | 8/19 | docker-compose | 2 题 LC |
| 45 | 8/20 | Linux 部署 | 2 题 LC |
| 46 | 8/21 | Go 八股 20 题 | 2 题 LC |
| 47 | 8/22 | MySQL 八股 20 题 | 2 题 LC |
| 48 | 8/23 | Redis 计网八股 | 2 题 LC |
| 49 | 8/24 | 模拟面试 | 2 题 LC |
| 50 | 8/25 | 分布式 ID/锁 | 2 题 LC |
| 51 | 8/26 | MQ、gRPC 概念 | 2 题 LC |
| 52 | 8/27 | 写简历 | 2 题 LC |
| 53 | 8/28 | 整理 GitHub | 2 题 LC |
| 54 | 8/29 | 调研 JD | 2 题 LC |
| 55 | 8/30 | 模拟笔试 | 2 题 LC |
| 56 | 8/31 | 总复盘 | — |

---

## 19. 面试八股清单（Go 后端向）

### 19.1 Go 语言（30 题）

1. slice 和 array 区别？slice 底层结构？  
2. slice 扩容规则？  
3. map 是否线程安全？如何实现线程安全？  
4. map 删除元素内存会缩吗？  
5. channel 有缓冲和无缓冲区别？  
6. 关闭 channel 后读写行为？  
7. select 用法？  
8. goroutine 和线程区别？  
9. GMP 模型简述？  
10. defer 执行顺序？defer 和 return 谁先？  
11. interface 底层？空 interface 和带方法 interface？  
12. 值接收者和指针接收者区别？  
13. Go 如何实现面向对象？  
14. make 和 new 区别？  
15. Go GC 算法？如何减少 GC 压力？  
16. context 用途？  
17. goroutine 泄漏怎么排查？  
18. 如何优雅关闭 Go 服务？  
19. Go 错误处理最佳实践？  
20. sync.Map 适用场景？  
21. 原子操作 atomic 用过吗？  
22. Go 泛型了解吗？  
23. init 函数执行顺序？  
24. Go module 版本选择规则？  
25. 如何组织 Go 项目目录？  
26. Go 如何实现单例？  
27. 字符串底层？为什么 string 不可变？  
28. rune 和 byte 区别？  
29. panic 和 recover？  
30. Go 内存逃逸是什么？

### 19.2 MySQL（20 题）

1. 索引数据结构？为什么 B+ 树？  
2. 聚簇索引和非聚簇索引？  
3. 最左前缀原则？  
4. 覆盖索引？  
5. EXPLAIN 关键字段？  
6. 事务 ACID？  
7. 隔离级别分别解决什么问题？  
8. MVCC 是什么？  
9. 脏读、幻读、不可重复读？  
10. 行锁表锁间隙锁？  
11. 慢 SQL 怎么优化？  
12. 主从复制原理？  
13. 分库分表什么时候需要？  
14. 乐观锁悲观锁？  
15. redo log 和 binlog？  
16. 两阶段提交？  
17.  varchar 和 char？  
18. 为什么推荐自增主键？  
19. 联合索引 (a,b,c) 哪些查询能用？  
20. COUNT(*)、COUNT(1) 区别？

### 19.3 Redis（15 题）

1. 五种数据类型及场景？  
2. 为什么快？  
3. 持久化 RDB vs AOF？  
4. 缓存穿透击穿雪崩？  
5. 缓存一致性？  
6. 分布式锁实现？  
7. 过期键删除策略？  
8. 主从哨兵 Cluster 区别？  
9. 热 key 问题？  
10. 大 key 问题？  
11. Redis 单线程为什么还能高并发？  
12. Lua 脚本用途？  
13. 缓存双写不一致怎么办？  
14. Redis 和 Memcached 区别？  
15. 如何用 Redis 实现排行榜？

### 19.4 计算机网络（15 题）

1. OSI 七层 / 五层模型？  
2. TCP vs UDP？  
3. 三次握手四次挥手？  
4. TIME_WAIT 作用？  
5. TCP 可靠传输机制？  
6. HTTP 状态码常见？  
7. GET vs POST？  
8. HTTP 1.1 vs 2.0？  
9. HTTPS 握手？  
10. 从 URL 到页面？  
11. DNS 解析过程？  
12. Cookie Session Token？  
13. JWT 结构？  
14. 跨域是什么？怎么解决？  
15. WebSocket 和 HTTP 区别？

### 19.5 操作系统（10 题）

1. 进程 vs 线程？  
2. 协程 vs 线程？  
3. 死锁条件与预防？  
4. 虚拟内存？  
5. 页面置换算法？  
6. 用户态内核态？  
7. 系统调用流程？  
8. select poll epoll？  
9. 零拷贝？  
10. 线程池参数？

### 19.6 项目 / 场景题

1. 短链怎么生成？冲突怎么办？  
2. 高并发跳转怎么优化？  
3. 为什么用 Redis？命中率多少？  
4. JWT 存在哪？怎么刷新？  
5. 如果 QPS 10 万怎么扩容？（思路即可）

---

## 20. 推荐资源

### 20.1 书籍

| 书名 | 用途 |
|------|------|
| 《Go 程序设计语言》（The Go Programming Language） | Go 主教材 |
| 《Go 语言精进之路》 | 进阶 |
| 《MySQL 必知必会》 | SQL 入门 |
| 《高性能 MySQL》（选读章节） | 索引/事务 |
| 《Redis 设计与实现》（选读） | 深入 Redis |
| 《图解 HTTP》 | 计网入门 |
| 《Unix/Linux 命令参考》（任意简明手册） | Linux |

### 20.2 视频 / 课程（选 1–2 套，勿贪多）

- B站：搜索「Go Zero 微服务」了解生态（不必全学）  
- B站：MySQL 索引/事务专题（任意高赞）  
- 官方：https://go.dev/tour/ （Tour of Go）  
- 官方：https://go.dev/doc/effective_go  

### 20.3 网站

| 网站 | 用途 |
|------|------|
| https://leetcode.cn | 算法 + SQL |
| https://github.com | 代码托管 |
| https://go.dev/pkg/ | 标准库文档 |
| https://gin-gonic.com/docs/ | Gin 文档 |
| https://gorm.io/docs/ | GORM 文档 |
| https://redis.io/docs/ | Redis 文档 |
| 牛客网 | 八股 + 面经 |

### 20.4 工具安装清单

- [ ] Go 1.22+  
- [ ] Git  
- [ ] VS Code / GoLand + Go 插件  
- [ ] WSL2 Ubuntu 22.04  
- [ ] MySQL 8.0（或 Docker）  
- [ ] Redis（或 Docker）  
- [ ] Docker Desktop  
- [ ] Postman 或 curl  
- [ ] DBeaver（SQL 客户端）  

---

## 21. 大二学年延续计划

### 9 月–10 月（大二上）

- 课内：OS、计网、数据库系统 → **与八股对齐**  
- 项目：短链维护 + 开始**第二项目**（秒杀 / Todo 协作 / 简易 IM 选一）  
- 算法：每周 5 题  
- 八股：每天 30min  
- **开始投日常实习**（中小厂、远程、字节跳动/美团等 Go 岗）

### 11 月–1 月（期末 + 寒假）

- 期末降低 side project 强度  
- 寒假：**实习冲刺**，每日 6h 面试准备  
- 目标：拿到第一份后端实习 offer  
- 可选：给 Go 开源提 PR

### 3 月–6 月（大二下）

- 若有实习：全力实习 + 学到生产环境经验  
- 若无实习：继续投 + 完善项目 + 开源  
- 准备大三暑期实习（大厂关键战场）

---

## 22. 自检清单

### 暑假结束前总检查

**工程**
- [ ] GitHub 有 shorturl 项目，commit 记录清晰  
- [ ] README 含：功能、技术栈、架构图、如何运行、压测数据  
- [ ] docker-compose up 能一键启动  
- [ ] 能在 WSL/Linux 独立部署  

**语言**
- [ ] 用 Go 独立写 Gin 服务无教程辅助  
- [ ] 能写 worker pool + context 超时控制  
- [ ] 能解释 GMP、slice、channel  

**数据库**
- [ ] 手写 SQL JOIN 无压力  
- [ ] 能解释索引、事务隔离级别  
- [ ] 项目中表结构自己设计  

**缓存**
- [ ] Redis 五种类型能举例场景  
- [ ] 能讲穿透/击穿/雪崩  

**网络**
- [ ] 能讲 URL 到页面 + TCP 握手  

**算法**
- [ ] LeetCode Go 80+ 题  
- [ ] 热题 100 完成 60+  

**简历**
- [ ] 一页 PDF，竞赛成绩 + 项目 + 技能栈  
- [ ] 技能栈诚实，不写「精通」  

---

## 附录 A：Windows 开发环境快速命令

```powershell
# 安装 Go 后验证
go version

# 创建项目
mkdir F:\study\code\projects\shorturl
cd F:\study\code\projects\shorturl
go mod init github.com/你的用户名/shorturl

# 安装常用依赖
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/mysql
go get github.com/redis/go-redis/v9
go get github.com/golang-jwt/jwt/v5
go get go.uber.org/zap
go get golang.org/x/crypto/bcrypt
```

```powershell
# WSL 安装（管理员 PowerShell）
wsl --install -d Ubuntu-22.04
```

---

## 附录 B：简历技能栈写法（参考）

```
语言：Go（主）、C++（竞赛）
基础：数据结构与算法（CCPC 省金 / ICPC 省银）、计算机网络、操作系统
后端：Gin、GORM、RESTful、JWT
存储：MySQL、Redis
工具：Git、Linux、Docker、Postman
```

---

## 附录 C：常见坑提醒

1. **教程收藏夹吃灰**：只选 1 套 Go 教程 + 1 本书，做完项目再换  
2. **项目贪大**：短链 MVP 先跑通，再加功能  
3. **算法过猛**：暑假不是冲 ACM，热题 + 中等足够  
4. **不写 daily-log**：三天后会忘进度，务必记录  
5. **Windows 写完不部署 Linux**：面试说会 Linux 却只会 Windows 会露馅  
6. **简历造假**：压测数据、命中率要真实或合理，能解释  
7. **忽视八股**：Go 岗一样考 MySQL/Redis/计网，暑假第 7 周要开始背  

---

## 附录 D：与本仓库现有资料的交叉引用

> 你的 `F:\study` 里已有大量后端笔记，Go 计划**不重复造轮子**，以下模块可直接对照阅读。

| 本计划模块 | 仓库内已有资料 |
|------------|----------------|
| 数据结构 / 算法 | [数据结构 00](后端学习/数据结构/00-学习路线图与说明.md) |
| Linux | [Linux 00](后端学习/Linux/00-学习路线图与说明.md) · [Docker](后端学习/Linux/12-Docker容器基础.md) |
| 计网 / HTTP | [计算机网络](前端学习/计算机网络/) |
| MySQL / SQL | [Java/06 MySQL](后端学习/Java/06-MySQL基础索引与事务.md) |
| Redis | [Java/07 Redis](后端学习/Java/07-Redis核心原理与缓存实战.md) |
| 短链项目设计 | [系统设计/08 短链服务设计](后端学习/系统设计/08-短链服务设计.md) ⭐ |
| 系统设计方法论 | [系统设计 01](后端学习/系统设计/01-系统设计方法论与面试框架.md) |
| 高并发 / 分布式 | [Java/12 高并发](后端学习/Java/12-高并发与分布式系统基础.md) |
| 后端总览 | [后端路线总览](后端学习/00-后端路线总览.md) |

**说明**：Java 目录里的 MySQL/Redis/并发笔记，语言无关，Go 面试同样考。

---

*文档版本：v1.1 · 生成日期：2026-07-08 · 路径：`F:\study\go-backend-learning-plan.md`*
