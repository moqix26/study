# Redis 与 go-redis 缓存实战

<!-- 修改说明: 2026-07-08 按 EXPANSION-STANDARD 新建 §0、FAQ≥10、闭卷自测；理论交叉引用 Java/07；2026-07-14 补充可验证降级、TTL 抖动、singleflight、原子限流与故障测试 -->

> **文件编码**：UTF-8。  
> **定位**：Go 后端「缓存层」——`github.com/redis/go-redis/v9` 接 Redis，实现 Cache Aside 与三大经典问题对策。  
> **理论前置**：[Java 07 Redis 核心原理与缓存实战](../Java/07-Redis核心原理与缓存实战.md)（穿透/击穿/雪崩、数据结构、持久化在本章以交叉引用 + Go 代码为主）。  
> **代码前置**：[07 GORM 与 MySQL 实战](./07-GORM与MySQL实战.md)。

---

## 0. 读前导读（零基础也能跟上）

### 0.1 用一句话弄懂本章

**一句话**：把热点数据放 **Redis 抽屉**，读先 Redis 后 MySQL，写先 MySQL 再 **删** Redis——应用自己管缓存叫 **Cache Aside**。

**生活类比**（与 [Java 07](../Java/07-Redis核心原理与缓存实战.md) 一致）：

| 对比 | Redis | MySQL |
|------|-------|-------|
| 速度 | 微秒级 | 毫秒级 |
| 容量 | 小 | 大 |
| 断电 | 可丢（可 AOF） | 可靠落盘 |
| 场景 | 短链映射、Session、限流 | 订单、用户主数据 |

---

### 0.2 你需要提前知道什么

| 水平 | 建议 |
|------|------|
| 学完 07 GORM | 跟做 Cache Aside |
| 学过 Java 07 | 重点 go-redis API |
| 不懂穿透/击穿/雪崩 | 先读 Java 07 §9 |

---

### 0.3 本章知识地图（学完后应能勾选全部 ☐→☑）

- [ ] Docker 启动 Redis 7 并用 redis-cli 验证
- [ ] go-redis 连接池配置与 `Ping`
- [ ] 实现短链 `short_code → original_url` Cache Aside
- [ ] 写路径：UPDATE/INSERT MySQL 后 DEL 缓存
- [ ] 口述穿透/击穿/雪崩及一种 Go 侧对策
- [ ] Redis 超时/断连时受控降级 MySQL，而不是把缓存故障直接变成接口故障
- [ ] 用 TTL jitter + `singleflight` 抑制同时回源
- [ ] 用 Lua 实现原子限流并说清 fail-open/fail-closed 边界
- [ ] 用 SETNX 实现简易分布式锁（知道局限）
- [ ] 闭卷自测 ≥ 8/10

---

### 0.4 建议学习时长与节奏

| 阶段 | 时间 | 内容 |
|------|------|------|
| 环境 | 2 h | Docker + redis-cli |
| go-redis | 3 h | §2～§4 |
| Cache Aside | 3 h | §5 短链缓存 |
| 三大问题 | 2 h | §6 |
| 自测 | 1 h | FAQ + 闭卷 |

---

### 0.5 学完本章你能做什么

1. `redis-cli PING` → `PONG`；Go 程序 `client.Ping` 成功。
2. 第二次 `GET /{code}` 跳转路径命中 Redis（日志或断点）。
3. 更新长链后旧缓存失效，再查得到新 URL。
4. 白板画 Cache Aside 读/写路径（面试 3 分钟）。

---

### 0.6 redis-cli 手把手

| 步骤 | 命令 | 预期 |
|------|------|------|
| 1 | `docker run -d --name study-redis -p 6379:6379 redis:7` | Up |
| 2 | `docker exec -it study-redis redis-cli` | `127.0.0.1:6379>` |
| 3 | `SET link:abc https://example.com EX 3600` | OK |
| 4 | `GET link:abc` | URL 字符串 |
| 5 | `INCR stats:abc:clicks` | 整数自增 |

