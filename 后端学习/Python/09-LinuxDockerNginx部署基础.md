# Linux、Docker、Nginx 部署基础

> **文件编码**：UTF-8。Dockerfile、nginx.conf、shell 脚本建议 UTF-8。

<!-- 修改说明: 2026-06-30 按 EXPANSION-STANDARD 扩充 §0、FAQ≥12、闭卷自测、费曼检验 -->

## 0. 读前导读（零基础也能跟上）

### 0.1 用一句话弄懂本章

08 章你在本机开了五六个进程才能跑通 demo-api——**Docker 把「环境 + 代码 + 依赖」打包成镜像**，**docker-compose 一条命令拉起 MySQL/Redis/MQ/API/Worker**，**Nginx 做对外大门**，让别人通过 `http://你的域名` 访问，而不是 `127.0.0.1:8000`。

### 0.2 你需要提前知道什么（真不会就先跳到哪一章）

| 你已会 | 可以直接学本章 |
|--------|----------------|
| 04～08 章 demo-api 本地能跑 | ✅ 本章 |
| 会改 `.env`、会用 PowerShell 终端 | ✅ 本章 |
| 没跑通过 demo-api | 先完成 **[08 Celery](./08-Celery与消息队列实战.md)** 或至少 04～07 |
| Linux 完全零基础 | 先读 §2 命令表，边做边查；可配合 [Linux 01](../Linux/01-Linux入门与环境搭建.md) |

### 0.3 本章知识地图（学完后应能勾选全部 ☐→☑）

- [ ] 会用 10+ Linux 命令查日志、进程、端口
- [ ] 能写 demo-api 的 Dockerfile 并 `docker build`
- [ ] 能用 docker-compose 一键起 mysql + redis + rabbitmq + api + worker
- [ ] 知道 uvicorn 开发模式 vs gunicorn 生产模式的区别
- [ ] 能配置 Nginx 反代 `/api` 到 uvicorn :8000
- [ ] 能托管 Vue/React 静态 `dist` 并与后端同域
- [ ] 部署失败能按 §13 报错表逐步排查
- [ ] 能说明与 Java 09 部署链路的异同

### 0.4 建议学习时长与节奏

| 阶段 | 内容 | 建议时长 |
|------|------|----------|
| 第 1 天 | §1～§3 Linux 命令 + Docker 基础 | 2～3 小时 |
| 第 2 天 | §4～§6 Dockerfile + compose 手把手 | 3～4 小时 |
| 第 3 天 | §7～§9 Nginx + 生产启动 + 联调 | 2～3 小时 |
| 复盘 | FAQ + 闭卷 + 费曼 | 30 分钟 |

### 0.5 学完本章你能做什么（可验证的具体动作）

1. `docker compose up -d` 后，浏览器访问 Swagger 或 Nginx 反代后的 `/api/docs`
2. 重启 api 容器后 MySQL 里用户数据仍在（volume 持久化）
3. 用 `docker compose logs api` 定位 502 问题
4. 向同学演示「一条命令起全栈」并解释每个 service 的作用
5. 对照 Java 09 说出 Python 用 gunicorn+uvicorn 而非 `java -jar` 的原因

---

## 本章与上一章的关系

08 章你在本机同时开了 FastAPI、Celery Worker、MySQL、Redis、RabbitMQ——五个进程，换一台电脑就要重复配环境。线上服务器更不会让你 `uvicorn --reload` 挂着开发。

这一章解决 **「怎么部署」**：Linux 基础命令、Docker 镜像与 Compose、**uvicorn/gunicorn** 生产启动、**Nginx** 反向代理，让 demo-api 和中间件能一键拉起、对外提供稳定 HTTP 服务。09 章是 04～08 的「上线收官」。

**与 Java 路线对照**：部署流程与 [Java 09 Linux Docker Nginx 部署基础](../Java/09-LinuxDockerNginx部署基础.md) 平行；Java 用 `java -jar`，Python 用 **uvicorn / gunicorn**。

---

## 本章衔接

| 上一章（08） | 本章（09） | 下一章（10） |
|--------------|------------|--------------|
| Celery + RabbitMQ 本地跑通 | docker-compose 全栈一键启动 | 完整项目实战 + 联调 |
| 多终端手动起服务 | Nginx 反代 + 静态前端 | 登录、商品、下单串起来 |
| Windows 开发环境 | Linux 命令 + 生产部署清单 | 面试项目讲解 |

