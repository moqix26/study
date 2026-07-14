# Docker 与 Compose：把 Go 短链服务装进可重复运行的环境

这一章不把 Docker 当作“背命令”的独立知识点，而是用它解决一个具体问题：如何让同一份 Go/Gin 短链服务，在你的电脑、Linux 虚拟机和云服务器上以相同方式运行。

最终架构如下：

```text
互联网
   |
   | 80 / 443
   v
宿主机 Nginx
   |
   | 127.0.0.1:8080
   v
Go/Gin 容器 ---- Compose 私有网络 ---- MySQL 容器
     |
     +------------------------------- Redis 容器
```

这里有三条不能打折的边界：

- 公网只开放 SSH、HTTP 和 HTTPS；`8080` 只绑定宿主机回环地址。
- MySQL 和 Redis 不发布宿主机端口，只允许 Compose 网络内的应用访问。
- 容器里的 Go 进程不使用 root 身份运行，密码不写进镜像和 Git。

> 本章采用“宿主机 Nginx + 容器化应用和中间件”的方案。这样 TLS、证书续期和应用容器解耦，也更适合单机学习。全容器化并非错误，但不要在同一台机器上同时让两个 Nginx 抢占 80/443。

---

## 1. 先建立正确的容器心智模型

### 1.1 镜像、容器、卷和网络分别是什么

| 对象 | 可以把它理解为 | 是否应保存业务数据 |
|---|---|---|
| Image 镜像 | 只读的软件包和启动说明 | 否 |
| Container 容器 | 镜像启动后形成的进程及隔离环境 | 否，容器可随时重建 |
| Volume 卷 | 独立于容器生命周期的数据目录 | 是 |
| Network 网络 | 容器之间按服务名通信的虚拟网络 | 不涉及 |
| Registry 仓库 | 存放和分发镜像的服务 | 存镜像，不存运行数据 |

容器不是一台轻量虚拟机。它本质上仍是宿主机上的进程，只是借助 namespace、cgroup 等机制隔离了文件系统视图、网络和资源。进入容器执行 shell 只是排查手段，不应把容器当作长期手工维护的服务器。

如果你修复了一个容器内的文件，但没有修改 Dockerfile 并重新构建镜像，下次重建容器时修复就会消失。这正是容器“不可变部署”的核心：变更进入镜像，数据进入卷，配置从运行环境注入。

### 1.2 Dockerfile 和 Compose 各自负责什么

- `Dockerfile` 回答：Go 应用镜像如何构建、包含哪些文件、以什么用户启动。
- `compose.yaml` 回答：应用、MySQL、Redis 如何组合，使用什么网络、卷、配置和健康检查。
- Nginx、DNS、TLS 仍属于部署边界，见下一章。

### 1.3 容器内的 `localhost` 不是宿主机

在 `app` 容器中：

- `127.0.0.1` 指向 `app` 容器自身；
- MySQL 主机名应写 Compose 服务名 `mysql`，端口为容器端口 `3306`；
- Redis 主机名应写 `redis`，端口为 `6379`；
- 宿主机 Nginx 通过发布端口 `127.0.0.1:8080` 访问应用。

因此，容器内的数据库 DSN 不应写成 `127.0.0.1:3306`。

---

## 2. 安装与权限：能不用 sudo 不等于没有 root 权限

以下命令适用于 Ubuntu，仓库地址和发行版代号应以 Docker 官方安装文档为准：

```bash
sudo apt update
sudo apt install -y ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" \
  | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null

sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io \
  docker-buildx-plugin docker-compose-plugin docker-ce-rootless-extras

sudo systemctl enable --now docker
sudo docker version
sudo docker compose version
```

不要从不明脚本一键安装生产环境。安装后先确认客户端和服务端版本都能正常显示。

### 2.1 `docker` 用户组的真实权限

开发教程常见下面的做法：

```bash
sudo usermod -aG docker "$USER"
```

重新登录后，当前用户可以不加 `sudo` 执行 Docker。但这不是普通权限：能控制 Docker daemon 的用户通常可以挂载宿主机目录、启动特权容器，因此实际上接近宿主机 root。