---

## 本章与上一章的关系

07 章短链查 MySQL；08 章在 Repository 之上加 **CacheService**，跳转读路径优先 Redis。

```mermaid
flowchart TD
    Req[GET /abc] --> C{Redis GET link:abc}
    C -->|hit| R302[302 跳转]
    C -->|miss| DB[(MySQL)]
    DB --> SET[SET link:abc EX TTL]
    SET --> R302
```

| 上一章（07） | 本章（08） | 下一章（09） |
|--------------|------------|--------------|
| MySQL 持久化 | Redis 热点缓存 | JWT 登录态 |
| GORM | go-redis/v9 | Token 存 Redis 可选 |

---

## 1. go-redis 连接

```go
import "github.com/redis/go-redis/v9"

func NewRedis(cfg Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           0,
		PoolSize:     20,
		MinIdleConns: 5,
		DialTimeout:  500 * time.Millisecond,
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
		PoolTimeout:  200 * time.Millisecond,
		MaxRetries:   1,
	})
}

func main() {
	rdb := NewRedis(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		// 本项目把 Redis 定义成可降级依赖：记录告警后仍可启动，读路径回源 MySQL。
		// 若你的业务把 Redis 当唯一 Session/队列存储，则应改成启动失败或 readiness=false。
		log.Printf("redis unavailable, start in degraded mode: %v", err)
	}
}
```

**v9 要点**：所有命令第一个参数 `context.Context`，支持超时与取消。客户端超时是防线，请求内仍应给单次缓存操作更短的 child context；否则 Redis 故障会先耗尽整个 HTTP deadline，根本来不及回源 MySQL。

---

## 2. Key 命名规范

```
link:{short_code}     → 原 URL 字符串
user:{id}             → 用户 JSON
lock:create_link:{uid}→ 分布式锁
rl:ip:{ip}            → 限流计数（11 章）
```

规则：业务前缀 + 冒号 + id。**缓存类 key 默认应有 TTL**；确实代表持久状态的集合、配置或幂等记录要单独设计生命周期，不能机械给所有 key 同一个过期时间。

---

## 3. Cache Aside 读路径

```go
const linkKeyPrefix = "link:"
const linkTTL = 24 * time.Hour
const notFoundValue = "\x00" // 合法 URL 不会是单个 NUL 字节

type LinkCache struct {
	rdb  *redis.Client
	repo *repository.LinkRepository
	load singleflight.Group
}

func (c *LinkCache) GetOriginalURL(ctx context.Context, code string) (string, error) {
	key := linkKeyPrefix + code
	cacheCtx, cancel := context.WithTimeout(ctx, 80*time.Millisecond)
	val, err := c.rdb.Get(cacheCtx, key).Result()
	cancel()
	if err == nil {
		if val == notFoundValue {
			return "", apperr.ErrNotFound
		}
		return val, nil // cache hit
	}
	fillCache := errors.Is(err, redis.Nil)
	if !fillCache {
		// Redis 是加速层，不是短链映射的 source of truth。
		// 记录 redis_degraded_total / 延迟并做采样日志，继续查 DB；不能在这里直接返回 500。
		log.Printf("redis get degraded key=%q: %v", key, err)
	}

	return c.loadFromDB(ctx, key, code, fillCache)
}

func (c *LinkCache) loadOneAndMaybeFill(ctx context.Context, key, code string, fillCache bool) (string, error) {
	link, err := c.repo.GetByShortCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("load link from mysql: %w", err)
	}
	now := time.Now()
	if link == nil || link.Status != model.LinkStatusActive ||
		(link.ExpiresAt != nil && !link.ExpiresAt.After(now)) {
		if fillCache {
			fillCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 80*time.Millisecond)
			defer cancel()
			if err := c.rdb.Set(fillCtx, key, notFoundValue, ttlWithJitter(3*time.Minute, time.Minute)).Err(); err != nil {
				log.Printf("redis negative fill degraded key=%q: %v", key, err)
			}
		}
		return "", apperr.ErrNotFound
	}

	if fillCache {
		cacheTTL := ttlWithJitter(linkTTL, 2*time.Hour)
		if link.ExpiresAt != nil {
			// 缓存绝不能比业务链接活得更久，否则 DB 已过期但 Redis 仍会继续跳转。
			remaining := link.ExpiresAt.Sub(now)
			if remaining < cacheTTL {
				cacheTTL = remaining
			}
		}
		fillCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 80*time.Millisecond)
		defer cancel()
		if err := c.rdb.Set(fillCtx, key, link.OriginalURL, cacheTTL).Err(); err != nil {
			log.Printf("redis fill degraded key=%q: %v", key, err)
		}
	}
	return link.OriginalURL, nil
}

func ttlWithJitter(base, jitter time.Duration) time.Duration {
	if jitter <= 0 {
		return base
	}
	return base + time.Duration(rand.Int63n(int64(jitter))) // 非安全随机，仅用于错峰
}
```

