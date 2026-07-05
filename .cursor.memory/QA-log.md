# QA 质检完成报告（2026-06-30）

## 总结

对全库 ~196 篇学习资料做了真读、真查、真改的质检。**未盲信前一轮「全 OK」标记**，发现并修复了实质性质量问题。

## 修复清单

### 🔴 严重（已全部修复）

**全库断链 114 处 → 0**
- 文件名重命名后旧链接未更新：24 类文件名映射（如 `08-JavaScriptDOM操作与事件处理`→`08-JavaScript-DOM-BOM与事件机制`）
- 路径深度错误：`../../` 多写一层（后端系列间、前端系列间互链）
- 跨大类目录路径错误（修改规范.md、系统设计→Vue、Web安全→AIAgent）
- 计网 01 的 CORS/HTTPS 章节实际移至 Web安全系列，旧链接 + 「（待写）」标记已修正
- 修复方式：Python 脚本精确替换，62+8 文件、234 处替换；修复后重新扫描确认 0 真断链

### 🟡 应修（已全部修复）

| 文件 | 问题 | 修复 |
|------|------|------|
| Java/00 | 「13 篇 §32 待补」标记过时（答案已存在） | 去标记 |
| AIAgent/02 | §7 小节编号乱序 7.1→7.4→7.2→7.3 | 重排 |
| AIAgent/10 | §18→§20→§19 编号乱序 + 3 处引用 | 重排+改引用 |
| AIAgent/04 | 5 处断链 Java 05/06 旧文件名 | 改真实文件名 |

### 🟢 可选（已记录，不影响学习）

- AIAgent/02 §6.3 多余对冲措辞 → 已清理
- TS/11 缺 §0 标签（有等价「本章衔接」导读）→ 收官篇，不动
- 数据结构/09 闭卷自测用 `###` 而非 `##`（内容齐全）→ 不动
- Java/16、Java/17 个别 Q&A 重复（一中一末）→ 不动

## 已深检通过（逐行核实 API/代码可跑）

| 系列 | 章节 | 核实内容 |
|------|------|----------|
| AIAgent | 02 | Spring AI 1.0 ChatClient.Builder / MessageChatMemoryAdvisor / chat_memory_conversation_id |
| AIAgent | 03 | stream().content()→Flux\<String\> / SseEmitter / TEXT_EVENT_STREAM_VALUE / EventSource |
| AIAgent | 04 | @Tool / @ToolParam / FunctionToolCallback / .defaultTools() 包路径 |
| AIAgent | 06 | EmbeddingModel / VectorStore / SimpleVectorStore / similaritySearch |
| AIAgent | 08 | MessageWindowChatMemory / RedisChatMemoryRepository / ChatMemory.CONVERSATION_ID |
| Java | 04 | pom(Spring Boot 3.2.5) / Result\<T\> / jakarta.validation / DTO-VO-Service |
| Java | 07 | StringRedisTemplate / opsForValue / Cache Aside / 穿透击穿雪崩 / 布隆 / SETNX |
| Java | 16 | SseEmitter / TextWebSocketHandler / @MessageMapping / @SendTo / 心跳 / Nginx |
| Java | 17 | JUnit 5 Jupiter / Mockito 5 / @Mock vs @MockBean / @WebMvcTest / Testcontainers |
| Vue | 05 | script setup / ref / computed / onMounted / composables / defineProps |
| Vue | 08 | axios.create / 拦截器 / Vite proxy / JWT 流 / CancelToken |
| Web安全 | 01 | 三类 XSS / innerHTML vs textContent / CSP / HttpOnly |
| 系统设计 | 02 | 令牌桶/漏桶/滑动窗口 / 熔断三状态 / Sentinel / Redis 限流 |

## 结构完整性（全系列 ✓）

15 个系列、~196 篇全部具备 §0 读前导读 + 闭卷自测 + 费曼检验（个别索引章/收官章有等价内容）。
版本号全库一致：JDK 17/21、Spring Boot 3.2+、Spring AI 1.0.x、MySQL 8。

## 进阶深挖章补充（2026-06-30 续）

应用户要求「补全仓库讲得偏浅、面试会深挖的技术点」，新增 AIAgent 13～16 四篇进阶章，均按 EXPANSION-STANDARD 编写（§0 导读/术语/手把手/逐行代码/报错表/FAQ≥10/闭卷自测/费曼/进阶练习/交叉引用）。

