# 附录 A：版本与供应商能力矩阵

```yaml
last_verified: 2026-07-15
scope: Go-first Agent 学习路线
policy: 运行前再次核验供应商官方文档；模型名、价格、限额和区域可用性不视为长期事实
```

## 1. 这张表应该怎样使用

本附录不是“模型排行榜”，也不承诺某个供应商永远支持某项能力。

它解决三个工程问题：

1. 接入前应该核验什么；
2. 如何把供应商差异隔离在适配器中；
3. 如何用最小探测请求证明当前账号、区域和模型确实可用。

同一供应商的不同模型、API 形态、账号等级和区域可能不同。

因此，“文档写了支持”仍不等于“当前项目配置一定支持”。

最终证据按可信度从高到低排列：

1. 当前账号发出的真实探测请求及响应；
2. 供应商官方 API 文档与模型页；
3. 官方 SDK 类型和发布说明；
4. 第三方框架文档；
5. 博客、课程和转述。

## 2. 项目版本策略

| 对象 | 建议策略 | 验证命令或证据 |
|---|---|---|
| Go | 在 `go.mod` 固定最低版本 | `go version`、`go env` |
| Go 依赖 | 提交 `go.mod` 与 `go.sum` | `go mod verify` |
| PostgreSQL | Compose 或部署清单固定主版本 | `SELECT version();` |
| pgvector | 固定镜像或扩展版本 | `SELECT extversion FROM pg_extension WHERE extname='vector';` |
| 模型 API | 模型名来自环境变量，不写死在业务代码 | 启动日志只记录非敏感配置 |
| JSON Schema | Schema 进入仓库并版本化 | 单元测试验证合法与非法样例 |
| MCP | 记录协商得到的协议版本 | 初始化日志、能力快照 |
| A2A | 记录 Agent Card 和协议版本 | 契约测试快照 |
| Ollama | 固定部署版本与模型摘要 | `ollama --version`、模型清单 |
| vLLM | 固定容器 tag 或 Python lock | 服务启动参数、健康检查 |

不要在学习笔记里写“最新版”后长期不维护。

更稳妥的记录方式是：

```text
verified_at=2026-07-15
provider=<供应商>
api_surface=<接口形态>
model=<运行时配置值>
region=<区域>
capabilities=<探测通过的能力>
evidence=<测试记录或链接>
```

## 3. 供应商能力核验矩阵

下表的“核验项”是测试清单，不是对所有模型的永久承诺。

