# 05 Shell 与可靠发布脚本

Shell 最适合做“把已有工具可靠地串起来”：

- 执行测试和编译。
- 校验构建产物。
- 创建发布目录。
- 切换版本。
- 重启服务。
- 做健康检查和回滚。

Shell 不适合承载复杂业务逻辑。短链生成、数据库事务、并发控制等仍应写在 Go 中。

这一章的目标不是堆语法，而是写出失败时行为可预测的自动化。

---

## 1. 脚本是另一种程序

一个可靠脚本也需要：

- 明确输入。
- 校验前置条件。
- 可理解的日志。
- 正确的退出码。
- 对临时资源做清理。
- 对重复执行和并发执行有定义。
- 失败后不留下“半新半旧”的状态。

常见危险想法：

- “就几条命令，不会出错。”
- “失败了手工补一下。”
- “加 set -e 就可靠了。”
- “发布失败就把旧目录复制回来。”

可靠性来自明确设计，不来自某一个神奇选项。

---

## 2. Bash、执行方式与 shebang

### 2.1 指定解释器

如果脚本使用 Bash 的数组、[[ ]]、pipefail 等特性：

~~~bash
#!/usr/bin/env bash
~~~

这表示从当前 PATH 查找 bash，适合开发脚本。系统管理脚本若要求固定系统 Bash，也可使用：

~~~bash
#!/bin/bash
~~~

不要写着 Bash 语法却用：

~~~bash
sh script.sh
~~~

sh 会忽略脚本 shebang，并使用 /bin/sh；Ubuntu 的 /bin/sh 通常是 dash，不支持很多 Bash 语法。

正确执行：

~~~bash
chmod +x script.sh
./script.sh
~~~

或明确：

~~~bash
bash script.sh
~~~

### 2.2 先做静态检查

~~~bash
bash -n script.sh
shellcheck script.sh
~~~

bash -n 只检查语法，不执行。ShellCheck 能发现大量引用、数组、管道和可移植性问题：

~~~bash
sudo apt update
sudo apt install -y shellcheck
~~~

格式化可选用 shfmt，但格式化不能代替逻辑验证。

---

## 3. 变量展开：大多数 Shell bug 的来源

### 3.1 赋值等号两边不能有空格

~~~bash
name='shortlink'
port=8080
~~~

下面不是赋值，而是在尝试执行名为 name 的命令：

~~~text
name = shortlink
~~~

### 3.2 默认给展开加双引号

~~~bash
config_file='/etc/shortlink/shortlink.env'
cp -- "$config_file" "$backup_dir/"
~~~

没有双引号时，变量内容可能发生：

- 单词分割。
- 通配符展开。
- 空值导致参数消失。

错误示例：

~~~bash
rm $target
~~~

更安全：

~~~bash
rm -- "$target"
~~~

-- 表示后续内容不再按选项解析，避免文件名以减号开头时被当作命令选项。

### 3.3 单引号、双引号和无引号

~~~bash
name='shortlink'

printf '%s\n' '$name'
printf '%s\n' "$name"
printf '%s\n' $name
~~~

结果：

- 单引号：原样输出 $name。
- 双引号：展开变量，但保持为一个参数。
- 无引号：展开后还会做分词和通配符展开。

构造命令参数时，优先使用数组，而不是把整条命令拼成字符串。

### 3.4 数组保存参数

~~~bash
curl_args=(
  --silent
  --show-error
  --fail
  --connect-timeout 2
  --max-time 5
)

curl "${curl_args[@]}" \
  'http://127.0.0.1:8080/healthz'
~~~

每个数组元素保持为一个参数。不要使用 eval 执行用户可控字符串，eval 很容易产生命令注入。

### 3.5 位置参数

~~~bash
#!/usr/bin/env bash

printf 'script=%s\n' "$0"
printf 'first=%s\n' "${1:-}"
printf 'count=%s\n' "$#"

for arg in "$@"; do
  printf 'arg=%s\n' "$arg"
done
~~~

"$@" 能保持每个参数的边界；"$*" 会把所有参数合成一个字符串，含义不同。

### 3.6 参数默认值与必填值

~~~bash
environment="${ENVIRONMENT:-development}"
artifact="${1:?usage: deploy.sh ARTIFACT CHECKSUM}"
checksum_file="${2:?usage: deploy.sh ARTIFACT CHECKSUM}"
~~~