合理选择如下：

- 个人学习虚拟机：加入 `docker` 组可以接受，但要清楚风险。
- 多用户服务器：不要把普通账号随意加入 `docker` 组。
- 安全要求更高的环境：考虑 Rootless Docker，或只让受控部署账号通过 CI/CD 操作。

尤其不要把 `/var/run/docker.sock` 挂进来源不可信的容器；这几乎等同于把宿主机交给该容器控制。

### 2.2 Rootless Docker

Rootless 模式让 daemon 和容器均由普通用户运行，可降低 daemon 被利用后的影响面：

```bash
sudo apt install -y uidmap dbus-user-session docker-ce-rootless-extras
dockerd-rootless-setuptool.sh install
systemctl --user enable --now docker
sudo loginctl enable-linger "$USER"
docker context use rootless
docker info | grep -i rootless
```

它也有现实限制，例如默认不能直接绑定小于 1024 的特权端口，部分网络和存储能力与 rootful 模式不同。本章本来就让 Go 应用使用 `8080`、让宿主机 Nginx 监听 `80/443`，因此较容易适配 Rootless 模式。

不要把“容器进程使用非 root 用户”和“Rootless Docker”混为一谈：

- `USER 10001` 只限制容器内应用进程；
- Rootless Docker 进一步让宿主机上的 daemon 也不以 root 运行；
- 两者可以同时采用，防护层次不同。

---

## 3. 为 Go/Gin 应用编写可交付镜像

假设仓库结构如下：

```text
shortlink/
├── cmd/server/main.go
├── internal/
├── migrations/
├── go.mod
├── go.sum
├── Dockerfile
└── .dockerignore
```

### 3.1 多阶段 Dockerfile

下面的例子使用 Alpine 作为运行层。构建镜像版本要与 `go.mod` 中声明的 Go 版本保持一致；示例版本仅作为写法演示。

```dockerfile
# syntax=docker/dockerfile:1.7

ARG GO_IMAGE=golang:1.26-alpine
FROM ${GO_IMAGE} AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w \
        -X main.version=${VERSION} \
        -X main.commit=${COMMIT} \
        -X main.buildTime=${BUILD_TIME}" \
      -o /out/shortlink-server ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 shortlink \
    && adduser -S -D -H -u 10001 -G shortlink shortlink

WORKDIR /app
COPY --from=build --chown=10001:10001 /out/shortlink-server /app/shortlink-server

USER 10001:10001
EXPOSE 8080

ENTRYPOINT ["/app/shortlink-server"]
```

构建：

```bash
docker build \
  --build-arg VERSION=v0.1.0 \
  --build-arg COMMIT="$(git rev-parse --short HEAD)" \
  --build-arg BUILD_TIME="$(date -u +%FT%TZ)" \
  -t shortlink:v0.1.0 .
```

这里的关键点不是“镜像越小越好”，而是：

- 编译器和源码只留在构建层，运行层只有二进制和运行所需证书。
- 使用固定 UID/GID 的非 root 用户，方便卷权限保持可预测。
- `CGO_ENABLED=0` 适用于纯 Go 依赖；如果项目使用必须依赖 C 库的驱动，应在兼容的构建和运行环境中编译，不能盲目关闭 CGO。
- `ca-certificates` 使程序能够验证外部 HTTPS 证书。
- `tzdata` 只在确实需要时保留。业务时间建议存 UTC，在展示层转换时区。
- 不在镜像中复制 `.env`、私钥、数据库备份或 Git 目录。

`scratch` 镜像更小，但没有 shell、CA 证书和常用排障工具；distroless 也有类似取舍。学习阶段使用精简 Alpine 更直观。生产环境真正重要的是及时更新基础镜像、扫描漏洞和固定版本，而不是为了几 MB 牺牲可维护性。

### 3.2 `.dockerignore`

```gitignore
.git
.gitignore
.env
.env.*
secrets/
backups/
tmp/
bin/
coverage.out
*.log
*.md
```