| 接入面 | 官方入口 | 接入前必须核验 | Go 侧隔离点 |
|---|---|---|---|
| OpenAI Responses API | [Responses API reference](https://platform.openai.com/docs/api-reference/responses) | 文本、流式事件、工具调用、结构化输出、输入类型、用量字段、保留策略 | `internal/llm` provider adapter |
| OpenAI 模型 | [Models](https://platform.openai.com/docs/models) | 模型 ID、上下文、工具与结构化输出支持、价格、限额 | 配置中的 model 与 capability probe |
| Anthropic Messages API | [API docs](https://docs.anthropic.com/en/api/overview) | 工具、流式事件、内容块、缓存、模型可用区域 | 内容块到统一事件的转换 |
| Google Gemini API | [Gemini API docs](https://ai.google.dev/gemini-api/docs) | function calling、结构化输出、流式、多模态、模型 ID | request/response adapter |
| DeepSeek API | [API docs](https://api-docs.deepseek.com/) | OpenAI 风格兼容范围、工具、JSON 输出、流式事件、模型 ID | 不假设完全兼容，单独契约测试 |
| Azure OpenAI | [Azure OpenAI docs](https://learn.microsoft.com/azure/ai-services/openai/) | deployment name、API version、区域、身份认证、内容过滤 | endpoint、deployment、credential adapter |
| AWS Bedrock | [Bedrock docs](https://docs.aws.amazon.com/bedrock/) | Converse API、模型区域、IAM、工具能力、流式 | SigV4/SDK 与统一 LLM 接口 |
| Ollama | [API docs](https://docs.ollama.com/api) | 本机模型、上下文、工具/结构化输出的模型与版本限制 | 本地 endpoint adapter |
| vLLM | [OpenAI-compatible server](https://docs.vllm.ai/en/latest/serving/openai_compatible_server.html) | 实际支持的路由与字段、chat template、并发与显存 | 兼容端点也保留能力开关 |

### 3.1 为什么“OpenAI 兼容”仍需单独测试

“兼容”通常只表示某些请求与响应形状接近，不意味着：

- 每个字段都被接受；
- 每种流式事件都一致；
- 工具调用 ID 和结束原因完全一致；
- strict JSON Schema 约束完全一致；
- 错误码、重试语义和用量统计一致；
- Responses API 与 Chat Completions API 可以互换；
- 服务端保留、缓存和安全策略一致。

项目应把“协议形状”和“能力开关”分开：

```go
type Capabilities struct {
	Streaming        bool
	ToolCalling      bool
	StrictJSONSchema bool
	UsageInStream    bool
	VisionInput      bool
}
```

这个结构只代表探测结果或配置，不代表供应商的永久属性。

## 4. OpenAI 接入的当前基线

本路线的新代码优先把 OpenAI 接入建立在 Responses API 抽象上。

官方入口：

- [Responses API reference](https://platform.openai.com/docs/api-reference/responses)
- [Responses guide](https://platform.openai.com/docs/guides/responses-vs-chat-completions)
- [Function calling](https://platform.openai.com/docs/guides/function-calling)
- [Structured Outputs](https://platform.openai.com/docs/guides/structured-outputs)
- [Streaming Responses](https://platform.openai.com/docs/guides/streaming-responses)
- [Models](https://platform.openai.com/docs/models)

接入时不要从本附录复制模型名。

模型名由部署当天的官方模型页和账号可见列表决定，例如：

```powershell
$env:AGENTGO_LLM_PROVIDER = "openai"
$env:AGENTGO_LLM_MODEL = "<从当前官方模型页与账号核验>"
```

需要记录的 Responses API 差异包括：

- 输入由哪些 item 组成；
- 输出可能包含哪些 item；
- 工具调用参数如何返回；
- 工具结果如何关联 call ID；
- 流式事件的事件名和顺序；
- 是否启用服务端状态以及数据保留配置；
- 错误响应中哪些字段可安全记录。

## 5. 最小能力探测

供应商适配器至少运行以下契约测试。

### 5.1 普通文本

输入一个确定性较强的问题，断言：

- HTTP 成功；
- 有非空文本；
- 能取得请求 ID 或本地 trace ID；
- 超时可以由 `context.Context` 取消。

### 5.2 流式输出

断言：

- 首个事件能在总响应结束前到达；
- 事件可以按顺序组装；
- 客户端取消后连接释放；
- 中途错误不会被当作正常结束。

### 5.3 工具调用

提供一个无副作用的测试工具：

```text
name=get_test_clock
input={"timezone":"Asia/Shanghai"}
```

断言：

- 模型生成可解析参数；
- 未知字段被拒绝或明确处理；
- call ID 能关联工具结果；
- 最大循环次数生效；
- 工具异常不会把密钥或堆栈原样送回模型。

### 5.4 严格结构化输出

使用附录 B 的 schema，分别提交：

- 一个正常请求；
- 一个容易诱导额外字段的请求；
- 一个要求违反 schema 的请求。

只有实际响应持续通过本地校验，才能把能力标记为可用。

### 5.5 限流与重试

断言：

- 只重试可重试错误；
- 指数退避带随机抖动；
- 尊重服务端重试提示；
- 已产生副作用的工具不被盲目重放；
- 到达总 deadline 后停止。

## 6. Embedding 与向量检索核验

Embedding 模型更换会影响向量维度和语义空间。

上线前记录：

| 字段 | 含义 |
|---|---|
| provider | embedding 供应商 |
| model | 精确模型 ID |
| dimension | 实际返回维度 |
| normalization | 是否归一化 |
| distance | cosine、inner product 或 L2 |
| corpus_version | 语料版本 |
| embedded_at | 向量生成时间 |

不要把不同模型生成的向量无标记地放入同一索引。

维度、距离函数和索引类型必须与 pgvector 表定义及查询一致。

## 7. 本地推理能力核验

### Ollama

适合个人开发、离线实验和快速换模型。

核验：

- API 版本和 endpoint；
- 模型是否已经拉取；
- 模型上下文和内存是否足够；
- 工具或 JSON 能力是否由当前模型支持；
- 首次加载与常驻后的延迟差异。

### vLLM

适合有 GPU 服务化需求的场景，但性能取决于模型、量化、显卡、并发、上下文和启动参数。

核验：

- chat template 是否正确；
- GPU 架构与精度是否兼容；
- KV cache 可用空间；
- 最大上下文与并发组合；
- OpenAI 兼容路由的实际字段支持；
- 监控指标与压测条件。

未经同机同模型压测，不写“比某方案快多少”。

## 8. 每次升级的检查单

- [ ] 阅读官方 release notes 和迁移说明。
- [ ] 更新锁文件或镜像 tag。
- [ ] 在测试环境运行文本、流式、工具、schema 契约测试。
- [ ] 检查错误类型和重试分类是否变化。
- [ ] 检查 token 用量字段是否变化。
- [ ] 检查数据保留与隐私设置。
- [ ] 检查模型弃用日期。
- [ ] 检查 embedding 维度与索引兼容性。
- [ ] 检查 MCP/A2A 协议协商结果。
- [ ] 更新 `last_verified` 和证据链接。

## 9. 结论

能力矩阵的价值不在于填满勾号，而在于让每个勾号都有可复现证据。

学习项目可以只实现一个 mock provider 和一个真实 provider；生产设计则必须允许能力探测失败、模型被替换以及协议逐步演进。
