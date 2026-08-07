# 09 · FAQ 与踩坑

## 1. nil pointer / Ping 崩

- `context` 必须是 `context.Background()`，不能是 nil  
- 见你学 Redis 时的教训

## 2. connection refused :8379

端口写错。Redis 默认 **6379**。

## 3. PUT /ai/users 那种笔误

路由字符串写错 → 404。短链里注意 `/:code` 不要抢掉 `/api`（handler 里对 `api`/`health` 做了防护）。

## 4. PowerShell 里 curl

用 `curl.exe`。`Invoke-RestMethod` 对 302 会自动跟随，不好观察 `X-Cache`。

## 5. 中文显示 ??

控制台编码问题；库里 utf8mb4 一般正常。用 `docker exec ... mysql` 查。

## 6. 独立 go.mod 后怎么跑

必须在 `F:\study\Code\shortlink` 下 `go run .`，不要混用上层 module 路径习惯。

## 7. 缓存穿透（概念，V1 未专门防）

大量查不存在的 code → 每次打 DB。以后可对「空结果」缓存短 TTL。详见 study.md 加餐节。

## 8. gin.Default 重复中间件

规范版用 `gin.New()` + Recovery + 自写 Logger，避免双日志。
