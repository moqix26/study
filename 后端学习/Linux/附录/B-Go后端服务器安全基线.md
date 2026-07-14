# Go 后端服务器安全基线

本附录面向单台 Ubuntu 24.04 云服务器上的学习项目。它不是企业完整合规规范，但能避免简历 demo 最常见的高风险错误。

## 1. 先定义信任边界

推荐的单机结构：

```text
公网
  ↓ 80/443
Nginx
  ↓ 127.0.0.1:8080
Go / Gin API
  ↓ 127.0.0.1 或 Docker 内部网络
MySQL + Redis
```

默认公网暴露：

- 80/TCP：用于 HTTP 跳转或 ACME 验证。
- 443/TCP：正式 HTTPS 流量。
- 22/TCP：只允许你的固定 IP；IP 不固定时至少启用密钥、限速和日志监控。

默认不公网暴露：

- Go 应用 8080。
- MySQL 3306。
- Redis 6379。
- Docker daemon 2375/2376。
- 调试、pprof、管理后台和内部 metrics 端口。

## 2. 账号与 SSH

1. 使用普通 sudo 用户管理服务器，不直接使用 root 做日常操作。
2. 为应用创建不可登录的专用系统用户：

```bash
sudo useradd --system --home /var/lib/shortlink --create-home \
  --shell /usr/sbin/nologin shortlink
```

3. 客户端密钥使用 Ed25519，并设置 passphrase：

```bash
ssh-keygen -t ed25519 -a 64
```

4. 首次连接核对服务器 host key 指纹，不盲目输入 yes。
5. 确认密钥登录成功并保留救援会话后，再关闭密码和交互式认证：

```text
PubkeyAuthentication yes
PasswordAuthentication no
KbdInteractiveAuthentication no
PermitRootLogin no
```

6. 使用 `sshd -t` 检查语法，用 `sshd -T` 检查最终生效配置；云镜像可能通过 `sshd_config.d` 覆盖主文件。

## 3. 最小权限文件布局

```text
/opt/shortlink/releases/<version>/  root:shortlink 0750
/opt/shortlink/current              原子软链接
/etc/shortlink/shortlink.env        root:shortlink 0640
/var/lib/shortlink/                 shortlink:shortlink 0750
/var/backups/shortlink/             root:root 0700
```

服务进程只需要读取二进制和配置，并写入明确的数据目录。不要把整个 `/opt`、`/etc` 或 `/var/log` chown 给登录用户。

## 4. systemd 隔离

最低建议：

```ini
User=shortlink
Group=shortlink
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
CapabilityBoundingSet=
RestrictSUIDSGID=true
UMask=0027
```

每增加一个限制，都要通过实际运行和日志确认没有阻断合法功能。不要为了“安全参数更多”盲目复制不理解的沙箱项。

`EnvironmentFile=` 适合学习项目，但环境变量并不是秘密保险箱：root、调试工具或具有相应权限的进程仍可能读取。更高要求可使用 systemd credentials、云密钥服务或专门的 secret manager。

## 5. 绑定地址

绑定地址取决于进程边界：

| 场景 | 推荐监听 |
|------|----------|
| 同机 Nginx → Go | `127.0.0.1:8080` |
| 容器内 Go | `0.0.0.0:8080`，但宿主机映射为 `127.0.0.1:8080:8080` |
| 局域网实验直接访问 | 具体 VM IP，防火墙仅放宿主机网段 |
| 公网直接暴露 Go | 通常不推荐；至少要 TLS、限流和完整安全配置 |

`0.0.0.0` 表示监听所有本机 IPv4 接口，不等于“互联网一定能访问”，也不等于“生产必须这样配置”。

## 6. MySQL 与 Redis

- 业务账号只授予自己的库，例如 `shortlink.*`，不要授予 `*.*`。
- 不允许远程 root 登录。
- 不把 3306 和 6379 发布到公网。
- Redis 即使设置密码，也不应把它当成可安全暴露公网的数据库。
- 容器网络只提供网络隔离，不代替身份认证、备份和最小权限。
- MySQL 账号、Redis 密码、JWT 密钥必须随机生成，示例密码不能进入生产。

MySQL 最小授权示例：

```sql
CREATE DATABASE shortlink CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
CREATE USER 'shortlink'@'localhost' IDENTIFIED BY 'replace-with-random-password';
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, INDEX
ON shortlink.* TO 'shortlink'@'localhost';
```

项目稳定后，应把建表权限与运行期读写权限进一步拆分，由迁移账号执行 schema migration。

## 7. Docker

- `docker` 组基本等价于宿主机 root 权限，加入前必须理解风险。
- 容器进程使用非 root 用户。
- 尽量使用只读根文件系统、`cap_drop: [ALL]` 和 `no-new-privileges`。
- 数据库使用命名卷，并有独立备份；volume 不是备份。
- 镜像 tag 至少固定到明确版本，正式环境记录 digest。
- 不把 `.env`、私钥、云凭据复制进镜像层。
- 不把 Docker socket 挂进普通业务容器。

## 8. Nginx 与 TLS

- HTTP 自动跳转 HTTPS。
- 只启用 TLS 1.2/1.3。
- HSTS 仅在确认 HTTPS 长期可用后开启；错误启用可能让浏览器长期拒绝 HTTP 回退。
- 限制请求体大小、连接和读取超时。
- 正确传递 `Host`、`X-Forwarded-For`、`X-Forwarded-Proto`。
- Gin 只信任明确的代理地址，否则客户端可以伪造转发头。
- 限流用于保护入口，不替代业务鉴权和配额。

## 9. 日志与隐私

禁止记录：

- Authorization、Cookie、密码、JWT 完整值。
- MySQL DSN 中的密码。
- 身份证、手机号等不必要的个人信息。

建议记录：

- request_id。
- HTTP 方法、规范化路由、状态码、耗时、响应字节数。
- 版本、commit、启动时间和关停原因。

日志保留必须有上限。使用 journald 配额、logrotate 或容器日志轮转，防止磁盘被写满。

## 10. 更新、备份与恢复

- 安全更新前先确认变更范围和维护窗口，不在发布脚本中无条件执行整机 `apt upgrade`。
- 数据库备份必须加密、限制权限，并存一份到另一台机器或对象存储。
- 定期做恢复演练；没有恢复测试的备份只能算“可能可用”。
- 发布使用版本化目录与原子软链接，健康检查失败自动回滚。
- 数据库迁移要考虑向后兼容；仅回滚二进制不能自动回滚破坏性 schema。

## 11. 上线前检查表

- [ ] SSH 密钥登录成功，密码登录和 root 登录已按计划关闭。
- [ ] 云安全组的 22 仅允许可信来源。
- [ ] 只有 80/443 对公网开放。
- [ ] Go 进程使用专用用户，监听 loopback 或内部网络。
- [ ] MySQL/Redis 没有公网端口。
- [ ] 配置文件权限正确，Git 中无真实秘密。
- [ ] Nginx 配置通过 `nginx -t`。
- [ ] TLS 证书有效，续期 dry-run 通过。
- [ ] `/healthz`、`/readyz`、业务 smoke test 均通过。
- [ ] 重启服务器后服务能够恢复。
- [ ] 备份生成、校验和恢复演练均成功。
- [ ] 日志不会记录密码和 token，且有轮转或配额。
