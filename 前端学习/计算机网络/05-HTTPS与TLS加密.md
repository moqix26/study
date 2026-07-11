# HTTPS 与 TLS 加密

> **文件编码**：UTF-8。  
> **定位**：Go 后端 **0 基础** 读者的计网 **05 章**——在 [04-HTTP协议深入](./04-HTTP协议深入.md) 掌握 HTTP 报文之后，讲清 **为什么生产必须 HTTPS**、**TLS 如何握手**、**证书链如何建立信任**，以及 **443 端口、重定向、HSTS** 等后端必知配置；为 [Go 05 net/http](../../后端学习/Go/05-Go标准库与HTTP基础.md) 与 [Go 09 JWT](../../后端学习/Go/09-JWT认证与用户体系.md) 打地基。  
> **前置**：[04-HTTP协议深入](./04-HTTP协议深入.md)、[02-TCP与UDP](./02-TCP与UDP.md)（建议 [03 IP 与 DNS](./03-IP地址与DNS解析.md)）  
> **工具**：**PowerShell + curl -v**（本章主工具；Chrome DevTools Security 仅作可选补充）

---

## 0. 读前导读（零基础也能跟上）

### 0.1 用一句话弄懂本章

**HTTPS** = 在 HTTP 外面套一层 **TLS 加密隧道**——像把挂号信放进**带密码锁的保险箱**再寄出；同时用 **CA 证书**证明「收件人确实是 `api.example.com` 本人」，不是钓鱼冒充。

**核心类比：HTTPS = 银行柜台 + 身份证 + 密封信封**

| 生活场景 | HTTPS 对应 |
|----------|------------|
| 明文 HTTP 寄明信片 | 路人都能读 password、JWT |
| TLS 对称加密（AES-GCM） | 密封信封，只有双方有钥匙 |
| 非对称加密（ECDHE）握手 | 先通过公开渠道**安全商量**信封密码 |
| CA 证书链 | 公安局盖章的身份证，客户端认「真服务器」 |
| 443 端口 | 银行专用通道号（HTTP 是 80 号普通窗口） |
| 混合内容拦截 | HTTPS 页面里不许偷偷塞 HTTP 明信片（浏览器策略，面试常问） |
| HSTS | 银行贴告示：「本域只走加密通道，别走旁门」 |

**与 04 章关系**：04 讲「信怎么写」（HTTP 报文）；本章讲「信怎么**安全寄**」（TLS + 证书）。Go 客户端发 `Authorization: Bearer <token>` 时，**没有 HTTPS 时该头明文可被 Wi-Fi 嗅探**。

### 0.2 你需要提前知道什么

| 前置 | 对应章节 | 本章是否必须 |
|------|----------|--------------|
| HTTP 请求/响应、状态码 | [04 章](./04-HTTP协议深入.md) | ✅ 必须先读 |
| TCP 三次握手、端口 | [02 章](./02-TCP与UDP.md) | ✅ TLS 在 TCP 之后 |
| DNS 域名解析 | [03 章](./03-IP地址与DNS解析.md) | 建议有 |
| Go 基础语法 | [Go 00～04](../../后端学习/Go/) | ✅ 能读简单代码即可 |
| JWT 是什么 | [Go 09](../../后端学习/Go/09-JWT认证与用户体系.md) | 本章会预告，不必须 |

**不会 HTTP 报文？** 先读完 04 章 §1～§3，再回来。  
**还没写 Go HTTP 服务？** 正常；本章用 **curl** 观察 HTTPS，Go 05 再写代码。

### 0.3 本章知识地图（学完后应能勾选全部 ☐→☑）

```text
☐ 能说出 HTTP 明文的三类风险（窃听、篡改、冒充）及 HTTPS 对策
☐ 能区分对称/非对称加密及在 TLS 握手中的分工
☐ 能口述证书链：根 CA → 中间 CA → 叶子证书，客户端如何验证
☐ 能对比 TLS 1.2（~2 RTT）与 TLS 1.3（~1 RTT）
☐ 知道 443 端口、Nginx TLS 终结、内网 HTTP 反代架构
☐ 能解释混合内容 Active/Passive（浏览器面试语境）
☐ 能说明 HSTS 防 SSL Stripping 的原理
☐ 会用 curl -v 查看 TLS 版本与证书信息
☐ 能说明 JWT Bearer Token 为何必须走 HTTPS 传输
☐ 知道 HTTPS 防传输层窃听，不防 XSS / 代码层漏洞
☐ 闭卷自测 ≥ 8/10
```

### 0.4 建议学习时长与节奏

| 阶段 | 内容 | 时间 |
|------|------|------|
| 动机 + 加密基础 | §0～§3 | 45 分钟 |
| TLS 握手 + 端口部署 | §4～§6 | 50 分钟 |
| 混合内容 + HSTS + Token | §7～§9 | 40 分钟 |
| curl 实操 | §10 | 35 分钟 |
| 自测复盘 | FAQ、闭卷、费曼 | 30 分钟 |

**动手优先**：读完 TLS 1.3 握手后，立刻在 PowerShell 执行 `curl.exe -v https://httpbin.org/get`，找 `TLSv1.3` 与 `SSL connection using` 两行。

### 0.5 学完本章你能做什么（可验证的具体动作）

1. 用 `curl.exe -v` 访问任意 HTTPS 站，说出 TLS 版本与 cipher 套件名称。
2. 解释：短链 API 若生产环境用 `http://` 传 JWT，攻击者能做什么。
3. 白板画 TLS 1.3 简化握手（ClientHello → ServerHello+证书+Finished → Finished）。
4. 写出 Nginx 三行核心：`80→301 https`、`443 ssl`、`proxy_pass` 反代 Go 8080。
5. 向同学 2 分钟口述 HTTPS 完整流程（对照 §17 费曼检验）。

---

## 本章衔接

### 与上一章的关系

| 04 章你已学会 | 05 章继续深入 |
|---------------|---------------|
| HTTP 明文报文、Header、状态码 | 同样报文，但经 **TLS 加密** 后 outsiders 不可读 |
| `Authorization: Bearer token` | token 在 HTTPS 下防**中间人窃听** |
| curl `http://localhost:8080` | 生产 `https://api.example.com:443` |
| HTTP/2 多路复用概念 | 现代浏览器 **HTTPS 才开 HTTP/2**（localhost 除外） |

