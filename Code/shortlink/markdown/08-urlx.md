# 08 · urlx URL 校验精读

> 主线请跟 [`../study.md`](../study.md)；本文是精读加餐。

> 对应源码：`internal/pkg/urlx/urlx.go`  
> 目标：把「用户提交的 URL 怎么洗、什么算合法、什么会被拒」讲透。

---

## 0. 这个包在整体里干什么

创建短链时，客户端 POST 的 `{"url":"..."}` 必须先变成**可信的长链字符串**再入库。脏数据若进 MySQL，后续 302 `Location` 可能跳到 javascript:、file: 等危险目标。

```text
POST /api/links  body.url
  → handler.CreateLink（ShouldBindJSON）
  → service.Create(rawURL)
  → urlx.Normalize(rawURL)  ← 本文
  → shortcode.Random → repo.Create
```

**上游**：`handler.CreateLink` 把 `req.URL` 原样交给 `service.Create`。  
**下游**：规范化后的 `longURL` 写入 `model.Link.LongURL`，并出现在响应 JSON 的 `long_url` 字段；跳转时 `Resolve` 读出的也是这条串。

本包**只做校验与规范化**，不访问网络（不 HEAD 请求、不 DNS 解析）。

---

## 1. 完整源码（逐块对照）

```go
package urlx

import (
	"errors"
	"net/url"
	"strings"
)

// Normalize 校验并规范化用户提交的 URL。
func Normalize(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("url required")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", errors.New("invalid url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("url must be http or https")
	}
	return u.String(), nil
}
```

---

## 2. `package urlx` 与 `import`

```go
package urlx
```

- `url` + `x`（extension）表示 URL 相关小工具，路径 `internal/pkg/urlx`。
- 与 `shortcode` 一样：无外部 IO 依赖，纯函数，易单测。

### import 表


| 包 | 本文件里干什么 |
| --- | --- |
| `errors` | `errors.New` 构造固定文案的业务错误 |
| `net/url` | `url.Parse` 解析 URL 为结构体 `*url.URL` |
| `strings` | `TrimSpace` 去掉首尾空白 |

---

## 3. `Normalize(raw string) (string, error)` 逐步拆解

### 3.1 签名


| 参数/返回 | 类型 | 含义 |
| --- | --- | --- |
| `raw` | `string` | 用户原始输入（可能含首尾空格、无 scheme 等） |
| 成功返回 | `string` | 规范化后的 URL 字符串，可安全入库 |
| 失败返回 | `error` | 三种固定消息（见下表） |

### 3.2 第一步：`TrimSpace`

```go
raw = strings.TrimSpace(raw)
```

| 输入示例 | 处理后 |
| --- | --- |
| `"  https://a.com  "` | `"https://a.com"` |
| `"\n\thttps://a.com"` | `"https://a.com"` |
| `"   "` | `""`（进入下一步报错） |

**为什么先 Trim？**  
前端/脚本常多打空格；不 trim 会导致 `url.Parse` 把空格算进 URL 或整串判无效，体验差。

### 3.3 第二步：空串检查

```go
if raw == "" {
	return "", errors.New("url required")
}
```

| 触发条件 | 错误文案 | HTTP 层映射 |
| --- | --- | --- |
| trim 后长度为 0 | `url required` | `400 Bad Request`（handler 按 `err.Error()` 匹配） |

**为什么不用 `invalid url`？**  
区分「没填」和「填了但格式不对」，便于客户端提示。

### 3.4 第三步：`url.Parse` + Scheme/Host

```go
u, err := url.Parse(raw)
if err != nil || u.Scheme == "" || u.Host == "" {
	return "", errors.New("invalid url")
}
```

| 名字 | 含义 |
| --- | --- |
| `u` | 解析结果 `*url.URL` |
| `u.Scheme` | 协议，如 `https`、`http`；**必须非空** |
| `u.Host` | 主机名（可含端口），如 `www.example.com:443`；**必须非空** |
| `err != nil` | 解析器认为字符串不是合法 URL |

**会被拒的示例**（`invalid url`）：


| 输入 | 原因 |
| --- | --- |
| `not-a-url` | 无 scheme，常无 host |
| `://missing` | parse 失败或 host 空 |
| `https://` | host 空 |
| `/only/path` | 相对路径，无 scheme/host |

**注意**：`https://www.example.com/path?q=1` 合法；`url.Parse` 会拆出 Path、RawQuery 等，最后 `u.String()` 再拼回去。

**与旧版单文件差异**：历史 `normalizeURL` **没有**限制 scheme 只能是 http/https；分层版多了白名单（见下一节），更安全。

### 3.5 第四步：只允许 http / https

```go
if u.Scheme != "http" && u.Scheme != "https" {
	return "", errors.New("url must be http or https")
}
```

| scheme | 结果 |
| --- | --- |
| `http` | 通过 |
| `https` | 通过 |
| `javascript` | 拒绝（防 XSS 跳转） |
| `file` | 拒绝 |
| `ftp` | 拒绝 |
| `HTTP`（大写） | `url.Parse` 后 scheme 通常已小写；若为大写需看 Parse 行为，一般仍可比 |

