# Go 系统设计实验

这些程序只使用标准库，用来验证容易被口号掩盖的事实。它们不是完整业务项目。

## 实验列表

| 目录 | 验证内容 |
|------|----------|
| 01-context-retry | deadline、指数退避、jitter、有限重试 |
| 02-token-bucket | 令牌桶容量、补充速率和突发 |
| 03-shard-routing | 4 库 × 16 表的正确 64 槽路由 |
| 04-base62-math | Base62 空间、碰撞概率和存储单位 |
| 05-float64-precision | Snowflake ID 作为 ZSet score 的精度问题 |

## 运行

~~~powershell
Set-Location F:\study\后端学习\系统设计\Go实验
go test ./...
go run .\01-context-retry
go run .\02-token-bucket
go run .\03-shard-routing
go run .\04-base62-math
go run .\05-float64-precision
~~~

程序会在关键不变量不成立时 panic，因此“能运行”同时也是一个最小断言。

