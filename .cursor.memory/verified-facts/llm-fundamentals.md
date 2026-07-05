# 大模型岗/AI开发岗进阶章已核实事实（2026-06-30 核对）

> 来源：官方博客 + 论文 + 社区技术文。写 17～22 章引用，避免幻觉。

## 微调方法（17/20 章用）

- **LoRA**（Hu et al. 2021, arXiv 2106.09685）：权重更新具低秩性，把 ΔW 分解为 B(d×r)·A(r×k)，r ≪ d,k。冻结基座 W，只训 A、B。参数减少 99%+。
- **QLoRA**（Dettmers et al. 2023, arXiv 2305.14314）：在 4-bit NF4 量化基座上训 LoRA，加 paged optimizers + double quantization。单 48GB GPU 可微调 65B 模型。
- **SFT**：监督微调，教指令遵循，学习率 ~2e-4。
- **RLHF**：三阶段（SFT → 奖励模型训练 → PPO 策略优化）。复杂、不稳、4 模型同时在内存。ChatGPT/Claude/Gemini 背后技术。
- **DPO**（Rafailov et al. 2023, arXiv 2305.18290）：跳过奖励模型，直接用偏好对训练，分类式 loss。更稳、2 模型在内存。学习率 ~1e-6~1e-5，beta~0.1 控 KL 散度。2025-2026 成主流偏好对齐法。还有变体 IPO/KTO/ORPO。
- **GRPO**（DeepSeek 提出）：去掉奖励模型和 critic，组相对策略优化，能诱导 CoT/自纠。配合 RWR（Verifiable Rewards）。
- **典型栈**：base → SFT → DPO；RLHF 是升级路径非默认。决策树：要指令遵循→SFT；要对齐人类偏好→DPO 或 RLHF。

## vLLM 推理优化（19/21 章用）

- **PagedAttention**（Kwon et al. 2023, vLLM 论文）：KV cache 分块存储（仿 OS 虚拟内存/分页），非连续，内存浪费 <4%（vs 传统连续分配 60-80% 碎片）。比 HF Transformers 高 24x 吞吐。
- **Continuous Batching**：迭代级调度，每个 decode step 后动态加入/移除请求（vs 静态批处理等最慢序列完成）。吞吐 2-4x。三个队列：WAITING/RUNNING/SWAPPED。
- **Chunked Prefill**：长 prompt 切 ~512 token 片与 decode 交错，避免长 prompt 冻结所有 decode token。
- **Speculative Decoding**：小 draft 模型提议 k 个 token，大模型并行验证。方法：EAGLE/MTP/n-gram，`--speculative-config` 配置。加速取决于接受率。
- 模式：NAIVE（单请求）/ STATIC（padding 等待）/ CONTINUOUS / CONTINUOUS+CHUNKED（混合流量最优）。

## MCP 协议（21 章用）

- Anthropic 2024 开源，开放标准，"AI 的 USB-C"。灵感来自 LSP（语言服务器协议）。
- 客户端-服务器架构，JSON-RPC 2.0。
- 角色：Host（LLM 应用）/ Client（连接器）/ Server（提供能力）。
- 传输：stdio（本地进程）/ Streamable HTTP（远程，原 HTTP+SSE）。
- Server 三原语：**Tools**（可执行函数，如 API/查库）、**Resources**（数据/上下文，如文件/日志）、**Prompts**（可复用模板/few-shot）。
- Client 可提供：Sampling（让 server 触发 LLM）、Roots（URI 边界）、Elicitation（向用户要信息）。
- 握手：initialize/initialized → tools/list、resources/list、prompts/list 动态发现。变更通知 notifications/tools/list_changed。

## A2A 协议（21 章用）

- Google Cloud 提出，agent 间通信。
- 任务状态机：SUBMITTED → WORKING → COMPLETED/FAILED/CANCELED/REJECTED/INPUT_REQUIRED/AUTH_REQUIRED。
- **Agent Cards**：自描述（能力、协议、接受的请求），用于发现协作方。
- 绑定：JSON-RPC 2.0（常用）/ gRPC / HTTP+JSON。
- **MCP vs A2A**：MCP 连 agent↔tool（单 agent 扩展能力）；A2A 连 agent↔agent（团队协作）。互补，常组合：A2A 编排多 agent，每个 agent 内部用 MCP 连工具。

## 模型生态（22 章用，需以发布时实际版本为准）

- 开源主流：Llama（Meta）、Qwen（阿里）、DeepSeek、GLM（智谱）、Mistral。
- 闭源：GPT（OpenAI）、Claude（Anthropic）、Gemini（Google）。
- 趋势：MoE 架构（DeepSeek/Mixtral）、长上下文（百万 token）、多模态、推理模型（o1/DeepSeek-R1 类 CoT 强化）。
