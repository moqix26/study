# internal/repo/link.go 逐块精讲

> 主线请跟 [`../study.md`](../study.md)；本文是精读加餐。

> 对应源码：`internal/repo/link.go`  
> 目标：搞清 `LinkRepo` 每个方法、GORM 用法、`IncrClick` 的原子更新、`IsDuplicate` 如何判断唯一冲突。

---

## 0. Repo 层在整体里的位置

```text
service（业务）
    ↓ 调用
repo（持久化）          ← 本文
    ↓ GORM
MySQL 表 links
```

**职责边界：** repo **只知道**「怎么存、怎么取」；不负责 URL 校验、随机码、Redis、HTTP 状态码。这样 service 可以组合 repo + cache 实现 Cache Aside。

---

## 1. 完整源码

```go
package repo

import (
	"errors"
	"strings"

	"shortlink/internal/model"

	"gorm.io/gorm"
)

type LinkRepo struct {
	db *gorm.DB
}

func NewLinkRepo(db *gorm.DB) *LinkRepo {
	return &LinkRepo{db: db}
}

func (r *LinkRepo) AutoMigrate() error {
	return r.db.AutoMigrate(&model.Link{})
}

func (r *LinkRepo) Create(link *model.Link) error {
	return r.db.Create(link).Error
}

func (r *LinkRepo) FindByCode(code string) (*model.Link, error) {
	var link model.Link
	err := r.db.Where("code = ?", code).First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *LinkRepo) IncrClick(code string) error {
	return r.db.Model(&model.Link{}).Where("code = ?", code).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
}

func IsDuplicate(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique") || strings.Contains(msg, "1062")
}
```

---

## 2. `import` 一览

| 包 | 用途 |
|----|------|
| `errors` | `errors.Is` 判断 `gorm.ErrDuplicatedKey` |
| `strings` | 错误信息字符串兜底匹配 |
| `shortlink/internal/model` | `model.Link` |
| `gorm.io/gorm` | ORM API、`gorm.Expr` |

---

## 3. `LinkRepo` 结构体与构造函数

```go
type LinkRepo struct {
	db *gorm.DB
}

func NewLinkRepo(db *gorm.DB) *LinkRepo {
	return &LinkRepo{db: db}
}
```

| 符号 | 含义 |
|------|------|
| `LinkRepo` | 短链表的仓储；持有 DB 句柄 |
| `db *gorm.DB` | 连接池指针；所有 SQL 经此发出 |
| `NewLinkRepo` | 工厂函数；在 `app.Run` 里 `NewLinkRepo(db)` 注入 |

**为何不用全局 `var db`？**

- 测试可传入内存 SQLite 或 mock DB。
- 多 repo 共存时边界清晰（以后若有 `UserRepo` 等同理）。

---

## 4. `AutoMigrate`

```go
func (r *LinkRepo) AutoMigrate() error {
	return r.db.AutoMigrate(&model.Link{})
}
```

| 符号 | 含义 |
|------|------|
| `AutoMigrate` | 根据 `model.Link` 的 gorm tag 创建或**增量对齐**表结构 |
| `&model.Link{}` | 传类型信息；不插入数据 |
| 返回值 | 迁移失败时的 error |

**会做什么（学习项目层面）：**

- 表 `links` 不存在 → 创建
- 缺列 → 尝试添加（如后来加了 `click_count`）
- **不会**随意删列、改已有列类型（GORM 偏保守）

**调用时机：** `app.Run` 在 `NewLinkRepo` 后立刻调用，早于 HTTP 监听。

**生产注意：** 正式环境常用 golang-migrate、Flyway 等版本化迁移；AutoMigrate 适合本地练习。

---

## 5. `Create`

```go
func (r *LinkRepo) Create(link *model.Link) error {
	return r.db.Create(link).Error
}
```

| 符号 | 含义 |
|------|------|
| `link *model.Link` | 待插入行；至少应有 `Code`、`LongURL` |
| `r.db.Create(link)` | 生成 `INSERT INTO links ...` |
| `.Error` | GORM 把执行错误放在 `Error` 字段 |

**成功后副作用：** GORM 常回填 `link.ID`、`link.CreatedAt` 等。

**上游：**

```go
link := &model.Link{Code: code, LongURL: longURL}
err = s.repo.Create(link)
```

**失败两类：**

| 类型 | 处理方 |
|------|--------|
| 唯一索引冲突（code 重复） | `IsDuplicate` → service 重试 |
| 连接断开、语法错误等 | service 直接返回 500 |

---

## 6. `FindByCode`

```go
func (r *LinkRepo) FindByCode(code string) (*model.Link, error) {
	var link model.Link
	err := r.db.Where("code = ?", code).First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}
```

| 符号 | 含义 |
|------|------|
| `var link model.Link` | 栈上空结构体，接收查询结果 |
| `Where("code = ?", code)` | 参数化查询，防 SQL 注入 |
| `First(&link)` | `SELECT ... LIMIT 1`；找不到返回 `gorm.ErrRecordNotFound` |
| `return &link` | 返回堆上数据的指针（link 逃逸到堆） |

**上游：** `service.Resolve` 在 Redis miss 后调用。

**错误约定：**

| error | service 处理 |
|-------|----------------|
| `gorm.ErrRecordNotFound` | 当作「短码不存在」，返回空 longURL |
| 其他 | 基础设施错误，500 |

**为何不在 repo 里吞掉 `NotFound`？**

- repo 保持「诚实」：查不到就是 error。
- 「不存在」是否算业务正常，由 service 决定——分层惯例。

---

## 7. `IncrClick`（重点）

```go
func (r *LinkRepo) IncrClick(code string) error {
	return r.db.Model(&model.Link{}).Where("code = ?", code).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
}
```

逐段：

