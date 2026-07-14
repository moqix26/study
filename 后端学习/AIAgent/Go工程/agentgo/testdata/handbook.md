# AgentGo 演示手册

AgentGo 是本学习库的可运行 Go 基线工程。

默认模式使用 Mock Provider 和内存向量存储，因此不需要 API Key，也不需要 Docker。真实模型模式使用 OpenAI Responses API；模型名通过环境变量配置。

系统提供普通聊天、POST SSE 流式聊天、知识库写入与检索问答、受控 Tool Calling 和有界 Agent 循环。所有 `/api` 路由都要求 `X-User-ID`，用于演示服务端身份边界。

生产环境不能直接相信客户端提供的用户 ID，应由 JWT、Session 或网关认证结果写入服务端上下文。知识库检索必须带 tenant 条件，避免跨用户召回。