```mermaid
flowchart TB
    User[浏览器] --> Nginx[Nginx :80]
    Nginx -->|/api/*| API[uvicorn demo-api :8000]
    Nginx -->|/| Static[Vue 静态 dist]
    API --> MySQL[(MySQL)]
    API --> Redis[(Redis)]
    API --> MQ[RabbitMQ]
    Worker[Celery Worker] --> MQ
    Worker --> Redis
```

---

## 1. 为什么后端要学 Linux 和部署

后端不仅要「代码能跑」，还要：

- 在 Linux 服务器上查日志、查进程、查端口
- 用 Docker 统一环境，避免「我机器上能跑」
- 用 Nginx 做反向代理、负载均衡、托管前端静态文件

Python 后端部署常见链路：

```text
代码 → Docker 镜像 → Compose 编排 → Nginx 反代 → 用户访问
```

### 1.1 术语三件套：Docker 与容器

**Docker**：把应用及其依赖打包成**镜像**，在隔离的**容器**里运行。
**生活类比**：**标准化集装箱**——不管里面装的是 Python 还是 MySQL，码头（服务器）用同一套工具装卸。
**为什么重要**：消除「我机器能跑」；09 章起 demo-api 与中间件版本全员一致。
**本章用到的地方**：§4 Dockerfile、§6 docker-compose

### 1.2 术语三件套：Nginx 反向代理

**Nginx 反向代理**：对外接收浏览器请求，**转发**到内网 uvicorn；可同时托管静态 `dist`。
**生活类比**：**公司前台**——访客只找前台（80 端口），前台再内线转接各部门（api:8000）。
**为什么重要**：隐藏内网拓扑、TLS 终结、负载均衡、前后端同域。
**本章用到的地方**：§7～§8 nginx.conf

---

## 2. Linux 常用命令速查

部署排查离不开这些命令（云服务器多为 Linux；本地 Windows 可用 WSL 或 Git Bash 练习）：

```bash
# 文件与目录
ls -la
cd /var/log
tail -f app.log              # 实时看日志
grep "ERROR" app.log | tail -20

# 进程与端口
ps -ef | grep uvicorn
ss -tlnp | grep 8000         # 或 netstat -tlnp
kill -15 <pid>               # 优雅停止

# 磁盘与内存
free -h
df -h
du -sh logs/

# 权限
chmod +x start.sh
```

### Windows 本地对应（PowerShell）

```powershell
Get-Process | Where-Object {$_.ProcessName -like "*python*"}
netstat -ano | findstr :8000
Get-Content app.log -Wait -Tail 20
```

---

## 3. Docker 核心概念

| 概念 | 说明 |
|------|------|
| 镜像 Image | 只读模板（如 `python:3.12-slim`） |
| 容器 Container | 镜像的运行实例 |
| 卷 Volume | 持久化数据（MySQL 数据目录） |
| 网络 Network | 容器间互联（Compose 默认网） |

常用命令：

```powershell
docker ps                    # 运行中容器
docker logs study-mysql      # 看日志
docker exec -it study-redis redis-cli
docker stop study-mysql
docker rm study-mysql
```

---

## 4. uvicorn 生产部署

### 4.1 开发 vs 生产

| 模式 | 命令 | 说明 |
|------|------|------|
| 开发 | `uvicorn app.main:app --reload --port 8000` | 热重载，单进程 |
| 生产 | `gunicorn app.main:app -k uvicorn.workers.UvicornWorker -w 4 -b 0.0.0.0:8000` | 多 worker |

### 4.2 安装 gunicorn（Linux / WSL 生产环境）

```bash
pip install gunicorn uvicorn[standard]
```

```bash
gunicorn app.main:app \
  -k uvicorn.workers.UvicornWorker \
  -w 4 \
  -b 0.0.0.0:8000 \
  --access-logfile - \
  --error-logfile -
```

**预期输出**：

```text
[2025-06-18 10:00:00 +0800] [1] [INFO] Starting gunicorn 21.x
[2025-06-18 10:00:00 +0800] [1] [INFO] Listening at: http://0.0.0.0:8000
[2025-06-18 10:00:00 +0800] [1] [INFO] Using worker: uvicorn.workers.UvicornWorker
[2025-06-18 10:00:00 +0800] [8] [INFO] Booting worker with pid: 8
...
```

