# GORM 与 MySQL 实战

<!-- 修改说明: 2026-07-08 按 EXPANSION-STANDARD 新建 §0、FAQ≥10、闭卷自测、费曼检验；理论交叉引用 Java/06；2026-07-14 补充完整短链模型、版本迁移、并发唯一约束与 keyset 分页；2026-07-26 按审查报告修订：修复 import/Config 缺失、Updates 零值语义、链式复用表述等 9 处问题；新增 First/Find 差异、链式复用安全、gorm.Config 选项、Preload 实操、批量写入与 Upsert、Raw SQL、Hooks、优雅关闭、Repository 集成测试、软删除×唯一索引等小节；所有核心示例补成完整可编译清单（已在 Go 1.26 + GORM v1.31 实际编译验证），命令行统一 Windows PowerShell；2026-07-27 去水化精简：删除知识地图、学习时长、学完你能做什么、学完标准、闭卷自测、费曼检验、章节衔接等模板板块；FAQ 拆解——sqlx 选型并入 §8.3、事务禁跨 goroutine 共用 tx 并入 §6、金额 decimal 并入 §2.2、ctx 承载 trace 并入 §1.7，其余条目正文已覆盖故删；「本章与上一章的关系」的分层说明与图并入 §4 开头；0.6/0.7 重编为 0.2/0.3，练习建议重编为 §13；正文讲解与全部代码清单原样保留 -->

> **文件编码**：UTF-8。  
> **定位**：Go 后端「持久化层」——GORM ORM 接 MySQL，完成 Model、CRUD、事务、迁移。  
> **理论前置**：[Java 06 MySQL 基础索引与事务](../Java/06-MySQL基础索引与事务.md)（索引、ACID、EXPLAIN 在本章不重复展开，以交叉引用为主）。  
> **代码前置**：[06 Gin 框架核心与中间件](./06-Gin框架核心与中间件.md)。  
> **环境约定**：命令默认在 **Windows 11 + PowerShell** 中执行；标注「服务器上」的命令属于 Linux 部署环境（13 章）。

---

## 0. 导读与环境准备

### 0.1 用一句话弄懂本章

**一句话**：**GORM = Go 的 MyBatis-Plus**——用 struct 映射表行，用链式 API 写 CRUD，Service 不再碰裸 SQL（复杂查询仍可 Raw，见 §8.3）。

**生活类比**：

| GORM 概念 | MyBatis / Java 对照 | 含义 |
|-----------|---------------------|------|
| `model.User` | Entity / PO | 一行记录的形状 |
| `db.Create(&u)` | `insert` | 插入 |
| `db.First(&u, id)` | `selectById` | 按主键查 |
| `db.Transaction` | `@Transactional` | 多步绑一起 |
| `AutoMigrate` | Flyway 简化版 | 自动建表/加列 |

**为什么重要**：06 章内存 map 无法持久化；GORM 是 Go 实习项目标配 ORM。

---

### 0.2 Docker 起 MySQL 手把手

前提：已安装并启动 **Docker Desktop**（Windows 上装它即可，本章只用它跑 MySQL；部署篇详见 [13 章](./13-Docker与Linux部署Go服务.md)）。以下命令都在 PowerShell 中执行。

先确认 3306 端口空闲（没有输出就是空闲）：

```powershell
Get-NetTCPConnection -LocalPort 3306 -ErrorAction SilentlyContinue
```

> Linux 服务器上等价命令是 `ss -lntp | grep 3306`。

| 步骤 | 命令/动作 | 预期 | 若不对 |
|------|-----------|------|--------|
| 1 | `docker run -d --name study-mysql -e MYSQL_ROOT_PASSWORD=root123 -e MYSQL_DATABASE=shortlink -p 3306:3306 mysql:8.0` | `docker ps` 显示 Up | 3306 被占改 `-p 3307:3306`（后文 DSN 端口同步改） |
| 2 | `docker exec -it study-mysql mysql -uroot -proot123 -e "SHOW DATABASES;"` | 列表里有 shortlink | 检查密码与容器名一致 |
| 3 | 按 §1.1 装依赖、§1.2 跑第一个程序 | 输出 `id=1 content=hello gorm ...` | 对照 §1.2 的 DSN 格式排查 |

说明：

- 第一次 `docker run` 会先拉取 `mysql:8.0` 镜像，可能要几分钟；拉取慢可配置国内镜像加速（13 章有讲）。
- 容器刚起来的十几秒内 MySQL 还在初始化，连接被拒绝属正常，稍等重试。
- 以后每次开机只需 `docker start study-mysql`，数据保留在容器里。

---

### 0.3 本章代码组织与「完整清单 / 片段」约定

本章继续在 **06 章创建的项目**里写代码（模块名 `github.com/you/shortlink-api`）。学完本章后目录长这样：

```text
shortlink-api/
├── go.mod
├── quickstart/
│   └── main.go            # §1.2 第一个 GORM 程序（一次性练习，可留可删）
├── cmd/
│   └── server/
│       └── main.go        # §10 Gin 集成 + 优雅关闭
├── internal/
│   ├── apperr/
│   │   └── apperr.go      # §4.1 领域哨兵错误
│   ├── model/
│   │   ├── user.go        # §2.1 User / UserProfile
│   │   └── short_link.go  # §2.3 ShortLink（含 §9 的 Hook）
│   ├── repository/
│   │   ├── db.go          # §1.3 连接 + AutoMigrate
│   │   ├── user_repo.go   # §4.2 / §4.4
│   │   ├── user_repo_test.go # §11 集成测试
│   │   └── link_repo.go   # §4.5 / §4.6
│   └── service/
│       └── link_service.go # §4.6 短码生成与重试
└── migrations/            # §3 版本化迁移（部署用）
```

两种代码块的约定：

- **完整清单**：给出整个文件内容（package + import + 全部代码），标注文件路径，可以直接新建文件粘贴，保证能编译。
- **片段**：节选，会注明「属于哪个文件、完整版在哪一节」；片段不含 import，不要单独粘贴运行。

本章全部完整清单已在 Go 1.26 + `gorm.io/gorm v1.31` 下实际编译通过。

---

## 1. 连接与配置

### 1.1 准备项目与安装依赖（PowerShell 手把手）

在 06 章的项目根目录（`go.mod` 所在目录）执行：

```powershell
# 可选：本会话内使用国内代理加速拉包（只影响当前窗口，不改全局配置）
$env:GOPROXY = "https://goproxy.cn,direct"

go get gorm.io/gorm
go get gorm.io/driver/mysql
```

版本说明：

- 我们用的是 **GORM v2**。注意一个容易懵的点：GORM v2 的模块版本号仍叫 `v1.x`（如本章验证用的 `gorm.io/gorm v1.31.2`、`gorm.io/driver/mysql v1.6.0`）——“v2” 指的是 `gorm.io/gorm` 这个新模块路径，老的 v1 在 `github.com/jinzhu/gorm`，**不要装错**。
- `go get` 后 `go.mod` 里会多出这两行依赖即成功；只要是 `gorm.io/gorm v1.25` 以上，本章代码全部适用。

---

### 1.2 第一个能跑通的 GORM 程序（完整清单）

先用最短的代码把「连接 → 建表 → 插入 → 查询」整条链路打通，建立手感。

**文件：`quickstart/main.go`（完整清单）**

```go
package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Note 是本章第一个演示用模型：GORM 会把它映射到表 notes
type Note struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	Content   string `gorm:"size:255;not null"`
	CreatedAt time.Time
}

func main() {
	dsn := "root:root123@tcp(127.0.0.1:3306)/shortlink?charset=utf8mb4&parseTime=True&loc=UTC"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect mysql: %v", err)
	}
	log.Println("database connected")

	if err := db.AutoMigrate(&Note{}); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}
	if err := db.Create(&Note{Content: "hello gorm"}).Error; err != nil {
		log.Fatalf("insert note: %v", err)
	}
	var latest Note
	if err := db.Order("id DESC").First(&latest).Error; err != nil {
		log.Fatalf("query note: %v", err)
	}
	fmt.Printf("id=%d content=%s created_at=%s\n",
		latest.ID, latest.Content, latest.CreatedAt.Format(time.RFC3339))
}
```

运行：

```powershell
go run ./quickstart
```

预期输出（时间会不同）：

```text
2026/07/26 12:00:00 database connected
id=1 content=hello gorm created_at=2026-07-26T04:00:00Z
```

> `created_at` 以 `Z` 结尾表示 UTC，比北京时间“早 8 小时”是**正常的**：我们统一存 UTC，展示层再转时区。别急着去“修”它，§12 错误表里有完整说明。

**逐段讲解**：

