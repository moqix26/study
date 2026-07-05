# LangChain 与 LlamaIndex 应用层

> **文件编码**：UTF-8。  
> **前置**：[12 HuggingFace Transformers](12-HuggingFace-Transformers入门.md)、[15 LoRA/PEFT](15-微调SFT与LoRA-PEFT.md)；RAG 概念可先读 [AIAgent 06 RAG 基础](../AIAgent/06-RAG检索增强生成基础.md)。  
> **对照**：[AIAgent 09 LangChain4j](../AIAgent/09-LangChain4j进阶.md)（Java 栈）；推理部署见 [LLMInfra 14 引擎导读](../LLMInfra/14-vLLM-TensorRT-LLM-llama.cpp架构导读.md)。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**LangChain / LlamaIndex = Python 侧「搭积木」框架**：把 LLM、Embedding、向量库、Prompt、Chain、Agent 串成 RAG 与对话应用——你专注业务编排，不必从零写 ingest/retrieve 流水线。

### 0.2 你需要提前知道什么

| 情况 | 建议 |
|------|------|
| 只会调 OpenAI API | 先读 [AIAgent 01](../AIAgent/01-大模型基础与API调用入门.md) |
| 不懂 Embedding / 向量检索 | 先读 [AIAgent 06～07](../AIAgent/06-RAG检索增强生成基础.md) |
| 已会 Spring AI RAG | ✅ 本章学 Python 等价物与差异 |
| 要做生产 Agent | 本章 + [AIAgent 05 ReAct](../AIAgent/05-Agent架构与ReAct模式.md) |

### 0.3 本章知识地图

- [ ] 能说出 LangChain 与 LlamaIndex 的定位差异
- [ ] 能画 Python RAG 索引/查询两阶段流水线
- [ ] 会用 LCEL（LangChain Expression Language）拼 Chain
- [ ] 会用 LlamaIndex 做 Document → Index → QueryEngine
- [ ] 能对比 Python LangChain vs Java Spring AI / LangChain4j
- [ ] 知道何时用框架、何时裸写 Transformers + FAISS

### 0.4 建议学习时长

| 阶段 | 内容 | 时间 |
|------|------|------|
| 概念 | §0～§2 | 45 分钟 |
| LangChain | §3～§5 | 2 小时 |
| LlamaIndex | §6～§7 | 1.5 小时 |
| 对照 Java | §8 | 45 分钟 |
| FAQ + 自测 | §9～§10 | 30 分钟 |

### 0.5 学完本章你能做什么

1. 用 50 行 Python 跑通「Markdown 知识库 → 向量检索 → 带引用回答」。
2. 向面试官解释：为什么 Java 产品用 Spring AI，研究/脚本用 LangChain。
3. 把本章 Demo 的 ingest 逻辑迁移到 [24 章项目](24-项目实战微调小型语言模型.md) 的 Gradio 前端。

### 0.6 与 LLMInfra / AIAgent 分工

```text
训练微调（本系列 12～15）→ 权重 .safetensors
  ↓
Python 应用层（本章 21）→ RAG / Agent 编排
  ↓
Java 产品 API（AIAgent 02～07）→ Spring Boot 对外服务
  ↓
C++ 推理引擎（LLMInfra 14～16）→ vLLM / TensorRT-LLM 高吞吐
```

---

## 1. 为什么需要应用层框架

裸写 RAG 需要自行处理：文档加载、分块、Embedding 批处理、向量库 CRUD、Prompt 模板、重试、流式输出、多轮记忆。LangChain 与 LlamaIndex 把这些封装成可组合模块。

| 维度 | 裸写 Transformers + FAISS | LangChain / LlamaIndex |
|------|---------------------------|-------------------------|
| 上手速度 | 慢，但可控 | 快，抽象多 |
| 定制深度 | 完全掌控 | 需读源码绕坑 |
| 生态集成 | 自己接 | 100+ VectorStore / Loader |
| 面试表述 | 「底层实现我清楚」 | 「工程效率与标准模式」 |

**原则**：算法/微调岗以 HF 为主；应用/全栈 Demo 用框架加速；生产 Java 栈见 AIAgent。