- ${VAR:-default}：未设置或为空时使用默认值。
- ${VAR:?message}：未设置或为空时打印错误并退出。

对密码等秘密值，不要在错误信息中打印实际内容。

---

## 4. set -Eeuo pipefail：有用，但不是保险

常见开头：

~~~bash
set -Eeuo pipefail
IFS=$'\n\t'
~~~

### 4.1 每项含义

- -e：未被条件结构处理的命令返回非零时退出。
- -E：让 ERR trap 在函数、命令替换和子 shell 中更容易继承。
- -u：使用未设置变量时退出。
- pipefail：管道中任一命令失败，管道整体失败。
- IFS 调整：减少意外按空格分割；但正确引用仍然是核心。

### 4.2 -e 有上下文例外

下面的失败是条件判断的一部分，不会直接退出：

~~~bash
if curl -fsS http://127.0.0.1:8080/healthz; then
  echo 'healthy'
else
  echo 'unhealthy'
fi
~~~

这正是需要的行为。

但不要以为 -e 能捕获所有逻辑错误。命令放在 if、while、until、&&、|| 或某些子 shell 上下文时，规则会变化。关键操作仍应显式判断。

### 4.3 pipefail 的价值

没有 pipefail：

~~~bash
generate_data | gzip > backup.gz
~~~

如果 generate_data 失败而 gzip 正常结束，管道可能返回成功，留下不完整备份。

启用 pipefail 后，前面的失败会传播出来。

### 4.4 不要滥用 || true

~~~bash
important_command || true
~~~

这会吞掉所有错误。只有“失败确实可接受”时才使用，并解释原因：

~~~bash
if ! grep -q '^optional=' "$config_file"; then
  log 'optional setting is absent; using default'
fi
~~~

### 4.5 ERR trap 只做诊断

~~~bash
on_error() {
  local exit_code=$?
  printf 'ERROR line=%s command=%q exit=%s\n' \
    "${BASH_LINENO[0]}" \
    "$BASH_COMMAND" \
    "$exit_code" >&2
}

trap on_error ERR
~~~

ERR trap 的触发也受 -e 上下文规则影响。它适合补充诊断，不应作为唯一回滚机制。

---

## 5. 条件、循环和函数

### 5.1 使用 [[ ]]

~~~bash
if [[ -f "$artifact" && -x "$artifact" ]]; then
  echo 'artifact is an executable file'
fi

if [[ "$environment" == 'production' ]]; then
  echo 'production deployment'
fi
~~~

常用文件判断：

- -e：路径存在。
- -f：普通文件。
- -d：目录。
- -L：符号链接。
- -r、-w、-x：可读、可写、可执行。
- -s：文件存在且非空。

### 5.2 算术循环

~~~bash
for ((attempt = 1; attempt <= 10; attempt++)); do
  if curl -fsS --max-time 2 "$health_url" >/dev/null; then
    echo 'healthy'
    break
  fi
  sleep 1
done
~~~

自动化中的重试应有：

- 最大次数或总时限。
- 每次操作自己的超时。
- 最终明确失败。
- 不重试明显不可恢复的输入错误。

### 5.3 函数返回状态

~~~bash
is_healthy() {
  curl \
    --silent \
    --show-error \
    --fail \
    --connect-timeout 1 \
    --max-time 2 \
    "$1" >/dev/null
}

if is_healthy 'http://127.0.0.1:8080/healthz'; then
  echo 'ok'
fi
~~~

函数用 return 返回 0 到 255 的状态码，用 stdout 返回文本。不要用 return 返回业务字符串。

### 5.4 日志函数

~~~bash
log() {
  printf '%s %s\n' \
    "$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    "$*" >&2
}

die() {
  log "ERROR: $*"
  exit 1
}
~~~

日志写 stderr，正常数据写 stdout，便于调用方分别处理。

永远不要记录：

- 完整 DSN。
- Authorization 请求头。
- 数据库和 Redis 密码。
- 私钥内容。

---

## 6. heredoc：是否展开由分隔符决定

### 6.1 带引号的分隔符：内容原样保留

~~~bash
cat <<'EOF'
user=$USER
time=$(date)
EOF
~~~

输出中的 $USER 和 $(date) 不会展开。

这适合生成：

