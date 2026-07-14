# 缓存、Cookie 与会话机制

> 本章目标：分清浏览器缓存、服务端缓存和登录状态，理解 Cookie、Session、JWT 以及 CORS 的真实工作方式。
>
> 前置：[04 · HTTP 协议](./04-HTTP协议深入.md)、[05 · HTTPS 与 TLS](./05-HTTPS与TLS加密.md)　下一章：[07 · 复习与面试总表](./07-面试专题与知识点总表.md)

---

## 1. 先记住五个结论

1. **浏览器 HTTP 缓存保存的是响应副本；Redis 缓存通常保存的是服务端业务数据。两者不是同一个缓存。**
2. **Cookie 是浏览器保存并按规则自动携带的小段数据。Session 是服务端状态，Cookie 常只保存 session ID。**
3. **JWT 是一种自包含的令牌格式，不等于加密，也不天然比 Session 更安全。**
4. **CORS 是浏览器对前端 JavaScript 的跨源读取限制；curl 和服务端请求不会因为 CORS 被浏览器拦截。**
5. **认证回答“你是谁”，授权回答“你能做什么”。**

---

## 2. 为什么这一章最容易混

一次登录后的接口请求，可能同时出现：

```text
浏览器缓存        → 可能不请求服务器，直接使用旧响应
Cookie           → 浏览器自动放进请求头
Session          → 服务端根据 session ID 查登录状态
JWT              → 客户端放在 Authorization 头中
Redis            → 服务端可能存 Session 或业务缓存
CORS             → 浏览器决定前端 JavaScript 能否读取响应
HTTPS            → 保护这些内容在网络传输中不被窃听和篡改
```

它们处于不同位置、解决不同问题。学习时不要只背名词，要先问：

> 这个数据存在哪里？谁自动发送？谁验证？解决性能还是身份问题？

---

## 3. 先分清三类“缓存”

### 3.1 浏览器 HTTP 缓存

位置：用户浏览器。

常缓存：

- CSS、JavaScript；
- 图片、字体；
- 某些允许缓存的 GET 响应。

目的：减少网络请求和下载，加快页面加载。

由 HTTP Header 控制，例如 `Cache-Control`、`ETag`。

### 3.2 CDN 缓存

位置：离用户较近的边缘节点。

目的：让很多用户从边缘节点获取相同内容，减少源站压力和跨地域延迟。

CDN 是否缓存取决于响应头、CDN 规则、请求方法、Cookie 等因素。

### 3.3 服务端业务缓存

位置：服务器内存、Redis 等。

例如短链跳转：

```text
short_code: abc123 → https://example.com/long/path
```

目的：减少数据库查询和业务计算。

服务端缓存对客户端通常是透明的。客户端只看到 HTTP 响应，不知道数据来自 Redis 还是 MySQL。

### 对比

| 类型 | 在哪里 | 由谁控制 | 常见对象 |
|---|---|---|---|
| 浏览器缓存 | 用户设备 | HTTP Header + 浏览器 | 静态资源、GET 响应 |
| CDN 缓存 | 边缘节点 | HTTP Header + CDN 配置 | 公共静态内容 |
| Redis 业务缓存 | 服务端 | 后端代码 | 热点业务数据、Session |

---

## 4. HTTP 强缓存

服务器返回：

```http
Cache-Control: public, max-age=3600
```

表示响应可以被公共缓存保存 3600 秒。在未过期期间，浏览器可能直接使用缓存，不向服务器验证。

这叫强缓存。

### 常见指令

#### max-age

```http
Cache-Control: max-age=3600
```

响应从生成后可新鲜使用 3600 秒。

#### public

允许浏览器和共享缓存保存。

#### private

只允许用户自己的私有缓存保存，不应被共享 CDN 缓存。

#### no-store

```http
Cache-Control: no-store
```

表示不要存储响应。敏感接口、登录响应或需要每次获取最新内容的接口常使用。

#### no-cache

名称容易误解。它通常不是“完全不缓存”，而是：

> 可以保存，但每次使用前必须向服务器验证是否仍有效。

#### immutable

告诉浏览器在新鲜期内资源不会改变，适合文件名带内容哈希的静态资源：

```text
app.a8f31c.js
```

内容变化后文件名也变化，因此旧文件可以长时间缓存。

---

## 5. 协商缓存与 304

缓存过期后，客户端不一定重新下载完整内容，可以询问服务器：