**为什么只允许 http/https？**  
短链核心是「浏览器可跟的 Web 跳转」。其它 scheme 在 302 场景有安全风险或客户端行为不一致。V1 明确收窄范围，错误信息也固定，handler 可映射 400。

### 3.6 第五步：返回 `u.String()`

```go
return u.String(), nil
```

| 调用 | 含义 |
| --- | --- |
| `u.String()` | 按 RFC 3986 规则把结构体编码回字符串 |

**规范化效果举例**：

- 可能统一某些转义（如空格、`%` 编码）
- 保留 path、query、fragment（若 Parse 能解析）

入库的是**这一串**，不是用户原始 `raw`（若二者有细微差别，以 `u.String()` 为准）。

---

## 4. 错误文案与 HTTP 映射（上下游）


```text
urlx.Normalize 错误
       ↓
service.Create → return nil, err（不区分类型，原样往上）
       ↓
handler.CreateLink:
  msg := err.Error()
  if msg == "url required" || msg == "invalid url" || msg == "url must be http or https"
    → 400 + {"error": msg}
  else
    → 500
```

| `err.Error()` | HTTP | 谁产生 |
| --- | --- | --- |
| `url required` | 400 | 空 URL |
| `invalid url` | 400 | Parse/缺 scheme/host |
| `url must be http or https` | 400 | scheme 白名单 |
| 其它（如 DB 错） | 500 | 非 urlx |

**设计点**：urlx 用**字符串常量**错误，handler 用**字符串相等**判断 400。生产项目可改为 `errors.Is` + 哨兵错误类型；V1 保持简单。

---

## 5. 本包刻意不做的事


| 不做 | 为什么 |
| --- | --- |
| 不自动补 `https://` | 避免把 `example.com` 静默变成合法；强迫用户写全 scheme，规则清晰 |
| 不请求目标站 | 创建要快；可达性由用户负责 |
| 不校验域名黑名单 | V1 范围外 |
| 不限制 URL 长度 | 长度由 `model.Link.LongURL` `size:2048` 与 MySQL 列约束；超长 INSERT 会失败 → 500 |

---

## 6. 常见坑


| 坑 | 现象 | 对策 |
| --- | --- | --- |
| POST 缺 `url` 字段 | `""` → `url required` | 客户端检查 JSON |
| 只写 `www.baidu.com` 无 scheme | `invalid url` | 必须 `https://www.baidu.com` |
| `javascript:alert(1)` | `url must be http or https` | 白名单生效 |
| handler 新增 urlx 错误文案却忘记加 400 分支 | 校验失败变 500 | 同步改 `CreateLink` 的 if |
| 以为 Normalize 会跟随重定向 | 不会 | 存的就是用户给的 URL |
| Windows 下误用 `\` | 通常 `invalid url` | 用正斜杠 URL |

---

## 7. 验收

### 7.1 合法创建

```powershell
Invoke-RestMethod -Uri http://localhost:8080/api/links -Method POST `
  -ContentType "application/json" -Body '{"url":"https://www.bilibili.com"}'
# 期望 201，long_url 为规范化 https URL
```

### 7.2 非法用例（期望 400）


```powershell
# 空
Invoke-RestMethod -Uri http://localhost:8080/api/links -Method POST `
  -ContentType "application/json" -Body '{"url":""}' 
# error: url required

# 无 scheme
Invoke-RestMethod -Uri http://localhost:8080/api/links -Method POST `
  -ContentType "application/json" -Body '{"url":"www.example.com"}'
# error: invalid url

# 危险 scheme
Invoke-RestMethod -Uri http://localhost:8080/api/links -Method POST `
  -ContentType "application/json" -Body '{"url":"javascript:alert(1)"}'
# error: url must be http or https
```

（PowerShell 对 400 可能抛异常，可用 `try/catch` 或 `curl.exe`：）

```powershell
curl.exe -i -X POST http://localhost:8080/api/links -H "Content-Type: application/json" -d "{\"url\":\"\"}"
```

### 7.3 期望表


| 请求 body | 状态码 | `error` 字段 |
| --- | --- | --- |
| 合法 https URL | 201 | 无 |
| `""` | 400 | `url required` |
| 无 scheme | 400 | `invalid url` |
| `javascript:...` | 400 | `url must be http or https` |

---

## 8. 与 Resolve / 跳转的关系

`Normalize` **只在创建时**执行一次。之后 `GET /:code` 和 `GET /api/links/:code` 只读已存的 `long_url`，不再过 urlx。

若 DB 里被手工改成非法 URL，跳转仍会 `302` 到该串——**真源是 MySQL**，urlx 是创建入口的闸门。

---

## 口述题

1. `Normalize` 里为什么要先 `TrimSpace` 再判空？
2. `url.Parse` 之后为什么还要单独检查 `Scheme` 和 `Host`？各举一个非法例子。
3. 为什么拒绝 `javascript:` scheme？如果放过会发生什么？
4. 创建失败时，handler 怎么区分 400 和 500？三种 urlx 错误分别是什么？
5. `return u.String()` 和用户原始输入 `raw` 一定完全相同吗？什么情况下可能不同？
