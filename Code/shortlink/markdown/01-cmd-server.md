# `cmd/server` 入口逐块精讲

> 主线请跟 [`../study.md`](../study.md)；本文是入口的精读加餐。
>
> 当前项目只有一个服务入口：`cmd/server/main.go`。历史单文件架构请看
> [`H-main-singlefile.md`](./H-main-singlefile.md)，它不对应当前根目录源码。

---

## 0. 入口在整体架构里的位置

```text
cmd/server/main.go          进程入口，只负责启动和报告错误
        |
        v
internal/app/app.go         组装配置、MySQL、Redis、路由并运行 HTTP 服务
        |
        v
internal/config / repo / cache / service / handler / middleware
```

入口保持很薄，业务代码放在 `internal/`，这样启动方式和业务实现彼此独立。

## 1. 完整源码

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

## 2. 逐块理解

### `package main`

`main` 是 Go 可执行程序的入口包。`cmd/server` 是一个独立的命令目标，所以它可以被单独运行或构建。

### `shortlink/internal/app`

`go.mod` 声明了模块名 `shortlink`，因此导入路径使用 `shortlink/internal/app`，而不是文件系统相对路径。

### `app.Run()`

`Run` 返回 `error`，启动流程大致是：

```text
读取配置
  -> 连接 MySQL 并 AutoMigrate
  -> 连接 Redis
  -> 组装 repo/cache/service/handler
  -> 注册 Gin 路由
  -> 监听 HTTP 端口
```

入口不直接创建数据库连接，也不注册路由，只负责调用启动流程。

### `log.Fatal(err)`

启动依赖不可用时服务无法正常工作，因此打印错误并以非零状态退出。后续如果加入优雅停机，应把资源清理放到 `app` 的生命周期管理中，不要依赖 `log.Fatal` 执行 `defer`。

## 3. 为什么使用 `cmd/server`

`cmd/` 目录用于放可执行程序，每个子目录对应一个独立命令：

```text
cmd/
  server/main.go    -> HTTP 服务
  migrate/main.go   -> 未来可选的数据库迁移命令
  worker/main.go    -> 未来可选的异步消费者
```

当前只保留 `cmd/server`，避免同一个服务有两个入口、两套运行文档和两个构建目标。

## 4. 如何运行和构建

必须在包含 `go.mod` 的项目目录执行：

```powershell
cd F:\study\Code\shortlink
go run ./cmd/server
```

构建：

```powershell
go build -o server.exe ./cmd/server
```

成功启动时可以看到：

```text
mysql ok
redis ok
:8080 is on
```

另开终端检查：

```powershell
curl.exe http://localhost:8080/health
# 期望：{"status":"ok"}
```

## 5. 常见问题

| 现象 | 原因 / 修法 |
|------|-------------|
| `go: go.mod file not found` | 没有进入 `shortlink` 项目目录 |
| `mysql: ... connection refused` | MySQL 没启动，或 DSN 端口不是 3307 |
| `redis: ... connection refused` | Redis 没启动，或地址不是 6379 |
| `migrate: ...` | 数据库 `study` 不存在或连接配置错误 |
| 改 config.env 后仍监听 8080 | 配置只在启动时读取，需要重启服务 |

## 6. 与历史单文件代码的关系

以前的单文件版本把配置、数据库、缓存、路由和业务都写在一个 `main.go` 中。现在这些职责已经拆到 `internal/`，入口只保留 `app.Run()` 调用。

需要理解这次重构时，阅读 [`H-main-singlefile.md`](./H-main-singlefile.md) 做概念对照即可；不要把其中的旧命令或旧源码当作当前项目的运行方式。

## 7. 口述检查

1. 当前项目的唯一启动命令是什么？为什么放在 `cmd/server`？
2. `main.go` 为什么不直接连接 MySQL 或注册路由？
3. `app.Run()` 返回错误时，入口如何处理？
4. 如果未来增加迁移命令，应该放在哪里？
