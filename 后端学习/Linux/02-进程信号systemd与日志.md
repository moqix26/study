# 02 进程、信号、systemd 与日志

Go 源码不会直接“在服务器上运行”。它先被编译成二进制，内核加载二进制后创建进程；进程拥有 PID、用户身份、文件描述符、内存和网络 socket。生产环境还需要一个服务管理器负责启动、停止、重启、资源限制和日志收集，这正是 systemd 的职责。

本章完成后，你应该能够：

- 观察进程的 PID、父子关系、状态、资源和打开文件。
- 分清前台任务、后台任务与正式系统服务。
- 正确使用 `SIGINT`、`SIGTERM`、`SIGKILL`，理解 Go 优雅关闭。
- 编写和验证一个安全的 systemd service unit。
- 使用 journald 查启动日志、实时日志和历史故障。
- 沿“服务状态 → 进程 → 端口 → 本机请求 → 日志”定位问题。

涉及 systemd 的主实验建议在 Ubuntu 24.04 VM 或云服务器完成。WSL2 可以辅助练习，但不要把它当成唯一的部署结论。

## 1. 程序、进程和线程

- **程序**：磁盘上的可执行文件，例如 `/opt/shortlink/bin/shortlink`。
- **进程**：程序的一次运行实例，拥有独立 PID 和运行时资源。
- **线程**：进程内由内核调度的执行单元。
- **goroutine**：Go 运行时调度的轻量并发任务，多个 goroutine 会复用若干操作系统线程。

同一个二进制可以启动多次，产生不同 PID。删除或替换磁盘上的二进制，不会自动改变已经加载并运行的旧进程；要让新版本生效，需要按发布流程重启服务。

查看当前 Shell 和它启动的进程：

```bash
printf 'shell pid=%s\n' "$$"
ps -p $$ -o pid,ppid,user,stat,lstart,cmd
pstree -ap $$
```

若 `pstree` 不存在：

```bash
sudo apt update
sudo apt install -y psmisc
```

## 2. 查找和观察进程

### 2.1 `ps` 是某一时刻的快照

```bash
ps aux | less
ps -eo pid,ppid,user,stat,%cpu,%mem,etime,cmd --sort=-%cpu | head -n 20
```

重点字段：

| 字段 | 含义 |
|---|---|
| PID | 进程 ID |
| PPID | 父进程 ID |
| USER | 进程实际所属用户 |
| STAT | 状态与附加标志 |
| %CPU / %MEM | 采样区间内 CPU 与内存占比 |
| ELAPSED / ETIME | 运行时长 |
| CMD | 启动命令和参数 |

不要用 `ps aux | grep shortlink` 后不加判断就杀第一个 PID，因为结果可能包含 grep 自身、编辑器或其他相似命令。更明确地查进程：

```bash
pgrep -a -x shortlink
systemctl show -p MainPID --value shortlink.service
```

systemd 托管的服务优先通过 unit 找主进程，不要凭模糊文本匹配。

### 2.2 `top` / `htop` 是动态视图

```bash
top
```

`top` 中按 `P` 按 CPU 排序、`M` 按内存排序、`1` 展开每个 CPU、`q` 退出。若喜欢更直观界面：

```bash
sudo apt install -y htop
htop
```

一次瞬时高 CPU 不等于故障。应观察持续时间、请求量、延迟和日志，再决定是否处理。

### 2.3 常见进程状态

`STAT` 的主要首字母：

- `R`：正在运行或可运行。
- `S`：可中断睡眠，等待事件；服务大部分时间处于 `S` 很正常。
- `D`：不可中断睡眠，常在等待内核 I/O；长时间大量出现需要关注磁盘或网络存储。
- `T`：被作业控制或调试信号暂停。
- `Z`：僵尸进程，子进程已退出但父进程尚未回收退出状态。

僵尸进程本身不继续消耗普通用户内存，但占用 PID 表项。应找父进程为何不执行 wait，而不是对僵尸 PID 使用 `kill -9`；僵尸已经退出，无法再被信号“杀死”。

## 3. `/proc`：内核提供的进程视图

每个进程在 `/proc/<PID>` 下都有一组虚拟文件。选择一个当前 Shell：

```bash
pid=$$
tr '\0' ' ' < "/proc/$pid/cmdline"
printf '\n'
tr '\0' '\n' < "/proc/$pid/environ" | head
ls -l "/proc/$pid/fd" | head
cat "/proc/$pid/status" | head -n 20
```

