# HTTP 协议深入

> **文件编码**：UTF-8。  
> **定位**：⭐ **Go 05 net/http 最重要前置**——从报文结构到 curl 实操，把 [02 TCP 与 UDP](./02-TCP与UDP.md) 的「TCP 接通电话」升级到能读 HTTP 报文、选对方法、解释状态码，并在 [Go 05 标准库与 HTTP 基础](../../后端学习/Go/05-Go标准库与HTTP基础.md) 用 `net/http` 写服务。  
> **前置**：[02 TCP 与 UDP](./02-TCP与UDP.md)；建议 [03 IP 地址与 DNS 解析](./03-IP地址与DNS解析.md)  
> **下一章**：[05 HTTPS 与 TLS 加密](./05-HTTPS与TLS加密.md)

---

## 0. 读前导读（零基础也能跟上）

### 0.1 用一句话弄懂本章

**HTTP（HyperText Transfer Protocol，超文本传输协议）** = 客户端与服务器之间「写信的格式」——信封上写清楚：什么操作（GET/POST）、找谁（Host）、内容类型（Content-Type）、是否带登录凭证（Authorization）。

**核心类比：HTTP = 寄信 / 快递单格式**

| 信件部分 | HTTP 对应 |
|----------|-----------|
| 信封上的「收/寄、地址」 | **请求行** `GET /api/users HTTP/1.1` + **Host** 头 |
| 邮票、挂号、保价标签 | **Header**（Content-Type、Authorization、Cookie…） |
| 信纸正文 | **Body**（JSON、表单） |
| 邮局回执「已签收 / 查无此人」 | **状态码** 200 / 404 / 500 |
| 挂号信必须走可靠通道 | 跑在 **TCP** 上（02 章） |

### 0.1.1 本章还会出现的词

| 词 | 是什么 |
|----|--------|
| **REST** | Representational State Transfer，用 URL 表资源、HTTP 方法表操作的 API 设计风格 |
| **CRUD** | Create / Read / Update / Delete，增删改查 |
| **幂等** | 同一操作执行 1 次和 N 次，资源最终状态相同（GET/PUT/DELETE 通常幂等，POST 创建通常不幂等） |
| **Bearer** | HTTP 认证方案名，写在 `Authorization: Bearer <token>` 里，意为「持票人凭此 token 通行」 |
| **Keep-Alive** | HTTP 长连接，同一 TCP 连接上连续发多个请求，省掉反复握手 |
| **队头阻塞** | 前一个响应没传完，后面的请求必须排队等（HTTP/1.1 常见问题） |
| **httpbin.org** | 公开的 HTTP 测试网站，会把你发的请求原样回显，适合练 curl |

---

| 前置 | 章节 | 必须？ |
|------|------|--------|
| TCP、端口、三次握手 | [02 章](./02-TCP与UDP.md) | ✅ |
| DNS、localhost | [03 章](./03-IP地址与DNS解析.md) | 建议 |
| Go 基础、struct | [Go 00～03](../../后端学习/Go/) | ✅ |
| goroutine、context | [Go 04 并发](../../后端学习/Go/04-Go并发编程goroutine与channel.md) | 建议 |

**最低门槛**：知道「客户端发请求、服务器回 JSON」；本章以 **curl** 为主调试工具，**浏览器 DevTools Network** 作选修对照（见 §10.1）。

### 0.3 本章知识地图（学完后应能勾选全部 ☐→☑）

```text
☐ 能画出 HTTP 请求/响应结构（起始行、头、空行、Body）
☐ 能解释 Host、Content-Type、Authorization、Cookie
☐ 能区分 GET/POST/PUT/DELETE 语义与何时使用
☐ 能解释 200/201/301/302/400/401/403/404/500/502/503
☐ 能对比 GET 与 POST 的语义差异与 curl 写法
☐ 能说明 HTTP/1.1 Keep-Alive 与 HTTP/2 多路复用（一句话）
☐ 能设计简单 RESTful URI（名词 + HTTP 方法）
☐ 能用 curl -v 完成 GET/POST/PUT/DELETE + Bearer Token
☐ 能把 HTTP 概念对应到 Go Handler、Request、ResponseWriter
☐ 闭卷自测 ≥ 8/10
```

