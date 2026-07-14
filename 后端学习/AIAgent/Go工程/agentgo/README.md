# AgentGo：可运行的 Go Agent/RAG 基线工程

AgentGo 为 `后端学习/AIAgent` 的代码验证基线。它刻意分成三个演进层级：

- **V1**：Mock Provider + 内存会话 + 内存 RAG，离线即可运行；
- **V2**：OpenAI Responses API + PostgreSQL/pgvector；
- **V3**：队列、Redis、完整鉴权、评估与可观测性，由文档说明升级边界。

工程不会在没有测试证据时声称“生产可用”，也不会把模型生成的 Thought 暴露给 API。Agent trace 只记录 Tool 名、经过校验的参数、结果摘要、状态与时延。

---

## 1. 目录结构

```text
agentgo/
├── cmd/server/                 # 进程入口、依赖装配、优雅关闭
├── internal/config/            # 环境变量与边界校验
├── internal/llm/               # Mock 与 OpenAI Responses Provider
├── internal/memory/            # 有界会话窗口
├── internal/tool/              # Tool Registry、计算器、知识库检索
├── internal/rag/               # 分块、Embedding、Memory/pgvector Store
├── internal/agent/             # 有界 Tool Call 状态机
├── internal/httpapi/           # Gin API、POST SSE、身份边界
├── testdata/                    # 可公开测试资料
├── docker-compose.yml          # 可选 pgvector
├── .env.example
└── go.mod
```

---

## 2. V1：无需密钥运行

```powershell
cd F:\study\后端学习\AIAgent\Go工程\agentgo
go mod tidy
go test ./...
$env:AI_PROVIDER="mock"
$env:RAG_STORE="memory"
go run ./cmd/server
```

成功日志类似：

```text
agentgo listening on :8080 provider=mock rag=memory
```

健康检查：

```powershell
Invoke-RestMethod http://localhost:8080/health
```

预期：

```text
status provider rag_store
------ -------- ----------
ok     mock     memory
```

---

## 3. API 验收

所有 `/api` 路由要求演示身份 Header：

```powershell
$headers = @{ "X-User-ID" = "demo-user" }
```

演示模式把 `X-User-ID` 同时作为用户与租户命名空间，并主动拒绝客户端传入 `X-Tenant-ID`。这仍不是真实认证：生产环境必须将演示 Header 替换为 JWT、Session 或网关认证上下文，由服务端注入 tenant 与 user。

本机如果使用 Windows PowerShell 5.1，直接把含中文的 JSON 字符串传给 `-Body` 可能按旧编码发送并变成 `?`。下面统一显式转为 UTF-8 字节；PowerShell 7 通常也可沿用这种写法。

### 3.1 普通聊天

```powershell
$body = @{ message = "你好，解释一下 context.Context" } | ConvertTo-Json
Invoke-RestMethod `
  -Uri http://localhost:8080/api/chat `
  -Method POST `
  -Headers $headers `
  -ContentType "application/json; charset=utf-8" `
  -Body ([Text.Encoding]::UTF8.GetBytes($body))
```

第二轮可把返回的 `conversation_id` 带回，服务端会读取同一用户的有界窗口。

### 3.2 POST SSE 流式聊天

PowerShell 的 `Invoke-RestMethod` 不适合观察逐块到达，推荐：

```powershell
curl.exe -N -X POST http://localhost:8080/api/chat/stream `
  -H "Content-Type: application/json; charset=utf-8" `
  -H "X-User-ID: demo-user" `
  -d '{"message":"stream an explanation of Go context cancellation"}'
```

服务端事件：

```text
event: meta
event: delta
event: done
```

客户端断开会取消请求上下文，下游 Provider 应停止读取流。

### 3.3 写入知识库

```powershell
$content = Get-Content -LiteralPath .\testdata\handbook.md -Raw
$body = @{
  id = "agentgo-handbook"
  title = "AgentGo 演示手册"
  content = $content
  metadata = @{ source = "testdata/handbook.md" }
} | ConvertTo-Json -Depth 5

Invoke-RestMethod `
  -Uri http://localhost:8080/api/rag/ingest `
  -Method POST `
  -Headers $headers `
  -ContentType "application/json; charset=utf-8" `
  -Body ([Text.Encoding]::UTF8.GetBytes($body))
```

### 3.4 RAG 问答

```powershell
$body = @{ question = "AgentGo 默认为什么不需要 API Key？"; top_k = 3 } | ConvertTo-Json
Invoke-RestMethod `
  -Uri http://localhost:8080/api/rag/ask `
  -Method POST `
  -Headers $headers `
  -ContentType "application/json; charset=utf-8" `
  -Body ([Text.Encoding]::UTF8.GetBytes($body))
```

返回的 `citations` 是检索候选，不自动证明最终答案忠实。真正项目必须通过评测集检查答案是否被证据支持。

### 3.5 Agent Tool 循环