```mermaid
flowchart TB
    subgraph 明文["HTTP（不安全）"]
        H1["GET /api/login"]
        H1 --> Wire1["网络传输：可被窃听"]
    end
    subgraph 加密["HTTPS = HTTP + TLS"]
        H2["GET /api/login"]
        H2 --> TLS["TLS 加密层"]
        TLS --> Wire2["密文传输"]
        Wire2 --> TLS2["TLS 解密"]
        TLS2 --> Server["Go net/http 服务"]
    end
```

**实战链路（短链项目典型架构）**：

```text
curl / 浏览器 → Nginx :443（TLS 终结）→ 内网 HTTP → Go Gin :8080
```

JWT 在 `Authorization` 头里传输时，**浏览器/curl 到 Nginx 这一段必须是 HTTPS**。

### 与 Go 系列的关系

| 章节 | 本章铺垫什么 |
|------|-------------|
| [Go 05 net/http](../../后端学习/Go/05-Go标准库与HTTP基础.md) | 本地 `http://localhost:8080` 开发；生产前面加 Nginx TLS |
| [Go 06 Gin](../../后端学习/Go/06-Gin框架核心与中间件.md) | 路由、中间件；HTTPS 在 Nginx 层终结 |
| [Go 09 JWT](../../后端学习/Go/09-JWT认证与用户体系.md) | Bearer Token 必须 HTTPS 传输 |

---

## 1. 为什么需要 HTTPS？

### 1.1 HTTP 的三类风险

| 风险 | 攻击者能做什么 | 没有 HTTPS 的例子 |
|------|----------------|-------------------|
| **窃听（Eavesdropping）** | 看到密码、JWT、个人信息 | 公共 Wi-Fi 抓包读到 `password=123456` |
| **篡改（Tampering）** | 改响应 JSON、注入内容 | 把 `{"clicks":99}` 改成 `{"clicks":1}` |
| **冒充（Spoofing）** | 伪装成 `api.example.com` | 钓鱼站点骗你输入账号密码 |

### 1.2 HTTPS 提供的保障

| 保障 | 机制 | 用户可见表现 |
|------|------|--------------|
| **机密性** | 对称加密传输 | 浏览器地址栏小锁（客户端视角） |
| **完整性** | MAC / AEAD 校验 | 被改包解密失败 |
| **身份认证** | 证书 + CA 链 | 证书有效则域名可信 |

### 1.3 为什么 Go 后端必须关心 HTTPS？（深入解释 ①）

**不只是运维配证书的事**，写 API 的你也需要知道：

1. **传输安全**：JWT、Session ID、密码在 HTTP 下明文飞过链路每个节点
2. **部署架构**：Go 服务通常监听 8080 HTTP，**对外 443 由 Nginx/Caddy 终结 TLS**
3. **客户端行为**：浏览器对 HTTPS 页请求 HTTP API 会 **block**（混合内容）
4. **Cookie 属性**：`Secure` 标记只在 HTTPS 发送（Go 09 若用 Cookie 方案）

**短链登录流程示例**：

```text
curl POST /api/login {username, password}
  → 响应 { "token": "eyJ..." }
  → 后续请求 Header: Authorization: Bearer eyJ...
```

若走 HTTP，**同一局域网**的攻击者用抓包工具即可拿到 token，随后伪造请求调用「创建短链」等需登录接口。

```mermaid
sequenceDiagram
    participant C as curl / 客户端
    participant A as 攻击者
    participant S as Go 服务器
    Note over C,S: HTTP：明文
    C->>S: POST password=123456
    A->>A: 截获明文密码
    Note over C,S: HTTPS：TLS 隧道
    C->>S: 加密密文
    A->>A: 只见乱码，无法还原
```

### 1.4 HTTPS 解决不了什么

| HTTPS 能防 | HTTPS 不能防 |
|-----------|-------------|
| 链路上窃听 JWT | XSS 脚本读 localStorage 里的 token |
| 中间人冒充服务器（证书校验通过时） | 服务器代码 SQL 注入 |
| 传输途中篡改 HTTP 体 | 弱密码、逻辑漏洞 |

**结论**：HTTPS 是**必要条件**，不是**充分条件**。应用层安全见 [计网 06](./06-缓存Cookie与会话机制.md) 与后续 Web 安全系列。

---

## 2. 加密基础：对称 vs 非对称

### 2.1 对称加密（Symmetric Encryption）

**对称加密（Symmetric Encryption）**：**同一把密钥**加密和解密。  
**生活类比**：你和室友共用一把抽屉钥匙——开锁快，但钥匙怎么交给对方是个问题。  
**为什么重要**：HTTP 数据量大，传输阶段必须快。  
**本章用到的地方**：§2.3、§4 TLS 握手后的数据加密。

| 优点 | 缺点 |
|------|------|
| 速度快，适合大量数据 | 密钥如何安全传给对方？ |

常见算法：**AES-128-GCM**、**AES-256-GCM**、ChaCha20-Poly1305（TLS 1.3 常用）。

```text
明文 + 密钥 K → 密文
密文 + 密钥 K → 明文
```

### 2.2 非对称加密（Asymmetric Encryption）

**非对称加密（Asymmetric Encryption）**：**公钥**加密 / 验签，**私钥**解密 / 签名；密钥成对出现。  
**生活类比**：邮箱上的投信口（公钥）谁都能投，只有主人钥匙（私钥）能取。  
**为什么重要**：解决「在不安全信道上商量对称密钥」的问题。  
**本章用到的地方**：§4 TLS 握手。

| 角色 | 持有 | 用途 |
|------|------|------|
| 服务器 | 私钥 | 解密、证明身份 |
| 客户端 | 公钥（来自证书） | 加密、验证签名 |

常见算法：**RSA**（老方案）、**ECDHE**（临时椭圆曲线密钥交换，前向安全）。

```text
密文 = Encrypt(明文, 公钥)
明文 = Decrypt(密文, 私钥)   // 只有私钥持有者能解
```

### 2.3 为什么 TLS 要混用两种？（深入解释 ②）

| 阶段 | 用什么 | 原因 |
|------|--------|------|
| **握手** | 非对称（ECDHE 等） | 在**不安全信道**上协商出只有双方知道的秘密 |
| **传输 HTTP** | 对称（AES-GCM） | 数据量大，对称加密 **CPU 开销低** |

