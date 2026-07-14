# 04 APT 与 Go、MySQL、Redis 环境

这一章完成短链项目最基本的服务器环境：

- 用 Ubuntu 24.04 的 APT 正确管理软件。
- 安装并验证 Go。
- 安装 MySQL，区分迁移账户和运行账户。
- 安装 Redis，使用本机绑定和 ACL 最小权限。
- 把密钥从代码、Git 和 systemd unit 正文中移走。

目标不是“把软件装上就算结束”，而是每安装一个组件，都能回答：

1. 它从哪个可信来源安装？
2. 它由哪个服务管理？
3. 它监听什么地址和端口？
4. 应用使用哪个账户、拥有哪些权限？
5. 配置和数据在哪里？
6. 出错时看哪条命令和哪份日志？

---

## 1. Ubuntu 软件管理的几层概念

### 1.1 dpkg 与 APT

dpkg 负责本机的 Debian 软件包：

~~~bash
dpkg -l
dpkg -L redis-server
dpkg -S /usr/bin/redis-server
~~~

APT 在 dpkg 之上处理：

- 软件源元数据。
- 依赖关系。
- 版本选择。
- 下载、安装、升级和卸载。
- 软件源签名验证。

日常优先使用 apt，不要到陌生网站下载一个 deb 后直接安装。

### 1.2 更新索引不等于升级软件

~~~bash
sudo apt update
apt list --upgradable
sudo apt upgrade
~~~

- apt update 只刷新“仓库里有哪些版本”的索引。
- apt upgrade 才会安装可升级的软件包。
- apt full-upgrade 允许为解决依赖而安装或删除包，执行前必须审阅计划。

生产服务器不要把大版本升级当成无风险例行命令。数据库、内核和 OpenSSH 升级都应有维护窗口、备份和回退方案。

### 1.3 查询而不是猜包名

~~~bash
apt search '^redis-server$'
apt show redis-server
apt policy redis-server
apt-cache depends redis-server
~~~

apt policy 可以看：

- 已安装版本。
- 候选版本。
- 候选版本来自哪个仓库。

### 1.4 安装、卸载和清理

~~~bash
sudo apt install redis-server
sudo apt remove redis-server
sudo apt purge redis-server
sudo apt autoremove
~~~

- remove 通常保留系统级配置。
- purge 同时移除由包管理器管理的配置。
- autoremove 会删除“APT 认为不再需要”的自动依赖，确认列表后再执行。

卸载软件包通常不会自动删除业务数据目录。对 MySQL、Redis 等数据服务，先确认备份和数据路径，绝不能把 purge 当作普通重装按钮。

---

## 2. Ubuntu 24.04 的 deb822 软件源

### 2.1 先查看现状

Ubuntu 24.04 常用 deb822 格式，主配置通常是：

~~~bash
cat /etc/apt/sources.list.d/ubuntu.sources
ls -l /etc/apt/sources.list /etc/apt/sources.list.d/
~~~

一个常见的 ubuntu.sources 片段：

~~~text
Types: deb
URIs: http://archive.ubuntu.com/ubuntu/
Suites: noble noble-updates noble-backports
Components: main restricted universe multiverse
Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg

Types: deb
URIs: http://security.ubuntu.com/ubuntu/
Suites: noble-security
Components: main restricted universe multiverse
Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg
~~~

不同地区镜像、云镜像和安装方式的 URI 可能不同。不要因为教程里的内容不同，就整份覆盖自己的文件。

deb822 关键字段：

- Types：通常是 deb，若需要源码包可写 deb-src。
- URIs：仓库地址。
- Suites：noble、noble-updates、noble-security 等发行套件。
- Components：main、universe 等组件。
- Signed-By：只信任指定 keyring 对这个仓库签名。

### 2.2 修改前备份与验证

如果确实要换镜像：

~~~bash
sudo cp -a \
  /etc/apt/sources.list.d/ubuntu.sources \
  /etc/apt/sources.list.d/ubuntu.sources.backup