### 0.4 建议学习时长与节奏

| 阶段 | 内容 | 时间 |
|------|------|------|
| 报文与方法 | §1～§4 | 50 分钟 |
| 状态码 + GET vs POST | §5～§6 | 40 分钟 |
| Keep-Alive + 版本 + REST | §7～§9 | 30 分钟 |
| curl 实操 + Go 映射 | §10～§11 | 50 分钟 |
| 自测复盘 | 报错表、FAQ、闭卷、费曼 | 40 分钟 |

**建议**：每节 curl 示例立刻跟做；暂无 Go 服务可用 `https://httpbin.org`（§10）。

### 0.5 学完本章你能做什么（可验证的具体动作）

1. 读 `curl -v` 输出，区分 `>`（发出）与 `<`（收到）行。
2. 设计「用户 CRUD」REST 接口表（Method + URI），URI 不含动词。
3. 用 curl 模拟 401（缺 token）与 404（路径错）。
4. 向朋友 3 分钟讲清：HTTP 无状态、401 vs 403、GET vs POST。
5. 打开 Go 05，看到 `http.ResponseWriter` 时不困惑「这是在写响应哪一部分」。

---

## 本章衔接

| 章节 | 基础 | 本章补什么 |
|------|------|------------|
| [02 TCP](./02-TCP与UDP.md) | 端口、三次握手 | 应用层 HTTP 报文在 TCP 之上 |
| [Go 04](../../后端学习/Go/04-Go并发编程goroutine与channel.md) | goroutine、context | 每请求通常一个 goroutine |
| **本章** | — | 报文、方法、状态码、curl、REST |
| [Go 05 net/http](../../后端学习/Go/05-Go标准库与HTTP基础.md) | — | 标准库读写本章概念 |
| [05 HTTPS](./05-HTTPS与TLS加密.md) | 明文 HTTP | TLS、443 |

```mermaid
flowchart TB
    subgraph 应用层["应用层（本章）"]
        HTTP[HTTP 报文 / 方法 / 状态码]
        REST[RESTful 设计]
    end
    subgraph 传输层["传输层（02 章）"]
        TCP[TCP :8080]
    end
    subgraph Go["Go 05"]
        H[Handler / Request / ResponseWriter]
    end
    HTTP --> TCP
    H --> HTTP
    REST --> HTTP
```

---

## 1. HTTP 是什么

HTTP 定义**客户端与服务器交换消息的格式和语义**。TCP 负责「电话接通」（02 章）；HTTP 负责「电话里说什么」。

| 特征 | 含义 | 对 Go 后端 |
|------|------|------------|
| **无状态** | 服务器不记上次请求 | 登录态靠 Cookie / Token 每次自带 |
| **请求-响应** | 一问一答 | Handler 处理 Request，写 Response |
| **可扩展 Header** | 元数据在头部 | Content-Type、Authorization |
| **基于 TCP** | 默认 80/443 | `ListenAndServe(":8080", ...)` |

**无状态的代价**：登录、权限需 Cookie+Session 或 JWT（放 `Authorization: Bearer xxx`）。

```mermaid
sequenceDiagram
    participant C as 客户端 curl / http.Client
    participant S as Go http.Server
    participant H as Handler

    Note over C,S: TCP 三次握手（02 章）已完成
    C->>S: POST /api/login HTTP/1.1<br/>Host: localhost:8080<br/>Content-Type: application/json<br/><br/>{"username":"admin"}
    S->>H: 解析为 *http.Request
    H->>H: 读 Body、校验
    H->>S: WriteHeader(200) + JSON
    S->>C: HTTP/1.1 200 OK + Body
    Note over C,S: Keep-Alive 可复用 TCP
```