- Go 源码。
- 包含 $ 的配置模板。
- 需要稍后由另一个 shell 展开的文件。

例如：

~~~bash
sudo tee /etc/profile.d/go.sh >/dev/null <<'EOF'
export PATH=/usr/local/go/bin:$PATH
EOF
~~~

我们希望用户登录时再展开 $PATH，所以必须使用带引号的 EOF。

### 6.2 不带引号的分隔符：当前 shell 展开

~~~bash
cat <<EOF
user=$USER
time=$(date)
EOF
~~~

适合确实要把当前值写入文件的场景。

风险是文件中的 $、反斜杠和命令替换可能被意外处理。秘密值还可能进入生成文件或日志。使用前应明确每个展开。

### 6.3 here-string

~~~bash
read -r first_word rest <<< "$line"
~~~

短输入可以使用 <<<，大文本仍优先 heredoc 或文件。

---

## 7. 临时文件、trap 与原子替换

### 7.1 使用 mktemp

不要自己猜一个 /tmp/result.tmp：

~~~bash
tmp_file="$(mktemp)"
tmp_dir="$(mktemp -d)"
~~~

固定名字会产生并发冲突和符号链接攻击风险。

### 7.2 保证清理

~~~bash
tmp_dir="$(mktemp -d)"

cleanup() {
  rm -rf -- "$tmp_dir"
}

trap cleanup EXIT
~~~

这里 tmp_dir 由 mktemp 直接生成，仍建议在真实脚本中验证它非空且位于预期父目录后再递归删除。对动态计算出来的任意路径，绝不能直接 rm -rf。

更保守的文件清理：

~~~bash
tmp_file="$(mktemp)"
trap 'rm -f -- "$tmp_file"' EXIT
~~~

### 7.3 同一文件系统内原子切换

写配置或发布指针时，先准备临时对象，再 rename：

~~~bash
tmp_link='/opt/shortlink/.current.new'
ln -s '/opt/shortlink/releases/20260715T120000Z' "$tmp_link"
mv -Tf -- "$tmp_link" /opt/shortlink/current
~~~

同一文件系统内的 rename 是原子的：观察者看到旧链接或新链接，不会看到“复制到一半”的链接。

它不代表整个发布事务原子。切换后服务重启、健康检查和数据库迁移仍可能失败，所以还需要回滚设计。

### 7.4 防止并发发布

~~~bash
exec 9>/opt/shortlink/.deploy.lock

if ! flock -n 9; then
  echo 'another deployment is running' >&2
  exit 1
fi
~~~

文件描述符 9 在脚本存活期间保持锁。脚本退出后内核自动释放。

---

## 8. 非交互环境与可重复执行

脚本在你的终端能跑，不代表在 systemd、cron 或 CI 中能跑。非交互环境常见差异：

- PATH 更短。
- 当前目录不同。
- 没加载 .bashrc。
- 没有 TTY，无法等待密码。
- locale 不同。
- 环境变量更少。

脚本应主动确定自身位置：

~~~bash
SCRIPT_DIR="$(
  cd -- "$(dirname -- "${BASH_SOURCE[0]}")"
  pwd -P
)"
~~~

设置可预测环境：

~~~bash
export PATH='/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin'
export LC_ALL=C.UTF-8
umask 027
~~~

不要依赖“我刚好在项目根目录”。显式 cd：

~~~bash
PROJECT_ROOT="$(cd -- "$SCRIPT_DIR/.." && pwd -P)"
cd "$PROJECT_ROOT"
~~~

### 幂等性

幂等不是“忽略错误”，而是重复执行能收敛到同一状态。

~~~bash
install -d -m 0755 /opt/shortlink/releases
~~~

目录已存在时仍可成功。

反例：

~~~bash
echo 'export PATH=...' >> ~/.profile
~~~

每执行一次都会追加重复内容。应检查、使用受管理文件，或由配置管理工具维护。

---

## 9. 构建短链服务

假设项目结构：

~~~text
shortlink/
├── cmd/server/main.go
├── go.mod
├── go.sum
├── internal/
└── scripts/build.sh
~~~

完整 build.sh：

~~~bash
#!/usr/bin/env bash
set -Eeuo pipefail
IFS=$'\n\t'

export PATH='/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin'
export LC_ALL=C.UTF-8
umask 027

log() {
  printf '%s %s\n' \
    "$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    "$*" >&2
}

