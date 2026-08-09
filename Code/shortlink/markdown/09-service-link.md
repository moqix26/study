# 09 · service/link.go 业务层精读

> 主线请跟 [`../study.md`](../study.md)；本文是精读加餐。

> 对应源码：`internal/service/link.go`（全文）  
> 目标：把 LinkService、创建、解析、缓存旁路、异步点击的设计讲透。

---

## 0. 这个文件在分层里的位置

```text
handler（HTTP 翻译）
    ↓ 调方法、认状态码
service（业务规则）  ← 本文
    ↓
repo（MySQL）  +  cache（Redis）
    ↑
pkg: urlx、shortcode
```

**原则**：handler 不写 SQL、不写 Redis key；service 编排「校验 → 生成 → 存 → 读 → 缓存策略」。

---

## 1. 完整源码

```go
package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"shortlink/internal/cache"
	"shortlink/internal/config"
	"shortlink/internal/model"
	"shortlink/internal/pkg/shortcode"
	"shortlink/internal/pkg/urlx"
	"shortlink/internal/repo"

	"gorm.io/gorm"
)

type LinkService struct {
	cfg   config.Config
	repo  *repo.LinkRepo
	cache *cache.LinkCache
}

func NewLinkService(cfg config.Config, r *repo.LinkRepo, c *cache.LinkCache) *LinkService {
	return &LinkService{cfg: cfg, repo: r, cache: c}
}

type CreateResult struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
}

func (s *LinkService) Create(rawURL string) (*CreateResult, error) {
	longURL, err := urlx.Normalize(rawURL)
	if err != nil {
		return nil, err
	}

	for i := 0; i < s.cfg.MaxRetries; i++ {
		code, err := shortcode.Random(s.cfg.CodeLength)
		if err != nil {
			return nil, err
		}
		link := &model.Link{Code: code, LongURL: longURL}
		err = s.repo.Create(link)
		if err == nil {
			return &CreateResult{
				Code:     link.Code,
				ShortURL: s.cfg.BaseURL + "/" + link.Code,
				LongURL:  link.LongURL,
			}, nil
		}
		if !repo.IsDuplicate(err) {
			return nil, err
		}
	}
	return nil, errors.New("failed to allocate code")
}

// Resolve 返回长链与是否缓存命中。
func (s *LinkService) Resolve(ctx context.Context, code string) (longURL string, cacheHit bool, err error) {
	if len(code) != s.cfg.CodeLength {
		return "", false, nil
	}

	if v, ok, err := s.cache.Get(ctx, code); err != nil {
		log.Println("redis get error:", err)
	} else if ok {
		return v, true, nil
	}

	link, err := s.repo.FindByCode(code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", false, nil
		}
		return "", false, err
	}

	if err := s.cache.Set(ctx, code, link.LongURL); err != nil {
		log.Println("redis set error:", err)
	}
	return link.LongURL, false, nil
}

// IncrClickAsync 异步增加点击（失败只打日志，不影响跳转）。
func (s *LinkService) IncrClickAsync(code string) {
	go func() {
		if err := s.repo.IncrClick(code); err != nil {
			log.Println("incr click error:", err)
		}
	}()
}

func (s *LinkService) ShortURL(code string) string {
	return fmt.Sprintf("%s/%s", s.cfg.BaseURL, code)
}
```

---

## 2. import 与依赖


| 包 | 用途 |
| --- | --- |
| `context` | `Resolve` 把 `ctx` 传给 Redis GET/SET（取消/超时） |
| `errors` / `gorm` | 判断 `ErrRecordNotFound`；创建失败 `failed to allocate code` |
| `fmt` | `ShortURL` 拼接 |
| `log` | Redis/点击失败打日志，**不向上抛**（降级） |
| `cache` / `repo` / `config` / `model` | 基础设施与领域模型 |
| `shortcode` / `urlx` | 纯函数工具 |

组装入口见 `internal/app/app.go`：

```go
svc := service.NewLinkService(cfg, linkRepo, linkCache)
```

---

## 3. `LinkService` 结构体

```go
type LinkService struct {
	cfg   config.Config
	repo  *repo.LinkRepo
	cache *cache.LinkCache
}
```

| 字段 | 类型 | 含义 |
| --- | --- | --- |
| `cfg` | `config.Config` | `BaseURL`、`CodeLength`、`MaxRetries` 等 |
| `repo` | `*repo.LinkRepo` | MySQL：Create / FindByCode / IncrClick |
| `cache` | `*cache.LinkCache` | Redis：`link:{code}` → longURL |

**为什么注入而不是全局变量？**  
测试可 mock repo/cache；`app.Run` 里一次构造，多 handler 共享同一 `svc`。

---

## 4. `CreateResult`

```go
type CreateResult struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
}
```

