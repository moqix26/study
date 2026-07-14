# 11 Go 端到端 Agent 项目：从内存 V1 到可演进服务

```yaml
project: Go工程/agentgo
language: Go
last_reviewed: 2026-07-15
truth_policy: 只有命令实际通过，才把能力记为已完成；V2/V3 不包装成现成功能
```

## 1. 项目定位

`agentgo` 是本路线的可运行练习工程。

它不是“复制即生产”的脚手架，也不声称有未经测试的性能指标。

项目分三阶段：

| 阶段 | 状态定义 | 存储与依赖 | 目标 |
|---|---|---|---|
| V1 | 已在 2026-07-15 通过普通与 race 测试，可无 Key 运行 | mock LLM、内存 memory、内存 RAG | 跑通 HTTP、SSE、Tool、Agent loop、测试 |
| V2 | pgvector/Responses adapter 已有代码和单测；外部集成需容器或凭据另验 | PostgreSQL + pgvector、可选 OpenAI 文本 provider；当前 embedding 仍为 Hash | 学习持久化、迁移、检索评估 |
| V3 | 尚未完成的生产增强清单 | 真实认证、限流、队列、观测、评估、安全 | 按实际需求逐项实现和验证 |

“V2/V3”表示演进路线，不表示仓库已经完成生产部署。

## 2. V1 验收标准

在 `Go工程/agentgo` 下运行：

```powershell
go test ./...
go run ./cmd/server
```

默认：

```text
AI_PROVIDER=mock
RAG_STORE=memory
```

因此没有云 API Key 也应能启动。

最低验收：

- [ ] `go test ./...` 通过；
- [ ] `GET /health` 成功；
- [ ] 普通聊天可返回；
- [ ] SSE 流可以逐事件读取并正常结束；
- [ ] 文档可写入内存 RAG；
- [ ] RAG 问答能返回基于已写入内容的结果；
- [ ] Agent 能调用允许的本地工具；
- [ ] 达到最大步骤、timeout 或取消时安全停止；
- [ ] 响应不暴露隐藏 Thought；
- [ ] 未配置 Key 时不会误连真实云 API。

若某项未用命令验证，就保持未勾选。

## 3. 项目目录

目标结构：

```text
Go工程/agentgo/
├─ cmd/server/                 # 组合依赖并启动 HTTP 服务
├─ internal/
│  ├─ config/                  # 环境变量解析与校验
│  ├─ llm/                     # 统一模型接口、mock、Responses API adapter
│  ├─ memory/                  # 会话历史抽象与内存实现
│  ├─ tool/                    # 工具定义、注册、校验和执行
│  ├─ rag/                     # chunk、embedding、memory/pgvector store
│  ├─ agent/                   # 有界 Agent loop
│  └─ httpapi/                 # 路由、JSON、SSE、错误映射
├─ testdata/                   # 固定测试样例
├─ docker-compose.yml          # V2 可选 PostgreSQL/pgvector
├─ .env.example               # 非敏感配置样例
├─ go.mod
└─ README.md                   # 实际运行状态与命令
```

最终以磁盘上的真实目录和测试为准。

章节讲的是依赖边界，不要求为了“好看”创建空包。

## 4. 总体架构

```text
HTTP Client
    ↓
internal/httpapi
    ↓
┌────────────────────────────────────┐
│ Chat Service / Agent Runner        │
│  ├─ memory.Store                   │
│  ├─ llm.Provider                   │
│  ├─ tool.Registry                  │
│  └─ rag.Store                      │
└────────────────────────────────────┘
    ↓                  ↓
mock / OpenAI       memory / pgvector
```

依赖方向应从外层指向抽象。

业务包不直接读取全局环境变量，也不直接依赖某个供应商完整响应类型。

## 5. 为什么默认 mock

mock 不是为了假装模型能力，而是为了稳定测试控制流。

它能验证：

- HTTP 参数和错误映射；
- memory 读写；
- Tool loop 的停止条件；
- SSE 事件组装；
- context 取消；
- RAG 数据流；
- 关键回归用例。

它不能证明：

- 真实模型回答质量；
- 真实工具选择准确率；
- 云 API 延迟与成本；
- provider 在当前模型上支持 strict schema；
- prompt 对真实模型足够稳健。

因此 V1 测控制流，真实 provider 另做契约测试与离线评估。

## 6. 配置

工程约定的环境变量：

