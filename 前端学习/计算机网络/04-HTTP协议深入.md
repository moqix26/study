# HTTP 协议深入

> 本章目标：能逐行读懂 HTTP 请求和响应，并把方法、路径、Header、Body、状态码与 Go Handler 对应起来。
>
> 前置：[02 · TCP 与 UDP](./02-TCP与UDP.md)、[03 · IP 地址与 DNS](./03-IP地址与DNS解析.md)　下一章：[05 · HTTPS 与 TLS](./05-HTTPS与TLS加密.md)

---

## 1. 先记住三个结论

1. **HTTP 是请求—响应协议。客户端先发请求，服务器再返回响应。**
2. **一条 HTTP 消息由起始行、Header、空行和可选 Body 组成。**
3. **状态码说明处理结果，响应体提供具体数据或错误信息。状态码不是业务数据的替代品。**

---

## 2. HTTP 在整条链路中的位置

访问：

```text
http://localhost:8080/health
```

大致顺序是：

```text
解析主机和端口
  → 建立 TCP 连接
  → 发送 HTTP 请求
  → Go 服务处理
  → 返回 HTTP 响应
```

访问 HTTPS 时，会在 TCP 与 HTTP 之间增加 TLS：

```text
DNS → TCP → TLS → HTTP → 业务处理
```

HTTP 负责描述：

- 想对哪个资源做什么；
- 客户端携带了哪些附加信息；
- 是否有请求体；
- 服务器处理结果是什么；
- 返回的内容是什么格式。

HTTP 不负责 IP 路由和 TCP 重传。

---

## 3. 逐行读懂一个 GET 请求

运行：

```powershell
curl.exe -v http://localhost:8080/health
```

可能看到：

```http
GET /health HTTP/1.1
Host: localhost:8080
User-Agent: curl/8.x
Accept: */*

```

### 第一行：请求行

```http
GET /health HTTP/1.1
```

包含三部分：

| 部分 | 示例 | 含义 |
|---|---|---|
| 方法 | `GET` | 客户端想做什么 |
| 请求目标 | `/health` | 想访问哪个路径和查询参数 |
| 版本 | `HTTP/1.1` | 使用哪版 HTTP 规则 |

### Host

```http
Host: localhost:8080
```

同一 IP 可能托管多个域名，服务器需要 Host 判断客户端想访问哪个站点。HTTP/1.1 请求通常必须携带 Host。

### User-Agent

```http
User-Agent: curl/8.x
```

表示客户端类型。服务端可以记录，但不应把它当成可靠身份认证，因为客户端能自行伪造。

### Accept

```http
Accept: */*
```

表示客户端愿意接收的响应媒体类型。`*/*` 表示都可以。

### 最后的空行

空行表示 Header 结束。如果有 Body，Body 从空行后开始。

---

## 4. 逐行读懂一个响应

服务端可能返回：

```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 16
Date: Tue, 14 Jul 2026 10:00:00 GMT

{"status":"ok"}
```

### 状态行

```http
HTTP/1.1 200 OK
```

包含版本、状态码和原因短语。

程序主要依据数字状态码判断结果，`OK` 主要方便人阅读。

### Content-Type

```http
Content-Type: application/json
```

告诉客户端 Body 的媒体类型。JSON API 应返回正确的 `application/json`，否则客户端可能按错误格式处理。

### Content-Length

表示 Body 的字节长度，帮助接收方确定消息边界。HTTP 也可以使用分块传输等其他方式表达长度。

### Body

```json
{"status":"ok"}
```

Body 承载实际内容。并不是所有响应都有 Body，例如 204、HEAD 响应和 304 通常不带正常实体 body。

---

## 5. POST JSON 请求怎样组成

发送：

```powershell
curl.exe -v -X POST http://localhost:8080/api/users `
  -H "Content-Type: application/json" `
  -d '{"name":"Alice"}'
```

请求大致是：

```http
POST /api/users HTTP/1.1
Host: localhost:8080
Content-Type: application/json
Content-Length: 16

{"name":"Alice"}
```

这里要同时理解：

- `POST` 表示创建或提交处理；
- `/api/users` 是目标资源集合；
- `Content-Type` 告诉服务端请求体是 JSON；
- Body 是真正提交的数据。

### Content-Type 错了会怎样

如果 Body 是 JSON，却写成：

```http
Content-Type: text/plain
```

某些框架会拒绝解析，或无法自动绑定结构体。常见响应是 400 或 415。

`Content-Type` 描述**你实际发送的 Body 是什么**，不是一句随便写的标签。