1. **DSN（Data Source Name）**是连接字符串，解剖如下：

   ```text
   root : root123 @tcp( 127.0.0.1:3306 ) / shortlink ? 参数们
   用户    密码      协议(地址:端口)       库名        连接选项
   ```

   | 参数 | 作用 |
   |------|------|
   | `charset=utf8mb4` | 完整 Unicode（含 emoji），MySQL 的 `utf8` 是残缺三字节版 |
   | `parseTime=True` | 把 `DATETIME` 列解析成 Go 的 `time.Time`（不加则只能拿 `[]byte`） |
   | `loc=UTC` | 驱动解析时间时使用的时区 |

2. **`gorm.Open`** 建立连接（底层是连接池，见 §1.4）；第二个参数 `&gorm.Config{}` 是 GORM 全局配置，§1.3/§1.6 会往里加东西。
3. **`AutoMigrate(&Note{})`** 按 struct 定义建表：表名自动取结构体名的**蛇形复数**（`Note`→`notes`），字段名转蛇形（`CreatedAt`→`created_at`）。
4. **`Create`** 插入后会把自增主键**回填**到 `ID` 字段；`.Error` 是 GORM 取错误的方式——**每次数据库操作后都必须检查**。
5. **`Order("id DESC").First(&latest)`** 生成 `SELECT ... ORDER BY id DESC LIMIT 1`。

验证表真的建出来了：

```powershell
docker exec -it study-mysql mysql -uroot -proot123 -e "USE shortlink; SHOW TABLES; SELECT * FROM notes;"
```

---

### 1.3 项目版连接代码（完整清单）

quickstart 的 DSN 是写死的，项目版需要：配置注入、连接池参数、启动时 Ping 自检、SQL 日志。

**文件：`internal/repository/db.go`（完整清单）**

```go
package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/you/shortlink-api/internal/model"
)

// Config 汇总连接 MySQL 需要的五个参数；12 章会改成从配置文件/环境变量加载。
type Config struct {
	User     string // 数据库用户，如 root
	Password string // 密码
	Host     string // 地址，本机 Docker 就是 127.0.0.1
	Port     int    // 端口，默认 3306
	DBName   string // 库名，如 shortlink
}

func NewDB(cfg Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=UTC&time_zone=%%27%%2B00%%3A00%%27",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Info), // 开发期打印每条 SQL
		TranslateError: true, // 将重复键等驱动错误翻译为 gorm.ErrDuplicatedKey（§4 依赖它）
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}

// AutoMigrate 建表/加列；仅本地开发用，部署走 §3 的版本化迁移。
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&model.User{}, &model.UserProfile{}, &model.ShortLink{})
}
```

> 此文件 import 了 `internal/model`，三个 Model 在 §2 给出完整清单；先把本文件抄下来，§2 抄完即可一起编译。

**逐点讲解**：

- **DSN 里的 `%%27%%2B00%%3A00%%27` 是什么鬼**：这是两层转义。`fmt.Sprintf` 里 `%` 要写成 `%%`，所以真实字符串是 `time_zone=%27%2B00%3A00%27`；再做 URL 解码，`%27` 是 `'`、`%2B` 是 `+`，最终 MySQL 收到的是 `time_zone='+00:00'`——把**会话时区**设为 UTC。`loc=UTC` 管驱动怎么**解析**时间，`time_zone` 管 MySQL 侧 `NOW()` 等函数怎么**生成**时间，两者统一成 UTC 才不会差 8 小时。
- **`TranslateError: true`**：MySQL 唯一键冲突原始错误是 `Error 1062`，开了这个开关 GORM 会把它翻译成跨数据库统一的 `gorm.ErrDuplicatedKey`，§4 的错误判断全靠它。
- **`db.DB()`** 取出底层 `*sql.DB` 设置连接池（§1.4 详解），并用带 3 秒超时的 `PingContext` 做启动自检——连不上就立刻失败，而不是等第一个请求才炸。
- **`AutoMigrate`** 放在这里统一维护，§3 讨论它的适用边界。

---

### 1.4 `*gorm.DB` 与 `*sql.DB` 分工

- `*gorm.DB` 是查询构造、模型映射和回调层；其中 `WithContext`、`Session`、`Debug` 这三个 **New Session 方法**会返回可安全复用的新会话句柄（注意：`Where` 等链式方法**不是**，区别见 §1.5）。
- 底层 `*sql.DB` **不是一条连接**，而是并发安全的连接池。
- 一次查询会从池中借连接，使用完归还；事务会在其生命周期内占用同一条连接。

连接池参数不能照抄固定数字：

| 参数 | 作用 | 设错的现象 |
|------|------|------------|
| `MaxOpenConns` | 最大打开连接数，0 表示不限制 | 太小请求排队；太大打爆 MySQL |
| `MaxIdleConns` | 最大空闲连接数 | 太小频繁建连；太大闲置资源 |
| `ConnMaxLifetime` | 连接最大总寿命 | 应小于数据库/代理强制断开周期并加抖动 |
| `ConnMaxIdleTime` | 空闲多久后关闭 | 控制低峰期闲置连接 |

运行中观察 `sqlDB.Stats()`：`OpenConnections`、`InUse`、`Idle`、`WaitCount`、`WaitDuration`。若 `WaitCount` 和等待时间持续上升，说明池成为瓶颈；先确认慢 SQL 和事务时长，再决定是否加连接。

---

### 1.5 链式调用复用安全：哪些 `*gorm.DB` 能存、哪些不能

GORM 的方法分三类，这是理解「什么时候会踩坑」的钥匙：

| 类别 | 例子 | 行为 |
|------|------|------|
| **Chain 方法** | `Where` / `Order` / `Limit` / `Model` / `Select` | 只往当前语句（Statement）上**追加**条件，不执行 SQL |
| **Finisher 方法** | `Find` / `First` / `Count` / `Create` / `Updates` / `Delete` | 真正生成并执行 SQL |
| **New Session 方法** | `Session` / `WithContext` / `Debug` | 返回一个“干净起点”，之后的第一次 Chain 调用会克隆出新语句 |

关键规则：**从新 session 出发的第一次 Chain 调用会克隆语句；同一条链上后续的 Chain 调用都在这条语句上原地追加**。于是把中间结果存进变量反复用，条件就会越积越多：

```go
// ❌ 反例：语句污染（片段，仅演示）
q := db.Where("status = ?", 1)
q.Where("user_id = ?", 1).Find(&a) // 实际条件: status=1 AND user_id=1
q.Where("user_id = ?", 2).Find(&b) // 污染！条件累积成 status=1 AND user_id=1 AND user_id=2
```

第二次查询把第一次的 `user_id=1` 也带上了，结果必然为空——这类 bug 没有报错，只有“查不到”，非常难排查。安全写法：

```go
// ✅ 写法一：每次从 db 重新起链（最常用）
db.Where("status = ?", 1).Where("user_id = ?", 2).Find(&b)

// ✅ 写法二：确要复用公共条件时，用 Session 定格成安全起点
base := db.Where("status = ?", 1).Session(&gorm.Session{})
base.Where("user_id = ?", 1).Find(&a) // 各自从 base 克隆，互不污染
base.Where("user_id = ?", 2).Find(&b)
```

**那 §4.4 分页里 `q := db.WithContext(ctx).Model(...)` 先 `Count` 再 `Find` 为什么安全？** 因为它是「一次性构建、当场连用两个 Finisher、用完即弃」：`Count` 执行完会清掉自己临时加的 `SELECT count(*)`，`Find` 是 q 的最后一次使用，之后 q 被丢弃。危险的是把 Chain 产物存成**包级变量/结构体字段**跨请求复用。

**口诀**：可以长期保存的只有两种——`gorm.Open` 返回的 `db`，以及 `Session`/`WithContext` 的返回值；`Where` 之后的一律用完即弃。

---

### 1.6 gorm.Config 实用选项：性能与慢查询日志

除了 §1.3 用到的两项，这三个选项面试和调优常见：

```go
// 片段 · 可替换 internal/repository/db.go 中的 &gorm.Config{...}
&gorm.Config{
	SkipDefaultTransaction: true, // 跳过单条写操作的默认事务
	PrepareStmt:            true, // 缓存预编译语句
	TranslateError:         true,
	Logger:                 newLogger, // 见下方定义
}
```

- **`SkipDefaultTransaction`**：GORM 默认把每个写操作（Create/Updates/Delete）包在一个小事务里执行，为的是保证「Hook + 写入」整体原子。若你的写操作没有复杂 Hook，跳过它可提速（官方给的参考数字约 30%，以自己压测为准）。注意：跳过后单条语句本身仍是原子的（InnoDB 特性）；**多条语句需要原子时永远显式 `db.Transaction`（§6）**。
- **`PrepareStmt`**：把 SQL 预编译并缓存，同一形状的查询反复执行时省去解析开销。
- **慢查询日志**：`logger.Default` 不能配阈值，用 `logger.New` 自定义：