---

## 2. HTTP 报文结构

消息三部分（`\r\n` 换行）：

```text
起始行 → 头部（键值对）→ 空行 → Body（可选）
```

### 2.1 请求报文

```http
POST /api/login HTTP/1.1
Host: localhost:8080
Content-Type: application/json
Content-Length: 42
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9

{"username":"admin","password":"123456"}
```

| 部分 | 说明 |
|------|------|
| **请求行** | 方法 + 路径 + 版本，如 `GET /api/users?page=1 HTTP/1.1` |
| **Host** | HTTP/1.1 **必填**，虚拟主机路由 |
| **Header** | Content-Type、Authorization、Cookie 等 |
| **Body** | POST/PUT 常见，JSON 字符串 |

### 2.2 响应报文

```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8
Content-Length: 58

{"id":1,"name":"张三"}
```

| 部分 | 说明 |
|------|------|
| **状态行** | 版本 + **状态码** + 原因短语 |
| **Header** | Content-Type、Set-Cookie 等 |
| **Body** | JSON 等业务数据 |

### 2.3 逐行读对照（改错会怎样）

| 行 | 改错后果 |
|----|----------|
| 请求行方法/路径错 | 404 / 405 |
| 缺 Host | 400 或路由错误 |
| 缺 Content-Type（有 JSON Body） | 415 或解析失败 |
| Content-Length 不对 | 挂起或截断 |

---

## 3. 常见 Header

### 3.1 请求头（Go 后端必会）

| Header | 典型值 | 作用 | Go 读取 |
|--------|--------|------|---------|
| **Host** | `localhost:8080` | 目标主机 | `r.Host` |
| **Content-Type** | `application/json` | Body 格式 | `r.Header.Get("Content-Type")` |
| **Authorization** | `Bearer <JWT>` | 认证 | `r.Header.Get("Authorization")` |
| **Cookie** | `sessionid=abc` | 会话 | `r.Cookie("sessionid")` |

### 3.2 响应头（Go 写入）

| Header | 作用 | Go 写入 |
|--------|------|---------|
| **Content-Type** | 响应体类型 | `w.Header().Set("Content-Type", "application/json")` |
| **Set-Cookie** | 写 Cookie | `http.SetCookie(w, cookie)` |
| **Location** | 301/302 目标 | `http.Redirect(w, r, url, code)` |

### 3.3 Content-Type 常见值

| 值 | 场景 |
|----|------|
| `application/json` | REST API（**最常用**） |
| `application/x-www-form-urlencoded` | 传统表单 |
| `multipart/form-data` | 文件上传 |

**Authorization vs Cookie**：API 前后端分离常用 JWT + `Authorization: Bearer ...`；传统 Web 用 Cookie + Session。

---

## 4. HTTP 方法

| 方法 | 语义 | Body | 幂等 | 何时用 |
|------|------|------|------|--------|
| **GET** | 获取 | 否 | ✅ | 查询列表、详情 |
| **POST** | 创建/提交 | 是 | ❌ | 登录、创建用户 |
| **PUT** | 全量替换 | 是 | ✅ | 更新整个对象 |
| **DELETE** | 删除 | 可选 | ✅ | 删除资源 |
| **PATCH** | 部分更新 | 是 | 视实现 | 改备注、改状态字段 |
| **HEAD** | 只要响应头 | 否 | ✅ | 探活、查 Content-Length，不发 body |

> **幂等**：执行 1 次与 N 次，资源最终状态相同。

Go 1.22+ 路由示例（Go 05 详讲）：

```go
mux.HandleFunc("GET /api/users/{id}", getUserHandler)
mux.HandleFunc("POST /api/users", createUserHandler)
mux.HandleFunc("PUT /api/users/{id}", updateUserHandler)
mux.HandleFunc("DELETE /api/users/{id}", deleteUserHandler)
```

---

## 5. 状态码（本章必背）