---

## 6. URL 的每一部分

```text
https://api.example.com:443/users/42?include=profile#bio
│       │                │  │        │               │
协议    主机              端口 路径     查询字符串        片段
```

### 路径参数

```text
/users/42
```

通常表示 ID 为 42 的用户。在 Gin 中可能定义：

```go
r.GET("/users/:id", handler)
```

### 查询参数

```text
/users?page=2&page_size=20
```

适合过滤、排序、分页或可选条件。

### 片段

```text
#bio
```

片段主要由浏览器在本地使用，通常不会发送到服务器。因此后端 Handler 一般看不到 `#bio`。

### URL 编码

空格、中文和保留字符不能直接随意出现在 URL 中，会进行百分号编码。例如空格常表示为 `%20`。

不要手工拼接用户输入形成 URL，使用标准库编码，避免歧义和安全问题。

---

## 7. HTTP 方法怎么选择

| 方法 | 常见含义 | 是否通常有 Body | 是否通常幂等 |
|---|---|---:|---:|
| GET | 查询资源 | 否 | 是 |
| POST | 创建资源或提交动作 | 是 | 否 |
| PUT | 整体替换指定资源 | 是 | 是 |
| PATCH | 部分更新资源 | 是 | 不一定 |
| DELETE | 删除资源 | 可选 | 是 |
| HEAD | 只获取响应头 | 否 | 是 |
| OPTIONS | 查询通信能力，CORS 预检常用 | 通常否 | 是 |

### 安全方法与幂等方法

**安全**表示按语义不应修改服务器状态，例如 GET、HEAD。

**幂等**表示同一个请求执行一次或多次，资源最终状态相同。

例如：

```text
PUT /users/42  把 name 设置为 Alice
```

重复执行，最终仍是 Alice，所以通常幂等。

```text
POST /orders  创建新订单
```

重复执行可能创建多个订单，所以通常不幂等。

幂等不等于响应完全相同，也不等于没有日志、计数等副作用。它关注目标资源的预期最终状态。

### GET 为什么不应修改数据

浏览器预取、缓存、搜索引擎爬虫和代理可能自动发 GET。如果 GET 会删除或支付，会造成严重风险。

---

## 8. 状态码怎样理解

状态码第一位表示类别：

| 范围 | 含义 |
|---|---|
| 1xx | 临时信息 |
| 2xx | 请求成功处理 |
| 3xx | 重定向或缓存相关 |
| 4xx | 客户端请求有问题 |
| 5xx | 服务器处理失败 |

### 2xx：成功

| 状态码 | 常见用法 |
|---|---|
| 200 OK | 查询、更新等成功，并返回内容 |
| 201 Created | 成功创建资源 |
| 202 Accepted | 已接收，稍后异步处理 |
| 204 No Content | 成功，但无响应体 |

创建资源时可以返回：

```http
HTTP/1.1 201 Created
Location: /api/users/42
Content-Type: application/json
```

### 3xx：重定向与缓存

| 状态码 | 常见用法 |
|---|---|
| 301 | 永久重定向，客户端和缓存可能长期记住 |
| 302 | 临时重定向，短链跳转常见 |
| 307 | 临时重定向，并明确保持原方法和 Body |
| 308 | 永久重定向，并明确保持原方法和 Body |
| 304 | 缓存仍有效，不返回正常实体 body |

短链常用 302，是因为目标地址可能调整，不希望客户端永久缓存映射。

### 4xx：客户端请求问题

| 状态码 | 含义 |
|---|---|
| 400 | 请求格式、字段或参数无效 |
| 401 | 未通过身份认证 |
| 403 | 已识别身份，但没有权限 |
| 404 | 路由或资源不存在 |
| 405 | 路径存在，但方法不允许 |
| 409 | 资源状态冲突，例如唯一键冲突 |
| 415 | 不支持请求体媒体类型 |
| 422 | 请求格式能解析，但业务校验不通过 |
| 429 | 请求过多，被限流 |

### 5xx：服务器问题

| 状态码 | 含义 |
|---|---|
| 500 | 未归类的服务端错误 |
| 502 | 网关从上游收到无效响应 |
| 503 | 服务暂时不可用 |
| 504 | 网关等待上游超时 |

不要把所有失败都返回 200，再在 JSON 里写 `success:false`。这样会破坏 HTTP 语义，也让监控、重试和客户端处理更困难。

---

## 9. 常见 Header 逐个理解

### Host

目标域名和可选端口。用于虚拟主机路由。

### Content-Type