sudoedit /etc/apt/sources.list.d/ubuntu.sources
sudo apt update
~~~

验证点：

- noble、noble-updates、noble-security 没有漏。
- 架构和发行代号正确。
- apt update 没有签名错误。
- 镜像支持 HTTPS 或处于可信网络；软件包签名仍然是必要校验。

### 2.3 第三方仓库与 Signed-By

现代 APT 不应再用 apt-key add 把第三方密钥放进全局信任库。更安全的模式是：

1. 从供应商官方 HTTPS 地址下载公钥。
2. 通过供应商文档或独立可信渠道核对指纹。
3. 放入 /etc/apt/keyrings/供应商.gpg。
4. 在该仓库条目中使用 Signed-By 精确引用。

示意流程：

~~~bash
sudo install -d -m 0755 /etc/apt/keyrings

curl -fLo /tmp/vendor.asc \
  https://vendor.example/repository-signing-key.asc

gpg --show-keys --with-fingerprint /tmp/vendor.asc

gpg --dearmor \
  --output /tmp/vendor.gpg \
  /tmp/vendor.asc

sudo install -m 0644 \
  /tmp/vendor.gpg \
  /etc/apt/keyrings/vendor.gpg
~~~

随后在 deb822 文件中使用：

~~~text
Signed-By: /etc/apt/keyrings/vendor.gpg
~~~

vendor.example 只是格式示例，不能照抄执行。添加第三方源意味着信任对方发布的软件包；能用 Ubuntu 官方仓库或软件官方发布包解决时，不要无意义堆仓库。

### 2.4 基础工具

~~~bash
sudo apt update
sudo apt install -y \
  build-essential \
  ca-certificates \
  curl \
  git \
  gnupg \
  jq \
  openssl \
  unzip
~~~

- ca-certificates：验证 HTTPS 证书。
- build-essential：gcc、make 等基础编译工具。
- jq：查看 JSON。
- openssl：生成随机值、检查证书等。

安装命令中的 -y 适合明确且可复现的包列表。对会删除或大规模替换依赖的操作，不要为了省一步确认而盲目使用 -y。

---

## 3. 安装 Go

### 3.1 两种合理方式

方式 A：Ubuntu 仓库

~~~bash
sudo apt install golang-go
go version
apt policy golang-go
~~~

优点是升级和卸载统一由 APT 管理；缺点是版本节奏由 Ubuntu 仓库决定，未必满足项目指定版本。

方式 B：Go 官方归档

适合项目需要明确 Go 版本。必须：

- 只从 go.dev 官方下载。
- 校验 SHA-256。
- 用版本目录保存，避免直接覆盖造成残留文件。
- 明确升级和回滚方式。

短链项目建议选择一种方式并记录版本，不要让 APT 版和手工版在 PATH 中互相抢优先级。

### 3.2 安装官方归档

先到 https://go.dev/dl/ 确认项目要用的版本。下面的版本号只是写法示例，执行前替换为项目已确认的版本：

~~~bash
GO_VERSION=1.26.5

case "$(dpkg --print-architecture)" in
  amd64) GO_ARCH=amd64 ;;
  arm64) GO_ARCH=arm64 ;;
  *)
    echo "unsupported architecture: $(dpkg --print-architecture)" >&2
    exit 1
    ;;
esac

FILE="go$GO_VERSION.linux-$GO_ARCH.tar.gz"
WORK_DIR="$(mktemp -d)"
cd "$WORK_DIR"

curl -fLO "https://go.dev/dl/$FILE"
curl -fLO "https://go.dev/dl/$FILE.sha256"

echo "$(cat "$FILE.sha256")  $FILE" | sha256sum -c -
~~~

预期：

~~~text
go1.26.5.linux-amd64.tar.gz: OK
~~~

只有校验成功才继续：

~~~bash
if [ -e "/opt/go/$GO_VERSION" ]; then
  echo "target version already exists; inspect it instead of overwriting" >&2
  exit 1
fi