可以观察：

- `cmdline`：启动参数。
- `environ`：进程环境变量。
- `fd/`：打开的文件描述符和 socket。
- `status`：UID、GID、线程数、内存等。
- `limits`：文件描述符、进程数等限制。

安全提醒：不要把数据库密码放在命令行参数中，它可能出现在 `ps` 和 `/proc/<PID>/cmdline`。环境变量也不是“加密保险箱”；同一用户或特权用户可能读取。秘密配置需要权限控制，并在更成熟环境中使用凭据/密钥管理方案。

查看进程打开的网络与文件：

```bash
sudo lsof -p <PID> | less
sudo lsof -Pan -p <PID> -i
```

## 4. 前台、后台和 Shell 作业控制

在当前终端启动：

```bash
sleep 300
```

它占据前台。按 `Ctrl+Z` 发送暂停信号后：

```bash
jobs -l
bg %1
jobs -l
fg %1
```

在前台按 `Ctrl+C` 发送 `SIGINT`，通常终止 `sleep`。

命令末尾的 `&` 直接放入后台：

```bash
sleep 300 &
job_pid=$!
printf 'pid=%s\n' "$job_pid"
jobs -l
kill -TERM "$job_pid"
wait "$job_pid"
```

`$!` 是最近一个后台任务 PID。`wait` 回收并取得退出状态。

### 4.1 后台任务不等于服务

以下做法适合临时实验，不适合生产：

```bash
nohup ./shortlink > shortlink.log 2>&1 &
disown
```

它缺少统一的用户隔离、依赖关系、自动重启、资源限制、状态查询和结构化日志。SSH 会话断开后“进程还活着”只是最低要求，正式 Go 服务应交给 systemd 或容器编排。

## 5. 退出状态

进程正常结束会向父进程提供退出状态。Shell 中：

```bash
true
printf 'status=%s\n' "$?"
false
printf 'status=%s\n' "$?"
```

通常 `0` 表示成功，非零表示某种失败。要立即保存状态，因为下一条命令会覆盖 `$?`：

```bash
some-command
status=$?
printf 'status=%s\n' "$status"
```

Shell 常用 `128 + 信号编号` 表示命令因信号结束，例如 SIGTERM 是 15，某些 Shell 场景中会看到 `143`。这是一种 Shell 表示约定，不应把所有 `143` 都机械加入 systemd 的 `SuccessExitStatus=`。更好的 Go 服务会捕获 SIGTERM、完成清理并以 `0` 正常退出。

## 6. 信号：请求进程改变状态

查看信号表：

```bash
kill -l
```

高频信号：

| 信号 | 常见来源 | 用途 |
|---|---|---|
| `SIGINT` | 终端 `Ctrl+C` | 请求中断前台程序 |
| `SIGTERM` | `kill PID`、systemd stop | 请求程序优雅终止 |
| `SIGHUP` | 终端断开或显式发送 | 传统上表示挂起，部分服务用它重新加载配置 |
| `SIGSTOP` | 显式发送 | 内核强制暂停，不能被捕获 |
| `SIGCONT` | 显式发送 | 恢复暂停进程 |
| `SIGKILL` | `kill -KILL PID` | 内核立即终止，不能捕获或清理 |

`kill` 命令的名字容易误导：默认并不是“强杀”，而是发送 SIGTERM：

```bash
kill -TERM <PID>
```

只有进程无法在合理时间内退出，并且你已经了解未刷盘数据、未完成请求和锁状态的风险时，才考虑：

```bash
kill -KILL <PID>
```

先 TERM、等待、观察日志，再 KILL。不要把 `kill -9` 当成固定第一步。

### 6.1 Go 为什么需要优雅关闭

若进程被立即终止，可能发生：

- 正在处理的 HTTP 请求被中断。
- 数据库事务未完成。
- 缓冲日志尚未刷新。
- 后台 goroutine 没有机会结束。

优雅关闭的基本流程：

```text
收到 SIGTERM
  → 停止接收新请求
  → 给现有请求一个有限完成时间
  → 关闭数据库、缓存和后台任务
  → 返回 0
```

“有限时间”很重要。无限等待会让发布和故障恢复永久卡住，因此应用与 systemd 都应设置停止超时。

## 7. 最小 Go HTTP 服务：正确处理 SIGTERM