die() {
  log "ERROR: $*"
  exit 1
}

script_dir="$(
  cd -- "$(dirname -- "${BASH_SOURCE[0]}")"
  pwd -P
)"
project_root="$(cd -- "$script_dir/.." && pwd -P)"
output_dir="$project_root/dist"
output_file="$output_dir/shortlink"

command -v go >/dev/null 2>&1 || die 'go is not in PATH'
[[ -f "$project_root/go.mod" ]] || die 'go.mod not found'
[[ -d "$project_root/cmd/server" ]] || die 'cmd/server not found'

goos="${GOOS:-linux}"
goarch="${GOARCH:-amd64}"

case "$goos/$goarch" in
  linux/amd64|linux/arm64) ;;
  *) die "unsupported target: $goos/$goarch" ;;
esac

if version_from_git="$(
  git -C "$project_root" describe \
    --tags \
    --always \
    --dirty 2>/dev/null
)"; then
  version="$version_from_git"
else
  version='development'
fi

mkdir -p "$output_dir"
tmp_binary="$(mktemp "$output_dir/.shortlink.XXXXXX")"

cleanup() {
  rm -f -- "$tmp_binary"
}
trap cleanup EXIT

cd "$project_root"

log 'running tests'
go test ./...

log 'running go vet'
go vet ./...

log "building version=$version target=$goos/$goarch"
CGO_ENABLED=0 \
GOOS="$goos" \
GOARCH="$goarch" \
go build \
  -trimpath \
  -ldflags='-s -w' \
  -o "$tmp_binary" \
  ./cmd/server

chmod 0755 "$tmp_binary"
mv -f -- "$tmp_binary" "$output_file"

(
  cd "$output_dir"
  sha256sum shortlink > shortlink.sha256
)

log "artifact=$output_file"
log "checksum=$output_file.sha256"
go version -m "$output_file"
~~~

说明：

- 先 test、vet，再构建。
- 在 dist 内创建临时文件，最终 mv 在同一文件系统中完成。
- 生成 SHA-256，传输到服务器后重新校验。
- CGO_ENABLED=0 适合不依赖 cgo 的纯 Go 服务，便于生成独立二进制；如果项目依赖 cgo，必须按目标系统构建和测试，不能机械关闭。
- -s -w 会减少调试信息。若线上需要更完整堆栈或符号，可移除。
- version 目前只记录在日志。若要注入 Go 变量，需要项目中存在可注入的包级字符串，并使用准确的完整包路径。

SHA-256 能发现传输损坏，但如果攻击者能同时替换二进制和校验文件，它不能证明发布者身份。生产制品还应来自受控仓库，并使用签名或其他可信发布链验证真实性。

运行：

~~~bash
chmod +x scripts/build.sh
./scripts/build.sh
sha256sum -c dist/shortlink.sha256
~~~

---

## 10. 发布目录与 systemd 前提

推荐布局：

~~~text
/opt/shortlink/
├── current -> /opt/shortlink/releases/20260715T120000Z-a1b2c3d4
└── releases/
    ├── 20260714T100000Z-11223344/
    │   └── shortlink
    └── 20260715T120000Z-a1b2c3d4/
        └── shortlink

/etc/shortlink/shortlink.env
/var/lib/shortlink/
~~~

原则：

- release 目录创建后不再原地改二进制。
- current 只是版本指针。
- 配置和业务数据不放 release 目录。
- 旧 release 暂时保留，作为二进制回滚目标。

示例 systemd unit：

~~~ini
[Unit]
Description=Shortlink Go API
After=network-online.target mysql.service redis-server.service
Wants=network-online.target

[Service]
Type=simple
User=shortlink
Group=shortlink
WorkingDirectory=/var/lib/shortlink
EnvironmentFile=/etc/shortlink/shortlink.env
ExecStart=/opt/shortlink/current/shortlink

Restart=on-failure
RestartSec=2s
TimeoutStopSec=20s
KillSignal=SIGTERM

UMask=0027
NoNewPrivileges=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=strict
ReadWritePaths=/var/lib/shortlink
CapabilityBoundingSet=

[Install]
WantedBy=multi-user.target
~~~

保存后：

~~~bash
sudo systemd-analyze verify /etc/systemd/system/shortlink.service
sudo systemctl daemon-reload
sudo systemctl enable shortlink
~~~