这里才是真正的“Redis 故障降级”：**缓存 miss 与缓存 error 分开观测，但二者都可进入受保护的 DB 回源**。若 MySQL 也失败，统一错误层返回 503；不能把“数据库故障”伪装成 404。

降级不是无限放行。Redis 全挂时所有请求都会压到 MySQL，因此还要配合 §6.3 `singleflight`、数据库连接池上限、短超时、熔断/本地应急限流和告警。目标是保住核心流量，不是承诺依赖全挂仍维持原 QPS。

---

## 4. 写路径：先 DB 后删缓存

```go
func (s *LinkService) UpdateURL(ctx context.Context, code, newURL string) error {
	if err := s.repo.UpdateOriginalURL(ctx, code, newURL); err != nil {
		return err
	}
	key := linkKeyPrefix + code
	invalidateCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 100*time.Millisecond)
	defer cancel()
	if err := s.rdb.Del(invalidateCtx, key).Err(); err != nil {
		// DB 已成功，不能简单告诉客户端“整个更新失败”并诱导盲目重试。
		// 记录指标/日志，并把失效任务写入可靠重试队列或 outbox。
		log.Printf("invalidate link cache code=%s: %v", code, err) // 项目中换成结构化日志
	}
	return nil
}
```

**为何 DEL 不是 SET？** 避免并发下旧值覆盖新值；与 [Java 07 Cache Aside](../Java/07-Redis核心原理与缓存实战.md) 一致。

### 4.1 Cache Aside 仍可能出现短暂旧值

典型竞态：

1. 请求 A 缓存 miss，读到 DB 旧值。
2. 请求 B 更新 DB，并删除缓存。
3. 请求 A 最后把旧值写回缓存。

对策按业务成本选择：

- 短链映射尽量设计为**创建后不可修改**，修改时生成新短码，直接消除最难的一致性路径。
- 回填缓存前携带版本号，Lua 比较版本，只允许新版本覆盖旧版本。
- 更新后做延迟二次删除，用于降低上述窗口，但它仍不是数学意义的强一致。
- 对一致性要求更高的场景使用消息/outbox 驱动失效，或直接绕过缓存读取主库。

缓存是副本，数据库才是 source of truth。要先明确业务能容忍多久的旧值，再选择方案。

### 4.2 哪些写操作必须失效缓存

| MySQL 操作 | 成功后动作 | 原因 |
|------------|------------|------|
| 创建短链 | `DEL link:{code}` | 清除之前可能缓存的“不存在”标记 |
| 修改 URL/过期时间 | `DEL link:{code}` | 下次按新数据回填 |
| 禁用/删除 | `DEL link:{code}` | 避免旧 URL 继续跳转 |
| 恢复启用 | `DEL link:{code}` | 清除禁用期间的空值缓存 |

