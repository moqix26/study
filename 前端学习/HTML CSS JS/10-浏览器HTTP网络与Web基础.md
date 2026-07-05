# 浏览器、HTTP、网络与 Web 基础

> **系统学习计算机网络**（OSI/TCP、三次握手、HTTP/2、HTTPS、DNS、CORS 原理）请学 [计算机网络系列](../计算机网络/00-学习路线图与说明.md)（00～07）。  
> 本章是 **Web 开发必知的网络入门**：能看 Network 面板、懂状态码和跨域现象；**深入原理**在 [计网 04 — HTTP 协议深入](../计算机网络/04-HTTP协议深入.md)。  
> 前端发请求实战：[09 fetch](./09-JavaScript异步编程网络请求与本地存储.md) → [Vue 08 Axios](../Vue/08-Axios网络请求与前后端联调.md)。

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0 导读、FAQ 12 题、闭卷自测、费曼检验；强化链 计网 04 / Vue 08 -->

## 0. 读前导读（零基础也能跟上）

> **读者假设**：你已会用 [09 章 fetch](./09-JavaScript异步编程网络请求与本地存储.md) 发请求并在 Console 见过 CORS 红字。本章不重复造 TCP/IP 轮子，而是教你在 **Network 面板**里当侦探——和 [计网 04](../计算机网络/04-HTTP协议深入.md) 同一套 HTTP，视角从「开发者排查」出发。

### 0.1 用一句话弄懂本章

**一句话**：浏览器按 **URL** 找到服务器，用 **HTTP** 一问一答拿 HTML/JSON；**状态码**告诉你成没成，**CORS/缓存/Cookie** 解释「为什么有时拿不到、为什么有时还是旧的」——Vue 08 联调卡壳时，80% 要先打开 Network。

**生活类比**：

| 本章概念 | 类比 | 深入阅读 |
|---------|------|----------|
| **URL** | 快递完整地址（协议+门牌+房间号） | 计网 04 §URL |
| **HTTP 请求/响应** | 寄件单 + 回执 | [计网 04](../计算机网络/04-HTTP协议深入.md) 全文 |
| **状态码 404/401** | 回执上「查无此人/未带证件」 | 本章 §7 + 计网 04 |
| **CORS** | 浏览器保安：外域包裹不让进 JS 房间 | 计网 04 CORS 节 |
| **Network 面板** | 快递跟踪 App | 本章 §24、§35 |

---

### 0.2 你需要提前知道什么

| 水平 | 建议 |
|------|------|
| 刚学 fetch | 先 [09 章](./09-JavaScript异步编程网络请求与本地存储.md) §7c，再来本章对照 Network |
| 要面试计网 | 本章扫一遍后精读 [计网 04](../计算机网络/04-HTTP协议深入.md)、[05 HTTPS](../计算机网络/05-HTTPS与TLS加密.md) |
| Vue 08 联调 | 本章 §20 CORS + §35 Network 步骤 + Vue 08 代理配置 |
| 完全零基础 | 可先看 §3 URL、§7 状态码，再回头补 09 |

---

### 0.3 本章知识地图（☐→☑）

- [ ] 能拆解 URL 四部分（协议、域名、路径、查询）
- [ ] 熟记 200/301/304/400/401/403/404/500 含义与排查方向
- [ ] 理解同源策略与 CORS 报错在说什么
- [ ] 区分 Cookie / Session / Token 使用场景
- [ ] 会用 Network 看 Method、Status、Headers、Timing
- [ ] 知道强缓存 vs 协商缓存对「刷新仍旧」的影响
- [ ] 闭卷自测 ≥ 8/10

---

### 0.4 建议学习时长与节奏

| 阶段 | 时间 |
|------|------|
| URL + HTTP 基础 §3～§9 | 1.5 小时 |
| 状态码 + 缓存 §7、§11～§12、§26 | 2 小时 |
| 同源/CORS §20 | 1 小时 |
| Network 实操 §24、§35 | 1.5 小时 |
| 安全与性能速览 §31～§33 | 1 小时 |
| 自测 | 0.5 小时 |

---

### 0.5 学完本章你能做什么（可验证）

