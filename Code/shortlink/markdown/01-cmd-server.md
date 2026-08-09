# 入口 main.go 与 cmd/server/main.go 逐块精讲

> 主线请跟 [`../study.md`](../study.md)；本文是精读加餐。

> 对应源码：根目录 `main.go`、`cmd/server/main.go`  
> 目标：搞清「程序从哪启动」「为什么有两个入口」「`go run` 怎么选」。

---

## 0. 这两份文件在整体架构里的位置

分层之后，**可执行程序的入口**只剩薄薄一层 `main` 包；真正的启动逻辑在 `internal/app/app.go` 的 `Run()` 里。

```text
main.go / cmd/server/main.go   ← 你正在读的这一层（入口）
        ↓ 调用
internal/app/app.go            ← 组装依赖、挂路由、监听端口
        ↓ 调用
config / repo / cache / service / handler / middleware
```

**为什么拆成两层？**

- `main` 包只做一件事：把进程拉起来，出错就退出。
- `app` 包可被测试、可被别的入口复用（例如以后加 `cmd/migrate` 只跑迁移，不必复制粘贴启动代码）。

---

## 1. 根目录 `main.go`（完整源码）

```go
// 兼容入口：在 shortlink 目录执行 `go run .` 即可。
// 规范入口同样是：go run ./cmd/server
package main

import (
	"log"

	"shortlink/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
```

---

## 2. `cmd/server/main.go`（完整源码）

```go
package main

import (
	"log"

	"shortlink/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
```

**和根目录 `main.go` 的区别：** 只有根目录多了两行注释；`import` 与 `main()` 函数体**完全相同**。

---

## 3. 逐块拆解

### 3.1 注释（仅根目录有）

```go
// 兼容入口：在 shortlink 目录执行 `go run .` 即可。
// 规范入口同样是：go run ./cmd/server
```

| 片段 | 含义 |
|------|------|
| `兼容入口` | 老习惯：在项目根目录直接 `go run .`，不必记子路径 |
| `规范入口` | Go 社区惯例：可执行文件放在 `cmd/<名字>/main.go` |

两行注释是给人看的，编译器会忽略。

### 3.2 `package main`

```go
package main
```

| 符号 | 含义 |
|------|------|
| `package` | 声明这个文件属于哪个包 |
| `main` | **可执行程序专用包名**；`go build` 时若包名是 `main` 且含 `func main()`，会生成二进制 |

同一模块里可以有多个 `package main`，但**每个 `main` 包只能有一个 `func main()`**，且通常各自对应一个可执行目标（根目录一个、`cmd/server` 一个）。

### 3.3 `import` 块

```go
import (
	"log"

	"shortlink/internal/app"
)
```

| 导入路径 | 类型 | 本文件里干什么 |
|----------|------|----------------|
| `"log"` | 标准库 | 启动失败时 `log.Fatal` 打印错误并 `os.Exit(1)` |
| `"shortlink/internal/app"` | 本项目包 | 调用 `app.Run()` 启动整个 HTTP 服务 |

**为什么 `shortlink/...` 而不是 `./internal/app`？**

- `go.mod` 第一行写了 `module shortlink`，这是**模块名**（import 前缀）。
- Go 用模块名 + 相对模块根的路径来定位包，与当前文件在磁盘上的相对位置无关。

```go
// go.mod
module shortlink
```

因此无论从根 `main.go` 还是 `cmd/server/main.go` 引用，都写 `shortlink/internal/app`。

### 3.4 `func main()`

```go
func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
```

| 符号/调用 | 含义 |
|-----------|------|
| `func main()` | 进程入口；无参数、无返回值（Go 规定） |
| `app.Run()` | 返回 `error`：成功 `nil`，失败带原因 |
| `err := ...` | 短变量声明，只在 `if` 块内可见 |
| `err != nil` | Go 惯用错误检查 |
| `log.Fatal(err)` | 打印 `err` 到 stderr，然后 `os.Exit(1)` |

**为什么用 `log.Fatal` 而不是 `panic`？**

- `app.Run()` 已经把错误包装成普通 `error` 返回（如 `mysql: ...`、`redis: ...`）。
- `log.Fatal` 适合「启动阶段失败、无法继续服务」：日志清晰、退出码非 0，方便脚本/CI 判断。
- `panic` 更适合「绝不应该发生的编程错误」；启动失败用 `error` 返回更地道。

**`app.Run()` 里大致做了什么（下游，详见 `03-app-wire.md`）：**

```text
Load 配置 → 连 MySQL → AutoMigrate → 连 Redis → 组装 service/handler
→ gin 路由 → r.Run(HTTPAddr)
```

入口层**不关心**这些细节，只关心「成没成」。

---

## 4. 为何双入口？

| 入口 | 路径 | 典型命令 |
|------|------|----------|
| 兼容入口 | `main.go`（模块根） | `go run .` |
| 规范入口 | `cmd/server/main.go` | `go run ./cmd/server` |

**历史原因：** 项目从「根目录单文件 `main.go`」演进到分层后，保留根入口让你 `cd shortlink` 后直接 `go run .`，不用改 muscle memory。

**工程原因：** `cmd/` 目录可以挂多个二进制：

```text
cmd/
  server/main.go    → HTTP 服务
  migrate/main.go   → （未来）只跑数据库迁移
  worker/main.go    → （未来）异步消费者
```