`DEL` 失败时 DB 已提交，不能盲目回滚 HTTP 结果。项目版至少记录结构化日志和指标，并通过 outbox/可靠任务重试失效；重试操作天然幂等。测试必须覆盖“创建新短码前曾被缓存为空”的场景，否则用户会在数分钟内错误地看到 404。

```mermaid
sequenceDiagram
    participant App
    participant Redis
    participant MySQL

    Note over App,MySQL: 写路径
    App->>MySQL: UPDATE url
    App->>Redis: DEL link:code

    Note over App,MySQL: 读路径 miss
    App->>Redis: GET
    Redis-->>App: nil
    App->>MySQL: SELECT
    App->>Redis: SET + TTL
```

---

## 5. 序列化：String vs JSON

短链映射用 **纯 String**（值即 URL）最快。用户对象可用 JSON：

```go
data, err := json.Marshal(user)
if err != nil {
	return fmt.Errorf("marshal cached user: %w", err)
}
if err := rdb.Set(ctx, "user:1", data, time.Hour).Err(); err != nil {
	return fmt.Errorf("cache user: %w", err)
}
```

---

## 6. 穿透 / 击穿 / 雪崩

| 问题 | 含义 | Go 侧对策 |
|------|------|-----------|
| **穿透** | 查不存在 key，打穿 DB | 空值缓存短 TTL；布隆过滤器（[系统设计 08](../系统设计/08-短链服务设计.md)） |
| **击穿** | 热点 key 过期瞬间并发打 DB | 互斥锁 `SetNX` 只有一个回源；逻辑过期 |
| **雪崩** | 大量 key 同时过期 | TTL 加随机 jitter；多级缓存 |

### 6.1 空值防穿透

```go
if link == nil {
	if err := c.rdb.Set(ctx, key, notFoundValue, ttlWithJitter(3*time.Minute, time.Minute)).Err(); err != nil {
		log.Printf("cache not-found marker: %v", err) // 项目中改为采样日志 + 指标
	}
	return "", apperr.ErrNotFound
}
// 读时
if val == notFoundValue {
	return "", apperr.ErrNotFound
}
```

### 6.2 多实例必要时再做互斥回源（击穿）

普通短链先使用 §6.3 的 `singleflight`。若压测证明多实例同时回源仍会打爆数据库，再复用 §7 的 `WithLock`：拿锁后必须 **再次 GET 缓存**，仍 miss 才查 DB；未拿到锁的请求做有上限、带抖动的等待，不能递归无限重试。锁 value 必须是唯一 token，释放必须用 Lua 比较后删除，绝不能使用 `SETNX "1"` + 直接 `DEL` 的写法。

### 6.3 单机先用 `singleflight` 合并重复回源

同一个 Go 实例内，`golang.org/x/sync/singleflight` 可以让相同 key 的并发 miss 只执行一次查询：

```bash
go get golang.org/x/sync
```

```go
func (c *LinkCache) loadFromDB(ctx context.Context, key, code string, fillCache bool) (string, error) {
	resultCh := c.load.DoChan(code, func() (any, error) {
		// 不让“最先进入的请求取消”连带取消所有等待者；但共享回源必须有硬上限。
		loadCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 500*time.Millisecond)
		defer cancel()
		return c.loadOneAndMaybeFill(loadCtx, key, code, fillCache) // §3 的 DB 查询与可选回填
	})

	select {
	case result := <-resultCh:
		if result.Err != nil {
			return "", result.Err
		}
		return result.Val.(string), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
```

它只合并**当前进程**的请求，多实例仍可能同时回源。`DoChan` 让每个调用者可独立响应自己的取消；共享加载使用 `WithoutCancel + 硬超时`，避免首个调用者取消造成整组失败，也避免后台任务无限运行。很多业务先用 singleflight + 随机 TTL 就足够；不要一上来就为普通缓存 miss 引入复杂分布式锁。

---

## 7. 简易分布式锁