| 码 | 含义 | 场景 |
|----|------|------|
| **200** | 成功 | GET/POST 成功 |
| **201** | 已创建 | POST 创建资源 |
| **301** | 永久重定向 | 域名更换 |
| **302** | 临时重定向 | 未登录跳登录 |
| **400** | 参数错误 | JSON 非法、校验失败 |
| **401** | 未认证 | 无 token / token 无效 |
| **403** | 无权限 | 有 token 但角色不够 |
| **404** | 资源不存在 | 路径未注册、id 不存在 |
| **500** | 服务端内部错误 | panic、数据库失败 |
| **502** | 网关上游挂了 | Nginx 后 Go 进程挂 |
| **503** | 服务不可用 | 维护、过载 |

**401 vs 403（必背）**：401「你是谁？」；403「我知道你是谁，但不能干这个」。

```mermaid
flowchart TD
    Start[收到请求] --> Auth{需登录?}
    Auth -->|无 token| R401[401]
    Auth -->|无权限| R403[403]
    Auth -->|通过| Route{路径/方法对?}
    Route -->|否| R404[404]
    Route -->|是| Valid{参数合法?}
    Valid -->|否| R400[400]
    Valid -->|是| Biz[业务逻辑]
    Biz -->|成功| R200[200/201]
    Biz -->|异常| R500[500]
```

---

## 6. GET vs POST

| 维度 | GET | POST |
|------|-----|------|
| 语义 | **安全**查询 | **非安全**提交 |
| 参数 | Query `?page=1` | Body（JSON） |
| 幂等 | ✅ | ❌ |
| 重复提交 | 一般无害 | 可能重复下单 |

**curl 对比**：

```bash
curl -v "http://localhost:8080/api/users?page=1"

curl -v -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"李四\",\"age\":22}"
```

**误解**：POST 不比 GET 安全——明文 HTTP 下都不安全，靠 **HTTPS**（05 章）。

---

## 7. Keep-Alive 与 HTTP 版本

**HTTP/1.1** 默认 **Keep-Alive**：同一 TCP 复用多个请求，减少握手。

```http
Connection: keep-alive
```

缺点：同连接响应须按序返回（**队头阻塞**）。

**HTTP/2（一句话）**：在一条 TCP 上用二进制帧实现**多路复用（multiplexing）**——多请求/响应并行，并压缩重复 Header。

| | HTTP/1.1 | HTTP/2 |
|--|----------|--------|
| 格式 | 文本，curl 可读 | 二进制帧 |
| 多请求 | 串行 | 多路复用 |

Go 开发默认 HTTP/1.1；生产 HTTPS 常由 Nginx 协商 h2。HTTP/3（QUIC）了解即可。

---

## 8. RESTful 概念（简要）

REST = 用 **URI 表资源**，用 **HTTP 方法表操作**。

| ❌ 非 REST | ✅ RESTful |
|-----------|-----------|
| `POST /api/getUserById` | `GET /api/users/1` |
| `POST /api/deleteUser` | `DELETE /api/users/1` |

| 操作 | 方法 | URI |
|------|------|-----|
| 列表 | GET | `/api/users` |
| 详情 | GET | `/api/users/{id}` |
| 创建 | POST | `/api/users` |
| 更新 | PUT | `/api/users/{id}` |
| 删除 | DELETE | `/api/users/{id}` |

**两种错误风格**：纯 REST 用真实 404；国内常见 HTTP 200 + `{"code":1,"message":"..."}`。团队应统一约定。

---

## 9. curl 实操（`-v` 必读）

| 符号 | 含义 |
|------|------|
| `>` | 客户端**发出**的请求行/头 |
| `<` | 服务器**返回**的状态行/头 |

### 9.1 公共练习：httpbin.org

**GET**：

```bash
curl -v "https://httpbin.org/get?name=go"
```

预期：`< HTTP/1.1 200 OK`，Body 含 `"args":{"name":"go"}`。

**POST JSON**：

```bash
curl -v -X POST https://httpbin.org/post \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"123456\"}"
```

