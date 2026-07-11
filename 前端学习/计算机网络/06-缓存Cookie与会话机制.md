# 缓存、Cookie 与会话机制

> **文件编码**：UTF-8。  
> **定位**：计算机网络系列 **06 章**——讲清 HTTP 缓存（`Cache-Control` / `ETag` / 304）、CDN、Cookie / Session / JWT、以及 **CORS 完整排查**；实操以 **curl** 为主，浏览器 DevTools / fetch / Spring / Flask 作多栈对照。  
> **前置**：[05-HTTPS与TLS加密](./05-HTTPS与TLS加密.md)、[04-HTTP协议深入](./04-HTTP协议深入.md)、[Go 00～04](../../后端学习/Go/00-学习路线图与说明.md)。  
> **交叉**：[Go 06 Gin CORS 中间件](../../后端学习/Go/06-Gin框架核心与中间件.md) §6、[Go 09 JWT 认证](../../后端学习/Go/09-JWT认证与用户体系.md)、[Go 10～11 短链项目](../../后端学习/Go/10-短链服务项目实战上.md)。

---

## 0. 读前导读（零基础也能跟上）

### 0.1 用一句话弄懂本章

**一句话**：浏览器和 CDN 会**记住**看过的响应（**缓存**）；服务器用 **Cookie** 让浏览器自动出示「会员手环」，前后端分离 API 常用 **JWT + Authorization 头**；**CORS** 是浏览器保安——**curl 没有保安，所以 curl 从不报 CORS 错**。

**核心类比表**

| 概念 | 生活类比 | Go 后端关联 |
|------|----------|-------------|
| **强缓存** | 冰箱便当，保质期内**不打电话问饭店** | 你在响应头设 `max-age`，客户端直接用 |
| **协商缓存（304）** | 过期后**打电话问**「还是昨天那批吗？」 | 你用 `ETag` 判断，未变则 304 无 body |
| **CDN** | 连锁便利店在各小区设分店 | 静态资源走 CDN 域名，API 默认 `no-store` |
| **Cookie** | 健身房**手环**，进门自动识别 | `Set-Cookie` 响应头；`net/http` / Gin 设置 |
| **Session** | 前台**登记簿**（服务端 Redis） | 手环只存 `sessionId`，状态在 Redis |
| **JWT Token** | 自带信息的**电子票** | `Authorization: Bearer`，见 [Go 09](../../后端学习/Go/09-JWT认证与用户体系.md) |
| **CORS** | 浏览器保安：别的域 JS **不许读**响应 | Gin `cors` 中间件声明 `Allow-Origin` |
| **curl** | 没有保安的**快递员** | 直连 `localhost:8080`，跨域也不拦 |

### 0.2 你需要提前知道什么

| 前置 | 对应章节 | 本章是否必须 |
|------|----------|--------------|
| HTTP 状态码 200/304/401 | [04 章](./04-HTTP协议深入.md) | ✅ |
| HTTPS 与 Cookie `Secure` | [05 章](./05-HTTPS与TLS加密.md) | ✅ |
| Go `net/http`、Gin 路由 | [Go 05](../../后端学习/Go/05-Go标准库与HTTP基础.md)、[Go 06](../../后端学习/Go/06-Gin框架核心与中间件.md) | ✅ 建议 |
| JWT 签发与中间件 | [Go 09](../../后端学习/Go/09-JWT认证与用户体系.md) | 本章讲概念，09 跟做代码 |
| 浏览器 / Vue / Axios | 无要求 | CORS 面试必会，联调时再学 |

### 0.3 本章知识地图（学完后应能勾选全部 ☐→☑）

```text
☐ 能口述强缓存 vs 协商缓存，304 有没有 body
☐ 能解释 Cache-Control 至少 5 个指令（max-age/no-cache/no-store/private/immutable）
☐ 能说明 API 响应 vs 静态资源（hash JS）的缓存策略差异
☐ 能在 Go/Gin 里为不同路由设置合适的缓存响应头
☐ 能对比 Session+Cookie、JWT Header、JWT HttpOnly Cookie（Go 后端选型）
☐ 能解释为什么 JWT 放 Header 不会自动出现——要客户端手动带
☐ 能说明 curl 不报 CORS、浏览器报 CORS 的根本原因
☐ 能排查 OPTIONS 预检失败，对照 Gin cors 配置修复
☐ 能用 curl -I / curl -v 查看 Cache-Control、ETag、304
☐ 能独立验收短链登录链路（§12.7 表全过）
```

### 0.4 建议学习时长与节奏

| 阶段 | 内容 | 时间 |
|------|------|------|
| HTTP 缓存决策 | §1～§5 | 55 分钟 |
| CDN + Cookie/Session/JWT | §6～§9 | 60 分钟 |
| CORS 原理 + Gin 配置 | §10～§11 | 50 分钟 |
| curl 实操 + 自测复盘 | §12～章末 | 45 分钟 |

### 0.5 学完本章你能做什么（可验证的具体动作）

1. `curl -I http://localhost:8080/health` 读出 `Cache-Control`，并解释该不该缓存该 API。
2. 对带 `ETag` 的静态文件第二次请求，用 `curl -H "If-None-Match: ..."` 打出 **304**。
3. `curl -X POST http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d '{"username":"u","password":"p"}'` 拿到 token，再 `curl -H "Authorization: Bearer <token>"` 调受保护接口。
4. 向面试官解释：「Postman/curl 正常、浏览器 Console 报 CORS」——原因与修法。
5. 完成 §12.7 短链认证验收表 5 步全绿。

---

## 本章衔接

01～05 章你已建立 **OSI → IP/TCP → HTTP/HTTPS → DNS** 的大图景；[Go 05](../../后端学习/Go/05-Go标准库与HTTP基础.md) 教你用 `net/http` 写 Handler。本章把 **缓存响应头、会话凭证、跨域** 三块补全——这是写 Gin API、接前端、过面试的必备拼图。

| 上章产出 | 本章怎么用 |
|----------|------------|
| HTTP 请求/响应、状态码 | 理解 200/304/401 在缓存与登录里的含义 |
| HTTPS 加密通道 | Cookie `Secure`、JWT 传输安全 |
| DNS 解析 | CDN 边缘节点、域名与 Cookie `Domain` |
| Go Handler 返回 JSON | 给 API 加 `Cache-Control: no-store`；静态文件加 `ETag` |

**学完本章你应该能**：用 curl 判断缓存行为；为短链 API 与静态资源配置合理响应头；说清楚 Session/Cookie/JWT 选型；独立排查 CORS 预检失败并在 Gin 里修复。