**面试标准答法**：非对称解决「密钥怎么安全商量」；对称负责 bulk 数据加密；TLS 握手结束后用协商好的**会话密钥（session key）**加密 HTTP 报文。

```mermaid
flowchart LR
    A["非对称握手"] --> B["得到会话密钥 session key"]
    B --> C["对称加密 HTTP 数据"]
```

### 2.4 数字签名 vs 加密

- **加密**：保密内容
- **签名**：用私钥签名，公钥验证 → 证明「确实是持有私钥的服务器发的」且内容未被改

证书里的签名由 **CA 私钥**签发，客户端用 **CA 公钥**验证。

### 2.5 术语三件套：AEAD

**AEAD（Authenticated Encryption with Associated Data，带关联数据的认证加密）**：同时加密与校验完整性的算法族（如 AES-GCM）。  
**生活类比**：带封蜡的信——拆开即知是否被拆过。  
**为什么重要**：TLS 1.3 仅允许 AEAD，废弃弱 cipher。  
**本章用到的地方**：§4.2 TLS 1.3 改进。

---

## 3. CA 与证书链

### 3.1 为什么需要 CA（Certificate Authority）？

若允许任意自签证书，任何人都能造 `api.bank.com` 的证书——客户端无法区分真假。

**CA（Certificate Authority，证书颁发机构）** 是受操作系统 / 浏览器信任的第三方，验证域名归属后签发证书。

### 3.2 证书链结构

```text
根证书（Root CA）           ← 预装在 Windows / macOS / 浏览器
    └── 中间证书（Intermediate CA）
            └── 叶子证书（Leaf）  ← 你的 api.example.com
```

客户端（curl、浏览器）验证路径：

1. 用**中间 CA 公钥**验证叶子证书签名
2. 用**根 CA 公钥**验证中间证书
3. 根在**信任库**里 → 整条链可信

```mermaid
flowchart BT
    Root["Root CA 自签"]
    Inter["Intermediate CA"]
    Leaf["api.example.com 叶子证书"]
    Root --> Inter
    Inter --> Leaf
    Client["客户端信任库"] --> Root
    Client -->|"验证链"| Leaf
```

### 3.3 证书里有什么？

| 字段 | 示例 | 作用 |
|------|------|------|
| **Subject** | `CN=api.example.com` | 证书绑定的域名 |
| **Issuer** | `Let's Encrypt R3` | 签发 CA |
| **Validity** | 2026-01-01 ~ 2026-04-01 | 有效期 |
| **Subject Public Key** | EC P-256 | 服务器公钥 |
| **SAN** | `api.example.com`, `www.example.com` | 多域名（Subject Alternative Name） |
| **Signature** | ... | CA 对以上内容的签名 |

### 3.4 自签名 vs 正式证书

| 类型 | 场景 | curl / 浏览器表现 |
|------|------|-------------------|
| **自签名** | localhost 开发、内网 | ⚠️ 不受信任，curl 需 `-k` |
| **Let's Encrypt 等 DV** | 生产免费证书 | ✅ 验证通过 |
| **OV/EV** | 企业 | 更严格验证 |

本地开发可用 **mkcert** 生成本地信任的证书，避免每次跳过验证。

### 3.5 术语三件套：SAN

**SAN（Subject Alternative Name，主题备用名称）**：一张证书可绑多个域名。  
**生活类比**：身份证上除了姓名还列了曾用名——都能证明是同一人。  
**为什么重要**：API 域名 `api.example.com` 必须在 SAN 里，否则证书不匹配。  
**本章用到的地方**：§3.3、§13 报错表。

---

## 4. TLS 握手：1.2 vs 1.3 简化版

**TLS（Transport Layer Security，传输层安全）**：HTTP 与 TCP 之间的加密协议层。  
**生活类比**：寄信前先套密码锁信封。  
**为什么重要**：防窃听、篡改、冒充服务器。  
**本章用到的地方**：§4 全节、§10 curl 观测。

TLS 在 **TCP 连接建立后**、**HTTP 发送前**完成（见 [02 章](./02-TCP与UDP.md) 三次握手）。

### 4.1 TLS 1.2 握手（2-RTT 简化）

```mermaid
sequenceDiagram
    participant C as 客户端
    participant S as 服务器
    C->>S: ClientHello（支持的 cipher、随机数）
    S->>C: ServerHello + 证书 + ServerKeyExchange
    S->>C: ServerHelloDone
    C->>S: ClientKeyExchange + ChangeCipherSpec + Finished
    S->>C: ChangeCipherSpec + Finished
    Note over C,S: 之后：对称密钥加密 HTTP
    C->>S: GET /api/users（加密）
```

**步骤口述**：

1. **ClientHello**：客户端能力 + 随机数
2. **ServerHello + 证书**：选定算法，发证书链
3. 客户端**验证证书**，协商共享密钥（RSA 方案加密 pre-master secret；**ECDHE 等现代方案**用临时密钥交换，不总是 pre-master 这个词）
4. 双方算出**会话密钥**
5. **Finished** 消息后进入加密通信

**RTT（Round-Trip Time，往返时延）**：通常 **2 个往返** 才开始发 HTTP（不含 TCP 三次握手）。

### 4.2 TLS 1.3 握手（1-RTT，更快）

```mermaid
sequenceDiagram
    participant C as 客户端
    participant S as 服务器
    C->>S: ClientHello + KeyShare
    S->>C: ServerHello + 证书 + Finished
    C->>S: Finished
    Note over C,S: 1-RTT 后即可加密 HTTP
    C->>S: GET /api/users（加密）
```

**改进对比**：

| 对比项 | TLS 1.2 | TLS 1.3 |
|--------|---------|---------|
| 握手 RTT | ~2 | **~1** |
| 0-RTT 恢复 | 有限 | 支持（有重放风险，慎用） |
| 废弃算法 | RSA 密钥交换、弱 cipher | 仅 **AEAD**、Forward Secrecy |
| 加密范围 | 握手后半段才加密 | 更多握手消息加密 |

### 4.3 前向安全（Forward Secrecy）

使用 **ECDHE** 临时密钥：即使日后服务器私钥泄露，**过去**抓到的密文也无法解密。

**面试点**：TLS 1.3 强制前向安全；老 TLS 1.2 若配置不当用 RSA 密钥交换则不具备。

### 4.4 HTTPS 与 TCP 三次握手的关系

完整冷启动（简化）：