如果运行时需要迁移文件，不要把整个 `*.sql` 都忽略；应明确复制 `migrations/`，并确认其中没有真实数据。

### 3.3 检查镜像实际身份

```bash
docker run --rm --entrypoint id shortlink:v0.1.0
docker image inspect shortlink:v0.1.0 --format '{{.Config.User}}'
```

预期 UID/GID 为 `10001`，而不是 `0`。

---

## 4. 应用必须为容器运行做好准备

Docker 不能弥补应用自身缺失的运行能力。短链服务至少应具备：

1. `GET /healthz`：只判断进程和 HTTP 服务是否活着，不查询外部依赖。
2. `GET /readyz`：在很短的超时内检查 MySQL、Redis 等关键依赖，未就绪返回 `503`。
3. 收到 `SIGTERM` 后停止接收新请求，给在途请求一个有限的结束时间，然后退出码为 `0`。
4. 日志写标准输出/错误，不把日志固定写进容器文件系统。
5. 配置从环境变量或只读文件加载；密码支持 `*_FILE` 形式。
6. HTTP 客户端、数据库连接和 Redis 请求都有超时，不能无限等待。

Gin 的生产模式可通过非敏感配置传入：

```dotenv
APP_ENV=production
GIN_MODE=release
HTTP_ADDR=0.0.0.0:8080
DB_HOST=mysql
DB_PORT=3306
DB_NAME=shortlink
DB_USER=shortlink
REDIS_ADDR=redis:6379
```

容器内监听 `0.0.0.0:8080` 是合理的，因为进程需要接受容器网络接口的连接；真正控制宿主机暴露范围的是 Compose 中的 `127.0.0.1:8080:8080`。这与直接在宿主机运行 Go 时只监听 `127.0.0.1:8080` 并不矛盾。

### 4.1 `*_FILE` 的读取约定

环境变量易用，但敏感值会出现在进程环境、诊断输出或误打日志中。Compose secret 会把文件挂到 `/run/secrets/名称`，应用可以采用以下优先级：

```text
DB_PASSWORD_FILE 指向的文件内容
    > DB_PASSWORD 环境变量
    > 无默认值，启动失败
```

读取后要去掉文件末尾换行，不要把内容写入日志。配置校验只报告“缺少 DB_PASSWORD”，不要打印密码本身。

---

## 5. 一份有安全边界的 Compose 配置

目录建议：

```text
deploy/
├── compose.yaml
├── .env                     # 只有镜像标签等非敏感值
├── config/
│   └── app.env              # 非敏感运行配置，chmod 0640
└── secrets/                 # chmod 0700，整个目录加入 .gitignore
    ├── mysql_root_password.txt
    ├── mysql_app_password.txt
    └── redis_password.txt
```

`compose.yaml`：

```yaml
name: shortlink

services:
  app:
    image: "${APP_IMAGE:?set APP_IMAGE, for example shortlink:v0.1.0}"
    init: true
    user: "10001:10001"
    env_file:
      - ./config/app.env
    environment:
      DB_PASSWORD_FILE: /run/secrets/mysql_app_password
      REDIS_PASSWORD_FILE: /run/secrets/redis_password
    secrets:
      - mysql_app_password
      - redis_password
    ports:
      - "127.0.0.1:8080:8080"
    networks:
      - edge
      - data
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "-q", "-T", "2", "-O", "/dev/null", "http://127.0.0.1:8080/readyz"]
      interval: 10s
      timeout: 3s
      retries: 5
      start_period: 20s
    restart: unless-stopped
    read_only: true
    tmpfs:
      - /tmp:size=64m,noexec,nosuid,nodev
    cap_drop:
      - ALL
    security_opt:
      - no-new-privileges:true
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "5"

  mysql:
    image: mysql:8.4
    environment:
      MYSQL_DATABASE: shortlink
      MYSQL_USER: shortlink
      MYSQL_PASSWORD_FILE: /run/secrets/mysql_app_password
      MYSQL_ROOT_PASSWORD_FILE: /run/secrets/mysql_root_password
    secrets:
      - mysql_app_password
      - mysql_root_password
    volumes:
      - mysql_data:/var/lib/mysql
    networks:
      - data
    healthcheck:
      test:
        - CMD-SHELL
        - >-
          mysqladmin ping -h 127.0.0.1 -uroot
          -p"$$(cat /run/secrets/mysql_root_password)" --silent
      interval: 10s
      timeout: 5s
      retries: 10
      start_period: 30s
    restart: unless-stopped
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "5"

  redis:
    image: redis:7.4-alpine
    entrypoint: ["/bin/sh", "-c"]
    command:
      - >-
        exec redis-server
        --appendonly yes
        --requirepass "$$(cat /run/secrets/redis_password)"
    secrets:
      - redis_password
    volumes:
      - redis_data:/data
    networks:
      - data
    healthcheck:
      test:
        - CMD-SHELL
        - >-
          REDISCLI_AUTH="$$(cat /run/secrets/redis_password)"
          redis-cli ping | grep -q PONG
      interval: 10s
      timeout: 3s
      retries: 10
      start_period: 10s
    restart: unless-stopped
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "5"

volumes:
  mysql_data:
  redis_data:

networks:
  edge:
  data:
    internal: true

secrets:
  mysql_root_password:
    file: ./secrets/mysql_root_password.txt
  mysql_app_password:
    file: ./secrets/mysql_app_password.txt
  redis_password:
    file: ./secrets/redis_password.txt
```