这里没有写 SuccessExitStatus=143。Go 服务应捕获 SIGTERM、完成优雅关闭并正常返回。systemd 对非 oneshot 服务收到常见终止信号也有自己的成功判断；不能把 shell 常见的 128+信号编号解释机械套入所有 systemd 重启语义。

### Go 的优雅关闭

核心示例：

~~~go
server := &http.Server{
	Addr:              "127.0.0.1:8080",
	Handler:           router,
	ReadHeaderTimeout: 5 * time.Second,
}

serverErrors := make(chan error, 1)
go func() {
	serverErrors <- server.ListenAndServe()
}()

signals := make(chan os.Signal, 1)
signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
defer signal.Stop(signals)

select {
case err := <-serverErrors:
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("http server: %v", err)
	}
case sig := <-signals:
	log.Printf("received signal %s", sig)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		_ = server.Close()
		log.Fatalf("graceful shutdown: %v", err)
	}
}
~~~

完整项目还应关闭数据库、Redis、后台 goroutine，并停止接收新任务。

---

## 11. 完整可靠发布脚本

服务器脚本路径可设为 /usr/local/sbin/deploy-shortlink。它接收：

1. Go 二进制路径。
2. SHA-256 文件路径。

功能：

- 验证 root 权限和输入。
- 校验 SHA-256。
- 使用 flock 阻止并发发布。
- 创建不可变 release。
- 原子切换 current。
- 重启并等待健康检查。
- 失败时自动切回上一版本。
- 首次发布失败时停止服务并移除 current。

~~~bash
#!/usr/bin/env bash
set -Eeuo pipefail
IFS=$'\n\t'

export PATH='/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin'
export LC_ALL=C.UTF-8
umask 027

readonly app_root='/opt/shortlink'
readonly releases_dir="$app_root/releases"
readonly current_link="$app_root/current"
readonly service_name='shortlink.service'
readonly health_url='http://127.0.0.1:8080/healthz'
readonly health_attempts=20
readonly health_interval=1

tmp_link=''

log() {
  printf '%s %s\n' \
    "$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    "$*" >&2
}

die() {
  log "ERROR: $*"
  exit 1
}

cleanup() {
  if [[ -n "$tmp_link" && -L "$tmp_link" ]]; then
    unlink -- "$tmp_link"
  fi
}
trap cleanup EXIT

[[ "$EUID" -eq 0 ]] || die 'run as root'
[[ "$#" -eq 2 ]] || die 'usage: deploy-shortlink ARTIFACT CHECKSUM_FILE'

artifact="$1"
checksum_file="$2"

[[ -f "$artifact" && ! -L "$artifact" ]] ||
  die 'artifact must be a regular non-symlink file'
[[ -f "$checksum_file" && -s "$checksum_file" && ! -L "$checksum_file" ]] ||
  die 'checksum must be a non-empty regular non-symlink file'

artifact="$(readlink -f -- "$artifact")"
checksum_file="$(readlink -f -- "$checksum_file")"

command -v flock >/dev/null 2>&1 || die 'flock is required'
command -v sha256sum >/dev/null 2>&1 || die 'sha256sum is required'
command -v curl >/dev/null 2>&1 || die 'curl is required'
command -v systemctl >/dev/null 2>&1 || die 'systemctl is required'

expected_checksum="$(
  awk 'NR == 1 { print $1 }' "$checksum_file"
)"

[[ "$expected_checksum" =~ ^[[:xdigit:]]{64}$ ]] ||
  die 'checksum file does not start with a valid SHA-256'

actual_checksum="$(
  sha256sum "$artifact" | awk '{ print $1 }'
)"

[[ "${actual_checksum,,}" == "${expected_checksum,,}" ]] ||
  die 'artifact checksum mismatch'

install -d -o root -g root -m 0755 "$app_root"
install -d -o root -g root -m 0755 "$releases_dir"

exec 9>"$app_root/.deploy.lock"
if ! flock -n 9; then
  die 'another deployment is already running'
fi

release_id="$(
  printf '%s-%s' \
    "$(date -u +'%Y%m%dT%H%M%SZ')" \
    "${actual_checksum:0:8}"
)"
release_dir="$releases_dir/$release_id"

[[ ! -e "$release_dir" ]] ||
  die "release already exists: $release_dir"