1. 打开 Network，对一个 404 接口说出至少两个可能原因（路径错、方法错、后端未部署）。
2. 看到 CORS 错误，能判断该找前端改代理还是后端加 `Access-Control-Allow-Origin`。
3. 向 [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 联调：对照 Request Headers 里是否带了 `Authorization`。

---

### 0.6 核心术语三件套

**术语（HTTP 超文本传输协议）**：浏览器与服务器之间的应用层「对话规则」；无状态，靠 Cookie/Token 记登录。
**生活类比**：餐厅点菜的固定流程——第几号桌、要什么菜、厨房回「好了/没货」。
**为什么重要**：fetch/axios 全是 HTTP；不懂状态码和头字段无法联调。
**本章用到的地方**：§4～§9；详 [计网 04](../计算机网络/04-HTTP协议深入.md)。

**术语（同源策略 Same-Origin Policy）**：协议+域名+端口三者相同才默认同源；否则 JS 读响应受限制。
**生活类比**：小区门禁只认本小区住户——`localhost:5500` 和 `localhost:8080` 算不同小区。
**为什么重要**：跨域报错是前端 daily issue；不是 bug，是浏览器安全设计。
**本章用到的地方**：§20。

**术语（CORS 跨域资源共享）**：服务器通过响应头告诉浏览器「允许哪几个源读我」。
**生活类比**：外小区访客要在门卫处登记白名单才能进。
**为什么重要**：Vue 08 开发环境 proxy、生产 Nginx、Spring Boot `@CrossOrigin` 都在配这个。
**本章用到的地方**：§20；[Vue 08](../Vue/08-Axios网络请求与前后端联调.md)；[计网 04](../计算机网络/04-HTTP协议深入.md)。

---

## 1. 为什么前端也要学这些基础

很多初学者以为前端只要会写页面和 JS 就够了。

但你后面一定会不断碰到这些问题：

- 为什么接口请求失败
- 为什么有跨域
- 为什么缓存没更新
- 为什么页面第一次打开慢

这些问题背后往往和：

- 浏览器
- HTTP
- 网络基础

有关。

## 2. 浏览器是怎么工作的基础认知

你输入一个网址后，浏览器大致会做这些事：

1. 解析 URL
2. 发起网络请求
3. 收到 HTML
4. 解析 HTML
5. 解析 CSS
6. 执行 JavaScript
7. 渲染页面

这不是全部细节，但足够帮助你建立大图景。

## 3. URL 是什么

例如：

```text
https://www.example.com/products?id=1
```

你可以拆成：

- 协议：`https`
- 域名：`www.example.com`
- 路径：`/products`
- 查询参数：`?id=1`

## 4. HTTP 是什么

HTTP 是浏览器和服务器通信的协议。

你现在可以把它理解成：

- 前端请求后端时说的话和规则

## 5. 请求和响应

### 请求

浏览器发给服务器。

### 响应

服务器回给浏览器。

这是一来一回的通信过程。

## 6. 常见请求方法

### GET

查询数据。

### POST

提交数据。

### PUT

更新数据。

### DELETE

删除数据。

## 7. 状态码

前端必须认识常见状态码。

### `200`

请求成功。

### `201`

创建成功。

### `400`

请求参数有问题。

### `401`

未登录或认证失败。

### `403`

没有权限。

### `404`

资源不存在。

### `500`

服务器内部错误。

## 8. 请求头和响应头基础认知

请求头常见内容：

- `Content-Type`
- `Authorization`

响应头常见内容：

- `Content-Type`
- `Cache-Control`

## 9. Content-Type 基础认知

这个字段很常见。

它表示：

- 传输内容是什么格式

常见值：

- `application/json`
- `text/html`
- `multipart/form-data`

## 10. JSON 为什么这么常见

因为前后端分离里最常见的数据交换格式就是 JSON。

优点：

- 结构清晰
- 易于解析
- 语言无关

## 11. 浏览器缓存基础认知

浏览器为了提高性能，会缓存一部分资源。

常见好处：

- 页面更快
- 减少重复请求

但也会带来问题：

- 更新后用户看到旧资源

## 12. 强缓存和协商缓存基础印象

你现在先知道这两个概念即可：

- 强缓存：直接用本地缓存，不发请求
- 协商缓存：发请求问服务器资源有没有变

## 13. Cookie、Session、Token 基础认知

这三个词前端后端都会频繁看到。

### Cookie

浏览器可存的一小段数据。

### Session

更偏服务端会话状态。

### Token

前后端分离里非常常见。

你现在先知道它们都和登录态有关。

## 14. 跨域再细一点

当前端页面和接口地址不满足同源策略时，浏览器可能拦截请求结果。

你现在重点知道：

- 跨域是浏览器限制
- 服务端和代理层都可能参与解决

## 15. HTTPS 是什么

HTTPS 可以理解为更安全的 HTTP。

它主要带来：

- 加密传输
- 更安全的连接

## 16. DNS 基础认知

DNS 的作用可以简单理解为：

- 把域名解析成 IP 地址

你平时访问：

- `www.example.com`

浏览器最终要找到真实服务器地址。

## 17. 前端为什么要会看 Network 面板

浏览器开发者工具的 Network 面板很重要。

你至少要会看：

- 请求地址
- 请求方法
- 状态码
- 请求头
- 响应数据
- 请求耗时

## 18. 性能基础认知

页面变慢可能和很多因素有关：

- 图片太大
- 请求太多
- 资源未压缩
- JavaScript 太重
- CSS 阻塞渲染

## 19. Web 基础中的高频名词

你以后会反复看到这些词：

- CDN
- 域名
- IP
- 端口
- 协议
- 请求头
- 响应头
- 状态码

现在不必全部钻到底，但都要有基础印象。

## 20. 同源策略与跨域详解

### 什么是同源

协议、域名、端口**三者都相同**才叫同源。

| URL A | URL B | 是否同源 |
|-------|-------|----------|
| `http://a.com/page1` | `http://a.com/page2` | 是 |
| `http://a.com` | `https://a.com` | 否（协议不同） |
| `http://a.com` | `http://b.com` | 否（域名不同） |
| `http://a.com:80` | `http://a.com:8080` | 否（端口不同） |

### 跨域时浏览器报什么错

Console 常见：

```text
Access to fetch at 'http://api.example.com' from origin 'http://localhost:5500'
has been blocked by CORS policy
```

**重点**：请求可能已到达服务器，是浏览器**拒绝把响应交给 JS**。

### 常见解决方式（了解）

1. **后端配置 CORS**：响应头 `Access-Control-Allow-Origin`
2. **开发代理**：本地 dev server 转发 `/api` 到后端
3. **JSONP**（老方案，仅 GET，了解即可）

---

## 21. 一次 HTTP 请求长什么样

### 请求示例（GET）

```http
GET /api/users?page=1 HTTP/1.1
Host: api.example.com
Accept: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

### 响应示例（200）

```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Cache-Control: max-age=3600

{"code":0,"data":[{"id":1,"name":"张三"}]}
```

前端用 `fetch` 时主要关心：**状态码**、**Content-Type**、**body 里的 JSON**。

### 21.1 请求/响应报文逐行读（对照计网 04）

| 行号/字段 | 含义 | 改错会怎样 |
|-----------|------|------------|
| `GET /api/users?page=1 HTTP/1.1` | 请求行：方法 + 路径 + 协议版本 | 方法错 → 405；路径错 → 404 |
| `Host: api.example.com` | 虚拟主机必填（HTTP/1.1） | 缺 Host → 部分服务器 400 |
| `Accept: application/json` | 客户端期望的响应格式 | 后端忽略可能仍返回 JSON |
| `Authorization: Bearer ...` | 携带 JWT；[Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 拦截器自动加 | 缺 token → 401 |
| `HTTP/1.1 200 OK` | 状态行：协议 + 状态码 + 短语 | 看第一位判断成功/重定向/客户端错/服务端错 |
| `Content-Type: application/json` | body 是 JSON 文本 | 当 text/html parse → 报错 |
| `Cache-Control: max-age=3600` | 强缓存 1 小时 | 发版未改文件名 → 用户看旧 JS |
| `{"code":0,"data":[...]}` | 响应 body；[07 章](./07-JavaScript流程控制函数对象数组与ES6基础.md) 解构处理 | 字段名与前端不一致 → undefined |

完整报文结构、持久连接、HTTP/2 差异见 [计网 04 — HTTP 协议深入](../计算机网络/04-HTTP协议深入.md)。

---

## 22. GET 与 POST 区别（面试常问）

| | GET | POST |
|---|-----|------|
| 参数位置 | URL 查询字符串 `?a=1` | 通常在 body |
| 缓存 | 可能被缓存 | 一般不缓存 |
| 书签 | 可收藏 URL | 不可 |
| 语义 | **查询**数据 | **提交**数据 |
| 长度 | URL 有长度限制 | body 可较大 |

---

## 23. Cookie / Session / Token 对比

| | Cookie | Session | Token（如 JWT） |
|---|--------|---------|-----------------|
| 存哪 | 浏览器（也可 HttpOnly 服务端设） | 服务端内存/数据库 | 客户端常存 localStorage |
| 每次请求 | 自动带上 Cookie | 靠 SessionId | 手动放 `Authorization` 头 |
| 跨域 | 受同源限制较严 | - | 前后端分离常用 |
| 安全 | 要设 `HttpOnly` `Secure` | 服务端控制 | 防 XSS 泄露 token |

初学登录流程：**POST 登录 → 拿 token → localStorage 存 → 以后请求带 Authorization**。

### 23.1 与 Vue 08 / Java 04 联调时的登录态路径

```
用户提交表单
  → [09 fetch POST](./09-JavaScript异步编程网络请求与本地存储.md) 或 [Vue 08 axios.post](../Vue/08-Axios网络请求与前后端联调.md)
  → [Java 04](../../后端学习/Java/04-SpringBoot核心开发.md) 校验账号，返回 JWT
  → 前端 localStorage.setItem('auth', JSON.stringify({ token }))
  → 后续请求 Request Headers: Authorization: Bearer <token>
  → 401 时清 token 跳登录（Vue 08 响应拦截器）
```

| 步骤 | Network 里看什么 | 对应 [计网 04](../计算机网络/04-HTTP协议深入.md) |
|------|-------------------|--------------------------------------------------|
| 登录 POST | Status 200，Response body 含 token | POST body + JSON Content-Type |
| 带 token GET | Request Headers 有 Authorization | 自定义头、非简单请求可能预检 OPTIONS |
| 401 | 无权限或 token 过期 | 401 Unauthorized 语义 |
| Cookie 方案 | Set-Cookie 响应头，后续自动 Cookie 请求头 | HttpOnly、SameSite 属性 |

---

## 24. Network 面板实操步骤

1. `F12` → **Network** → 勾选 **Preserve log**（跳转后仍保留）
2. 刷新页面，看资源列表：
   - **Document**：HTML
   - **Stylesheet**：CSS
   - **Script**：JS
   - **Img**：图片
   - **Fetch/XHR**：接口
3. 点一条请求看：
   - **Headers**：Request URL、Method、Status Code、Request/Response Headers
   - **Preview / Response**：JSON 格式化查看
   - **Timing**：排队、DNS、等待服务器、下载各阶段耗时

### 根据状态码快速判断

| 状态码 | 你先查什么 |
|--------|------------|
| `(failed)` 红色 | 网络断、CORS、URL 错 |
| 404 | 路径、接口是否存在 |
| 401 | token 是否带上、是否过期 |
| 500 | 后端日志，看 Response body |

---

## 25. 浏览器渲染流程（简化）

```text
HTML → DOM 树
CSS  → CSSOM 树
     → 合成渲染树 → 布局(Layout) → 绘制(Paint) → 显示
JS 可能阻塞 HTML 解析（除非 defer/async）
```

知道即可：**CSS 放 head、JS 放底部或 defer**，有助于首屏更快。

---

## 26. 缓存：强缓存 vs 协商缓存

| 类型 | 机制 | 常见响应头 |
|------|------|------------|
| 强缓存 | 未过期直接用本地，**不发请求** | `Cache-Control: max-age=31536000` |
| 协商缓存 | 过期后带标识问服务器是否更新 | `ETag` / `Last-Modified` |

前端发版后用户看到旧页面？让运维/后端改缓存策略，或文件名加 hash（构建工具会做）。

---

## 27. 分级练习

**基础**：打开任意网站，Network 里找出 5 种类型资源  
**进阶**：用 jsonplaceholder 接口，看 Request/Response 完整头  
**挑战**：画一张「从输入 URL 到页面显示」的流程图（可手绘拍照）

---

## 28. FAQ

**Q1：前端能绕过跨域吗？**  
不能从浏览器端「关掉」同源策略；必须服务端 CORS、开发代理或 Nginx 转发。详见 [计网 04 CORS](../计算机网络/04-HTTP协议深入.md)。

**Q2：HTTP 和 HTTPS 默认端口？**  
80 和 443；URL 省略端口时用默认。

**Q3：JSON 和 JavaScript 对象一样吗？**  
很像，但 JSON 键必须双引号，不能有函数、undefined、注释。

**Q4：304 是错误吗？**  
不是。协商缓存命中，表示资源未变，用本地副本——发版后若仍旧，查强缓存/文件名 hash。

**Q5：401 和 403 区别？**  
401 未登录/ token 无效；403 已认证但无权限。Vue 08 联调时先查 Authorization 头。

**Q6：GET 和 POST 本质区别？**  
语义：GET 取资源（幂等），POST 提交/创建；POST 可有 body，GET 通常无 body。细节 [计网 04](../计算机网络/04-HTTP协议深入.md)。

**Q7：Cookie 和 localStorage 存 token 怎么选？**  
Cookie 可 HttpOnly 防 XSS 读；localStorage 实现简单但易被脚本窃取。前后端分离常见 Bearer Token + HTTPS。

**Q8：CORS 预检 OPTIONS 是什么？**  
非简单请求前浏览器先发 OPTIONS 问服务器是否允许；后端需正确响应 Allow-Methods/Headers。

**Q9：Network 里 `(failed)` 和 404 不同？**  
failed 常是 DNS/断网/CORS/混合内容；404 是服务器明确说「没有这资源」。

**Q10：本章和 [计网 04](../计算机网络/04-HTTP协议深入.md) 怎么分工？**  
本章 = 开发者 Network 排查；计网 04 = 报文结构、方法语义、缓存/CORS 原理与面试。

**Q11：[Vue 08](../Vue/08-Axios网络请求与前后端联调.md) devServer 代理解决什么？**  
开发时浏览器认为请求同源（如 `/api` → 后端），绕过 CORS；生产需真实 CORS 或网关。

**Q12：强缓存 `max-age=0` 呢？**  
通常走协商缓存；`no-cache` 每次都要验证；`no-store` 完全不存。

---

## 29. 练习建议

1. Network 面板分析一个真实网站的首屏请求
2. 对比同一接口 GET 与 POST 的 Request 差异
3. 故意请求错误 URL，观察 404 与 CORS 报错区别
4. 查看响应头里的 `Content-Type` 和 `Cache-Control`

---

## 30. 从输入 URL 到页面渲染（完整流程）

```
1. 输入 URL，浏览器解析协议/域名/路径
2. DNS 解析：域名 → IP 地址
3. TCP 三次握手建立连接
4. HTTPS 则进行 TLS 加密握手
5. 浏览器发送 HTTP 请求（请求行 + 请求头 + 可选 body）
6. 服务器处理并返回 HTTP 响应（状态行 + 响应头 + body）
7. 浏览器解析 HTML → 构建 DOM 树
8. 解析 CSS  → 构建 CSSOM 树
9. 合成渲染树 → 布局（Layout，计算位置大小）→ 绘制（Paint）
10. 显示页面。JS 执行可能阻塞 HTML 解析（除非 defer/async）
```

**前端核心关注**：步骤 5-10。理解这个流程有助于调试首屏加载、渲染阻塞等问题。

---

## 31. 前端安全基础（XSS 与 CSRF）

### XSS（跨站脚本攻击）

攻击者把恶意脚本注入页面。三种类型：

| 类型 | 发生位置 | 防御 |
|------|----------|------|
| 存储型 | 恶意数据存到服务器（如评论），其他用户访问时执行 | 服务端过滤输出，前端用 `textContent` |
| 反射型 | URL 参数中包含脚本，服务端回显 | 不信任 URL 参数 |
| DOM 型 | 前端 JS 把不可信数据插入 HTML | **绝不**用 `innerHTML` 插入用户输入 |

```js
// ❌ XSS 漏洞
const userInput = '<img src=x onerror=alert(1)>';
document.getElementById("box").innerHTML = userInput;

// ✅ 安全做法
document.getElementById("box").textContent = userInput;
// 如果必须插入 HTML，使用 DOMPurify 等库过滤
```

### CSRF（跨站请求伪造）

攻击者诱导用户点击链接，以用户身份向目标网站发请求。

**防御**：关键操作（支付、删改）要求 CSRF Token；SameSite Cookie；验证 Referer。

---

## 32. Web 性能指标速查

| 缩写 | 全称 | 含义 | 良好阈值 |
|------|------|------|----------|
| FP | First Paint | 首次绘制（任意像素） | < 1s |
| FCP | First Contentful Paint | 首次内容绘制 | < 1.8s |
| LCP | Largest Contentful Paint | 最大内容绘制 | < 2.5s |
| TBT | Total Blocking Time | 主线程阻塞总时长 | < 200ms |
| CLS | Cumulative Layout Shift | 累计布局偏移 | < 0.1 |
| TTI | Time to Interactive | 可交互时间 | < 3.8s |

用 Chrome Lighthouse 或 PageSpeed Insights 测量。

---

## 33. 性能优化清单

### 资源加载阶段
- [ ] 压缩 HTML/CSS/JS（构建工具处理）
- [ ] 图片压缩（WebP/AVIF 格式）
- [ ] 合理使用缓存（强缓存 / 协商缓存）
- [ ] CSS 放 head，JS 放底部或 `defer`/`async`
- [ ] 关键 CSS 内联，非关键 CSS 异步加载

### 首屏渲染阶段
- [ ] 懒加载图片（`loading="lazy"` 或 Intersection Observer）
- [ ] 代码分割（React.lazy / dynamic import）
- [ ] 减少首屏请求数量
- [ ] 预加载关键资源（`<link rel="preload">`）

### 运行时阶段
- [ ] 避免强制同步布局（读写 DOM 分离）
- [ ] 动画优先 `transform` 和 `opacity`
- [ ] 长任务拆分（`requestIdleCallback` 或 `setTimeout`）
- [ ] 防抖/节流高频事件

---

## 34. HTTPS 加密过程（简化版）

```
客户端                             服务器
  │──── 1. 客户端 Hello ──────────→│
  │     （支持的加密算法列表）        │
  │←─── 2. 服务器 Hello + 证书 ────│
  │     （选算法 + 公钥）            │
  │──── 3. 客户端验证证书 ──────────│
  │──── 4. 生成对称密钥            │
  │     用服务器公钥加密发送 ──────→│
  │←─── 5. 之后都用对称密钥通信 ───→│
```

前端要知道的：HTTPS 不是"更慢的 HTTP"，现代 TLS 1.3 握手只需 1-RTT。生产环境必须用 HTTPS。

---

## 35. 完整实战：Network 面板实操指南

1. **F12** → **Network** 标签 → 勾选 **Preserve log**（页面跳转后仍保留记录）
2. 刷新页面，观察资源列表：
   - **Doc**：HTML 文档
   - **CSS**：样式表
   - **JS**：脚本
   - **Img**：图片
   - **Fetch/XHR**：接口请求
3. 点一个请求，查看：
   - **Headers**：URL、Method、Status Code、Req/Res Headers
   - **Preview/Response**：返回内容（JSON 会格式化展示）
   - **Timing**：排队 → DNS → 连接 → 等待（TTFB）→ 下载
4. 筛选技巧：
   - 筛选框输入 `api` 只看接口
   - 右键 → "Clear browser cache" 强制重新加载

### 35.1 Vue 08 联调排查六步（Network + 计网 04）

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 项目跑起来，F12 Network，Filter **Fetch/XHR** | 出现 `/api/...` 请求 | baseURL、proxy 是否配置 |
| 2 | 看 **General → Request URL** | 与后端 [Java 04](../../后端学习/Java/04-SpringBoot核心开发.md) 路径一致 | 少 `/api` 前缀或多斜杠 |
| 3 | 看 **Status Code** | 200/201 成功；401 查 token | 对照 [计网 04 状态码](../计算机网络/04-HTTP协议深入.md) |
| 4 | **Request Headers** 是否有 `Authorization: Bearer` | 登录后应有 | 拦截器未挂载或 localStorage 无 token |
| 5 | **Response** Preview 是否为预期 JSON | `{ code, data }` 结构与前端解构一致 | 字段 camelCase vs snake_case |
| 6 | Console 若 CORS 红字 | 开发环境查 vite proxy；生产查后端 CORS 头 | 见 §20 与计网 04 |

### 根据状态码快速判断

| 状态码 | 含义 | 先查什么 |
|--------|------|----------|
| `(failed)` 红色 | 网络不通 / CORS 拦截 / URL 写错 | 检查 URL、CORS 配置 |
| 404 | 资源不存在 | 检查路径和方法 |
| 401 | 未认证 | Token 是否带上、是否过期 |
| 403 | 没权限 | 确认权限配置 |
| 500 | 服务器错误 | 叫后端看日志 |
| 304 | 未修改（缓存命中） | 正常！内容来自缓存 |

---

## 36. 学完标准（扩充）

- [ ] 能拆解 URL，理解 HTTP 请求/响应结构
- [ ] 熟记常见状态码及排查方向（200/301/304/400/401/403/404/500）
- [ ] 理解同源策略、跨域、CORS 基本概念
- [ ] 区分 Cookie / Session / Token / JWT 使用场景
- [ ] 会用 Network 面板分析接口与静态资源
- [ ] 了解「输入 URL 到页面渲染」完整流程
- [ ] 知道 XSS 和 CSRF 的基本原理和防御
- [ ] 了解 Web 性能指标（FCP/LCP/CLS）
- [ ] 对缓存、HTTPS、DNS 有基础印象
- [ ] 能独立完成 Network 面板实操
- [ ] 能与 [计网 04](../计算机网络/04-HTTP协议深入.md) 对照，不矛盾地描述 HTTP 请求/响应
- [ ] 联调 [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 时能独立排查 CORS 与 401

---

## 37. 与计网 04 / Vue 08 对照表

| 话题 | 本章（开发者视角） | [计网 04](../计算机网络/04-HTTP协议深入.md) | [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) |
|------|-------------------|---------------------------------------------|--------------------------------------------------|
| 请求行 Method/URL | Network Headers 里看 | 报文格式、REST 语义 | axios `method`、`url` |
| 状态码 | §7、§35 速查表 | 分类与场景详解 | 拦截器里统一处理 401 |
| Request Body JSON | Preview 面板 | Content-Type、序列化 | `axios.post(url, data)` |
| CORS 报错 | §20 Console 红字 | 简单/预检请求 | vite.config proxy |
| Authorization | Request Headers | Bearer/JWT 概念 | 请求拦截器自动带 token |
| 缓存 304 | 刷新仍旧 | ETag/Cache-Control 原理 | 构建 hash 文件名 |

学完本章 Network 实操 + 计网 04 原理 + Vue 08 工程封装，前后端联调链路就闭环。

---

## 38. 闭卷自测

1. URL `https://api.example.com:8443/v1/users?page=2` 各部分分别是什么？
2. 同源的三要素是什么？举一对不同源的例子。
3. 200、201、204 分别常表示什么？
4. 强缓存和协商缓存，哪个在「未过期」时完全不发 HTTP 请求？
5. CORS 错误是「请求没到服务器」吗？正确理解是什么？
6. GET 和 POST 在「是否有 body」「幂等性」上各怎么说？
7. `Content-Type: application/json` 表示什么？
8. **动手**：Network 里找一个 Fetch 请求，写出 Method、Status、Content-Type。
9. **动手**：故意跨域 fetch 一次，复制 Console 里 CORS 关键词一句。
10. **综合**：Vue 08 开发环境 200 但生产 CORS 失败，最可能差在哪一环？

### 38.1 自测参考答案

1. 协议 https；域名 api.example.com；端口 8443；路径 /v1/users；查询 page=2。
2. 协议、域名、端口全同；如 http://a.com vs https://a.com。
3. 200 OK；201 Created；204 No Content（成功无 body）。
4. 强缓存（如 max-age 未过）。
5. 不是；请求可能已到服务器，浏览器拒绝把响应交给 JS。
6. GET 通常无 body、幂等；POST 常有 body、非幂等（创建）。
7. body 是 JSON 格式。
8. 学生自选记录，三项齐全即可。
9. 含 `CORS policy` / `Access-Control-Allow-Origin` 等。
10. 开发用了 proxy 同源，生产未配后端 CORS 或网关。

---

## 39. 费曼检验

3 分钟向朋友解释「点一下网页背后发生了什么」：

1. **URL 是地址，HTTP 是寄信规则**——浏览器问服务器要文件或 JSON，服务器用**状态码**回信。
2. **保安叫同源策略**——别的域名来的数据，默认不让 JS 看，除非服务器给 CORS 白名单（[计网 04](../计算机网络/04-HTTP协议深入.md) 详讲）。
3. **Network 面板是快递跟踪**——联调 [Vue 08](../Vue/08-Axios网络请求与前后端联调.md) 卡住，先 F12 看这一屏，比猜代码快十倍。

> 深入 TCP、HTTP/2、TLS：[计网 04](../计算机网络/04-HTTP协议深入.md) → [计网 05 HTTPS](../计算机网络/05-HTTPS与TLS加密.md)。动手联调：[09 fetch](./09-JavaScript异步编程网络请求与本地存储.md) → [Vue 08 Axios](../Vue/08-Axios网络请求与前后端联调.md)。
