# HTTP 协议深入

> 面向前端 / 全栈面试：从报文结构到 HTTP/3，把你在 [HTML CSS JS 10-浏览器HTTP网络与Web基础](../HTML%20CSS%20JS/10-浏览器HTTP网络与Web基础.md) 里建立的「HTTP 是一来一回」认知，升级到能读原始报文、选对方法、解释状态码、对比 HTTP 版本，并与 [Java 04-SpringBoot核心开发](../../后端学习/Java/04-SpringBoot核心开发.md) 的 `Result<T>` 和 [Vue 08-Axios网络请求与前后端联调](../Vue/08-Axios网络请求与前后端联调.md) / [React 08-Axios网络请求与前后端联调](../React/08-Axios网络请求与前后端联调.md) 联调场景对齐。

---

## 本章衔接

### 与前面章节的关系

| 章节 | 你已具备的基础 | 本章补什么 |
|------|----------------|------------|
| [HTML 10](../HTML%20CSS%20JS/10-浏览器HTTP网络与Web基础.md) | URL、GET/POST、常见状态码、Network 面板初识 | **完整报文结构**、全部状态码分类、幂等性、HTTP/2/3、REST |
| [HTML 09](../HTML%20CSS%20JS/09-JavaScript异步编程网络请求与本地存储.md) | fetch、Promise、localStorage 存 token | 理解 Axios/fetch 底层发的是什么 HTTP |
| [Java 04](../../后端学习/Java/04-SpringBoot核心开发.md) | Spring Boot Controller、`Result<T>`、CORS | 把 JSON 响应对应到 HTTP 状态码 + Header 设计 |
| 计算机网络 01～03（建议先读） | TCP/IP、DNS、三次握手 | 本章专注**应用层 HTTP**，传输层细节见 03 章 |

```mermaid
flowchart TB
    subgraph 应用层["应用层（本章）"]
        HTTP[HTTP 报文 / 方法 / 状态码]
        REST[RESTful 设计]
    end
    subgraph 传输层["传输层（03 章）"]
        TCP[TCP 连接]
        QUIC[QUIC / UDP]
    end
    subgraph 前端["前端联调"]
        Axios[Vue/React 08 Axios]
        DevTools[Network 面板]
    end
    HTTP --> TCP
    HTTP --> QUIC
    Axios --> HTTP
    DevTools --> HTTP
    REST --> HTTP
```

**学习建议**：本地启动 Java 04 的 `demo` 项目（`8080`），配合本章所有 `curl` 命令跟做；浏览器打开 shop-vue 的 Network 面板对照报文。

---

## 1. HTTP 是什么：无状态的应用层协议

HTTP（HyperText Transfer Protocol，超文本传输协议）定义了**客户端与服务器之间交换消息的格式和语义**。

### 1.1 核心特征

| 特征 | 含义 | 对前端的影响 |
|------|------|--------------|
| **无状态** | 服务器默认不记住上一次请求 | 登录态靠 Cookie / Session / Token 自己带 |
| **请求-响应模型** | 一问一答 | Axios 的 then/catch 对应一次完整往返 |
| **可扩展 Header** | 元数据放在头部 | `Authorization`、`Content-Type` 等 |
| **分层** | 建立在 TCP（或 QUIC）之上 | 慢有时不是 HTTP 本身，而是连接层 |

### 1.2 为什么 HTTP 要设计成无状态？（深入解释 ①）

**原因**：早期 Web 主要是静态文档浏览，无状态让服务器实现简单、易水平扩展——任意一台机器都能处理任意请求，不必查「这个用户上一秒干了什么」。

**代价**：登录、购物车、权限都要**额外机制**：

- Cookie + Session：SessionId 在 Cookie，状态在服务端
- JWT Token：状态编码在 Token 里，放 `Authorization` 头（shop-vue 08 章做法）

```mermaid
sequenceDiagram
    participant C as 浏览器 / Axios
    participant S as Spring Boot
    Note over C,S: 无状态：每个请求自带凭证
    C->>S: GET /api/users<br/>Authorization: Bearer xxx
    S->>C: 200 + Result JSON
    C->>S: POST /api/login<br/>Body: {username, password}
    S->>C: 200 + {token: "eyJ..."}
```

---

## 2. HTTP 报文结构：请求与响应

一条 HTTP 消息由三部分组成（CRLF `\r\n` 换行）：

```text
起始行（Start Line）
头部字段（Headers）      ← 键值对，空行结束
空行（CRLF）
消息体（Body，可选）
```

### 2.1 HTTP 请求报文

```http
POST /api/login HTTP/1.1
Host: localhost:8080
Content-Type: application/json
Accept: application/json
Content-Length: 42
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9

{"username":"admin","password":"123456"}
```

| 部分 | 示例 | 说明 |
|------|------|------|
| **请求行** | `POST /api/login HTTP/1.1` | 方法 + 路径 + 协议版本 |
| **Host** | `localhost:8080` | HTTP/1.1 **必填**，虚拟主机路由 |
| **Content-Type** | `application/json` | Body 格式 |
| **Content-Length** | `42` | Body 字节长度 |
| **Body** | JSON 字符串 | POST/PUT/PATCH 常见 |

### 2.2 HTTP 响应报文

```http
HTTP/1.1 200 OK
Content-Type: application/json;charset=UTF-8
Transfer-Encoding: chunked
Date: Thu, 18 Jun 2026 08:00:00 GMT

{"code":0,"message":"success","data":{"token":"eyJ..."}}
```

| 部分 | 示例 | 说明 |
|------|------|------|
| **状态行** | `HTTP/1.1 200 OK` | 版本 + **状态码** + 原因短语 |
| **Content-Type** | `application/json` | 告诉客户端如何解析 Body |
| **Body** | JSON | 业务数据 |

### 2.3 请求行三要素详解