预期：Body 里 `"json"` 字段含提交内容。

**PUT / DELETE**：

```bash
curl -v -X PUT https://httpbin.org/put -H "Content-Type: application/json" -d "{\"name\":\"王五\"}"
curl -v -X DELETE https://httpbin.org/delete
```

**Authorization**：

```bash
curl -v https://httpbin.org/bearer -H "Authorization: Bearer my-test-token"
```

预期：Body 含 `"token": "my-test-token"`。

**只看响应头**：`curl -I https://httpbin.org/get`

### 9.2 本地 Go 服务（Go 05 之后）

```bash
curl -v http://localhost:8080/health

curl -v -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"李四\",\"age\":22}"

curl -v http://localhost:8080/api/users/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 9.3 错误模拟

```bash
curl -v http://localhost:8080/api/not-exist          # 404
curl -v -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" -d "{}"       # 400（视校验）
```

### 9.4 curl 步骤表

| 步骤 | 命令 | 预期 | 若不对 |
|------|------|------|--------|
| 1 | `curl -v https://httpbin.org/get` | 200 + JSON | 查网络 |
| 2 | POST + Content-Type | Body 回显 json | 缺 Header |
| 3 | Authorization Bearer | token 回显 | 头名拼写 |
| 4 | 本地 `:8080/health` | JSON | Connection refused → 未启动 |

---

## 10. 映射到 Go net/http

| HTTP 概念 | Go API |
|-----------|--------|
| 服务器 | `http.ListenAndServe` |
| **Handler** | `func(w ResponseWriter, r *Request)` |
| **Request** | `r.Method`、`r.URL`、`r.Header`、`r.Body` |
| **ResponseWriter** | `WriteHeader`、`Header().Set`、`Write` |
| 读 Body | `json.NewDecoder(r.Body).Decode(&v)` |
| 写 JSON | `w.Header().Set(...)` → `WriteHeader(200)` → `Encode(data)` |
| 客户端 | `http.Client`（Go 05 §5） |

**最小 Handler（预习 Go 05）**：

```go
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(User{ID: 1, Name: "张三"})
}
```

对照 §2：`WriteHeader(200)` → 状态行；`Header().Set` → 响应头；`Encode` → Body。

### 10.1 多栈对照：浏览器 fetch / Java Spring / Python Flask

| 名词 | 是什么 |
|------|--------|
| **fetch** | 浏览器里用 JavaScript 发 HTTP 请求的 API |
| **Spring Boot** | Java 最常用的 Web 框架，用注解注册路由 |
| **Flask** | Python 的轻量 Web 框架 |

HTTP 报文格式与语言无关；下表是**同一 REST 语义**在不同栈的写法：

| 操作 | curl | 浏览器 fetch | Spring Boot | Flask |
|------|------|--------------|-------------|-------|
| GET 列表 | `curl /api/users` | `fetch('/api/users')` | `@GetMapping("/api/users")` | `@app.get("/api/users")` |
| POST 创建 | `curl -X POST -d '{...}'` | `fetch(url, {method:'POST', body:...})` | `@PostMapping` + `@RequestBody` | `request.get_json()` |
| 设响应头 | `-H` 发 / 读 `<` 行 | `response.headers.get(...)` | `ResponseEntity` / `HttpHeaders` | `response.headers[...]` |
| 401 | 读状态行 | `response.status === 401` | `@ExceptionHandler` / `ResponseStatus` | `abort(401)` |

**fetch 读响应示例**（前端联调时对照 curl）：

```javascript
const res = await fetch('http://localhost:8080/api/users', {
  headers: { 'Authorization': 'Bearer ' + token }
});
console.log(res.status);        // 对应 curl 的 < HTTP/1.1 200
const data = await res.json();  // 对应 curl body
```

学完 Go 05 后，用 curl 和 fetch **各打一遍同一接口**，能加深对「报文 vs 框架封装」的理解。前后端联调选修可参考 [Vue 08-Axios 联调](../../前端学习/Vue/08-Axios网络请求与前后端联调.md)。