```mermaid
flowchart TB
    subgraph 上章["01～05 + Go 05"]
        A[HTTP/HTTPS]
        B[net/http Handler]
    end
    subgraph 本章["06 章"]
        C[强缓存 / 协商缓存 / 304]
        D[CDN 边缘缓存]
        E[Cookie / Session / JWT]
        F[CORS + OPTIONS 预检]
        G[短链认证链路]
    end
    subgraph 下章["Go 06 / 09 / 10"]
        H[Gin CORS 中间件]
        I[JWT AuthMiddleware]
        J[短链项目实战]
    end
    A --> C
    B --> E
    C --> H
    E --> I
    F --> H
    G --> J
```

---

## 1. 为什么 Go 后端必须懂 HTTP 缓存

### 1.1 缓存解决什么问题

浏览器、CDN、反向代理（Nginx）都会缓存 HTTP 响应：

- **减少重复下载**：同一份 `app.[hash].js`、同一张 logo 不必每次回源
- **降低延迟**：边缘节点或本地磁盘比跨城回源快几个数量级
- **减轻源站压力**：静态资源由 CDN 扛，Go 进程专注 API

### 1.2 后端设错缓存头的后果

| 现象 | 常见原因 | 谁该修 |
|------|----------|--------|
| 用户看到旧版前端 | `index.html` 被强缓存 | Nginx/CDN 配置 + 构建 hash |
| API 返回了别人的数据 | 带 `Authorization` 的响应被 CDN 缓存 | **Go/Nginx 设 `private, no-store`** |
| 改代码后 curl 仍旧 JSON | 中间层缓存了 GET 接口 | 检查 `Cache-Control` |
| 304 但业务数据过期 | 协商缓存命中，ETag 未随内容更新 | Handler 或文件服务 ETag 逻辑 |

**后端职责**：为 **API 响应** 显式声明「不要缓存」或「短时公共缓存」；为 **静态资源**（若由 Go `http.FileServer` 或 Nginx 提供）配合运维设长缓存 + hash 文件名。

### 1.3 缓存存在哪一层

```text
客户端请求
  → 浏览器内存缓存
  → 浏览器磁盘缓存（HTTP Cache）
  → CDN 边缘节点
  → Nginx / 反向代理
  → Go Gin API（通常 API 本身不「缓存」，但可返回 Cache-Control 指令）
```

每一层都读响应头里的 `Cache-Control`、`ETag` 等；**最严格的一层**往往决定最终行为。

---

## 2. 强缓存 vs 协商缓存

### 2.1 一句话对比

| 类型 | 是否发请求到服务器 | 关键响应头 | 典型状态码 |
|------|-------------------|------------|------------|
| **强缓存** | **不发**（直接用本地） | `Cache-Control`、`Expires` | 200（curl 看不到「from cache」字样，浏览器 DevTools 会标） |
| **协商缓存** | **发**，带验证字段 | `ETag` / `Last-Modified` | **304 Not Modified**（**无 body**） |

### 2.2 浏览器决策流程（面试必画）

```mermaid
sequenceDiagram
    participant B as 浏览器
    participant S as Go/Nginx 服务器

    B->>B: 查本地缓存
    alt 强缓存未过期
        B->>B: 直接使用缓存（无网络请求）
    else 强缓存过期或不存在
        B->>S: 请求 + If-None-Match / If-Modified-Since
        alt 资源未变
            S-->>B: 304 Not Modified
            B->>B: 用本地副本
        else 资源已变
            S-->>B: 200 + 新 body + 新 ETag
            B->>B: 替换本地缓存
        end
    end
```

**curl 视角**：curl **没有**浏览器那套磁盘缓存（默认每次真发请求），所以你能稳定看到 200/304，适合后端调试。

### 2.3 优先级

1. 先看 **强缓存**是否命中且未过期  
2. 过期或无强缓存 → 走 **协商缓存**  
3. 协商 miss → 正常 **200** 拉全量  

**口诀**：先「能不能直接用」，再「问服务器变没变」。

---

## 3. 强缓存：`Cache-Control` 与 `Expires`

### 3.1 `Expires`（HTTP/1.0，了解即可）

```http
Expires: Wed, 21 Oct 2026 07:28:00 GMT
```

绝对过期时间，依赖客户端时钟；HTTP/1.1 下被 `Cache-Control` **覆盖**。

### 3.2 `Cache-Control` 常见指令

| 指令 | 含义 | Go API 典型场景 |
|------|------|-----------------|
| `max-age=31536000` | 多少**秒**内强缓存有效 | 带 content-hash 的静态 JS/CSS |
| `no-cache` | 可存本地，但**使用前必须**向服务器验证 | `index.html` 入口 |
| `no-store` | **不存**任何缓存 | 登录、用户信息、支付 |
| `private` | 仅浏览器可缓存，CDN 不应缓存 | 含用户信息的页面 |
| `public` | 浏览器和 CDN 都可缓存 | 公共静态资源 |
| `immutable` | 缓存期内不必因刷新重新验证 | `app.[hash].js` |
| `must-revalidate` | 过期后必须验证，不能用 stale | 严格一致性资源 |

**短链 API（默认不缓存）**：

```http
Cache-Control: no-store, private
```

**带 hash 的静态资源（Nginx 或 Go FileServer）**：

```http
Cache-Control: public, max-age=31536000, immutable
```

### 3.3 `no-cache` vs `no-store`（高频面试）

| | no-cache | no-store |
|---|----------|----------|
| 能否存本地 | 能 | **不能** |
| 使用前 | 必须验证（常 304） | 每次都完整请求 |
| Go API | 可接受但必须验证的 GET | **登录/鉴权/写操作响应必用** |

### 3.4 在 Go 里设置 Cache-Control

```go
// 方式一：标准库
func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Cache-Control", "no-store")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte(`{"status":"ok"}`))
}

// 方式二：Gin
func Health(c *gin.Context) {
    c.Header("Cache-Control", "no-store, private")
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
```

---

## 4. 协商缓存：`ETag` 与 `Last-Modified`

### 4.1 `Last-Modified` / `If-Modified-Since`

**首次响应**：

```http
Last-Modified: Tue, 15 Apr 2025 10:00:00 GMT
```

**再次请求**：

```http
If-Modified-Since: Tue, 15 Apr 2025 10:00:00 GMT
```

服务端比较文件修改时间：未变 → **304**；变了 → **200** + 新 body。

**缺点**：精度只到秒；内容未变但 touch 了 mtime 也会失效。

### 4.2 `ETag` / `If-None-Match`（推荐）

**首次响应**：

```http
ETag: "33a64df551425fcc55e4d42a148795d9f25f89d4"
```

**再次请求**：

```http
If-None-Match: "33a64df551425fcc55e4d42a148795d9f25f89d4"
```

- **强 ETag**：字节级一致  
- **弱 ETag**：`W/"..."` 语义等价即可  

同时存在时，现代实现优先 **ETag**。