```text
1. TCP SYN / SYN-ACK / ACK          ← 1 RTT（02 章）
2. TLS 1.3 握手                      ← 1 RTT（本章）
3. HTTP 请求/响应                    ← 1 RTT（04 章）
```

**HTTP/3 + QUIC** 把传输层握手与 TLS 进一步合并（见 04 章 HTTP/3），首包延迟更低——Go 1.21+ 标准库对 HTTP/3 有实验性支持，初学先掌握 TCP + TLS 1.3 即可。

### 4.5 密钥交换算法演进（了解）

| 机制 | 前向安全 | 说明 |
|------|----------|------|
| **RSA 密钥传输** | ❌ | 老方案，用服务器 RSA 公钥加密 pre-master secret |
| **DHE / ECDHE** | ✅ | 临时 Diffie-Hellman，每次握手新密钥 |
| **TLS 1.3** | ✅ 强制 | 移除 RSA key transport |

---

## 5. HTTPS 端口 443 与 URL

### 5.1 默认端口

| 协议 | 默认端口 | URL 示例 |
|------|----------|----------|
| HTTP | **80** | `http://example.com`（等价 :80） |
| HTTPS | **443** | `https://example.com`（等价 :443） |

省略端口时客户端用默认值。Go `net/http` 内嵌服务默认 **8080 HTTP**；生产由 **Nginx/Caddy** 在 443 终结 TLS，反代到 8080。

```mermaid
flowchart LR
    Client["curl / 浏览器 :443 HTTPS"]
    Nginx["Nginx 443 TLS 终结"]
    Go["Go Gin :8080 HTTP"]
    Client --> Nginx
    Nginx -->|"内网 HTTP"| Go
```

### 5.2 开发 vs 生产

| 环境 | 常见做法 |
|------|----------|
| 本地 Go 开发 | `http://localhost:8080`，curl 直接测 |
| 本地 HTTPS 实验 | Go 8443 自签名，或 mkcert |
| 生产 | `https://api.example.com` → Nginx 443 → 8080 |

**注意**：内网反代段可以是 HTTP，但**客户端到 Nginx** 必须是 HTTPS，用户才安全。8080 **不能**直接暴露公网。

### 5.3 Go 服务监听与 TLS 的关系

Go 05 典型写法（HTTP，开发用）：

```go
// go
package main

import "net/http"

func main() {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("ok"))
    })
    // 开发：明文 HTTP
    http.ListenAndServe(":8080", nil)
}
```

生产常见两种模式：

| 模式 | 谁终结 TLS | Go 监听 |
|------|-----------|---------|
| **推荐** | Nginx/Caddy 443 | `:8080` HTTP（内网） |
| 直连 | Go 自己 | `:8443` HTTPS（需证书文件） |

初学 **先用 Nginx 终结**，Go 代码保持简单 HTTP。

---

## 6. HTTP 到 HTTPS 重定向

### 6.1 为什么要重定向？

- 用户习惯输入 `http://example.com`
- SEO 与 Cookie 域统一
- 避免明文泄露

### 6.2 Nginx 常见实现

```nginx
server {
    listen 80;
    server_name api.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.example.com;
    ssl_certificate     /etc/letsencrypt/live/api.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.example.com/privkey.pem;
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Nginx 配置逐行读**：

| 行/指令 | 含义 | 改错会怎样 |
|---------|------|------------|
| `listen 80` | 监听 HTTP 端口 | 用户输入 http 进不来 |
| `return 301 https://...` | 永久跳转到 HTTPS | 若写错域名，全站跳错 |
| `listen 443 ssl http2` | HTTPS + 可选 HTTP/2 | 443 未开则 HTTPS 不可用 |
| `ssl_certificate` | 公钥链（含中间证书） | 链不完整 → 浏览器报错 |
| `ssl_certificate_key` | 私钥 | 必须与证书配对 |
| `proxy_pass http://127.0.0.1:8080` | 反代 Go 服务 | 端口错 → 502 |
| `X-Forwarded-Proto $scheme` | 告诉 Go 原始协议是 https | Go 生成绝对 URL 可能错 |

### 6.3 301 vs 302 重定向到 HTTPS

| 状态码 | 使用 |
|--------|------|
| **301** | 永久迁移，搜索引擎更新索引 |
| **302** | 临时（一般不用于全站 HTTPS） |

客户端 `baseURL` 应直接写 **`https://`**，避免先 HTTP 再 301 多一次往返。

### 6.4 Go 直连 HTTPS（可选，了解即可）

```go
// go
package main

import (
    "log"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("hello https"))
    })
    // 需事先准备 cert.pem + key.pem
    log.Fatal(http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", nil))
}
```

`ListenAndServeTLS` 在 Go 进程内终结 TLS——小项目可以，生产更常见 Nginx 443。

---

## 7. 混合内容（Mixed Content，浏览器语境）

### 7.1 定义

**混合内容（Mixed Content）**：**HTTPS 页面**加载了 **HTTP 资源**（脚本、样式、图片、API 请求等）。

> 本章主读者是 Go 后端：你主要写 API，不写 HTML 页面。但**面试常问**，且将来若有前端联调需知道浏览器规则。

### 7.2 浏览器策略

| 类型 | 示例 | 现代浏览器 |
|------|------|------------|
| **Active Mixed** | `<script src="http://...">`、XHR/fetch `http://api` | **blocked** |
| **Passive Mixed** | `<img src="http://...">` | 可能允许但 ⚠️ 警告 |

Console 典型报错：

```text
Mixed Content: The page at 'https://example.com/' was loaded over HTTPS,
but requested an insecure XMLHttpRequest endpoint 'http://api.example.com/users'.
This request has been blocked; the content must be served over HTTPS.
```

### 7.3 后端开发者如何避免

| 配置 | 正确 | 错误 |
|------|------|------|
| API 公网地址 | `https://api.example.com` | `http://api.example.com` |
| Nginx 对外 | 443 TLS | 仅 80 明文 |
| 返回的绝对 URL | `https://...` | 硬编码 `http://...` |

**纯 API + curl 测试**不存在混合内容——那是浏览器安全策略。但 **Postman / 浏览器调 Swagger** 若页面是 HTTPS、API 是 HTTP，仍会出问题。

---

## 8. HSTS（HTTP Strict Transport Security）

### 8.1 是什么？

**HSTS（HTTP Strict Transport Security，HTTP 严格传输安全）**：响应头告诉浏览器：**在 max-age 内，强制用 HTTPS 访问该域**，即使用户输入 `http://`。