> 我手里的版本还有效吗？

### ETag

首次响应：

```http
HTTP/1.1 200 OK
ETag: "v1-a8f31c"
Content-Type: application/javascript

...完整内容...
```

下次请求：

```http
If-None-Match: "v1-a8f31c"
```

内容未变化时服务器返回：

```http
HTTP/1.1 304 Not Modified
ETag: "v1-a8f31c"
```

客户端继续使用本地副本，省去响应 Body 传输。

### Last-Modified

首次响应：

```http
Last-Modified: Tue, 14 Jul 2026 08:00:00 GMT
```

后续请求：

```http
If-Modified-Since: Tue, 14 Jul 2026 08:00:00 GMT
```

未修改则可返回 304。

ETag 通常比秒级修改时间更精确，但生成和比较策略由服务端决定。

### 304 是不是错误

不是。304 是缓存验证结果，告诉客户端内容没有变化。它通常没有正常实体 body。

---

## 6. API 应该怎样缓存

不能简单规定“所有 GET 都缓存”。要看数据性质。

### 适合公共长缓存

- 带内容哈希的 CSS、JS、图片；
- 版本固定的公开文档；
- 所有用户看到都相同的公共资源。

### 适合私有或短缓存

- 当前用户自己的资料；
- 短时间不变的列表；
- 有明确失效策略的查询。

### 通常 no-store

- 登录响应；
- 包含 token 或敏感个人信息的响应；
- 必须每次得到最新状态的操作结果。

如果响应会因 Authorization、Cookie 或 Accept-Encoding 不同而变化，缓存还需要正确使用 `Vary`，避免把一个用户的响应错误返回给另一个用户。

例如：

```http
Vary: Accept-Encoding
```

---

## 7. Cookie 到底是什么

Cookie 是浏览器保存的一小段键值数据。服务器通过响应头设置：

```http
Set-Cookie: session_id=abc123; Path=/; HttpOnly; Secure; SameSite=Lax
```

浏览器以后访问匹配的域名和路径时，自动发送：

```http
Cookie: session_id=abc123
```

### 关键特征

- Cookie 由浏览器管理；
- 浏览器按 Domain、Path、Secure、SameSite 等规则决定是否携带；
- Cookie 每次随匹配请求发送，不能放太大；
- Cookie 内容在客户端，用户可以查看或修改，不能盲目信任；
- 敏感 Cookie 必须通过 HTTPS。

### Cookie 常见属性

#### Domain

控制哪些域名可收到 Cookie。省略时通常是更严格的 host-only Cookie。

不要把敏感 Cookie 的 Domain 放得过宽，否则更多子域会共享它。

#### Path

```text
Path=/api
```

表示请求路径匹配 `/api` 时才发送。Path 主要是发送范围规则，不是安全权限边界。

#### Max-Age / Expires

控制持久时间。没有持久属性的 Cookie 通常是会话 Cookie，浏览器会话结束后可能删除。

#### Secure

只通过 HTTPS 发送。

#### HttpOnly

阻止前端 JavaScript 通过 `document.cookie` 直接读取，降低部分 XSS 窃取 Cookie 的风险。

它不能阻止浏览器自动携带 Cookie，也不能修复 XSS 本身。

#### SameSite

控制跨站请求时是否自动携带 Cookie：

- Strict：最严格，跨站通常不带；
- Lax：一些顶级导航可带，常作为实用默认；
- None：允许跨站，但必须同时 Secure。

SameSite 能降低 CSRF 风险，但仍需要根据业务设计完整防护。

---

## 8. Session 登录怎样工作

Session 方案中，登录状态主要保存在服务端。

### 登录

```text
1. 客户端 POST /login 提交账号密码
2. 服务端验证成功
3. 服务端生成随机 session ID
4. 服务端保存 session ID → user ID、过期时间等
5. 响应 Set-Cookie: session_id=...
```

### 后续请求

```text
1. 浏览器自动发送 Cookie: session_id=...
2. 服务端读取 session ID
3. 到内存、数据库或 Redis 查询 Session
4. 得到当前用户身份
5. 再进行权限判断
```

```mermaid
sequenceDiagram
    participant B as 浏览器
    participant API as Go API
    participant R as Session 存储/Redis
    B->>API: POST /login 账号密码
    API->>API: 验证密码
    API->>R: 保存 session_id → user_id
    API-->>B: Set-Cookie: session_id=...; HttpOnly; Secure
    B->>API: GET /profile + Cookie
    API->>R: 查询 session_id
    R-->>API: user_id=42
    API-->>B: 200 + 用户资料
```