### 4.3 304 响应长什么样

```http
HTTP/1.1 304 Not Modified
Date: Sun, 12 Jul 2026 08:00:00 GMT
ETag: "abc123"
Cache-Control: no-cache
```

- **无响应体**（省带宽）  
- 浏览器用本地副本；curl 只收到响应头  

### 4.4 Go 标准库自动 ETag（静态文件）

`http.FileServer` 和 `http.ServeContent` 会为静态文件自动处理 `If-None-Match` / `If-Modified-Since`，命中时返回 304：

```go
// 静态资源目录：自动协商缓存
fs := http.FileServer(http.Dir("./static"))
http.Handle("/static/", http.StripPrefix("/static/", fs))
```

自定义 API 若要做 304（少见，多用于可缓存的公共 GET）：

```go
func publicStatsHandler(w http.ResponseWriter, r *http.Request) {
    body := []byte(`{"pv":1000}`)
    etag := fmt.Sprintf(`"%x"`, sha256.Sum256(body))
    w.Header().Set("ETag", etag)
    w.Header().Set("Cache-Control", "public, max-age=60")
    if r.Header.Get("If-None-Match") == etag {
        w.WriteHeader(http.StatusNotModified)
        return
    }
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write(body)
}
```

---

## 5. API 与静态资源的缓存策略（短链项目视角）

### 5.1 经典「入口不缓存、资源长缓存」

| 资源 | 谁提供 | 建议 Cache-Control |
|------|--------|-------------------|
| `index.html` | Nginx / 前端部署 | `no-cache` 或 `max-age=0, must-revalidate` |
| `assets/index-[hash].js` | Nginx / CDN | `public, max-age=31536000, immutable` |
| `GET /api/v1/links/:code` 跳转 | Go 302 | **通常不缓存**（要统计 PV）或极短 `max-age=0` |
| `POST /api/v1/auth/login` | Go Gin | `no-store` |
| `GET /api/v1/users/me` | Go Gin + JWT | `no-store, private` |

**原理**：HTML 是前端入口指针；JS 文件名带 hash，内容变则 URL 变，可长期强缓存。

### 5.2 Nginx 示例（短链前后端部署）

```nginx
# 前端静态
location / {
    root /usr/share/nginx/html;
    try_files $uri $uri/ /index.html;
    add_header Cache-Control "no-cache";
}

location /assets/ {
    root /usr/share/nginx/html;
    add_header Cache-Control "public, max-age=31536000, immutable";
}

# Go API 反代
location /api/ {
    proxy_pass http://127.0.0.1:8080;
    # API 缓存策略主要由 Go 响应头决定；这里可加兜底：
    add_header Cache-Control "no-store" always;
}
```

### 5.3 发版后用户仍旧版怎么办

1. 确认 `index.html` **没有**被 CDN/浏览器长期 `max-age` 强缓存  
2. 确认构建产物 **filename 含 hash**（Vite/Webpack 默认）  
3. 运维对 CDN 做 **purge**（仅 HTML 或全站，按策略）  
4. **后端 API 发版**不涉及浏览器 JS 缓存，但若改了 OpenAPI 契约，前端需重新构建  

---

## 6. CDN 概念（后端必知）

### 6.1 CDN 是什么

**内容分发网络**：把静态资源复制到离用户近的**边缘节点**。用户访问 `cdn.example.com/logo.png` 时，优先命中边缘；未命中再**回源**到源站（OSS/Nginx/Go 静态目录）。

```mermaid
flowchart LR
    U[用户 北京] --> E1[边缘节点 北京]
    U2[用户 广州] --> E2[边缘节点 广州]
    E1 -->|MISS 回源| O[源站 OSS/Nginx]
    E2 -->|HIT| E2
```

### 6.2 CDN 与浏览器缓存的关系

- CDN 节点也遵守 `Cache-Control`、`Expires`  
- `private` 的资源 CDN **不应**缓存  
- 带 **Cookie / Authorization** 的个性化响应：配置 CDN **不缓存**或按 Vary 分流（进阶）

### 6.3 缓存键（Cache Key）

CDN 判断「是否同一份缓存」常看：

- URL 路径 + Query（`?v=2` 会变成不同缓存项）  
- 部分 Header（如 `Accept-Encoding`）  
- **一般不看** `Authorization`——因此 **绝不要把带用户身份的 API 响应交给 CDN 公共缓存**

### 6.4 Go 后端要注意

| 做法 | 说明 |
|------|------|
| 静态资源走 CDN 域名 | 与 API 域名分离，减少 Cookie 误携带 |
| API 统一 `Cache-Control: no-store` | 防止敏感 JSON 被边缘缓存 |
| 公共只读 GET 若允许短缓存 | 显式 `public, max-age=60`，需业务同意 |
| 短链 302 跳转 | 通常 **不 CDN 缓存**（统计、换目标 URL） |

### 6.5 API 能否 CDN 缓存？

- **GET 公共只读**且允许短时缓存：可以 `max-age=60` + `public`（如公开配置字典）  
- **带 Cookie / Authorization 的**：`private, no-store`  
- **默认**：短链等业务 API **不 CDN 缓存**，只缓存静态资源  

---

## 7. Cookie 基础与属性（Go 后端视角）

### 7.1 Cookie 是什么

服务器通过 **`Set-Cookie`** 响应头让客户端存键值；之后同域请求客户端自动带 **`Cookie`** 头。

```http
Set-Cookie: sessionId=abc123; Path=/; HttpOnly; Secure; SameSite=Lax
```

```http
Cookie: sessionId=abc123
```

### 7.2 常见属性详解

| 属性 | 作用 | Go 后端必知 |
|------|------|-------------|
| **Name=Value** | 键值 | 不要存明文密码 |
| **Domain** | 哪些子域可带 | `.example.com` 含 `api.example.com` |
| **Path** | 哪些路径可带 | 默认请求路径 |
| **Expires / Max-Age** | 过期时间 | `Max-Age` 优先级更高 |
| **HttpOnly** | JS 读不到 | 防 XSS 偷 SessionId |
| **Secure** | 仅 HTTPS 发送 | 生产必开 |
| **SameSite** | 跨站是否携带 | 防 CSRF 关键 |

### 7.3 `SameSite` 三种值

| 值 | 行为 | 场景 |
|----|------|------|
| **Strict** | 跨站导航不带 Cookie | 高安全后台 |
| **Lax**（现代默认） | 顶级 GET 导航可带；跨站 Ajax **不带** | 大多数站点 |
| **None** | 跨站都带，须配合 **Secure** | 第三方登录、跨域嵌入 |

### 7.4 Go 设置 Cookie

