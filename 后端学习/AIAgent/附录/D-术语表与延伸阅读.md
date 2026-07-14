# 附录 D：术语表与延伸阅读

```yaml
last_reviewed: 2026-07-15
reading_policy: 协议和 API 先读官方文档；论文用于理解方法边界；实现细节以当前版本为准
```

## 1. LLM 与推理术语

| 术语 | 简明定义 |
|---|---|
| Token | 模型处理的离散符号 ID；不等同于字、词或 UTF-8 字节。 |
| Tokenizer | 文本与 token ID 序列之间的编码/解码系统。 |
| Vocabulary | tokenizer 可直接表示的 token 集合。 |
| Embedding | 把离散对象映射到连续向量；输入 embedding 与检索 embedding 用途不同。 |
| Transformer | 以注意力、前馈网络、残差连接和归一化等组件构成的序列模型架构族。 |
| Self-Attention | 同一序列位置之间基于 Q/K/V 计算信息聚合权重。 |
| Causal Mask | 阻止当前位置在自回归生成时访问未来 token。 |
| MHA | Multi-Head Attention，多组注意力头并行学习不同投影。 |
| MQA | Multi-Query Attention，多个 query 头共享较少的 K/V 头。 |
| GQA | Grouped-Query Attention，query 头按组共享 K/V 头。 |
| RoPE | Rotary Position Embedding，把位置信息编码进注意力相关旋转。 |
| Context Window | 一次推理可处理的 token 范围；实际可用输入还需给输出和协议开销留空间。 |
| Prefill | 对整段已有输入并行计算并建立 KV cache 的阶段。 |
| Decode | 基于已有 cache 逐 token 生成后续输出的阶段。 |
| KV Cache | 保存历史层的 key/value，避免每个 decode 步骤重复计算全部历史。 |
| TTFT | Time To First Token，从请求到首个可见 token/事件的时间。 |
| TPOT | Time Per Output Token，生成阶段相邻输出 token 的平均时间。 |
| Throughput | 单位时间处理的请求、token 或任务量；必须说明口径。 |
| Quantization | 用较低精度表示权重或激活，以降低资源占用，可能影响质量和兼容性。 |
| Continuous Batching | 运行中动态加入和移除请求的批处理调度方式。 |
| Hallucination | 输出缺乏依据、与事实或提供资料冲突；不是单一可完全消除的故障类型。 |

## 2. 适配与训练术语

| 术语 | 简明定义 |
|---|---|
| Pre-training | 在大规模数据上学习通用预测目标的训练阶段。 |
| SFT | Supervised Fine-Tuning，用输入与目标输出对进行监督适配。 |
| PEFT | Parameter-Efficient Fine-Tuning，只训练较小部分新增或选定参数的方法族。 |
| LoRA | 用低秩增量矩阵适配部分权重的 PEFT 方法。 |
| QLoRA | 在量化基座上进行 LoRA 训练的一类做法；训练细节仍需核对实现。 |
| Preference Data | 对候选回答的偏好、排序或比较数据。 |
| RLHF | 以人类反馈构建奖励信号并使用强化学习优化策略的一类流程。 |
| DPO | Direct Preference Optimization，直接从偏好对优化策略的目标。 |
| RLVR | Reinforcement Learning with Verifiable Rewards，用可程序验证结果构造奖励的范式。 |
| Distillation | 用教师模型输出或分布训练较小/不同学生模型。 |
| Catastrophic Forgetting | 适配新数据后，原有能力明显退化。 |
| Data Contamination | 测试数据或其近似内容进入训练/提示，导致评估失真。 |

## 3. Agent 与工具术语

| 术语 | 简明定义 |
|---|---|
| Agent | 围绕目标，组合模型、状态、工具和控制循环的应用系统。 |
| Tool Calling | 模型输出结构化调用意图，由应用校验并执行工具。 |
| Tool Result | 工具执行后的结构化结果；它仍是不可信外部数据。 |
| Agent Loop | 模型决策、工具执行、结果回填、再决策的有界循环。 |
| Planner | 产生任务分解或计划的组件；计划不等于授权。 |
| Router | 根据输入把请求分发到模型、工具或子流程的组件。 |
| Handoff | 把任务上下文和责任移交给另一个 Agent/流程。 |
| Guardrail | 输入、输出、工具或流程上的约束与检测；不能只靠 prompt。 |
| Idempotency Key | 标识同一逻辑操作，避免重试导致重复副作用的键。 |
| Human-in-the-loop | 在高风险或低置信环节引入人工审核或确认。 |
| Thought | 模型内部推理的泛称；应用不应依赖或公开隐藏推理文本。 |
| Trace | 跨模型、检索、工具和数据库阶段关联的可观察记录。 |

## 4. RAG 与检索术语

| 术语 | 简明定义 |
|---|---|
| RAG | Retrieval-Augmented Generation，先检索外部资料，再让模型基于资料生成。 |
| Chunk | 被索引和召回的文本片段。 |
| Dense Retrieval | 使用语义向量相似度的检索。 |
| Sparse Retrieval | 使用词项权重等稀疏表示的检索，如 BM25。 |
| Hybrid Search | 组合稠密与稀疏信号。 |
| Reranker | 对初召回候选重新打分排序的模型或规则。 |
| Top-k | 返回得分最高的 k 个候选；k 大不等于答案更好。 |
| Recall@k | 相关项是否进入前 k 个结果的检索指标。 |
| MRR | Mean Reciprocal Rank，首个相关结果倒数排名的均值。 |
| Groundedness | 回答中的主张是否被给定资料支持。 |
| Citation | 回答主张与来源之间的可核验关联。 |
| pgvector | PostgreSQL 的向量数据类型、操作符和索引扩展。 |