创建密码文件时避免把密码留在 shell 历史中。可以使用交互式读取：

```bash
install -d -m 0700 secrets
read -rsp 'MySQL root password: ' MYSQL_ROOT_PASSWORD; echo
printf '%s' "$MYSQL_ROOT_PASSWORD" > secrets/mysql_root_password.txt
unset MYSQL_ROOT_PASSWORD
chmod 0600 secrets/mysql_root_password.txt
```

另外两个密码采用同样方式创建。确认 `secrets/` 已加入 `.gitignore`。

> Compose 的本地 `file:` secret 提供了“以文件挂载、避免直接放进环境变量”的使用方式，但它本身不是云端密钥保险箱，宿主机上的源文件仍需靠权限、磁盘加密和备份策略保护。更成熟的平台应接入专门的 secret manager。

### 5.1 为什么只发布应用的回环端口

```yaml
ports:
  - "127.0.0.1:8080:8080"
```

三个端口的含义依次是：宿主机地址、宿主机端口、容器端口。

- `127.0.0.1:8080:8080`：只有宿主机自身可访问，Nginx 可以反代。
- `8080:8080`：通常等价于在所有宿主机接口发布，可能绕过 Nginx 和 TLS。
- MySQL/Redis 没有 `ports`：只能通过 Compose 网络访问，宿主机和公网均无监听端口。

如果临时排查 MySQL，不要为方便长期添加 `3306:3306`，使用：

```bash
docker compose exec mysql mysql -ushortlink -p shortlink
docker compose exec redis sh
```

需要从宿主机 GUI 客户端临时连接时，也应绑定 `127.0.0.1` 并在使用后删除映射，不能开放到 `0.0.0.0`。

### 5.2 为什么分成 `edge` 和 `data` 网络

- `app` 同时加入 `edge` 与 `data`。
- MySQL、Redis 只加入设置了 `internal: true` 的 `data` 网络。
- 数据层没有直接访问外部网络的常规路径，也不会被其他无关 Compose 项目自动发现。

网络隔离不是数据库鉴权的替代品，因此仍设置独立应用账号和密码。应用账号只应拥有短链数据库需要的权限，不应使用 MySQL root。

### 5.3 版本标签必须可回滚

生产环境不要只用：

```yaml
image: your/shortlink:latest
```

`latest` 不能告诉你当前究竟运行哪次构建，也可能在不同机器上指向不同内容。应使用不可变版本，例如：

```dotenv
APP_IMAGE=ghcr.io/example/shortlink:v0.3.2
```

更严格时固定镜像 digest：

```text
ghcr.io/example/shortlink@sha256:真实摘要
```

版本标签便于人阅读，digest 保证内容不可变。发布记录应同时保存两者。