```go
// 片段 · 属于 internal/repository/db.go（需 import "log"、"os"）
newLogger := logger.New(
	log.New(os.Stdout, "\r\n", log.LstdFlags),
	logger.Config{
		SlowThreshold:             200 * time.Millisecond, // 超过即打 SLOW SQL 警告
		LogLevel:                  logger.Warn,            // 生产建议 Warn；开发可 Info 看全量 SQL
		IgnoreRecordNotFoundError: true,                   // 查不到不算错误日志
	},
)
```

---

### 1.7 context 超时实测：查询到底会不会被取消

`WithContext(ctx)` 不是仪式，它真的能掐断慢查询。把下面片段塞进 `quickstart/main.go` 的 main 末尾试一次（`Raw` 的完整讲解见 §8.3，这里只借用它执行一条慢 SQL）：

```go
// 片段 · 演示用（需 import "context"）
ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
defer cancel()
var one int
err := db.WithContext(ctx).Raw("SELECT SLEEP(2)").Scan(&one).Error
fmt.Println(err) // context deadline exceeded
```

`SELECT SLEEP(2)` 要睡 2 秒，但 500ms 时 context 到期，调用立刻返回 `context.DeadlineExceeded`（用 `errors.Is(err, context.DeadlineExceeded)` 判断）。此时：

- 这条被取消的**连接会被驱动废弃**，连接池随后补充新连接——不用担心“脏连接”被复用；
- 若发生在事务中，整个事务回滚；
- 这就是每个 Repository 方法都写 `WithContext(ctx)` 的原因：上游（Gin 请求、调用方超时）一取消，数据库层立刻止损，而不是白白算完。

超时/取消之外，`ctx` 还是链路追踪（trace）信息的载体：中间件生成的 trace id 靠它一路传到 SQL 层，接入 tracing 插件后才能把慢查询挂到具体请求上——这是 `WithContext(ctx)` 必写的另一个理由。

---

## 2. Model 定义

### 2.1 User 与 UserProfile（完整清单）

**文件：`internal/model/user.go`（完整清单）**

```go
package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string         `gorm:"size:32;uniqueIndex;not null" json:"username"`
	Password  string         `gorm:"size:128;not null" json:"-"` // 09 章 bcrypt，JSON 不返回
	Email     string         `gorm:"size:128" json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 软删除

	// has many 关联：一个用户拥有多条短链（§7 Preload 用）。
	// 它不会在 users 表里生成列，外键在 short_links.user_id 上。
	Links []ShortLink `gorm:"foreignKey:UserID" json:"-"`
}

// UserProfile 与 User 一对一（§6 事务示例用）
type UserProfile struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;uniqueIndex" json:"user_id"` // 唯一索引保证 1:1
	Bio       string    `gorm:"size:255;not null;default:''" json:"bio"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

**先记三条命名/约定规则**（GORM 的“约定优于配置”）：

1. 表名 = 结构体名的蛇形复数：`User`→`users`、`ShortLink`→`short_links`；想改就给类型实现 `func (User) TableName() string`。
2. 列名 = 字段名转蛇形：`CreatedAt`→`created_at`。
3. 叫 `CreatedAt` / `UpdatedAt` 的字段由 GORM **自动填充**创建/更新时间；类型为 `gorm.DeletedAt` 的字段自动启用软删除（§5.2）。

---

### 2.2 常用 gorm tag 速查

| tag 写法 | 含义 |
|----------|------|
| `primaryKey` | 主键 |
| `autoIncrement` | 自增 |
| `size:32` | `varchar(32)` |
| `type:text`、`type:varchar(16) ...` | 直接指定完整列类型（复杂类型这么写） |
| `not null` | `NOT NULL` |
| `default:0` | 列默认值 |
| `uniqueIndex` | 唯一索引；`uniqueIndex:uk_xxx` 指定索引名 |
| `index` | 普通索引 |
| `index:idx_xxx,priority:2` | 参与名为 idx_xxx 的**联合索引**，按 priority 从小到大排列字段 |
| `column:xx` | 自定义列名 |
| `-` | 该字段不映射到表 |
| `foreignKey:UserID` | 关联的外键字段（§7） |

两套标签别混淆：`gorm:"..."` 管数据库映射，`json:"..."` 管 `encoding/json` 序列化（06 章讲过），互不影响。`Password` 上的 `json:"-"` 只是不出现在 API 响应里，数据库照存。

**列类型选择的一个高频雷区——金额**：不要用 float/double（二进制浮点表示不了大多数十进制小数，累加会出分位误差），列用 `type:decimal(10,2)`，Go 侧用 `shopspring/decimal` 或字符串承载；原理展开见 [Java 06](../Java/06-MySQL基础索引与事务.md) 金额章节。

---

### 2.3 ShortLink（完整清单）——10 章直接复用

**文件：`internal/model/short_link.go`（完整清单）**

```go
package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type LinkStatus uint8

const (
	LinkStatusActive   LinkStatus = 1
	LinkStatusDisabled LinkStatus = 2
)

type ShortLink struct {
	ID          int64          `gorm:"primaryKey;autoIncrement;index:idx_links_user_page,priority:4" json:"id"`
	UserID      int64          `gorm:"not null;index:idx_links_user_page,priority:1" json:"user_id"`
	ShortCode   string         `gorm:"type:varchar(16) CHARACTER SET ascii COLLATE ascii_bin;not null;uniqueIndex:uk_short_links_code" json:"short_code"`
	OriginalURL string         `gorm:"type:text;not null" json:"original_url"`
	Title       string         `gorm:"size:128;not null;default:''" json:"title"`
	Status      LinkStatus     `gorm:"type:tinyint unsigned;not null;default:1" json:"status"`
	ExpiresAt   *time.Time     `gorm:"index:idx_links_expires_at" json:"expires_at,omitempty"`
	ClickCount  int64          `gorm:"not null;default:0" json:"click_count"`
	Version     uint64         `gorm:"not null;default:1" json:"version"`
	CreatedAt   time.Time      `gorm:"not null;index:idx_links_user_page,priority:3" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index:idx_links_user_page,priority:2" json:"-"`
}

// BeforeCreate 在 INSERT 前被 GORM 自动调用（与 INSERT 同处默认事务中）。
// 先照抄，它是 §9 Hooks 的主角，讲解见 §9。
func (l *ShortLink) BeforeCreate(tx *gorm.DB) error {
	if l.OriginalURL == "" {
		return fmt.Errorf("original url must not be empty")
	}
	if l.Status == 0 {
		l.Status = LinkStatusActive
	}
	if l.Title == "" {
		l.Title = "未命名链接"
	}
	return nil
}
```

**字段为何这样设计**：

| 字段/索引 | 项目作用 |
|-----------|----------|
| `short_code` + `uk_short_links_code` | 数据库作为最终防线，杜绝两个长链指向同一短码 |
| `ascii_bin` | Base62 大小写敏感；避免 MySQL 默认排序规则把 `a` 和 `A` 当相同 |
| `status` / `expires_at` | 支持禁用与过期，跳转前必须校验 |
| `version` | 乐观锁与缓存版本控制，防止并发覆盖 |
| `(user_id, deleted_at, created_at, id)` | 支持“我的短链”稳定倒序 keyset 分页（§4.5） |
| `click_count` | 最终汇总值；高频点击不在每次跳转中直接同步更新 MySQL |

注意 `idx_links_user_page` 这个联合索引是靠四个字段上的同名 `index` tag + `priority` 拼出来的：priority 1→4 依次是 `user_id, deleted_at, created_at, id`。

软删除后唯一短码仍被占用，这是刻意的安全策略：旧书签不应在未来突然跳到另一个用户的新资源。如果业务要复用短码，必须单独设计回收期和审计流程，不能仅把 `deleted_at` 塞进唯一索引。短链表的容量估算与整体设计取舍见 [系统设计 08 短链服务设计](../系统设计/08-短链服务设计.md)。

---

### 2.4 软删除 × 唯一索引：users 表的同名重注册问题（经典面试题）

上面说 `short_code` 软删后**故意**不释放；但 `users.username` 恰恰相反——用户注销（软删）后，业务通常希望这个用户名**可以被重新注册**。而现在的定义做不到：软删只是给 `deleted_at` 填了时间，行还在，`username` 的唯一索引仍然拦住同名新用户。

**为什么不能简单把唯一索引改成 `(username, deleted_at)`？** 因为 `gorm.DeletedAt` 未删除时是 `NULL`，而 MySQL 唯一索引把 NULL 彼此视为**不同**——两个活跃用户 `(alice, NULL)` 和 `(alice, NULL)` 都能插进去，唯一性直接失效。

两种正确做法：

**方案 A：注销时改写用户名**（简单，无新依赖）——软删的同时把 `username` 改成不可能冲突的值，如 `alice#deleted#1024`（拼上主键 ID）；在同一个事务里完成（§6）。

