# 11 · middleware/logger.go 请求日志精读

> 主线请跟 [`../study.md`](../study.md)；本文是精读加餐。

> 对应源码：`internal/middleware/logger.go`（全文）  
> 目标：搞清 Gin 中间件执行顺序、`c.Next` 前后能读到什么、如何观察 `X-Cache`。

---

## 0. 这个文件在整体里干什么

每个 HTTP 请求进入 Gin 后，在到达 `handler.Health / CreateLink / ...` **之前和之后**各打一行日志，便于本地调试：方法、路径、最终状态码、缓存头。

```text
客户端请求
  → gin.Recovery（panic 恢复）
  → middleware.Logger  [IN]
  → handler 业务
  → middleware.Logger  [OUT]（含 Status、X-Cache）
```

注册见 `internal/app/app.go`：

```go
r := gin.New()
r.Use(gin.Recovery())
r.Use(middleware.Logger())
```

**顺序**：Recovery 在外层，Logger 在内层；panic 时 Recovery 先兜，Logger 的 OUT 行可能打不全（视 panic 时机而定）。

---

## 1. 完整源码

```go
package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path
		fmt.Println("[IN ]", method, path)
		c.Next()
		fmt.Println("[OUT]", method, path, "->", c.Writer.Status(), "X-Cache=", c.Writer.Header().Get("X-Cache"))
	}
}
```

---

## 2. `Logger()` 返回什么

```go
func Logger() gin.HandlerFunc
```

| 类型 | 含义 |
| --- | --- |
| `gin.HandlerFunc` | 类型别名：`func(c *gin.Context)` |
| 返回值 | **闭包**：Gin 对每个请求调用一次 |

**为什么不是 `func Logger(c *gin.Context)`？**  
`r.Use` 要的是「工厂」：注册时调用 `Logger()` 得到 handler；每个请求执行的是内层 `func(c *gin.Context)`。

---

## 3. `c.Next()` 之前：入站日志

```go
method := c.Request.Method
path := c.Request.URL.Path
fmt.Println("[IN ]", method, path)
```

| 变量/字段 | 示例 | 含义 |
| --- | --- | --- |
| `c` | — | 当前请求的 Gin 上下文 |
| `c.Request` | `*http.Request` | 标准库请求对象 |
| `Method` | `GET`、`POST` | HTTP 方法 |
| `URL.Path` | `/api/links/BaLrEf` | 路径（不含 query） |

示例输出：

```text
[IN ] POST /api/links
[IN ] GET /BaLrEf
```

**此时还没有**：最终状态码、`X-Cache`（handler 还没跑）。

---

## 4. `c.Next()`：放行

```go
c.Next()
```

| 行为 | 说明 |
| --- | --- |
| 调用时机 | IN 日志之后、OUT 日志之前 |
| 内部 | 执行**下一个**中间件，或最终路由 handler |
| 返回 | 整条链（含 handler）跑完后才返回到这里 |
| 嵌套 | 多个 `Use` 形成洋葱模型：先进 IN，Next 往里，再层层 OUT |

对本项目：

```text
Logger IN
  → Recovery（可能直接 Next）
  → 匹配路由 → handler.CreateLink / Redirect / ...
Logger OUT
```

**注意**：若 handler 里 `return` 提前结束，只要没 panic，仍会回到 `Next()` 之后打 OUT。

---

## 5. `c.Next()` 之后：出站日志

```go
fmt.Println("[OUT]", method, path, "->", c.Writer.Status(), "X-Cache=", c.Writer.Header().Get("X-Cache"))
```

| 调用 | 必须在 Next **之后**的原因 |
| --- | --- |
| `c.Writer.Status()` | 状态码由 handler 的 `c.JSON` / `c.Redirect` 等写入 ResponseWriter |
| `c.Writer.Header().Get("X-Cache")` | `handler.setCacheHeader` 在写 body 前设置 |

| 字段 | 示例 | 含义 |
| --- | --- | --- |
| `Status()` | `200`、`201`、`302`、`404` | 最终 HTTP 状态码 |
| `X-Cache` | `HIT`、`MISS`、空串 | 仅 `GetLinkJSON` / `Redirect` 设置；Health/Create 通常为空 |