```http
GET /api/users?page=1&size=10 HTTP/1.1
│   │                           │
│   └─ 请求 URI（路径 + 查询串）  └─ 协议版本
└─ 方法
```

- **方法**：动词，表达操作语义（GET、POST…）
- **URI**：统一资源标识；REST 里常对应资源路径
- **版本**：`HTTP/1.1` 最常见；HTTP/2 在帧层封装，调试工具仍常显示语义等价的 1.1 风格

---

## 3. 常见请求头与响应头

### 3.1 前端必须熟悉的请求头

| Header | 典型值 | 作用 | shop-vue / Java 04 场景 |
|--------|--------|------|-------------------------|
| **Host** | `localhost:8080` | 目标主机，HTTP/1.1 必需 | Vite proxy 转发时 Host 可能变 |
| **Content-Type** | `application/json` | 声明 Body 格式 | POST `/api/login` 提交 JSON |
| **Accept** | `application/json` | 客户端期望的响应类型 | 告诉 Spring 返回 JSON 而非 HTML |
| **Authorization** | `Bearer <JWT>` | 认证凭证 | 登录后 Axios 拦截器自动附加 |
| **Cookie** | `sessionid=abc` | 自动携带 Cookie | Session 方案；JWT 方案较少用 |
| **User-Agent** | `Mozilla/5.0 ...` | 客户端标识 | 服务端统计、兼容 |
| **Origin** | `http://localhost:5173` | 请求来源 | CORS 预检时后端要比对 |
| **Referer** | 上一页 URL | 来源页面 | 防盗链、统计 |

### 3.2 常见响应头

| Header | 典型值 | 作用 |
|--------|--------|------|
| **Content-Type** | `application/json;charset=UTF-8` | 响应体类型 |
| **Set-Cookie** | `token=...; HttpOnly; Secure` | 服务端写 Cookie |
| **Cache-Control** | `no-cache` / `max-age=3600` | 缓存策略 |
| **Access-Control-Allow-Origin** | `*` 或具体域名 | CORS 允许的来源 |
| **Location** | `https://example.com/new` | 302/301 重定向目标 |
| **ETag** | `"abc123"` | 协商缓存指纹 |
| **Transfer-Encoding** | `chunked` | 分块传输 |

### 3.3 Content-Type 常见取值

| Content-Type | 使用场景 |
|--------------|----------|
| `application/json` | REST API、前后端分离（**最常用**） |
| `application/x-www-form-urlencoded` | 传统表单、`URLSearchParams` |
| `multipart/form-data` | 文件上传 |
| `text/html` | 服务端渲染页面 |
| `text/plain` | 纯文本 |

### 3.4 Authorization 与 Java 04 / shop-vue

Java 04 登录接口返回 JWT 后，前端存储并携带：

```javascript
// shop-vue axios 拦截器（简化）
axios.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})
```

对应 HTTP 原始头：

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIn0.xxx
```

**面试点**：`Bearer` 是认证方案名；还有 `Basic`（用户名密码 Base64）、`Digest` 等。

---

## 4. HTTP 方法：语义、幂等与安全性

### 4.1 方法一览

| 方法 | 语义 | 有 Body | 幂等 | 安全 | 典型用途 |
|------|------|---------|------|------|----------|
| **GET** | 获取资源 | 否（规范不推荐） | ✅ | ✅ | 查询列表、详情 |
| **POST** | 创建 / 提交 | 是 | ❌ | ❌ | 登录、下单、创建用户 |
| **PUT** | 全量替换资源 | 是 | ✅ | ❌ | 更新整个用户对象 |
| **PATCH** | 部分更新 | 是 | ❌* | ❌ | 只改 `name` 字段 |
| **DELETE** | 删除资源 | 可选 | ✅ | ❌ | 删除用户 |
| **HEAD** | 同 GET 但不返回 Body | 否 | ✅ | ✅ | 检查资源是否存在 |
| **OPTIONS** | 询问支持的方法 | 否 | ✅ | ✅ | **CORS 预检** |

> *PATCH 严格说是否幂等取决于实现；面试一般答「可能不幂等」。

### 4.2 幂等（Idempotent）是什么？

**定义**：同一请求执行 **1 次** 与执行 **N 次**，资源**最终状态相同**（副作用一致）。

| 示例 | 是否幂等 | 解释 |
|------|----------|------|
| `GET /api/users/1` | ✅ | 只读，不改变服务器状态 |
| `PUT /api/users/1` Body 完整对象 | ✅ | 多次 PUT 同样内容，结果一样 |
| `POST /api/users` | ❌ | 每调一次可能多创建一条记录 |
| `POST /api/login` | ❌ | 可能产生新 session / 刷新 token |
| `DELETE /api/users/1` | ✅ | 删第一次成功，再删 404，「资源不存在」状态一致 |

### 4.3 为什么 PUT 幂等而 POST 不幂等？（深入解释 ②）

**REST 设计意图**：

- **POST** 面向「**集合**」：`POST /api/users` → 「往用户列表里**加**一个」，每次加一个新的
- **PUT** 面向「**具体 URI**」：`PUT /api/users/1` → 「把 id=1 设成这个内容」，设十次结果相同

Spring Boot 示例（Java 04 风格）：

```java
@PostMapping("/api/users")
public Result<UserVO> create(@RequestBody UserDTO dto) { ... }  // 非幂等

@PutMapping("/api/users/{id}")
public Result<UserVO> update(@PathVariable Long id, @RequestBody UserDTO dto) { ... }  // 幂等