### 4.3 Windows 本地生产模拟

Windows 上 gunicorn 支持有限，本地可用：

```powershell
uvicorn app.main:app --host 0.0.0.0 --port 8000 --workers 1
```

真正多 worker 生产部署建议放 **Linux 容器** 内。

### 4.4 环境变量

```powershell
$env:DATABASE_URL="mysql+asyncmy://root:123456@127.0.0.1:3306/study_db"
$env:REDIS_URL="redis://127.0.0.1:6379/0"
$env:CELERY_BROKER_URL="amqp://guest:guest@127.0.0.1:5672//"
uvicorn app.main:app --host 0.0.0.0 --port 8000
```

---

## 5. Dockerfile 示例（demo-api）

```dockerfile
FROM python:3.12-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple

COPY . .

EXPOSE 8000

CMD ["gunicorn", "app.main:app", "-k", "uvicorn.workers.UvicornWorker", "-w", "2", "-b", "0.0.0.0:8000"]
```

**逐行读 Dockerfile**：

| 行/指令 | 含义 | 改错会怎样 |
|---------|------|------------|
| `FROM python:3.12-slim` | 基础镜像 | 版本与本地不一致可能依赖冲突 |
| `WORKDIR /app` | 容器内工作目录 | 错则 COPY 路径与 import 混乱 |
| `PYTHONUNBUFFERED=1` | 日志即时输出 | 不设则 docker logs 延迟 |
| `COPY requirements.txt` + `RUN pip` | 先装依赖利用缓存层 | 每次改代码都重装依赖变慢 |
| `COPY . .` | 拷贝项目源码 | `.dockerignore` 应排除 `.venv` |
| `EXPOSE 8000` | 文档声明端口 | 不映射 ports 仍无法外访 |
| `CMD gunicorn ... -w 2` | 2 个 uvicorn worker | `-w 1` 无法利用多核 |

构建与运行：

```powershell
cd f:\study\demo-api
docker build -t demo-api:1.0 .
docker run -d --name demo-api -p 8000:8000 `
  -e DATABASE_URL=mysql+asyncmy://root:123456@host.docker.internal:3306/study_db `
  -e REDIS_URL=redis://host.docker.internal:6379/0 `
  demo-api:1.0
```

```powershell
curl http://127.0.0.1:8000/docs
# 预期：200，OpenAPI 文档页
```

---

## 6. docker-compose 全栈部署（核心）

在 `demo-api/deploy/` 目录创建 `docker-compose.yml`：

```yaml
services:
  mysql:
    image: mysql:8.0
    container_name: study-mysql
    environment:
      MYSQL_ROOT_PASSWORD: "123456"
      MYSQL_DATABASE: study_db
    ports:
      - "3306:3306"
    volumes:
      - mysql-data:/var/lib/mysql
      - ./sql/init.sql:/docker-entrypoint-initdb.d/init.sql:ro
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-p123456"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7
    container_name: study-redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

  rabbitmq:
    image: rabbitmq:3-management
    container_name: study-rabbitmq
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest

  api:
    build:
      context: ..
      dockerfile: deploy/Dockerfile
    container_name: demo-api
    ports:
      - "8000:8000"
    environment:
      DATABASE_URL: mysql+asyncmy://root:123456@mysql:3306/study_db?charset=utf8mb4
      REDIS_URL: redis://redis:6379/0
      CELERY_BROKER_URL: amqp://guest:guest@rabbitmq:5672//
      CELERY_RESULT_BACKEND: redis://redis:6379/1
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
      rabbitmq:
        condition: service_started

  worker:
    build:
      context: ..
      dockerfile: deploy/Dockerfile
    container_name: demo-worker
    command: celery -A celery_app worker --loglevel=info --concurrency=2
    environment:
      DATABASE_URL: mysql+asyncmy://root:123456@mysql:3306/study_db?charset=utf8mb4
      REDIS_URL: redis://redis:6379/0
      CELERY_BROKER_URL: amqp://guest:guest@rabbitmq:5672//
      CELERY_RESULT_BACKEND: redis://redis:6379/1
    depends_on:
      - api
      - rabbitmq
      - redis