示例：

```text
[OUT] POST /api/links -> 201 X-Cache=
[OUT] GET /api/links/BaLrEf -> 200 X-Cache= MISS
[OUT] GET /api/links/BaLrEf -> 200 X-Cache= HIT
[OUT] GET /BaLrEf -> 302 X-Cache= HIT
```

（`fmt.Println` 多个参数之间会自动加空格，故 `X-Cache= MISS` 中间有空格。）

---

## 6. 与 handler `X-Cache` 的配合


```text
handler.GetLinkJSON / Redirect
  → service.Resolve → hit bool
  → setCacheHeader(c, hit)  // c.Header("X-Cache", ...)
  → c.JSON / c.Redirect
       ↓
middleware Logger OUT 读取 c.Writer.Header().Get("X-Cache")
```

| 路由 | 典型 OUT 中 X-Cache |
| --- | --- |
| `GET /health` | 空 |
| `POST /api/links` | 空 |
| 首次 `GET /api/links/:code` | `MISS` |
| 二次同 URL | `HIT` |
| 首次 `GET /:code` 跳转 | `MISS` |

**用途**：终端里对比两次请求，无需每次手打 `curl -i`（但验收清单仍建议 curl 看原始头）。

---

## 7. 为什么用 `fmt.Println` 而不是 gin 默认 Logger

| 点 | 说明 |
| --- | --- |
| `gin.New()` | 不用 `gin.Default()`，避免内置 Logger 重复 |
| 自定义格式 | `[IN]`/`[OUT]` 一眼区分；带上 `X-Cache` |
| 学习向 | 代码短，看清中间件本质 |

生产会换结构化日志（zap/slog）、request id、耗时等；V1 保持最小。

---

## 8. 上下游


| 方向 | 关系 |
| --- | --- |
| 上游 | `app.Run` 里 `r.Use(middleware.Logger())` |
| 下游 | 所有注册的路由 handler |
| 依赖 handler | OUT 行的 Status / X-Cache 由 handler 写入 |

本包**不**调 service、不碰 DB/Redis。

---

## 9. 常见坑


| 坑 | 说明 |
| --- | --- |
| 在 `Next()` **前**读 `Status()` | 往往是 `200` 默认值，误导 |
| 以为 Create 会有 X-Cache | 创建不走 Resolve，OUT 里为空正常 |
| 中间件顺序反了 | Logger 若在 Recovery 外，panic 时行为不同；当前 Recovery 先注册=外层，合理 |
| 用 `URL.String()` 代替 `Path` | 会把 query 打进来，日志噪 |
| 期望 IN 里有 body | 本中间件不读 body（读 body 只能读一次，需 TeeReader） |

---

## 10. 验收

启动服务后本地请求，观察终端：

```powershell
# 终端应出现 IN/OUT
Invoke-RestMethod http://localhost:8080/health
# [IN ] GET /health
# [OUT] GET /health -> 200 X-Cache=

$r = Invoke-RestMethod -Uri http://localhost:8080/api/links -Method POST `
  -ContentType "application/json" -Body '{"url":"https://example.com"}'
$code = $r.code

curl.exe -s -o NUL "http://localhost:8080/api/links/$code"
curl.exe -s -o NUL "http://localhost:8080/api/links/$code"
# 第二次 OUT 应 X-Cache= HIT
```

| 检查 | 期望 |
| --- | --- |
| 每个请求两行 | IN + OUT |
| POST 创建 | OUT `201`，X-Cache 空 |
| 两次 GET JSON | 先 MISS 后 HIT |
| GET /:code | OUT `302`，带 X-Cache |

---

## 口述题

1. `c.Next()` 返回时，说明下游哪些代码已经执行完了？
2. 为什么 `c.Writer.Status()` 必须在 `c.Next()` 之后读？
3. `POST /api/links` 的 OUT 行里 `X-Cache` 为什么通常是空的？
4. `gin.Recovery()` 和 `middleware.Logger()` 的注册顺序对请求处理有什么影响？
5. 若要在 OUT 行增加「耗时」，应把计时变量放在 `Next()` 的哪一侧？