```go
import "net/http"

func setSessionCookie(w http.ResponseWriter) {
    http.SetCookie(w, &http.Cookie{
        Name:     "sessionId",
        Value:    "abc123",
        Path:     "/",
        MaxAge:   3600,
        HttpOnly: true,
        Secure:   true, // 生产 HTTPS
        SameSite: http.SameSiteLaxMode,
    })
}
```

Gin 等价：

```go
c.SetCookie("sessionId", "abc123", 3600, "/", "", true, true)
// 参数：name, value, maxAge, path, domain, secure, httpOnly
// SameSite 需 Go 1.18+ 通过 c.Writer.Header().Add 或封装
```

### 7.5 Cookie 大小与数量

- 单条约 **4KB**  
- 每域约 **20～50** 条  
- 每次请求**全带上**——不要把巨大 JWT 链塞进多个 Cookie  

---

## 8. Session vs Cookie vs JWT（Go 后端选型）

### 8.1 概念对齐

| 概念 | 本质 | 状态放哪 |
|------|------|----------|
| **Cookie** | 客户端存储/传输**机制** | 客户端存，自动随请求发 |
| **Session** | 服务端**会话** | 服务端 Redis/内存；浏览器常只存 **SessionId** Cookie |
| **JWT** | 自包含**凭证** | 服务端无状态验签；客户端存 Header / Cookie / 内存 |

**易混点**：Session 往往**借助 Cookie** 传 SessionId；JWT 也**可以**放 HttpOnly Cookie，但前后端分离更常见 **Authorization: Bearer**。

### 8.2 Session + Cookie 流程

```mermaid
sequenceDiagram
    participant C as 客户端
    participant G as Go Gin
    participant R as Redis

    C->>G: POST /login 账号密码
    G->>R: SET sessionId → userInfo
    G-->>C: Set-Cookie: sessionId=xxx; HttpOnly
    C->>G: GET /api/v1/users/me Cookie: sessionId=xxx
    G->>R: GET sessionId
    R-->>G: userInfo
    G-->>C: 200 JSON
```

**优点**：服务端可随时删 Session 作废；敏感数据在服务端。  
**缺点**：分布式要 Redis 共享 Session；跨域 Cookie 麻烦（要 CORS `credentials` + 具体 Origin）。

### 8.3 JWT Token 流程（短链项目主线）

```mermaid
sequenceDiagram
    participant C as 客户端/curl
    participant G as Go Gin
    participant J as JWTManager

    C->>G: POST /api/v1/auth/login JSON
    G->>G: bcrypt 校验密码
    G->>J: IssueAccess(user)
    J-->>G: access_token
    G-->>C: {"access_token":"eyJhbG..."}
    C->>G: POST /api/v1/links Authorization: Bearer eyJ...
    G->>J: ParseAccess
    J-->>G: userID
    G-->>C: 201 创建成功
```

详见 [Go 09 JWT](../../后端学习/Go/09-JWT认证与用户体系.md) §3～§6。

**优点**：无状态、易水平扩展、跨域用 Header 即可。  
**缺点**：签发后难即时作废（需黑名单/短过期 + Refresh）；payload **不要**存密码。

### 8.4 对比总表（面试背这张）

| 维度 | Session + Cookie | JWT（Authorization Header） | JWT（HttpOnly Cookie） |
|------|------------------|----------------------------|------------------------|
| 状态 | 有状态（Redis） | 无状态 | 无状态 |
| 存储 | SessionId Cookie | 客户端内存 / 手动带 Header | HttpOnly Cookie |
| 自动携带 | 是（同域） | **否**，需客户端加 Header | 是 |
| XSS | HttpOnly 可防读 SessionId | 内存稍好；localStorage 高危 | HttpOnly 较安全 |
| CSRF | 需 SameSite / CSRF Token | Header 方式 CSRF 风险低 | 需 CSRF 防护 |
| 跨域 | Cookie 域 + CORS credentials | **简单**（ACAO + Authorization） | CORS + credentials |
| 注销 | 删 Redis Session | 黑名单 / 短 Access + Refresh | 清 Cookie + 黑名单 |
| 短链项目 | 可选 | **推荐主线**（Go 09） | 可选加强 |

### 8.5 Refresh Token（了解，Go 09 §7）

- **Access Token**：短过期（15min～2h），放 `Authorization: Bearer`  
- **Refresh Token**：长过期，HttpOnly Cookie 或安全存储  
- Access 过期 → 调 `/auth/refresh` 换新的，用户无感  

---

## 9. 为什么 JWT 不会「自动出现在请求里」

### 9.1 HTTP 规范只自动带 Cookie

```bash
# curl 手动带 Token——这就是客户端该做的事
curl http://localhost:8080/api/v1/links \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

| 存储 | 是否随 HTTP 自动发送 | Go 后端怎么读 |
|------|---------------------|---------------|
| Cookie | **是**（同域规则） | `c.Cookie("name")` / `r.Cookie` |
| JWT in Header | **否**，客户端必须加 | `c.GetHeader("Authorization")` |
| Redis Session | 通过 Cookie 间接 | 中间件读 Cookie → 查 Redis |

### 9.2 Go 09 AuthMiddleware 核心逻辑

```go
auth := c.GetHeader("Authorization")
if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
    response.Fail(c, http.StatusUnauthorized, "未登录")
    c.Abort()
    return
}
tokenStr := strings.TrimPrefix(auth, "Bearer ")
claims, err := jwt.ParseAccess(tokenStr)
```

**面试一句话**：Cookie 是浏览器内置的「自动附加器」；JWT 放 Header 是应用层约定，**curl、移动端、别的服务**都要自己带。

### 9.3 选型建议（短链 / Go API）

| 数据 | 推荐 |
|------|------|
| 登录 Access JWT | `Authorization: Bearer`（Go 09 主线） |
| Refresh Token | HttpOnly Cookie 或 Redis 存 Refresh |
| 服务端 SessionId | HttpOnly + Secure + SameSite Cookie |
| 限流 / 黑名单 | Redis（Go 08） |

---

## 10. CORS 完整原理（面试必修，curl 不受影响）

### 10.1 同源策略

**同源** = 协议 + 域名 + 端口**全部相同**。

| URL A | URL B | 是否同源 |
|-------|-------|----------|
| `http://localhost:8080` | `http://localhost:8080` | ✅ |
| `http://localhost:5173` | `http://localhost:8080` | ❌ 端口不同 |
| `https://api.example.com` | `https://www.example.com` | ❌ 主机不同 |

**重点**：跨域时，请求常常**已经到达** Go 服务器；浏览器检查响应头后，若不符合 CORS 规则，**不把响应 body 交给页面里的 JavaScript**。

### 10.2 为什么 curl 从不报 CORS

