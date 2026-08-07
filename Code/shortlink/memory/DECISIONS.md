# DECISIONS

## 目录位置

- 批准路径：`F:\study\Code\shortlink`

## 模块

- 独立 `go.mod`（module `shortlink`），便于当简历独立仓库拷贝。
- 不再依赖上层 `Code/go.mod` 的 `go run ./shortlink`；改为在本目录 `go run ./cmd/server`。

## API（V1 最佳合理版）

| 方法 | 路径 | 作用 |
|------|------|------|
| GET | `/health` | 健康检查 |
| POST | `/api/links` | 创建短链 |
| GET | `/api/links/:code` | JSON 查映射 + X-Cache |
| GET | `/:code` | 302 跳转 + 异步访问计数 |

## 数据

- 表 `links`：code 唯一、long_url、click_count、created_at
- Redis key：`link:{code}` → 长 URL 字符串，TTL 默认 1h

## 旧单文件

- 原根目录 `main.go` 改为指向 `cmd/server` 的说明性入口，或删除后仅保留 cmd。
- 决策：根 `main.go` 保留为 **兼容薄封装**（调用与 cmd 相同的 bootstrap），方便习惯 `go run .`。

## 延后（写进 study.md，不做进 V1）

- JWT 用户体系、限流网关、分布式 ID、K8s、完整监控。