```go
var ErrLockBusy = errors.New("lock busy")

func WithLock(ctx context.Context, rdb *redis.Client, key string, ttl time.Duration, fn func() error) error {
	token := uuid.NewString()
	ok, err := rdb.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return fmt.Errorf("acquire redis lock: %w", err)
	}
	if !ok {
		return ErrLockBusy
	}
	defer func() {
		// Lua：只删自己的 token
		script := `if redis.call("get",KEYS[1])==ARGV[1] then return redis.call("del",KEYS[1]) else return 0 end`
		releaseCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := rdb.Eval(releaseCtx, script, []string{key}, token).Err(); err != nil {
			log.Printf("release redis lock key=%q: %v", key, err)
		}
	}()
	return fn()
}
```

面试至少说清以下边界：

1. `SET key token NX PX ttl` 必须一次完成；`SetNX(..., ttl)` 已表达该语义。
2. value 必须唯一，释放时 Lua“比较 token 后删除”，不能直接 `DEL` 别人的锁。
3. 临界区执行时间超过 TTL 时锁会提前失效；需要合理上限、续期，或让任务天然幂等。
4. Redis 故障转移和网络分区下，普通锁不能自动变成严格一致的分布式互斥。
5. 对扣款、库存等正确性敏感场景，应使用数据库约束/事务、幂等键或 fencing token，不能只靠“拿到 Redis 锁”。

库可以减少实现错误，但不能替你定义故障语义。

### 7.1 Pipeline、事务与 Lua 的区别

```go
pipe := rdb.Pipeline()
pipe.Incr(ctx, "stats:clicks")
pipe.Expire(ctx, "stats:clicks", 24*time.Hour)
_, err := pipe.Exec(ctx)
```

| 工具 | 核心能力 | 是否原子 |
|------|----------|----------|
| Pipeline | 批量发送，减少网络 RTT | 否，其他命令可穿插 |
| `TxPipeline` / MULTI EXEC | 一组命令顺序执行，中间不被其他客户端命令插入 | 是，但不支持读取结果后再决定后续命令 |
| Lua | 在 Redis 内执行“读 → 判断 → 写” | 脚本执行期间原子 |

不要把 Pipeline 当事务。释放锁、限流等依赖“检查后修改”的逻辑通常用 Lua。

### 7.2 Lua 原子限流：计数与首次过期必须一起做

若把 `INCR` 和 `EXPIRE` 分两次发送，应用可能在中间崩溃，留下永不过期的限流 key。固定窗口可先用 Lua 原子实现：

```go
var fixedWindow = redis.NewScript(`
local current = redis.call('INCR', KEYS[1])
if current == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
local ttl = redis.call('PTTL', KEYS[1])
if current > tonumber(ARGV[2]) then
  return {0, current, ttl}
end
return {1, current, ttl}
`)

func Allow(ctx context.Context, rdb *redis.Client, key string, window time.Duration, limit int64) (bool, error) {
	if window <= 0 || limit <= 0 {
		return false, apperr.ErrInvalidArgument
	}
	result, err := fixedWindow.Run(ctx, rdb, []string{key}, window.Milliseconds(), limit).Int64Slice()
	if err != nil {
		return false, fmt.Errorf("run rate limiter: %w", err)
	}
	if len(result) != 3 {
		return false, fmt.Errorf("unexpected limiter result length: %d", len(result))
	}
	return result[0] == 1, nil
}
```

要说清四个边界：

1. 固定窗口在边界处允许瞬时双倍流量；更平滑可用滑动窗口或 token bucket。
2. Lua 只保证 **Redis 内脚本** 原子，不能把“限流成功 + MySQL 写入”变成跨系统事务；限流只是准入控制。
3. Redis Cluster 多 key 脚本要求 key 在同一 slot，可用 `{userID}` hash tag；上例只有一个 key。
4. Redis 故障策略按接口选：公开跳转通常 **fail-open + 本地应急限流 + DB 保护**；登录/发短信等安全接口可 **fail-closed** 或使用更严格本地后备。策略必须有指标和测试，不能在代码里默默决定。