sudo install -d -m 0755 "/opt/go/$GO_VERSION"
sudo tar \
  -xzf "$FILE" \
  -C "/opt/go/$GO_VERSION" \
  --strip-components=1
~~~

检查 /usr/local/go：

~~~bash
ls -ld /usr/local/go 2>/dev/null || true
~~~

如果它不存在或本来就是你管理的版本符号链接：

~~~bash
sudo ln -sfn "/opt/go/$GO_VERSION" /usr/local/go
~~~

如果 /usr/local/go 是一个真实目录，不要让 ln 在目录内部创建意外链接。先用 type -a go、dpkg -S 和 ls 判断旧安装来源，再制定迁移方案。

写入全局 PATH：

~~~bash
sudo tee /etc/profile.d/go.sh >/dev/null <<'EOF'
export PATH=/usr/local/go/bin:$PATH
EOF

source /etc/profile.d/go.sh
go version
type -a go
go env GOROOT GOPATH GOMODCACHE GOCACHE
~~~

这里使用带引号的 EOF，故意不在创建文件时展开 $PATH；登录 shell 加载文件时才展开。

### 3.3 不要乱设 GOPATH

现代 Go 项目使用 modules。通常不需要：

- 把项目放进 GOPATH/src。
- 把 GOPATH 设置成项目目录。
- 关闭 modules。

初始化项目：

~~~bash
mkdir -p ~/projects/shortlink
cd ~/projects/shortlink
go mod init example.com/shortlink
~~~

依赖版本写入 go.mod，校验写入 go.sum。

### 3.4 模块代理与校验

查看当前设置：

~~~bash
go env GOPROXY GOSUMDB GOPRIVATE
~~~

默认公共模块校验机制有助于发现内容被替换。不要为了“下载快”随手设置：

~~~text
GOSUMDB=off
GOINSECURE=*
~~~

如果网络需要替代代理，应了解代理运营方的信任和隐私边界，并尽量保留 sum.golang.org 校验。公司私有模块使用 GOPRIVATE 精确声明域名：

~~~bash
go env -w GOPRIVATE='git.example.com/company/*'
~~~

不要把所有模块都标成 private，否则会失去公共模块代理和校验带来的收益。

### 3.5 最小 Go 验证

~~~bash
mkdir -p ~/labs/go-check
cd ~/labs/go-check
go mod init go-check