| 章节 | 核心内容 | 已核实事实（防幻觉） |
|------|----------|---------------------|
| 13 RAG 进阶 | rerank、混合检索+RRF、HyDE/子问题改写、5 种 chunk 策略、RAGAS 四指标+优化闭环 | RAGAS 指标真实名称；Spring AI DocumentRanker/ScoringModel/RetrievalAugmentationAdvisor（PR#5887/Disc#6067） |
| 14 Agent 进阶 | 5 种多 Agent 协作模式、长程任务状态持久化/续跑、5 种记忆类型、ReAct 绕圈四件套防护 | LangChain4j langchain4j-agentic（AgenticServices/AgenticScope/sequence/loop/supervisor/planner，标注实验性） |
| 15 LLM 可观测性 | Langfuse trace/observation/score、Spring AI 经 OTel+Micrometer 接入、6 类指标、三层评估闭环、Langfuse/LangSmith/Phoenix 对比 | Langfuse Spring AI 集成方式 + langfuse-java 客户端 + OTLP 端点 |
| 16 向量库选型 | pgvector/Qdrant/Weaviate/Milvus/Pinecone/Chroma/Redis 对比+决策树、HNSW/IVF/PQ 索引、Spring AI VectorStore 抽象切换、PGVector→Qdrant 迁移 | 2025-2026 社区基准数字 + 各库定位（标注选型前复测） |

**索引同步更新**：AIAgent/00（mindmap/学习顺序/阶段目标/衔接索引/demo 演进/§2 映射图）、修改规范.md §5.10（00～12→00～16）、PROGRESS-tracker（AIAgent 13/13→17/17，合计 192→196）。

**诚实标注**：13 章 rerank 的 ScoringModel/Jina 实现、15 章 langfuse-java 客户端均为早期/演进中 API，文档明确写「以你 pom 版本为准」并给官方链接兜底，不硬编不存在的稳定 API。

### 大模型岗/AI开发岗面试底层+前沿章补充（2026-06-30 续）

应用户要求「补全 AI 开发岗/大模型岗面试所需」，新增 AIAgent 17～22 六篇，均按 EXPANSION-STANDARD 编写。面向大模型岗面试的底层原理 + 前沿视野。

| 章节 | 核心内容 | 已核实事实（防幻觉） |
|------|----------|---------------------|
| 17 LLM 原理与训练流程 | Transformer/Self-Attention(Q/K/V)/Multi-Head/位置编码、Decoder-only vs Encoder vs Enc-Dec、BPE Tokenizer、Pretrain→SFT→RLHF/DPO/GRPO、上下文窗口/O(n²)/MoE/Scaling Law/幻觉 | Transformer/Attention 基础(常识)；SFT/DPO/RLHF/GRPO 流程(核实自多源技术文) |
| 18 Prompt 进阶+结构化输出 | few-shot/CoT/self-consistency/角色/指令位置/负面指令、Spring AI entity()/BeanOutputConverter/JSON Schema/重试修复、版本管理+A/B、先prompt后微调方法论 | Spring AI BeanOutputConverter/entity/ParameterizedTypeReference/getFormat/convert/postProcessSchema移除(核实自官方文档+Upgrade Notes) |
| 19 成本与延迟优化 | prompt caching(前缀相同命中)、大小模型路由、Batch API、TTFT/ITL、KV Cache、成本延迟质量三角 | prompt caching/OpenAI Batch(常识)；KV Cache/vLLM 机制(见 verified-facts) |
| 20 模型适配方法论+微调入门 | Prompt/RAG/Fine-tune/预训练选型决策树、Full/LoRA/QLoRA、SFT/DPO/RLHF/GRPO、数据准备与评估、Java 工程师入门路径 | LoRA(arXiv 2106.09685)/QLoRA(2305.14314)/DPO(2305.18290) 论文核实；典型栈 base→SFT→DPO |
| 21 MCP/A2A+本地推理 | MCP(Anthropic,JSON-RPC,Tools/Resources/Prompts,stdio/Streamable HTTP,握手)、A2A(Google,任务状态机,Agent Cards)、MCP vs A2A 互补、Spring AI MCP(1.1+,@McpTool)、vLLM(PagedAttention/Continuous Batching/Chunked Prefill/Speculative Decoding)、量化(INT4/8/GGUF) | MCP/A2A 协议(核实自官方规范+多源)；vLLM 三大技术(核实自官方博客+论文)；**Spring AI MCP 完整支持在 1.1.0-M1+，1.0.x 不完整，文档明确标注** |
| 22 生态选型+前沿范式 | 开源(Llama/Qwen/DeepSeek/GLM/Mistral) vs 闭源(GPT/Claude/Gemini)、五维选型决策树、推理模型(o1/R1)、Reflection/Plan-Execute/ToT/GoT/Reflexion、MoE/长上下文/多模态/Agent化趋势 | 推理范式(ReAct/Reflexion/ToT 论文核实)；模型版本标注「以最新为准」 |