**方案 B：把软删标记从 NULL 换成 0**——用官方插件 `gorm.io/plugin/soft_delete` 的时间戳标记：未删除 = `0`（不是 NULL），删除 = Unix 时间戳。活跃行都是 `(alice, 0)`，唯一索引能拦住同名活跃用户；删除后变成 `(alice, 1721980800)`，不再挡路。

```powershell
go get gorm.io/plugin/soft_delete
```

```go
// 片段 · 可选方案演示（替换 User 的相应字段时使用）
import "gorm.io/plugin/soft_delete"

type User struct {
	ID        int64                 `gorm:"primaryKey;autoIncrement"`
	Username  string                `gorm:"size:32;not null;uniqueIndex:uk_users_username,priority:1"`
	DeletedAt soft_delete.DeletedAt `gorm:"uniqueIndex:uk_users_username,priority:2"` // 未删除=0
	// ...其余字段同 §2.1
}
```

本章正文继续用 `gorm.DeletedAt`（与 GORM 默认生态最兼容）；面试时能讲出「NULL 在唯一索引里不判重 → 所以要么改名要么用 0 标记」这条推理链，就是加分项。

---

## 3. 迁移：开发 AutoMigrate，项目交付用版本化 SQL

本地开发用 §1.3 里定义的 `repository.AutoMigrate`（内含 User / UserProfile / ShortLink 三个模型）即可。它与生产级迁移工具的分工：

| 能力 | AutoMigrate | 版本化迁移工具 |
|------|-------------|-------------|
| 建表/加列 | ✅ | ✅ |
| 删列/改类型 | ❌ 不安全 | ✅ 版本脚本 |
| 本地第一遍练习 | 够用 | 可选 |
| 简历项目/部署环境 | 不作为发布方案 | 必须，结构变更可审计、可重复 |

推荐使用 `golang-migrate`，迁移文件一旦发布就不再修改。Windows 上安装 CLI：

```powershell
go install -tags mysql github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate -version   # 有版本号输出即成功（migrate.exe 装在 $env:USERPROFILE\go\bin）
```

> `-tags mysql` 表示编译进 MySQL 驱动，缺了它运行时会报 unknown driver。

目录结构：

```text
migrations/
├── 000001_create_users.up.sql
├── 000001_create_users.down.sql
├── 000002_create_short_links.up.sql
└── 000002_create_short_links.down.sql
```

`000002_create_short_links.up.sql` 的核心结构应与 Model 对齐：

```sql
CREATE TABLE short_links (
    id           BIGINT NOT NULL AUTO_INCREMENT,
    user_id      BIGINT NOT NULL,
    short_code   VARCHAR(16) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
    original_url TEXT NOT NULL,
    title        VARCHAR(128) NOT NULL DEFAULT '',
    status       TINYINT UNSIGNED NOT NULL DEFAULT 1,
    expires_at   DATETIME(3) NULL,
    click_count  BIGINT NOT NULL DEFAULT 0,
    version      BIGINT UNSIGNED NOT NULL DEFAULT 1,
    created_at   DATETIME(3) NOT NULL,
    updated_at   DATETIME(3) NOT NULL,
    deleted_at   DATETIME(3) NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uk_short_links_code (short_code),
    KEY idx_links_user_page (user_id, deleted_at, created_at, id),
    KEY idx_links_expires_at (expires_at),
    CONSTRAINT fk_short_links_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

> `DATETIME(3)` 的 3 是毫秒精度，与 GORM 默认生成的时间列精度一致。

本机执行迁移（PowerShell）：

```powershell
$env:MIGRATE_DATABASE_URL = "mysql://root:root123@tcp(127.0.0.1:3306)/shortlink"
migrate -path ./migrations -database $env:MIGRATE_DATABASE_URL up
migrate -path ./migrations -database $env:MIGRATE_DATABASE_URL version
```

服务器上（Linux/CI）等价命令：

```bash
migrate -path ./migrations -database "$MIGRATE_DATABASE_URL" up
migrate -path ./migrations -database "$MIGRATE_DATABASE_URL" version
```

注意 `MIGRATE_DATABASE_URL` 用的是迁移工具的 URL 格式（带 `mysql://` 前缀），**与 GORM 的 `user:pass@tcp(...)` DSN 不是同一种格式**，不要互相照抄。迁移应在发布步骤中单独执行，应用启动只检查 schema 版本，不让每个副本同时 `AutoMigrate`。`down.sql` 只用于本地/测试回滚；生产含数据回滚通常要写前向修复迁移，并先备份和演练。

---

## 4. CRUD 与 Repository

06 章 Handler 调 Service，数据在 `map` 里。07 章在 Service 与 MySQL 之间加 **Repository 层**（可选但推荐），GORM 负责 SQL 生成与映射。

```mermaid
flowchart LR
    H[Handler Gin] --> S[Service]
    S --> R[Repository GORM]
    R --> M[(MySQL)]
    J06[Java/06 索引事务理论] -.-> M
```

**理论分工**：B+ 树、最左前缀、ACID、隔离级别等原理 → [Java 06](../Java/06-MySQL基础索引与事务.md)；本章专注 **Go 代码怎么写对**。

### 4.1 错误先行：apperr 哨兵错误（完整清单）

Repository 会遇到“查不到”“重复了”这类**业务可预期**的失败，不能把 GORM 的错误直接漏到 Handler。做法是定义一组**哨兵错误**（sentinel error，即预定义的可比较错误值，02 章 errors.Is 的用武之地）：

**文件：`internal/apperr/apperr.go`（完整清单）**

```go
package apperr

import "errors"

// 领域哨兵错误：Repository/Service 返回它们，Handler 用 errors.Is 判断后映射 HTTP 状态码。
var (
	ErrNotFound    = errors.New("resource not found")              // -> 404
	ErrConflict    = errors.New("resource conflict")               // -> 409
	ErrUnavailable = errors.New("service temporarily unavailable") // -> 503
)
```

配合 `fmt.Errorf("...: %w", apperr.ErrNotFound)` 包装：既保留了“哪个操作、什么参数”的上下文，`errors.Is` 又依然能命中哨兵。Handler 层的映射见 §10。

---

### 4.2 UserRepository（完整清单）

**文件：`internal/repository/user_repo.go`（完整清单）**

```go
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/you/shortlink-api/internal/apperr"
	"github.com/you/shortlink-api/internal/model"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("username already exists: %w", apperr.ErrConflict)
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user id=%d: %w", id, apperr.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get user id=%d: %w", id, err)
	}
	return &u, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, name string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Where("username = ?", name).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user username=%q: %w", name, apperr.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get user username=%q: %w", name, err)
	}
	return &u, nil
}

// List 见 §4.4 讲解
func (r *UserRepository) List(ctx context.Context, page, size int) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	var users []model.User
	var total int64
	q := r.db.WithContext(ctx).Model(&model.User{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	if err := q.Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	return users, total, nil
}
```

**讲解**：

- `First(&u, id)` 的第二个参数是**主键**内联条件，生成 `WHERE id = ? LIMIT 1`；按其他列查就用 `Where("username = ?", name).First(&u)`。
- `Create` 里的 `gorm.ErrDuplicatedKey` 依赖 §1.3 的 `TranslateError: true`——没开这个开关就永远匹配不上。
- **为什么“查不到”返回 `apperr.ErrNotFound` 而不是 `(nil, nil)`**：`(nil, nil)` 模式要求每个调用方都记得先判 `if u == nil`，忘一次就是 nil 指针解引用 panic；Go 社区（如 Uber 风格指南）普遍建议用显式错误表达“不存在”。这也和 §5.4 的错误风格保持一致：全链路 `errors.Is(err, apperr.ErrNotFound)` 一种写法走到底。

---

### 4.3 First / Take / Find：查不到时行为完全不同（高频坑）

| 方法 | 排序 | 查不到一行时 |
|------|------|--------------|
| `First(&u)` | 主键升序取第一条 | 返回 `gorm.ErrRecordNotFound` |
| `Take(&u)` | 无排序取一条 | 返回 `gorm.ErrRecordNotFound` |
| `Last(&u)` | 主键降序取第一条 | 返回 `gorm.ErrRecordNotFound` |
| `Find(&users)` | 无 | **`err == nil`**！只是切片为空 |

```go
// 片段 · 演示（可放进 quickstart 测试）
var u model.User
err := db.First(&u, 999).Error
fmt.Println(err) // record not found

var users []model.User
res := db.Where("id > ?", 999_999).Find(&users)
fmt.Println(res.Error, res.RowsAffected, len(users)) // <nil> 0 0
```

