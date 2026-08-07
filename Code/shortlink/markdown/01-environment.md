# 01 · 环境与依赖

## Docker

```powershell
docker start study-mysql
docker start study-redis
docker ps
```

期望：

- MySQL：`0.0.0.0:3307->3306/tcp`（本机若 3306 被占，练习用 3307）
- Redis：`0.0.0.0:6379->6379/tcp`

探活：

```powershell
docker exec study-mysql mysql -uroot -proot123 -e "SELECT 1"
docker exec study-redis redis-cli PING
```

## Go 模块

本项目**自有** `go.mod`（module `shortlink`）：

```powershell
cd F:\study\Code\shortlink
go mod tidy
go run .
```

不要在 `F:\study\Code` 根目录用旧的 `go run ./shortlink` 指望自动找到子模块依赖（有独立 go.mod 后应进入本目录运行）。

## 默认连接串（仅本地练习）

| 项 | 默认值 |
|----|--------|
| MySQL DSN | `root:root123@tcp(127.0.0.1:3307)/study?...` |
| Redis | `127.0.0.1:6379` |
| HTTP | `:8080` |

覆盖方式见 [07-config.md](./07-config.md)。**生产勿用示例密码。**

## 启动成功标志

```text
mysql ok
redis ok
:8080 is on
```

若 Redis panic：`ctx` 必须非 nil；Addr 必须是 **6379** 不是 8379（你之前踩过）。