```http
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
```

### 8.2 防什么攻击？

**SSL Stripping**：中间人把用户的 HTTPS 链接降级成 HTTP。HSTS 后浏览器**直接发 HTTPS**，不给首次明文机会（首次访问除外，故还有 **HSTS Preload List**）。

```mermaid
flowchart TD
    A["用户输入 http://bank.com"] --> B{"HSTS 缓存?"}
    B -->|有| C["自动改 https://bank.com"]
    B -->|无| D["可能明文 HTTP"]
    D --> E["中间人攻击窗口"]
    C --> F["TLS 保护"]
```

### 8.3 后端注意

- 本地 `localhost` 一般不加 HSTS
- `includeSubDomains` 会影响所有子域
- 提交 **preload** 后很难撤销，需慎重
- Nginx 用 `add_header Strict-Transport-Security "max-age=31536000" always;`

---

## 9. Cookie / JWT 与传输安全（衔接 Go 09）

### 9.1 Set-Cookie 安全 flags

| 属性 | 作用 |
|------|------|
| **Secure** | 仅 HTTPS 发送 |
| **HttpOnly** | JS 无法 `document.cookie` 读取，防 XSS 偷 Cookie |
| **SameSite=Strict/Lax** | 降低 CSRF 风险 |

Go 09 若用 Cookie 存 Session（示意）：

```go
// go
http.SetCookie(w, &http.Cookie{
    Name:     "session_id",
    Value:    sessionID,
    HttpOnly: true,
    Secure:   true, // 生产必须 true
    SameSite: http.SameSiteLaxMode,
    Path:     "/",
})
```

### 9.2 JWT Bearer Token（Go 09 常见）

| 存法 / 传法 | XSS 风险 | CSRF | HTTPS 必要性 |
|-------------|----------|------|--------------|
| Header `Authorization: Bearer` | 若客户端存 localStorage，XSS 可读 | 较低 | **必须 HTTPS 防传输窃听** |
| HttpOnly Cookie 存 JWT | JS 读不到 | 需 SameSite / CSRF Token | 必须 Secure |

**Go 09 典型登录响应**：

```json
{
  "code": 0,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

客户端后续请求：

```http
GET /api/shortlinks HTTP/1.1
Host: api.example.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

攻击者在链路上只能看到 **TLS 密文**——前提是 URL 是 `https://`。

### 9.3 curl 带 JWT 访问 HTTPS API

```powershell
# PowerShell
$token = "eyJhbGciOiJIUzI1NiIs..."
curl.exe -v `
  -H "Authorization: Bearer $token" `
  https://api.example.com/api/shortlinks
```

**`-v`** 会显示 TLS 握手过程；`Authorization` 头在加密隧道内，抓包只见密文。

### 9.4 HTTPS 与 JWT 验签的分工

| 层次 | 负责什么 |
|------|----------|
| **TLS** | 传输加密 + 服务器身份（证书） |
| **JWT 签名（HS256/RS256）** | 证明 token 未被篡改、是谁签发的 |

两者互补：TLS 防路上被偷看；JWT 签名防 token 内容被伪造（在应用层验证）。

---

## 10. 手把手：PowerShell + curl 验证 HTTPS

### 10.1 为什么用 curl.exe

Windows PowerShell 里 `curl` 可能是 **Invoke-WebRequest 别名**。请显式用 **`curl.exe`**，输出与 Linux/macOS 一致，便于对照教程。

### 10.2 访问公网 HTTPS

```powershell
curl.exe -v https://httpbin.org/get
```

**预期（节选，不同环境 cipher 名可能略有差异）**：

```text
* TLSv1.3 (OUT), TLS handshake, Client hello
* TLSv1.3 (IN), TLS handshake, Server hello
* SSL connection using TLSv1.3 / TLS_AES_256_GCM_SHA384
> GET /get HTTP/2
> Host: httpbin.org
<
< HTTP/2 200
```

**如何读输出**：

| 前缀 | 含义 |
|------|------|
| `*` | curl 调试信息（TLS 握手在这里） |
| `>` | 客户端发出的 HTTP 头 |
| `<` | 服务器返回的 HTTP 头 |

### 10.3 只看响应头（-I）

```powershell
curl.exe -vI https://www.baidu.com 2>&1 | Select-String -Pattern "subject|issuer|SSL|TLS"
```

PowerShell 用 `Select-String` 过滤；Linux/macOS 可用 `grep -i`。

### 10.4 查看证书信息（openssl，可选）

若已安装 Git Bash 或 OpenSSL：

```bash
openssl s_client -connect www.baidu.com:443 -servername www.baidu.com </dev/null 2>/dev/null | openssl x509 -noout -subject -issuer -dates
```

**预期输出示例**（域名与日期随时间变化，格式类似）：

```text
subject=CN = baidu.com
issuer=C = US, O = DigiCert Inc, ...
notBefore=Mar  1 00:00:00 2026 GMT
notAfter=Mar  1 23:59:59 2027 GMT
```

### 10.5 跳过证书验证（仅调试自签名）

```powershell
curl.exe -k https://localhost:8443/health
```

**`-k` / `--insecure`**：不验证证书链——**切勿在生产脚本中习惯使用**。

### 10.6 指定 TLS 版本（排查用）

```powershell
curl.exe --tlsv1.2 -v https://example.com
curl.exe --tls-max 1.3 -v https://example.com
```

### 10.7 测试本地 Go HTTPS 服务

```powershell
# 终端 1：运行 Go ListenAndServeTLS（需 cert.pem / key.pem）
go run main.go

# 终端 2
curl.exe -k -v https://localhost:8443/health
```

自签名证书会报 `SSL certificate problem`——加 `-k` 仅本地调试。

### 10.8 TLS 握手观测步骤表

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | `curl.exe -v https://httpbin.org/get 2>&1` | `TLSv1.3 (OUT), Client hello` | 查网络/防火墙/代理 |
| 2 | 继续往下看 | `Server hello` + 证书相关信息 | 见 §13 报错表 |
| 3 | 找 `SSL connection using` | 如 `TLS_AES_256_GCM_SHA384` | 服务器只开老 TLS |
| 4 | 找 `> GET` 行 | HTTP 请求已发出 | TLS 握手失败则不会有 |
| 5 | 找 `< HTTP/2 200` 或 `< HTTP/1.1 200` | 响应成功 | 4xx/5xx 是 HTTP 层问题 |