在 `~/linux-lab/02-process` 新建 `main.go`，内容如下：

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	addr := os.Getenv("APP_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s", addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	case <-ctx.Done():
		log.Printf("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
			_ = server.Close()
		}
	}

	log.Printf("server stopped")
}
```

创建模块并编译：

```bash
mkdir -p ~/linux-lab/02-process
cd ~/linux-lab/02-process
# 先在编辑器中保存上面的 main.go
go mod init example.com/shortlink-lab
go fmt ./...
go build -trimpath -o shortlink-lab .
```

若 `go` 尚未安装，可先阅读本章，在第 04 章准备 Go 环境后再回来完成实验。

前台运行：

```bash
APP_ADDR=127.0.0.1:8080 ./shortlink-lab
```

另开终端验证：

```bash
curl -i http://127.0.0.1:8080/healthz
pgrep -a -x shortlink-lab
ss -ltnp 'sport = :8080'
```

预期状态码为 `200`，正文为 `ok`。回到服务终端按 `Ctrl+C`，应看到收到信号和服务停止日志，而不是立即被强制中断。

## 8. systemd 的职责

Ubuntu 24.04 使用 systemd 作为 PID 1。它不仅“开机启动服务”，还负责：

- 根据 unit 描述启动和停止进程。
- 按依赖关系排序。
- 以指定用户、工作目录和环境运行。
- 监控主进程并按策略重启。
- 把进程放入 cgroup，统一追踪子进程和资源。
- 将 stdout/stderr 接入 journald。
- 提供安全沙箱和资源限制。

确认：

```bash
ps -p 1 -o pid,comm,args
systemctl is-system-running
systemctl --failed
```

### 8.1 unit 文件来自哪里

高频路径：

```text
/usr/lib/systemd/system/       软件包安装的 unit（部分系统兼容路径为 /lib/systemd/system）
/etc/systemd/system/           管理员自定义 unit 和覆盖配置
/run/systemd/system/           本次启动期间的临时 unit
```

不要直接修改软件包管理的 unit；升级可能覆盖。自定义服务放 `/etc/systemd/system`，修改现有服务用 drop-in：

```bash
systemctl cat ssh.service
sudo systemctl edit ssh.service
```

### 8.2 `start`、`enable` 和 `restart`

```bash
sudo systemctl start shortlink.service
sudo systemctl stop shortlink.service
sudo systemctl restart shortlink.service
sudo systemctl enable shortlink.service
sudo systemctl disable shortlink.service
```

- `start` 影响当前运行状态，不自动设置下次开机。
- `enable` 创建启动依赖关系，不一定立即启动。
- `enable --now` 同时启用并立即启动。
- `restart` 停止旧进程后启动新进程。
- `reload` 只有服务明确支持并配置 `ExecReload=` 时才有意义，不能把它当成轻量 restart。

`daemon-reload` 让 systemd 重新读取 unit 文件，它不会自动重启业务进程：

```bash
sudo systemctl daemon-reload
sudo systemctl restart shortlink.service
```

## 9. 把实验程序安装成系统服务

以下步骤会修改 VM 或云服务器系统状态，需要 sudo。先确认当前目录中的二进制是你刚刚编译并测试过的版本。

### 9.1 创建专用系统用户

```bash
getent passwd shortlink || \
  sudo useradd --system \
    --user-group \
    --home-dir /var/lib/shortlink \
    --shell /usr/sbin/nologin \
    shortlink
```

专用用户不用于交互登录，只给应用进程最小权限。确认：

```bash
getent passwd shortlink
id shortlink
```

### 9.2 安装二进制和配置

在 `~/linux-lab/02-process` 中执行：

```bash
sudo install -d -o root -g root -m 0755 /opt/shortlink/bin
sudo install -o root -g shortlink -m 0750 \
  ./shortlink-lab /opt/shortlink/bin/shortlink

sudo install -d -o root -g shortlink -m 0750 /etc/shortlink
printf 'APP_ADDR=127.0.0.1:8080\n' | \
  sudo tee /etc/shortlink/shortlink.env > /dev/null
sudo chown root:shortlink /etc/shortlink/shortlink.env
sudo chmod 0640 /etc/shortlink/shortlink.env
```

检查：

```bash
namei -l /opt/shortlink/bin/shortlink
namei -l /etc/shortlink/shortlink.env
sudo -u shortlink env APP_ADDR=127.0.0.1:18080 \
  timeout --signal=TERM 3s /opt/shortlink/bin/shortlink &
