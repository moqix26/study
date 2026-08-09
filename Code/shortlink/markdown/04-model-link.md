# internal/model/link.go 逐块精讲

> 主线请跟 [`../study.md`](../study.md)；本文是精读加餐。

> 对应源码：`internal/model/link.go`  
> 目标：搞清每个字段、json/gorm tag、`ClickCount`、`size:16` 与默认短码长度 6 的关系。

---

## 0. Model 层在整体里的位置

```text
HTTP JSON  ←→  model.Link（结构体 + tag）
                    ↕
              GORM AutoMigrate / CRUD
                    ↕
              MySQL 表 links
```

`model` 包**只放数据结构**，不写 HTTP、不写 SQL 字符串（除了通过 GORM 标签声明约束）。它是 **DB 行** 与 **API JSON** 的「共同外形」。

---

## 1. 完整源码

```go
package model

import "time"

// Link 对应 MySQL 表 links。
type Link struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Code       string    `json:"code" gorm:"size:16;uniqueIndex;not null"`
	LongURL    string    `json:"long_url" gorm:"size:2048;not null"`
	ClickCount int64     `json:"click_count" gorm:"not null;default:0"`
	CreatedAt  time.Time `json:"created_at"`
}
```

---

## 2. `package model` 与 `import`

```go
package model

import "time"
```

| 符号 | 含义 |
|------|------|
| `package model` | 领域模型包；被 `repo`、`service` 引用 |
| `import "time"` | 仅 `CreatedAt` 字段需要 `time.Time` |

本文件无第三方依赖——保持 model **纯净**，避免循环 import。

---

## 3. `Link` 结构体总览

```go
type Link struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Code       string    `json:"code" gorm:"size:16;uniqueIndex;not null"`
	LongURL    string    `json:"long_url" gorm:"size:2048;not null"`
	ClickCount int64     `json:"click_count" gorm:"not null;default:0"`
	CreatedAt  time.Time `json:"created_at"`
}
```

GORM 默认表名：结构体 `Link` → 复数蛇形 **`links`**。

---

## 4. 字段逐列精讲

### 4.1 `ID`

```go
ID uint `json:"id" gorm:"primaryKey"`
```

| 部分 | 含义 |
|------|------|
| `uint` | 无符号整数；GORM 映射 MySQL 常见为 `BIGINT UNSIGNED` 自增 |
| `json:"id"` | 序列化成 JSON 时字段名为 `id`（小写） |
| `gorm:"primaryKey"` | 主键；`Create` 成功后 GORM 会回填 `ID` |

**业务上：** 对外主要用 `Code` 查链；`ID` 是数据库内部行标识。

**为何不用 `Code` 做主键？**

- 短码是随机字符串，作聚簇主键不如自增整型友好（索引、JOIN 习惯）。
- 短码用 **唯一索引** 保证不重复即可。

---

### 4.2 `Code`

```go
Code string `json:"code" gorm:"size:16;uniqueIndex;not null"`
```

| tag 片段 | 含义 |
|----------|------|
| `json:"code"` | API 里叫 `code` |
| `size:16` | 列最长 16 字符（`VARCHAR(16)` 一类） |
| `uniqueIndex` | 唯一索引：两个相同 `code` 不能同时存在 |
| `not null` | 不允许 NULL |

**`size:16` 与默认 `CodeLength=6` 的关系：**

| 概念 | 值 | 说明 |
|------|-----|------|
| 运行时生成的短码长度 | **6**（`SHORTLINK_CODE_LEN` / `config.CodeLength`） | `shortcode.Random(6)`、`Resolve` 里 `len(code)!=6` 直接当不存在 |
| 数据库列上限 | **16**（`gorm:"size:16"`） | 列比当前业务码**更长**，留扩展空间 |

```text
现在：6 位码  BaLrEf     → 远小于 16，安全
以后：若改成 8～10 位   → 仍可容纳，不必立刻改表（在 16 以内）
若配置 CodeLength=20   → 可能截断或插入失败，与模型不一致
```

**原则：** `CodeLength` ≤ `gorm size`；改配置长度时要同时检查 service 校验、随机码生成、模型 `size`。

**唯一索引与创建流程：**

`service.Create` 循环插入，撞 `uniqueIndex` → `repo.IsDuplicate` → 重试。见 [`05-repo-link.md`](./05-repo-link.md)。

---

### 4.3 `LongURL`

```go
LongURL string `json:"long_url" gorm:"size:2048;not null"`
```

| 部分 | 含义 |
|------|------|
| `json:"long_url"` | JSON 蛇形命名，与 Go 字段 `LongURL` 解耦 |
| `size:2048` | 最长 2048 字符；防超长 URL 撑爆行 |
| `not null` | 必须有目标长链 |

**上游：** `service.Create` 里 `urlx.Normalize` 校验后再写入。  
**下游：** 跳转时 `Resolve` 返回给 handler 做 302；Redis 缓存的值也是这个字符串。

---

### 4.4 `ClickCount`

```go
ClickCount int64 `json:"click_count" gorm:"not null;default:0"`
```

| 部分 | 含义 |
|------|------|
| `int64` | 点击次数可能很大，用 64 位 |
| `json:"click_count"` | API 若返回整行时可带此字段 |
| `not null;default:0` | 新行默认 0 次点击 |

**何时增加？** 不在 model 里——在 `repo.IncrClick` + `service.IncrClickAsync`（跳转成功后异步 `UPDATE click_count = click_count + 1`）。

**为何异步？** 跳转要快；计数失败只打日志，不阻塞 302。