## 5. 协议与服务术语

| 术语 | 简明定义 |
|---|---|
| JSON-RPC | 用 JSON 表达请求、响应和通知的 RPC 协议。 |
| MCP | Model Context Protocol，连接 AI 应用与上下文/工具服务器的开放协议。 |
| MCP Host | 承载 AI 应用并管理一个或多个 MCP client 的宿主。 |
| MCP Client | 与一个 MCP server 建立会话并协商能力的协议组件。 |
| MCP Server | 暴露 tools、resources、prompts 等能力的一端。 |
| A2A | Agent2Agent，用于独立 Agent 之间能力发现、消息与任务协作的开放协议。 |
| Agent Card | A2A 中描述 Agent endpoint、能力、skill 和认证需求的元数据。 |
| SSE | Server-Sent Events，服务端通过 HTTP 持续发送文本事件。 |
| Streamable HTTP | MCP 的一种 HTTP 传输方式；具体会话与安全语义以当前规范为准。 |
| OpenAI-compatible | 对部分 OpenAI 风格接口的兼容描述，不保证全部字段和行为一致。 |

## 6. Go 工程术语

| 术语 | 简明定义 |
|---|---|
| `context.Context` | 跨调用传递取消、deadline 和请求范围值的标准机制。 |
| Goroutine | 由 Go runtime 调度的轻量并发执行单元。 |
| Channel | goroutine 间传递值和同步的类型化通道。 |
| Data Race | 多个 goroutine 未同步访问同一内存且至少一个写入。 |
| Interface | 由方法集合描述行为的 Go 类型；应由使用方按需定义小接口。 |
| Middleware | 包装 HTTP handler 处理认证、日志、恢复、限流等横切逻辑。 |
| Table-driven Test | 用测试用例表覆盖多组输入与期望的 Go 测试写法。 |
| Fuzzing | 自动生成和变异输入以发现崩溃或不变量破坏。 |
| Backpressure | 下游处理能力不足时，系统限制上游继续生产的机制。 |

## 7. 官方协议与 API 阅读

### Go

- [The Go Programming Language Specification](https://go.dev/ref/spec)
- [Go memory model](https://go.dev/ref/mem)
- [Effective Go](https://go.dev/doc/effective_go)
- [`context` package](https://pkg.go.dev/context)

### OpenAI

- [Responses API reference](https://platform.openai.com/docs/api-reference/responses)
- [Function calling](https://platform.openai.com/docs/guides/function-calling)
- [Structured Outputs](https://platform.openai.com/docs/guides/structured-outputs)
- [Streaming Responses](https://platform.openai.com/docs/guides/streaming-responses)
- [Models](https://platform.openai.com/docs/models)

### MCP

- [MCP documentation](https://modelcontextprotocol.io/docs/)
- [MCP specification](https://modelcontextprotocol.io/specification/)
- [MCP specification repository](https://github.com/modelcontextprotocol/modelcontextprotocol)
- [Official Go SDK repository](https://github.com/modelcontextprotocol/go-sdk)

协议版本会演进。

实现前从 specification 首页进入当前稳定版本，不要只依赖本课程给出的历史路径。

### A2A

- [A2A documentation](https://a2a-protocol.org/)
- [A2A specification](https://a2a-protocol.org/latest/specification/)
- [A2A project repository](https://github.com/a2aproject/A2A)

### 本地推理与向量库

- [Ollama API](https://docs.ollama.com/api)
- [vLLM documentation](https://docs.vllm.ai/)
- [vLLM OpenAI-compatible server](https://docs.vllm.ai/en/latest/serving/openai_compatible_server.html)
- [pgvector repository and documentation](https://github.com/pgvector/pgvector)
- [PostgreSQL documentation](https://www.postgresql.org/docs/)

## 8. 基础论文

以下论文用于理解概念来源，不代表工程中必须复现训练。

- Transformer：[Attention Is All You Need](https://arxiv.org/abs/1706.03762)
- BPE 在 NMT 中的应用：[Neural Machine Translation of Rare Words with Subword Units](https://arxiv.org/abs/1508.07909)
- SentencePiece：[SentencePiece](https://arxiv.org/abs/1808.06226)
- RoPE：[RoFormer](https://arxiv.org/abs/2104.09864)
- MQA：[Fast Transformer Decoding: One Write-Head is All You Need](https://arxiv.org/abs/1911.02150)
- GQA：[GQA: Training Generalized Multi-Query Transformer Models](https://arxiv.org/abs/2305.13245)
- RAG：[Retrieval-Augmented Generation for Knowledge-Intensive NLP Tasks](https://arxiv.org/abs/2005.11401)
- LoRA：[LoRA: Low-Rank Adaptation of Large Language Models](https://arxiv.org/abs/2106.09685)
- QLoRA：[QLoRA: Efficient Finetuning of Quantized LLMs](https://arxiv.org/abs/2305.14314)
- RLHF/InstructGPT：[Training language models to follow instructions with human feedback](https://arxiv.org/abs/2203.02155)
- DPO：[Direct Preference Optimization](https://arxiv.org/abs/2305.18290)
- PagedAttention/vLLM：[Efficient Memory Management for Large Language Model Serving with PagedAttention](https://arxiv.org/abs/2309.06180)

## 9. 阅读方法

面对一个新术语，按下面顺序记录：

1. 它解决什么具体问题；
2. 输入和输出是什么；
3. 哪些部分是协议，哪些是实现；
4. 有哪些前提和失败模式；
5. 如何在 `agentgo` 中做一个最小实验；
6. 用什么测试证明理解没有停留在名词。

不要以背缩写结束学习。

能画出数据流、写出失败用例并解释取舍，才算掌握到工程可用程度。
