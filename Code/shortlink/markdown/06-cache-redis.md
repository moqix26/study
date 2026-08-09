# internal/cache/redis.go 逐块精讲

> 主线请跟 [`../study.md`](../study.md)；本文是精读加餐。

> 对应源码：`internal/cache/redis.go`  
> 目标：搞清 key 前缀、`Get` 对 `redis.Nil` 的处理、`Set` TTL、`Del`、以及 **ctx 为何不能 nil**。

---

## 0. Cache 层在整体里的位置

```text
service.Resolve（Cache Aside 读路径）
    ├─ cache.Get   → 命中则直接返回 longURL
    ├─ repo.FindByCode（miss）
    └─ cache.Set   → 回填 Redis

service（未来若删链/改链）
    └─ cache.Del   → 失效缓存
```

Redis 里**只存长链字符串**（`longURL`），不存整个 `model.Link` JSON——读跳转只需 URL，值小、解析简单。

---

## 1. 完整源码

```go
package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type LinkCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewLinkCache(rdb *redis.Client, ttl time.Duration) *LinkCache {
	return &LinkCache{rdb: rdb, ttl: ttl}
}

func key(code string) string {
	return "link:" + code
}

func (c *LinkCache) Get(ctx context.Context, code string) (string, bool, error) {
	val, err := c.rdb.Get(ctx, key(code)).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if val == "" {
		return "", false, nil
	}
	return val, true, nil
}

func (c *LinkCache) Set(ctx context.Context, code, longURL string) error {
	return c.rdb.Set(ctx, key(code), longURL, c.ttl).Err()
}

func (c *LinkCache) Del(ctx context.Context, code string) error {
	return c.rdb.Del(ctx, key(code)).Err()
}
```

---

## 2. `import` 一览

| 包 | 用途 |
|----|------|
| `context` | 所有 Redis 命令第一个参数 `ctx` |
| `time` | `ttl time.Duration` |
| `github.com/redis/go-redis/v9` | 官方推荐 Go Redis 客户端（v9 起强制 context） |

---

## 3. `LinkCache` 结构体

```go
type LinkCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewLinkCache(rdb *redis.Client, ttl time.Duration) *LinkCache {
	return &LinkCache{rdb: rdb, ttl: ttl}
}
```

| 字段 | 含义 |
|------|------|
| `rdb` | 在 `app.Run` 里 `redis.NewClient` 创建，Ping 通过后注入 |
| `ttl` | 来自 `config.CacheTTL`（默认 `1h`）；每次 `Set` 用这个过期时间 |

**为何 ttl 放在 struct 而不是每次 `Set` 传参？**

- 全项目统一缓存策略；改 `SHORTLINK_CACHE_TTL` 一处生效。
- `Set` 签名更短；若以后按链接等级不同 TTL 再扩展不迟。

---

## 4. `key(code)` — Redis 键名

```go
func key(code string) string {
	return "link:" + code
}
```

| 部分 | 示例 | 含义 |
|------|------|------|
| 前缀 `link:` | 命名空间 | 避免与同一 Redis 里其它业务 key 冲突 |
| `code` | `BaLrEf` | 短码 |
| 完整 key | `link:BaLrEf` | `GET`/`SET`/`DEL` 用这个字符串 |

**为什么不用 `code` 裸作 key？**

- Redis 常见单库多业务；前缀便于 `KEYS link:*` 排查（生产慎用 KEYS，应用 SCAN）。
- 一眼看出 key 归属。

**小写私有函数 `key`：** 仅包内 `Get/Set/Del` 使用，外部不直接拼 key，避免散落魔法字符串。

---

## 5. `Get` — 读缓存（核心）

```go
func (c *LinkCache) Get(ctx context.Context, code string) (string, bool, error) {
	val, err := c.rdb.Get(ctx, key(code)).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if val == "" {
		return "", false, nil
	}
	return val, true, nil
}
```

### 5.1 返回值三元组

| 返回值 | 含义 |
|--------|------|
| `string` | 长链 URL；未命中或空值为 `""` |
| `bool` | **`true` = 缓存命中（HIT）**；`false` = miss 或当空处理 |
| `error` | 基础设施错误（网络、Redis 宕机）；**不是**「key 不存在」 |

与 `service.Resolve` 对接：