`Find` 的语义是“查一个集合”，空集合是合法结果，所以**不报错**——判断“没查到”要用 `len(users) == 0` 或 `res.RowsAffected == 0`，等 `ErrRecordNotFound` 是等不来的。反过来的坑同样常见：用 `Find(&u)` 传单个 struct 查询，查不到也不报错，你拿着一个全零值的 `u` 继续跑，直到某处逻辑莫名其妙。**查单条用 First/Take，查集合用 Find 并判空。**

---

### 4.4 Offset 分页

`List` 的实现在 §4.2 清单里，要点：

- **参数归一化**：`page < 1` 时 GORM 会收到负 Offset——它会**静默忽略**，等价于第一页，不报错；`size = 0` 会生成 `LIMIT 0` **静默返回空数组**；size 无上限则可能被恶意参数一次拉全表。三种都不报错，所以必须在入口处归一化（`page=1`、`size=20`、上限 100），这与 §4.5 keyset 版本的校验标准一致。
- **Count + Find 复用 `q`**：这是 §1.5 讲过的“一次性构建、当场用完即弃”特例，安全；但不要把 `q` 存起来跨请求用。

---

### 4.5 深分页：Offset 为什么越往后越慢，用 keyset 替代

`OFFSET 100000 LIMIT 20` 往往仍需扫描/跳过大量索引记录，而且并发插入会让翻页重复或漏项。短链列表按 `created_at DESC, id DESC` 排序，用复合游标（cursor）与 §2.3 的联合索引配合。

**文件：`internal/repository/link_repo.go`（完整清单，含 §4.6 的 Create）**

```go
package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/you/shortlink-api/internal/model"
)

var ErrShortCodeConflict = errors.New("short code conflict")

type LinkRepository struct {
	db *gorm.DB
}

func NewLinkRepository(db *gorm.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

// Create 直接插入，由唯一索引裁决短码冲突（§4.6 讲解）
func (r *LinkRepository) Create(ctx context.Context, link *model.ShortLink) error {
	err := r.db.WithContext(ctx).Create(link).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrShortCodeConflict
	}
	if err != nil {
		return fmt.Errorf("create short link: %w", err)
	}
	return nil
}

// LinkCursor 是复合游标：上一页最后一行的排序键
type LinkCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        int64     `json:"id"`
}

// EncodeCursor 把游标编码成不透明的 Base64URL 字符串返回给客户端
func EncodeCursor(c LinkCursor) string {
	b, _ := json.Marshal(c) // 结构固定，Marshal 不会失败
	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodeCursor 解析客户端原样带回的游标；解不开一律按坏请求处理
func DecodeCursor(s string) (*LinkCursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("bad cursor: %w", err)
	}
	var c LinkCursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("bad cursor: %w", err)
	}
	return &c, nil
}

func (r *LinkRepository) ListByUser(
	ctx context.Context,
	userID int64,
	cursor *LinkCursor,
	size int,
) ([]model.ShortLink, bool, error) {
	if size <= 0 || size > 100 {
		size = 20
	}
	var links []model.ShortLink
	q := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Limit(size + 1)
	if cursor != nil {
		q = q.Where(
			"(created_at < ? OR (created_at = ? AND id < ?))",
			cursor.CreatedAt, cursor.CreatedAt, cursor.ID,
		)
	}
	if err := q.Find(&links).Error; err != nil {
		return nil, false, fmt.Errorf("list links by user: %w", err)
	}
	hasMore := len(links) > size
	if hasMore {
		links = links[:size]
	}
	return links, hasMore, nil
}
```

**讲解**：

- **keyset 的思路**：不数“跳过多少行”，而是记住上一页最后一行的排序键 `(created_at, id)`，下一页直接 `WHERE 排序键 < 游标`——索引直达起点，翻到第几页都一样快，且并发插入不会造成重复/漏项。
- **`id` 是并列裁判**：同一毫秒可能创建多条，只按 `created_at` 会翻页错乱；补上唯一的 `id` 保证全序。
- **`Limit(size + 1)` 多查一条**：第 size+1 条存在就说明还有下一页（`hasMore`），砍掉后返回——省去一次 COUNT。
- **游标要不透明**：响应把 `EncodeCursor` 的 Base64URL 字符串发给客户端，下一页原样传回，`DecodeCursor` 解不开就按 400 处理。不要让客户端看懂/构造游标内容，更**不允许客户端传 `user_id`**——它必须来自 09 章鉴权 Context。

---

### 4.6 唯一约束才是并发下的最终裁判

“先 `SELECT short_code`，不存在再 `INSERT`”存在 TOCTOU 竞态（Time-of-check to time-of-use：检查和使用之间世界变了）：两个请求可能同时查到“不存在”，然后都去插入。正确做法是**直接尝试插入，由唯一索引裁决**，只对短码冲突做有限重试。Repository 侧的 `Create` 已在 §4.5 清单中给出，Service 侧：

**文件：`internal/service/link_service.go`（完整清单）**

```go
package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/you/shortlink-api/internal/apperr"
	"github.com/you/shortlink-api/internal/model"
	"github.com/you/shortlink-api/internal/repository"
)

const base62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// randomCode 用密码学随机源生成 n 位 Base62 短码。
// 取模有轻微分布偏差，对短码场景无影响；发号器方案见 10 章。
func randomCode(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read random: %w", err)
	}
	for i, b := range buf {
		buf[i] = base62[int(b)%len(base62)]
	}
	return string(buf), nil
}

type LinkService struct {
	repo *repository.LinkRepository
}

func NewLinkService(repo *repository.LinkRepository) *LinkService {
	return &LinkService{repo: repo}
}

func (s *LinkService) Create(ctx context.Context, userID int64, rawURL string) (*model.ShortLink, error) {
	for attempt := 0; attempt < 5; attempt++ {
		code, err := randomCode(7)
		if err != nil {
			return nil, err
		}
		link := &model.ShortLink{UserID: userID, ShortCode: code, OriginalURL: rawURL}
		err = s.repo.Create(ctx, link)
		if err == nil {
			return link, nil
		}
		if !errors.Is(err, repository.ErrShortCodeConflict) {
			return nil, err
		}
		// 短码撞了：换一个码重试，最多 5 次
	}
	return nil, fmt.Errorf("generate unique short code after retries: %w", apperr.ErrUnavailable)
}
```

**讲解**：

- `crypto/rand` 是密码学随机源（不可预测）；`math/rand` 可预测，用户能猜出别人的短码，不要用。
- 重试**只**针对 `ErrShortCodeConflict`：7 位 Base62 空间约 62⁷≈3.5 万亿，正常几乎不撞；连撞 5 次说明系统异常，返回 503（`ErrUnavailable`）而不是死循环。
- 若表里还有用户名等其他唯一索引，Repository 应根据具体约束名/错误上下文映射成不同领域错误，不能看到任何 1062 都当“短码碰撞”重试。

---

## 5. 更新与软删除

### 5.1 Updates 的三种形式：struct、map、Save 的零值语义（必考）

GORM 更新有三种写法，**对零值字段（`""`、`0`、`false`）的处理完全不同**，混淆它们是线上事故高发区：

```go
// 片段 · 演示（user 是已查出的 *model.User）

// ① struct 形式：只更新“非零值”字段
db.Model(user).Updates(model.User{Email: "new@x.com"}) // UPDATE users SET email='new@x.com'

// struct + 零值 = 什么都不更新！想清空 Email 这样写是无效的：
db.Model(user).Updates(model.User{Email: ""}) // 没有任何 SET 子句

// ② map 形式：map 里列出的字段“全部”写入，包括零值
db.Model(user).Updates(map[string]any{"email": ""}) // email 被真正清空

// ③ Select 指定列 + struct：强制把所选列写入（即使是零值）
db.Model(user).Select("email").Updates(model.User{Email: ""}) // 也能清空

// ④ Save：按主键把“整个对象的所有字段”写回去
db.Save(user) // UPDATE 全列；若 user 没有主键，退化为 INSERT

// ⑤ 单列更新
db.Model(user).Update("email", "single@x.com")
```

| 写法 | 零值字段 | 触发 Hook / 自动更新 updated_at |
|------|----------|----------------------------------|
| `Updates(struct)` | **跳过** | 是 |
| `Updates(map)` | **写入** | 是 |
| `Select("col").Updates(struct)` | 写入所选列 | 是 |
| `Save(&obj)` | 全字段写入 | 是；无主键时退化为 Create |
| `UpdateColumn / UpdateColumns` | 写入 | **否**（也不更新 updated_at，见 §6.3/§9） |

**怎么选**：更新字段清单明确时用 **map**（或 `Select`+struct），语义是“我说改哪列就改哪列”；`Updates(struct)` 只适合“非零字段全是要改的”场景，且要牢记它清不掉零值；`Save` 是整对象覆盖，容易把并发修改也覆盖掉（§5.4 的乐观锁就是为此服务），项目里慎用。面试经典问题“为什么我的 bool 字段改不成 false？”——答案就是 `Updates(struct)` 跳过了零值。