### 10.9 Chrome DevTools Security（建议配合 curl 使用）

若你用浏览器访问 HTTPS 页面，建议与 curl 对照：

1. `F12` → **Security** 标签
2. 查看 **Overview** 是否 secure、**View certificate** 看 Issuer 与有效期

**多栈证书信任**（面试常问「为什么 Java/Python 报证书错」）：

| 环境 | 信任根证书来源 |
|------|----------------|
| 浏览器 / curl（Win） | 系统证书存储 |
| **Java** | JVM `cacerts` 信任库；企业内网常要导入 `keytool -import` |
| **Python requests** | 依赖 `certifi` 包或系统 CA；`verify=False` 仅调试 |
| **Go** | 默认用系统根；`InsecureSkipVerify` 仅本地调试 |

后端日常调试 **优先 curl**；DevTools 适合对照「用户浏览器看到了什么证书」。

---

## 11. Go 短链 API 全链路安全示意

```mermaid
sequenceDiagram
    participant C as curl HTTPS
    participant N as Nginx 443
    participant G as Go Gin 8080
    C->>N: TLS 加密 POST /api/login
    N->>G: 明文 HTTP（内网）
    G->>N: JSON token
    N->>C: TLS 加密响应
    C->>N: Authorization Bearer（TLS 内）
    N->>G: 转发
    G->>C: 200 短链列表
```

| 环节 | 安全要点 |
|------|----------|
| 客户端 ↔ Nginx | TLS 1.2+、有效证书、可选 HSTS |
| Nginx ↔ Go | 内网隔离；8080 不暴露公网 |
| 登录接口 | HTTPS + 密码 bcrypt 哈希（Go 09） |
| token 传输 | 仅 HTTPS；短过期 + Refresh |
| CORS | 不用 `*` 配 credentials；明确 Origin（计网 06） |

详见 [Go 09 JWT](../../后端学习/Go/09-JWT认证与用户体系.md)、[Go 06 Gin 中间件](../../后端学习/Go/06-Gin框架核心与中间件.md)。

---

## 12. HTTP 与 HTTPS 并排对照（复习 04 + 05）

| 维度 | HTTP | HTTPS |
|------|------|-------|
| 默认端口 | 80 | 443 |
| 传输 | 明文 | TLS 密文 |
| 证书 | 不需要 | 需要有效证书链 |
| 地址栏（浏览器） | Not secure | 小锁 |
| HTTP/2（浏览器策略） | localhost 除外一般不行 | 支持 |
| 混合内容 | 无此概念 | HTTPS 页不宜引 HTTP 资源 |
| JWT / 密码 | 链路可嗅探 | 链路加密 |
| curl | `http://` | `https://`，自签名需 `-k` |
| Go 本地开发 | `:8080` 常见 | Nginx 443 或 `ListenAndServeTLS` |
| 性能 | 少 TLS 握手 | 多 1～2 RTT（TLS 1.3 已优化） |

---

## 13. 常见报错与排查表

| 现象 | 原因 | 排查 | 解决 |
|------|------|------|------|
| `SSL certificate problem: unable to get local issuer certificate` | 自签名 / 链不完整 | `curl -v` 看证书 | 用正规 CA；或本地 `-k` 仅调试 |
| `SSL certificate problem: certificate has expired` | 证书过期 | 看 notAfter | 续签 Let's Encrypt |
| `SSL: no alternative certificate subject name matches` | 域名与 CN/SAN 不匹配 | 证书 Subject/SAN | 重签含正确域名 |
| `curl: (35) schannel: next InitializeSecurityContext failed` | Windows 证书/TLS 协商失败 | 换 `--tlsv1.2` 试 | 更新系统根证书；查服务器 TLS 配置 |
| Mixed Content blocked（浏览器） | HTTPS 页请求 HTTP API | 浏览器 Console | API 改 https 或同域反代 |
| `Connection refused` :443 | 443 未监听 / 安全组未开 | `curl -v` 连不上 | Nginx 监听 443；云防火墙放行 |
| 只有 HTTP 能访问 | 未配 443 或未跳转 | `curl -I http://` | 配 80→301→443 |
| 301 循环 | HTTP/HTTPS 互相重定向 | `curl -I` 跟踪 Location | 修正 Nginx 配置 |
| HSTS 导致本地 http 失败 | 曾访问过带 HSTS 的域 | 换测试域名 | 开发用未 preload 的域 |
| `502 Bad Gateway` | Nginx 连不上 Go 8080 | Go 进程是否运行 | 启动 Go；查 proxy_pass 端口 |
| TLS 版本过低 | 服务器只开 TLS 1.0/1.1 | `curl --tlsv1.2 -v` | Nginx `ssl_protocols TLSv1.2 TLSv1.3;` |
| token 疑似泄露 | 曾用 HTTP 上线 | 访问日志、换密钥 | 强制 HTTPS + 轮换 JWT secret |

---

## 14. 常见问题 FAQ

**Q1：HTTPS 是不是只要买了证书就安全了？**  
不是。证书解决**身份**与**传输加密**；应用层 XSS 仍能偷客户端存的 token，还需输入校验、CSP 等。内网 Nginx→Go 段若是 HTTP，要依赖**内网隔离**，8080 不能暴露公网。

**Q2：自签名证书开发能用吗？**  
可以。curl 用 `-k`；浏览器会警告。推荐 **mkcert** 安装本地信任根，体验更接近生产。

**Q3：TLS 1.3 一定比 1.2 快吗？**  
握手通常 **1-RTT vs 2-RTT**，首包更快；小 API 响应时延迟仍可能 dominated 后端逻辑。用 `curl -v` 区分 TLS 阶段与 HTTP 等待。

**Q4：为什么 HTTP/2 浏览器常要求 HTTPS？**  
历史安全策略 + 避免中间人 downgrade。localhost 开发是例外。

**Q5：HSTS 首次访问还有风险吗？**  
有。用户**第一次**访问该域若被 SSL Stripping，HSTS 尚未写入。`preload` 列表可缓解，但提交 preload 后难撤销。

**Q6：JWT 放 Header 还要 Cookie Secure 吗？**  
Bearer 方案不依赖 Cookie，但生产仍应全站 HTTPS；`Secure` 针对 Set-Cookie 场景。