```mermaid
sequenceDiagram
    participant curl as curl
    participant S as http.Server
    participant H as Handler

    curl->>S: GET /api/users/1 HTTP/1.1
    S->>H: r.Method=GET, r.URL.Path=/api/users/1
    H->>S: WriteHeader(200) + JSON
    S->>curl: HTTP/1.1 200 OK + Body
```

---

## 11. 常见报错与排查表

| 现象 | 可能原因 | 怎么查 | 解决 |
|------|----------|--------|------|
| Connection refused | 服务未启动、端口错 | `curl -v localhost:8080` | `go run .` |
| 404 Not Found | URL 与路由不一致 | 对照 HandleFunc | 改路径 |
| 405 Method Not Allowed | 方法不对 | 看 r.Method | 改 curl -X |
| 400 Bad Request | JSON 非法、缺字段 | Handler 校验 | 修正 -d |
| 401 Unauthorized | 缺 Authorization | curl -v 看 `>` | 加 Bearer 头 |
| 403 Forbidden | 有 token 无权限 | 查角色 | 换账号 |
| 415 Unsupported Media Type | 未设 Content-Type | 请求头 | 加 application/json |
| 500 Internal Server Error | panic、DB 错误 | Go 终端日志 | 修代码 + Recovery |
| 502 Bad Gateway | Nginx 后进程挂 | 进程状态 | 重启服务 |
| 503 Service Unavailable | 过载、维护 | 监控 | 扩容/熔断 |
| 响应乱码 | 未设 Content-Type | 响应头 | Header().Set |
| EOF / Body 读不完 | 客户端断开 | 服务端日志 | 查 Content-Length |

---

## 12. FAQ（≥10）

**Q1：HTTP 和 HTTPS？** HTTP 明文；HTTPS = HTTP + TLS。见 [05 章](./05-HTTPS与TLS加密.md)。

**Q2：HTTP 和 TCP？** TCP 是电话线（02 章）；HTTP 是通话格式（本章）。

**Q3：GET 能带 Body？** 不推荐；很多服务器忽略。参数放 Query。

**Q4：幂等？** GET/PUT/DELETE 通常幂等；POST 创建不幂等。

**Q5：401 vs 403？** 401 未认证；403 已认证无权限。

**Q6：301 vs 302？** 301 永久；302 临时。

**Q7：200 vs 201？** 200 一般成功；201 表示创建了资源。

**Q8：500/502/503？** 500 程序内部错；502 网关连不上后端；503 主动不可用。

**Q9：Host 为何必填？** 一机多域名，Host 指定虚拟主机。

**Q10：Cookie vs Authorization？** API 常用 JWT+Authorization；传统 Web 用 Cookie+Session。

**Q11：WriteHeader 顺序？** 先 Header().Set，再 WriteHeader，最后 Write/Encode。

**Q12：Go 要配 HTTP/2 吗？** 开发默认 1.1；生产 HTTPS 常由 Nginx 协商 h2。

---

## 13. 练习与参考答案

### 基础题

1. 画出请求报文四部分并各写一行示例。
2. 401 与 403 区别，各举 Go API 场景。
3. 为何 POST /api/orders 不幂等？取消订单用什么方法？
4. 列出 10 个必背状态码及含义。
5. curl -v 中 `>` 与 `<` 各是什么？

### 进阶题

6. 设计商品 REST 接口表（列表/详情/创建/更新/删除）。
7. Keep-Alive 与 HTTP/2 多路复用各一句话。
8. Go Handler 返回 404 怎么写？JSON 前先设什么 Header？
9. 写两步 curl：POST 登录 + Bearer 访问受保护接口。
10. 对比「HTTP 404」与「HTTP 200 + code:1」两种风格。

### 参考答案（节选）

**2.** 401：访问 /api/admin 无 Authorization；403：普通用户 token 访问 admin 接口。