| 工具 | 是否执行同源策略 | 原因 |
|------|-----------------|------|
| **curl** | ❌ 不执行 | 命令行客户端，无「页面 JS 读响应」的安全模型 |
| **Postman** | ❌ 不执行 | 同上 |
| **浏览器 fetch/XHR** | ✅ 执行 | 保护用户不在 `evil.com` 的脚本读取 `bank.com` 的数据 |

所以：**后端用 curl 联调一切正常，前端一调就红字**——不是 Go 没收到请求，是**浏览器拦了 JS 读响应**。

**浏览器 fetch 最小示例**（对照 curl `-H`）：

```javascript
// JWT：须手动带头（不会像 Cookie 自动带）
await fetch('http://localhost:8080/api/v1/links', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer ' + token
  },
  body: JSON.stringify({ original_url: 'https://example.com' })
});

// Session Cookie：须 credentials，且服务端 ACAO 不能是 *
await fetch('http://localhost:8080/api/me', { credentials: 'include' });
```

### 10.3 简单请求 vs 需预检请求

**简单请求**（同时满足）：

- 方法：GET、HEAD、POST  
- Header 仅：`Accept`、`Accept-Language`、`Content-Language`、`Content-Type`（限三种）  
- `Content-Type` 只能是：`text/plain`、`multipart/form-data`、`application/x-www-form-urlencoded`

**否则** → 浏览器先发 **OPTIONS 预检**。

短链项目典型 **非简单请求**（必预检）：

- `Content-Type: application/json`（登录、创建短链）  
- 自定义头 `Authorization: Bearer ...`  
- 方法 PUT、DELETE  

### 10.4 预检 OPTIONS 流程

```mermaid
sequenceDiagram
    participant B as 浏览器
    participant G as Go Gin

    B->>G: OPTIONS /api/v1/auth/login
    Note right of B: Origin: http://localhost:5173<br/>Access-Control-Request-Method: POST<br/>Access-Control-Request-Headers: authorization,content-type
    G-->>B: 204 + ACAO + ACAM + ACAH + Max-Age
    B->>G: POST /api/v1/auth/login（真实请求）
    G-->>B: 200 + ACAO + JSON body
```

关键响应头：

| 响应头 | 含义 |
|--------|------|
| `Access-Control-Allow-Origin` | 允许的来源（`credentials` 时不能 `*`） |
| `Access-Control-Allow-Methods` | 允许的方法 |
| `Access-Control-Allow-Headers` | 允许的请求头（**必须含 Authorization**） |
| `Access-Control-Allow-Credentials` | 是否允许带 Cookie |
| `Access-Control-Max-Age` | 预检结果缓存秒数 |
| `Access-Control-Expose-Headers` | 允许 JS 读到的额外响应头 |

### 10.5 常见 Console 报错

```text
Access to fetch at 'http://localhost:8080/api/v1/auth/login' from origin
'http://localhost:5173' has been blocked by CORS policy:
Response to preflight request doesn't pass access control check:
No 'Access-Control-Allow-Origin' header is present on the requested resource.
```

**排查顺序**（即使你现在只用 curl，面试/联调必背）：

1. Network 里有没有 **OPTIONS**，状态码多少  
2. OPTIONS 响应是否含 `Access-Control-Allow-Origin`  
3. 真实 POST 响应是否**也**带 CORS 头（有人只配了 POST 漏 OPTIONS）  
4. `credentials: true` 时是否错误用了 `Allow-Origin: *`  
5. Auth 中间件是否把 **OPTIONS 拦成 401**（最常见 Go 坑之一）

---

## 11. 在 Gin 里修复 CORS

### 11.1 方案对比

| 方案 | 谁解决跨域 | 适用 |
|------|------------|------|
| **curl / Postman 直连** | 不涉及 CORS | **Go 后端日常开发首选** |
| **Gin CORS 中间件** | 后端加响应头 | 浏览器直连 API、生产跨域 |
| **Nginx 反代同域** | `/api` 与页面同域 | 生产推荐：`https://short.example.com/api/` |
| **Vite dev proxy**（选修） | 开发服务器转发，浏览器认为同源 | 前端同学本地用；**不是 Go 必修** |

### 11.2 Gin CORS 配置（对照 Go 06 §6）

```go
import "github.com/gin-contrib/cors"

func SetupCORS(r *gin.Engine) {
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))
}
```

**注意**：

- `AllowCredentials: true` 时 `AllowOrigins` **不能**写 `*`，要用**具体源**  
- CORS 中间件应挂在 **Auth 中间件之前**（或 Auth 对 OPTIONS 放行）  
- 确保 **OPTIONS** 返回 204/200 且带齐 CORS 头  

### 11.3 Auth 中间件必须放行 OPTIONS

```go
func AuthMiddleware(jwt *JWTManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 预检请求不带 Authorization，必须放行
        if c.Request.Method == http.MethodOptions {
            c.Next()
            return
        }
        auth := c.GetHeader("Authorization")
        if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
            response.Fail(c, http.StatusUnauthorized, "未登录")
            c.Abort()
            return
        }
        // ... ParseAccess ...
        c.Next()
    }
}
```

更干净的做法：公开路由与受保护路由**分组**，CORS 全局，Auth 只挂在受保护组上（见 [Go 09](../../后端学习/Go/09-JWT认证与用户体系.md)）。

### 11.4 手写最小 CORS 中间件（理解原理）

```go
func SimpleCORS(allowedOrigin string) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", allowedOrigin)
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Header("Access-Control-Allow-Credentials", "true")
        if c.Request.Method == http.MethodOptions {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }
        c.Next()
    }
}
```

### 11.5 生产 Nginx 同域（推荐）

```nginx
server {
    listen 443 ssl;
    server_name short.example.com;

    location / {
        root /usr/share/nginx/html;  # 前端静态
    }

    location /api/ {
        proxy_pass http://127.0.0.1:8080;  # Go Gin
        proxy_set_header Host $host;
    }
}
```

浏览器只访问 `https://short.example.com`，**同源，无 CORS**。  
若前后端必须不同域（`api.example.com` + `www.example.com`），则 **Gin CORS 必配**。

### 11.6 Vite proxy（选修，给前端联调参考）

```javascript
// vite.config.js — 前端本地开发可选用，Go 后端无需配置
export default defineConfig({
  server: {
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
    },
  },
})
```

原理：浏览器只请求 `http://localhost:5173/api/...`（同源），Vite 在 Node 侧转发到 `8080`。**你的 Go 服务仍建议配 CORS**，否则别人绕过 proxy 直连仍会失败。

---

## 12. 手把手：curl 验证缓存与短链认证

### 12.1 目标

完成：**看 Cache-Control → 打 304 → 登录拿 token → 带 Bearer 创建短链 → 无 token 得 401**。

### 12.2 查看响应头（强缓存 / no-store）