**Q7：Nginx 443 终结 TLS，Go 8080 HTTP 安全吗？**  
**客户端到 Nginx** 必须 HTTPS；Nginx 到 Go 在内网/VPC 可 HTTP，前提是 8080 外网不可达。

**Q8：curl -k 和浏览器点「继续访问」一样吗？**  
类似：都跳过证书链验证。仅调试自签名，勿用于生产脚本。

**Q9：Go 里要不要自己实现 TLS？**  
不要。用 **Nginx/Caddy** 或 `ListenAndServeTLS` + 正规库；密码学勿手写。

**Q10：HTTPS 能防 SQL 注入吗？**  
不能。SQL 注入是应用层问题，靠参数化查询、ORM 规范（Go 07 GORM）。

**Q11：Let's Encrypt 免费证书生产能用吗？**  
能。DV 证书验证域名控制权，90 天有效，需自动续期（certbot）。

**Q12：怎么知道对方用的是 TLS 1.2 还是 1.3？**  
`curl.exe -v` 输出里找 `SSL connection using TLSv1.x`。

---

## 15. 术语速查（术语三件套汇总）

**TLS（Transport Layer Security，传输层安全）**  
- **定义**：HTTP 与 TCP 之间的加密协议层。  
- **生活类比**：寄信前先套密码锁信封。  
- **为什么重要**：防窃听、篡改、冒充服务器。  
- **本章用到的地方**：§4 握手、§10 curl 观测。

**CA（Certificate Authority，证书颁发机构）**  
- **定义**：受信任的第三方，验证域名后签发数字证书。  
- **生活类比**：公安局给身份证盖章。  
- **为什么重要**：没有 CA，客户端无法区分真假服务器。  
- **本章用到的地方**：§3 证书链。

**Session Key（会话密钥）**  
- **定义**：TLS 握手结束后，双方协商出的对称加密密钥。  
- **生活类比**：每次见面临时换的新抽屉锁密码。  
- **为什么重要**：用它加密 HTTP 正文，速度快。  
- **本章用到的地方**：§2.3、§4。

**Forward Secrecy（前向安全）**  
- **定义**：即使长期私钥日后泄露，过去会话密文仍无法解密。  
- **生活类比**：每次对话用一次性密码，旧密码作废。  
- **为什么重要**：ECDHE / TLS 1.3 的核心安全属性。  
- **本章用到的地方**：§4.3、§4.5。

**ALPN（Application-Layer Protocol Negotiation，应用层协议协商）**  
- **定义**：TLS 握手中协商上层协议（如 h2 或 http/1.1）。  
- **生活类比**：打电话先问「说中文还是英文」再聊正文。  
- **为什么重要**：决定 HTTPS 之上跑 HTTP/2 还是 HTTP/1.1。  
- **本章用到的地方**：§10 curl 输出里的 `HTTP/2`。

---

## 16. 闭卷自测（10 题）

### 概念题（6 题）

1. 用「**保险箱 + 身份证**」类比：TLS、CA 证书、对称加密各对应什么？
2. HTTP 明文传输的三类风险是什么？HTTPS 各如何解决？
3. 为什么 TLS 握手用非对称、数据传输用对称？
4. 什么是混合内容？Active 与 Passive 区别？
5. HSTS 的 `max-age` 和 `preload` 各做什么？
6. JWT Bearer Token 为什么必须 HTTPS？HTTPS 能防 XSS 偷 token 吗？

### 动手题（2 题）

7. 写出 Nginx `80→301 https` 与 `443 ssl` 反代 `127.0.0.1:8080` 的最小 server 块骨架（各 3 行核心）。
8. 如何用 PowerShell + curl 查看一次 HTTPS 请求的 TLS 版本？

### 综合题（2 题）

9. 短链项目生产环境 API 误用 `http://api.example.com`，客户端是 HTTPS 网页，会发生什么？Go 后端要改什么？
10. 2 分钟口述：`curl -v https://api.example.com/api/login` 经历的 TCP + TLS + HTTP 分层过程。

### 自测参考答案

**1.** TLS = 密封保险箱；CA 证书 = 身份证验真；对称加密 = 双方共用的箱锁密码（session key）。

**2.** 窃听 → 加密；篡改 → MAC/AEAD；冒充 → 证书链验证。

**3.** 非对称解决不安全信道上的密钥协商；对称加密 bulk 数据 CPU 开销低。

**4.** HTTPS 页加载 HTTP 资源。Active：脚本/XHR → block；Passive：图片等 → 可能警告。

**5.** max-age 内浏览器强制 HTTPS；preload 提交浏览器内置列表，首次也强制。

**6.** Bearer 在 Header 里，HTTP 下明文可被嗅探。HTTPS 不防 XSS；XSS 靠转义、CSP、HttpOnly Cookie 等。

**7.**  
`listen 80; return 301 https://$host$request_uri;`  
`listen 443 ssl; ssl_certificate ...; ssl_certificate_key ...;`  
`location / { proxy_pass http://127.0.0.1:8080; }`

**8.** `curl.exe -v https://httpbin.org/get 2>&1`，找 `SSL connection using TLSv1.x / ...`。

**9.** 浏览器 block 混合内容，API 调不通；Nginx 配 443 证书，对外只暴露 https，80 跳 301。

**10.** DNS 解析 → TCP 443 三次握手 → TLS 1.3 握手验证书 → 协商 session key → AES 加密 HTTP POST → Nginx 解密 → 反代 Go 8080 → 原路返回加密响应。

---

## 17. 费曼检验：3 分钟讲给零基础朋友

**请不看资料，向没学过编程的朋友解释 HTTPS。对照是否讲到：**

1. **HTTP 像明信片谁都能看**；HTTPS 把信放进**带锁保险箱**，钥匙是握手时商量出来的（session key）。
2. **CA 证书像身份证**，电脑只认「公安局（根 CA）」盖章的；假网站证书不对会被拦。
3. **小锁不等于万能**：防路上被偷听，**防不了**你电脑里恶意程序偷 token；写 Go API 还要防 SQL 注入、校验输入。
4. **443 是 HTTPS 专用门牌号**；Go 程序常在 8080，门口 Nginx 负责上锁（TLS）。
5. **curl -v** 能亲眼看到「先 TLS 握手，再发 HTTP」的顺序。

---

## 18. 练习建议

### 基础题

