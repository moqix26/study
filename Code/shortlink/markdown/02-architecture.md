# 02 · 分层架构

## 请求怎么走

### 创建短链

```text
HTTP POST /api/links
  → handler.CreateLink        解析 JSON
  → service.Create            校验 URL、生成短码、调 repo
  → repo.Create               INSERT
  → JSON 201 返回
```

### 跳转

```text
HTTP GET /:code
  → handler.Redirect
  → service.Resolve
       → cache.Get            HIT 则返回
       → repo.FindByCode      MISS 查库
       → cache.Set            回填
  → IncrClickAsync            后台 +1（不挡响应）
  → c.Redirect 302
```

## 每层职责（面试口述）

| 层 | 可以做 | 不要做 |
|----|--------|--------|
| handler | 状态码、Header、绑定 JSON | 直接写 SQL / Redis 细节 |
| service | 业务规则、重试、编排 | 依赖 `gin.Context` |
| repo | GORM | 拼 HTTP 响应 |
| cache | go-redis | 知道业务校验规则 |

## 对应代码

- 组装：`internal/app/app.go`
- HTTP：`internal/handler/http.go`
- 业务：`internal/service/link.go`

## 和单文件 main.go 的关系

以前全写在一个 `main.go` 里能跑，但难维护。现在逻辑等价，只是**搬了家**。旧精讲仍在 [main.go.md](./main.go.md) 可对照。