| 字段 | JSON | 来源 |
| --- | --- | --- |
| `Code` | `code` | 插入成功的 `link.Code` |
| `ShortURL` | `short_url` | `cfg.BaseURL + "/" + code` |
| `LongURL` | `long_url` | 规范化后的长链 |

handler 直接 `c.JSON(201, res)`，靠 struct tag 序列化。

**重要**：`Create` **成功后不写 Redis**（Cache Aside：写路径只落库，读路径第一次再回填）。见 §7。

---

## 5. `Create(rawURL string)` 创建流程

### 5.1 规范化

```go
longURL, err := urlx.Normalize(rawURL)
if err != nil {
	return nil, err
}
```

| 结果 | 行为 |
| --- | --- |
| 成功 | 得到可入库的 `longURL` |
| 失败 | `nil, err` → handler 映射 400 |

### 5.2 碰撞重试循环

```go
for i := 0; i < s.cfg.MaxRetries; i++ {
	code, err := shortcode.Random(s.cfg.CodeLength)
	...
	link := &model.Link{Code: code, LongURL: longURL}
	err = s.repo.Create(link)
	if err == nil {
		return &CreateResult{...}, nil
	}
	if !repo.IsDuplicate(err) {
		return nil, err
	}
}
return nil, errors.New("failed to allocate code")
```

| 步骤 | 含义 |
| --- | --- |
| `shortcode.Random` | 本轮候选短码 |
| `repo.Create` | `INSERT`；GORM 可能回填 `link.ID` |
| `err == nil` | 成功，立即返回，**不 SET Redis** |
| `repo.IsDuplicate(err)` | 唯一索引冲突 → `continue` 下一轮 |
| 非 duplicate 错误 | 连接断开、字段过长等 → 直接失败 |
| 循环耗尽 | `"failed to allocate code"` → handler 500 |

**配置**：`MaxRetries` 默认 8（环境变量 `SHORTLINK_MAX_RETRIES`）。

### 5.3 创建路径数据流

```text
rawURL
  → urlx.Normalize → longURL
  → [循环] shortcode.Random → code
  → repo.Create(Link{code, longURL})
  → CreateResult（无 cache.Set）
```

---

## 6. `Resolve(ctx, code)` — Cache Aside 读路径

签名：

```go
func (s *LinkService) Resolve(ctx context.Context, code string) (longURL string, cacheHit bool, err error)
```

| 返回值 | 含义 |
| --- | --- |
| `longURL` | 长链；**不存在时为 `""`** |
| `cacheHit` | `true` = Redis HIT；`false` = MISS 或没走缓存 |
| `err` | **仅基础设施错误**（如 MySQL 挂了）；「找不到」**不是 error** |

### 6.1 长度门禁

```go
if len(code) != s.cfg.CodeLength {
	return "", false, nil
}
```

| 例子 | 结果 |
| --- | --- |
| `code` 长度 6（默认） | 继续 |
| `health`、`api`、5 位、7 位 | `("", false, nil)` → handler 当 404 |

**注意**：`/:code` 路由会把 `health` 当 code 传进来；handler 另有黑名单，service 层用长度+查库双保险。

### 6.2 Redis GET（可降级）

```go
if v, ok, err := s.cache.Get(ctx, code); err != nil {
	log.Println("redis get error:", err)
} else if ok {
	return v, true, nil
}
```

| `cache.Get` 结果 | 行为 |
| --- | --- |
| `ok == true` | **HIT**：`(v, true, nil)`，不打 MySQL |
| `err != nil` | 打日志，**继续查 MySQL**（降级） |
| miss（`ok==false`, `err==nil`） | 继续 MySQL |

`cache.Get` 内部：`redis.Nil` → miss；其它 error → 向上；有值 → HIT。

### 6.3 MySQL 查询

```go
link, err := s.repo.FindByCode(code)
if err != nil {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", false, nil
	}
	return "", false, err
}
```

| 情况 | 返回 |
| --- | --- |
| 找到 | 继续回填缓存 |
| `ErrRecordNotFound` | `("", false, nil)` — **空串，非 error** |
| 其它 DB 错误 | `("", false, err)` → handler 500 |

**设计意图**：「没有这条短码」是正常业务结果，用 `longURL==""` 表达，handler 统一 404。

### 6.4 回填 Redis

```go
if err := s.cache.Set(ctx, code, link.LongURL); err != nil {
	log.Println("redis set error:", err)
}
return link.LongURL, false, nil
```

| 点 | 说明 |
| --- | --- |
| `cacheHit == false` | 本次是 MISS（即使 SET 成功，响应头仍标 MISS） |
| SET 失败 | 只打日志；仍返回 DB 里的长链 |
| TTL | `cfg.CacheTTL`，默认 1h（`cache` 包内 `SET` 带过期） |

### 6.5 Resolve 流程图

```text
len(code) != CodeLength? → ("", false, nil)
        ↓
Redis GET
  HIT → (longURL, true, nil)
  error → log，降级
  MISS → MySQL FindByCode
        NotFound → ("", false, nil)
        DB error → ("", false, err)
        Found → SET Redis（失败只 log）→ (longURL, false, nil)
```