test_job=$!
sleep 1
curl -fsS http://127.0.0.1:18080/healthz
wait "$test_job" || true
```

如果 8080 已被其他进程占用，不要直接杀未知进程。先用 `sudo ss -ltnp 'sport = :8080'` 查明身份，再调整实验端口或停止明确的旧实例。

### 9.3 编写 unit

创建 `/etc/systemd/system/shortlink.service`：

```bash
sudo nano /etc/systemd/system/shortlink.service
```

```ini
[Unit]
Description=Shortlink Go HTTP service
Wants=network-online.target
After=network-online.target
StartLimitIntervalSec=60
StartLimitBurst=5

[Service]
Type=simple
User=shortlink
Group=shortlink
WorkingDirectory=/var/lib/shortlink
ExecStart=/opt/shortlink/bin/shortlink
EnvironmentFile=/etc/shortlink/shortlink.env

Restart=on-failure
RestartSec=2s
TimeoutStopSec=15s
KillSignal=SIGTERM
UMask=0027

StateDirectory=shortlink
StateDirectoryMode=0750

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictSUIDSGID=true
CapabilityBoundingSet=

[Install]
WantedBy=multi-user.target
```

几个关键点：

- `User` / `Group`：服务不以 root 运行。
- `ExecStart`：直接指向绝对路径，不经过 Shell，不会解释 `|`、`>`、`$VAR` 等 Shell 语法。
- `EnvironmentFile`：缺失会导致启动失败，这里故意要求配置必须存在。
- `Restart=on-failure`：异常退出时重启，正常 stop 不会反复拉起。
- `TimeoutStopSec=15s`：比 Go 内部 10 秒关闭超时略长，给应用清理时间。
- `StateDirectory=shortlink`：由 systemd 创建 `/var/lib/shortlink` 并交给服务用户。
- `ProtectSystem=strict`：系统目录只读；`StateDirectory` 仍是服务被允许写入的状态位置。
- 空的 `CapabilityBoundingSet=`：清除 Linux capabilities；监听 8080 不需要特权端口能力。

`network-online.target` 只表达启动排序意图，不证明 MySQL、Redis 或外部网络已经真正可用。应用仍需要连接超时、重试和就绪检查。

如果应用以后确实需要写其他路径，应精确增加 `ReadWritePaths=` 或使用 systemd 提供的 `CacheDirectory=`、`LogsDirectory=` 等，不要通过关闭所有沙箱换取方便。

### 9.4 验证并启动

先做静态验证：

```bash
sudo systemd-analyze verify /etc/systemd/system/shortlink.service
```

无错误后：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now shortlink.service
systemctl status shortlink.service --no-pager -l
curl -i http://127.0.0.1:8080/healthz
```

查看 systemd 实际解析结果：

```bash
systemctl cat shortlink.service
systemctl show shortlink.service \
  -p User -p Group -p MainPID -p ActiveState -p SubState -p ExecMainStatus
```

`active (running)` 只说明主进程仍在，不保证业务路由和依赖正常，所以还要 curl 健康检查。

### 9.5 验证优雅关闭

一个终端持续看日志：

```bash
sudo journalctl -u shortlink.service -f
```

另一个终端：

```bash
sudo systemctl restart shortlink.service
```

日志应先出现 `shutdown signal received`、`server stopped`，随后新 PID 启动。确认 PID 已变化：

```bash
systemctl show -p MainPID --value shortlink.service
```

Go 程序捕获 SIGTERM 后正常返回，因此一般不需要把 `143` 声明为成功退出状态。

## 10. service unit 的常见类型

- `Type=simple`：`ExecStart` 启动的进程就是主进程；普通 Go HTTP 服务常用。
- `Type=exec`：与 simple 接近，但 systemd 会等到 `execve` 成功后再认为启动成功，现代 systemd 中可作为更严格选择。
- `Type=notify`：服务主动向 systemd 通知已经就绪，需要应用实现通知协议。
- `Type=oneshot`：一次性任务，允许多个顺序执行的 `ExecStart=`。
- `Type=forking`：传统程序自行 fork 到后台，现代 Go 服务不应采用这种模式。

不要让 Go 程序自行 daemonize。让它保持前台、把日志写 stdout/stderr，由 systemd 管理生命周期。

## 11. 重启策略不是万能药