---

## 2. LangChain 核心概念

### 2.1 模块地图

```text
Document Loaders → Text Splitters → Embeddings → VectorStore
                                                      ↓
Retriever ←─────────────────────────────────── Query
     ↓
Prompt Template + LLM → Output Parser → Chain / Agent
```

### 2.2 LCEL 链式组合

LangChain Expression Language 用 `|` 管道连接 Runnable：

```python
from langchain_openai import ChatOpenAI, OpenAIEmbeddings
from langchain_community.vectorstores import FAISS
from langchain_core.prompts import ChatPromptTemplate
from langchain_core.output_parsers import StrOutputParser
from langchain_core.runnables import RunnablePassthrough

llm = ChatOpenAI(model="gpt-4o-mini", temperature=0)
prompt = ChatPromptTemplate.from_template(
    "仅根据资料回答。\n资料：{context}\n问题：{question}"
)

def format_docs(docs):
    return "\n\n".join(d.page_content for d in docs)

chain = (
    {"context": retriever | format_docs, "question": RunnablePassthrough()}
    | prompt
    | llm
    | StrOutputParser()
)
answer = chain.invoke("年假有多少天？")
```

### 2.3 Retriever 与 Agent

- **Retriever**：只负责「找资料」，不参与生成。
- **Agent**：LLM 决定调哪些 Tool（搜索、计算器、SQL），适合多步任务；对应 [AIAgent 05 ReAct](../AIAgent/05-Agent架构与ReAct模式.md)。
- **Memory**：`ConversationBufferMemory` 存多轮；生产需截断 + 摘要，见 [AIAgent 08](../AIAgent/08-对话记忆与会话管理.md)。

### 2.4 常用 VectorStore 对照

| 存储 | LangChain 类 | 场景 |
|------|--------------|------|
| FAISS | `FAISS` | 本地实验、无服务端 |
| Chroma | `Chroma` | 轻量持久化 |
| PGVector | `PGVector` | 与 Postgres 一体，对照 AIAgent 07 |
| Milvus | `Milvus` | 大规模、过滤丰富 |

---

## 3. LangChain RAG 完整 Demo

### 3.1 索引阶段

```python
from langchain_community.document_loaders import DirectoryLoader, TextLoader
from langchain_text_splitters import RecursiveCharacterTextSplitter
from langchain_openai import OpenAIEmbeddings
from langchain_community.vectorstores import FAISS

loader = DirectoryLoader("./kb-docs", glob="**/*.md", loader_cls=TextLoader)
docs = loader.load()
splitter = RecursiveCharacterTextSplitter(chunk_size=500, chunk_overlap=80)
chunks = splitter.split_documents(docs)
embeddings = OpenAIEmbeddings(model="text-embedding-3-small")
vectorstore = FAISS.from_documents(chunks, embeddings)
vectorstore.save_local("./faiss_index")
retriever = vectorstore.as_retriever(search_kwargs={"k": 4})
```

### 3.2 查询阶段（带引用）

```python
from langchain_core.runnables import RunnableParallel

rag_chain = RunnableParallel({
    "context": retriever,
    "question": RunnablePassthrough(),
}).assign(
    answer=(
        lambda x: prompt.invoke({
            "context": format_docs(x["context"]),
            "question": x["question"],
        })
        | llm
        | StrOutputParser()
    )
)
result = rag_chain.invoke("试用期内能否请年假？")
# result["context"] 即 citations 来源
```

### 3.3 流式输出

```python
for chunk in chain.stream("问题"):
    print(chunk, end="", flush=True)
```

对照 [AIAgent 03 SSE 流式](../AIAgent/03-流式对话与SSE实战.md)：Python 侧 `stream()` 等价于 Java `ChatClient.stream()`。

---

## 4. LlamaIndex 核心概念

LlamaIndex **以 Index 为中心**，更偏「数据 + 检索」；LangChain 更偏「Chain + Agent 编排」。

### 4.1 三步走