cat > main.go <<'EOF'
package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Printf("go=%s os=%s arch=%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
EOF

go fmt ./...
go test ./...
go run .
go build -o go-check .
file go-check
~~~

注意 heredoc 使用带引号的 EOF，Go 源码中的符号不会被 shell 展开。

---

## 4. 安装并约束 MySQL

### 4.1 安装与服务状态

Ubuntu 24.04 官方仓库：

~~~bash
sudo apt update
sudo apt install -y mysql-server

mysql --version
systemctl status mysql --no-pager
sudo ss -lntp 'sport = :3306'
~~~

如果没有启动：

~~~bash
sudo systemctl enable --now mysql
journalctl -u mysql -n 100 --no-pager
~~~

先确认 Ubuntu 官方包是否满足项目版本。只有确实需要另一个版本时，才考虑 MySQL 官方仓库，并严格验证仓库密钥、发行版兼容性和升级路径。

### 4.2 确认监听地址

~~~bash
sudo grep -R --line-number \
  --include='*.cnf' \
  '^[[:space:]]*bind-address' \
  /etc/mysql

sudo ss -lntp 'sport = :3306'
~~~

同机部署时应监听：

~~~text
bind-address = 127.0.0.1
~~~

常见配置文件是 /etc/mysql/mysql.conf.d/mysqld.cnf，但应先检查 include 关系和现有配置，不要凭路径猜。

修改后：

~~~bash
sudo mysqld --validate-config
sudo systemctl restart mysql
sudo ss -lntp 'sport = :3306'
~~~

某些发行包或版本可能不支持同样的验证参数；若命令报“未知选项”，不要跳过检查，至少查看日志并在重启前保留现有 SSH 会话和配置备份。

绝不要为了从本机连接方便，把 MySQL 改成 0.0.0.0 并对公网开放 3306。

### 4.3 root 登录方式

Ubuntu 包常让本机系统 root 通过 Unix socket 管理 MySQL：

~~~bash
sudo mysql
~~~

这和使用 MySQL 密码从网络登录不是同一件事。应用绝不能使用数据库 root 账户。

### 4.4 创建数据库和最小权限账户

先生成并安全保存两个不同的密码：

~~~bash
openssl rand -hex 32
openssl rand -hex 32
~~~

进入 MySQL：

~~~bash
sudo mysql
~~~

执行，替换占位密码：

~~~sql
CREATE DATABASE shortlink
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_0900_ai_ci;

CREATE USER 'shortlink_migrator'@'127.0.0.1'
  IDENTIFIED BY 'replace-with-a-long-random-migrator-password';

GRANT SELECT, INSERT, UPDATE, DELETE,
      CREATE, ALTER, INDEX, DROP, REFERENCES
  ON shortlink.*
  TO 'shortlink_migrator'@'127.0.0.1';

CREATE USER 'shortlink_app'@'127.0.0.1'
  IDENTIFIED BY 'replace-with-a-different-long-random-app-password';

GRANT SELECT, INSERT, UPDATE, DELETE
  ON shortlink.*
  TO 'shortlink_app'@'127.0.0.1';

SHOW GRANTS FOR 'shortlink_migrator'@'127.0.0.1';
SHOW GRANTS FOR 'shortlink_app'@'127.0.0.1';
~~~

为什么分两个账户：

- migrator 只在发布迁移阶段使用，可以改表结构。
- app 是长期运行账户，不能随意 DROP 或 ALTER。

不要授予：

~~~sql
GRANT ALL ON *.* ...
~~~

执行 CREATE USER 和 GRANT 后不需要额外 FLUSH PRIVILEGES；这些 SQL 会立即更新权限系统。FLUSH PRIVILEGES 主要用于直接修改权限表等特殊场景，而直接改系统权限表本身也不推荐。

### 4.5 127.0.0.1 与 localhost 的账户区别

MySQL 账户由“用户名 + 来源主机”组成：

~~~text
'shortlink_app'@'127.0.0.1'
'shortlink_app'@'localhost'
'shortlink_app'@'%'
~~~

它们不是同一个账户。命令行中的 localhost 通常优先使用 Unix socket，而 127.0.0.1 明确走 TCP：

~~~bash
mysql -h 127.0.0.1 -u shortlink_app -p shortlink
~~~

Go 的 DSN 使用 tcp(127.0.0.1:3306) 时，应和 @127.0.0.1 账户匹配。不要用 @% 逃避主机匹配问题，@% 会扩大允许来源。

### 4.6 建立短链表

使用 migrator 账户：

~~~bash
mysql -h 127.0.0.1 -u shortlink_migrator -p shortlink
~~~

示例表：

~~~sql
CREATE TABLE short_links (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(16)
    CHARACTER SET ascii
    COLLATE ascii_bin
    NOT NULL,
  original_url TEXT NOT NULL,
  created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  expires_at TIMESTAMP(6) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_short_links_code (code),
  KEY idx_short_links_expires_at (expires_at)
) ENGINE=InnoDB;
~~~

code 使用大小写敏感的 ascii_bin，避免 aB3 和 Ab3 被错误视为相同。真实项目的字段长度、删除策略和索引要根据需求确定。

用运行账户验证：

~~~sql
SELECT CURRENT_USER();
SHOW GRANTS;
SELECT COUNT(*) FROM short_links;
CREATE TABLE should_fail (id INT);
~~~

前三条应成功，CREATE TABLE 应因权限不足失败。这正是最小权限生效，而不是环境坏了。

### 4.7 MySQL 常用检查

~~~bash
mysqladmin -h 127.0.0.1 -u shortlink_app -p ping
sudo mysql -e "SHOW PROCESSLIST"
sudo mysql -e "SHOW VARIABLES LIKE 'bind_address'"
journalctl -u mysql -n 100 --no-pager
~~~

不要在共享命令中使用 -p明文密码，因为它会进入 shell 历史和进程参数。

---

## 5. 安装并约束 Redis

### 5.1 安装与监听

~~~bash
sudo apt install -y redis-server
redis-server --version
systemctl status redis-server --no-pager
sudo ss -lntp 'sport = :6379'
~~~

常见配置文件：

~~~bash
sudoedit /etc/redis/redis.conf
~~~

同机部署至少确认：

~~~text
bind 127.0.0.1 -::1
protected-mode yes
port 6379
~~~

- 127.0.0.1 只允许本机 IPv4。
- -::1 中的减号表示没有该地址时不因此启动失败。
- protected-mode 是额外安全保护，不替代 bind 和 ACL。

验证配置中实际生效的相关行：

~~~bash
sudo grep -E \
  '^[[:space:]]*(bind|protected-mode|port|aclfile)[[:space:]]' \
  /etc/redis/redis.conf

sudo systemctl restart redis-server
sudo ss -lntp 'sport = :6379'
redis-cli PING
~~~

在配置 ACL 之前，本机默认用户可能无需密码。因为服务仅绑定回环，这可以作为很短的初始化窗口，但不应当作最终生产状态。

### 5.2 使用 ACL，而不是让应用成为 Redis 管理员

Redis 6 及以上支持 ACL。我们建立：

- shortlink_admin：只供管理员维护 ACL。
- shortlink_app：只能访问 short:* 键，并只允许项目当前需要的命令。
- default：初始化完成后关闭。

先在 redis.conf 中增加或确认：

~~~text
aclfile /etc/redis/users.acl
~~~

首次创建 ACL 文件前先确认它不存在。已有文件代表机器可能已经有 ACL，必须先审阅和合并，不能覆盖：

~~~bash
sudo ls -l /etc/redis/users.acl 2>/dev/null || true

sudo sh -c '
  set -C
  umask 027
  printf "%s\n" "user default on nopass ~* &* +@all" \
    > /etc/redis/users.acl
'

sudo chown redis:redis /etc/redis/users.acl
sudo chmod 0640 /etc/redis/users.acl
sudo systemctl restart redis-server
redis-cli PING
~~~

set -C 开启 noclobber；如果文件已经存在，重定向会拒绝覆盖。这里的 default 全权限只用于本机回环地址上的短暂初始化，完成下面步骤后会关闭。

生成两个不同密码，并先存入你的受控密码管理位置：

~~~bash
REDIS_ADMIN_PASSWORD="$(openssl rand -hex 32)"
REDIS_APP_PASSWORD="$(openssl rand -hex 32)"

printf 'admin=%s\napp=%s\n' \
  "$REDIS_ADMIN_PASSWORD" \
  "$REDIS_APP_PASSWORD"
~~~

终端输出也可能被录屏或记录。真实生产初始化应使用组织的 secret 管理流程。

创建用户。双引号用于防止 > 被 shell 当作重定向：

~~~bash
redis-cli ACL SETUSER shortlink_admin \
  reset on ">$REDIS_ADMIN_PASSWORD" \
  '~*' '&*' '+@all'

redis-cli ACL SETUSER shortlink_app \
  reset on ">$REDIS_APP_PASSWORD" \
  '~short:*' '&short:*' \
  '+ping' \
  '+get' '+mget' \
  '+set' '+setex' '+psetex' \
  '+del' '+unlink' \
  '+exists' \
  '+expire' '+pexpire' \
  '+ttl' '+pttl' \
  '+incr' '+decr'

redis-cli ACL SETUSER default off

REDISCLI_AUTH="$REDIS_ADMIN_PASSWORD" \
  redis-cli --user shortlink_admin ACL SAVE
~~~

ACL SAVE 会把当前规则写入 aclfile，密码以哈希形式保存。必须确认命令返回 OK；如果失败，不要退出当前管理会话，先修复 aclfile 路径和权限。确认：

~~~bash
REDISCLI_AUTH="$REDIS_ADMIN_PASSWORD" \
  redis-cli --user shortlink_admin ACL LIST

sudo ls -l /etc/redis/users.acl
~~~

不要把 shortlink_admin 的凭据交给应用。

### 5.3 验证应用权限

~~~bash
REDISCLI_AUTH="$REDIS_APP_PASSWORD" \
  redis-cli --user shortlink_app PING

REDISCLI_AUTH="$REDIS_APP_PASSWORD" \
  redis-cli --user shortlink_app \
  SET short:test ok EX 60

REDISCLI_AUTH="$REDIS_APP_PASSWORD" \
  redis-cli --user shortlink_app \
  GET short:test

REDISCLI_AUTH="$REDIS_APP_PASSWORD" \
  redis-cli --user shortlink_app \
  CONFIG GET bind
~~~

预期：

- PING、SET 和 GET 成功。
- CONFIG GET 因 NOPERM 失败。
- 操作 other:test 也应因键模式不匹配失败。

如果之后项目使用 Lua、Sorted Set、Hash 或发布订阅，应根据实际访问增加精确命令和 key/channel pattern。不要为了省事改成 +@all。

### 5.4 密码参数的边界

以下写法会把密码直接放在命令参数里：

~~~text
redis-cli -a 明文密码
~~~

redis-cli 会警告这可能不安全。REDISCLI_AUTH 可避免出现在命令文本和部分进程参数中，但环境变量仍不是绝对秘密，同用户或 root 仍可能观察。自动化应从权限严格的 secret 文件或密钥系统短暂注入，并控制日志。

### 5.5 不把 Redis 暴露到公网

不要设置：

~~~text
bind 0.0.0.0
protected-mode no
~~~

也不要在云安全组或 UFW 中对全网开放 6379。远程维护优先使用私网、VPN 或 SSH 本地隧道。ACL 是纵深防御，不是公开 Redis 的理由。

---

## 6. 应用配置与密钥

### 6.1 创建服务账户和配置目录

~~~bash
sudo adduser \
  --system \
  --group \
  --home /var/lib/shortlink \
  shortlink

sudo install -d \
  -o root \
  -g shortlink \
  -m 0750 \
  /etc/shortlink

sudo touch /etc/shortlink/shortlink.env
sudo chown root:shortlink /etc/shortlink/shortlink.env
sudo chmod 0640 /etc/shortlink/shortlink.env
~~~

touch 不会清空已经存在的配置；不要用会覆盖目标的安装命令重复初始化秘密文件。

用 sudoedit 编辑，不要先写进自己的临时文件再忘记清理：

~~~bash
sudoedit /etc/shortlink/shortlink.env
~~~

示例结构：

~~~text
APP_ADDR=127.0.0.1:8080

DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=shortlink
DB_USER=shortlink_app
DB_PASSWORD=replace-with-real-secret

REDIS_ADDR=127.0.0.1:6379
REDIS_USERNAME=shortlink_app
REDIS_PASSWORD=replace-with-real-secret
~~~

为什么分字段而不直接拼 DSN：

- 密码中的 @、:、/ 等字符可能需要编码。
- 程序可以用驱动提供的配置结构安全生成 DSN。
- 日志脱敏更容易。

### 6.2 不要把密码写进这些地方

- Git 仓库中的 .env。
- Go 源码和测试快照。
- systemd unit 的 Environment=明文。
- Dockerfile。
- Shell 脚本。
- README、命令截图和聊天记录。
- 数据库连接失败时的完整 DSN 日志。

EnvironmentFile 也不是密钥保险箱，但配合 root:服务组 和 0640，至少能把普通本机用户挡在外面。更高要求应使用云 Secret Manager、Vault、systemd credentials 等机制。

### 6.3 权限验证

~~~bash
sudo -u shortlink test -r /etc/shortlink/shortlink.env
sudo -u shortlink test ! -w /etc/shortlink/shortlink.env
namei -l /etc/shortlink/shortlink.env
~~~

预期：服务账户可读但不可写，其他普通用户不可读。

---

## 7. 实验：Go 同时连接 MySQL 与 Redis

### 7.1 准备项目

~~~bash
mkdir -p ~/labs/storage-check
cd ~/labs/storage-check
go mod init storage-check
go get github.com/go-sql-driver/mysql
go get github.com/redis/go-redis/v9
~~~

main.go：

~~~go
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

func mustEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("missing environment variable %s", name)
	}
	return value
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mysqlConfig := mysql.NewConfig()
	mysqlConfig.User = mustEnv("DB_USER")
	mysqlConfig.Passwd = mustEnv("DB_PASSWORD")
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = mustEnv("DB_ADDR")
	mysqlConfig.DBName = mustEnv("DB_NAME")
	mysqlConfig.ParseTime = true
	mysqlConfig.Collation = "utf8mb4_0900_ai_ci"

	db, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(3 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("mysql ping: %v", err)
	}

	var one int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil {
		log.Fatalf("mysql query: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     mustEnv("REDIS_ADDR"),
		Username: mustEnv("REDIS_USERNAME"),
		Password: mustEnv("REDIS_PASSWORD"),
	})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping: %v", err)
	}

	key := "short:env-check"
	if err := rdb.Set(ctx, key, "ok", 30*time.Second).Err(); err != nil {
		log.Fatalf("redis set: %v", err)
	}

	value, err := rdb.Get(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Fatalf("redis get: %v", err)
	}

	fmt.Printf("mysql=%d redis=%s\n", one, value)
}
~~~