```powershell
# 只看响应头
curl -I http://localhost:8080/health

# 详细输出（找 > Cache-Control）
curl -v http://localhost:8080/health 2>&1 | Select-String -Pattern "Cache-Control|ETag|< HTTP"
```

**预期**：健康检查 API 应类似：

```http
HTTP/1.1 200 OK
Cache-Control: no-store
Content-Type: application/json
```

### 12.3 协商缓存 304（静态文件或自定义 ETag 接口）

```powershell
# 第一次：记下 ETag
curl -v http://localhost:8080/static/logo.png 2>&1

# 第二次：带 If-None-Match
curl -v http://localhost:8080/static/logo.png `
  -H "If-None-Match: `"abc123`"" 2>&1 | Select-String -Pattern "< HTTP|ETag"
```

**预期**：`< HTTP/1.1 304 Not Modified`，且无 body。

### 12.4 短链登录与鉴权（curl 全流程）

```powershell
# 1. 注册（若项目有）
curl -X POST http://localhost:8080/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d '{"username":"demo","password":"demo123456"}'

# 2. 登录拿 token
curl -X POST http://localhost:8080/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{"username":"demo","password":"demo123456"}'

# 3. 无 token 创建短链 → 401
curl -X POST http://localhost:8080/api/v1/links `
  -H "Content-Type: application/json" `
  -d '{"long_url":"https://example.com"}'

# 4. 带 Bearer 创建短链 → 201/200
curl -X POST http://localhost:8080/api/v1/links `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer <粘贴 access_token>" `
  -d '{"long_url":"https://example.com"}'
```

### 12.5 用 curl 模拟预检（理解 CORS，非浏览器真实行为）

curl **不会**自动发 OPTIONS，但你可以手动模拟浏览器预检：

```powershell
curl -v -X OPTIONS http://localhost:8080/api/v1/auth/login `
  -H "Origin: http://localhost:5173" `
  -H "Access-Control-Request-Method: POST" `
  -H "Access-Control-Request-Headers: authorization,content-type" 2>&1
```

**预期响应头含**：

```http
Access-Control-Allow-Origin: http://localhost:5173
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Origin, Content-Type, Authorization
```

若没有这些头，浏览器里的 fetch 会失败——**与 curl POST 是否成功无关**。

### 12.6 浏览器 DevTools（选修：CORS 可视化）

1. 前端或简单 HTML 页面从 `http://localhost:5173` fetch `http://localhost:8080/api/v1/auth/login`  
2. Network → 先出现 **OPTIONS**（预检），再 **POST**  
3. 若 OPTIONS 4xx/缺 ACAO → Console CORS 红字  
4. Application → Cookies：Session 方案能看到 Cookie；JWT Header 方案通常**无** token Cookie  

### 12.7 验收步骤表

| 步骤 | curl / 工具预期 | 说明 |
|------|-----------------|------|
| 1. `curl -I /health` | 有 `Cache-Control: no-store` | API 不应被缓存 |
| 2. 静态资源带 ETag 再请求 | 304 或 200 | 理解协商缓存 |
| 3. POST login | 200 + `access_token` JSON | 无 CORS 问题（curl） |
| 4. POST /links 无 Bearer | **401** | Auth 中间件生效 |
| 5. POST /links 有 Bearer | **201/200** | 完整认证链 |
| 6. OPTIONS 预检（手动 curl） | 响应含 ACAO、ACAH | 联调前端前必验 |

### 12.8 端到端时序图（短链创建）

```mermaid
sequenceDiagram
    participant C as curl/客户端
    participant G as Gin
    participant J as JWTManager
    participant DB as MySQL

    C->>G: POST /api/v1/auth/login
    G->>DB: 查用户 + bcrypt
    G->>J: IssueAccess
    G-->>C: access_token
    C->>G: POST /api/v1/links + Bearer
    G->>J: ParseAccess → userID
    G->>DB: INSERT link
    G-->>C: 201 short_url
```

---

## 13. 常见错误与排查表

| # | 现象 | 原因 | 解决 |
|---|------|------|------|
| 1 | 用户看到旧版前端 | `index.html` 被 `max-age` 强缓存 | HTML 改 `no-cache`；资源用 hash |
| 2 | API 响应被 CDN 缓存串号 | 带 Authorization 的 GET 未设 `no-store` | Go 统一 `Cache-Control: no-store, private` |
| 3 | curl 看到 304 但内容不对 | ETag 未随 body 更新 | 修 ETag 计算或禁用该路径缓存 |
| 4 | Cookie 设置了但请求不带 | Domain/Path/SameSite 不匹配 | 对照 Set-Cookie 与请求 URL |
| 5 | Session 跨域失败 | 未配 CORS `Allow-Credentials` + 具体 Origin | Gin cors `AllowCredentials: true` |
| 6 | 有 token 仍 401 | Header 不是 `Bearer ` 前缀 / 中间件顺序错 | 对照 Go 09 `AuthMiddleware` |
| 7 | OPTIONS 404 或 401 | 无 OPTIONS 路由或被 Auth 拦截 | CORS 全局；Auth 放行 OPTIONS |
| 8 | CORS 红字但 curl 正常 | 浏览器独有同源策略 | 配 ACAO/ACAH；见 §11 |
| 9 | `Allow-Origin: *` + credentials | 规范冲突 | 指定具体 Origin |
| 10 | JWT 注销仍能用 | 无状态 token 未过期 | 短 Access + Refresh 黑名单（Go 09 §7） |
| 11 | 预检通过 POST 仍失败 | POST 响应缺 ACAO | 所有 Method 都要回 CORS 头 |
| 12 | 敏感 JSON 出现在 CDN | 误配 `public` 缓存 API | API 默认 `no-store` |

---

## 14. 深入与安全（合并）

304 仍有 **一次 RTT**，强缓存未过期则零请求——hash 静态用 `immutable`，API 用 `no-store`。

| 风险 | 防护 |
|------|------|
| XSS 偷 token | HttpOnly Cookie 存 Refresh；输入转义 |
| CSRF | SameSite；JWT Header 降低风险 |
| 中间人 | HTTPS + Secure Cookie（[05 章](./05-HTTPS与TLS加密.md)） |
| CDN 缓存用户数据 | `private, no-store` |
| CORS 过宽 | `AllowOrigins` 白名单 |
| 日志泄露 | 不记录完整 `Authorization`（[Go 09](../../后端学习/Go/09-JWT认证与用户体系.md) §8） |

---

## 15. 与 Go 章节对照