**3.** 每次 POST 可能多订单；取消用 PUT/PATCH 改状态或 DELETE。

**6.**

| 操作 | Method | URI |
|------|--------|-----|
| 列表 | GET | /api/products |
| 详情 | GET | /api/products/{id} |
| 创建 | POST | /api/products |
| 更新 | PUT | /api/products/{id} |
| 删除 | DELETE | /api/products/{id} |

**9.**

```bash
curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"123456\"}"
curl -v http://localhost:8080/api/users/1 -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 14. 学完标准

- [ ] 画出 HTTP 请求/响应结构（起始行、头、Body）
- [ ] 解释 Host、Content-Type、Authorization、Cookie
- [ ] 区分 GET/POST/PUT/DELETE，知道何时用
- [ ] 背诵 200/201/301/302/400/401/403/404/500/502/503
- [ ] 说明 GET vs POST 差异；Keep-Alive 与 HTTP/2 多路复用各一句
- [ ] 设计 RESTful URI；curl -v 完成 CRUD + Bearer
- [ ] 对应 Handler、Request、ResponseWriter
- [ ] 闭卷 ≥8/10；费曼 3 分钟

---

## 15. 闭卷自测（10 题）

**概念（6）**

1. 寄信类比：请求行、Header、Body 各对应什么？
2. HTTP 无状态？登录态靠什么？
3. GET vs POST：语义、幂等、参数位置？
4. 401 vs 403？Go 场景各一例。
5. Keep-Alive 解决什么？HTTP/2 多做了什么？
6. 为何反对 POST /api/getUserById？应改什么？

**动手（2）**

7. 写 POST 登录 curl（含 Content-Type + JSON Body）。
8. curl -v 如何区分请求头与响应头？

**综合（2）**

9. Go 返回 404 调用什么？写 JSON 前先设什么 Header？
10. 从 curl GET 到 Go 返回 JSON，经过哪些层？

### 参考答案

1. 地址；标签；正文。 2. 服务器不记上次；Cookie/JWT。 3. GET 查/幂等/Query；POST 提交/非幂等/Body。 4. 401 无 token；403 无权限。 5. 复用 TCP；多 Stream 并行。 6. GET /api/users/{id}。 7. 见 §13 题 9。 8. `>` 发出；`<` 收到。 9. WriteHeader(404)；Content-Type。 10. TCP→HTTP 报文→Server→Handler→ResponseWriter→curl。

---

## 16. 费曼检验（3 分钟）

对照是否讲到：

1. HTTP = 信纸格式：方法、Host、Body、状态码。
2. HTTP 跑在 TCP 上（02 章管接通）。
3. 无状态 → 每次自带 token（Authorization/Cookie）。
4. 401 vs 403：没证件 vs 有证件不让进。
5. Go 05：Handler 收 Request、写 ResponseWriter。

---

## 17. 下一章预告

- [05 HTTPS 与 TLS 加密](./05-HTTPS与TLS加密.md)：生产为何必须 HTTPS
- [Go 05 net/http](../../后端学习/Go/05-Go标准库与HTTP基础.md)：ListenAndServe、中间件、Client

**建议**：本章 → Go 05 写服务 → 05 HTTPS → Go 06 Gin。

---

## 附录：速查卡

```text
GET→查  POST→建  PUT→改  DELETE→删
200成功 201创建 301永久 302临时
400参数 401未登录 403无权限 404无资源
500后端 502网关 503不可用
```

| 主题 | 文档 |
|------|------|
| TCP | [02](./02-TCP与UDP.md) |
| DNS | [03](./03-IP地址与DNS解析.md) |
| HTTPS | [05](./05-HTTPS与TLS加密.md) |
| Go net/http | [Go 05](../../后端学习/Go/05-Go标准库与HTTP基础.md) |

*Go 05 前置标准：§0、报错表 12 条、FAQ 12 条、闭卷 10 题、费曼、Mermaid 序列图、curl 实操、net/http 映射。*