`Restart=on-failure` 通常适合网络服务，但需要理解边界：

- 进程非零退出、异常信号或超时通常触发重启。
- `systemctl stop` 是管理员明确操作，不会立即重新拉起。
- 如果程序启动后马上崩溃，systemd 会反复尝试，达到 StartLimit 后进入 failed，避免无限高速循环。
- 重启不能修复错误配置、数据库迁移失败或磁盘已满，只会重复制造日志。

查看重启次数和失败状态：

```bash
systemctl show shortlink.service \
  -p NRestarts -p Result -p ExecMainCode -p ExecMainStatus
```

修复后清除 failed 状态：

```bash
sudo systemctl reset-failed shortlink.service
sudo systemctl start shortlink.service
```

## 12. journald：服务日志的第一入口

systemd 默认把服务的 stdout/stderr 交给 journald。Go 标准库 `log.Printf` 的输出会自动带上 unit、PID、启动批次等元数据，无需把日志手工重定向到某个文件。

### 12.1 高频查询

```bash
# 当前 unit 最近 100 行
sudo journalctl -u shortlink.service -n 100 --no-pager

# 只看本次系统启动
sudo journalctl -u shortlink.service -b --no-pager

# 最近 10 分钟
sudo journalctl -u shortlink.service --since '10 minutes ago' --no-pager

# 明确时间范围
sudo journalctl -u shortlink.service \
  --since '2026-07-15 10:00:00' \
  --until '2026-07-15 10:30:00' --no-pager

# 实时跟随，Ctrl+C 退出
sudo journalctl -u shortlink.service -f

# 内核日志
sudo journalctl -k -b --no-pager
```

按优先级过滤：

```bash
sudo journalctl -u shortlink.service -p warning..alert --since today
```

但 Go 默认文本日志通常没有自动映射为 journald 的 warning/error 优先级，因此关键字搜索仍有价值：

```bash
sudo journalctl -u shortlink.service --since today --no-pager | \
  rg -i 'error|fatal|panic|timeout'
```

### 12.2 查看结构化字段

```bash
sudo journalctl -u shortlink.service -n 1 -o verbose
sudo journalctl -u shortlink.service -n 5 -o json-pretty
```

常见字段包括 `_SYSTEMD_UNIT`、`_PID`、`_UID`、`_BOOT_ID` 和 `MESSAGE`。这也是 journald 比单纯文本文件更方便关联服务和启动批次的原因。

### 12.3 磁盘占用与持久化

```bash
journalctl --disk-usage
```

Ubuntu 的 `Storage=auto` 会在 `/var/log/journal` 存在时持久化，否则可能主要使用 `/run/log/journal`。先查看实际状态，不要假设重启后一定保留全部日志：

```bash
ls -ld /var/log/journal /run/log/journal 2>/dev/null
journalctl --list-boots
```

需要明确限制时，可以创建 `/etc/systemd/journald.conf.d/shortlink-retention.conf`：

```ini
[Journal]
SystemMaxUse=500M
MaxRetentionSec=14day
```

然后验证配置并重启 journald：

```bash
systemd-analyze cat-config systemd/journald.conf
sudo systemctl restart systemd-journald
```

容量和保留期应根据磁盘、审计需求和集中日志方案决定，不要机械使用示例值。`journalctl --vacuum-time=` 会真正删除旧归档日志，执行前先确认保留要求。

## 13. 日志应该记录什么

有用的后端日志至少应便于回答：

- 什么时候发生。
- 哪个请求或任务发生。
- 哪个阶段失败。
- 错误类型和必要上下文是什么。
- 耗时、状态码、依赖目标是什么。

推荐字段：

```text
time level service request_id method path status duration_ms error
```

不要记录：

- 密码、完整令牌、Cookie、私钥。
- 完整数据库连接串中的密码。
- 没有必要的身份证号、手机号等敏感信息。
- 用户提交的完整长 URL 中可能含有的授权查询参数。

错误日志应保留可行动上下文，但不能以泄露秘密为代价。生产中可使用 JSON 日志交给集中式日志系统；本章先学会使用 journald 证据链。

## 14. cgroup 与资源限制

systemd 将 unit 的进程组织到 cgroup 中，因此不仅跟踪一个 PID，还能管理其子进程。查看：

```bash
systemd-cgls
systemctl status shortlink.service
```

可以通过 drop-in 添加资源上限：

```bash
sudo systemctl edit shortlink.service
```