```go
if v, ok, err := s.cache.Get(ctx, code); err != nil {
	log.Println("redis get error:", err)  // 降级继续查 MySQL
} else if ok {
	return v, true, nil                   // HIT
}
```

### 5.2 `redis.Nil` 是什么？

| 情况 | `err` | 应如何处理 |
|------|-------|------------|
| key 不存在 | `redis.Nil` | **正常 miss**，不是故障 |
| key 存在 | `nil` | 看 `val` |
| 连接失败等 | 其它 error | 上报/日志，可降级 |

`go-redis` 用哨兵错误 `redis.Nil` 表示「GET 不到 key」，区别于 `err == nil`。

**因此：**

```go
if err == redis.Nil {
	return "", false, nil   // 第二个 false = miss，第三个 nil = 无基础设施错误
}
```

**初学者常犯错误：** 把 `redis.Nil` 当真正的异常返回上去 → 每次 miss 都 500。本项目在 service 里对非 nil 基础设施错误才打日志并降级。

### 5.3 空字符串 `val == ""`

```go
if val == "" {
	return "", false, nil
}
```

理论上正常 `Set` 不应存空 URL；若人为 `SET link:xxx ""`，当作 miss 更安全，避免跳转到空 Location。

---

## 6. `Set` — 写缓存 + TTL

```go
func (c *LinkCache) Set(ctx context.Context, code, longURL string) error {
	return c.rdb.Set(ctx, key(code), longURL, c.ttl).Err()
}
```

| 参数 | 含义 |
|------|------|
| `ctx` | 超时/取消传播 |
| `code` | 短码 |
| `longURL` | 要缓存的值 |
| `c.ttl` | 过期时间；`0` 表示永不过期（本项目默认非 0） |

Redis 命令概念：`SET link:BaLrEf "https://..." EX 3600`（若 ttl 为 1h）。

**调用时机：** `service.Resolve` 在 MySQL `FindByCode` 成功后：

```go
if err := s.cache.Set(ctx, code, link.LongURL); err != nil {
	log.Println("redis set error:", err)
}
return link.LongURL, false, nil  // MISS（刚从 DB 读出）
```

**Set 失败不阻断返回：** 下次读会再 miss、再查 DB——Cache Aside 的容错。

**创建短链时不写 Redis：** 只有**第一次被访问**才回填，减少无用 key。

---

## 7. `Del` — 删缓存

```go
func (c *LinkCache) Del(ctx context.Context, code string) error {
	return c.rdb.Del(ctx, key(code)).Err()
}
```

| 符号 | 含义 |
|------|------|
| `Del` | 删除 key；不存在也不报错（删除计数 0） |

**V1 跳转路径未调用 `Del`**；预留给「更新/删除长链时失效缓存」场景（写后删缓存模式）。

---

## 8. ctx 为何不能 nil？

`go-redis/v9` **要求**每个命令传入 `context.Context`：

```go
c.rdb.Get(ctx, key(code))
```

| 若传 `nil` | 后果 |
|------------|------|
| 运行时常 **panic** 或不可预测行为 | 库内部会调 `ctx.Done()` 等 |

**正确用法：**

| 场景 | 推荐 ctx |
|------|----------|
| HTTP 请求内 | `c.Request.Context()`（handler 传给 service） |
| 启动 Ping | `context.Background()` |
| 带超时探活 | `context.WithTimeout(parent, 3*time.Second)` |

本项目 handler：

```go
h.svc.Resolve(c.Request.Context(), code)
```

请求取消或客户端断开时，可中断正在进行的 Redis 调用（取决于调用点是否尊重 ctx）。

**`context.Background()`：** 非 nil 的空上下文，无 deadline，作为根。

---

## 9. Cache Aside 全流程（与 cache 包对应）

```text
Resolve(code)
  │
  ├─ cache.Get
  │     ├─ HIT  → 返回 (longURL, true, nil)
  │     └─ miss / redis.Nil
  │
  ├─ repo.FindByCode
  │     ├─ NotFound → ("", false, nil)
  │     └─ 有数据
  │
  ├─ cache.Set（忽略错误日志）
  └─ 返回 (longURL, false, nil)   // MISS
```

第二次访问同 code：`Get` HIT，不打 MySQL。

---

## 10. 与上下游怎么接

### 10.1 上游

