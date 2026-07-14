# Nginx、TLS 与反向代理：给 Go 短链服务建立公网边界

Go/Gin 可以直接监听 HTTP，但公网服务通常仍在前面放一层 Nginx。它负责：

- 监听公网 `80/443`；
- 终止 TLS，管理证书；
- 把请求转发给只监听回环地址的 Go 服务；
- 统一请求头、超时、限流、安全响应头和访问日志；
- 在不暴露应用端口的前提下提供稳定入口。

本章使用以下边界：

```text
Client
  |
  | HTTPS :443
  v
Nginx                         公网边界
  |
  | HTTP 127.0.0.1:8080      仅本机
  v
Go / Gin shortlink-server
  |
  +---- MySQL / Redis         不开放公网
```

如果 Go 服务原生运行在宿主机，它应监听 `127.0.0.1:8080`。如果 Go 服务在 Docker 中，它在容器内监听 `0.0.0.0:8080`，但 Compose 只发布 `127.0.0.1:8080:8080`。两种部署最终都让 Nginx 使用相同 upstream。

---

## 1. Nginx 和 Gin 的职责边界

| 能力 | Nginx 更适合 | Gin 更适合 |
|---|---:|---:|
| TLS 证书与协议 | 是 | 通常否 |
| 公网端口监听 | 是 | 仅内部端口 |
| 通用请求大小/连接限制 | 是 | 可做第二层校验 |
| 业务鉴权 | 否 | 是 |
| 短码查询和权限判断 | 否 | 是 |
| 参数校验和业务错误 JSON | 否 | 是 |
| 统一访问日志 | 是 | 也应有应用日志 |
| 数据库和缓存访问 | 否 | 是 |
| 业务级限流 | 粗粒度 | 精细规则更适合 |

不要把业务逻辑塞进 Nginx rewrite，也不要让 Gin 自己承担所有公网防护。二者是分层，不是二选一。

对短链服务尤其要注意：跳转状态码和缓存策略属于业务语义。默认优先由 Go 返回 `302` 或 `307`。`301/308` 可能被浏览器和中间缓存长期记住，一旦短链目标需要修改、禁用或风控拦截，客户端仍可能绕过服务器跳到旧地址。

---

## 2. 上线前的前置条件

### 2.1 域名与 DNS

假设短链域名是 `s.example.com`：

1. 域名已完成需要的注册、实名或备案流程。
2. DNS `A` 记录指向服务器 IPv4。
3. 如果配置 `AAAA`，服务器必须真的具备可用 IPv6；否则部分客户端会优先走坏掉的 IPv6。
4. 云安全组和主机防火墙允许 TCP 80、443。
5. Go 服务已在本机 `127.0.0.1:8080` 健康运行。

验证：

```bash
dig +short A s.example.com
dig +short AAAA s.example.com
curl -fsS http://127.0.0.1:8080/healthz
ss -lntp | grep -E ':(80|443|8080)\b'
```

DNS 修改后存在 TTL 缓存，不同递归 DNS 看到新地址的时间可能不同。申请公网证书前必须确保域名确实解析到当前服务器。

### 2.2 端口策略

建议的入站规则：

| 端口 | 来源 | 用途 |
|---|---|---|
| 22 | 你的固定 IP 或受控跳板机 | SSH |
| 80 | 公网 | HTTP 跳转和 ACME 验证 |
| 443 | 公网 | HTTPS |
| 8080 | 不开放公网 | Nginx 到 Go |
| 3306 | 不开放公网 | MySQL |
| 6379 | 不开放公网 | Redis |

UFW 示例：

```bash
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow from 你的管理IP to any port 22 proto tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
sudo ufw status verbose
```

启用防火墙前必须先确认 SSH 规则有效，否则可能把自己锁在服务器外。云服务器还要同步检查云安全组；UFW 放行不代表云安全组已放行，反之亦然。

---

## 3. 安装并认识配置结构

Ubuntu：

```bash
sudo apt update
sudo apt install -y nginx
sudo systemctl enable --now nginx
nginx -v
systemctl status nginx --no-pager
```

常见目录：

