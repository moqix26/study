# 07 · 配置

`internal/config.Load()` 读环境变量，都有本地默认值。

| 变量 | 默认 | 含义 |
|------|------|------|
| `SHORTLINK_HTTP_ADDR` | `:8080` | 监听 |
| `SHORTLINK_BASE_URL` | `http://localhost:8080` | 拼 short_url |
| `SHORTLINK_MYSQL_DSN` | 见 example | MySQL |
| `SHORTLINK_REDIS_ADDR` | `127.0.0.1:6379` | Redis |
| `SHORTLINK_CACHE_TTL` | `1h` | 缓存过期 |
| `SHORTLINK_CODE_LEN` | `6` | 短码长度 |
| `SHORTLINK_MAX_RETRIES` | `8` | 碰撞重试 |

示例文件：[`../configs/config.example.env`](../configs/config.example.env)

PowerShell 临时覆盖：

```powershell
$env:SHORTLINK_REDIS_ADDR="127.0.0.1:6379"
go run .
```

**安全**：示例密码仅学习；公开仓库注意不要提交真实云数据库口令。