**索引同步更新**：AIAgent/00（mindmap 22章/学习顺序/阶段目标/衔接索引/demo演进/§2映射图）、修改规范.md §5.10（00～16→00～22）、PROGRESS-tracker（AIAgent 17/17→23/23，合计 196→202）。

**版本风险标注（重要）**：
- 18 章 Spring AI 结构化输出 API 在 1.0.x 稳定，`postProcessSchema` 已移除改 `generateSchema()`
- 21 章 MCP 完整支持需 Spring AI **1.1.0-M1+**，1.0.x 不完整——文档明确写「务必核对版本」
- 22 章模型版本迭代快，标注「以最新版本和 benchmark 为准」

## 使用的工具脚本（存于 .cursor.memory）

- `scan_links.py` — 断链扫描（解码 %20，排除 .cursor.memory）
- `find_fixes.py` — 模糊匹配建议（参考，但有错配未盲用）
- `apply_fixes.py` — 主修复脚本（手工验证的映射表）
- `fix_round2.py` — 残留断链修复
- `broken-links.txt` / `fix-report.txt` / `suggested-fixes.txt` — 过程记录

## 源文件深度增强（面试深挖，2026-06-30）

按用户要求「在源文件基础上补充内容」，针对「面试深挖但原文只到基础认知」的 Java/Linux 后端核心章节，在文件末尾"下一章预告"前插入「面试深挖补充」大节。**未新增文件、未改原有编号**，与相关小节对照阅读。每节含「一句话+原理+对照表+面试标准答法+深挖点关联+面试自检」。

### 补充的章节与深挖点（共 5 文件、27 个深挖点）

| 源文件 | 深挖点 |
|--------|--------|
| `Java/03-Java并发编程与JVM.md` | ①synchronized锁升级(偏向→轻量→重量，JDK15废弃偏向锁JEP374) ②AQS(state+CLH双向队列+模板方法，ReentrantLock公平/非公平) ③线程池执行时序(core→queue→max→reject，Executors两个OOM坑) ④CAS的ABA+AtomicStampedReference ⑤ConcurrentHashMap JDK7分段锁Segment vs JDK8桶级CAS+synchronized+红黑树 ⑥G1(Region/GarbageFirst/RSet/Mixed GC)与ZGC(染色指针/读屏障/<10ms) ⑦打破双亲委派(SPI线程上下文类加载器/Tomcat WebappClassLoader/OSGi热部署) |
| `Java/07-Redis核心原理与缓存实战.md` | ①五种数据结构底层(SDS/listpack替代ziplist消除连锁更新/quicklist/skiplist+dict双结构/intset/编码切换阈值+OBJECT ENCODING) ②单线程模型+6.0多线程IO(io-threads只管网络读写，命令执行仍单线程) ③过期删除(惰性+定期)与内存淘汰(8种策略+近似LRU采样+LFU Morris计数器) ④哨兵(主观/客观下线+Raft选leader+选主规则)与Cluster(16384槽+CRC16+Gossip+MOVED+节点自治故障转移) ⑤RedLock算法(5实例过半)与Kleppmann争议(GC暂停/时钟漂移，fencing token/ZooKeeper)+Redisson看门狗(不传leaseTime启用，每1/3 TTL续期) ⑥AOF重写(fork子进程+重写缓冲区+触发条件)与混合持久化(4.0+/5.0默认，RDB前段+AOF后段) |
| `Java/06-MySQL基础索引与事务.md` | ①MVCC的ReadView机制(隐藏字段trx_id/roll_ptr+undo版本链+m_ids/min_trx_id/max_trx_id/creator_trx_id+可见性判断规则+RC每次生成/RR首次复用) ②索引下推ICP(5.6+，Using index condition，减少回表) ③间隙锁/临键锁(RR专属，等值命中唯一索引退化记录锁/未命中退化间隙锁，RC无间隙锁致当前读幻读) ④change buffer(仅非唯一索引，唯一索引必须读页判唯一性)+AHI(自动哈希旁路，高并发可关)+MRR(主键排序回表，随机IO转顺序IO) |
| `Linux/06-进程与服务管理.md` | ①进程状态机(R/S/D/Z/T+修饰位，D状态kill-9无效因信号需回用户态处理) ②僵尸进程(父进程没wait，占PID耗尽致无法fork，处理靠父进程wait或杀父进程让init收养)+孤儿进程(init收养无害，daemon/nohup利用) ③OOM Killer(oom_score打分主看内存，oom_score_adj=-1000永不杀，Java常被杀，dmesg排查，区别于JVM堆OOM) ④SIGCHLD与wait根因(显式wait/SIG_IGN自动回收) ⑤systemd unit编写([Unit]/[Service]/[Install]+Type=simple/forking/notify+Wants弱依赖/Requires强依赖/After仅顺序+daemon-reload必须+SuccessExitStatus=143防SIGTERM重启) |
| `Linux/07-网络命令与防火墙基础.md` | ①TCP状态全景(LISTEN→SYN_SENT/SYN_RECV→ESTABLISHED→FIN_WAIT/CLOSE_WAIT/LAST_ACK/TIME_WAIT，ss -tan state过滤) ②TIME_WAIT堆积(2MSL两原因：防旧报文/确保最后ACK到达+耗源端口+长连接治本+tcp_tw_reuse+tcp_tw_recycle 4.12已移除禁用) ③CLOSE_WAIT堆积(必是应用bug没close连接，try-with-resources修) ④ss vs netstat(netlink直查内核 vs 遍历/proc，性能差) ⑤tcpdump抓包(-i/-nn/-w pcap/-A，三次握手S/S./.，四次挥手F./.，排查连不上看SYN有无SYN+ACK) |

