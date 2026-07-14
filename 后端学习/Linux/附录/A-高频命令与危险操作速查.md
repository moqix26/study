# 高频命令与危险操作速查

这份附录用于“已经做过实验，临时忘了命令”时查询。它不是入门章节，也不建议从头背诵。命令中的路径、服务名、网卡名必须先替换成自己的实际值。

## 1. 先确认环境

```bash
whoami
id
hostnamectl
cat /etc/os-release
uname -r
pwd
```

遇到教程和本机输出不一致，先确认三个变量：

1. 发行版和版本，例如 Ubuntu 24.04。
2. 当前用户是否有 sudo 权限。
3. 命令是在 Windows PowerShell、WSL2、容器还是云服务器里执行。

## 2. 文件与目录

```bash
pwd
ls -lah
ls -li
cd /path/to/dir
mkdir -p ./a/b/c
cp -a source destination
mv old new
install -d -m 0750 /opt/shortlink/releases
install -m 0755 ./shortlink-api /opt/shortlink/releases/v1/shortlink-api
stat file
file binary
realpath path
```

删除前采用固定动作：

```bash
pwd
realpath ./target
ls -la ./target
```

确认无误后才执行删除。不要把未经检查的变量直接交给 `rm -rf`：

```bash
# 危险反例
rm -rf "$TARGET"

# 至少验证非空、限定父目录，并拒绝根目录
[[ -n "${TARGET:-}" ]]
resolved="$(realpath -- "$TARGET")"
[[ "$resolved" == /opt/shortlink/releases/* ]]
[[ "$resolved" != / ]]
rm -rf -- "$resolved"
```

## 3. 查看与搜索文本

```bash
cat file
less file
head -n 30 file
tail -n 100 file
tail -F /var/log/nginx/error.log
grep -n --color=auto 'ERROR' app.log
grep -RIn --exclude-dir=.git '127.0.0.1:8080' .
sed -n '120,180p' file
awk '{print $1}' access.log
sort input | uniq -c | sort -nr
```

`tail -F` 会在日志轮转后继续跟踪同名新文件；`tail -f` 更偏向跟踪当前文件描述符。

## 4. 权限与用户

```bash
id shortlink
getent passwd shortlink
namei -l /opt/shortlink/current/shortlink-api
chmod 0750 directory
chmod 0640 /etc/shortlink/shortlink.env
chown root:shortlink /etc/shortlink/shortlink.env
sudo -l
getfacl path
```

常见模式：

| 模式 | 常见用途 | 风险边界 |
|------|----------|----------|
| `0755` | 公共可遍历目录、可执行程序 | 所有人可读和执行 |
| `0750` | 服务目录 | 仅 owner 和 group 可进入 |
| `0644` | 非敏感配置、静态文件 | 所有人可读，不适合秘密 |
| `0640` | 服务环境文件 | 配合专用组读取 |
| `0600` | 私钥、个人凭据 | 仅 owner 可读写 |

`chmod 777` 不是通用修复方案。权限错误要先用 `namei -l` 找出路径上哪一级目录阻止访问，再修正确切的 owner/group/mode。

## 5. 进程、信号与资源

```bash
ps -eo pid,ppid,user,stat,%cpu,%mem,etime,cmd --sort=-%cpu | head
pgrep -a shortlink-api
top
free -h
vmstat 1
df -hT
df -i
du -xhd1 /var | sort -h
ulimit -n
cat /proc/1234/limits
```

优雅停止优先发送 SIGTERM：

```bash
kill -TERM 1234
```

只有进程无响应、确认可以放弃清理时才考虑 SIGKILL：

```bash
kill -KILL 1234
```

不要用宽泛的 `pkill -f java`、`pkill -f go` 或 `killall` 处理生产服务；应让 systemd 根据明确的 unit 管理进程。

## 6. systemd 与 journal

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now shortlink.service
systemctl is-active shortlink.service
systemctl is-enabled shortlink.service
systemctl status shortlink.service --no-pager
sudo systemctl restart shortlink.service
sudo systemctl reload nginx
sudo systemctl cat shortlink.service
systemctl show shortlink.service -p MainPID -p User -p FragmentPath

journalctl -u shortlink.service -n 100 --no-pager
journalctl -u shortlink.service -f
journalctl -u shortlink.service --since '30 minutes ago'
journalctl -p err..alert -b --no-pager
journalctl --disk-usage
```

改 unit 文件后必须 `daemon-reload`。改 Nginx 配置则应先 `nginx -t`，再 reload。

## 7. 网络与 HTTP

```bash
ip -brief address
ip route
resolvectl status
getent hosts example.com
dig example.com A
ss -lntup
ss -lntp '( sport = :8080 )'
nc -vz 127.0.0.1 8080
curl --fail --show-error --silent http://127.0.0.1:8080/healthz
curl -v --max-time 5 http://127.0.0.1:8080/healthz
```

从外到内的常用排查顺序：

```text
DNS → 云安全组 → 主机防火墙 → Nginx → 127.0.0.1:8080 → Go 进程 → MySQL/Redis
```

从应用主机本地开始验证通常更快：

```bash
curl -v http://127.0.0.1:8080/healthz
ss -lntp '( sport = :8080 )'
systemctl status shortlink.service
journalctl -u shortlink.service -n 100
```

## 8. SSH 与传输

```bash
ssh -vvv shortlink-prod
ssh-keygen -t ed25519 -a 64
ssh-keygen -lf ~/.ssh/id_ed25519.pub
ssh-keygen -R old.example.com
scp ./shortlink-api shortlink-prod:/tmp/
rsync -a --info=progress2 ./dist/ shortlink-prod:/srv/site/
sftp shortlink-prod
```

删除或接受新的 host key 前，应先通过云控制台或其他可信渠道核对指纹。不要用 `StrictHostKeyChecking=no` 消除安全提示。

## 9. Docker 与 Compose

```bash
docker version
docker info
docker image ls
docker container ls -a
docker volume ls
docker network ls
docker inspect shortlink-lab-app-1
docker logs --tail 100 -f shortlink-lab-app-1
docker stats

docker compose config
docker compose pull
docker compose build --pull
docker compose up -d --wait
docker compose ps
docker compose logs --tail 100 app
docker compose down
```

慎用以下命令：

```bash
docker compose down -v
docker system prune -a --volumes
```

它们可能删除数据库卷或仍需使用的镜像。先运行 `docker system df`、`docker volume ls` 和 `docker compose config --volumes`，明确影响范围。

## 10. Nginx 与 TLS

```bash
sudo nginx -t
sudo systemctl reload nginx
curl -I http://127.0.0.1
curl -vk https://s.example.com/healthz
openssl s_client -connect s.example.com:443 -servername s.example.com </dev/null
sudo certbot certificates
sudo certbot renew --dry-run
```

502 不等于“后端一定没启动”。它还可能来自上游超时、上游提前断开、错误的 Unix socket 权限、DNS 解析或 TLS 上游配置。必须结合 Nginx error log 和直连 upstream 的结果判断。

## 11. 数据库备份

```bash
mysqldump --defaults-extra-file=/etc/shortlink/mysql-backup.cnf \
  --single-transaction --quick shortlink | gzip > shortlink.sql.gz
gzip -t shortlink.sql.gz
sha256sum shortlink.sql.gz
```

“生成了压缩文件”不等于备份有效。至少要验证 gzip、校验和，并定期恢复到独立测试库。生产备份还要有异机或对象存储副本。