写入：

```ini
[Service]
MemoryMax=512M
CPUQuota=80%
TasksMax=200
LimitNOFILE=65536
```

保存后：

```bash
sudo systemctl daemon-reload
sudo systemctl restart shortlink.service
systemctl show shortlink.service \
  -p MemoryMax -p CPUQuotaPerSecUSec -p TasksMax -p LimitNOFILE
```

限制值必须通过负载测试确定。太低会把正常峰值误判成故障，太高则失去保护意义。若服务因内存限制被杀，结合以下证据：

```bash
systemctl show shortlink.service -p Result -p OOMPolicy
sudo journalctl -u shortlink.service -b --no-pager
sudo journalctl -k -b --no-pager | rg -i 'oom|out of memory|killed process'
```

## 15. systemd 启动失败的证据链

不要看到 `failed` 就重复 restart。按顺序收集：

```bash
systemctl status shortlink.service --no-pager -l
sudo journalctl -u shortlink.service -b -n 100 --no-pager
systemctl cat shortlink.service
systemctl show shortlink.service \
  -p Result -p ExecMainCode -p ExecMainStatus -p MainPID
sudo systemd-analyze verify /etc/systemd/system/shortlink.service
```

常见状态：

| 提示 | 常见原因 | 检查方向 |
|---|---|---|
| `status=203/EXEC` | 二进制不存在、无执行权限、架构/解释器问题 | `namei -l`、`file`、`ldd`、ExecStart 路径 |
| `status=217/USER` | User/Group 不存在或身份设置失败 | `getent passwd`、`getent group` |
| `status=200/CHDIR` | WorkingDirectory 不存在或不可进入 | `namei -l` 目标目录 |
| `address already in use` | 端口已有监听者 | `sudo ss -ltnp 'sport = :8080'` |
| `Permission denied` | 路径权限或沙箱限制 | `namei -l`、unit hardening、journal |
| `Start request repeated too quickly` | 连续启动失败触发速率限制 | 先修最早错误，再 `reset-failed` |

`systemctl status` 只显示部分日志，真正原因常在更早的 journal 行中。重点找“第一次失败”，后续错误可能只是连锁反应。

## 16. 服务运行但请求失败怎么查

使用固定顺序减少猜测：

### 16.1 服务状态

```bash
systemctl is-active shortlink.service
systemctl show -p MainPID --value shortlink.service
```

### 16.2 端口是否监听

```bash
sudo ss -ltnp 'sport = :8080'
```

应该看到 `127.0.0.1:8080` 和对应进程。若监听 `127.0.0.1`，只能从本机或同一网络命名空间访问，这是同机 Nginx 反代时的预期安全设计。

### 16.3 本机 HTTP 是否成功

```bash
curl -v --max-time 5 http://127.0.0.1:8080/healthz
```

- `Connection refused`：该地址端口没有监听者，或监听在别的地址族/命名空间。
- 超时：请求被丢弃、网络路径异常，或应用接收后卡住。
- HTTP 404：连接和 HTTP 已成功，但路由不匹配。
- HTTP 500：请求已到应用，继续看应用和依赖日志。

### 16.4 同时观察日志

```bash
sudo journalctl -u shortlink.service -f
```

在另一个终端发请求，观察是否产生请求日志。完全没有日志可能表示请求未到应用，也可能表示应用没有记录该路径，不能只凭这一点下结论。

## 17. 主动制造三个可控故障

这些实验只在 VM 或学习机执行，每次改动后都恢复。

### 17.1 错误的二进制路径

创建一个只用于本实验、名称明确的 drop-in：

```bash
sudo install -d -m 0755 /etc/systemd/system/shortlink.service.d
sudo nano /etc/systemd/system/shortlink.service.d/90-lab-bad-exec.conf
```

```ini
# /etc/systemd/system/shortlink.service.d/90-lab-bad-exec.conf
[Service]
ExecStart=
ExecStart=/opt/shortlink/bin/not-exist
```

这里第一行空 `ExecStart=` 用于清除原值。然后：

```bash
sudo systemctl daemon-reload
sudo systemctl restart shortlink.service
systemctl status shortlink.service --no-pager -l
sudo journalctl -u shortlink.service -n 30 --no-pager
```

观察 `203/EXEC`。恢复：

```bash
sudo rm -- /etc/systemd/system/shortlink.service.d/90-lab-bad-exec.conf
sudo systemctl daemon-reload
sudo systemctl reset-failed shortlink.service
sudo systemctl start shortlink.service
```

