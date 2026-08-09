# 07 · shortcode 短码生成精读

> 主线请跟 [`../study.md`](../study.md)；本文是精读加餐。

> 对应源码：`internal/pkg/shortcode/shortcode.go`  
> 目标：把「6 位 base62 短码怎么生成、为什么这样写」讲透。

---

## 0. 这个包在整体里干什么

短链服务的**短码**是用户可见的 ID（如 `BaLrEf`）。创建流程里：

```text
POST /api/links
  → handler.CreateLink
  → service.Create
  → urlx.Normalize（校验长链）
  → shortcode.Random(6)（生成本次候选短码）  ← 本文
  → repo.Create（INSERT，唯一索引兜底碰撞）
```

**上游**：`internal/service/link.go` 的 `Create` 循环调用 `shortcode.Random`。  
**下游**：生成的 `code` 写入 `model.Link.Code`，经 GORM 落到 MySQL `links.code` 唯一索引。

本包**只做一件事**：给定长度 `n`，返回 `n` 位密码学安全随机 base62 字符串。不负责碰撞重试、不负责存库。

---

## 1. 完整源码（逐块对照）

```go
package shortcode

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Random 生成 n 位 base62 短码（密码学安全随机）。
func Random(n int) (string, error) {
	var b strings.Builder
	b.Grow(n)
	max := big.NewInt(int64(len(alphabet)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b.WriteByte(alphabet[idx.Int64()])
	}
	return b.String(), nil
}
```

---

## 2. `package shortcode` 与 `import`

```go
package shortcode
```

- 包名 `shortcode`，路径 `shortlink/internal/pkg/shortcode`。
- 放在 `internal/pkg/` 表示**可复用的小工具**，不依赖 HTTP、DB、Redis；单元测试时可直接 `shortcode.Random(6)`。

### import 表


| 包 | 本文件里干什么 |
| --- | --- |
| `crypto/rand` | 密码学安全随机源；`rand.Reader` 从操作系统熵池读随机字节 |
| `math/big` | `rand.Int` 返回 `*big.Int`；用 `big.NewInt` 表示上界 62 |
| `strings` | `strings.Builder` 高效拼接 `n` 个 byte，避免 `+=` 反复分配 |

**为什么不用 `math/rand`？**  
`math/rand` 默认可预测（种子固定则序列固定）。短码若可预测，攻击者可枚举或撞库。`crypto/rand` 适合「当作唯一标识符」的场景。

---

## 3. 常量 `alphabet`

```go
const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
```

| 名字 | 值 | 含义 |
| --- | --- | --- |
| `alphabet` | 62 个字符 | 数字 10 + 小写 26 + 大写 26 = **base62** |

| 为什么 base62 | 说明 |
| --- | --- |
| URL 友好 | 不含 `+`、`/`、`=` 等需编码字符，可直接拼在路径 `/:code` |
| 信息密度 | 每位 62 种；6 位空间 \(62^6 \approx 5.68 \times 10^{10}\)，练习项目碰撞极低 |
| 可读性 | 比 base64 少符号，复制粘贴少踩坑 |

**与配置的关系**：`config.CodeLength` 默认 `6`，`service.Create` 传入 `shortcode.Random(s.cfg.CodeLength)`。改长度要同时考虑路由校验（`Resolve` 里 `len(code) != s.cfg.CodeLength`）和表字段 `gorm:"size:16"`。

**与历史单文件对照**：旧 `main.go` 里叫 `codeAlphabet`，逻辑相同；分层后抽到独立包。

---

## 4. `Random(n int) (string, error)` 全文拆解

### 4.1 函数签名


| 参数/返回 | 类型 | 含义 |
| --- | --- | --- |
| `n` | `int` | 短码位数；业务上由 `cfg.CodeLength` 传入（默认 6） |
| 返回值 1 | `string` | 生成的短码；失败时为 `""` |
| 返回值 2 | `error` | 随机源异常时非 nil；正常恒为 nil |

### 4.2 `strings.Builder` 与 `Grow`

```go
var b strings.Builder
b.Grow(n)
```

| 调用 | 含义 |
| --- | --- |
| `var b strings.Builder` | 可变长度 byte 缓冲区，专门用于拼字符串 |
| `b.Grow(n)` | 预分配至少 `n` 字节容量，循环 `n` 次 `WriteByte` 时少扩容 |

**为什么不用 `s += string(c)`？**  
Go 里字符串不可变，每次 `+=` 可能分配新底层数组，6 次虽小，但 `Builder` 是惯用写法，和旧版 `randomCode` 一致。

### 4.3 `big.NewInt` 与上界 `max`

```go
max := big.NewInt(int64(len(alphabet)))
```

| 名字 | 值 | 含义 |
| --- | --- | --- |
| `len(alphabet)` | `62` | 字母表长度 |
| `max` | `*big.Int` 值为 62 | `rand.Int` 的**上界（不包含）**，即下标范围 `[0, 61]` |

**为什么用 `big.Int`？**  
`crypto/rand.Int(rand.Reader, max)` 的 API 要求 `max` 为 `*big.Int`，在 `[0, max)` 上均匀采样。这里 max=62，恰好对应 62 个字符下标。

### 4.4 循环：每一位怎么抽

```go
for i := 0; i < n; i++ {
	idx, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	b.WriteByte(alphabet[idx.Int64()])
}
```