| 组件 | 关系 |
|------|------|
| `app.Run` | `NewLinkCache(rdb, cfg.CacheTTL)` |
| `service.Resolve` | `Get` / `Set` |
| `handler` | 传 `Request.Context()` |

### 10.2 下游

| 组件 | 关系 |
|------|------|
| `redis.Client` | TCP 连 `cfg.RedisAddr` |
| Redis 进程 | 独立 Docker 容器 `study-redis` |

cache 包**不** import repo/model——单向依赖。

---

## 11. 常见坑

| 坑 | 现象 | 修法 |
|----|------|------|
| `ctx` 传 `nil` | panic | 用 `Background` 或 `Request.Context()` |
| 把 `redis.Nil` 当 500 | 每次查询都失败 | `Nil` → miss，返回 `false, nil` |
| 创建时就 `Set` | 大量从未访问的 key | V1 只在 Resolve 时回填 |
| 无 TTL 永久缓存 | 长链改了仍跳旧地址 | 必须 `Set` 带 `c.ttl`；改链时要 `Del` 或更新 |
| key 无前缀 | 与其它项目 key 冲突 | 保持 `link:` |
| Redis 挂了直接 500 | 服务不可用 | service 已对 Get 错误降级查 MySQL |
| 用 `KEYS *` 在生产查 key | 阻塞 Redis | 本地练习可用；生产用 SCAN |
| 缓存穿透（恶意不存在 code） | 每次都打 DB | V1 用 `len(code)!=CodeLength` 快速挡；以后可加布隆过滤器 |

---

## 12. 本地怎么验证

### 12.1 HIT / MISS 头

```powershell
# 创建
$r = Invoke-RestMethod -Method POST -Uri http://localhost:8080/api/links `
  -ContentType "application/json" -Body '{"url":"https://example.com/cache"}'
$code = $r.code

# 第一次 JSON 查询
curl.exe -i http://localhost:8080/api/links/$code
# X-Cache: MISS

# 第二次
curl.exe -i http://localhost:8080/api/links/$code
# X-Cache: HIT
```

### 12.2 Redis 里看 key 和 TTL

```powershell
docker exec -it study-redis redis-cli GET link:$code
docker exec -it study-redis redis-cli TTL link:$code
# TTL 应接近 3600（默认 1h）或你设的 SHORTLINK_CACHE_TTL
```

### 12.3 模拟 miss（删 key）

```powershell
docker exec -it study-redis redis-cli DEL link:$code
curl.exe -i http://localhost:8080/api/links/$code
# 再次 MISS，且终端/日志会打 MySQL 查询
```

### 12.4 验证 redis.Nil 路径

key 从未被访问过：

```powershell
docker exec -it study-redis redis-cli EXISTS link:neverrr
# 0

curl.exe -i http://localhost:8080/api/links/neverrr
# 404；不应 500（Resolve 长度校验或 DB not found）
```

### 12.5 停 Redis 验证降级（可选）

```powershell
docker stop study-redis
# 若服务仍在跑，Get 报错应打日志，仍可从 MySQL 返回（若之前未重启过进程）
# 新启动 go run ./cmd/server 会直接 redis Ping 失败退出——启动期与运行期策略不同
docker start study-redis
```

---

## 13. 与旧版单文件对照

| 旧 main | 现 cache |
|---------|----------|
| 全局 `rdb` + `linkKey` | `LinkCache` + 私有 `key()` |
| `loadLongURL` 内联 | `service.Resolve` + `cache.Get/Set` |
| `_ = rdb.Set` 忽略错误 | service 里 `log` Set 错误 |

语义一致，职责拆到 `internal/cache`。

---

## 14. 口述检查（2～3 题）

1. **`Get` 返回的三个值分别代表什么？`redis.Nil` 时三个值分别是什么？**  
   （期望：longURL, 是否HIT, 基础设施error；Nil 时 `"", false, nil`。）

2. **为什么 `Set` 必须带 TTL？创建短链时为什么不立刻 `Set`？**  
   （期望：防止脏数据永久存在；未访问的链接不占缓存，第一次读再回填。）

3. **go-redis v9 为什么 ctx 不能 nil？HTTP 请求里应该传哪个 ctx？**  
   （期望：库实现依赖 context；传 `c.Request.Context()` 支持取消与超时。）