### 7.2 临时注入配置

只用于个人实验终端：

~~~bash
export DB_ADDR='127.0.0.1:3306'
export DB_NAME='shortlink'
export DB_USER='shortlink_app'
read -rsp 'MySQL app password: ' DB_PASSWORD
export DB_PASSWORD
printf '\n'

export REDIS_ADDR='127.0.0.1:6379'
export REDIS_USERNAME='shortlink_app'
read -rsp 'Redis app password: ' REDIS_PASSWORD
export REDIS_PASSWORD
printf '\n'

go run .
~~~

预期：

~~~text
mysql=1 redis=ok
~~~

实验后：

~~~bash
unset DB_PASSWORD REDIS_PASSWORD
~~~

read -s 可以避免密码回显和写进 shell 历史，但它仍会进入当前进程环境。生产由服务管理器读取受保护配置，不依赖人工 export。

---

## 8. 故障排查

### 8.1 APT 报锁被占用

先确认是否真的有更新进程：

~~~bash
ps aux | grep -E '[a]pt|[d]pkg'
systemctl status apt-daily.service apt-daily-upgrade.service --no-pager
~~~

不要一看到锁文件就直接删除。另一个 dpkg 正在写数据库时删除锁，可能破坏包管理状态。

如果上次安装被意外中断，在确认没有其他包管理进程后：