---

### 5.2 软删除与物理删除

```go
// 片段 · 演示
db.Delete(&model.User{}, id)            // 软删除：UPDATE users SET deleted_at=NOW() WHERE id=?
db.Unscoped().Delete(&model.User{}, id) // 物理删除（慎用）：DELETE FROM users WHERE id=?

var all []model.User
db.Unscoped().Find(&all)                // 查询含已软删的全部行
```

因为 `User` 带 `gorm.DeletedAt` 字段，`Delete` 自动变成打标记；之后所有常规查询自动追加 `WHERE deleted_at IS NULL`，像这行数据消失了一样。要连已删的一起查/真删，用 `Unscoped()`。

---

### 5.3 Select / Omit：别把 TEXT 大字段每次都捞出来

`short_links.original_url` 是 TEXT，列表页根本不展示它，却在每次 `Find` 时都从磁盘读出来、走一遍网络。查询也可以用 `Select`/`Omit` 控制列：

```go
// 片段 · 属于 internal/repository/link_repo.go（可作为列表查询的优化版）
var links []model.ShortLink
err := db.WithContext(ctx).
	Select("id", "short_code", "title", "status", "created_at").
	Where("user_id = ?", userID).
	Find(&links).Error

// 或者反着写：除了 original_url 都要
err = db.WithContext(ctx).
	Omit("original_url").
	Where("user_id = ?", userID).
	Find(&links).Error
```

注意：未选中的字段是**零值**（如 `OriginalURL == ""`），不是“数据库里是空”。序列化返回前要清楚哪些字段没查，避免把零值当真值用。

---

### 5.4 带 owner 与 version 的受控更新（乐观锁）

更新短链时同时带 owner 与 version，既防越权又防并发覆盖：

```go
// 片段 · 属于 internal/repository/link_repo.go（错误哨兵来自 §4.1 apperr）
res := db.WithContext(ctx).Model(&model.ShortLink{}).
	Where("id = ? AND user_id = ? AND version = ?", id, userID, oldVersion).
	Updates(map[string]any{
		"title":      title,
		"version":    gorm.Expr("version + 1"),
		"updated_at": time.Now(),
	})
if res.Error != nil {
	return fmt.Errorf("update owned link: %w", res.Error)
}
if res.RowsAffected != 1 {
	var owned int64
	if err := db.WithContext(ctx).Model(&model.ShortLink{}).
		Where("id = ? AND user_id = ?", id, userID).
		Count(&owned).Error; err != nil {
		return fmt.Errorf("check owned link after update conflict: %w", err)
	}
	if owned == 0 {
		return apperr.ErrNotFound
	}
	return apperr.ErrConflict // 资源属于当前用户，但 version 已变化
}
```

只检查 `.Error` 不够：UPDATE/DELETE 命中 0 行通常**不会报 SQL 错误**，必须检查 `RowsAffected`。这里 `RowsAffected != 1` 有两种可能——资源不存在/不属于该用户（404），或 version 被并发改掉了（409）——再补一次查询区分开。

---

## 6. 事务

```go
// 片段 · 属于 internal/repository/user_repo.go（UserProfile 定义见 §2.1）
func (r *UserRepository) CreateWithProfile(ctx context.Context, u *model.User, bio string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 事务内所有操作必须用 tx，用回 r.db 就逃出事务了！
		if err := tx.Create(u).Error; err != nil {
			return err // 自动 Rollback
		}
		profile := model.UserProfile{UserID: u.ID, Bio: bio}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}
		return nil // Commit
	})
}
```

对照 [Java 06 事务](../Java/06-MySQL基础索引与事务.md)：**要么全成功要么全撤销**；Go 里 return error 即回滚。注意 `u.ID` 在第一个 `Create` 后已被回填，第二条记录才拿得到外键。

另一条硬规则：**不要在事务函数里开 goroutine 共用 `tx`**。事务在其生命周期内绑定连接池中的同一条连接（§1.4），`tx` 句柄不是并发安全的；需要并行的工作要么各自开事务，要么放到事务提交之后再做。

### 6.1 事务边界怎么划

事务应覆盖“必须原子完成的数据库操作”，不要把以下慢操作包在事务中：外部 HTTP、Redis、发送 MQ、文件上传、等待用户输入。事务越长，连接占用和锁持有越久，死锁与超时概率越高。

如果数据库提交成功后还要发消息，可用 Outbox：在同一事务写业务表和 outbox 表，后台可靠投递；不要假设“先提交 DB，再发 MQ”天然原子。

### 6.2 隔离级别与锁

GORM 最终使用 `database/sql`，需要时可指定事务选项（需 import `"database/sql"`）：

```go
// 片段 · 演示
err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
	// 查询和更新都必须使用 tx，而不是外层 db。
	return tx.Model(&model.ShortLink{}).
		Where("id = ?", id).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
}, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
```

隔离级别不是越高越好：更强隔离通常意味着更高冲突成本。先根据业务不变量选择，再结合索引和锁范围验证。重试死锁时，整个事务函数必须可安全重放，并加入次数上限和随机退避。

### 6.3 避免“先读再写”的丢失更新

错误模式：读出 `ClickCount`，在 Go 中 `+1`，再 Save；两个并发请求可能都从 10 写成 11。使用数据库原子表达式：

```go
// 片段 · 演示
err := db.WithContext(ctx).
	Model(&model.ShortLink{}).
	Where("id = ?", id).
	UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
```

`gorm.Expr` 把 `click_count + 1` 原样送进 SQL，加法在数据库里原子完成。这里用 `UpdateColumn` 而非 `Updates`，顺带避免了自动更新 `updated_at`（点击不算“编辑”）。若更新依赖旧状态，可用乐观锁版本号：`WHERE id=? AND version=?`，并检查 `RowsAffected` 是否为 1（§5.4）。

```mermaid
sequenceDiagram
    participant S as Service
    participant G as GORM
    participant M as MySQL

    S->>G: Transaction(func)
    G->>M: BEGIN
    G->>M: INSERT user
    G->>M: INSERT profile
    alt 任一步 err
        G->>M: ROLLBACK
    else 全 OK
        G->>M: COMMIT
    end
```

---

## 7. 关联与预加载：把 N+1 讲透

### 7.1 声明关联

§2.1 的 `User` 里已经埋了这行：

```go
// 片段 · 属于 internal/model/user.go
Links []ShortLink `gorm:"foreignKey:UserID" json:"-"`
```

这声明了 **has many**（一对多）：一个 User 拥有多条 ShortLink，外键是 `short_links.user_id`。它不在 users 表生成任何列，只是告诉 GORM 两张表怎么连。

> 附带效应：默认配置下 `AutoMigrate` 会为声明的关联**自动创建外键约束**。不想让 GORM 管外键（交给 §3 的迁移 SQL），可在 `gorm.Config` 设 `DisableForeignKeyConstraintWhenMigrating: true`。

### 7.2 N+1 问题现场重现

需求：列出 10 个用户和各自的短链。直觉写法：

```go
// ❌ 片段 · N+1 反例：1 次查用户 + 10 次查短链 = 11 条 SQL
var users []model.User
db.Limit(10).Find(&users)
for i := range users {
	db.Where("user_id = ?", users[i].ID).Find(&users[i].Links)
}
```

用户 10 个就是 11 条 SQL，1000 个就是 1001 条——每条都要一次网络往返，列表接口被拖垮。这就是 **N+1 问题**：1 次主查询 + N 次关联查询。开着 §1.3 的 Info 日志跑一次，SQL 刷屏会给你留下深刻印象。

### 7.3 Preload 正解

```go
// ✅ 片段 · Preload：固定 2 条 SQL
var users []model.User
db.Preload("Links").Limit(10).Find(&users)
```

GORM 实际执行的是：

```sql
SELECT * FROM users WHERE deleted_at IS NULL LIMIT 10;
SELECT * FROM short_links WHERE user_id IN (1,2,3,...,10) AND deleted_at IS NULL;
```

先查主表，再用一条 `IN` 查询把所有关联行取回，按 `user_id` 拆回各自的 `Links` 字段。**用户再多也是 2 条 SQL**。`Preload` 的参数 `"Links"` 是**字段名**（不是表名）。还可以给关联加条件：

```go
// 片段 · 只预加载启用状态的短链
db.Preload("Links", "status = ?", model.LinkStatusActive).Limit(10).Find(&users)
```

注意 Preload 的第二条 SQL 同样自动带软删除过滤。

### 7.4 Joins 的适用场景与取舍