| 名字 | 含义 |
| --- | --- |
| `i` | 当前正在生成第几位（0 到 n-1） |
| `rand.Reader` | 全局 `io.Reader`，读 OS 密码学随机字节 |
| `rand.Int(rand.Reader, max)` | 均匀随机整数 `idx`，`0 <= idx < 62` |
| `idx.Int64()` | `*big.Int` → `int64`，用作切片下标 |
| `alphabet[idx.Int64()]` | 取一个字符（`byte`） |
| `b.WriteByte(...)` | 写入 Builder |

**均匀性**：每一位独立均匀；6 位共 \(62^6\) 种组合（近似；若 `n` 变长则指数增长）。

**错误处理**：`rand.Int` 在极罕见环境（熵池耗尽等）可能失败，直接 `return "", err`，由 `service.Create` 向上返回 500，而不是生成弱随机码。

### 4.5 返回

```go
return b.String(), nil
```

| 调用 | 含义 |
| --- | --- |
| `b.String()` | 一次性分配最终 `string` |
| `nil` | 成功 |

---

## 5. 碰撞与重试（不在本包，但必须一起理解）

`shortcode.Random` **不保证**全局唯一。两批随机可能相同，概率约 \(1/62^6\)。

碰撞由**上层 + 数据库**处理：

```text
service.Create:
  for i := 0; i < MaxRetries; i++ {
    code := shortcode.Random(...)
    err := repo.Create(link)
    if err == nil → 成功
    if repo.IsDuplicate(err) → 下一轮换 code
    else → 其它 DB 错误，失败
  }
```

| 层级 | 职责 |
| --- | --- |
| `shortcode` | 只负责「像随机 ID」 |
| `links.code` UNIQUE | 最终权威：重复 INSERT 失败 |
| `MaxRetries`（默认 8） | 防止无限循环；8 次仍撞 → `"failed to allocate code"` |

**口述要点**：应用层重试 + DB 唯一约束是双保险；不要指望 Random  alone 去重。

---

## 6. 上下游数据流


```text
config.CodeLength (6)
       ↓
service.Create(rawURL)
       ↓
shortcode.Random(6) → "xY3kPq"
       ↓
model.Link{ Code: "xY3kPq", LongURL: "https://..." }
       ↓
repo.Create → INSERT INTO links
       ↓
CreateResult{ code, short_url, long_url }  // 不写 Redis
```

| 阶段 | 是否调用 shortcode |
| --- | --- |
| 创建 | 是，每轮重试重新 Random |
| Resolve / 跳转 | 否，只用已有 `code` 查缓存和库 |
| IncrClickAsync | 否 |

---

## 7. 常见坑


| 坑 | 现象 | 原因 / 对策 |
| --- | --- | --- |
| 用 `math/rand` 且无种子 | 重启后短码序列可预测 | 必须用 `crypto/rand` |
| `n <= 0` | 返回空串 `""`，`Resolve` 因长度不对直接当不存在 | 本函数未校验 `n`；依赖配置 `CodeLength > 0` |
| 以为 Random 保证唯一 | 极低概率撞唯一索引 | 靠 `IsDuplicate` 重试 |
| 改 `alphabet` 长度却忘记改 `max` | 下标越界 panic | `max` 必须用 `len(alphabet)` |
| 把短码当加密 | 62^6 可被离线枚举（若攻击者有意扫） | V1 练习够用；生产要加 rate limit、更长码或哈希 |

---

## 8. 验收（本包可单测；联调看创建接口）

### 8.1 联调（经 HTTP）

```powershell
cd F:\study\Code\shortlink
# 确保 mysql/redis 已启，go run ./cmd/server
$r = Invoke-RestMethod -Uri http://localhost:8080/api/links -Method POST `
  -ContentType "application/json" -Body '{"url":"https://www.example.com"}'
$r.code.Length   # 期望 6
$r.code -match '^[0-9a-zA-Z]{6}$'   # 期望 True
```

### 8.2 期望


| 检查项 | 期望 |
| --- | --- |
| `code` 长度 | 与 `SHORTLINK_CODE_LEN` 一致（默认 6） |
| 字符集 | 仅 `0-9a-zA-Z` |
| 多次创建 | `code` 不同（极高概率） |
| HTTP 状态 | `201 Created` |

### 8.3 可选：包内快速验证

在 `shortcode_test.go`（若你自写练习）里循环 `Random(6)` 100 次，断言长度与字符集；本仓库 V1 可不强制。

---

## 9. 与旧文档 / 单文件对照


| 旧名（main.go.md） | 现名 |
| --- | --- |
| `codeAlphabet` | `alphabet` |
| `randomCode(n)` | `shortcode.Random(n)` |
| `codeLen` | `config.CodeLength` |

行为等价；分层后便于单测和复用。

---

## 口述题

1. 为什么用 `crypto/rand` 而不是 `math/rand`？`rand.Reader` 是什么？
2. `rand.Int(rand.Reader, max)` 里 `max=62` 时，合法下标范围是多少？为什么用 `big.Int`？
3. `strings.Builder` 的 `Grow(n)` 解决什么问题？
4. 短码碰撞时，本包会不会重试？谁在重试？最多几次？
5. 6 位 base62 大约有多少种组合？为什么不用自增 ID 当短码？