~~~bash
sudo dpkg --configure -a
sudo apt -f install
~~~

### 8.2 APT 签名错误

检查：

- 系统时间是否正确。
- 仓库是否支持当前 noble。
- Signed-By 路径和文件权限。
- 下载的密钥指纹是否与官方一致。
- 是否错误混用了不同发行版的源。

不要用 trusted=yes 或关闭签名验证来“修复”。

### 8.3 go 版本与预期不符

~~~bash
type -a go
command -v go
readlink -f "$(command -v go)"
go env GOROOT
dpkg -S "$(command -v go)" 2>/dev/null || true
~~~

常见原因是 /usr/bin/go 和 /usr/local/go/bin/go 同时存在，PATH 顺序不同。

### 8.4 MySQL Connection refused

~~~bash
systemctl status mysql --no-pager
sudo ss -lntp 'sport = :3306'
mysql -h 127.0.0.1 -u shortlink_app -p shortlink
journalctl -u mysql -n 100 --no-pager
~~~

Connection refused 先看服务和监听，不是先重置密码。

### 8.5 MySQL Access denied

在管理员会话检查：

~~~sql
SELECT User, Host, plugin
FROM mysql.user
WHERE User LIKE 'shortlink%';

SHOW GRANTS FOR 'shortlink_app'@'127.0.0.1';
~~~