`Joins` 用一条 `LEFT JOIN` SQL 搞定，但它适合 **belongs to / has one**（一对一）——比如“短链带上它的作者”。对 has many 用 JOIN 会让主表每行重复 N 遍（行数膨胀），GORM 需要去重拼装，得不偿失。经验法则：**一对一用 `Joins`，一对多用 `Preload`**；确有复杂聚合需求就手写 SQL（§8.3）。

---

## 8. 批量写入、Upsert 与 Raw SQL

### 8.1 CreateInBatches：批量插入

循环调 `Create` 插 1 万行 = 1 万次网络往返 + 1 万个默认小事务。批量写法：

```go
// 片段 · 演示（links 为 []model.ShortLink，需已填好必填字段）
if err := db.CreateInBatches(links, 500).Error; err != nil {
	return fmt.Errorf("batch insert links: %w", err)
}
```

每批生成一条多值 `INSERT INTO ... VALUES (...),(...),...`，500 行一趟。批大小别贪大：单条 SQL 受 `max_allowed_packet` 限制，超大批还会拉长锁持有和 undo 日志。几百到一千是常见档位。批量插入后自增主键同样会回填到每个元素。

### 8.2 Upsert：clause.OnConflict

“存在就更新，不存在就插入”（比如从 CSV 导入短链，重复短码要覆盖标题）。SQL 层是 `INSERT ... ON DUPLICATE KEY UPDATE`，GORM 写法（需 import `"gorm.io/gorm/clause"`）：

```go
// 片段 · 冲突则更新指定列
err := db.Clauses(clause.OnConflict{
	Columns:   []clause.Column{{Name: "short_code"}},
	DoUpdates: clause.AssignmentColumns([]string{"title", "original_url", "updated_at"}),
}).Create(&links).Error

// 片段 · 冲突则静默跳过（导入去重）
err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&links).Error
```

说明：

- `AssignmentColumns` 表示“这些列用**新插入值**覆盖旧行”。
- MySQL 的 `ON DUPLICATE KEY UPDATE` 由**任意唯一键**触发，并不看 `Columns` 写了谁（那是给 PostgreSQL 等库用的）；但建议写上表达意图，代码可跨库。
- Upsert 与 `CreateInBatches` 可叠加：`db.Clauses(...).CreateInBatches(links, 500)`。

### 8.3 Raw / Scan / Exec：复杂查询的逃生门

开篇说过“复杂查询仍可 Raw”。统计类 SQL 用 ORM 硬拼不如直接写：

```go
// 片段 · 近 30 天每日新增短链数（放 LinkRepository 即可）
type DailyCount struct {
	Day   string
	Total int64
}

var rows []DailyCount
err := db.WithContext(ctx).Raw(`
    SELECT DATE(created_at) AS day, COUNT(*) AS total
    FROM short_links
    WHERE user_id = ? AND deleted_at IS NULL
    GROUP BY DATE(created_at)
    ORDER BY day DESC
    LIMIT ?`, userID, 30).Scan(&rows).Error
```

- `Scan` 按**列别名 → 字段名**（蛇形转驼峰：`day`→`Day`）把结果填进任意 struct 切片，struct 不必是 Model。
- `?` 占位符依然由驱动参数化，**防注入**；永远不要 `fmt.Sprintf` 拼 SQL。
- **Raw 不会自动加软删除过滤**——`deleted_at IS NULL` 要自己写，这是从 ORM 便利区走出来的代价。

无结果集的语句用 `Exec`：

```go
// 片段 · 清零某用户所有短链点击数
err := db.WithContext(ctx).
	Exec("UPDATE short_links SET click_count = 0 WHERE user_id = ?", userID).Error
```

**选型注（GORM vs sqlx）**：如果团队偏好全程手写 SQL，`sqlx` 是常见替代（薄封装 + 结果映射，无 ORM 语义）。本教程的取舍是：业务 CRUD 用 GORM 提效，复杂查询与极致性能路径走本节的 Raw/Exec 逃生门，不必再引第二个库。

---

## 9. Hooks：模型生命周期回调

§2.3 的 `BeforeCreate` 就是一个 Hook——只要给 Model 定义特定签名的方法，GORM 会在对应时机自动调用：

```go
// 片段 · 属于 internal/model/short_link.go（完整清单见 §2.3）
func (l *ShortLink) BeforeCreate(tx *gorm.DB) error {
	if l.OriginalURL == "" {
		return fmt.Errorf("original url must not be empty")
	}
	if l.Status == 0 {
		l.Status = LinkStatusActive
	}
	if l.Title == "" {
		l.Title = "未命名链接"
	}
	return nil
}
```

可用的 Hook：`BeforeSave` / `BeforeCreate` / `AfterCreate` / `AfterSave`、`BeforeUpdate` / `AfterUpdate`、`BeforeDelete` / `AfterDelete`、`AfterFind`。适合做：兜底默认值、不变量校验、归一化（如 URL 去空格）。

**规则与坑**：

| 要点 | 说明 |
|------|------|
| 返回 error 即中断 | 本次操作取消；因默认事务（§1.6）存在，已执行部分回滚 |
| 与写入同事务 | Hook 里需要再查库时用参数 `tx`，保持同一事务；别用全局 db |
| 不触发的场景 | `UpdateColumn(s)`、`Raw`/`Exec`、`Session(&gorm.Session{SkipHooks: true})` |
| 批量也触发 | `CreateInBatches` 对每个元素都调 Hook |
| 别做外部 IO | Hook 在事务里执行，放 HTTP/Redis/MQ 会拖长事务（§6.1 同理） |

跳过 Hook 的写法（如数据修复脚本）：

```go
// 片段 · 演示
db.Session(&gorm.Session{SkipHooks: true}).Create(&link)
```

Hook 是“模型级”逻辑，只放**跟这行数据自身有关**的规则；跨模型的业务流程仍归 Service 层——别把业务全塞进 Hook，那会变成谁也找不到入口的暗逻辑。

---

## 10. 与 Gin 集成与优雅关闭

把本章零件装配成能跑的服务，并补上 06 章埋的伏笔：显式 `http.Server` + **优雅关闭**（先停 HTTP 再关连接池）。

**文件：`cmd/server/main.go`（完整清单）**

```go
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/you/shortlink-api/internal/apperr"
	"github.com/you/shortlink-api/internal/model"
	"github.com/you/shortlink-api/internal/repository"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func loadConfig() repository.Config {
	cfg := repository.Config{
		User:     getenv("DB_USER", "root"),
		Password: getenv("DB_PASSWORD", "root123"),
		Host:     getenv("DB_HOST", "127.0.0.1"),
		Port:     3306,
		DBName:   getenv("DB_NAME", "shortlink"),
	}
	if p := os.Getenv("DB_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			cfg.Port = n
		}
	}
	return cfg
}

type createUserReq struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	Email    string `json:"email" binding:"omitempty,email"`
}

func main() {
	db, err := repository.NewDB(loadConfig())
	if err != nil {
		log.Fatalf("init database: %v", err)
	}
	// 本地练习用 AutoMigrate；部署版由发布步骤运行版本化迁移（§3）。
	if os.Getenv("APP_ENV") != "prod" {
		if err := repository.AutoMigrate(db); err != nil {
			log.Fatalf("migrate database: %v", err)
		}
	}
	userRepo := repository.NewUserRepository(db)

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	// 为让清单自包含，这里在闭包里直接调 Repository；
	// 项目版应按 06 章分层：Handler -> Service -> Repository。
	r.POST("/api/v1/users", func(c *gin.Context) {
		var req createUserReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		u := &model.User{
			Username: req.Username,
			Password: req.Password, // 明文仅为演示，09 章换 bcrypt
			Email:    req.Email,
		}
		if err := userRepo.Create(c.Request.Context(), u); err != nil {
			if errors.Is(err, apperr.ErrConflict) {
				c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
				return
			}
			log.Printf("create user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		c.JSON(http.StatusCreated, u)
	})
	r.GET("/api/v1/users/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		u, err := userRepo.GetByID(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, apperr.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			log.Printf("get user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		c.JSON(http.StatusOK, u)
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve http: %v", err)
		}
	}()
	log.Println("listening on :8080, press Ctrl+C to stop")

	// Windows 下 Ctrl+C 触发 os.Interrupt；SIGTERM 用于 Linux 容器 stop（13 章）。
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil { // 1) 先停 HTTP：不再接新请求，等在途请求完成
		log.Printf("http shutdown: %v", err)
	}
	if sqlDB, err := db.DB(); err == nil { // 2) 再关连接池：此时不再有查询进来
		if cerr := sqlDB.Close(); cerr != nil {
			log.Printf("close mysql pool: %v", cerr)
		}
	}
	log.Println("bye")
}
```