---

## 6. 启动、验收和日常操作

首次启动前先让 Compose 渲染配置：

```bash
docker compose config --quiet
docker compose config --services
```

注意：完整执行 `docker compose config` 可能把插值后的普通环境变量显示在终端或 CI 日志中，不要在公共日志里输出含敏感值的渲染结果。

启动与查看状态：

```bash
docker compose pull
docker compose up -d
docker compose ps
docker compose logs --tail=100 app
```

从宿主机验收：

```bash
curl --fail --show-error http://127.0.0.1:8080/healthz
curl --fail --show-error http://127.0.0.1:8080/readyz
ss -lntp | grep ':8080'
```

`ss` 应看到 `127.0.0.1:8080`，不应看到 `0.0.0.0:8080` 或 `[::]:8080`。

常用操作：

```bash
# 查看最近日志并持续跟踪
docker compose logs --tail=200 -f app

# 查看容器最终配置、网络和挂载
docker inspect shortlink-app-1

# 进入临时 shell；生产容器未必提供 shell
docker compose exec app sh

# 只重建应用，不动数据库卷
docker compose up -d --no-deps app

# 停止并删除容器、保留命名卷
docker compose down

# 同时删除数据卷：会丢数据，不能当普通清理命令使用
docker compose down --volumes
```

最后一个命令必须格外警惕。容器可重建，数据库卷不可随意删除。

---

## 7. Healthcheck、depends_on 和 restart 的真实语义

这三者最容易被误解。

### 7.1 Healthcheck 只负责报告健康状态

容器健康状态可能是：

- `starting`
- `healthy`
- `unhealthy`

查看细节：

```bash
docker inspect shortlink-app-1 \
  --format '{{json .State.Health}}'
```

Docker Engine 不会仅仅因为容器变成 `unhealthy` 就自动重启它。`restart` 策略主要针对主进程退出，而不是健康检查失败。需要基于不健康状态自动替换实例，通常由编排平台或额外监控处理。

因此不能故意让健康检查执行 `kill 1` 来“触发修复”。先查明依赖超时、连接池耗尽或死锁等根因。

### 7.2 `depends_on` 不是运行期依赖保证

`condition: service_healthy` 可以让 Compose 首次启动时等 MySQL/Redis 健康后再启动应用，但它不保证：

- MySQL 永远不会重启；
- 应用启动后依赖永远在线；
- 依赖恢复后业务连接一定自动恢复。

Go 应用仍需实现连接超时、有限重试、退避和连接重建。依赖暂时不可用时，readiness 应返回 `503`，而不是让请求无限挂起。

### 7.3 restart 策略如何选

| 策略 | 主进程退出后的行为 | 适用场景 |
|---|---|---|
| `no` | 不自动重启 | 一次性任务、迁移任务 |
| `on-failure` | 非零退出时重启 | 希望正常退出后保持停止 |
| `unless-stopped` | 通常持续重启，管理员手动停后保持停止 | 单机长期服务 |
| `always` | daemon 重启后也再次拉起，包括曾被手动停止的差异行为 | 需充分理解后使用 |

短链服务、MySQL、Redis 在单机 Compose 中可以使用 `unless-stopped`。数据库迁移不应配置无限重启，失败后应停下等待人工判断。

如果应用因为错误配置立刻退出，重启策略只会产生 crash loop，不会修复配置。此时查看：

```bash
docker compose ps -a
docker compose logs --tail=200 app
docker inspect shortlink-app-1 --format '{{.RestartCount}}'
```

---

## 8. 发布和回滚容器版本

假设当前为 `v0.3.1`，准备发布 `v0.3.2`：

1. 在 CI 构建一次镜像并推送，保存 digest。
2. 在服务器拉取新镜像，不立即删除旧镜像。
3. 修改 `.env` 中的 `APP_IMAGE` 为新版本。
4. 重建应用容器。
5. 通过回环地址和 Nginx 两条路径做健康检查。
6. 失败则把 `APP_IMAGE` 改回旧版本并再次重建。