注意错误信息里的 user@host，它可能和你创建的账户不同。

### 8.6 Redis Connection refused、NOAUTH、NOPERM

~~~bash
systemctl status redis-server --no-pager
sudo ss -lntp 'sport = :6379'
journalctl -u redis-server -n 100 --no-pager
~~~

- Connection refused：服务或监听地址问题。
- NOAUTH：没有认证或凭据错误。
- WRONGPASS：用户名/密码错误或用户关闭。
- NOPERM：ACL 命令或 key pattern 不允许，先确认应用真实需求，不要直接授予全部权限。

管理员查看 ACL：

~~~bash
REDISCLI_AUTH="$REDIS_ADMIN_PASSWORD" \
  redis-cli --user shortlink_admin ACL GETUSER shortlink_app
~~~

### 8.7 本机能连，容器里不能连

容器内的 127.0.0.1 指向容器自己，不是宿主机。需要：

- 使用 Compose 网络中的服务名。
- 或使用明确的宿主机网关方案。
- 同时调整数据库监听和防火墙到最小必要范围。

不能简单把数据库改成 0.0.0.0 并公开端口。容器网络将在 Docker 章节单独处理。

---

## 9. 升级、备份与环境边界

### 9.1 Go 升级

用新的 /opt/go/版本目录安装并验证，再切换 /usr/local/go 符号链接。项目执行：