### Session ID 应具备什么性质

- 足够随机，难以猜测；
- 登录后重新生成，防止会话固定攻击；
- 有过期时间；
- 登出、改密码或风险事件时可撤销；
- 不直接使用递增用户 ID 作为 session ID。

### Redis 在这里的角色

多实例 Go 服务不能只把 Session 放在某一台机器内存里，否则下次请求被负载均衡到另一台就找不到。

Redis 可以作为共享 Session 存储：

```text
session:abc123 → { user_id: 42, expires_at: ... }
```

这里 Redis 保存的是服务端登录状态，不是浏览器 HTTP 缓存。

---

## 9. JWT 登录怎样工作

JWT 通常由三部分组成：

```text
header.payload.signature
```

### Header

描述令牌类型和签名算法。

### Payload

保存声明，例如：

```json
{
  "sub": "42",
  "exp": 1784025600,
  "role": "user"
}
```

### Signature

服务端对前两部分签名，防止客户端擅自修改 payload 后仍被接受。

### 重要：JWT 通常不是加密

Header 和 Payload 只是 Base64URL 编码，拿到 token 的人通常能解码查看内容。

不要放：

- 密码；
- 身份证号；
- 私钥；
- 不希望客户端看到的敏感信息。

### 请求怎样携带

常见方式：

```http
Authorization: Bearer <access_token>
```

服务端：

1. 解析 token；
2. 检查签名算法；
3. 验证签名；
4. 检查过期时间、签发者、受众等；
5. 取得用户 ID；
6. 进行权限判断。

### JWT 的优点

- 服务端可不按 access token 保存完整 Session；
- 多服务之间可按统一规则验证；
- 适合 API 和分布式场景中的某些需求。

### JWT 的难点

- 签发后在过期前不容易立即撤销；
- 权限变化后旧 token 可能仍携带旧声明；
- token 过长，每次请求都要发送；
- 密钥轮换、算法校验、刷新机制更复杂；
- token 被窃取后仍可被冒用。

因此 JWT 不是“免数据库、绝对无状态、天然更安全”的万能方案。

---

## 10. Access Token 与 Refresh Token

常见设计：

- Access Token：寿命较短，用于访问 API；
- Refresh Token：寿命较长，只用于换取新 Access Token。

好处是 Access Token 泄露后的有效窗口较短，同时用户不必频繁登录。

Refresh Token 风险更高，应：

- 安全存储；
- 只发送到刷新端点；
- 支持轮换；
- 检测重复使用；
- 支持撤销；
- 通过 HTTPS 传输。

是否需要这套机制取决于项目。简单服务不必为了“看起来高级”引入过度复杂的 token 体系。

---

## 11. Session 与 JWT 怎么选

| 对比 | Session | JWT Access Token |
|---|---|---|
| 主要状态位置 | 服务端 | token 内包含部分声明 |
| 撤销 | 删除服务端 Session 较直接 | 常需黑名单、版本号或等过期 |
| 多实例共享 | 需要共享 Session 存储或粘性会话 | 各实例可验证签名 |
| 每次请求大小 | Cookie 中 ID 较小 | token 可能较大 |
| 权限变化 | 服务端查询可快速生效 | 旧 token 可能延迟到过期 |
| 复杂点 | Session 存储和扩展 | 密钥、刷新、撤销、轮换 |

选择原则：

- 浏览器传统网站、强调即时撤销：Session 很自然；
- 多客户端 API、已有成熟 token 基础设施：JWT 可能合适；
- 两者也可以组合，例如短期 Access Token + 服务端保存 Refresh Session。

安全性取决于完整实现，不取决于名字。

---

## 12. 认证与授权

### 认证 Authentication

回答：

> 你是谁？

例如验证 Session 或 JWT 后得到 `user_id=42`。

### 授权 Authorization

回答：

> 你能不能操作这个资源？

例如：

```text
用户 42 请求删除短链 100
  → 查询短链 100 是否属于用户 42
  → 属于才允许删除
```

有有效 JWT 不代表能访问所有数据。后端查询必须包含资源归属条件，不能只相信客户端传来的 `user_id`。

---

## 13. 同源策略是什么

浏览器把下面三项组合视为源：