```text
/etc/nginx/nginx.conf                 主配置，包含 http 块
/etc/nginx/conf.d/*.conf              http 级公共配置
/etc/nginx/sites-available/            站点配置源文件
/etc/nginx/sites-enabled/              已启用站点的符号链接
/var/log/nginx/access.log              默认访问日志
/var/log/nginx/error.log               错误日志
/var/www/                              静态文件/ACME 目录
```

修改前查看 Nginx 实际加载的完整配置：

```bash
sudo nginx -T | less
```

`nginx -T` 比只打开某一个文件更可靠，因为错误可能来自被 include 的其他文件。

禁用 Ubuntu 默认站点时只删除启用链接，不必删除源文件：

```bash
sudo rm /etc/nginx/sites-enabled/default
```

执行前确认它确实是默认站点链接，并确保你已经准备好自己的站点配置。

---

## 4. 先用纯 HTTP 打通反向代理

在申请证书前，先确认 Nginx 能通过回环地址访问 Go。创建 `/etc/nginx/sites-available/shortlink`：

```nginx
upstream shortlink_backend {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    listen [::]:80;
    server_name s.example.com;

    access_log /var/log/nginx/shortlink_access.log;
    error_log  /var/log/nginx/shortlink_error.log warn;

    location / {
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Request-ID $request_id;
        proxy_set_header Connection "";

        proxy_connect_timeout 2s;
        proxy_send_timeout 10s;
        proxy_read_timeout 10s;

        proxy_pass http://shortlink_backend;
    }
}
```

启用并检查：

```bash
sudo ln -s /etc/nginx/sites-available/shortlink \
  /etc/nginx/sites-enabled/shortlink
sudo nginx -t
sudo systemctl reload nginx
curl -v -H 'Host: s.example.com' http://127.0.0.1/healthz
curl -v http://s.example.com/healthz
```

每次配置变更都遵循：

```bash
sudo nginx -t && sudo systemctl reload nginx
```

`reload` 会让新 worker 使用新配置，旧 worker 尽量处理完已有连接后退出，通常比无必要的 `restart` 更平滑。如果 `nginx -t` 失败，后面的 reload 不会执行。

### 4.1 `proxy_pass` 末尾斜杠不是装饰

下面两种写法在带前缀的 location 中语义不同：

```nginx
location /api/ {
    proxy_pass http://shortlink_backend;
}
```

请求 `/api/v1/links` 通常保持原 URI 转发。

```nginx
location /api/ {
    proxy_pass http://shortlink_backend/;
}
```

这会用 `/` 替换匹配到的 `/api/` 前缀，请求可能变成 `/v1/links`。除非你明确要去前缀，否则不要随意多加斜杠。

本章让 Gin 路由与公网路径一致，因此使用不带 URI 部分的 `proxy_pass http://shortlink_backend;`。

### 4.2 为什么传这些请求头

- `Host`：让应用知道用户访问的域名，生成短链时必须使用受信任配置，不能盲目信任任意 Host。
- `X-Real-IP`：直接客户端地址。
- `X-Forwarded-For`：代理链地址列表。
- `X-Forwarded-Proto`：告诉应用外部连接是 HTTPS，避免生成 HTTP 链接。
- `X-Request-ID`：为一次请求提供跨 Nginx、Gin、数据库慢日志的关联 ID。
- 清空 `Connection`：允许 Nginx 与 upstream 使用 HTTP/1.1 keepalive。

应用生成短链的公开域名最好来自固定配置，例如 `PUBLIC_BASE_URL=https://s.example.com`，而不是直接拼接用户可控的 `Host` 请求头，否则可能产生 Host Header Injection。

---

## 5. 申请和续期 TLS 证书