```powershell
$body = @{ message = "请计算 12+8" } | ConvertTo-Json
Invoke-RestMethod `
  -Uri http://localhost:8080/api/agent/run `
  -Method POST `
  -Headers $headers `
  -ContentType "application/json; charset=utf-8" `
  -Body ([Text.Encoding]::UTF8.GetBytes($body))
```

Mock Provider 会请求 `calculator`，服务端用安全表达式解析器执行，再把结构化结果交回模型。`steps` 不包含模型私有思维链。

---

## 4. V2：OpenAI Responses API

先确认账号、模型和数据策略，再设置：

```powershell
$env:AI_PROVIDER="openai"
$env:AI_MODEL="填写当前实际可用模型"
$env:OPENAI_API_KEY="你的密钥"
$env:OPENAI_BASE_URL="https://api.openai.com"
go run ./cmd/server
```

代码使用 `/v1/responses`，处理：

- `output_text`；
- `function_call` 与 `call_id`；
- `function_call_output`；
- `store:false` 下的 `reasoning.encrypted_content` 与完整 output replay；
- `response.output_text.delta`；
- `response.completed`；
- `response.failed`、`response.incomplete`、refusal 与顶层 `error`。

Agent Tool 循环使用无状态续接：每轮保留原始输入历史和全部 `response.output` item，再追加 `function_call_output`。加密 reasoning item 只回传给模型，不会作为 API 响应或 trace 暴露。

不要把其他供应商的 Chat Completions 兼容地址直接填进来并假设 Responses、Tool、Embedding 都兼容。应为不同能力实现和测试独立 adapter。

---

## 5. V2：PostgreSQL + pgvector

启动数据库：

```powershell
docker compose up -d
docker compose ps
```

切换 Store：

```powershell
$env:RAG_STORE="pgvector"
$env:DATABASE_URL="postgres://agent:agent@localhost:5432/agent?sslmode=disable"
go run ./cmd/server
```

当前 pgvector adapter：

- 启动时创建扩展、表和 HNSW 索引；
- 每个查询强制包含 `tenant_id`；
- 文档更新先在事务外完成分块和 Embedding，再在数据库事务内替换 chunks；
- 使用确定性 Hash Embedding，便于离线验证。

接入真实 Embedding Provider 时，换模型或维度必须重建向量；不要把 Chat 模型地址当成 Embedding 地址。

---

## 6. 配置表

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `SERVER_ADDR` | `:8080` | HTTP 监听地址 |
| `AI_PROVIDER` | `mock` | `mock` 或 `openai` |
| `AI_MODEL` | 空 | OpenAI 模式必须显式设置 |
| `OPENAI_API_KEY` | 空 | 仅从环境读取 |
| `OPENAI_BASE_URL` | `https://api.openai.com` | Responses API 基址 |
| `REQUEST_TIMEOUT` | `45s` | 单请求总预算 |
| `MAX_AGENT_STEPS` | `4` | 允许范围 1～12 |
| `RAG_STORE` | `memory` | `memory` 或 `pgvector` |
| `DATABASE_URL` | 空 | pgvector 模式必填 |

---

## 7. 安全边界

- Mock/测试数据不得含真实凭据或隐私。
- Tool 只通过 Registry 白名单暴露。
- JSON Schema 之后仍执行 Go 业务校验。
- Principal 从服务端上下文进入 Tool，不从模型参数读取。
- 计算器不使用 `eval`、Shell 或脚本引擎。
- RAG 检索始终带 tenant 条件。
- 文档片段和 Tool 返回都按不可信数据处理。
- API 不返回原始 Chain-of-Thought。
- Agent 具有最大步数、总 deadline 和单 Tool timeout。
- SSE 使用 POST，避免问题进入 URL、代理日志和浏览器历史。

---

## 8. 测试说明

```powershell
go test ./...
go test -race ./...
go vet ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

当前测试覆盖：

- 配置边界；
- Responses API 的无状态 reasoning/Tool 续接、拒答、失败与流式终态；
- 会话窗口、用户隔离、失败不提交与同会话串行；
- 请求体限制、演示身份边界与 SSE 提交语义；
- 真正的 chunk overlap；
- RAG tenant 隔离；
- pgvector 向量序列化与 HNSW 维度边界；
- Tool 严格参数与计算器；
- Agent 调用循环；
- HTTP 健康检查、身份校验和聊天接口。

真实模型和数据库集成测试需要凭据/容器，默认单元测试不会访问外部系统。

---

## 9. 已知限制

- Hash Embedding 只用于验证流程，不代表真实语义质量。
- 内存会话和 RAG 在重启后清空。
- 内存实现限制了单会话消息数和单请求体大小，但会话总数、文档总数仍没有持久化配额或 TTL；它是本地学习基线，不是公网部署方案。
- pgvector 示例尚未包含异步 ingest、outbox 和文档删除 API。
- Mock Provider 只覆盖可预测测试路径。
- OpenAI Provider 当前聚焦文本、Tool 和 HTTP SSE，不覆盖图像、音频、WebSocket 和托管工具。
- 演示身份 Header 不是生产鉴权。

这些限制是后续 V2/V3 的任务清单，不应在简历中写成已经实现。
