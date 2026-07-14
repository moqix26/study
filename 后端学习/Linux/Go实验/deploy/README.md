# 部署样例使用说明

这里的文件与 [09 Go 短链服务完整部署](../../09-Go短链服务完整部署.md) 配套。建议先在 Ubuntu VM 完成 systemd 路线，再尝试 Compose 路线；两条路线用于理解不同边界，不要求在同一台机器同时运行。

## 1. systemd 路线

### 1.1 构建 Linux 二进制

在 Go 实验根目录执行：

```bash
mkdir -p bin
version="v0.1.0"
commit="$(git rev-parse --short HEAD 2>/dev/null || printf none)"
build_time="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -trimpath \
  -ldflags "-s -w -X main.version=$version -X main.commit=$commit -X main.buildTime=$build_time" \
  -o bin/shortlink-api ./cmd/shortlink-demo
sha256sum bin/shortlink-api
```

ARM 云主机把 `GOARCH` 改为 `arm64`。不要不看架构就上传 amd64 二进制。

### 1.2 创建服务账号和目录

```bash
sudo useradd --system --home /var/lib/shortlink --create-home \
  --shell /usr/sbin/nologin shortlink

sudo install -d -o root -g shortlink -m 0750 /opt/shortlink/releases
sudo install -d -o root -g shortlink -m 0750 /etc/shortlink
sudo install -d -o shortlink -g shortlink -m 0750 /var/lib/shortlink
```

### 1.3 安装配置

```bash
sudo install -o root -g shortlink -m 0640 \
  deploy/shortlink.env.example /etc/shortlink/shortlink.env
sudoedit /etc/shortlink/shortlink.env
```

使用随机值替换占位符。配置文件模式 0640 允许 root 修改、shortlink 组读取，其他用户不可读。

### 1.4 安装 unit

```bash
sudo install -o root -g root -m 0644 \
  deploy/shortlink.service /etc/systemd/system/shortlink.service
sudo systemd-analyze verify /etc/systemd/system/shortlink.service
sudo systemctl daemon-reload
sudo systemctl enable shortlink.service
```

服务第一次还不能启动，因为 `/opt/shortlink/current` 尚未指向 release。

### 1.5 第一次发布

`deploy.sh` 需要写 `/opt/shortlink` 并重启服务。学习机可先用 sudo 执行：

```bash
sudo bash deploy/deploy.sh ./bin/shortlink-api v0.1.0
```

正式自动化应配置受限 sudoers，只允许部署账号执行明确的 `systemctl restart shortlink.service`，而不是授予无密码执行任意命令的 sudo。

验证：

```bash
readlink -f /opt/shortlink/current
systemctl status shortlink.service --no-pager
journalctl -u shortlink.service -n 50 --no-pager
curl --fail --silent --show-error http://127.0.0.1:8080/healthz
```

### 1.6 发布失败如何回滚

脚本先记录旧 `current` 指向，再安装新 release，通过临时软链接和 `mv -T` 原子替换。服务重启或健康检查失败时，会把 `current` 切回上一版本并重启。

它无法自动解决：

- 破坏性数据库迁移。
- 新版本已写入旧版本无法识别的数据。
- 外部依赖协议不兼容。

因此正式迁移必须遵守 expand/contract 等向后兼容策略。

## 2. Nginx 路线

VM 内没有域名时，使用 HTTP 配置：

```bash
sudo install -o root -g root -m 0644 \
  deploy/nginx/shortlink-http.conf /etc/nginx/sites-available/shortlink
sudo ln -sfn /etc/nginx/sites-available/shortlink /etc/nginx/sites-enabled/shortlink
sudo nginx -t
sudo systemctl reload nginx
```

云服务器有域名和证书后，复制 HTTPS 模板并替换全部 `s.example.com`。证书文件真实存在前，`nginx -t` 会失败，这是正确的保护。

应用保持监听 `127.0.0.1:8080`，公网只进入 Nginx 的 80/443。

## 3. Compose 路线

### 3.1 准备环境变量

```bash
cd deploy
cp .env.example .env
chmod 0600 .env
```

生成随机密码示例：

```bash
openssl rand -base64 36
```

分别填写 MySQL root、业务账号和 Redis 密码，不要三个服务复用同一个值。

### 3.2 验证与启动

```bash
docker compose config --quiet
docker compose build --pull
docker compose up -d --wait
docker compose ps
docker compose logs --tail 100 app
curl http://127.0.0.1:8080/healthz
```

Compose 有意不发布 3306/6379；MySQL、Redis 只能通过 `backend` 内部网络访问。App 的宿主机端口也只绑定到 127.0.0.1。

### 3.3 限制说明

`.env` 能避免密码写进 YAML 和 Git，但密码仍会出现在容器环境或 inspect 元数据中。企业环境应使用 Docker secrets、云密钥服务或编排平台的 secret 机制。

`docker compose down` 默认保留命名卷；`docker compose down -v` 会删除数据库卷，不应在有价值数据的环境随意执行。

## 4. 数据库备份与恢复演练

建立仅供备份使用的 MySQL option file：

```ini
[client]
host=127.0.0.1
user=shortlink_backup
password=replace-me
```

```bash
sudo install -o root -g root -m 0600 mysql-backup.cnf /etc/shortlink/mysql-backup.cnf
sudo bash deploy/mysql-backup.sh
sudo bash deploy/restore-check.sh /var/backups/shortlink/mysql/shortlink.TIMESTAMP.sql.gz
```

备份账号应具备完成 dump 所需的最小权限。恢复检查必须使用独立数据库名，脚本会拒绝把检查库设成源库。

## 5. 清理实验

先确认当前环境没有有价值数据，再逐项清理：

```bash
sudo systemctl disable --now shortlink.service
sudo rm -f /etc/systemd/system/shortlink.service
sudo systemctl daemon-reload

docker compose down
```

是否删除 `/opt/shortlink`、`/etc/shortlink`、数据库卷和备份必须由你单独确认，本资料不提供“一键全删”脚本。