描述当前消息 Body 的格式：

```text
application/json
text/html; charset=utf-8
application/x-www-form-urlencoded
multipart/form-data
```

### Accept

表示客户端希望接收什么类型：

```http
Accept: application/json
```

### Authorization

携带认证信息：

```http
Authorization: Bearer <token>
```

Bearer Token 必须通过 HTTPS 传输。任何拿到 token 的人通常都能以持有者身份使用它。

### Cookie 与 Set-Cookie

服务器通过响应头设置 Cookie：

```http
Set-Cookie: session_id=abc123; HttpOnly; Secure; SameSite=Lax
```

浏览器之后按规则自动发送：

```http
Cookie: session_id=abc123
```

### Cache-Control

控制缓存行为：

```http
Cache-Control: no-store
```

或：

```http
Cache-Control: public, max-age=3600
```

### Location

重定向目标或新资源地址：

```http
Location: https://example.com/new-path
```

### User-Agent

客户端标识，可用于兼容性分析和日志，但不能作为可信认证依据。

### X-Request-ID / Traceparent

用于追踪一次请求经过多个组件的日志。它们不是 HTTP 强制标准业务字段，但在工程化系统中非常有用。

---

## 10. Header 大小写与重复值

HTTP Header 名不区分大小写：

```text
Content-Type
content-type
CONTENT-TYPE
```

语义相同。Go 标准库会把常见 Header 名规范化显示。

有些 Header 可以出现多次或包含多个值。不要简单假设所有 Header 都只有一个字符串，也不要随意用逗号拼接 `Set-Cookie`。

---

## 11. 请求体怎样确定长度

HTTP/1.1 常见两种方式：

### Content-Length

```http
Content-Length: 16
```

表示随后读取固定字节数。

### Transfer-Encoding: chunked

数据分块传输，每块带自己的长度，最后用结束块表示完成。适合生成内容时不提前知道总长度。

应用层框架通常替你处理这些细节，但理解它能帮助排查代理、上传和响应截断问题。

---

## 12. REST 只是设计风格，不是新协议

REST API 仍然使用 HTTP。常见设计：

```text
GET    /users       查询用户列表
POST   /users       创建用户
GET    /users/42    查询用户 42
PATCH  /users/42    部分更新用户 42
DELETE /users/42    删除用户 42
```

建议：

- URL 用名词表示资源；
- 方法表示操作；
- 状态码表达结果；
- JSON 返回稳定字段；
- 错误响应有机器可识别的错误码和人类可读信息。

示例错误响应：

```json
{
  "code": "USER_NOT_FOUND",
  "message": "user does not exist",
  "request_id": "req_123"
}
```

HTTP 状态码可以是 404，业务 `code` 再细分具体原因。

---

## 13. HTTP 是无状态的是什么意思

HTTP 协议本身不会自动记住：

> 这次请求与上一次请求来自同一个已登录用户。

每个请求都应携带处理它所需的信息，或者携带能让服务端找到状态的凭证，例如：

- Cookie 中的 session ID；
- Authorization 中的 JWT；
- API key。

“HTTP 无状态”不等于应用不能有用户状态，而是状态需要通过额外机制维护。06 章会详细讲。

---

## 14. 长连接和连接复用

如果每个 HTTP 请求都重新建立 TCP，握手成本会很高。

HTTP/1.1 默认倾向于复用连接：

```text
建立一次 TCP
  → 请求 1 / 响应 1
  → 请求 2 / 响应 2
  → 请求 3 / 响应 3
```

连接复用能减少：

- TCP 握手；
- TLS 握手；
- 系统 socket 创建；
- 延迟和 CPU 消耗。

Go `http.Client` 和 `Transport` 会管理连接池。应复用 Client，而不是每个请求创建新的 Transport。

### HTTP/1.1 的限制

同一连接上的请求响应处理容易受到队头阻塞影响。浏览器常创建多条连接缓解。

### HTTP/2

在一条连接中多路复用多个流，并对 Header 压缩。多个请求可以并发进行。

但因为通常仍基于一条 TCP，底层丢包可能影响连接中的多个流。

### HTTP/3

基于 QUIC/UDP，把多路流和 TLS 1.3 等能力结合，减少传输层队头阻塞和建连延迟。

当前阶段先掌握 HTTP/1.1 报文和语义，版本演进理解到这里即可。

---

## 15. 映射到 Go net/http

下面这段代码展示请求与响应各字段在哪里：

```go
func createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "content type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var input struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id":   42,
		"name": input.Name,
	})
}
```