### 7.3 故障测试：证明“能降级”，不能只画图

| 场景 | 操作 | 应断言 |
|------|------|--------|
| Redis 断开 | `docker stop study-redis` 或 Toxiproxy reset | 缓存操作在短超时内失败，命中 MySQL 后仍 302；`redis_degraded_total` 增加 |
| Redis 高延迟 | Toxiproxy 注入 300ms latency | 请求不会等待完整 300ms，80ms 左右转 DB |
| Redis 断开 + MySQL miss | 请求不存在短码 | 返回 404，不 panic |
| Redis 断开 + MySQL 故障 | 同时停止两者 | 返回 503，DB 连接池不被无限请求打满 |
| 并发热点 miss | 100 goroutine 请求同一 code | 单实例 DB 查询接近 1 次，所有结果一致 |
| 写后失效失败 | 让 `DEL` 报错 | DB 更新仍成功，重试任务最终删除旧缓存 |
| TTL 抖动 | 生成大量 TTL | 都落在 `[base, base+jitter)`，不是同一秒过期 |

`miniredis` 适合快速测命中、miss、TTL 和基本命令；网络超时、连接池、故障恢复必须用真实 Redis + Toxiproxy/容器故障集成测试。

---

## 8. 常见错误对照表

| 现象 | 原因 | 处理 |
|------|------|------|
| `context deadline exceeded` | 无超时 | `context.WithTimeout` |
| 缓存不一致 | 先删缓存后写 DB | 先 DB 后 DEL |
| 内存暴涨 | 缓存 key 无生命周期 | 缓存 key 默认 TTL；持久状态单独设计 |
| `MOVED` 错误 | Cluster 模式 | 用 ClusterClient |
| 热 key | 单 key QPS 过高 | 本地缓存 + Redis；11 章 CDN |
| Redis 挂后接口全 500 | 把非 `redis.Nil` 直接返回 | 短超时记录降级并受控回源 DB |
| 大量 key 同秒失效 | 固定 TTL 无抖动 | `baseTTL + jitter` |

---

## 9. 与 Gin 集成

```go
type LinkHandler struct {
	cache *service.LinkCache
}

func (h *LinkHandler) Redirect(c *gin.Context) {
	code := c.Param("code")
	url, err := h.cache.GetOriginalURL(c.Request.Context(), code)
	if err != nil {
		response.WriteError(c, err) // not found→404；MySQL/依赖故障→503/500
		return
	}
	c.Redirect(http.StatusFound, url) // 302，11 章详述
}
```

---

## 10. FAQ

**Q1：go-redis v8 和 v9？**  
新项目 **v9**，Context 一等公民。

**Q2：Redis 和 Memcached？**  
Redis 数据结构丰富、持久化；Go 后端几乎全 Redis。

**Q3：缓存和 DB 强一致？**  
Cache Aside 是 **最终一致**；强一致用分布式事务（过重，短链不需要）。

**Q4：TTL 设多长？**  
短链映射可从 24h～7d + jitter 起步，但 TTL 必须取 `min(缓存策略 TTL, 距业务 expires_at 的剩余时间)`；无业务过期的热点才可考虑长期缓存 + 主动失效。

**Q5：Pipeline 有什么用？**  
批量命令减 RTT；统计写入可 Pipeline。

**Q6：Redis 单线程为何快？**  
内存 + IO 多路复用；无磁盘随机读。

**Q7：持久化 RDB/AOF？**  
实习认知即可；详见 [Java 07](../Java/07-Redis核心原理与缓存实战.md)。

**Q8：Token 放 Redis？**  
09 章 JWT 可无状态；黑名单/Refresh 可用 Redis。

**Q9：布隆过滤器 Go 库？**  
`bits-and-blooms/bloom/v3`；设计见 [系统设计 08](../系统设计/08-短链服务设计.md)。

**Q10：缓存 null 占内存吗？**  
占，但远小于 DB 压力；TTL 要短。

