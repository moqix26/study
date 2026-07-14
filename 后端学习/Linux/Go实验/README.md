# Linux 配套 Go 实验

本目录不是新的业务项目，而是 Linux 章节的可运行验证材料。它解决一个常见问题：文档里的命令看懂了，但手上没有一个行为明确、能够被 systemd、Docker、Nginx 管理的 Go 进程。

## 实验目录

| 路径 | 目的 | 对应章节 |
|------|------|----------|
| `cmd/shortlink-demo` | Gin API、健康检查、JSON 日志、优雅退出、版本信息 | 02、03、08、09、10 |
| `cmd/signal-demo` | 直接观察 SIGINT / SIGTERM 与清理过程 | 02 |
| `cmd/env-check` | 验证环境变量和敏感文件权限，不打印秘密本身 | 04、06、09 |
| `deploy/shortlink.service` | 非 root 的 systemd 服务与基础沙箱 | 02、09 |
| `deploy/Dockerfile` | 多阶段构建、非 root 运行 | 07 |
| `deploy/compose.yaml` | App + MySQL + Redis 的内部网络与健康检查 | 07、09 |
| `deploy/nginx/shortlink-http.conf` | 本地 VM 的 HTTP 反向代理 | 08 |
| `deploy/nginx/shortlink-https.conf` | 云服务器 TLS 入口模板 | 08、09 |
| `deploy/deploy.sh` | 版本化 release、原子切换、健康检查、失败回滚 | 05、09 |

## 本机运行

```powershell
cd F:\study\后端学习\Linux\Go实验
go mod download
go run ./cmd/shortlink-demo
```

默认只监听 `127.0.0.1:8080`，适合与同机 Nginx 配合：

```powershell
Invoke-RestMethod http://127.0.0.1:8080/healthz
Invoke-RestMethod -Uri http://127.0.0.1:8080/api/links `
  -Method POST `
  -ContentType 'application/json' `
  -Body '{"url":"https://go.dev/doc/"}'
```

## Linux 构建

```bash
cd ~/study/后端学习/Linux/Go实验
go test ./...
go build -trimpath -ldflags "-s -w -X main.version=lab" \
  -o ./bin/shortlink-api ./cmd/shortlink-demo
./bin/shortlink-api
```

收到 `SIGTERM` 时，程序先把 readiness 设为失败，再等待正在处理的请求结束：

```bash
kill -TERM "$(pgrep -f '/shortlink-api$')"
```

## 明确边界

- 数据存储在内存，重启后会清空；它用于练部署，不代替正式短链项目。
- Compose 中 MySQL、Redis 用于练环境与网络；演示服务不会把“端口通”冒充“业务已正确使用数据库”。
- 示例密码必须通过本地 `.env` 注入，`.env` 不应提交 Git。
- 云服务器上线必须使用真实域名、TLS 和受限 SSH 来源；不要照抄示例占位符。