@DeleteMapping("/api/users/{id}")
public Result<String> delete(@PathVariable Long id) { ... }  // 幂等
```

### 4.4 GET vs POST 再深入（面试高频）

| 维度 | GET | POST |
|------|-----|------|
| 参数位置 | Query String | Body（或 Query，不推荐） |
| 书签 / 分享 | URL 含参数，可收藏 | 不可 |
| 浏览器缓存 | 可能被缓存 | 通常不缓存 |
| 长度限制 | URL 长度受限（约 2K～8K） | Body 可很大 |
| 语义 | **安全**查询 | **非安全**提交 |
| 刷新 / 重试 | 重复 GET 一般无害 | 重复 POST 可能重复下单 |

**注意**：「POST 比 GET 安全」是**误解**——两者都是明文 HTTP；安全靠 **HTTPS**（见 [05-HTTPS与TLS加密](./05-HTTPS与TLS加密.md)）。

---

## 5. HTTP 状态码：1xx～5xx 全分类

状态码是**三位数字**，按首位分类：

```mermaid
flowchart LR
    A[1xx 信息] --> B[2xx 成功]
    B --> C[3xx 重定向]
    C --> D[4xx 客户端错误]
    D --> E[5xx 服务端错误]
```

### 5.1 1xx 信息响应

| 状态码 | 含义 | 前端何时见到 |
|--------|------|--------------|
| **100 Continue** | 继续发送 Body | 大文件上传前，较少见 |
| **101 Switching Protocols** | 切换协议 | WebSocket 升级 |

### 5.2 2xx 成功

| 状态码 | 含义 | 典型场景 |
|--------|------|----------|
| **200 OK** | 成功 | GET 查询、POST 登录成功 |
| **201 Created** | 已创建 | POST 创建用户成功 |
| **204 No Content** | 成功但无 Body | DELETE 成功 |
| **206 Partial Content** | 部分内容 | 视频断点续传、Range 请求 |

**与 Result JSON 的关系**（重要）：

Spring Boot 常**无论业务成败都返回 HTTP 200**，用 Body 里 `code` 区分：

```json
HTTP/1.1 200 OK
{"code":1,"message":"用户名不能为空","data":null}
```

前端 Axios 拦截器要**同时看 HTTP 状态码和 body.code**（Vue 08 章约定 `code === 0` 为成功）。

### 5.3 3xx 重定向

| 状态码 | 含义 | 缓存 | 方法是否变 GET |
|--------|------|------|----------------|
| **301 Moved Permanently** | 永久搬迁 | 可能被缓存 | 老客户端可能变 GET |
| **302 Found** | 临时重定向 | 一般不长期缓存 | 可能变 GET |
| **304 Not Modified** | 未修改，用缓存 | 协商缓存 | - |
| **307 Temporary Redirect** | 临时重定向 | - | **保持原方法** |
| **308 Permanent Redirect** | 永久重定向 | - | **保持原方法** |

### 5.4 4xx 客户端错误

| 状态码 | 含义 | 前端处理建议 |
|--------|------|--------------|
| **400 Bad Request** | 参数错误 | 提示用户检查输入；对照后端校验消息 |
| **401 Unauthorized** | 未认证 | 跳转登录；清 token |
| **403 Forbidden** | 已认证但无权限 | 提示无权限，勿与 401 混淆 |
| **404 Not Found** | 资源不存在 | 检查 URL、路由、后端 `@GetMapping` 路径 |
| **405 Method Not Allowed** | 方法不对 | 例如对 GET 接口发了 POST |
| **408 Request Timeout** | 请求超时 | 重试或提示网络 |
| **409 Conflict** | 冲突 | 如重复注册用户名 |
| **413 Payload Too Large** | Body 太大 | 压缩上传或分片 |
| **415 Unsupported Media Type** | Content-Type 不对 | 检查是否忘了 `application/json` |
| **429 Too Many Requests** | 限流 | 退避重试 |

**401 vs 403（面试必背）**：

- **401**：「你是谁？」→ 没带 token 或 token 无效
- **403**：「我知道你是谁，但你不能干这个」→ 普通用户访问 admin 接口

### 5.5 5xx 服务端错误

| 状态码 | 含义 | 前端处理 |
|--------|------|----------|
| **500 Internal Server Error** | 服务器内部异常 | 友好提示；DevTools 看响应 body |
| **502 Bad Gateway** | 网关上游挂了 | Nginx 后面 Java 进程挂了 |
| **503 Service Unavailable** | 服务不可用 | 维护或过载 |
| **504 Gateway Timeout** | 网关等待超时 | 后端慢查询、死锁 |

### 5.6 状态码决策流程（Mermaid）

```mermaid
flowchart TD
    Start[收到响应] --> Q1{HTTP 状态码?}
    Q1 -->|2xx| Q2{body.code === 0?}
    Q2 -->|是| OK[业务成功]
    Q2 -->|否| BizFail[业务失败: 展示 message]
    Q1 -->|401| Login[跳转登录]
    Q1 -->|403| Deny[无权限提示]
    Q1 -->|404| NotFound[资源不存在]
    Q1 -->|5xx| ServerErr[服务器错误 + 上报]
    Q1 -->|网络错误| NetErr[Axios error.response 可能为 undefined]