```python
from llama_index.core import VectorStoreIndex, SimpleDirectoryReader, Settings
from llama_index.embeddings.huggingface import HuggingFaceEmbedding
from llama_index.llms.huggingface import HuggingFaceLLM

Settings.embed_model = HuggingFaceEmbedding(model_name="BAAI/bge-small-zh-v1.5")
Settings.llm = HuggingFaceLLM(model_name="Qwen/Qwen2.5-0.5B-Instruct")

documents = SimpleDirectoryReader("./kb-docs").load_data()
index = VectorStoreIndex.from_documents(documents)
query_engine = index.as_query_engine(similarity_top_k=4)
response = query_engine.query("年假制度是什么？")
print(response.response)
print(response.source_nodes)  # 引用节点
```

### 4.2 Index 类型速查

| Index | 用途 |
|-------|------|
| VectorStoreIndex | 通用语义检索 |
| SummaryIndex | 全文摘要后问答 |
| TreeIndex | 层次化摘要，长文档 |
| KnowledgeGraphIndex | 实体关系，复杂知识图谱 |

### 4.3 Query Engine 进阶

- **RetrieverQueryEngine**：自定义 Retriever + Response Synthesizer。
- **SubQuestionQueryEngine**：复杂问题拆子问，类似 Agent 规划。
- **RouterQueryEngine**：多 Index 路由，对照 AIAgent Router 分意图。

---

## 5. LangChain vs LlamaIndex 选型

| 场景 | 推荐 |
|------|------|
| 快速 RAG Demo、多 Loader | LlamaIndex |
| 复杂 Agent、多 Tool、LCEL | LangChain |
| 本地 HF 小模型 | 两者均可 + `HuggingFacePipeline` |
| 与 Spring 团队对齐概念 | 先 LlamaIndex 理解 Index，再读 LangChain4j |
| 微调模型 serving | 接 [20 章 vLLM](20-vLLM-TGI与LMDeploy-Python侧.md) 或 [LLMInfra 14](../LLMInfra/14-vLLM-TensorRT-LLM-llama.cpp架构导读.md) |

---

## 6. 与 Java AIAgent 对照表

| 能力 | Python LangChain | Java Spring AI | Java LangChain4j |
|------|------------------|----------------|------------------|
| 文档加载 | `DirectoryLoader` | `MarkdownDocumentReader` | `DocumentLoader` |
| 分块 | `RecursiveCharacterTextSplitter` | `TokenTextSplitter` | `DocumentSplitter` |
| 向量库 | FAISS / Chroma / PGVector | `PgVectorStore` | `EmbeddingStore` |
| RAG 注入 | Retriever + Prompt | `QuestionAnswerAdvisor` | `RetrievalAugmentor` |
| 流式 | `chain.stream()` | `ChatClient.stream()` | `TokenStream` |
| Agent | `create_tool_calling_agent` | 自定义 ReAct | `AiServices` + `@Tool` |

**面试话术**：「业务 API 用 Spring AI 统一鉴权与监控；离线评估与数据脚本用 LangChain；模型权重与训练在本系列 12～15。」

---

## 7. 生产注意事项

### 7.1 与训练栈衔接

- Embedding 模型与微调 LLM **不必同厂**，但维度与语言需匹配。
- 领域 SFT 后 RAG 仍必要：权重不含最新私有文档，见 [15 章 LoRA](15-微调SFT与LoRA-PEFT.md)。

### 7.2 常见坑

| 问题 | 原因 | 处理 |
|------|------|------|
| 检索为空 | chunk 过大 / 问法差异 | 调 chunk_size、HyDE、重排 |
| 答案胡编 | Prompt 弱、temperature 高 | 拒答模板，对照 AIAgent 06 |
| 延迟高 | 串行 Embedding + LLM | 批处理、本地小 Embedding |
| 版本漂移 | LC 0.1→0.2 API 大变 | 锁定 `langchain-core` 版本 |

### 7.3 可观测性

对接 [22 章 wandb/mlflow](22-MLOps与实验跟踪wandb-mlflow.md) 记录：检索 hit@k、faithfulness、延迟 P99；对照 [AIAgent 15 可观测性](../AIAgent/15-LLM可观测性与评估体系.md)。