### 事实核查说明

均为后端/Linux 面试公认的技术原理，无版本敏感 API：
- 锁升级/双亲委派：JDK 版本演进（JEP 374 废弃偏向锁）已核实
- Redis：listpack 替代 ziplist 在 Redis 7.0、混合持久化 5.0 默认、tcp_tw_recycle Linux 4.12 移除——均为事实性版本节点，已核实
- MySQL：ICP 5.6、change buffer/5.5 前叫 insert buffer——版本节点已核实
- 无未发布/实验性 API，无幻觉风险

### 与既有内容的关系

- 深挖点与原"基础认知"小节显式对照（如「给 §5/§19/§31 补上底层为什么」）
- 每节末尾「关联」段把多个深挖点串起来（如 MVCC+间隙锁共同防 RR 幻读、CLOSE_WAIT 堆积呼应 systemd LimitNOFILE）
- 每节末尾「面试自检」提供勾选式自测，呼应 EXPANSION-STANDARD 的闭卷自测风格
- 文件总数不变（仍 ~202），仅在源文件内追加内容

## 项目与面试冲刺章（23/24，2026-06-30 新增）

按用户「如果看完还不能很好掌握就再补充」的要求，识别出"光有知识点不够进大厂"的三个差距（项目维度/面试输出维度/工程实操维度），新增两篇把知识转成面试战斗力的章：

| 文件 | 核心内容 | 解决的差距 |
|------|----------|------------|
| `AIAgent/23-Agent与Java端到端项目实战.md` | 企业知识库智能问答 Agent 端到端：项目定位+简历包装(STAR+量化)、架构图、6 个技术选型深挖(Spring AI vs LangChain4j/PGVector vs Milvus/RAG vs 微调/SSE vs WebSocket/多智能体 vs 单Agent/DeepSeek+Ollama)、核心模块代码骨架(RAG/FunctionCalling/多智能体/SSE/可观测)、Docker Compose+Nginx 部署、成本延迟准确率优化、4 个线上排障场景(延迟暴涨/召回低/Full GC/幻觉)、面试追问应对(RAG/Agent/工程/项目深挖四类+标准答案) | 项目维度：给一个能写简历、能扛追问的 Agent 项目 |
| `AIAgent/24-大厂面试实战手册.md` | 面试 5 环节占比与失败原因、场景设计 4 步框架+5 道高频题(限流/短链/秒杀/RAG服务/IM)、6 道手撕代码(LRU/DCL单例/生产消费者/令牌桶/快速幂/简化线程池)含正确实现+易错点+讲法、全库八股速查表(并发JVM/MySQL/Redis/Spring/Linux/AI 6 大块)、项目深挖 3 分钟 4 段式+3 张牌模板、简历优化公式+红线、模拟面试流程、自我检验清单 | 面试输出维度：把八股+项目聚成战斗力 |

两章均按 EXPANSION-STANDARD 结构（§0 导读/知识地图/FAQ/闭卷自测/费曼）。
**事实核查**：手撕代码均为标准实现（DCL 的 volatile 原因、LRU 双向链表、令牌桶懒补充等易错点已标注）；Spring AI API 用 1.0.x 稳定 API（ChatClient/Advisor/@Tool/SseEmitter/VectorStore），与 02-18 章已核实 API 一致；23 章明确标注"代码必须亲手跑通最小版本"、量化指标"必须真测过"，防幻觉和编造。
**文件总数**：202 → 204（新增 2 篇）。