```

---

## 6. 持久连接与 Keep-Alive

### 6.1 短连接 vs 持久连接

HTTP/1.0 默认：**一次请求 → 一次 TCP 连接 → 关闭**。

HTTP/1.1 默认：**持久连接（Persistent Connection）**。

```http
Connection: keep-alive
Keep-Alive: timeout=5, max=1000
```

| 模式 | 优点 | 缺点 |
|------|------|------|
| 短连接 | 实现简单 | 每次握手 + 慢启动，延迟高 |
| Keep-Alive | 复用 TCP，减少握手 | 占用连接数；HTTP/1.1 **队头阻塞** |

### 6.2 队头阻塞（Head-of-Line Blocking）

HTTP/1.1 上多个请求**复用同一 TCP 连接**，但响应必须**按请求顺序**返回。若第一个响应很慢，后面即使已就绪也要等——这就是 **HOL 阻塞**，也是 HTTP/2 **多路复用**要解决的问题。

---

## 7. HTTP/1.1 vs HTTP/2 vs HTTP/3

### 7.1 版本对比总表

| 特性 | HTTP/1.1 | HTTP/2 | HTTP/3 |
|------|----------|--------|--------|
| 传输层 | TCP | TCP | **QUIC（基于 UDP）** |
| 多路复用 | ❌（同连接串行） | ✅ 帧级并行 | ✅ 改进的 multiplex |
| 头部压缩 | ❌ 明文重复 | ✅ **HPACK** | ✅ **QPACK** |
| 服务器推送 | ❌ | ✅ Server Push（已少用） | - |
| 连接建立 | TCP 三次握手 + TLS | 同上（可 TLS 1.3 0-RTT） | **0-RTT / 1-RTT** 更快 |
| 队头阻塞 | 应用层 + TCP 层 | TCP 层仍阻塞 | 连接迁移、减少 TCP HOL |
| 浏览器支持 | 全部 | 全部现代浏览器 | 广泛支持 |
| 调试体感 | curl 默认可见明文 | 二进制帧，DevTools 解析 | 同 HTTP/2 |

### 7.2 HTTP/2 核心机制

```mermaid
flowchart TB
    subgraph HTTP2["HTTP/2 单连接"]
        S[Stream 1: GET /api/users]
        S2[Stream 3: GET /api/products]
        S3[Stream 5: POST /api/login]
    end
    TCP[一条 TCP 连接]
    S --> TCP
    S2 --> TCP
    S3 --> TCP
```

- **二进制分帧**：HEADERS 帧、DATA 帧
- **HPACK**：静态表 + 动态表压缩 Header，Cookie 重复发送体积大时效果明显
- **Stream ID**：奇数客户端发起，偶数服务端发起

### 7.3 HTTP/3 与 QUIC

- 把传输层从 TCP 换成 **QUIC**
- 连接 ID 不变，**切换网络（Wi-Fi → 4G）不断连**
- 丢包只阻塞对应 Stream，不拖整个连接

**面试一句话**：HTTP/2 解决了应用层队头阻塞，但 TCP 丢包仍阻塞所有 Stream；HTTP/3 用 QUIC 缓解。

### 7.4 前端需要做什么？

| 场景 | 你要做的 |
|------|----------|
| 开发 REST API | **几乎不用**选版本，Nginx/CDN 自动协商 |
| 性能优化 | 减少域名分片（HTTP/2 下多域名反而多连接） |
| 调试 | DevTools → Network → Protocol 列看 `h2` / `http/1.1` |
| 本地 demo | `localhost` 多为 HTTP/1.1，生产 HTTPS 常见 h2 |

---

## 8. RESTful 设计原则

REST（Representational State Transfer）是一种**架构风格**，不是协议。

### 8.1 六大约束（了解）

1. 客户端-服务器分离
2. **无状态**
3. 可缓存
4. 统一接口
5. 分层系统
6. 按需代码（可选）

### 8.2 统一接口：资源 + HTTP 方法

| 操作 | HTTP 方法 | URI 示例 | Java 04 风格 |
|------|-----------|----------|--------------|
| 列表 | GET | `/api/users` | `@GetMapping` |
| 详情 | GET | `/api/users/{id}` | `@GetMapping("/{id}")` |
| 创建 | POST | `/api/users` | `@PostMapping` |
| 全量更新 | PUT | `/api/users/{id}` | `@PutMapping` |
| 部分更新 | PATCH | `/api/users/{id}` | `@PatchMapping` |
| 删除 | DELETE | `/api/users/{id}` | `@DeleteMapping` |

### 8.3 RESTful 好 vs 非 REST 对比

| ❌ 非 REST | ✅ RESTful | 问题 |
|-----------|-----------|------|
| `POST /api/getUserById` | `GET /api/users/1` | 动词不应出现在 URI |
| `POST /api/deleteUser` | `DELETE /api/users/1` | 应用 HTTP 语义 |
| `GET /api/users/create` | `POST /api/users` | GET 不应有副作用 |

### 8.4 与 Spring Boot Result 的配合

```java
// 统一包装 — 见 Java 04 Result.java
public class Result<T> {
    private int code;      // 0 成功
    private String message;
    private T data;
}
```

**两种风格并存于业界**：

| 风格 | HTTP 状态码 | Body |
|------|-------------|------|
| **纯 REST** | 404 就是 404 | 错误体 `{error: "..."}` |
| **国内常见** | 多数 200 | `code !== 0` 表示业务失败 |

shop-vue 联调采用第二种；面试时能说清**为什么**：前端拦截器统一处理、网关层好包装。

---

## 9. 手把手：curl 实操（配合 Java 04 demo）

### 9.1 环境准备

```bash
# 终端 1：启动 Spring Boot
cd demo
./mvnw spring-boot:run