| 变量 | 当前默认或可选值 | 含义 |
|---|---|---|
| `SERVER_ADDR` | `:8080` | HTTP 监听地址 |
| `AI_PROVIDER` | `mock` / `openai` | 模型适配器 |
| `AI_MODEL` | 运行时核验的模型 ID | 不在教程写死 |
| `OPENAI_API_KEY` | secret | 只在真实 provider 需要 |
| `OPENAI_BASE_URL` | 官方或已核验端点 | 兼容端点也需契约测试 |
| `REQUEST_TIMEOUT` | `45s` | 单请求总 timeout，允许范围不超过 10 分钟 |
| `MAX_AGENT_STEPS` | `4` | Agent 工具循环硬上限，允许范围 1～12 |
| `RAG_STORE` | `memory` / `pgvector` | 检索存储实现 |
| `DATABASE_URL` | PostgreSQL DSN | V2 pgvector 使用 |

配置加载要遵守：

1. 默认值只用于安全、可离线的学习模式；
2. `AI_PROVIDER=openai` 时缺 Key 应尽早报错；
3. `RAG_STORE=pgvector` 时缺 DSN 应尽早报错；
4. duration 和整数必须有范围校验；
5. 日志只显示 `key_present=true/false`，不显示密钥；
6. 测试通过注入环境或配置对象，不依赖开发机全局状态。

### 6.1 PowerShell 临时配置

```powershell
$env:AI_PROVIDER = "mock"
$env:RAG_STORE = "memory"
$env:SERVER_ADDR = ":8080"
go run ./cmd/server
```

真实 provider：

```powershell
$env:AI_PROVIDER = "openai"
$env:AI_MODEL = "<从当前官方模型页与账号核验>"
$env:OPENAI_API_KEY = "<只放环境或 secret manager>"
go run ./cmd/server
```

不要把真实值提交进 `.env.example`。

## 7. LLM 抽象

核心接口只表达项目所需能力，例如：

```go
type Provider interface {
	Generate(ctx context.Context, req Request) (Response, error)
	Stream(ctx context.Context, req Request) (<-chan Event, error)
}
```

这是架构示意，精确类型以 `internal/llm` 为准。

接口设计重点：

- `context.Context` 是第一个参数；
- request 使用统一消息和工具类型；
- response 区分文本、工具调用和结束原因；
- stream channel 有明确关闭与错误事件语义；
- provider 原始响应用于调试时也必须脱敏；
- 业务层不判断供应商私有字段。

### 7.1 mock provider

mock 应可预测，但不能把复杂业务规则复制一遍。

适合使用输入标记或注入脚本响应，覆盖：

- 普通文本；
- 一个工具调用；
- 工具后给最终回答；
- provider 错误；
- 流中途错误；
- context 取消。

### 7.2 OpenAI Responses adapter

真实 OpenAI 接入以官方 Responses API 为基线：

