# 06 · 跳转与缓存

## Cache Aside（旁路缓存）

```text
读：Redis → 没有 → MySQL → 写回 Redis
写：本 V1 创建短链不主动 SET；靠第一次跳转回填
```

真源永远是 MySQL。Redis 挂了：`Resolve` 打日志后仍查库（降级）。

## 302

```go
c.Redirect(http.StatusFound, longURL) // 302 + Location
```

浏览器跟 `Location` 走。用 `curl.exe -i` 才能清楚看到头，不要用会自动跟随的客户端误判。

## X-Cache

- `HIT`：Redis 命中  
- `MISS`：走了 MySQL（并尝试回填）

## JSON 查询

`GET /api/links/:code`：同样走 `Resolve`，但不跳转，方便验收缓存。

## 异步计数

`IncrClickAsync`：跳转成功后 `go repo.IncrClick`。用 MySQL 看 `click_count` 是否增加。

## 相关代码

- `service.Resolve` / `IncrClickAsync`  
- `cache.Get/Set`  
- `handler.Redirect` / `GetLinkJSON`