# 预期最后几行包含：
Started DemoApplication in 2.xxx seconds
```

### 9.2 GET 查询用户列表

```bash
curl -v http://localhost:8080/api/users
```

**预期输出（节选）**：

```text
> GET /api/users HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/8.x
> Accept: */*
>
< HTTP/1.1 200
< Content-Type: application/json
<
{"code":0,"message":"success","data":[{"id":1,"name":"张三","age":20},...]}
```

### 9.3 POST 创建用户（JSON Body）

```bash
curl -v -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d "{\"name\":\"李四\",\"age\":22}"
```

**预期**：

```text
< HTTP/1.1 200
{"code":0,"message":"success","data":{"id":3,"name":"李四","age":22}}
```

### 9.4 POST 登录拿 Token

```bash
curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"123456\"}"
```

**预期**：

```json
{"code":0,"message":"success","data":{"token":"eyJhbGciOiJIUzI1NiJ9..."}}
```

### 9.5 带 Authorization 访问受保护接口

```bash
TOKEN="eyJhbGciOiJIUzI1NiJ9..."   # 替换为上一步 token

curl -v http://localhost:8080/api/users/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json"
```

**预期**：`200` + 用户 JSON；若 token 无效 → `401` 或 `code !== 0`（取决于后端实现）。

### 9.6 DELETE 删除

```bash
curl -v -X DELETE http://localhost:8080/api/users/3
```

**预期**：

```json
{"code":0,"message":"success","data":"删除成功"}
```

### 9.7 只看响应头

```bash
curl -I http://localhost:8080/api/users
```

**预期**：

```text
HTTP/1.1 200
Content-Type: application/json
...
```

### 9.8 模拟错误场景

```bash
# 404 路径不存在
curl -v http://localhost:8080/api/not-exist

# 400 参数校验失败（空 name）
curl -v -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"\",\"age\":20}"
```

---

## 10. 手把手：Chrome DevTools Network 面板

### 10.1 打开与基本操作

1. `F12` → **Network**
2. 勾选 **Preserve log**（路由跳转后记录仍在）
3. 勾选 **Disable cache**（开发时避免强缓存干扰）
4. 过滤：**Fetch/XHR** 只看接口

### 10.2 分析一次 shop-vue 登录请求

启动 shop-vue（`npm run dev`），打开登录页，输入账号密码，在 Network 里点 `login` 请求：

| 面板区域 | 看什么 |
|----------|--------|
| **Headers → General** | Request URL、Method、Status Code |
| **Headers → Request Headers** | `Content-Type`、`Authorization`（登录前无） |
| **Headers → Response Headers** | `Content-Type`、CORS 相关 |
| **Payload** | 提交的 JSON body |
| **Preview / Response** | `Result` 结构 `{code, message, data}` |
| **Timing** | DNS、连接、Waiting (TTFB)、Content Download |

### 10.3 查看 HTTP 版本

Network 表头右键 → 勾选 **Protocol** → 看到 `http/1.1` 或 `h2`。

### 10.4 Copy as cURL

右键请求 → **Copy** → **Copy as cURL** → 粘贴到终端可复现（注意 Windows 引号转义）。

```mermaid
flowchart LR
    A[用户点击登录] --> B[Axios POST /api/login]
    B --> C[Network 面板]
    C --> D[看 Status / Headers / Response]
    D --> E[Pinia 存 token]
    E --> F[后续请求带 Authorization]
```

---

## 11. CORS 预检与 OPTIONS（与 HTTP 方法联动）

跨域 POST 带自定义头 `Authorization` 时，浏览器先发 **OPTIONS 预检**：

```http
OPTIONS /api/users HTTP/1.1
Origin: http://localhost:5173
Access-Control-Request-Method: POST
Access-Control-Request-Headers: authorization,content-type
```

服务器正确响应：

```http
HTTP/1.1 204 No Content
Access-Control-Allow-Origin: http://localhost:5173
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: authorization, content-type
```

详见 Java 04 `CorsConfig` 与 Vue 08 Vite proxy 章节。

---

## 12. 常见报错与排查表

| 现象 | 可能原因 | 怎么查 | 解决办法 |
|------|----------|--------|----------|
| `Network Error` / 无 status | 后端未启动、端口错、CORS 被拦 | curl 直连 8080 | 启动 demo；配 proxy 或 CORS |
| `404 Not Found` | URL 路径与 Controller 不一致 | 对照 `@GetMapping` | 改前端 baseURL 或后端路径 |
| `415 Unsupported Media Type` | 未设 `Content-Type: application/json` | DevTools Request Headers | Axios 设 `headers['Content-Type']` |
| `401 Unauthorized` | token 缺失或过期 | 看 Authorization 头 | 重新登录；拦截器挂 token |
| `403 Forbidden` | 有 token 但角色不够 | 后端权限注解 | 换账号或改权限 |
| `500 Internal Server Error` | 后端空指针、SQL 错误 | 看 Java 控制台栈 | 修后端；前端友好提示 |
| `CORS policy` 红字 | 响应无 Allow-Origin | OPTIONS 预检失败 | Java 04 CorsConfig |
| `ERR_CONNECTION_REFUSED` | 端口无进程 | `curl localhost:8080` | 启动 Spring Boot |
| `413 Payload Too Large` | 上传文件过大 | Nginx `client_max_body_size` | 压缩或分片上传 |
| `302` 跟着跳到登录页 | Session 方案未带 Cookie | 看 Set-Cookie | 改 JWT 或 withCredentials |
| 响应乱码 | charset 未声明 | Content-Type 缺 charset | 后端 `;charset=UTF-8` |
| HTTP/1.1 下接口慢 | 串行请求太多 | Timing 瀑布图 | 合并接口；升级 HTTP/2 |

---

## 13. 面试高频问答速记

### Q1：HTTP 和 HTTPS 区别？

HTTP 明文；HTTPS = HTTP + **TLS** 加密 + 身份校验。详见 [05 章](./05-HTTPS与TLS加密.md)。

### Q2：GET 能带 Body 吗？

规范**未禁止**但**不推荐**；很多服务器、缓存、代理会忽略 GET Body。查询参数放 Query String。

### Q3：什么是幂等？哪些方法幂等？

见 §4.2；GET、PUT、DELETE 通常幂等；POST 不幂等。

### Q4：301 和 302 区别？

301 **永久**，SEO 权重转移；302 **临时**。307/308 保持原 HTTP 方法。

### Q5：HTTP/2 为什么更快？

多路复用、HPACK 头部压缩、二进制分帧；仍跑在 TCP 上。

### Q6：前后端分离为什么常用 JSON + Result？

语言无关、易解析；`code/message/data` 统一业务错误，配合 Axios 拦截器。

---

## 14. 练习建议

### 基础题

1. 画出 HTTP 请求报文的四个组成部分，并各写一行示例。
2. 说出 `401` 和 `403` 的区别，各举一个 shop-vue 场景。
3. 为什么 `POST /api/orders` 不是幂等的？什么方法适合「取消订单」？
4. 列出 5 个 4xx 和 2 个 5xx 状态码及含义。
5. 用 curl 向 `http://localhost:8080/api/users` 发 GET，说明 `-v` 输出里 `>` 和 `<` 行分别是什么。

### 进阶题

6. 解释 HTTP/1.1 队头阻塞；HTTP/2 如何缓解；HTTP/3 又做了什么。
7. 设计一套 RESTful 的「商品」接口（列表、详情、创建、更新、删除），写出 Method + URI 表格。
8. 说明 `Result<T>` 始终返回 HTTP 200 的利弊；若改成 404 返回真实 404，前端 Axios 要怎么改？
9. 在 DevTools 里找到一次 OPTIONS 预检，抄下 Request / Response 关键头。
10. 对比 `Connection: keep-alive` 与 HTTP/2 多路复用的关系。

### 挑战题

11. 写 bash 脚本：登录拿 token → 带 token 创建用户 → 删除该用户，全程 curl。
12. 读 Spring Boot 某 Controller 源码，对照 Network 面板验证每个 Header。

---

## 15. 练习参考答案（基础 + 进阶节选）

### 基础题 1 答案

四部分：**请求行** `GET /api/users HTTP/1.1`；**请求头** `Host: localhost:8080`；**空行**；**Body**（GET 通常无）。

### 基础题 2 答案

- **401**：未登录访问 `/api/users`，无 `Authorization` → 跳转登录。
- **403**：普通用户访问 `/api/admin/stats`，有 token 但非 admin → 提示无权限。

### 基础题 3 答案

每次 POST 可能创建新订单，状态「多一张订单」；取消订单用 **PUT/PATCH** 改状态为 cancelled，或 **DELETE**（视业务是否幂等设计）。

### 基础题 5 答案

- `>` 开头：客户端发出的请求行和请求头
- `<` 开头：服务器返回的状态行和响应头

### 进阶题 6 答案

HTTP/1.1 同连接响应顺序绑定；HTTP/2 多 Stream 并行；HTTP/3 QUIC 避免 TCP 丢包阻塞全连接，且支持连接迁移。

### 进阶题 7 答案（示例）

| 操作 | Method | URI |
|------|--------|-----|
| 列表 | GET | `/api/products?page=1&size=20` |
| 详情 | GET | `/api/products/{id}` |
| 创建 | POST | `/api/products` |
| 全量更新 | PUT | `/api/products/{id}` |
| 部分更新 | PATCH | `/api/products/{id}` |
| 删除 | DELETE | `/api/products/{id}` |

### 进阶题 8 答案

**利**：前端统一在 200 分支解析 `code`。**弊**：不符合 REST 纯语义，监控难靠 HTTP 码统计错误。若改真实 404：Axios 响应拦截器在 `error.response.status === 404` 分支处理，不能再假设只有 2xx 进 success 回调。

---

## 16. 学完标准

完成本章后，你应能：

- [ ] **独立画出** HTTP 请求/响应报文结构，说清起始行、头、Body
- [ ] **解释** Host、Content-Type、Accept、Authorization 的作用并会在 DevTools 里找到它们
- [ ] **区分** GET/POST/PUT/DELETE/PATCH 语义，说出幂等与安全方法概念
- [ ] **列举** 1xx～5xx 每类至少 2 个状态码，**区分 401 与 403**
- [ ] **对比** HTTP/1.1、HTTP/2（多路复用、HPACK）、HTTP/3（QUIC）核心差异
- [ ] **说明** Keep-Alive 与队头阻塞的关系
- [ ] **设计** RESTful URI，并对照 Java 04 Controller 注解
- [ ] **用 curl -v** 完成 GET/POST/DELETE + Bearer Token，并对照预期 JSON
- [ ] **在 Network 面板** 完成一次登录请求全链路分析（Timing、Headers、Response）
- [ ] **口述** `Result<T>` 与 HTTP 状态码的配合方式（面试 2 分钟版）

---

## 下一章预告

下一章 [05-HTTPS与TLS加密](./05-HTTPS与TLS加密.md) 将回答：

- 为什么明文 HTTP 不够，HTTPS 如何解决窃听与篡改
- 对称加密 vs 非对称加密、CA 证书链如何建立信任
- TLS 1.2 与 1.3 握手差异（几 RTT、为何更快）
- 443 端口、HTTP→HTTPS 重定向、混合内容、HSTS
- DevTools **Security** 面板如何看证书
- shop-vue 登录 token 在 HTTPS 下如何更安全（Secure、HttpOnly、防中间人）

**建议**：本章 curl 默认 `http://localhost`；05 章会讲生产环境必须 `https://`，并与 Nginx / Spring Boot 配置衔接。

---

## 附录：HTTP 方法 + 状态码速查卡

```text
┌─────────────────────────────────────────────────────────┐
│  GET → 查   POST → 建/提交   PUT → 全量改   PATCH → 部分改 │
│  DELETE → 删   OPTIONS → CORS 预检   HEAD → 只要头      │
├─────────────────────────────────────────────────────────┤
│  2xx 成了   3xx 跳转   4xx 你错了   5xx 服务器挂了       │
│  401 没登录  403 没权限  404 没资源  500 后端异常         │
└─────────────────────────────────────────────────────────┘
```

**交叉链接汇总**：

| 主题 | 文档 |
|------|------|
| HTTP 入门、跨域初识 | [HTML 10](../HTML%20CSS%20JS/10-浏览器HTTP网络与Web基础.md) |
| Axios 联调、token 拦截器 | [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) / [React 08](../React/08-Axios网络请求与前后端联调.md) |
| Result、CORS、Controller | [Java 04](../../后端学习/Java/04-SpringBoot核心开发.md) |
| HTTPS 与 TLS | [05-HTTPS与TLS加密](./05-HTTPS与TLS加密.md) |

---

## 17. 消息体编码：Content-Length 与 chunked

### 17.1 Content-Length

固定长度 Body，服务器提前告知字节数：

```http
Content-Length: 58

{"code":0,"message":"success","data":{"id":1,"name":"张三"}}
```

前端 Axios 据此知道何时收完；**Keep-Alive** 连接上必须读满 Content-Length 才能解析下一条消息。

### 17.2 Transfer-Encoding: chunked

流式分块，长度未知时（SSR 流、大 JSON 生成中）使用：

```http
Transfer-Encoding: chunked

5\r\n
{"id"\r\n
...
0\r\n
\r\n
```

HTTP/2 用 DATA 帧代替 chunked 语义，但 DevTools 仍可能显示 chunked 等价行为。

### 17.3 与 gzip 的关系

```http
Content-Encoding: gzip
```

Body 是压缩后的二进制；浏览器自动解压再给 JS。`Accept-Encoding: gzip, deflate, br` 在请求头里声明客户端能力。

---

## 18. 实体头与条件请求（和缓存衔接）

| Header | 方向 | 作用 |
|--------|------|------|
| **ETag** | 响应 | 资源版本指纹 `"abc123"` |
| **If-None-Match** | 请求 | 带上次的 ETag |
| **Last-Modified** | 响应 | 最后修改时间 |
| **If-Modified-Since** | 请求 | 协商缓存询问 |
| **Cache-Control** | 响应 | `max-age=3600` / `no-cache` |
| **Expires** | 响应 | 绝对过期时间（HTTP/1.0 遗留） |
| **Vary** | 响应 | 按 Accept-Encoding 等区分缓存 |

**304 Not Modified**：Body 为空，浏览器用本地副本——Network 面板显示 `(disk cache)` 或 `304`。

```bash
# 第一次
curl -v http://localhost:8080/api/users -H "Accept: application/json"

# 第二次带 ETag（从上次响应头复制）
curl -v http://localhost:8080/api/users \
  -H "Accept: application/json" \
  -H 'If-None-Match: "xyz"'
```

---

## 19. HTTP 语义扩展：Range、WebSocket 升级

### 19.1 Range 请求（206）

下载大文件、视频拖拽：

```http
GET /video.mp4 HTTP/1.1
Range: bytes=0-1023
```

响应 `206 Partial Content` + `Content-Range: bytes 0-1023/5000`。

### 19.2 Connection 头与升级

```http
Connection: Upgrade
Upgrade: websocket
```

`101 Switching Protocols` 后不再是 HTTP 语义，而是 WebSocket 帧（shop 实时通知、在线客服等场景）。

---

## 20. Spring Boot 与 HTTP 映射深入

### 20.1 注解与 HTTP 一一对应

```java
@RestController
@RequestMapping("/api/users")
public class UserController {

    @GetMapping                    // GET    /api/users
    public Result<List<UserVO>> list() { ... }

    @GetMapping("/{id}")           // GET    /api/users/1
    public Result<UserVO> getById(@PathVariable Long id) { ... }

    @PostMapping                   // POST   /api/users
    public Result<UserVO> create(@Valid @RequestBody UserDTO dto) { ... }

    @PutMapping("/{id}")           // PUT    /api/users/1
    public Result<UserVO> update(@PathVariable Long id, @RequestBody UserDTO dto) { ... }

    @PatchMapping("/{id}")        // PATCH  /api/users/1
    public Result<UserVO> patch(@PathVariable Long id, @RequestBody Map<String, Object> fields) { ... }

    @DeleteMapping("/{id}")        // DELETE /api/users/1
    public Result<String> delete(@PathVariable Long id) { ... }
}
```

### 20.2 @RequestBody vs @RequestParam

| 注解 | 数据位置 | Content-Type |
|------|----------|--------------|
| `@RequestParam` | Query `?page=1` 或表单 | `application/x-www-form-urlencoded` |
| `@RequestBody` | Body JSON | `application/json` |

前端 Axios：

```javascript
// Query
axios.get('/api/users', { params: { page: 1, size: 10 } })

// JSON Body
axios.post('/api/users', { name: '王五', age: 25 })
```

### 20.3 全局异常与 HTTP 状态码（两种风格）

**风格 A**：始终 200 + Result.fail（Java 04 demo 常见）

```java
@ExceptionHandler(Exception.class)
public Result<String> handle(Exception e) {
    return Result.fail(e.getMessage());
}
```

**风格 B**：REST 纯语义

```java
@ResponseStatus(HttpStatus.NOT_FOUND)
@ExceptionHandler(UserNotFoundException.class)
public ErrorBody notFound(UserNotFoundException e) {
    return new ErrorBody("USER_NOT_FOUND", e.getMessage());
}
```

前端需对应调整：风格 B 用 `catch (error) { if (error.response?.status === 404) ... }`。

---

## 21. Axios 错误对象与 HTTP 状态码映射

```javascript
axios.get('/api/users/999').catch((error) => {
  if (error.response) {
    // 服务器回了 4xx/5xx
    console.log(error.response.status)   // 404
    console.log(error.response.data)     // Result JSON 或 Spring 默认错误页
  } else if (error.request) {
    // 请求发出去了但没收到响应（网络断、CORS 被拦）
    console.log('Network / CORS')
  } else {
    console.log(error.message)
  }
})
```

| error.response.status | 典型处理 |
|-----------------------|----------|
| 400 | 展示 `message` 或字段级错误 |
| 401 | 清 token，路由 `/login` |
| 403 | Toast「无权限」 |
| 404 | 空状态页 |
| 500 | 「服务繁忙，请稍后」 |
| undefined | 检查后端是否启动、CORS、HTTPS 混合内容 |

---

## 22. 手把手续：curl 高级参数

### 22.1 保存 Cookie（Session 方案对比 JWT）

```bash
curl -c cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"123456\"}"

curl -b cookies.txt http://localhost:8080/api/users
```

JWT 方案则应用 `-H "Authorization: Bearer ..."` 而非 Cookie 文件。

### 22.2 跟随重定向

```bash
curl -L -v http://example.com
# 跟踪 301/302 的 Location
```

### 22.3 只发 HEAD

```bash
curl -I http://localhost:8080/api/users
# 等价 curl -X HEAD
```

### 22.4 自定义 Header 批量

```bash
curl http://localhost:8080/api/users \
  -H "Accept: application/json" \
  -H "Accept-Language: zh-CN" \
  -H "X-Request-Id: debug-001"
```

### 22.5 表单提交（对比 JSON）

```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=zhang&password=123456"
```

Spring 需 `@RequestParam` 而非 `@RequestBody` 接收。

---

## 23. DevTools 续：Timing 瀑布图精读

| 阶段 | 含义 | 优化方向 |
|------|------|----------|
| **Queueing** | 浏览器排队（HTTP/1.1 同域连接数限制） | HTTP/2、减少域名 |
| **DNS Lookup** | 域名解析 | DNS 缓存、预解析 |
| **Initial connection** | TCP + TLS 握手 | Keep-Alive、TLS 1.3 |
| **SSL** | TLS 握手 | 见 05 章 |
| **Request sent** | 请求已发出 | - |
| **Waiting (TTFB)** | 等到第一个字节 | 后端慢查询、无索引 |
| **Content Download** | 下载 Body | 压缩、减小 JSON |

shop-vue 商品列表慢：先看 **Waiting** 高 → 查 Java 接口；**Download** 高 → 分页、字段裁剪。

---

## 24. HTTP/2 Server Push（了解即可）

HTTP/2 允许服务器主动推资源：

```text
客户端 GET /index.html
服务器 PUSH /style.css、/app.js
```

因缓存策略复杂、推送常不准，**Chrome 已移除 Push 支持**。面试知道「曾经存在、现已式微」即可。

---

## 25. 场景题：从浏览器到 Spring Boot 完整链路

**场景**：shop-vue 点击「用户列表」，Network 看到 `GET /api/users` 200。

```mermaid
sequenceDiagram
    participant Vue as ProductList.vue
    participant Axios as Axios 实例
    participant Vite as Vite Dev Server
    participant SB as UserController
    participant Svc as UserService
    Vue->>Axios: get('/api/users')
    Axios->>Vite: GET http://localhost:5173/api/users
    Note over Vite: proxy 转发
    Vite->>SB: GET http://localhost:8080/api/users
    SB->>Svc: findAll()
    Svc-->>SB: List User
    SB-->>Vite: 200 Result JSON
    Vite-->>Axios: 200
    Axios-->>Vue: response.data
    Vue->>Vue: 渲染列表
```

逐步对照：

1. **General → Request URL**：开发环境是 `5173/api/users`（proxy），不是 8080
2. **Request Headers → Accept**：`application/json`
3. **Response → Preview**：`code === 0`，`data` 为数组
4. 若 401：检查 Pinia 里 token 是否过期、拦截器是否执行

---

## 26. 更多练习与参考答案

### 练习 13

用 curl 模拟缺少 Content-Type 的 POST，观察 Spring Boot 返回 415 还是 400。

**答案**：通常 **415 Unsupported Media Type**（无法解析 Body）；若后端未严格校验可能 400。

### 练习 14

解释为何 HTTP/2 下「域名分片」（img1.example.com、img2.example.com）反而可能变慢。

**答案**：HTTP/1.1 下分片绕过每域 6 连接限制；HTTP/2 单连接多路复用，多域名 = 多条 TCP+TLS 握手，开销更大。

### 练习 15

写一段 Axios 响应拦截器：HTTP 200 且 `code !== 0` 时 reject，并统一 Toast。

**答案**：

```javascript
api.interceptors.response.use(
  (response) => {
    const { code, message } = response.data ?? {}
    if (code !== 0) {
      ElMessage.error(message || '业务失败')
      return Promise.reject(new Error(message))
    }
    return response.data
  },
  (error) => {
    const status = error.response?.status
    if (status === 401) router.push('/login')
    return Promise.reject(error)
  }
)
```

---

## 27. 面试模拟：2 分钟 HTTP 口述稿

> HTTP 是应用层无状态协议，报文分起始行、头、Body。方法有 GET 查、POST 提交、PUT 全量改、DELETE 删；GET/PUT/DELETE 幂等。状态码 2 成功、3 跳转、4 客户端错、5 服务端错；401 未认证、403 无权限。HTTP/1.1 默认 Keep-Alive 但有队头阻塞；HTTP/2 多路复用和 HPACK；HTTP/3 基于 QUIC。我们项目 Spring Boot 返回 Result JSON，Axios 带 Bearer token，开发用 Vite proxy 解决跨域。

---

## 28. 本章知识点索引

| § | 关键词 |
|---|--------|
| 2 | 起始行、Header、Body、CRLF |
| 3 | Host、Content-Type、Accept、Authorization |
| 4 | 幂等、安全方法、REST 动词 |
| 5 | 1xx～5xx、401/403、304 |
| 6 | Keep-Alive、队头阻塞 |
| 7 | HTTP/2 multiplex、HPACK、HTTP/3 QUIC |
| 8 | RESTful URI 设计 |
| 9-10 | curl、DevTools Network |
| 20-21 | Spring Boot、Axios 错误处理 |
