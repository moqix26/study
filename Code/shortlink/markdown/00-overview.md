# 00 · 项目总览

## 一句话

把很长的 URL 变成短码；访问短码时 **302** 跳回长链。热点短码用 **Redis** 加速，真源在 **MySQL**。

## 和你已学过的对照

| gin-redis（用户） | shortlink（链接） |
|-------------------|-------------------|
| `User` 表 | `Link` 表 |
| `GET /api/users/:id` 返回 JSON | `GET /:code` **302 跳转** |
| Redis key `user:{id}` | Redis key `link:{code}` |
| PUT 后 DEL 缓存 | 本 V1 创建后不预写缓存，读时回填 |

你已经会的：`ShouldBindJSON`、GORM `Create/First`、`rdb.Get/Set`、`X-Cache` —— 短链是**同一套肌肉**换业务。

## 目录树（规范版）

```text
cmd/server/main.go          → 进程入口
internal/app/app.go         → 组装 MySQL/Redis/路由并 Run
internal/handler/           → 只懂 HTTP
internal/service/           → 业务：建链、解析、异步计数
internal/repo/              → 只懂 MySQL
internal/cache/             → 只懂 Redis
internal/model/             → 结构体 = 表
internal/config/            → 环境变量
internal/pkg/shortcode|urlx → 纯函数工具
```

**为什么分层？** 面试能讲清「请求怎么走进来」；以后加限流/JWT 只改 handler/service，不把 SQL 糊在路由里。

## API 一览

见 [`../README.md`](../README.md) 与 [`08-acceptance.md`](./08-acceptance.md)。

## 下一步读什么

环境：[01-environment.md](./01-environment.md)  
分层：[02-architecture.md](./02-architecture.md)