**Q11：Connection PoolSize？**  
默认 10×GOMAXPROCS；压测调。

**Q12：Redis 挂了怎么办？**  
缓存操作短超时后降级 MySQL，并用 singleflight、连接池、熔断/本地限流保护 DB；用故障测试验证，而不是无限制直查。

**Q13：Redis 故障时所有接口都应该 fail-open 吗？**
不是。公开读路径偏可用性可 fail-open；登录、验证码等安全接口常选 fail-closed 或严格后备策略。

**Q14：singleflight 能替代 Redis 锁吗？**
它只合并一个进程内相同 key 的调用；多实例仍会各回源一次，但通常已足够且故障面更小。

---

## 11. 练习建议

### 基础

1. 实现 `LinkCache.GetOriginalURL` 完整 Cache Aside
2. redis-cli 观察第二次 GET 命中

### 进阶

3. 更新 URL 后验证缓存失效
4. 空值缓存防穿透 demo

### 挑战

5. SETNX 互斥回源 + 单元测试（miniredis）
6. 对照 Java 07 写同流程时序图
7. 用 Lua 完成“每 IP 每分钟 60 次”并发测试，验证不会出现无 TTL key
8. 用 Toxiproxy 注入 Redis 延迟/断流，记录降级延迟与 DB 查询次数

---

## 12. 学完标准

- [ ] Docker Redis + go-redis Ping
- [ ] Cache Aside 读写正确
- [ ] 能解释穿透/击穿/雪崩
- [ ] key 命名 + TTL 规范
- [ ] Redis miss/error 语义分开，故障时受控回源且 MySQL 故障不伪装成 404
- [ ] TTL jitter、singleflight、缓存失效重试均有测试
- [ ] Lua 限流原子并能解释 fail-open/fail-closed
- [ ] 知道与 Java 07 理论对应关系

---

## 13. 闭卷自测

1. Cache Aside 读路径三步？
2. 写路径为何 DEL 不是 SET？
3. `redis.Nil` 表示什么？
4. 穿透与击穿区别？
5. 雪崩 TTL jitter 怎么加？（口述）
6. go-redis v9 命令签名特点？
7. 短链 key 如何命名？
8. SETNX 锁为何要唯一 token？
9. Redis 挂了短链怎么办？
10. 布隆过滤器解决什么问题？

### 参考答案

1. GET Redis → miss 查 MySQL → SET TTL。
2. 避免并发写乱序；以 DB 为准。
3. key 不存在（非网络错误）。
4. 穿透不存在；击穿热点过期。
5. `baseTTL + rand(0, 300s)`。
6. 第一个参数 `context.Context`。
7. `link:{short_code}`。
8. 防误删他人锁。
9. 降级 MySQL + 限流/熔断。
10. 快速判「一定不存在」，减穿透。

---

## 14. 费曼检验

3 分钟：**「抽屉（Redis）和仓库（MySQL）怎么配合？」**

用 Cache Aside 读/写时序 + 一个穿透例子；细节引用 [Java 07](../Java/07-Redis核心原理与缓存实战.md)。

---

## 15. 章节衔接

| 模块 | 链接 |
|------|------|
| 上一章 GORM | [07 GORM 与 MySQL](./07-GORM与MySQL实战.md) |
| Redis 理论 | [Java 07 Redis](../Java/07-Redis核心原理与缓存实战.md) |
| 下一章认证 | [09 JWT 认证与用户体系](./09-JWT认证与用户体系.md) |
| 短链设计 | [系统设计 08](../系统设计/08-短链服务设计.md) |
| 限流 | [系统设计 02 限流](../系统设计/02-限流熔断与降级.md) |

**下一章预告**：08 章解决了「读快」；09 章解决「谁有权创建短链」——**bcrypt + JWT + 鉴权中间件**。

---

*下一章：[09-JWT认证与用户体系](./09-JWT认证与用户体系.md)*
