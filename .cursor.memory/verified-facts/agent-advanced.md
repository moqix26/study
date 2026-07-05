# 进阶章已核实事实（2026-06-30 核对）

> 来源：官方文档 + GitHub PR/Discussion + 社区基准。写文档时引用，避免幻觉。

## RAGAS 指标（Python 库，LLM-as-judge）

四个核心指标：
- `faithfulness` — 答案是否基于检索上下文（量化幻觉）
- `answer_relevancy` — 答案是否切题
- `context_precision` — 检索 chunk 是否相关且排序合理
- `context_recall` — 是否检索到所有必要 chunk（需 ground_truth）

用法：
```python
from ragas import evaluate
from ragas.metrics import faithfulness, answer_relevancy, context_precision, context_recall
# dataset 需含 question, answer, contexts 列；context_recall 还需 ground_truth
scores = evaluate(dataset, metrics=[...])
```
注意：RAGAS 是 Python 生态，Java 项目可调子进程或用 HTTP 服务封装。faithfulness/LLMContextRecall 是高成本指标（多次 LLM 调用）。

## Spring AI rerank（1.0.x，版本演进中）

- `DocumentPostProcessor` 接口：对检索结果做后处理（rerank/过滤/压缩）
- `DocumentRanker` 接口：`org.springframework.ai.rag.postretrieval.ranking`，`rank(Query, List<Document>) -> List<Document>`
- `RetrievalAugmentationAdvisor`：模块化 RAG Advisor，需 `spring-ai-rag` 依赖
- `ScoringModel` 抽象 + `JinaScoringModel` 实现（Jina `/v1/rerank` 端点）
- `ScoringDocumentPostProcessor` 可接入 Advisor
- 配置：`spring.ai.jina.scoring.*`
- 注意：ScoringModel/Jina 实现可能在 1.0.x 后期或 1.1；文档写"以你 pom 版本为准"，给官方链接

## Langfuse（LLM 可观测性，开源可自部署）

- Spring AI 集成走 OpenTelemetry + Micrometer tracing
- 依赖：Spring Boot Actuator + Micrometer OTel
- OTLP endpoint：`https://cloud.langfuse.com/api/public/otel`（Basic Auth = public key/secret key）
- Java 客户端：`com.langfuse:langfuse-java`（GitHub Package Registry），可拉 prompt、提交 score
- 核心概念：trace（一次请求链路）、observation（span/generation/event）、score（Numeric/Categorical/Boolean）
- score 通过 trace_id 关联，可事后补打分
- 自部署：`http://localhost:3000`；Cloud：EU/US/JP/HIPAA 区

## 向量库选型（2025-2026 社区基准）

| 库 | 语言 | 适用规模 | 强项 | 弱点 |
|----|------|----------|------|------|
| pgvector | C(Postgres扩展) | <50M | 已有PG零成本、SQL join | >50M 性能降 |
| Qdrant | Rust | <100M | 低延迟、filtered search 强、单容器自托管 | 大规模需分片 |
| Weaviate | Go | <200M | hybrid search(BM25+向量)原生、内置vectorizer | 调参复杂 |
| Milvus | Go/C++ | billion级 | 分布式、K8s原生、多索引、GPU | 运维重，<100M不值 |
| Pinecone | 托管 | 任意 | 零运维、SLA | 贵、非开源 |
| Chroma | Python | <1M原型 | 简单 | 生产弱 |

选型口诀：已有PG用pgvector；新项目中等规模用Qdrant；要hybrid用Weaviate；billion级有运维用Milvus；要零运维用Pinecone；原型用Chroma。