1. 说出 HTTP 明文传输的三类风险，HTTPS 分别如何应对。
2. 对称加密和非对称加密的区别？TLS 为何混用？
3. 画证书链：根 CA → 中间 CA → 叶子证书，客户端如何验证？
4. HTTPS 默认端口？Go 开发常用端口？生产如何配合 Nginx？
5. 什么是混合内容？Active 和 Passive 区别？

### 进阶题

6. 对比 TLS 1.2 与 TLS 1.3 握手 RTT，说明 1.3 安全改进。
7. 解释 HSTS 防 SSL Stripping 的原理；`preload` 是什么？
8. 短链 API 若在 HTTP 上传 JWT，攻击者能做什么？如何用 curl 验证 HTTPS 已生效？
9. 说明 `Secure`、`HttpOnly`、`SameSite` 三个 Cookie 属性。
10. 用 `curl.exe -v` 访问 `https://httpbin.org/get`，记录 TLS 版本与是否 HTTP/2。

### 挑战题

11. 用 mkcert 或 openssl 为 localhost 生成证书，Go `ListenAndServeTLS` 开 8443，`curl -k -v` 访问。
12. 写完整 Nginx 配置：`80 → 301 https`，`443 http2` 反代 `8080`，并加 HSTS 头。

### 练习参考答案（节选）

**基础题 1**  
窃听 → 对称加密；篡改 → MAC/AEAD；冒充 → CA 证书链验证服务器身份。

**基础题 2**  
对称：同一密钥，快；非对称：公钥/私钥，慢但可解决密钥交换。TLS 握手用非对称协商 session key，传输用对称 AES-GCM。

**基础题 5**  
HTTPS 页面加载 HTTP 资源。**Active**：脚本、XHR → 阻塞。**Passive**：图片、音视频 → 可能警告。

**进阶题 8**  
攻击者嗅探拿到 Bearer token，可伪造创建短链等请求。修复：Nginx 443 + 证书，客户端改 `https://`；`curl -v` 应见 TLSv1.2/1.3 而非明文 HTTP。

---

## 19. 学完标准

完成本章后，你应能：

- [ ] **解释** 为什么生产环境必须 HTTPS，并与 Go JWT 传输关联
- [ ] **区分** 对称/非对称加密及在 TLS 中的分工
- [ ] **描述** CA 证书链验证流程
- [ ] **对比** TLS 1.2（~2 RTT）与 TLS 1.3（~1 RTT）握手
- [ ] **说明** 443 端口、Nginx TLS 终结、内网 HTTP 反代架构
- [ ] **理解** HTTP→HTTPS 301 重定向与 HSTS 作用
- [ ] **识别** 混合内容概念（浏览器面试语境）
- [ ] **使用** PowerShell + `curl.exe -v` 观察 TLS 握手与 HTTP/2
- [ ] **口述** Cookie Secure/HttpOnly 与 JWT Bearer 的安全要点
- [ ] **知道** HTTPS 不防 XSS / SQL 注入，还需应用层措施
- [ ] **闭卷自测** ≥ 8/10

---

## 20. 下一章预告

| 章节 | 主题 |
|------|------|
| [06 缓存 Cookie 与会话机制](./06-缓存Cookie与会话机制.md) | 强缓存/304、Cookie、JWT、CORS——与 Go 09 登录联调 |
| [07 面试专题](./07-面试专题与知识点总表.md) | 计网八股总复习 |
| [Go 05 net/http](../../后端学习/Go/05-Go标准库与HTTP基础.md) | 用 Go 写 HTTP 服务，把 04/05 章理论落地 |
| [Go 09 JWT](../../后端学习/Go/09-JWT认证与用户体系.md) | bcrypt + Bearer Token + Gin 鉴权中间件 |

**推荐顺序**：本章 → **Go 05**（写 HTTP 服务）→ 计网 **06**（Cookie/CORS）→ **Go 09**（JWT）。

---

## 附录 A：HTTPS 上线 Checklist（Go API）

```text
□ 证书有效且域名 SAN 匹配 api.example.com
□ 证书链完整（fullchain.pem 含中间证书）
□ 80 → 301 → 443
□ HSTS（可选 preload，确认后再加）
□ 公网 API 地址统一 https://
□ TLS 1.2+，禁用 1.0/1.1
□ Go 8080 仅内网监听，安全组不放行 8080
□ JWT 登录/刷新接口仅 HTTPS
□ 云安全组开放 443
□ curl.exe -v 验证 TLS 1.2/1.3 通过
```

---

## 附录 B：交叉链接汇总

| 主题 | 文档 |
|------|------|
| HTTP 报文、方法、状态码 | [04-HTTP协议深入](./04-HTTP协议深入.md) |
| TCP 三次握手、端口 | [02-TCP与UDP](./02-TCP与UDP.md) |
| DNS 与域名 | [03-IP地址与DNS解析](./03-IP地址与DNS解析.md) |
| Go net/http | [Go 05](../../后端学习/Go/05-Go标准库与HTTP基础.md) |
| Gin 中间件 | [Go 06](../../后端学习/Go/06-Gin框架核心与中间件.md) |
| JWT 认证 | [Go 09](../../后端学习/Go/09-JWT认证与用户体系.md) |
| Cookie / CORS | [06-缓存Cookie与会话机制](./06-缓存Cookie与会话机制.md) |
| Linux Nginx 部署 | [Linux 13](../../后端学习/Linux/13-Nginx与Web服务部署.md) |

---

## 附录 C：2 分钟面试口述稿（Go 后端版）

> HTTPS 在 HTTP 和 TCP 之间加了 TLS，提供加密、完整性和服务器身份验证。握手时用 ECDHE 等非对称算法协商 session key，之后用 AES-GCM 对称加密 HTTP 数据。证书由 CA 签发，客户端验证链到信任根。TLS 1.3 约 1-RTT，比 1.2 快。生产用 443，Nginx 终结 TLS 反代 Go 8080。JWT 放 Authorization 头里必须走 HTTPS 防嗅探；验签是应用层的事。HTTPS 不防 XSS 和 SQL 注入。我用 curl -v 看 TLS 版本和证书是否协商成功。

---

*本章已按 [修改规范.md](../../修改规范.md) 与 EXPANSION-STANDARD 编写：§0 导读、术语三件套、手把手步骤表、FAQ ≥12、闭卷自测 10 题、费曼检验、Mermaid ≥5、常见报错表 ≥12。*