删除前先用 `systemctl cat shortlink.service` 确认文件名。不要使用会一并清理其他本地覆盖的宽泛恢复操作。

### 17.2 配置端口冲突

先另开一个临时监听者：

```bash
python3 -m http.server 18080 --bind 127.0.0.1
```

把 `/etc/shortlink/shortlink.env` 临时改为 `APP_ADDR=127.0.0.1:18080`，重启服务并观察 `address already in use`。恢复配置到 8080，停止 Python，再启动 shortlink。

这个实验训练的是先用 `ss` 找占用者，而不是看到冲突就随机杀进程。

### 17.3 不存在的运行用户

创建 `/etc/systemd/system/shortlink.service.d/91-lab-bad-user.conf`：

```bash
sudo nano /etc/systemd/system/shortlink.service.d/91-lab-bad-user.conf
```

```ini
[Service]
User=shortlink-missing
```

重载并观察：

```bash
sudo systemctl daemon-reload
sudo systemctl restart shortlink.service
systemctl status shortlink.service --no-pager -l
sudo journalctl -u shortlink.service -n 30 --no-pager
```

应看到与用户身份有关的失败，常见为 `217/USER`。恢复：

```bash
sudo rm -- /etc/systemd/system/shortlink.service.d/91-lab-bad-user.conf
sudo systemctl daemon-reload
sudo systemctl reset-failed shortlink.service
sudo systemctl start shortlink.service
```

## 18. 更新二进制的正确最小流程

直接覆盖正在运行的二进制在 Linux 上可能可行，因为旧进程仍持有原 inode，但这不等于发布流程可靠。最低限度应：

1. 在独立路径构建并验证新二进制。
2. 使用 `install` 把权限和所有者一次设置正确。
3. 运行 unit 静态检查与二进制基本检查。
4. 重启服务。
5. 立刻检查 status、journal、端口和健康接口。
6. 失败时能够恢复上一版本。

本章只做手动更新：

```bash
go build -trimpath -o shortlink-lab .
sudo install -o root -g shortlink -m 0750 \
  ./shortlink-lab /opt/shortlink/bin/shortlink.new
sudo mv /opt/shortlink/bin/shortlink.new /opt/shortlink/bin/shortlink
sudo systemctl restart shortlink.service
systemctl status shortlink.service --no-pager -l
curl -fsS http://127.0.0.1:8080/healthz
```

完整发布会使用版本目录、原子符号链接切换、健康检查和自动回滚，见第 05 与第 09 章。不要把上面几条直接当成生产发布脚本。

## 19. 何时 reload，何时 restart

- **reload**：进程不退出，自己重新读取配置。只有应用实现对应机制且 unit 配置 `ExecReload=` 才能使用。
- **restart**：旧进程退出，新进程启动；代码更新一定需要这种生命周期切换。
- **daemon-reload**：systemd 自己重读 unit 文件，与业务配置 reload 完全不同。

不要随意把 SIGHUP 约定为配置 reload。Go 应用必须明确实现、测试并记录哪些配置可热更新。数据库 schema、监听地址、TLS 等配置通常更适合受控重启。

## 20. 本章检查点

完成本章后，应能独立回答和操作：

1. 程序、进程、线程和 goroutine 有什么区别？
2. 为什么 `kill` 默认不是强制杀死，什么时候才考虑 SIGKILL？
3. Go HTTP 服务收到 SIGTERM 后应该按什么顺序关闭？
4. 为什么后台 `nohup` 进程不能代替 systemd 服务？
5. `start`、`enable`、`daemon-reload`、`restart` 分别改变什么？
6. 为什么服务 `active (running)` 仍不能证明接口健康？
7. `status=203/EXEC`、`217/USER`、端口冲突分别先查什么？
8. 如何查询当前启动、最近十分钟和实时的 shortlink 日志？
9. 为什么应用不应以 root 运行，为什么二进制不应归应用用户可写？
10. 为什么手工 stop 通常不会被 `Restart=on-failure` 立即拉起？

能完成 Go 服务的前台运行、SIGTERM 验证、systemd 托管和三个故障实验，就已经具备后端部署最核心的进程与日志基础。下一章应继续学习网络、端口、DNS、防火墙和 curl，把“本机进程正常”扩展到“请求能够真正到达进程”。