**关闭顺序为什么重要**：反过来先 `Close` 连接池，在途请求的 SQL 会全部报错——优雅关闭的原则是**从入口往里关**：停接新请求 → 等在途请求 → 关下游资源。`Shutdown(ctx)` 的 10 秒是给在途请求的最后期限。

运行与验证（PowerShell）：

```powershell
go run ./cmd/server
```

另开一个 PowerShell 窗口：

```powershell
# PowerShell 5.1 里 curl 是 Invoke-WebRequest 的别名，测 JSON API 用 Invoke-RestMethod 最省心
Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8080/api/v1/users" `
  -ContentType "application/json" `
  -Body '{"username":"alice","password":"secret123","email":"alice@example.com"}'

Invoke-RestMethod "http://127.0.0.1:8080/api/v1/users/1"
```

预期：第一条返回 `id=1` 的用户 JSON（无 password 字段，`json:"-"` 生效）；重复执行第一条返回 409；GET 返回同一用户。也可用真正的 curl（注意是 `curl.exe`，引号要转义）：

```powershell
curl.exe -X POST http://127.0.0.1:8080/api/v1/users -H "Content-Type: application/json" -d "{\"username\":\"bob\",\"password\":\"secret123\"}"
```

**验证持久化与优雅关闭**：在服务端窗口按 `Ctrl+C`，应看到 `shutting down... / bye`；重新 `go run ./cmd/server` 后 `GET /api/v1/users/1` 依然返回 alice——数据在 MySQL 里，重启不丢，这就是 06→07 的质变。

**依赖方向**：Handler → Service → Repository → GORM，禁止 Handler 直接 `db.Create`。本清单为了自包含在闭包里调了 Repository，10 章项目实战会把这些闭包拆回 handler/service 包。

---

## 11. Repository 层怎么测试

写进简历的项目必须有测试。Repository 的测试有两派：

| 方案 | 原理 | 优点 | 缺点 |
|------|------|------|------|
| go-sqlmock | 伪造 `*sql.DB`，断言生成的 SQL 文本 | 快、无环境依赖 | 和 GORM 生成的 SQL 细节强耦合，重构就碎；测不出真实约束 |
| **真 MySQL 集成测试（推荐）** | 连一个专用测试库跑真 SQL | 唯一约束/软删除/时区全是真的 | 需要 Docker MySQL 在跑 |
| testcontainers-go | 测试代码自动起临时 MySQL 容器 | 适合 CI，环境自管理 | 首次配置略繁，留到 12 章之后 |

本机已有 study-mysql，直接给它建一个**独立测试库**（绝不能用开发库，测试会清表！）：

```powershell
docker exec study-mysql mysql -uroot -proot123 -e "CREATE DATABASE IF NOT EXISTS shortlink_test"
```

**文件：`internal/repository/user_repo_test.go`（完整清单）**

```go
package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/you/shortlink-api/internal/apperr"
	"github.com/you/shortlink-api/internal/model"
)

// newTestDB 连接专用测试库；未配置环境变量时自动跳过（不阻塞 go test ./...）。
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN 未设置，跳过集成测试")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		TranslateError: true,
	})
	if err != nil {
		t.Fatalf("connect test mysql: %v", err)
	}
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	// 清空表，让每次运行从干净状态开始（只清测试库！）
	if err := db.Exec("DELETE FROM short_links").Error; err != nil {
		t.Fatalf("clean short_links: %v", err)
	}
	if err := db.Exec("DELETE FROM users").Error; err != nil {
		t.Fatalf("clean users: %v", err)
	}
	return db
}

func TestUserRepository_CreateAndGet(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	u := &model.User{
		Username: fmt.Sprintf("alice_%d", time.Now().UnixNano()%1_000_000),
		Password: "not-a-real-hash",
		Email:    "alice@example.com",
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("create: %v", err)
	}
	if u.ID == 0 {
		t.Fatal("expected auto-increment id to be set")
	}

	got, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Username != u.Username {
		t.Fatalf("username mismatch: got %q want %q", got.Username, u.Username)
	}

	// 重复用户名 -> ErrConflict（真唯一索引裁决，不是 mock）
	dup := &model.User{Username: u.Username, Password: "x"}
	if err := repo.Create(ctx, dup); !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("want ErrConflict, got %v", err)
	}

	// 不存在 -> ErrNotFound
	if _, err := repo.GetByID(ctx, 99_999_999); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
```

运行（PowerShell）：

```powershell
$env:TEST_MYSQL_DSN = "root:root123@tcp(127.0.0.1:3306)/shortlink_test?charset=utf8mb4&parseTime=True&loc=UTC"
go test ./internal/repository/ -v
```

预期输出末尾：

```text
--- PASS: TestUserRepository_CreateAndGet (0.xx s)
PASS
ok      github.com/you/shortlink-api/internal/repository
```

设计要点：

- **环境变量开关 + `t.Skip`**：没配 DSN 时测试自动跳过，`go test ./...` 在任何机器上都不会红。
- **每次先清表**：测试必须从已知状态出发；清的是 `shortlink_test`，与开发库隔离。
- 这个测试真正验证了：自增回填、唯一索引→`ErrConflict` 映射（连 `TranslateError` 一起测了）、`ErrNotFound` 语义。这三条 sqlmock 都测不真。
- `go test` 的系统用法（表驱动、覆盖率）在 [12 章](./12-单元测试日志与配置工程化.md)展开，这里先会跑即可。

---

## 12. 常见错误对照表

| 错误 | 原因 | 处理 |
|------|------|------|
| `Error 1062 Duplicate entry` | 唯一索引冲突 | 开 `TranslateError` 后判 `gorm.ErrDuplicatedKey`，业务层转 409（§4.2） |
| `record not found` | First/Take 无行 | 判 `gorm.ErrRecordNotFound` 转 404，勿当 500 |
| `Find` 查不到但不报错 | Find 的语义是“空集合合法” | 判 `len(slice)==0` / `RowsAffected==0`（§4.3） |
| 第二次查询条件莫名变多 | 复用了 Chain 方法返回的 `q` | 每次从 db 重新起链，或 `Session` 定格（§1.5） |
| `Updates(struct)` 清不掉字段 | struct 形式跳过零值 | 用 map 或 `Select("col").Updates`（§5.1） |
| 时间差 8 小时 | 应用与 MySQL session 时区不一致 | 项目统一 `loc=UTC`，并把 session `time_zone` 设为 `+00:00`（§1.3） |
| 中文乱码 | charset 非 utf8mb4 | DSN + 表 utf8mb4 |
| 连接耗尽 | 未 SetMaxOpenConns | 配连接池（§1.4） |
| 慢查询 | 缺索引 | EXPLAIN，见 Java 06；配 SlowThreshold 抓现行（§1.6） |
| `a`/`A` 短码冲突 | 列使用大小写不敏感 collation | `ascii_bin` + 唯一索引（§2.3） |
| 更新返回成功但数据没变 | 没检查 `RowsAffected` | 要求恰好命中预期行数（§5.4） |
| 接口偶发 500 且无上下文 | 直接返回/忽略 `.Error` | 每一步检查并 `%w` 包装操作语义 |
| `undefined: time/context` | 照抄代码块漏 import | 用本章完整清单；片段不要单独编译 |

---

## 13. 练习建议

### 基础

1. 跑通 §10 完整清单，补一个 `DELETE /api/v1/users/:id`（软删除，返回 204）
2. 用户名唯一冲突返回 **409**（`ErrConflict`）——记住划分：400 给参数校验失败，409 给业务冲突

### 进阶

3. 实现 ShortLink 的 `GetByShortCode`（用 §4.2 的 First 模式 + `ErrNotFound`）
4. 分页 `GET /api/v1/links?page=1&size=10`（§4.4；故意传 `page=0`、`size=999` 验证归一化）
5. 用 `Preload` 实现 `GET /api/v1/users/:id/links`，开 Info 日志确认只有 2 条 SQL（§7）
6. 批量导入 100 条短链：`CreateInBatches` + `OnConflict DoNothing`，返回成功条数（§8）

### 挑战

7. 事务：注册用户 + 插入欢迎短链占位，第二步人为报错验证回滚（§6）
8. 对 `ListByUser` 做 EXPLAIN，对照 Java 06 确认命中 `idx_links_user_page`
9. 并发 100 个 goroutine 强制插入同一 short_code，用 §11 的测试方法断言数据库最终只有 1 行（§4.6）
10. 写 migration 集成测试：空库 `up` 后检查索引存在，再执行一次 `up` 确保版本不漂移（§3 + §11）

---

*下一章：[08-Redis与go-redis缓存实战](./08-Redis与go-redis缓存实战.md)*——07 章每次跳转都查 MySQL，高 QPS 扛不住；08 章用 **go-redis** 做 Cache Aside，理论对照 [Java 07 Redis](../Java/07-Redis核心原理与缓存实战.md)。