| 本章现象 / 技能 | Go 文档 |
|-----------------|---------|
| Gin CORS 中间件 | [06-Gin §6](../../后端学习/Go/06-Gin框架核心与中间件.md) |
| JWT 签发 / AuthMiddleware | [09-JWT §3～§6](../../后端学习/Go/09-JWT认证与用户体系.md) |
| Redis Session / 黑名单 | [08-Redis](../../后端学习/Go/08-Redis与go-redis缓存实战.md) |
| net/http 基础 / Header | [05-标准库 HTTP](../../后端学习/Go/05-Go标准库与HTTP基础.md) |
| 短链登录实战 | [10～11 短链项目](../../后端学习/Go/10-短链服务项目实战上.md) |
| 部署 Nginx 同域 | [13-Docker 与 Linux 部署](../../后端学习/Go/13-Docker与Linux部署Go服务.md) |

---

## 16. 练习与参考答案（精选）

**基础**：① 强缓存用 `Cache-Control`/`Expires`，协商用 `ETag`/`Last-Modified`；304 **无 body**。② `no-cache` 可存须验证；`no-store` 完全不存。③ HttpOnly / Secure / SameSite 防 XSS / 明文传输 / CSRF。④ curl 非浏览器，不执行同源策略。

**进阶**：⑤ 见 §12.8 时序图。⑥ Gin CORS 在跨域直连时加响应头；Nginx 同域让浏览器不触发跨域。⑦ API 含用户态须 `no-store`；hash JS URL 随内容变可长缓存。

**挑战**：⑧ HttpOnly Cookie 防 XSS 读 token 但须 CSRF 防护；Bearer Header CSRF 低但 XSS 仍能发请求——常见短 Access + Refresh HttpOnly。⑨ `curl -I` 看 `Cache-Control`；`curl -H "If-None-Match: ..."` 打 304。

---

## 17. 本章知识速查卡

| 主题 | 一句话 |
|------|--------|
| 强缓存 | 未过期不请求，`Cache-Control: max-age` |
| 协商缓存 | 过期后带 ETag 问，304 无 body |
| API 策略 | 默认 `no-store, private` |
| 静态策略 | HTML `no-cache`，hash 资源 `immutable` |
| CDN | 边缘缓存静态；API 禁止公共缓存 |
| Session | 状态在服务端 Redis，Cookie 常只带 id |
| JWT | 自包含凭证，`Authorization: Bearer` |
| CORS | 浏览器拦 JS 读跨域响应；curl 不受影响 |
| 预检 | 非简单请求先 OPTIONS |
| Gin 修复 | `gin-contrib/cors` + Auth 放行 OPTIONS |

---

## 18. 常见问题 FAQ（≥10）

**Q1：`no-cache` 到底缓不缓存？**  
**会存**本地，但每次用前必须向服务器验证（常 304）。真正不存的是 `no-store`。

**Q2：Go API 默认该加什么缓存头？**  
`Cache-Control: no-store, private`，除非明确的公共只读 GET 且业务允许短缓存。

**Q3：304 和强缓存有什么区别？**  
304 **有网络往返**（只下响应头）；强缓存命中通常不发完整 HTTP 请求（浏览器侧；curl 通常每次请求）。

**Q4：为什么 JWT 不自动出现在请求里？**  
HTTP 只自动带 Cookie；Bearer Token 是应用层约定，需客户端（或 curl `-H`）手动加。

**Q5：Session 和 JWT 选哪个？**  
前后端分离、Go 多实例无粘滞 → **JWT**（Go 09 主线）；要强注销、传统单体 → Session + Redis。

**Q6：CORS 是 Go 中间件拦请求吗？**  
不是。请求常**已到 Gin**；浏览器因响应头不合规**不把 body 交给 JS**。curl 不受影响。

**Q7：开发后端为什么要学 CORS？**  
你主要用 curl；但面试必考，且前端联调 Go API 时你是修 CORS 的人。

**Q8：CDN 会缓存带 Authorization 的 API 吗？**  
**不应**配置；响应须 `private, no-store`。

**Q9：OPTIONS 为什么要放行？**  
预检**不带** Authorization；若 Auth 中间件拦 OPTIONS，浏览器永远过不了预检。

**Q10：`Allow-Origin: *` 为什么和 Cookie 冲突？**  
规范要求：带 `credentials` 时必须**指定具体 Origin**，否则无法保证凭证安全。

**Q11：短链 302 跳转能 CDN 缓存吗？**  
一般**不缓存**——要统计点击、目标 URL 可能变；如需缓存须业务明确且接受统计偏差。

**Q12：Vite proxy 和 Gin CORS 二选一吗？**  
否。proxy 只帮**本地前端**；直连 API、测试环境、移动端仍要 Gin CORS 或 Nginx 同域。

---

## 19. Set-Cookie 逐行读（Session 方案）

```http
Set-Cookie: SESSION=abc123; Path=/; HttpOnly; Secure; SameSite=Lax; Max-Age=3600
```

| 字段 | 含义 | 改错会怎样 |
|------|------|------------|
| SESSION=abc123 | 会话 id | 明文存密码危险 |
| Path=/ | 全站路径携带 | Path 窄了 API 不带 |
| HttpOnly | JS 不可读 | XSS 能 `document.cookie` 偷 |
| Secure | 仅 HTTPS | HTTP 下 Cookie 不发 |
| SameSite=Lax | 跨站 POST 不带 | CSRF 风险变化 |
| Max-Age=3600 | 1 小时过期 | 永不过期风险 |

Go 对应：`http.SetCookie` / `c.SetCookie`，见 §7.4。  
Java 对应：`ResponseCookie` / Spring `Set-Cookie` 响应头。  
Python 对应：Flask `response.set_cookie(..., httponly=True, secure=True)`。

---

## 20. 模拟面试：缓存与会话六连问

1. **强缓存 vs 协商缓存？** → max-age 不请求；ETag 304 无 body。  
2. **API 为何默认 no-store？** → 防 CDN/浏览器缓存用户数据。  
3. **Cookie vs Bearer JWT？** → 自动携带 vs 客户端手动带。  
4. **OPTIONS 何时出现？** → 非简单请求（JSON + Authorization）。  
5. **curl 正常、浏览器 CORS 失败？** → 浏览器同源策略；配 ACAO/ACAH。  
6. **SameSite=Lax 防什么？** → 跨站携带 Cookie 的 CSRF。

---

## 21. 闭卷自测（10 题）

### 概念题（6 题）

1. 用「**冰箱便当**」类比强缓存；用「**打电话问饭店**」类比协商缓存。  
2. `no-cache` 与 `no-store` 区别？  
3. ETag 与 Last-Modified 优先级与精度差异？  
4. Session、Cookie、JWT 三概念如何区分？  
5. 简单请求与需预检请求各举短链一例。  
6. HttpOnly、Secure、SameSite 各防什么？  

### 动手题（2 题）

7. 写出 Gin `gin-contrib/cors` 允许 `Authorization` 的最小配置要点（3 项）。  
8. 如何用 curl 验证 API 返回了 `Cache-Control: no-store`？  