每个子目录一个 `main`，职责清晰；根 `main.go` 只是对 `cmd/server` 的**别名式兼容**。

**两者会冲突吗？**

- 不会同时编译进一个二进制。你 `go run .` 只编译根 `main.go`；`go run ./cmd/server` 只编译 `cmd/server/main.go`。
- `go build` 不带路径时，默认编译**当前目录**的 `main` 包（根目录），产物如 `shortlink.exe`。
- `go build -o server.exe ./cmd/server` 则编译规范入口。

---

## 5. `go run .` vs `go run ./cmd/server`

| 命令 | 编译谁 | 工作目录建议 | 行为 |
|------|--------|--------------|------|
| `go run .` | 根 `main.go` | `F:\study\Code\shortlink` | 与下等价（源码相同） |
| `go run ./cmd/server` | `cmd/server/main.go` | 同上 | 社区更常见的写法 |
| `go run ./cmd/server/main.go` | 同上 | 同上 | 显式到文件，效果一样 |

**注意：**

- 必须在**模块根**（有 `go.mod` 的目录）执行，否则找不到 `module shortlink`。
- 两者都调用同一个 `app.Run()`，运行时行为**完全一致**。
- 环境变量、`configs/config.example.env` 的加载方式与用哪个入口**无关**（配置在 `config.Load()` 里读环境变量）。

---

## 6. 与上下游怎么接

### 6.1 上游（谁调用入口）

| 调用方 | 场景 |
|--------|------|
| 你本地 | PowerShell 里 `go run .` |
| IDE | Run/Debug 配置指向 `main.go` 或 `cmd/server` |
| Docker / systemd | `ENTRYPOINT` 或 `ExecStart` 跑编译好的二进制 |
| CI | `go build -o bin/server ./cmd/server` 再跑测试 |

入口**不**被其他 Go 包 import（`package main` 不可被引用）。

### 6.2 下游（入口调用谁）

```text
main()
  └─ app.Run()          internal/app/app.go
       ├─ config.Load() internal/config/config.go
       ├─ repo / cache / service / handler
       └─ gin + r.Run()
```

下一关精读：[`02-config.md`](./02-config.md)（配置从哪来）、[`03-app-wire.md`](./03-app-wire.md)（`Run()` 全貌）。

---

## 7. 常见坑

| 坑 | 现象 | 原因 / 修法 |
|----|------|-------------|
| 在错误目录 `go run` | `go: go.mod file not found` | `cd` 到 `shortlink`（含 `go.mod`） |
| import 写成相对路径 | 编译错误 | 必须 `shortlink/internal/app`，不能 `./internal/app` |
| 改 `app.Run` 逻辑却只改了一个 main | 以为没生效 | 两个 main 相同，改 `app` 即可；不必改两处 |
| `log.Fatal` 后还想 defer 清理 | defer 不执行 | `Fatal` 直接 `Exit`；若需清理用 `log.Print` + `return` 或 `os.Exit` 前手动清理 |
| 根目录和 `cmd/server` 写出不同逻辑 | 行为不一致、难维护 | **保持同步**；业务只放 `internal/app` |
| `go build` 产出两个 exe 搞混 | 不知道跑哪个 | 约定：开发 `go run .`；发布 `go build -o server ./cmd/server` |

---

## 8. 本地怎么验证

```powershell
cd F:\study\Code\shortlink
docker start study-mysql
docker start study-redis

# 方式 A：兼容入口
go run .

# 方式 B：规范入口（另开一个终端对比）
go run ./cmd/server
```

**验收标准（两种命令应一致）：**

```text
mysql ok
redis ok
:8080 is on
```

另开终端：

```powershell
curl.exe http://localhost:8080/health
# 期望：{"status":"ok"}
```

**故意测失败路径（可选）：**

```powershell
# 停掉 Redis 后再 go run .
docker stop study-redis
go run .
# 期望：redis: ... 之类错误，log.Fatal 退出，进程非 0
docker start study-redis
```

**对比两个入口编译产物（可选）：**

```powershell
go build -o root.exe .
go build -o server.exe ./cmd/server
# 两者都应能 ./root.exe 或 ./server.exe 启动服务
```

---

## 9. 和旧版单文件 main 的对照

| 旧（`markdown/main.go.md` 时代） | 现（分层） |
|----------------------------------|------------|
| 根目录几千行 `main.go` | 根目录十几行 + `internal/*` |
| `main()` 里连 DB、挂路由 | `main()` 只调 `app.Run()` |
| 难单测、难复用 | `app`、`service`、`repo` 可单测 |

逻辑等价：现网功能 = 旧单文件 + 配置化 + `click_count` 等增强。

---

## 10. 口述检查（2～3 题）

1. **为什么项目里同时有根 `main.go` 和 `cmd/server/main.go`？实际业务代码应该写在哪？**  
   （期望：双入口是兼容 vs 规范；业务在 `internal/`，入口只调 `app.Run`。）

2. **`import "shortlink/internal/app"` 里的 `shortlink` 从哪来？在 `cmd/server` 里为什么不用 `../internal/app`？**  
   （期望：`go.mod` 的 `module` 名；Go 模块用模块路径 import，不用文件系统相对路径。）

3. **`app.Run()` 返回 error 时，`log.Fatal` 和 `panic` 你会选哪个？为什么？**  
   （期望：启动失败用 `error` + `log.Fatal`；可预期、退出码明确；`panic` 留给不可恢复的内部错误。）