使用 Let's Encrypt 和 Certbot：

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d s.example.com
```

选择把 HTTP 重定向到 HTTPS。Certbot 会验证域名、取得证书并修改 Nginx 配置。完成后验证：

```bash
sudo nginx -t
curl -I http://s.example.com
curl -I https://s.example.com
sudo certbot certificates
```

自动续期通常由 systemd timer 执行：

```bash
systemctl list-timers | grep certbot
systemctl status certbot.timer --no-pager
sudo certbot renew --dry-run
```

证书自动续期必须满足：

- 证书验证方式仍然可用；
- DNS 仍指向正确服务器；
- 80 端口没有被临时关闭；
- Nginx 配置能通过测试并重新加载；
- 定时器实际启用。

不要等证书过期当天才验证。可以为剩余有效期设置监控告警。

### 5.1 如果域名前面还有 CDN

CDN 代理可能改变证书验证和客户端 IP 获取方式：

- 确认 CDN 到源站也使用严格 TLS 校验，不要选择“前端 HTTPS、回源 HTTP”的宽松模式。
- HTTP-01 验证必须能到达正确源站，或改用受控 DNS-01 验证。
- 只有来自 CDN 官方 IP 段的请求头才可信，真实 IP 配置见第 9 节。
- 源站防火墙可进一步限制只接受 CDN 回源地址，但要预留证书验证和运维路径。

---

## 6. 一份适合短链 API 的 HTTPS 配置

为了避免在多个 location 中重复代理参数，先创建 `/etc/nginx/snippets/shortlink-proxy.conf`：

```nginx
proxy_http_version 1.1;
proxy_set_header Host $host;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Proto $scheme;
proxy_set_header X-Request-ID $request_id;
proxy_set_header Connection "";

proxy_connect_timeout 2s;
proxy_send_timeout 10s;
proxy_read_timeout 10s;

proxy_buffering on;
proxy_request_buffering on;
proxy_intercept_errors off;
```

再把站点整理成下面的形式。证书路径以 Certbot 实际生成结果为准：

```nginx
upstream shortlink_backend {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    listen [::]:80;
    server_name s.example.com;

    return 308 https://s.example.com$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name s.example.com;

    ssl_certificate     /etc/letsencrypt/live/s.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/s.example.com/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    server_tokens off;
    client_max_body_size 32k;

    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    add_header Referrer-Policy "no-referrer" always;
    add_header Permissions-Policy "camera=(), microphone=(), geolocation=()" always;
    add_header X-Request-ID $request_id always;

    access_log /var/log/nginx/shortlink_access.log shortlink_json;
    error_log  /var/log/nginx/shortlink_error.log warn;

    location = /healthz {
        allow 127.0.0.1;
        allow ::1;
        deny all;

        include snippets/shortlink-proxy.conf;
        proxy_pass http://shortlink_backend;
    }

    location = /readyz {
        allow 127.0.0.1;
        allow ::1;
        deny all;

        include snippets/shortlink-proxy.conf;
        proxy_pass http://shortlink_backend;
    }

    location = /api/v1/links {
        limit_req zone=shortlink_create burst=10 nodelay;
        limit_req_status 429;

        include snippets/shortlink-proxy.conf;
        proxy_pass http://shortlink_backend;
    }

    location / {
        include snippets/shortlink-proxy.conf;
        proxy_pass http://shortlink_backend;
    }
}
```

Ubuntu 24.04 自带的 Nginx 通常使用 `listen 443 ssl http2;`。较新的 Nginx 会提示该参数写法逐步弃用，并推荐单独写 `http2 on;`。应根据 `nginx -v` 选择一种写法，并以 `nginx -t` 为准，不要把两种写法同时叠加。Certbot 生成的 options 和 dhparam 路径也应以本机实际文件为准。

HTTP 到 HTTPS 使用 `308`，可以在重定向时保留 POST 等请求的方法和请求体；这和短链业务跳转使用 302/307 的缓存语义是两件事。

### 6.1 限流区和 JSON 日志必须放在 http 上下文

创建 `/etc/nginx/conf.d/shortlink-shared.conf`：

```nginx
limit_req_zone $binary_remote_addr
    zone=shortlink_create:10m
    rate=5r/s;

log_format shortlink_json escape=json
    '{'
      '"time":"$time_iso8601",'
      '"request_id":"$request_id",'
      '"remote_addr":"$remote_addr",'
      '"method":"$request_method",'
      '"uri":"$uri",'
      '"status":$status,'
      '"bytes_sent":$body_bytes_sent,'
      '"request_time":$request_time,'
      '"upstream_addr":"$upstream_addr",'
      '"upstream_status":"$upstream_status",'
      '"upstream_connect_time":"$upstream_connect_time",'
      '"upstream_response_time":"$upstream_response_time",'
      '"user_agent":"$http_user_agent"'
    '}';
