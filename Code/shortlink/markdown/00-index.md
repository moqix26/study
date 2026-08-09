# markdown 目录索引

> **怎么用本目录**  
> 1. **主线只跟** [`../study.md`](../study.md)（聊天式一步一步推）。  
> 2. study 某一步说「精读打开 xxx」时，再进本目录。  
> 3. 每篇都按当前源码逐块讲解：代码 + 表格 + 为什么 + 坑 + 验收。
> 4. 不要把本目录当第二套主线通读；会重复、也容易晕。

---

## 推荐阅读顺序（跟 study 同步）

| study 步骤 | 先精读 | 可加餐 |
|------------|--------|--------|
| S0 开跑 | [01-cmd-server](./01-cmd-server.md) · [02-config](./02-config.md) | — |
| S1 业务长什么样 | [12-redirect-cache-flow](./12-redirect-cache-flow.md) 前半 | [03-app-wire](./03-app-wire.md) 路由段 |
| S2 创建短链 | [08-urlx](./08-urlx.md) · [07-shortcode](./07-shortcode.md) · [04-model-link](./04-model-link.md) · [05-repo-link](./05-repo-link.md) · [09-service-link](./09-service-link.md) 的 Create · [10-handler-http](./10-handler-http.md) 的 CreateLink | — |
| S3 跳转+缓存 | [06-cache-redis](./06-cache-redis.md) · [09](./09-service-link.md) 的 Resolve · [10](./10-handler-http.md) 的 Redirect/GetLinkJSON · [12](./12-redirect-cache-flow.md) | [11-middleware-logger](./11-middleware-logger.md) |
| S4 分层与配置 | [03-app-wire](./03-app-wire.md) · [02-config](./02-config.md) | 全目录按层扫一眼 |
| S5 验收口述 | [12](./12-redirect-cache-flow.md) 验收段 | [H-main-singlefile](./H-main-singlefile.md) 对照旧单文件 |

---

## 全文件一览

| 文件 | 精讲对象 | 密度目标 |
|------|----------|----------|
| [01-cmd-server.md](./01-cmd-server.md) | `cmd/server/main.go` | 入口级 |
| [02-config.md](./02-config.md) | `internal/config` | 配置/DSN |
| [03-app-wire.md](./03-app-wire.md) | `internal/app` 组装 | 接线 |
| [04-model-link.md](./04-model-link.md) | `internal/model` | 表结构 |
| [05-repo-link.md](./05-repo-link.md) | `internal/repo` | MySQL |
| [06-cache-redis.md](./06-cache-redis.md) | `internal/cache` | Redis |
| [07-shortcode.md](./07-shortcode.md) | `pkg/shortcode` | 随机短码 |
| [08-urlx.md](./08-urlx.md) | `pkg/urlx` | URL 校验 |
| [09-service-link.md](./09-service-link.md) | `internal/service` | 业务编排 |
| [10-handler-http.md](./10-handler-http.md) | `internal/handler` | HTTP |
| [11-middleware-logger.md](./11-middleware-logger.md) | `internal/middleware` | 日志中间件 |
| [12-redirect-cache-flow.md](./12-redirect-cache-flow.md) | 跨文件全链路 | 串联+验收 |
| [H-main-singlefile.md](./H-main-singlefile.md) | 历史单文件版 | 对照用 |

---

## 和旧结构的关系

以前的 `00-overview`～`09-faq` 短文已废弃，内容并入 **study 主线** 与上表精读篇。  
历史单文件精讲保留为 `H-main-singlefile.md`（原 `main.go.md`）。