- [Responses API reference](https://platform.openai.com/docs/api-reference/responses)
- [Function calling](https://platform.openai.com/docs/guides/function-calling)
- [Streaming Responses](https://platform.openai.com/docs/guides/streaming-responses)

适配器负责：

```text
项目 Request
  → provider request
  → HTTP + timeout
  → provider output items / stream events
  → 项目 Response / Event
```

工具 call ID 必须正确关联工具结果。

当前 Agent Tool 续轮采用 `store:false`：adapter 每次请求 `reasoning.encrypted_content`，Runner 保留原始输入历史与上一轮完整 `response.output`，再追加 `function_call_output`。它不使用 `previous_response_id`，也不能只回传工具结果；加密 reasoning item 必须原样续传，但不能暴露给用户或写入普通业务日志。

不要把 Chat Completions 的字段猜到 Responses API 上，也不要假设第三方兼容端点支持全部事件。

## 8. Memory

V1 的内存 memory 用于理解会话读写。

当前接口的简化形状：

```go
type Store interface {
	History(tenantID, userID, conversationID string) ([]llm.InputItem, error)
	Append(tenantID, userID, conversationID string, item llm.InputItem) error
	Delete(tenantID, userID, conversationID string) error
}
```

关键不变量：

- 同一 conversation 由 tenant、user 与 conversation ID 共同隔离；
- 返回顺序明确；
- 读写并发安全；
- 同一会话的一整轮读历史、调用模型和提交结果需要串行或版本控制；
- 有最大消息数或字符/token 预算；
- 不把隐藏 Thought 保存成面向用户的历史；
- 内存实现重启后丢失，文档应明确。

### 8.1 为什么不能无限拼历史

历史越长会：

- 增加 token 和 TTFT；
- 挤掉 RAG 证据与输出预算；
- 引入过期事实；
- 放大敏感信息暴露范围。

V3 可加入摘要、窗口和长期记忆，但每种记忆都要定义来源、失效与删除策略。

## 9. Tool Registry

工具层至少分三部分：

```text
Definition：名称、描述、input schema
Policy：当前用户是否允许调用、是否需确认
Handler：校验后执行具体动作
```

概念接口：

```go
type Handler func(ctx context.Context, args json.RawMessage) (Result, error)

type Registry interface {
	Definitions(ctx context.Context, principal Principal) []Definition
	Execute(ctx context.Context, principal Principal, call Call) (Result, error)
}
```

模型只提出调用。

`Registry.Execute` 才是安全边界。

### 9.1 演示身份

V1 从 HTTP header 读取：

```text
X-User-ID: demo-user
```

这只是本地演示身份，不是认证。

当前演示模式不允许客户端另行指定受信任 tenant；服务端把租户边界绑定到演示用户。Header 本身仍可伪造，所以这不构成生产隔离证明。

生产环境必须由可信认证中间件验证 token/session，再把 principal 放入 context。

绝不能因为 header 名看起来像身份就信任公网请求。

### 9.2 Tool step 响应

Agent 的 steps 可以返回可观察摘要：

- tool；
- args；
- result 摘要；
- status；
- duration。

不得返回或依赖模型隐藏 Thought。

这是可观测动作，不是“完整思维链”。

## 10. Agent Loop

有界循环：

```text
构造模型输入
  ↓
调用模型
  ├─ 最终文本 → 返回
  └─ Tool calls
       ↓ 校验、授权、执行
       ↓ 回填结果
       └─ 下一轮模型调用
```

伪代码：

```go
for step := 0; step < maxSteps; step++ {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	resp, err := model.Generate(ctx, request)
	if err != nil {
		return Result{}, classify(err)
	}

	if len(resp.ToolCalls) == 0 {
		return final(resp.Text, steps), nil
	}

	functionOutputs := make([]InputItem, 0, len(resp.ToolCalls))
	for _, call := range resp.ToolCalls {
		toolResult := registry.Execute(ctx, principal, call)
		functionOutputs = append(functionOutputs, FunctionOutput(call.ID, toolResult))
		steps = append(steps, summarize(call, toolResult))
	}

	// store:false：先保留完整 resp.OutputItems，再追加工具结果。
	request = Request{
		Input: append(copyItems(resp.OutputItems), functionOutputs...),
		Tools: tools,
		Store: false,
	}
}

return Result{}, ErrMaxSteps
```

仍需补充的生产约束：

- 相同 tool+args 循环检测；
- 每个工具单独 timeout；
- 写操作幂等键；
- 并行工具的上限与取消；
- 结果大小限制；
- provider 和工具错误分类；
- 总预算。

## 11. RAG V1

V1 使用内存 store 和可测试 embedder。

数据流：

```text
文档
  → chunk
  → embed
  → memory store

问题
  → embed
  → similarity search top-k
  → 组装带 source ID 的上下文
  → LLM 回答
```

### 11.1 Chunk

切块要保留：

- document ID；
- chunk ID；
- 标题；
- 内容；
- 可选元数据；
- 顺序或位置。

不要只保存裸文本，否则引用、更新和删除很难处理。

### 11.2 测试 embedder

V1 的本地/确定性 embedder 用于验证 store 逻辑，不代表语义检索质量。

它适合测试：

- 维度检查；
- 插入与覆盖；
- top-k 边界；
- 排序稳定性；
- 空库行为；
- context 取消。

真实 embedding 上线前必须重建向量并运行相关性评估。

## 12. HTTP API

### 12.1 健康检查

```http
GET /health
```

健康检查应轻量。

生产中可进一步区分 liveness 和 readiness，但 V1 不必伪装完整编排平台。

### 12.2 普通聊天

```http
POST /api/chat
Content-Type: application/json
X-User-ID: demo-user

{
  "message": "解释 goroutine 和线程的区别",
  "conversation_id": "go-study-1"
}
```

`conversation_id` 可省略，具体生成与返回形式以实际 handler 为准。

### 12.3 流式聊天

```http
POST /api/chat/stream
Content-Type: application/json
X-User-ID: demo-user

{
  "message": "分三步解释 context 取消",
  "conversation_id": "go-study-1"
}
```

返回 SSE。

客户端必须按 SSE 事件解析，不按任意网络 chunk 解析。

### 12.4 写入 RAG

```http
POST /api/rag/ingest
Content-Type: application/json
X-User-ID: demo-user

{
  "id": "go-context-note",
  "title": "Go context 学习笔记",
  "content": "context 用于传递取消、deadline 和请求范围值。"
}
```

### 12.5 RAG 问答

```http
POST /api/rag/ask
Content-Type: application/json
X-User-ID: demo-user

{
  "question": "context 主要传递什么？",
  "top_k": 3
}
```

### 12.6 Agent 执行

```http
POST /api/agent/run
Content-Type: application/json
X-User-ID: demo-user

{
  "message": "使用可用工具完成这个任务"
}
```

响应中的 steps 只应是已执行动作摘要。

## 13. 验证方式

路由请求示例以项目 `README.md` 为准，SSE 使用 `curl.exe -N` 观察逐事件到达。

```powershell
go test ./...
go test -race ./...
go vet ./...
```

包级测试应覆盖配置边界、provider 映射、会话隔离、工具校验、RAG top-k、Agent 停止条件以及 HTTP/SSE 错误路径。当前实际覆盖项以 `go test ./...` 的测试源码为准；provider adapter 可用 `httptest.Server` 模拟 429、500、慢响应和断流而不消耗真实额度。

客户端必须解析 SSE 事件并区分 done/error；HTTP 错误应稳定映射，不能返回 provider 原始 body、SQL、路径或堆栈。

## 14. V2：pgvector

V2 的目标是替换 RAG store，而不是重写全部 Agent。

```text
rag.Store interface
├─ MemoryStore       # V1
└─ PGVectorStore     # V2
```

### 14.1 启动数据库

以仓库实际 `docker-compose.yml` 为准：

```powershell
docker compose up -d
docker compose ps
```

再配置：

```powershell
$env:RAG_STORE = "pgvector"
$env:DATABASE_URL = "<本机实际 DSN，不提交密码>"
```

### 14.2 Schema 必须记录的字段

至少考虑：

- tenant/user；
- document ID；
- chunk ID；
- title/content；
- metadata；
- embedding；
- embedding model；
- embedding dimension；
- corpus version；
- created/updated time。

### 14.3 向量维度

当前工程固定使用 128 维 Hash Embedding，并据此创建 `vector(128)`。它只验证流程，不代表真实语义质量。接入真实 embedding 后，维度由实际输出决定。

当前 Compose 使用 pgvector 0.8.1，adapter 创建 `vector` HNSW，因此支持范围受该索引限制；大于 2000 维不能直接沿用当前 DDL。

更换 embedding 模型时：

1. 建新版本索引或表；
2. 重算语料向量；
3. 双跑评估；
4. 切换读取；
5. 保留回滚；
6. 再清理旧版本。

不能把不同语义空间的向量直接混在一起比较。

### 14.4 检索评估

建立小型标注集：

```text
question
relevant_chunk_ids
forbidden_tenant_ids
expected_answer_points
```

分别计算：

- retrieval Recall@k；
- 排名指标；
- tenant 泄漏数必须为 0；
- grounded answer；
- citation 正确性；
- 延迟和成本。

没有评估集，不声称“召回率提升”。

## 15. V3：生产增强

V3 尚未完成，是按风险逐项实现的检查单：

| 维度 | 必要工作 |
|---|---|
| 身份 | OIDC/session/JWT、principal 进入 context、资源级授权 |
| 韧性 | 分层 timeout、安全重试、幂等、有界并发、优雅关闭 |
| 观测 | request/trace ID、模型与 token、检索/工具状态、耗时、错误分类 |
| 发布 | 回归集、prompt/schema 版本、小流量验证、失败阈值、回滚 |
| 成本 | 用户配额、输入输出上限、工具结果裁剪、异常告警 |

日志和 trace 必须脱敏，供应商 fallback 也不能造成无界重复计费。

未来与短链结合时优先增加只读统计工具；短链数据库仍是权威事实源，用户身份必须来自认证 context，而不是模型传入的 `user_id`。

完成 V1 后可以诚实描述为“用 Go 抽象 LLM、Memory、Tool、RAG 和有界 Agent loop，并以 mock/httptest 验证主要控制流”。具体声称 SSE、超时或错误路径已覆盖前，应先指出对应测试。只有完成标注集评估后，才声称 pgvector 的检索效果；没有压测和 baseline，不写 QPS 或提升百分比。

## 16. 验收题

1. 为什么默认 mock 仍有工程价值？
2. 为什么 provider adapter 不应泄漏供应商类型到 agent 包？
3. memory 为什么必须绑定 tenant、user 与 conversation？
4. Tool definition、policy、handler 为什么要分开？
5. Agent loop 至少有哪些停止条件？
6. SSE 客户端为什么不能按 TCP chunk 解析？
7. V1 的测试 embedder 为什么不能证明语义召回质量？
8. 更换 embedding 模型为什么通常需要重建向量？
9. `X-User-ID` 为什么只可用于演示？
10. steps 与 Thought 的边界是什么？

## 17. 本章完成条件

只有当你亲自运行、阅读关键包并解释一次失败路径，才算完成。

建议最终提交一份本地验证记录：

```text
commit/build:
go version:
go test ./...:
go test -race ./...:
provider:
rag store:
routes checked:
known limitations:
next step:
```

这份证据也会成为下一章面试表达的事实来源。