previous_target=''
if [[ -L "$current_link" ]]; then
  previous_target="$(readlink -f -- "$current_link" || true)"
  case "$previous_target" in
    "$releases_dir"/*) ;;
    *) die "current points outside releases: $previous_target" ;;
  esac
elif [[ -e "$current_link" ]]; then
  die 'current exists but is not a symbolic link'
fi

install -d -o root -g root -m 0755 "$release_dir"
install -o root -g root -m 0755 \
  "$artifact" \
  "$release_dir/shortlink"

installed_checksum="$(
  sha256sum "$release_dir/shortlink" | awk '{ print $1 }'
)"
[[ "$installed_checksum" == "$actual_checksum" ]] ||
  die 'installed artifact checksum mismatch'

atomic_switch() {
  local target="$1"

  [[ -d "$target" ]] ||
    die "switch target is not a directory: $target"

  case "$target" in
    "$releases_dir"/*) ;;
    *) die "refusing target outside releases: $target" ;;
  esac

  tmp_link="$app_root/.current.$BASHPID"
  [[ ! -e "$tmp_link" && ! -L "$tmp_link" ]] ||
    die "temporary link already exists: $tmp_link"

  ln -s -- "$target" "$tmp_link"
  mv -Tf -- "$tmp_link" "$current_link"
  tmp_link=''
}

wait_for_health() {
  local attempt

  for ((attempt = 1; attempt <= health_attempts; attempt++)); do
    if curl \
      --silent \
      --show-error \
      --fail \
      --connect-timeout 1 \
      --max-time 2 \
      "$health_url" >/dev/null; then
      log "health check passed attempt=$attempt"
      return 0
    fi

    log "health check failed attempt=$attempt/$health_attempts"
    sleep "$health_interval"
  done

  return 1
}

log "activating release=$release_id"
atomic_switch "$release_dir"

if ! systemctl restart "$service_name"; then
  log 'service restart command failed'
elif wait_for_health; then
  log "deployment succeeded release=$release_id"
  exit 0
else
  log 'new release did not become healthy'
fi

journalctl \
  -u "$service_name" \
  -n 80 \
  --no-pager >&2 || true

if [[ -n "$previous_target" && -d "$previous_target" ]]; then
  log "rolling back target=$previous_target"
  atomic_switch "$previous_target"

  if systemctl restart "$service_name" && wait_for_health; then
    log 'rollback succeeded'
  else
    log 'CRITICAL: rollback target is also unhealthy'
  fi
else
  log 'first deployment failed; removing active link and stopping service'

  if [[ -L "$current_link" ]]; then
    unlink -- "$current_link"
  fi

  systemctl stop "$service_name" || true
fi

die "deployment failed; failed release kept at $release_dir for inspection"
~~~

### 11.1 为什么这个脚本比“复制并重启”稳

校验在写入 release 前完成：

- 传输损坏不会进入 active 版本。

release 不覆盖：

- 旧版本仍可定位。
- 回滚不依赖“旧文件是否碰巧存在”。

符号链接原子替换：

- current 不会经历半更新。

首次发布单独处理：

- 没有 previous 时不会尝试回滚到空路径。

健康失败自动回滚：

- current 切回上一 release。
- 重启旧版本并再次健康检查。

发布失败的新 release 保留：

- 方便检查 ELF、配置兼容性和日志。
- 清理由单独的、可审计流程执行。

### 11.2 边界

这个脚本仍不能自动保证：

- 数据库迁移可回滚。
- 外部依赖健康。
- 新版本所有业务路径正确。
- 多实例流量无损切换。
- 配置变更和二进制完全兼容。

健康接口至少应证明：

- 进程事件循环可响应。
- 必要配置已加载。

是否把 MySQL、Redis 状态纳入 liveness 或 readiness 要谨慎。若 liveness 因数据库短暂抖动失败，服务可能陷入无意义重启。通常：

- liveness：进程自身是否还能工作。
- readiness：是否适合接收业务流量，可包含关键依赖检查。

单机 systemd 场景也应区分思路。

---

## 12. 上传并触发发布

在构建机：

~~~bash
./scripts/build.sh

upload_dir="/tmp/shortlink-upload-$(date -u +'%Y%m%dT%H%M%SZ')"

ssh shortlink-prod "install -d -m 0700 '$upload_dir'"

scp \
  dist/shortlink \
  dist/shortlink.sha256 \
  "shortlink-prod:$upload_dir/"

ssh -t shortlink-prod \
  "sudo /usr/local/sbin/deploy-shortlink \
    '$upload_dir/shortlink' \
    '$upload_dir/shortlink.sha256'"
~~~

这个示例假设 upload_dir 是脚本自身生成的安全格式，不接受任意用户文本。如果目录来自外部输入，必须额外校验，不能靠拼接引号防住所有情况。

更成熟的流程会：

- CI 构建并签名产物。
- 服务器从制品库拉取指定不可变版本。
- 审批后由受限部署身份触发。
- 记录发布人、版本、时间和结果。

### 限制 sudo 权限

不要让 deploy 用户获得无密码执行任意 root 命令的权限。可以只允许固定脚本路径，并确保该脚本及其父目录不可由 deploy 用户修改。

sudoers 修改必须使用：

~~~bash
sudo visudo -f /etc/sudoers.d/shortlink-deploy
~~~

具体授权要结合部署账户、参数控制和组织安全策略。仅限制命令名但允许脚本读取任意路径，仍可能形成提权面，所以发布脚本必须继续验证输入文件来源、所有权和路径范围。上面的教学脚本适合受控单机环境，生产 CI 还应进一步收紧上传目录和文件所有权。

---

## 13. 数据库迁移为什么不能跟二进制一起“无脑回滚”

假设新版本执行：

~~~sql
ALTER TABLE short_links DROP COLUMN old_field;
~~~

新版本失败后即使切回旧二进制，旧程序可能仍依赖 old_field，回滚会继续失败。

更安全的 expand-contract：

1. Expand：先新增表、列或索引，不删除旧结构。
2. 部署兼容新旧结构的代码。
3. 回填数据并观察。
4. 所有实例不再依赖旧结构后，下一次发布再 Contract。

迁移工具应：

- 记录已执行版本。
- 使用 shortlink_migrator，而不是运行账户。
- 每次变更可审查。
- 明确哪些 DDL 能否事务回滚。
- 发布前备份并做过恢复演练。

发布脚本可以调用迁移工具，但必须为迁移失败和二进制回滚设计清晰顺序。初学阶段宁可把迁移作为独立显式步骤，也不要伪装成“一键绝对安全”。

---

## 14. 版本保留与清理

不要把 rm -rf 塞进主发布路径。先列出：

~~~bash
find /opt/shortlink/releases \
  -mindepth 1 \
  -maxdepth 1 \
  -type d \
  -printf '%TY-%Tm-%Td %TH:%TM:%TS %p\n' |
  sort -r

readlink -f /opt/shortlink/current
~~~

清理策略可以是：

- 永远保留 current。
- 至少保留最近 3 到 5 个健康 release。
- 保留最近一次人工确认的稳定版本。
- 磁盘压力和审计要求共同决定期限。

删除前验证：

1. 解析后的路径位于 /opt/shortlink/releases/ 下。
2. 不是 current 指向目标。
3. 不是计划中的回滚版本。
4. 目录名符合脚本生成格式。
5. 打印 dry-run 列表并人工确认。

可靠脚本的一个重要品质，是拒绝对不确定路径执行递归删除。

---

## 15. 定时任务：优先考虑 systemd timer

cron 可以运行简单任务，但 systemd timer 更容易统一日志、依赖和资源限制。

例如定期检查短链过期数据，service：

~~~ini
[Unit]
Description=Shortlink expiration cleanup

[Service]
Type=oneshot
User=shortlink
Group=shortlink
EnvironmentFile=/etc/shortlink/shortlink.env
ExecStart=/opt/shortlink/current/shortlink cleanup-expired
~~~

timer：

~~~ini
[Unit]
Description=Run shortlink expiration cleanup hourly

[Timer]
OnCalendar=hourly
Persistent=true
RandomizedDelaySec=5m

[Install]
WantedBy=timers.target
~~~

验证：

~~~bash
sudo systemd-analyze verify \
  /etc/systemd/system/shortlink-cleanup.service \
  /etc/systemd/system/shortlink-cleanup.timer

sudo systemctl daemon-reload
sudo systemctl enable --now shortlink-cleanup.timer
systemctl list-timers --all
journalctl -u shortlink-cleanup.service
~~~

Persistent=true 表示机器关机错过任务后，重新启动会补跑。业务是否允许补跑要根据任务语义判断。

清理任务本身仍要：

- 分批处理。
- 设置超时。
- 防止并发执行。
- 记录数量而不泄露敏感数据。
- 保证重复执行安全。

---

## 16. 常见失败模式

### 16.1 Windows CRLF

报错：

~~~text
/usr/bin/env: ‘bash\r’: No such file or directory
~~~

检查：

~~~bash
file script.sh
sed -n '1l' script.sh
~~~

修复仓库的行尾规则，而不是每次上线手工转换。可在 .gitattributes 中为 shell 脚本指定 LF。

### 16.2 Permission denied

~~~bash
ls -l script.sh
namei -l /path/to/script.sh
mount | grep ' /opt '
~~~

可能是：

- 没有执行位。
- 父目录无 x 权限。
- 文件系统以 noexec 挂载。
- shebang 解释器不可执行。

### 16.3 command not found

~~~bash
command -v go
printf '%s\n' "$PATH"
~~~

在脚本中设置明确 PATH，不依赖交互 shell 配置。

### 16.4 脚本停住

常见原因：

- sudo 等待密码。
- ssh 等待首次 host key 确认。
- curl 没有超时。
- 命令等待 stdin。
- 包管理器弹出交互配置。

自动化前先消除交互，并为网络命令设置超时。

### 16.5 管道掩盖前序失败

启用 pipefail，并检查生成文件是否非空、校验是否通过。仅凭最后一个命令返回 0 不够。

### 16.6 source 一个不可信 .env

Shell 的 source 会执行文件中的命令：

~~~bash
source .env
~~~

只有当文件是受信任的 Shell 程序时才能这样做。用户上传或外部生成的键值文件不能直接 source。systemd EnvironmentFile 的语法也不等同于完整 Bash。

### 16.7 把 set -x 开在秘密环境

set -x 会打印展开后的命令，很容易泄露密码和 Token。诊断时也应对敏感段落关闭：

~~~bash
set +x
~~~

更好的做法是设计不包含秘密的结构化日志。

---

## 17. 小实验：验证失败传播和清理

建立脚本：

~~~bash
mkdir -p ~/labs/shell-reliability
cd ~/labs/shell-reliability

cat > demo.sh <<'EOF'
#!/usr/bin/env bash
set -Eeuo pipefail

tmp_file="$(mktemp)"

cleanup() {
  printf 'cleanup %s\n' "$tmp_file" >&2
  rm -f -- "$tmp_file"
}
trap cleanup EXIT

printf 'partial data\n' > "$tmp_file"

false | gzip > result.gz

mv -f -- "$tmp_file" result.txt
EOF

chmod +x demo.sh
./demo.sh
printf 'exit=%s\n' "$?"
ls -la
~~~

预期：

- false 使管道失败。
- pipefail 让脚本退出。
- EXIT trap 删除临时文件。
- result.txt 不存在。
- result.gz 可能已经被 shell 创建但内容无意义，说明“清理临时资源”和“清理所有外部副作用”不是同一回事。

改进方式是也先把 result.gz 写到临时路径，完整成功后再原子移动。

---

## 18. 本章验收

知识验收：

1. 为什么使用 "$@" 而不是 $*？
2. 带引号和不带引号的 heredoc 分隔符有什么差别？
3. set -e 为什么不能替代显式错误处理？
4. pipefail 解决什么问题？
5. 为什么发布目录不可原地覆盖？
6. 原子切换 current 为什么仍不等于整个发布原子？
7. 首次发布失败为什么要单独处理？
8. 为什么数据库迁移不能跟随二进制无脑回滚？

动手验收：

- 能通过 bash -n 和 ShellCheck 检查脚本。
- 能写一个使用 mktemp 和 EXIT trap 的脚本。
- 能用 flock 阻止同一脚本并发运行。
- 能构建 Go 二进制并生成、验证 SHA-256。
- 能解释发布脚本每一个失败分支。
- 能模拟健康检查失败并确认 current 回到旧 release。
- 能从 journalctl 找到新版本启动失败原因。

真正掌握这一章的标准不是记住 Bash 语法，而是发布失败时你知道系统停在什么状态、为什么，以及下一步怎样安全恢复。