```text
协议 + 主机 + 端口
```

例如：

```text
https://app.example.com:443
```

下列变化都会造成不同源：

- `http` 与 `https` 不同；
- `app.example.com` 与 `api.example.com` 不同；
- 8080 与 3000 不同。

浏览器的同源策略限制一个源的 JavaScript 随意读取另一个源的响应，防止恶意网站直接读取用户在其他网站的敏感数据。

注意：跨源请求可能已经发到服务器，限制重点常在“浏览器是否允许 JavaScript 读取响应”。

---

## 14. CORS 怎样放行跨源请求

CORS 是服务器通过响应头告诉浏览器：

> 哪些源、方法和 Header 可以读取我的响应。

### 简单请求

浏览器从：

```text
http://localhost:5173
```

请求：

```text
http://localhost:8080/api/users
```

服务端响应：

```http
Access-Control-Allow-Origin: http://localhost:5173
```

浏览器看到允许当前源，才把响应交给前端 JavaScript。

### 预检请求

如果请求包含 JSON POST、自定义 Header、Authorization 等，浏览器可能先发送 OPTIONS：

```http
OPTIONS /api/users HTTP/1.1
Origin: http://localhost:5173
Access-Control-Request-Method: POST
Access-Control-Request-Headers: content-type, authorization
```

服务端允许时返回：

```http
Access-Control-Allow-Origin: http://localhost:5173
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Max-Age: 600
```

浏览器随后才发送真实 POST。

### 为什么 curl 没有 CORS 错误

CORS 是浏览器安全策略。`curl`、Go 服务端、Postman 不受浏览器同源策略控制。

因此：

```text
curl 成功 + 浏览器失败
```

可能正是 CORS 配置问题，而不是接口没启动。

---

## 15. 带 Cookie 的跨源请求

前端需要显式允许携带凭证：

```javascript
fetch("https://api.example.com/profile", {
  credentials: "include"
})
```

服务端需要：

```http
Access-Control-Allow-Origin: https://app.example.com
Access-Control-Allow-Credentials: true
```

此时 `Access-Control-Allow-Origin` 不能使用 `*`，必须返回明确允许的源。

同时 Cookie 的 SameSite、Secure、Domain 规则也必须允许发送。CORS 放行不等于 Cookie 一定会带上。

---

## 16. CORS 不是什么

CORS 不是：

- 身份认证；
- 防火墙；
- 防止 curl 调接口；
- 防止其他服务器调用接口；
- 防止 CSRF 的完整方案；
- 防止 XSS 的方案。

攻击者完全可以用自己的服务器或 curl 请求公开 API。真正的安全仍依赖认证、授权、限流和业务校验。

不要配置：

```text
任意 Origin + 允许 Credentials
```

应维护明确的允许源列表，并正确处理 `Vary: Origin` 等缓存问题。

---

## 17. CSRF 与 XSS 为什么常和 Cookie/JWT 一起出现

### CSRF

浏览器会自动携带匹配 Cookie。恶意网站可能诱导用户浏览器向已登录网站发请求。

防护包括：

- SameSite Cookie；
- CSRF Token；
- 检查 Origin/Referer；
- 敏感操作使用正确方法并再次确认；
- 不使用 GET 修改数据。

### XSS

攻击者把恶意 JavaScript 注入你的页面。脚本可能读取页面数据、发起用户权限内操作，或窃取能被 JS 访问的 token。

防护包括：

- 输出编码和模板自动转义；
- 避免危险地插入 HTML；
- Content Security Policy；
- HttpOnly Cookie；
- 依赖和输入安全治理。

把 JWT 放到 localStorage 可以避免自动随请求携带造成的某些 CSRF 模式，但更容易被 XSS 脚本读取。不存在只靠换存储位置就解决所有安全问题的方案。

---

## 18. Go / Gin 中的落地点

### 设置 Cookie

标准库示例：

```go
http.SetCookie(w, &http.Cookie{
	Name:     "session_id",
	Value:    sessionID,
	Path:     "/",
	MaxAge:   3600,
	HttpOnly: true,
	Secure:   true,
	SameSite: http.SameSiteLaxMode,
})
```

本地纯 HTTP 调试时 Secure Cookie 不会被浏览器发送。生产必须使用 HTTPS，开发环境可以使用受控的环境配置，而不是忘记生产安全属性。

### 读取 Cookie