---

## 8. 进阶 RAG 方向（指针）

| 技术 | 说明 | 延伸阅读 |
|------|------|----------|
| Hybrid 检索 | BM25 + 向量 | [AIAgent 13](../AIAgent/13-RAG进阶-检索优化与评估.md) |
| Reranker | Cross-Encoder 重排 | 同上 |
| GraphRAG | 图结构检索 | LlamaIndex KG Index |
| Agentic RAG | 检索也交给 Agent | LangChain Agent + Retriever Tool |

---

## 9. FAQ

**Q1：LangChain 还值得学吗？**  
值得学 **概念与编排模式**；API 变动快，生产 Java 栈以 Spring AI 为主，Python 脚本与实验可用 LC/LlamaIndex。

**Q2：LlamaIndex 能替代 LangChain 吗？**  
纯 RAG 可以；复杂 Tool Agent 仍常选 LangChain。两者可混用：LlamaIndex 建 Index，LangChain 包 Agent。

**Q3：本地 Ollama 怎么接？**  
`ChatOllama(base_url="http://localhost:11434", model="qwen2.5:3b")`；Embedding 用 `OllamaEmbeddings` 或 HF 本地模型。

**Q4：和 24 章项目关系？**  
24 章微调后的模型可作为 LlamaIndex 的 `Settings.llm`；RAG 知识库与本章 Demo 结构相同。

**Q5：向量库选型与 AIAgent 07 一致吗？**  
原则一致：开发 FAISS/Chroma，生产 PGVector/Milvus，metadata 做租户隔离。

**Q6：如何做 citation？**  
LangChain 返回 `retriever` 的 `Document.metadata`；LlamaIndex 用 `response.source_nodes`。前端展示 filename + chunk_id。

**Q7：LangChain4j 与 LangChain Python 代码能复用吗？**  
不能直拷；概念（Retriever、Chain、Tool）可迁移，API 不同，见 [AIAgent 09](../AIAgent/09-LangChain4j进阶.md)。

**Q8：推理加速放哪一层？**  
应用层调 [20 章 vLLM OpenAI 兼容 API](20-vLLM-TGI与LMDeploy-Python侧.md)；内核优化见 [LLMInfra 07～16](../LLMInfra/07-大模型推理引擎架构概览.md)。

---

## 10. 闭卷自测

1. LangChain 与 LlamaIndex 各擅长什么？
2. LCEL 中 `|` 管道语义是什么？
3. RAG 索引阶段与查询阶段各做哪些事？
4. `as_retriever(search_kwargs={"k": 4})` 中 k 含义？
5. Spring AI 中哪组件等价于 LangChain Retriever + Prompt？
6. 为什么微调后仍需要 RAG？
7. FAISS 适合生产吗？替代方案？
8. 如何向 Java 面试官说明 Python 与 Spring 栈分工？

<details>
<summary>参考答案</summary>

1. LC 偏 Chain/Agent 编排；LlamaIndex 偏 Index/QueryEngine 与数据索引。
2. 前一个 Runnable 的输出作为后一个的输入，函数式管道组合。
3. 索引：加载→分块→Embedding→入库；查询：问句 Embedding→检索→拼 Prompt→LLM 生成。
4. Top-K，返回最相似的 4 个 chunk。
5. `QuestionAnswerAdvisor`（或 RetrievalAugmentor in LC4j）。
6. 权重不含最新私有文档，且长文档装不进 context；RAG 提供可更新外挂知识。
7. 不适合多租户生产；用 PGVector/Milvus + 持久化与 ACL。
8. Spring 负责产品 API、鉴权、监控；Python 负责训练、评估脚本、离线 RAG 实验；Infra C++ 负责高吞吐推理。

</details>

---

## 11. 下一章

[22 MLOps 与实验跟踪 wandb/mlflow](22-MLOps与实验跟踪wandb-mlflow.md)

并行复习：[AIAgent 06 RAG](../AIAgent/06-RAG检索增强生成基础.md)、[13 RAG 进阶](../AIAgent/13-RAG进阶-检索优化与评估.md)。