---

## 7. 为什么创建不写缓存？

| 策略 | 本项目的做法 |
| --- | --- |
| Cache Aside 写 | 创建只 INSERT MySQL |
| 第一次读/跳转 | MISS → MySQL → SET Redis |
| 好处 | 实现简单；避免「写缓存成功但事务回滚」不一致；创建后未必马上访问 |

若创建时 SET，还要处理：创建失败回滚、更新/删除时失效缓存等——V1 故意不做。

---

## 8. `IncrClickAsync(code)`

```go
func (s *LinkService) IncrClickAsync(code string) {
	go func() {
		if err := s.repo.IncrClick(code); err != nil {
			log.Println("incr click error:", err)
		}
	}()
}
```

| 点 | 说明 |
| --- | --- |
| 调用方 | `handler.Redirect` 在确认 `longURL` 非空后、`302` 前调用 |
| `go func()` | 新 goroutine，**不阻塞**响应 |
| 失败 | 只 `log`；用户仍收到 302 |
| SQL | `UPDATE links SET click_count = click_count + 1 WHERE code = ?` |

**坑**：进程退出时 goroutine 可能没跑完；练习可接受。生产常用队列或带 context 的 worker。

**注意**：`GetLinkJSON` **不**调用 `IncrClickAsync`，只有浏览器跳转路径计数。

---

## 9. `ShortURL(code)`

```go
func (s *LinkService) ShortURL(code string) string {
	return fmt.Sprintf("%s/%s", s.cfg.BaseURL, code)
}
```

| 用途 | 调用处 |
| --- | --- |
| 拼对外短链 | `Create` 的 `ShortURL` 字段；`GetLinkJSON` 响应 |

与 `Create` 里 `s.cfg.BaseURL + "/" + link.Code` 等价；集中一处便于改格式。

---

## 10. 上下游对照


| 方法 | 上游（谁调） | 下游（谁被调） |
| --- | --- | --- |
| `Create` | `handler.CreateLink` | `urlx`, `shortcode`, `repo.Create` |
| `Resolve` | `handler.GetLinkJSON`, `handler.Redirect` | `cache.Get/Set`, `repo.FindByCode` |
| `IncrClickAsync` | `handler.Redirect` | `repo.IncrClick`（异步） |
| `ShortURL` | `handler.GetLinkJSON` | 无 IO |

---

## 11. 常见坑


| 坑 | 说明 |
| --- | --- |
| 把 `Resolve` 的 `""` 当 error | 应判 `longURL==""` → 404 |
| 以为创建后 Redis 必有 key | 第一次访问前 `GET link:code` 为 nil |
| Redis 挂了整站 500 | 本项目 GET 失败会降级 MySQL，仅日志 |
| `IncrClickAsync` 在 JSON 接口也加 | 会导致 API 查询也涨点击 |
| `ctx` 传 `nil` | 应用里用 `c.Request.Context()`；nil 可能导致 redis 库异常 |
| 改 `CodeLength` 不配齐 | `Resolve` 长度门禁与 DB 不一致会全 404 |

---

## 12. 验收

```powershell
# 创建
$r = Invoke-RestMethod -Uri http://localhost:8080/api/links -Method POST `
  -ContentType "application/json" -Body '{"url":"https://www.example.com"}'
$code = $r.code

# 创建后 Redis 应无缓存（或 key 不存在）
docker exec study-redis redis-cli GET "link:$code"
# 期望 (nil)

# 第一次 JSON 查询 → MISS
curl.exe -i "http://localhost:8080/api/links/$code"
# X-Cache: MISS

# 第二次 → HIT
curl.exe -i "http://localhost:8080/api/links/$code"
# X-Cache: HIT

# 跳转后 click_count（略延迟）
docker exec study-mysql mysql -uroot -proot123 -e "SELECT click_count FROM study.links WHERE code='$code';"
```

| 检查 | 期望 |
| --- | --- |
| 创建成功 | 201，三字段齐全 |
| 创建后 Redis | 无 `link:{code}` |
| 首次 Resolve | MISS + 有 long_url |
| 二次 Resolve | HIT |
| 跳转两次 | `click_count` ≥ 2（异步，可多等 1s） |
| 不存在 code | `longURL` 空 → 404，非 500 |

---

## 口述题

1. `Resolve` 返回三个值，「短码不存在」时各是什么？为什么不用 `error` 表示 not found？
2. Redis GET 失败时服务会怎样？还算 HIT 吗？
3. 创建成功后为什么不 `cache.Set`？第一次读时谁负责回填？
4. `IncrClickAsync` 为什么用 `go func`？失败会影响用户跳转吗？
5. `Create` 循环里 `repo.IsDuplicate` 为 false 时为什么要立刻返回，而不是继续重试？