```bash
docker compose pull app
docker compose up -d --no-deps app

for i in $(seq 1 30); do
  if curl -fsS --max-time 2 http://127.0.0.1:8080/readyz >/dev/null; then
    echo 'release healthy'
    break
  fi
  sleep 2
done
```

手工回滚：

```bash
# 把 .env 的 APP_IMAGE 恢复为上一版本后执行
docker compose up -d --no-deps app
curl -fsS http://127.0.0.1:8080/readyz
```

应用镜像回滚不等于数据库回滚。数据库迁移必须遵循向后兼容的 expand/contract 策略：

1. 先增加新列、新表或兼容索引；
2. 发布同时兼容新旧结构的程序；
3. 完成数据回填并验证；
4. 等旧程序不再可能回滚后，再删除旧列或旧约束。

如果新版本一上线就执行破坏性 `DROP COLUMN`，即使旧镜像还在，也可能已经无法回滚。

---

## 9. 数据卷、备份与恢复

查看卷：

```bash
docker volume ls
docker volume inspect shortlink_mysql_data
docker compose exec mysql mysql -ushortlink -p -e 'SELECT COUNT(*) FROM shortlink.links' shortlink
```

不要把“卷还在”当作备份。卷会因误操作、磁盘损坏或宿主机丢失而一起消失。MySQL 应做逻辑备份或物理备份，并复制到另一存储位置；恢复演练见第 09 章。

Redis 在短链项目中最好只作为缓存，让 MySQL 成为真实数据源。这样 Redis 丢失后可以重建缓存。如果 Redis 承担计数、队列或未落库数据，就必须按其持久性要求设计 AOF/RDB 和备份，不能口头称它为缓存却在里面保存唯一数据。

### 9.1 不要直接复制正在写入的 MySQL 卷

直接 `cp -r` 活跃数据库的数据目录可能得到不一致副本。正确做法包括：

- 小型项目使用 `mysqldump --single-transaction`；
- 数据量大时使用 MySQL 支持的物理备份工具和一致性快照；
- 无论使用哪种方式，都要定期恢复到隔离实例验证。

---

## 10. 资源限制与日志边界

单机上一个异常容器可能吃光内存或磁盘。可根据压测结果设置资源边界：

```yaml
services:
  app:
    cpus: 1.0
    mem_limit: 512m
```

不同 Compose/编排模式对资源字段的支持有差异，应用配置后要用 `docker inspect` 验证是否生效。资源值不能照抄：设置过低会让正常高峰变成 OOM，设置过高则失去隔离意义。

日志采用 `json-file` 时必须限制大小，否则 `/var/lib/docker` 可能填满根分区。本章示例设置了 `max-size` 和 `max-file`。查看占用：

```bash
docker system df
sudo du -xh /var/lib/docker --max-depth=1 | sort -h
```

不要在未确认对象范围时执行 `docker system prune -a --volumes`。它可能删除未使用镜像、构建缓存和卷；生产环境清理前应先列出对象并确认备份。

---

## 11. 常见故障的定位顺序

### 11.1 容器不断重启

```bash
docker compose ps -a
docker compose logs --tail=200 app
docker inspect shortlink-app-1 \
  --format 'exit={{.State.ExitCode}} error={{.State.Error}} restarts={{.RestartCount}}'
```

重点检查：缺少 secret、配置名拼错、二进制架构不匹配、端口占用、数据库迁移失败、只读文件系统下错误写入。

### 11.2 应用连不上 MySQL

```bash
docker compose exec app getent hosts mysql
docker compose exec mysql mysqladmin ping -h 127.0.0.1 -uroot -p
docker network inspect shortlink_data
```

依次确认 DNS 服务名、容器健康、网络成员、账号权限和 DSN。不要第一反应就把 3306 发布到公网。

### 11.3 宿主机 curl 不通

```bash
docker compose ps
ss -lntp | grep ':8080'
curl -v --max-time 3 http://127.0.0.1:8080/healthz
docker compose logs --tail=100 app
```

如果容器内应用只监听 `127.0.0.1:8080`，Docker 的端口转发无法连接它；容器内应监听 `0.0.0.0:8080`。如果宿主机显示 `0.0.0.0:8080`，则是 Compose 发布地址写得过宽。

