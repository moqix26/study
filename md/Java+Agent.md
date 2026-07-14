# Java + Agent 旧路线说明

> 本文件名为历史遗留入口。仓库中的 AI Agent 学习路线已于 2026-07-15 重构为 **Go-first**，不再以 Java、Spring AI 或 LangChain4j 为前置。

新的唯一入口：

- [Go AI Agent 学习库](../后端学习/AIAgent/README.md)
- [Go AI Agent 学习路线图](../后端学习/AIAgent/00-学习路线图与说明.md)
- [可运行 AgentGo 工程](../后端学习/AIAgent/Go工程/agentgo/README.md)

## 当前路线关系

```text
Go 基础与并发
  → HTTP / Gin
  → MySQL / Redis
  → 短链等普通后端项目
  → AI Agent / Tool / RAG 应用工程（选修扩展）
```

Java 资料仍保留在 [`后端学习/Java`](../后端学习/Java/00-学习路线图与说明.md)，可用于 MySQL、Redis、并发和工程原理参考，但不是新版 AIAgent 的代码主线。

## 为什么不再维护旧路线正文

- 用户当前职业主线是 Go 后端；
- 同时维护 Spring Boot 与 Gin 会分散学习时间；
- 原路线的 Spring AI/LangChain4j 版本和部分示例已经失效；
- AI 应用中的 Tool、RAG、评估和安全原则可以先用语言无关方式学习，再由 Go 落地；
- 新路线已经配套真实可测试的 Go 工程，不再依赖 Markdown 中无法验证的 Java 片段。

如果未来需要 Java Agent 实现，应基于当时的 Spring AI 官方版本另建分轨，并提供可编译工程和测试，不应恢复旧文档中的过时 API。