```go
cookie, err := r.Cookie("session_id")
if err != nil {
	http.Error(w, "unauthorized", http.StatusUnauthorized)
	return
}
```

拿到 session ID 后还必须查询并验证 Session，不能只判断 Cookie 存在。

### 缓存 Header

敏感响应：

```go
w.Header().Set("Cache-Control", "no-store")
```

静态版本化资源可以由 Nginx/CDN 设置长期缓存，不必所有策略都塞进业务 Handler。

### CORS 中间件

中间件需要：

- 校验 Origin 是否在允许列表；
- 设置允许方法和 Header；
- 正确处理 OPTIONS；
- 带凭证时返回明确 Origin；
- 不盲目反射任意 Origin；
- 在实际路由和错误响应上都正确添加 Header。

使用成熟库也必须理解这些规则，不能只复制 `AllowAllOrigins: true`。

---

## 19. 登录接口排错顺序

### 登录成功但后续还是 401

检查：

1. 登录响应是否真的有 `Set-Cookie` 或 token；
2. 浏览器是否保存 Cookie；
3. Cookie 的 Domain、Path、Secure、SameSite 是否匹配；
4. 后续请求是否携带 Cookie 或 Authorization；
5. Session 是否存在且未过期；
6. JWT 是否过期、签名是否正确；
7. 代理是否移除了 Header；
8. 服务端时间是否正确。

### curl 成功，浏览器报 CORS

检查浏览器 Network：

1. 是否先发 OPTIONS；
2. OPTIONS 状态码是否 2xx；
3. Allow-Origin 是否精确匹配前端 Origin；
4. Allow-Methods 是否包含实际方法；
5. Allow-Headers 是否包含 Content-Type、Authorization；
6. 带 Cookie 时是否 Allow-Credentials 为 true；
7. 是否错误地同时使用 `*` 与 credentials；
8. 重定向或错误响应是否丢失 CORS Header。

### 页面一直显示旧数据

检查：

1. 浏览器是否命中强缓存；
2. Network 是否显示 304；
3. Service Worker 是否缓存；
4. CDN 是否缓存旧响应；
5. 服务端 Redis 是否缓存旧数据；
6. 数据库是否实际更新；
7. Cache-Control、ETag、缓存 key 和失效逻辑是否正确。

不要一看到“旧数据”就只清 Redis，缓存可能在多个层次。

---

## 20. 常见误解

### “Cookie 就是 Session”

不对。Cookie 在客户端；Session 通常在服务端。Cookie 可以只保存 Session ID。

### “JWT 比 Session 一定更先进”

不对。两者权衡不同，JWT 的撤销和刷新机制可能更复杂。

### “JWT payload 看不懂，所以是加密的”

不对。它通常只是编码，可以解码。签名负责防篡改。

### “no-cache 就是完全不缓存”

不准确。它通常允许存储，但使用前必须重新验证。

### “CORS 配好了，接口就安全了”

不对。CORS 只控制浏览器前端读取，不能替代认证和授权。

### “Redis 缓存和 304 是一回事”

不对。一个发生在服务端数据层，一个是 HTTP 客户端缓存验证。

---

## 21. 本章自测

1. 浏览器缓存、CDN 缓存和 Redis 缓存分别在哪里？
2. `no-store` 与 `no-cache` 有什么区别？
3. ETag 和 304 怎样配合？
4. Cookie 和 Session 分别存在哪里？
5. Session ID 为什么必须随机且可撤销？
6. JWT 的签名解决什么问题，不解决什么问题？
7. 认证和授权有什么区别？
8. 什么叫同源？
9. 为什么 JSON POST 常触发 OPTIONS 预检？
10. 为什么 curl 成功不能证明浏览器不会有 CORS 问题？

---

## 22. 学完标准

- [ ] 能分清浏览器、CDN、Redis 三类缓存；
- [ ] 能解释 Cache-Control、ETag 和 304；
- [ ] 能讲清 Cookie 属性与发送规则；
- [ ] 能完整描述 Session 登录链路；
- [ ] 知道 JWT 的结构、优点和撤销难点；
- [ ] 能区分认证和授权；
- [ ] 能解释同源、CORS 简单请求和预检；
- [ ] 知道 CORS、CSRF、XSS 不是同一个问题；
- [ ] 能排查登录后 401、浏览器 CORS 和旧缓存问题。

下一章：[07 · 面试专题与知识点总表](./07-面试专题与知识点总表.md)