```

Ubuntu 默认会在 `http {}` 中 include `conf.d/*.conf`，因此这里可以声明 `limit_req_zone` 和 `log_format`。执行 `sudo nginx -T` 确认实际 include 位置。

日志使用 `$uri` 而不是 `$request_uri`，避免默认记录查询字符串中的 token 或长 URL。更根本的规则是：认证信息不要放 URL 查询参数；创建短链的原始 URL 放 JSON 请求体，应用日志也不要原样打印敏感参数。

### 6.2 限流只是第一层

Nginx 的 IP 限流能缓解简单滥用，但不能处理账号套餐、API key 配额、分布式实例总额度等业务规则。短链创建接口还应在应用层结合账号、租户或 token 做限流。

`5r/s` 和 `burst=10` 只是示例，不是万能参数。应根据真实流量、压测和误伤成本调整。短链跳转接口通常读流量远大于创建接口，不应照搬同一个阈值。

---

## 7. 安全响应头如何选择

### 7.1 HSTS 要逐步启用

当 HTTPS 已稳定运行后，可以先用很短的时间观察：

```nginx
add_header Strict-Transport-Security "max-age=300" always;
```

确认所有入口、子域和证书续期都可靠后，再逐步提高到一年：

```nginx
add_header Strict-Transport-Security "max-age=31536000" always;
```

不要一开始就添加 `includeSubDomains; preload`。一旦预加载或长时间缓存，子域中尚未支持 HTTPS 的服务也可能无法访问，回退成本很高。

HSTS 只能放在 HTTPS 响应，不需要放在 80 端口跳转响应。

### 7.2 CSP 不能复制一行就算完成

如果域名只提供 JSON API 和跳转，可考虑非常严格的 CSP：

```nginx
add_header Content-Security-Policy "default-src 'none'; frame-ancestors 'none'; base-uri 'none'" always;
```

但将来若同域名托管管理前端，这条规则会阻止脚本、样式和资源加载。CSP 应根据实际页面资源制定并先使用 Report-Only 观察，不能机械加入。

### 7.3 不要让 CORS 与认证冲突

Nginx 可以加 CORS 头，但业务 API 更适合由 Gin 根据环境维护明确 allowlist。不要组合：

```text
Access-Control-Allow-Origin: *
Access-Control-Allow-Credentials: true
```

浏览器不接受这种组合，它也会模糊真实安全边界。服务端到服务端请求不受浏览器 CORS 保护，鉴权仍必须独立实现。

---

## 8. 超时、缓冲和请求大小

本章示例使用：

```nginx
proxy_connect_timeout 2s;
proxy_send_timeout 10s;
proxy_read_timeout 10s;
client_max_body_size 32k;
```

含义：

- `proxy_connect_timeout`：Nginx 与 Go 建立连接最多等多久。
- `proxy_send_timeout`：向 upstream 发送请求时，两次写操作之间允许的最长空闲时间。
- `proxy_read_timeout`：等待 upstream 两次读取操作之间的最长空闲时间，并不简单等于“整个请求总耗时”。
- `client_max_body_size`：过大请求在入口处拒绝，短链创建通常不需要大 body。

这些时间必须和应用、数据库、Redis 的超时形成层级。例如：

```text
数据库单次查询超时 1s
应用请求上下文 3s
Nginx upstream read timeout 5s
客户端总超时 8s
```

具体数字应通过业务 SLA 和压测决定，但内层通常应比外层更早、以可控错误结束。若数据库查询允许 30 秒，Nginx 却 5 秒就放弃，应用可能继续消耗资源处理一个客户端已经收不到的请求。

短链普通 API 可以启用请求/响应缓冲。Server-Sent Events、长轮询、流式下载或 WebSocket 需要单独配置，不能直接套用这里的超时和 buffering。

---

## 9. 客户端 IP 与可信代理

### 9.1 服务器直接暴露 Nginx

当客户端直接连接源站 Nginx 时，`$remote_addr` 就是对端地址。Nginx 应覆盖发给应用的 `X-Real-IP`，并用 `$proxy_add_x_forwarded_for` 追加代理链。

Gin 不能无条件信任任意代理头。原生 systemd 部署时，直接代理通常来自回环地址：

```go
if err := router.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
    return err
}
```

Docker 发布回环端口时，容器里看到的直接对端可能是 Docker bridge gateway，而不是 `127.0.0.1`。先通过访问日志和下面的命令确认：

```bash
docker network inspect shortlink_edge
```

然后只信任实际的代理地址或受控网段。不要为了省事使用 `0.0.0.0/0`，否则客户端可能伪造 `X-Forwarded-For` 绕过 IP 限流或污染审计日志。

### 9.2 Nginx 前面还有负载均衡/CDN

这时 Nginx 的直接对端是负载均衡器。只有在明确限制可信来源后，才能启用类似配置：

```nginx
# 以下网段只是语法示例，必须替换为供应商当前官方地址段
set_real_ip_from 192.0.2.0/24;
real_ip_header X-Forwarded-For;
real_ip_recursive on;
```

错误地信任全网会让攻击者自己声明“真实 IP”。供应商 IP 段还会变化，必须有更新流程。

---

## 10. 访问日志与请求关联

一次请求最好在三个位置使用同一个 request ID：

```text
Nginx access log
    request_id=abc
        |
        v
Gin structured log
    request_id=abc, route=/:code
        |
        v
数据库慢查询/外部调用日志
    request_id=abc（能传则传，不能传就记录操作上下文）
```

本章由 Nginx 在边缘生成 `$request_id` 并覆盖传给应用。Gin 中间件应读取 `X-Request-ID`、加入 context 和响应头，结构化记录：

- method、route 模板，而不是只记录带实际短码的原始路径；
- status、latency、request ID；
- 经过脱敏的用户/租户 ID；
- 错误类别，不记录密码、token 和完整原始 URL。

查看日志：

```bash
sudo tail -f /var/log/nginx/shortlink_access.log
sudo tail -f /var/log/nginx/shortlink_error.log
sudo journalctl -u shortlink.service -f
```

不要用“错误日志没有内容”证明服务正常。HTTP 500 可能是应用按正常 HTTP 流程返回，Nginx error log 未必记录；必须同时看 access log 的状态码和 upstream 时间。

---

## 11. 502、504 和 TLS 故障怎么查

### 11.1 502 Bad Gateway

502 表示 Nginx 没有从 upstream 得到有效响应。按链路查：

```bash
sudo nginx -t
curl -v --max-time 3 http://127.0.0.1:8080/healthz
ss -lntp | grep ':8080'
systemctl status shortlink.service --no-pager
sudo journalctl -u shortlink.service -n 100 --no-pager
sudo tail -n 100 /var/log/nginx/shortlink_error.log
```

常见根因：

- Go 服务没有启动或反复退出；
- upstream 地址/端口写错；
- 服务监听了错误接口；
- Unix/文件权限或 SELinux/AppArmor 限制；
- Go 在响应头发出前崩溃或关闭连接；
- Nginx 代理到了另一个旧进程。

### 11.2 504 Gateway Timeout

504 表示连接通常已建立，但 Nginx 在超时内没有收到所需响应。查：

- access log 的 `$request_time` 和 `$upstream_response_time`；
- Gin 请求日志和 pprof/指标；
- MySQL 慢查询、锁等待、连接池；
- Redis/外部服务超时；
- Go goroutine 是否泄漏或阻塞。

不要只把 `proxy_read_timeout` 从 10 秒改成 300 秒。延长超时可能只是让用户等得更久，同时积压更多连接。先找慢在哪里。

### 11.3 证书错误

```bash
openssl s_client \
  -connect s.example.com:443 \
  -servername s.example.com \
  -showcerts </dev/null

curl -vI https://s.example.com
sudo certbot certificates
date -u
```

检查证书域名、有效期、完整链、服务器时间和 SNI。直接用 IP 访问 HTTPS 时证书名通常不匹配，这并不表示域名访问也有问题。

### 11.4 HTTP 正常，HTTPS 连不上

依次确认：

1. `ss -lntp` 是否有 443；
2. `nginx -t` 是否成功；
3. UFW 是否放行 443；
4. 云安全组是否放行 443；
5. 域名是否包含错误 AAAA 记录；
6. 证书文件权限和路径是否正确。

---

## 12. 不建议一开始启用的“优化”

### 12.1 对短链跳转做 Nginx 缓存

缓存可以提高吞吐，但会引入失效问题：短链被禁用、过期、目标修改或进入风控名单后，缓存中的旧跳转可能继续生效。没有设计缓存 key、TTL、主动失效和一致性前，先让 Gin + Redis 负责缓存语义。

### 12.2 用 301 代替 302

301 看起来更“正式”，但浏览器可能长期缓存。可变、可封禁或需要统计点击的短链，通常先用 302。只有永久不可变且清楚缓存影响时才考虑 301/308。

### 12.3 随意信任 CDN 请求头

`X-Forwarded-For`、`X-Real-IP`、`CF-Connecting-IP` 都只是 HTTP 头。只有当请求确定来自受信任代理地址时，它们才可作为真实 IP 使用。

### 12.4 为掩盖慢请求无限增大 timeout

超时是资源边界，不是故障修复。应先通过日志、指标和慢查询找出瓶颈。

---

## 13. 变更与回滚流程

每次改 Nginx：

1. 备份或通过 Git/配置管理保存当前文件。
2. 修改 `sites-available`、snippet 或 `conf.d`。
3. 执行 `sudo nginx -t`。
4. 用 `sudo nginx -T` 确认实际配置。
5. `sudo systemctl reload nginx`。
6. 从本机和公网分别验证 HTTP、HTTPS、API、短链跳转。
7. 检查 error log 和 access log。

如果 reload 后行为异常，恢复上一版配置，再执行测试和 reload。不要在故障中反复无记录地手改多处文件，否则很难知道哪一项真正起作用。

最低验收命令：

```bash
curl -I http://s.example.com/example
curl -I https://s.example.com/example
curl -fsS https://s.example.com/api/v1/version
curl -fsS http://127.0.0.1:8080/readyz
sudo nginx -t
sudo certbot renew --dry-run
```

短链响应还应检查 `Location` 和状态码：

```bash
curl -sS -D - -o /dev/null https://s.example.com/abc123
```

---

## 14. 亲手实验

### 实验一：证明 Go 端口没有公网暴露

1. 在服务器本机 curl `127.0.0.1:8080`，应成功。
2. 在另一台机器 curl `服务器IP:8080`，应失败。
3. 通过 `https://s.example.com` 访问，应成功。
4. 用 `ss`、UFW 和云安全组解释三种结果。

### 实验二：制造并定位 502

1. 停止 Go 服务。
2. 访问 HTTPS，观察 502。
3. 查看 Nginx error log 中的 upstream 错误。
4. 用本机 curl 证明 8080 不通。
5. 启动 Go 服务并验证恢复。

### 实验三：观察 504 与应用超时

在测试路由中模拟一个超过 Nginx timeout 的处理，比较：

- 客户端耗时；
- Nginx access log 的 request/upstream time；
- Gin 是否在客户端断开后继续工作；
- 正确传播 request context 后有何变化。

实验只在本地或测试环境进行，不要在生产服务故意 sleep。

### 实验四：证书续期演练

运行 `certbot renew --dry-run`，检查 timer、日志和 Nginx reload。只有 dry-run 成功，才能认为自动续期链路基本可用。

---

## 15. 学完本章应能回答

1. 为什么 Go 服务应只通过回环地址暴露给 Nginx？
2. `proxy_pass` 末尾有无斜杠为什么可能改变 URI？
3. `X-Forwarded-For` 在什么条件下可信？
4. 为什么应用生成公开短链时不应直接信任 Host？
5. 502 与 504 分别说明链路的大致哪一段有问题？
6. 为什么 301 可能让短链禁用和目标修改失效？
7. HSTS 为什么不能第一天就 preload 全部子域？
8. Nginx 限流为什么不能代替账号/API key 级业务限流？
9. 为什么请求 ID 对跨层排障重要？
10. `nginx -t && systemctl reload nginx` 比直接 restart 好在哪里？

当你能独立完成 TLS 申请、反向代理、端口验证、一次 502 排查和一次续期演练时，这一章才算真正完成。
