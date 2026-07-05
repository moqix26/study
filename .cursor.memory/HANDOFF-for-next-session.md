# 接力说明（2026-06-30 QA + 两轮进阶章 + 源文件深度增强）

## 状态
全库质检完成。~202 篇 md 经真读+真查，修复 114 处断链 + 5 处编号/标记问题。
断链扫描归零（剩 2 处为 .cursor.memory 假阳性）。
AIAgent 新增 13～22 共十篇进阶章：
- 13～16：RAG/Agent/可观测/向量库「工程深挖」
- 17～22：LLM 原理/Prompt/成本/微调/MCP本地/生态范式「大模型岗底层+前沿」
均已核实事实、按 EXPANSION-STANDARD 编写。

另：在 5 个 Java/Linux **现有源文件**末尾「下一章预告」前插入「面试深挖补充」大节，补足"面试深挖但原文只到基础认知"的底层原理（未新增文件、未改编号，与原小节对照阅读）：
- `Java/03 并发JVM`：锁升级/AQS/线程池时序/CAS ABA/CHM JDK7vs8/G1与ZGC/打破双亲委派
- `Java/07 Redis`：数据结构底层/单线程+6.0多线程IO/过期删除+淘汰/哨兵+Cluster/RedLock争议+Redisson看门狗/AOF重写+混合持久化
- `Java/06 MySQL`：MVCC的ReadView/索引下推ICP/间隙锁临键锁/change buffer+AHI+MRR
- `Linux/06 进程`：进程状态机/僵尸+孤儿进程/OOM Killer/SIGCHLD/systemd unit编写
- `Linux/07 网络`：TCP状态全景/TIME_WAIT堆积/CLOSE_WAIT堆积/ss vs netstat/tcpdump抓包
共 27 个深挖点，每节含「一句话+原理+对照表+面试标准答法+关联+面试自检」。详见 QA-log.md「源文件深度增强」。

另：新增 AIAgent 23/24 两篇「项目与面试冲刺」章（应对"光有知识点不够进大厂"的差距）：
- 23 端到端项目实战：企业知识库智能问答 Agent，覆盖架构/选型深挖/编码/部署/优化/排障/面试追问应对，给一个能写简历、能扛追问的项目。
- 24 大厂面试实战手册：场景设计4步框架+5高频题、6道手撕代码、全库八股速查、项目深挖3分钟4段式+3张牌、简历优化、模拟面试。
AIAgent 系列从 23/23 升到 25/25，全库 202→204 篇。详见 QA-log.md「项目与面试冲刺章」。

## 用户主线
Agent+Java+Vue+HTML — 全部 ✅，API/代码已逐行核实无幻觉。
面向大模型岗/AI开发岗面试的底层原理(17) + 工程技巧(18/19) + 适配方法论(20) + 新兴协议与部署(21) + 生态视野(22) 已补全。

## 质检日志
详见 `QA-log.md`（含修复清单、已核实章节、两轮进阶章补充记录）。
进阶章核实事实存于 `verified-facts/agent-advanced.md` 和 `verified-facts/llm-fundamentals.md`。

## 若继续
- 单章精修：用户报章节+自测题号
- 版本更新：Spring AI/LangChain4j/Langfuse/MCP 大版本变更时核对 verified-facts
- 12 章面试总表可补「13～22 进阶考点」索引（可选，不阻塞）
- 新增章节：按 修改规范.md §4.2.1
- 断链复扫：`python .cursor.memory/scan_links.py`

## 进阶章 API 版本提醒（重要）
- 13 章：Spring AI `ScoringModel`/`JinaScoringModel` 在 1.0.x 不同小版本可用性不同，有「自封装 HTTP rerank」兜底
- 15 章：`langfuse-java` 为 0.0.x 早期版本，GitHub Package Registry 托管
- 18 章：Spring AI 结构化输出 `postProcessSchema` 已移除改 `generateSchema()`
- 21 章：**MCP 完整支持需 Spring AI 1.1.0-M1+，1.0.x 不完整**——最关键的版本红线
- 22 章：模型版本迭代快，以最新版本和 benchmark 为准

## 小瑕疵（不影响学习，可后续处理）
- TS/11 缺 §0 标签（有等价导读）
- 数据结构/09 闭卷自测用 ### 而非 ##

## 学习提醒
资料厚 ≠ 学会；必须手敲 + 闭卷自测 ≥7/10 再进下一章。
进阶章建议在对应基础章过关后再学：
- 06/07 → 13、16
- 05 → 14
- 11 → 15
- 01 → 17、18
- 02/11 → 19、21
- 17 → 20、22