对应关系：

| HTTP 内容 | Go 中的位置 |
|---|---|
| 方法 | `r.Method` |
| URL | `r.URL` |
| Header | `r.Header` |
| Body | `r.Body` |
| 响应 Header | `w.Header().Set(...)` |
| 状态码 | `w.WriteHeader(...)` |
| 响应 Body | `w.Write` 或 JSON Encoder |

### WriteHeader 的顺序

应先设置 Header，再写状态码，再写 Body：

```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
json.NewEncoder(w).Encode(result)
```

如果先写 Body，Go 会自动发送默认状态码 200，后续再调用 `WriteHeader(201)` 已经太晚。

---

## 16. curl 实操清单

### GET

```powershell
curl.exe -v http://localhost:8080/health
```

### 带查询参数

```powershell
curl.exe -v "http://localhost:8080/api/users?page=1&page_size=20"
```

### POST JSON

```powershell
curl.exe -v http://localhost:8080/api/users `
  -H "Content-Type: application/json" `
  -d '{"name":"Alice"}'
```

`-d` 已经会让 curl 使用 POST，一般不需要额外写 `-X POST`。

### 带 Authorization

```powershell
curl.exe -v http://localhost:8080/api/profile `
  -H "Authorization: Bearer test-token"
```

### 只看响应头

```powershell
curl.exe -I https://example.com
```

### 跟随重定向

```powershell
curl.exe -v -L http://example.com
```

不加 `-L` 时可以观察第一次 301/302 和 Location；加上后会继续访问目标地址。

### 保存响应体

```powershell
curl.exe -D headers.txt -o body.json http://localhost:8080/api/users
```

---

## 17. 常见问题怎样判断

### 400 Bad Request

可能原因：

- JSON 语法错误；
- 必填字段缺失；
- 路径或查询参数格式错误；
- 请求体超过限制；
- 服务端无法解析请求。

### 401 与 403

- 401：没有有效身份，例如 token 缺失或过期；
- 403：身份有效，但没有访问该资源的权限。

### 404 与 405

- 404：路径或资源不存在；
- 405：路径存在，但这个方法不允许。

### 502 与 504

常见于 Nginx、API 网关或负载均衡：

- 502：上游连接失败、响应无效或进程崩溃；
- 504：网关等待上游响应超时。

此时浏览器连接到网关可能成功，但网关到 Go 服务的上游链路有问题。

---

## 18. 常见误解

### “POST 比 GET 安全”

不对。POST Body 不是加密，只是通常不显示在 URL。真正的传输保密依赖 HTTPS。

### “状态码 200 表示业务一定正确”

不对。它只表示服务器选择用 200 表达成功。若代码错误地所有情况都返回 200，客户端仍会误判。

### “HTTP 是无连接协议”

说法容易误导。HTTP 是应用层请求—响应协议；HTTP/1.1、HTTP/2 通常运行在 TCP 连接上，而且连接可以复用。

### “Header 想写什么都行”

可以定义自有 Header，但标准 Header 有明确语义。认证、缓存、内容类型等不能随意解释。

### “GET 不能带 Body 是协议绝对禁止”

规范和实现对 GET Body 的互操作支持很差，语义也不明确。工程上不要依赖 GET Body，查询条件放 URL，复杂查询可设计 POST 查询接口。

---

## 19. 本章自测

1. HTTP 请求由哪四部分组成？
2. Content-Type 和 Accept 分别描述什么？
3. 201、204、302、400、401、403、404、409、429、500 各适合什么场景？
4. 为什么 POST 不天然安全？
5. 幂等是什么意思？POST 创建为什么通常不幂等？
6. 得到 404 时，TCP 连接是否大概率已经成功？
7. Cookie 与 Authorization 在 HTTP 报文的哪里？
8. Go 中为什么要在写 Body 前设置 Header 和状态码？

---

## 20. 学完标准

- [ ] 能逐行解释 GET 和 POST 的原始报文；
- [ ] 能正确选择常见 HTTP 方法；
- [ ] 能区分主要 2xx、3xx、4xx、5xx；
- [ ] 能解释 Content-Type、Accept、Authorization、Cookie、Location；
- [ ] 能理解 REST、无状态和幂等；
- [ ] 能使用 curl 构造 JSON、Header 和查询参数；
- [ ] 能把 HTTP 字段映射到 Go `net/http`；
- [ ] 能按状态码定位客户端、网关或服务端问题。

下一章：[05 · HTTPS 与 TLS](./05-HTTPS与TLS加密.md)