### 综合题（2 题）

9. 设计短链登录：JWT + Bearer + 401，写 5 步时序（文字即可）。  
10. 生产前后端不同域时，CORS 需配哪些响应头？`credentials: true` 时 Allow-Origin 能写 `*` 吗？  

### 自测参考答案

**1.** 强缓存=保质期内直接吃；协商=过期后问是否同一批，是则 304 不重送 body。  

**2.** no-cache 可存须验证；no-store 完全不存。  

**3.** ETag 内容级更准；同时存在优先 ETag。  

**4.** Cookie=机制；Session=服务端状态常借 Cookie 传 id；JWT=自包含凭证。  

**5.** 简单：GET `/health`；预检：POST login JSON + Bearer。  

**6.** HttpOnly 防 XSS 读 Cookie；Secure 仅 HTTPS；SameSite 减 CSRF。  

**7.** `AllowOrigins` 具体源、`AllowHeaders` 含 `Authorization`、`AllowMethods` 含 `OPTIONS`。  

**8.** `curl -I url` 或 `curl -v url 2>&1` 查找 `Cache-Control: no-store`。  

**9.** POST login→拿 token→POST links 无 token 401→带 Bearer 201→过期 401。  

**10.** ACAO、ACAM、ACAH、Allow-Credentials；不能 `*`，须具体 Origin。  

---

## 22. 费曼检验：3 分钟讲给零基础朋友

**对照是否讲到：**

1. **缓存像冰箱**：没过保质期不用问；过期了打电话问（304），没变就不重送整盒。API 像现做菜，**不能**放冰箱给下一位客人吃（`no-store`）。  
2. **Cookie 像手环自动进门**；JWT 像**票在口袋里**，检票要自己掏（`Authorization: Bearer`）——curl 也要你手写 `-H`。  
3. **跨域是浏览器保安**，curl 没有保安；修 CORS 头，或让前后端住同一个域名（Nginx）。

---

## 23. Cache-Control 组合决策表

| 资源类型 | 推荐组合 | 原因 |
|----------|----------|------|
| `index.html` | `no-cache` 或 `max-age=0, must-revalidate` | 入口指针及时更新 |
| `assets/*.[hash].js` | `public, max-age=31536000, immutable` | URL 变即新资源 |
| 用户私有 HTML | `private, no-cache` | CDN 不应共享 |
| 登录/用户 API | `private, no-store` | 禁止任何层缓存 |
| 公共 GET 字典（短） | `public, max-age=60` | 需业务同意 |
| 短链 302 | 不缓存或极短 | 统计与目标 URL |

---

## 24. CORS 响应头最小集（多栈对照）

| 头 | 示例值 | 缺了会怎样 |
|----|--------|------------|
| Access-Control-Allow-Origin | `http://localhost:5173` | Console CORS blocked |
| Access-Control-Allow-Methods | GET,POST,PUT,DELETE,OPTIONS | 预检失败 |
| Access-Control-Allow-Headers | Content-Type, Authorization | Bearer 被拒 |
| Access-Control-Allow-Credentials | true | Cookie 跨域失败 |
| Access-Control-Max-Age | 43200 | 预检频繁（性能） |

**credentials: true 时** Origin 不能为 `*`，必须写具体源。

| 栈 | 典型修法 |
|----|----------|
| **Go Gin** | `gin-contrib/cors` 中间件，见 [Go 06 §6](../../后端学习/Go/06-Gin框架核心与中间件.md) |
| **Spring Boot** | `@CrossOrigin(origins="http://localhost:5173", allowCredentials="true")` 或全局 `CorsConfiguration` |
| **Python Flask** | `Flask-CORS`：`CORS(app, origins=[...], supports_credentials=True)` |
| **浏览器 fetch** | `fetch(url, { credentials: 'include' })` 才会带 Cookie；JWT 仍要手动加 `Authorization` 头 |
| **Nginx 同域** | 前后端同一域名反代 `/api` → 浏览器不触发跨域，可省 CORS 配置 |

---

## 25. 本章总复习清单

1. `curl -I` 读 API 的 `Cache-Control`；静态资源打一次 304。  
2. 完成 §12.7 短链认证验收表 5 步。  
3. 口述 Session vs JWT 对比表（§8.4）。  
4. 闭卷自测 §21 ≥8/10。  
5. 费曼 3 分钟录音自评。  
6. 能解释为何 API 默认 `no-store`。  
7. 手动 curl OPTIONS，确认 CORS 头齐全。  
8. 对照 Go 09 完成 AuthMiddleware + Bearer 实战（或理解 Spring/Flask 等价写法）。

---

## 学完标准

| # | 标准 | 自检 |
|---|------|------|
| 1 | 能口述强缓存 vs 协商缓存，画出 304 时序 | ⬜ |
| 2 | 能解释 `Cache-Control` 至少 5 个指令 | ⬜ |
| 3 | 能说明 API vs hash 静态资源缓存策略 | ⬜ |
| 4 | 能对比 Session、Cookie、JWT 适用场景 | ⬜ |
| 5 | 能解释 HttpOnly / Secure / SameSite | ⬜ |
| 6 | 能说明 JWT 为何要客户端带 Bearer | ⬜ |
| 7 | 能排查 OPTIONS 预检失败 | ⬜ |
| 8 | 能完成 §12.7 短链认证验收表 | ⬜ |
| 9 | 能用 curl -I / -v 读缓存相关头 | ⬜ |
| 10 | §13 常见错误表至少能讲 8 条 | ⬜ |

---

## 下一章预告

- **计网 07**：[面试专题与知识点总表](./07-面试专题与知识点总表.md)——缓存、CORS、Cookie/JWT 口述题汇总  
- **Go 主线**：继续 [06 Gin](../../后端学习/Go/06-Gin框架核心与中间件.md) → [09 JWT](../../后端学习/Go/09-JWT认证与用户体系.md) → [10 短链实战](../../后端学习/Go/10-短链服务项目实战上.md)  

建议本章 curl 验收通过后再做 Go 09 编码，面试题可直接用短链登录链路举例。

---

*UTF-8 | 系列索引：[00 学习路线图](./00-学习路线图与说明.md) · 规范：[修改规范](../../修改规范.md)*

*本章已按 EXPANSION-STANDARD 扩充（§0～§0.5、FAQ≥12、闭卷自测、费曼、Mermaid、报错表≥12）。*

**EXPANSION-STANDARD 自检**：☑ §0 导读 ☑ 冰箱/手环类比 §0.1 ☑ curl 实操 §12 ☑ Set-Cookie 逐行 §19 ☑ FAQ §18（12 问）☑ 闭卷 10 题 §21 ☑ 费曼 §22 ☑ 报错表 §13（12 条）☑ 交叉 Go 06/09