~~~bash
go version
go test ./...
go vet ./...
go build ./...
~~~

不要只因为编译通过就认定升级完成，还要看依赖和运行时行为变化。

### 9.2 MySQL 和 Redis 升级

升级前至少明确：

- 当前版本和目标版本。
- 官方升级说明。
- 数据备份是否可恢复。
- 是否涉及不可逆数据格式变化。
- 停机或滚动策略。
- 回退是否真的可行。

快照不是数据库一致性备份的自动替代品。后续部署章节会结合短链项目安排备份和恢复演练。

### 9.3 远程数据库

如果数据库不与 Go 同机：

- 优先放在同一私有网络。
- 安全组只允许应用服务器来源。
- 使用 TLS 并验证服务器证书。
- 数据库用户仍按来源和权限收紧。
- 管理入口走 VPN、堡垒机或临时 SSH 隧道。

“有密码”不等于“适合公网暴露”。

---

## 10. 本章验收

知识验收：

1. apt update 和 apt upgrade 的区别是什么？
2. Ubuntu 24.04 的 ubuntu.sources 为什么使用 Signed-By？
3. 为什么安装 Go 官方归档前必须校验 SHA-256？
4. 为什么不能让应用使用 MySQL root？
5. shortlink_migrator 和 shortlink_app 为什么要分开？
6. MySQL 的 @localhost 和 @127.0.0.1 有什么实际差别？
7. Redis 的 bind、protected-mode、ACL 各解决哪一层问题？
8. 为什么不能把 MySQL 3306 和 Redis 6379 直接公开到公网？

动手验收：

- 能用 apt policy 解释一个包从哪里安装。
- 能用 type -a go 判断实际执行的是哪个 Go。
- 能用 ss 证明 MySQL 和 Redis 只监听回环地址。
- 能证明 MySQL 运行账户可读写业务表，但不能建表。
- 能证明 Redis 应用账户可操作 short:*，但不能执行 CONFIG。
- 能运行 Go 存储检查程序，同时连通 MySQL 与 Redis。
- 能确保代码库里没有真实数据库和 Redis 密码。

完成这些再进入 Gin 的数据库接入，会比“先把 root 密码塞进代码跑起来”稳得多。