| 片段 | 含义 |
|------|------|
| `Model(&model.Link{})` | 指定操作表 `links` |
| `Where("code = ?", code)` | 只更新该短码行 |
| `UpdateColumn(...)` | 更新单列，**不走** struct 零值覆盖其它字段 |
| `gorm.Expr("click_count + 1")` | SQL 表达式：在数据库侧 `click_count = click_count + 1` |

**为什么用 `UpdateColumn` + `Expr`，而不是 `link.ClickCount++` 再 `Save`？**

1. **原子性：** 单条 `UPDATE` 在 DB 内完成，避免「读-改-写」竞态（两次并发跳转不会丢一次计数）。
2. **不读整行：** 跳转路径不需要 `LongURL`，少一次 SELECT。
3. **避免零值陷阱：** `Updates(link)` 可能把未设字段写成零值；`UpdateColumn` 只动一列。

等价 SQL 概念：

```sql
UPDATE links SET click_count = click_count + 1 WHERE code = ?;
```

**上游：** `service.IncrClickAsync` 在 goroutine 里调用；失败只 `log.Println`。

---

## 8. `IsDuplicate`（包级函数）

```go
func IsDuplicate(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique") || strings.Contains(msg, "1062")
}
```

| 分支 | 含义 |
|------|------|
| `err == nil` | 没错误，当然不是重复 |
| `errors.Is(err, gorm.ErrDuplicatedKey)` | GORM v2 标准重复键错误 |
| 字符串包含 `duplicate` / `unique` | MySQL 英文报错兜底 |
| 包含 `1062` | MySQL 错误码 **1062 Duplicate entry** |

**为何三层判断？**

- 不同驱动、GORM 版本、MySQL 配置下错误类型/文案可能不同。
- 创建短链时**只有**重复 code 才值得重试；其它 DB 错误应立刻失败。

**上游：**

```go
if !repo.IsDuplicate(err) {
	return nil, err
}
// 否则 continue 下一轮随机 code
```

**注意：** `IsDuplicate` 是**启发式**，极端罕见误报可能把别的错误当重复；学习项目可接受。生产可解析 MySQL 具体 error number。

---

## 9. 与上下游怎么接

```text
app.Run
  └─ linkRepo := NewLinkRepo(db)
  └─ AutoMigrate()
  └─ service.NewLinkService(cfg, linkRepo, cache)

service.Create
  └─ repo.Create          // 可能 IsDuplicate

service.Resolve
  └─ repo.FindByCode      // cache miss

service.IncrClickAsync
  └─ repo.IncrClick       // 异步
```

repo **不**调用 cache、handler；单向依赖向下。

---

## 10. 常见坑

| 坑 | 现象 | 修法 |
|----|------|------|
| `First` 当不存在不检查 error | 空结构体当有效数据 | service 里 `errors.Is(NotFound)` |
| `Save` 覆盖整行 | 其它列被零值清空 | 计数用 `UpdateColumn` + `Expr` |
| 应用层 `count++` 写回 | 并发丢点击 | 必须用 SQL 原子 `+1` |
| 只依赖 `ErrDuplicatedKey` | 某些环境重试逻辑失效 | 保留字符串/1062 兜底 |
| 忘记 AutoMigrate | `Table doesn't exist` | 启动时 migrate |
| `Where` 拼接字符串 | SQL 注入风险 | 始终 `?` 占位符 |
| `Create` 传值非指针 | GORM 无法回填 ID | 传 `*model.Link` |

---

## 11. 本地怎么验证

### 11.1 Create + Find

```powershell
curl.exe -X POST http://localhost:8080/api/links `
  -H "Content-Type: application/json" `
  -d '{"url":"https://www.example.com/repo-test"}'
```

记下 `code`，然后：

```sql
SELECT * FROM links WHERE code = '<code>';
```

### 11.2 IncrClick

```powershell
curl.exe -I http://localhost:8080/<code>
curl.exe -I http://localhost:8080/<code>
```

```sql
SELECT click_count FROM links WHERE code = '<code>';
-- 应 ≥ 2
```

### 11.3 唯一约束（理解 IsDuplicate）

在 MySQL 手动插重复 code（仅实验）：

```sql
INSERT INTO links (code, long_url, click_count) VALUES ('duppp1', 'https://a.com', 0);
INSERT INTO links (code, long_url, click_count) VALUES ('duppp1', 'https://b.com', 0);
-- 第二条应失败：Duplicate entry
```

应用层 `Create` 撞唯一索引时，service 会自动换 code 重试，无需你手动处理。

### 11.4 FindByCode 不存在

```powershell
curl.exe http://localhost:8080/api/links/aaaaaa
# 404（Resolve 将 RecordNotFound 转为 not found）
```

---

## 12. 与旧版单文件对照

| 旧 main | 现 repo |
|---------|---------|
| 全局 `db.Create` | `LinkRepo.Create` |
| 全局 `db.First` | `FindByCode` |
| `isDuplicate` 小写私有 | `IsDuplicate` 包级导出供 service |
| 无 IncrClick | 独立原子更新 |

---

## 13. 口述检查（2～3 题）

1. **`FindByCode` 返回 `gorm.ErrRecordNotFound` 时，为什么 repo 不直接返回 `(nil, nil)`？**  
   （期望：repo 只报告存储事实；「不存在」的语义由 service 统一处理。）

2. **`IncrClick` 为什么用 `gorm.Expr("click_count + 1")` 而不是先 `Find` 再 `Save`？**  
   （期望：原子更新、防竞态、少一次读。）

3. **`IsDuplicate` 除了 `gorm.ErrDuplicatedKey` 为什么还要判断错误字符串和 1062？**  
   （期望：兼容不同驱动/版本的报错形态，确保碰撞重试逻辑稳定。）