**V1 JSON 接口** `GET /api/links/:code` 当前主要返回 `code/long_url/short_url`，未暴露 `click_count`；但字段已在表里，便于以后扩展统计 API。

---

### 4.5 `CreatedAt`

```go
CreatedAt time.Time `json:"created_at"`
```

| 部分 | 含义 |
|------|------|
| 无 gorm tag | GORM 约定：`CreatedAt` 自动维护创建时间 |
| `time.Time` | 依赖 DSN 里 `parseTime=True` |

`Create` 时若未手动赋值，GORM 通常写入当前时间。

---

## 5. json tag 与 gorm tag 分工

| 维度 | `json:"..."` | `gorm:"..."` |
|------|--------------|--------------|
| 谁读 | `encoding/json`、Gin `c.JSON` | GORM ORM |
| 作用 | API 字段名、是否导出 | 列类型、索引、约束 |
| 例子 | `long_url` vs `LongURL` | `uniqueIndex` 建唯一索引 |

**可以不一致：** 例如 Go 叫 `LongURL`，JSON 叫 `long_url`，MySQL 列名默认 `long_url`（GORM 蛇形）。

**创建 API 为何不直接 Bind `Link`？**

客户端 POST 只应传 `url`，不应让用户指定 `id`/`code`/`click_count`。handler 用单独的 `createReq{URL}`，service 构造 `&model.Link{Code, LongURL}`。

---

## 6. 与上下游怎么接

### 6.1 上游（谁构造 `Link`）

```go
// internal/service/link.go
link := &model.Link{Code: code, LongURL: longURL}
err = s.repo.Create(link)
```

只设 `Code`、`LongURL`；`ID`、`ClickCount`、`CreatedAt` 由 DB/GORM 填充。

### 6.2 下游（谁读写 `Link`）

| 组件 | 用法 |
|------|------|
| `repo.AutoMigrate(&model.Link{})` | 建表 |
| `repo.Create(link *model.Link)` | INSERT |
| `repo.FindByCode` | SELECT 整行 → `*model.Link` |
| `repo.IncrClick` | 只 UPDATE `click_count`，不加载整 struct |

### 6.3 数据流简图

```text
POST {"url":"..."}
  → service 生成 code
  → model.Link{Code, LongURL}
  → repo.Create → 表 links 一行

GET /:code
  → Resolve 主要用 LongURL（缓存存 string）
  → IncrClick 只动 click_count 列
```

---

## 7. 常见坑

| 坑 | 现象 | 修法 |
|----|------|------|
| `CodeLength` 改成 8，tag 仍写 `size:6` | 迁移后列长不够或行为混乱 | 保持 `size` ≥ 最大预期码长 |
| 只有 json tag 没有 gorm 约束 | 库里能插重复 code | 必须 `uniqueIndex` |
| `ClickCount` 用 `int` | 极大点击量溢出 | 用 `int64` |
| 忘记 `parseTime` | `CreatedAt` 解析异常 | DSN 加 `parseTime=True` |
| 把 `Link` 直接绑 POST body | 用户可伪造 `code` | 用独立 request DTO |
| 改字段名不迁移 | 旧表列对不上 | 开发环境可删表重建；生产用正式 migration |

---

## 8. 本地怎么验证

### 8.1 看表结构

启动服务后（`AutoMigrate` 已跑）：

```powershell
docker exec -it study-mysql mysql -uroot -proot123 study -e "SHOW CREATE TABLE links\G"
```

关注：

- `code` 是否有 **UNIQUE** 索引
- `code` 类型是否为 `varchar(16)` 左右
- `click_count` 默认 0
- `long_url` 足够长

### 8.2 插入与唯一约束

```powershell
# 创建一条
curl.exe -X POST http://localhost:8080/api/links `
  -H "Content-Type: application/json" `
  -d '{"url":"https://example.com/a"}'
```

MySQL 里查：

```sql
SELECT id, code, long_url, click_count, created_at FROM links ORDER BY id DESC LIMIT 1;
```

### 8.3 验证 ClickCount

多次访问短链：

```powershell
curl.exe -I http://localhost:8080/<code>
```

再查 DB：

```sql
SELECT code, click_count FROM links WHERE code = '<code>';
```

应看到 `click_count` 递增（异步，可能有极短延迟）。

### 8.4 验证 Code 长度与 Resolve

```powershell
curl.exe http://localhost:8080/api/links/abc
# code 长度不是 6 → 404 not found（service.Resolve 前置校验）
```

---

## 9. 与旧版单文件对照

| 旧 `main.go.md` | 现 `model/link.go` |
|-----------------|----------------------|
| `Code` `size:6` | `size:16`（更宽松） |
| 无 `ClickCount` | 有 `click_count` |
| 结构体写在 main 包 | 独立 `internal/model` |

---

## 10. 口述检查（2～3 题）

1. **`gorm:"size:16"` 和 `config.CodeLength` 默认 6 矛盾吗？为什么要这样设计？**  
   （期望：不矛盾；列上限 ≥ 业务生成长度，便于以后加长短码而不立刻改表。）

2. **`uniqueIndex` 在创建短链流程里起什么作用？撞了怎么办？**  
   （期望：保证 code 唯一；`IsDuplicate` 后重试随机码，最多 `MaxRetries` 次。）

3. **`ClickCount` 为什么放在 model 里，却在跳转时异步更新？**  
   （期望：持久化在 DB；跳转路径要快，用 `IncrClickAsync` 不阻塞 302。）