### 11.4 容器健康检查失败，但手工请求成功

查看 healthcheck 的实际输出和执行用户：

```bash
docker inspect shortlink-app-1 --format '{{range .State.Health.Log}}{{.End}} exit={{.ExitCode}} {{.Output}}{{println}}{{end}}'
docker compose exec --user 10001:10001 app \
  wget -q -T 2 -O /dev/null http://127.0.0.1:8080/readyz
```

常见原因是镜像没有 `wget`、路径写错、启动宽限期不足，或 readiness 查询外部依赖耗时过长。

### 11.5 镜像能在本机运行，服务器报 `exec format error`

通常是 CPU 架构不一致，例如在 ARM 电脑构建了 ARM 镜像，却部署到 AMD64 服务器。检查：

```bash
uname -m
docker image inspect shortlink:v0.1.0 --format '{{.Architecture}}/{{.Os}}'
```

使用 Buildx 构建明确平台或多架构镜像，不能靠修改文件名解决。

---

## 12. 安全检查清单

部署前逐项检查：

- [ ] Go 进程使用固定非 root UID，镜像配置中 `User` 不为 `0`。
- [ ] 未把 Docker socket 挂入应用容器。
- [ ] 应用只发布 `127.0.0.1:8080`。
- [ ] MySQL、Redis 没有宿主机 `ports`。
- [ ] 应用使用最小权限数据库账号，不使用 root。
- [ ] secret 文件权限为 `0600`，目录权限为 `0700`，且不在 Git 中。
- [ ] 镜像标签可追溯，生产发布记录保存 digest。
- [ ] 容器根文件系统尽可能只读，移除了不需要的 Linux capabilities。
- [ ] healthcheck 有合理的 timeout、start period 和失败阈值。
- [ ] 明白 unhealthy 不会自动触发 Docker 重启。
- [ ] 日志有大小和数量上限。
- [ ] 数据卷有异机备份，并完成过恢复验证。
- [ ] 防火墙/安全组没有开放 8080、3306、6379。

---

## 13. 建议你亲手完成的实验

### 实验一：证明端口边界

1. 使用 `127.0.0.1:8080:8080` 启动应用。
2. 在服务器执行 `curl 127.0.0.1:8080/healthz`，应成功。
3. 从另一台机器访问 `服务器IP:8080`，应失败。
4. 用 `ss -lntp` 解释为什么。

### 实验二：观察依赖失效

1. 正常启动全部服务。
2. 执行 `docker compose stop redis`。
3. 观察 `/healthz` 与 `/readyz` 的差异。
4. 验证应用是否有超时，而不是请求永久挂起。
5. 启动 Redis，观察应用能否自行恢复。

### 实验三：做一次真实回滚

1. 保留一个可用镜像 `v0.1.0`。
2. 构建一个故意启动失败的 `v0.1.1-broken`。
3. 发布坏版本，查看退出码和重启次数。
4. 把镜像版本恢复到 `v0.1.0` 并重建应用。
5. 记录整个过程中 MySQL 卷为何没有被删除。

---

## 14. 学完本章应能回答

1. 为什么容器内的 `localhost` 不是宿主机或 MySQL 容器？
2. 为什么加入 `docker` 组几乎等于获得 root 能力？
3. 容器内非 root 与 Rootless Docker 有什么区别？
4. 为什么应用容器内监听 `0.0.0.0`，宿主机却只发布到 `127.0.0.1`？
5. 为什么 MySQL 和 Redis 通常不需要 `ports`？
6. healthcheck 失败后 Docker 是否一定重启容器？
7. `depends_on: condition: service_healthy` 为什么不能代替应用重试？
8. `latest` 为什么妨碍追踪和回滚？
9. Compose secret 解决了什么，又没有解决什么？
10. 为什么直接复制正在运行的 MySQL volume 不是可靠备份？

当你能亲手完成三个实验并解释这些问题时，Docker 才真正进入了你的后端能力，而不只是命令列表。