volumes:
  mysql-data:
  redis-data:
```

### 6.1 用法（PowerShell）

```powershell
cd f:\study\demo-api\deploy
docker compose up -d --build
```

**预期输出**：

```text
[+] Running 6/6
 ✔ Network deploy_default    Created
 ✔ Volume deploy_mysql-data  Created
 ✔ Container study-mysql     Started
 ✔ Container study-redis     Started
 ✔ Container study-rabbitmq  Started
 ✔ Container demo-api        Started
 ✔ Container demo-worker     Started
```

```powershell
docker compose ps
# 预期：mysql、redis、rabbitmq、demo-api、demo-worker 均为 running
```

```powershell
docker compose logs -f api
# 预期：gunicorn/uvicorn 启动日志，Listening at :8000
```

```powershell
curl http://127.0.0.1:8000/api/products/1
# 预期：JSON 响应
```

停止：

```powershell
docker compose down
# 保留数据卷
docker compose down -v
# 删除数据卷（慎用）
```

---

## 7. Nginx 反向代理

### 7.1 为什么需要 Nginx

- 统一 80/443 入口
- 把 `/api` 转发到 uvicorn
- 托管 Vue 打包后的 `dist` 静态文件
- 后续可加 HTTPS、限流、负载均衡

### 7.2 配置示例

`deploy/nginx/demo.conf`：

```nginx
server {
    listen 80;
    server_name localhost;

    # 前端静态资源（Vue 08 打包产物）
    location / {
        root /usr/share/nginx/html;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    # 后端 API 反代
    location /api/ {
        proxy_pass http://api:8000/;   # 注意末尾斜杠：/api/users -> /users
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Compose 追加 nginx 服务：

```yaml
  nginx:
    image: nginx:1.25-alpine
    container_name: demo-nginx
    ports:
      - "80:80"
    volumes:
      - ./nginx/demo.conf:/etc/nginx/conf.d/default.conf:ro
      - ../frontend/dist:/usr/share/nginx/html:ro
    depends_on:
      - api
```

### 7.3 验证

```powershell
curl http://localhost/api/products/1
# 预期：与直连 :8000 相同 JSON

curl -I http://localhost/
# 预期：200，Content-Type text/html
```

### 7.4 与前端联调

[Vue 08 Axios 联调](../../前端学习/Vue/08-Axios网络请求与前后端联调.md) 中，生产环境应：

- 前端 `baseURL` 设为 `/api`（同域，无 CORS 问题）
- 或 Nginx 同时托管前后端

```javascript
// axios 实例
const request = axios.create({
  baseURL: '/api',
  timeout: 10000,
})
```

---

## 8. 部署检查清单

- [ ] Python 版本与 Dockerfile 一致（3.11+ / 3.12）
- [ ] 环境变量：DATABASE_URL、REDIS_URL、CELERY_BROKER_URL
- [ ] MySQL 已执行 `init.sql`（06 章建表）
- [ ] Worker 容器在跑，`docker compose logs worker` 无报错
- [ ] 防火墙 / 安全组放行 80、8000（若直连调试）
- [ ] Nginx `proxy_pass` 端口与 api 服务一致
- [ ] 日志目录可写（挂载 volume 或 stdout）

---

## 9. 部署失败排错顺序

```mermaid
flowchart TD
    A[访问失败] --> B{服务启动?}
    B -->|否| C[docker compose logs api]
    B -->|是| D{端口监听?}
    D -->|否| E[检查 CMD / 端口映射]
    D -->|是| F{DB/Redis 可连?}
    F -->|否| G[检查服务名 mysql/redis 与 URL]
    F -->|是| H{Nginx 502?}
    H -->|是| I[proxy_pass 地址与 api 容器名]
    H -->|否| J[查应用日志业务错误]
```

1. 服务是否启动成功
2. 端口是否监听
3. 日志是否报错
4. MySQL / Redis / RabbitMQ 是否可连
5. Nginx 配置是否正确

---

## 10. 数据卷与持久化

为什么需要 volume：

- `docker compose down` 删容器，**不删 volume 则 MySQL 数据仍在**
- 生产必须把 `mysql-data`、`redis-data` 挂到 volume 或宿主机目录

```yaml
volumes:
  mysql-data:
    driver: local
```

---

## 11. 健康检查与依赖

Compose 中 `depends_on` + `healthcheck` 避免 api 在 MySQL 未就绪时启动失败：

```yaml
depends_on:
  mysql:
    condition: service_healthy
```

FastAPI 可选健康接口：

```python
@app.get("/health")
async def health():
    return {"status": "ok"}
```

Nginx 或 K8s 可探测 `/health`。

---

## 12. HTTPS 认知（扩展）

生产应用 Let's Encrypt + Nginx：

```nginx
listen 443 ssl;
ssl_certificate     /etc/nginx/ssl/fullchain.pem;
ssl_certificate_key /etc/nginx/ssl/privkey.pem;
```

初学先把 HTTP 跑通，HTTPS 在实际上线时配置。

---

## 13. 常见报错与排查

| 报错信息（关键词） | 可能原因 | 解决方案 |
|-------------------|---------|---------|
| `ModuleNotFoundError: app.main` | WORKDIR 或 PYTHONPATH 错 | Dockerfile `WORKDIR /app`；COPY 完整项目 |
| `Connection refused mysql:3306` | MySQL 未就绪或 URL 错 | 等 healthcheck；URL 用服务名 `mysql` 非 localhost |
| `Can't connect to Redis` | Redis 容器未起 | `docker compose ps`；URL 用 `redis:6379` |
| `502 Bad Gateway` | 后端未启动或 upstream 错 | `docker compose logs api`；检查 `proxy_pass` |
| `404 on /api/xxx` | 路径前缀重复或缺少 | 统一 FastAPI `root_path` 或 Nginx 斜杠规则 |
| `Address already in use :8000` | 端口被占 | `netstat` 查 PID；改 ports 映射 |
| `docker compose` 找不到 | Docker Desktop 未启 | 启动 Docker；旧版用 `docker-compose` |
| `no such file init.sql` | 挂载路径错 | 确认 `./sql/init.sql` 相对 compose 文件存在 |
| Worker 不消费 | Broker URL 在容器内仍写 localhost | 改为 `rabbitmq` 主机名 |
| `gunicorn: command not found` | 未安装 gunicorn | requirements.txt 加入 gunicorn |
| 静态页 404 | dist 未挂载或路径错 | 确认 Vue build 产物路径 |
| `CORS error` 生产仍出现 | 前后端不同域 | 同域 Nginx 反代或配置 CORSMiddleware |

---

## 14. 分级练习

### 基础

本地 `uvicorn` 启动 demo-api，`curl http://127.0.0.1:8000/docs` 看到 Swagger。

### 进阶

`docker compose up` 起 mysql + redis + api，POST 用户/商品数据，重启 api 容器数据仍在。

### 挑战

Nginx 托管 Vue `dist` + 反代 `/api`，浏览器完成 [Vue 08](../../前端学习/Vue/08-Axios网络请求与前后端联调.md) 列表页联调。

---

## 15. 参考答案

### 基础

```powershell
cd f:\study\demo-api
.\.venv\Scripts\Activate.ps1
uvicorn app.main:app --host 127.0.0.1 --port 8000
curl http://127.0.0.1:8000/docs
# 预期：HTML 页面含 Swagger UI
```

### 进阶

```powershell
cd f:\study\demo-api\deploy
docker compose up -d mysql redis api
curl -X POST http://127.0.0.1:8000/users -H "Content-Type: application/json" -d "{\"username\":\"test\"}"
docker compose restart api
curl http://127.0.0.1:8000/users
# 预期：用户仍在
```

### 挑战：最小静态页 + Nginx

`frontend/dist/index.html`：

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="UTF-8"><title>demo</title></head>
<body>
  <button id="btn">测接口</button>
  <pre id="out"></pre>
  <script>
    document.getElementById('btn').onclick = () =>
      fetch('/api/products/1').then(r => r.json()).then(d =>
        document.getElementById('out').textContent = JSON.stringify(d, null, 2));
  </script>
</body>
</html>
```

浏览器打开 `http://localhost/`，点击按钮，`<pre>` 显示 JSON 即成功。

---

## 16. 与 Java 部署差异速查

| 维度 | Java（09 章） | Python（本章） |
|------|---------------|----------------|
| 打包 | `mvn package` → jar | 源码 + requirements.txt |
| 运行 | `java -jar app.jar` | gunicorn + uvicorn worker |
| 默认端口 | 8080 | 8000 |
| 异步任务 | Spring + RabbitMQ 消费者 | Celery Worker 容器 |
| 连接池 | HikariCP | SQLAlchemy pool |

中间件（MySQL、Redis、RabbitMQ、Nginx）配置思路**完全一致**。

### 16.1 部署排错步骤表（手把手）

| 步骤 | 你的动作 | 预期看到什么 | 若不对 |
|------|----------|--------------|--------|
| 1 | `docker compose ps` | 各 service Up (healthy) | 看哪个 Exited |
| 2 | `docker compose logs api --tail=50` | 无 Traceback | 按栈修 DATABASE_URL 等 |
| 3 | `curl http://127.0.0.1:8000/health` 或 `/docs` | 200 | api 未监听或端口错 |
| 4 | 容器内 `ping mysql` / `nc -zv redis 6379` | 通 | 网络/服务名错 |
| 5 | 浏览器访问 Nginx `/` 与 `/api/docs` | 静态页 + Swagger | 查 proxy_pass、try_files |
| 6 | POST 下单 + `logs worker` | Celery succeeded | Broker URL 仍写 localhost |

### 16.2 交叉索引

| 主题 | 本章 | 延伸阅读 |
|------|------|----------|
| Celery Worker 容器 | §6 compose | [08 Worker](./08-Celery与消息队列实战.md) |
| demo-api 项目结构 | §4 | [10 项目](./10-后端项目实战与面试准备.md) |
| Vue 同域联调 | §8 Nginx | [Vue 08](../../前端学习/Vue/08-Axios网络请求与前后端联调.md) |
| Linux 命令深入 | §2 | [Linux 01～14](../Linux/00-学习路线图与说明.md) |
| Java 部署对照 | §16 表 | [Java 09](../Java/09-LinuxDockerNginx部署基础.md) |

---

## 17. 高频知识点清单

- Linux：tail、grep、ps、ss、df
- Docker：run、ps、logs、exec、volume
- Dockerfile：FROM、COPY、CMD
- docker-compose：services、depends_on、healthcheck
- uvicorn vs gunicorn
- Nginx：proxy_pass、静态 root、try_files
- 部署排错顺序
- 前后端同域反代

---

## 18. 学完标准

- [ ] 会用 10+ Linux 命令查日志、进程、端口（或在 WSL 练习）
- [ ] 能写 demo-api 的 Dockerfile 并 `docker build`
- [ ] 能用 compose 一键起 mysql + redis + rabbitmq + api + worker
- [ ] 能配置 Nginx 反代 `/api` 到 uvicorn
- [ ] 知道 gunicorn + UvicornWorker 生产启动方式
- [ ] 部署失败能按清单逐步排查
- [ ] 能说明与 [Java 09](../Java/09-LinuxDockerNginx部署基础.md) 的异同

---

## 19. FAQ

**Q1：开发用 uvicorn --reload，生产用什么？**  
生产用 **gunicorn 管理多个 uvicorn worker 进程**，或容器内 `uvicorn --workers 4`；不要长期 `--reload`（性能差、不安全）。

**Q2：容器里连 MySQL 为什么不能用 localhost？**  
`localhost` 指**容器自身**；compose 里应写**服务名** `mysql:3306`，Docker 内网 DNS 解析。

**Q3：502 Bad Gateway 最先查什么？**  
`docker compose ps` 看 api 是否 Up → `logs api` 看崩溃栈 → Nginx `proxy_pass` 端口是否对。

**Q4：Windows 必须学 Linux 吗？**  
后端部署环境以 Linux 为主；Windows 可用 **WSL2 + Docker Desktop** 练习本章命令。

**Q5：Docker 和虚拟机有什么区别？**  
Docker 共享宿主机内核、启动秒级、镜像分层；VM 完整虚拟 OS、更重。后端部署普遍用容器。

**Q6：Nginx 和 uvicorn 各干什么？**  
Nginx：**对外 80/443**、静态文件、反代、负载均衡；uvicorn：跑 **Python ASGI 应用**。

**Q7：`.env` 在 compose 里怎么生效？**  
`env_file: .env` 或 `environment:` 注入；**不要把密码提交 Git**，用 `.gitignore`。

**Q8：Worker 容器 Broker URL 写 localhost 会怎样？**  
连的是自己容器，连不上 RabbitMQ；应写 **`rabbitmq:5672`**（compose 服务名）。

**Q9：数据会随 `docker compose down` 丢失吗？**  
若用了 **named volume** 挂载 MySQL 数据，`-v` 删除 volume 才会丢；练习时注意备份。

**Q10：前后端同域有什么好处？**  
避免 CORS 复杂配置；Cookie/Token 策略更简单；生产常见 Nginx 托管 `dist` + 反代 `/api`。

**Q11：和 Java jar 部署最大区别？**  
Java 打包成 fat jar 一个文件；Python 通常 **源码 + requirements.txt + 镜像**，运行时装依赖或用多阶段构建。

**Q12：healthcheck 有什么用？**  
compose 等 MySQL **真正就绪**再启动 api，避免 `Connection refused` 启动竞态。

---

## 20. 闭卷自测

1. **概念**：Docker 镜像与容器的关系是什么？
2. **概念**：docker-compose 里 `depends_on` 解决什么问题？有什么局限？
3. **概念**：Nginx `proxy_pass` 的作用？
4. **概念**：为什么生产不建议长期 `uvicorn --reload`？
5. **概念**：volume 在 MySQL 容器场景解决什么问题？
6. **概念**：列出 compose 里 demo-api 至少 4 个 service 及其职责。
7. **动手**：写出查看 api 容器最后 100 行日志的命令。
8. **动手**：容器内执行 redis-cli PING 的 docker 命令格式？
9. **综合**：部署后前端能打开静态页但 `/api` 404，排查步骤？
10. **综合**：对照 Java 09，画 Python 部署链路（从代码到用户浏览器）。

### 自测参考答案

1. 镜像是只读模板；容器是镜像的运行实例，可写层隔离。
2. 控制启动顺序；**不保证**依赖服务已 ready，需 healthcheck。
3. 把客户端请求转发到后端 upstream（如 `http://api:8000`）。
4. reload 监视文件变化、单进程、不适合多 worker 与高并发。
5. 数据库文件持久化到宿主机/volume，`down` 不删 volume 则数据保留。
6. mysql 存数据、redis 缓存、rabbitmq 消息、api HTTP、worker 异步任务（+ nginx 可选）。
7. `docker compose logs --tail=100 api`
8. `docker exec -it study-redis redis-cli PING` 或 compose 服务名。
9. 查 Nginx location 前缀、FastAPI root_path、proxy_pass 末尾斜杠、api 路由是否带 `/api` 前缀。
10. 代码 → Dockerfile build → compose up → Nginx:80 → uvicorn:8000 → MySQL/Redis/MQ（开放题，能串即可）。

---

## 21. 费曼检验

请用 **3 分钟**解释：**「为什么学完 FastAPI 还要学 Docker 和 Nginx？」**

**对照提纲**：

1. **开发能跑 ≠ 别人能访问**：本机 8000 端口外网打不开。
2. **Docker**：统一 MySQL/Redis/MQ 版本，换电脑 `compose up` 即可。
3. **Nginx**：对外统一入口、托管前端静态文件、HTTPS、反代后端。
4. **Python 特有**：gunicorn+uvicorn 多 worker 扛并发。
5. **面试**：能画「浏览器 → Nginx → uvicorn → DB/Redis/MQ」。

---

## 下一章预告

04～09 的技术栈已全部就绪——下一章（10 后端项目实战与面试准备）把 **demo-api 扩展为完整练手项目**：

- 用户注册登录、JWT 鉴权
- 商品列表/详情（MySQL + Redis）
- 下单 + Celery 异步通知
- 与 [Vue 08](../../前端学习/Vue/08-Axios网络请求与前后端联调.md) 完整联调
- 简历项目描述与面试讲法

09 章解决「能部署」，10 章解决「能讲清楚、能面试」。

---

*下一章：10 后端项目实战与面试准备*

*本章已按 EXPANSION-STANDARD 扩充（§0 导读 + FAQ 12 + 闭卷 + 费曼）。*

**EXPANSION-STANDARD 自检**：☑ §0.1～0.5 ☑ FAQ≥12 §19 ☑ 闭卷 10 题 §20 ☑ 费曼 §21
