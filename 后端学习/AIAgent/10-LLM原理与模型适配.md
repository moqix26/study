# 10 LLM 原理与模型适配：知道系统为什么慢、为什么错、该改哪一层

```yaml
last_reviewed: 2026-07-15
focus: 面向 Go Agent 工程的必要原理
claim_policy: 不对闭源专有模型结构做无官方来源断言
```

## 1. 学习目标

这一章不要求你训练基础模型。

目标是让你能做出工程判断：

- 为什么长输入会拖慢首 token；
- 为什么流式输出不能减少模型总计算；
- KV cache 保存了什么，为什么会占显存；
- tokenizer 为什么会影响成本、截断和切块；
- RAG、Prompt、Tool 和微调分别解决什么；
- LoRA、SFT、DPO、RLVR 为什么不是同一层概念；
- 怎样为模型更换建立可复现评估。

## 2. 从自回归生成看全链路

大多数文本生成模型按条件概率逐 token 生成：

```text
P(x1, x2, ..., xn)
= P(x1) · P(x2 | x1) · ... · P(xn | x1...x(n-1))
```

一次 Agent 请求可以粗略画成：

```text
文本 / 工具定义 / 检索片段
        ↓ tokenizer
token IDs
        ↓ Transformer prefill
首个 token 的概率分布 + KV cache
        ↓ sampling / decoding
下一个 token
        ↓ 反复 decode
输出 token IDs
        ↓ tokenizer decode
文本或结构化工具参数
```

模型没有直接“执行工具”。

它生成符合某种协议的结构化内容，由 Go 应用校验、授权并执行。

## 3. Tokenizer：最先发生、最容易被忽视

### 3.1 Token 不是字，也不是词

Tokenizer 把文本编码成词表中的整数 ID。

一个 token 可能是：

- 一个常见单词；
- 单词片段；
- 一个或多个汉字；
- 标点和空格组合；
- 字节回退片段；
- 特殊控制符号。

不同模型的 tokenizer 和词表不同。

不能用固定的“一个汉字等于几个 token”做预算。

必须使用目标模型的官方计数方法、tokenizer 或实际 API usage。

### 3.2 常见子词方法

#### BPE

BPE 类方法从较小单位开始，反复合并高频组合。

工程上要理解：

- 常见片段可能只占一个 token；
- 罕见字符串可能被拆得很细；
- 空格、大小写和 Unicode 形式会影响切分。

#### Unigram

Unigram 类方法从候选子词集合中选择概率较好的分词路径，并在训练中裁剪词表。

#### SentencePiece

SentencePiece 是直接在原始文本上训练和编码的工具/方法体系，可承载 BPE 或 unigram 等模型。

不要把 SentencePiece 与 BPE 简化成互斥的同级算法名。

### 3.3 Tokenizer 对 Agent 的影响

1. 工具 schema 也占上下文；
2. RAG 片段越多，输入 token 越多；
3. JSON key 重复会增加 token；
4. 截断可能切掉系统规则或关键证据；
5. 奇怪 ID、代码和 Base64 可能非常耗 token；
6. 模型迁移后，同一 prompt 的 token 数可能改变。

### 3.4 Go 服务中的预算

不要等 provider 返回“上下文过长”才处理。

预算应预留：

```text
系统与开发者规则
+ 用户消息
+ 会话历史
+ 工具定义
+ RAG 片段
+ 协议包装
+ 期望最大输出
<= 模型当前可用上下文
```

计数无法精确时，使用保守上限，并在真实响应记录实际 usage。

## 4. Transformer 的最小结构

经典 Transformer 论文：[Attention Is All You Need](https://arxiv.org/abs/1706.03762)

现代 LLM 具体结构会有差异，但常见积木包括：

- token embedding；
- 位置信息；
- 多层 attention；
- 前馈网络；
- 残差连接；
- normalization；
- 输出投影到词表 logits。

不要从某个开源实现推断所有闭源模型采用完全相同细节。

## 5. Self-Attention

输入隐藏状态经过线性投影得到 Query、Key、Value：

```text
Q = XWq
K = XWk
V = XWv
```

缩放点积注意力：

```text
Attention(Q, K, V)
= softmax(QKᵀ / √d_k + mask) V
```

直觉：

1. Query 表示当前位置想找什么；
2. Key 表示各历史位置可被怎样匹配；
3. 点积形成相关分数；
4. softmax 变成权重；
5. 对 Value 加权聚合。

### 5.1 Causal Mask

自回归模型不能在生成当前位置时偷看未来 token。

Causal mask 把未来位置屏蔽。

训练时可以并行计算一整段位置，但每个位置只能使用允许看到的前文。

### 5.2 Multi-Head Attention

多个头使用不同投影并行计算，再合并结果。

“每个头一定负责某种人类可命名能力”不是可靠工程假设。

### 5.3 MHA、MQA、GQA

- MHA：多个 query 头各有对应 K/V 头；
- MQA：多个 query 头共享更少的 K/V；
- GQA：query 头分组共享 K/V。

后两者可降低 KV cache 压力，但具体模型使用哪种结构，应查模型官方技术报告或配置。

参考：

- [Fast Transformer Decoding: One Write-Head is All You Need](https://arxiv.org/abs/1911.02150)
- [GQA: Training Generalized Multi-Query Transformer Models](https://arxiv.org/abs/2305.13245)

## 6. 位置信息

Attention 本身不天然理解 token 顺序。

模型需要注入位置信息。

常见方法包括绝对位置 embedding、相对位置方法和 RoPE。

RoPE 论文：[RoFormer](https://arxiv.org/abs/2104.09864)

工程边界：

- “标称支持很长上下文”不等于你的任务在远距离依赖上仍准确；
- 超出训练或适配范围的扩展方法可能有质量代价；
- 长上下文必须用真实数据测试检索、引用和指令保持能力。

## 7. 前馈网络、残差与归一化

Attention 负责位置间信息交换，前馈网络对每个位置做非线性变换。

残差连接让每层在已有表示上学习增量，也帮助深层训练。

归一化帮助控制数值尺度和优化稳定性。

不同模型可能使用不同激活、norm 类型、层顺序和稀疏专家结构。

没有公开来源时，不对专有模型作具体断言。

## 8. 从 logits 到下一个 token

最后一层输出被映射到词表维度，得到 logits。

经过采样策略选择下一个 token。

### 8.1 Temperature

Temperature 改变分布尖锐程度。

- 较低：通常更集中；
- 较高：通常更多样；
- 设为 0 的精确行为由 API 实现定义，不假定所有供应商一致。

### 8.2 Top-p 与 Top-k

- Top-p：从累计概率达到阈值的候选集合采样；
- Top-k：只在概率最高的 k 个候选中采样。

供应商可能只暴露部分参数或对组合做限制。

结构化工具参数通常更重视稳定性，但参数调低不能替代 schema 校验。

### 8.3 确定性边界

即使参数相同，仍可能因：

- 服务端实现更新；
- 硬件和数值计算；
- 并行调度；
- 隐藏系统配置；
- 模型版本变化；

得到不同结果。

测试应断言结构和业务性质，不要处处断言完整文本逐字一致。

## 9. Prefill 与 Decode

这是理解延迟最重要的分界。

### 9.1 Prefill

Prefill 对已有输入 token 进行计算，并为每层建立历史 K/V。

特点：

- 输入位置可高度并行；
- 长 prompt 增加计算量；
- 影响 TTFT；
- 工具定义、会话和 RAG 都属于输入负担。

减少无效上下文，通常能改善 TTFT 和成本。

### 9.2 Decode

Decode 每一步基于已有上下文生成一个新 token，并把新 K/V 追加到 cache。

特点：

- 时间上有自回归依赖；
- 每个请求逐 token 前进；
- 影响 TPOT 和总输出时间；
- 输出越长，decode 步骤越多。

流式传输让用户更早看到输出，但不会消除 decode 计算。

### 9.3 指标不要混用

| 指标 | 主要反映 | 常见误区 |
|---|---|---|
| TTFT | 排队、网络、prefill、首事件 | 用总耗时冒充首 token |
| TPOT | decode 阶段速度 | 不说明是否包含网络缓冲 |
| End-to-end | 用户等待总时间 | 不拆工具和检索耗时 |
| Tokens/s | 吞吐或单请求生成速率 | 不说明 input/output 与并发口径 |

一次 Agent 请求还要单独记录：

```text
retrieval_ms
llm_prefill_visible_ms（若 provider 可观测）
first_event_ms
tool_ms
llm_rounds
total_ms
```

## 10. KV Cache

若每次生成新 token 都重新计算全部历史 K/V，decode 会大量重复工作。

KV cache 保存各层历史 token 的 Key 和 Value。

概念上的存储量与以下因素近似成正比：

```text
层数
× 历史 token 数
× KV head 数
× head dimension
× K 和 V 两份
× 每个元素字节数
× 并发序列数
```

这是理解关系的公式，不是可直接套所有模型的精确显存计算器。

还要考虑：

- block/page 管理开销；
- allocator 碎片；
- prefix cache；
- beam 或并行序列；
- tensor parallel 分片；
- 模型实现的数据布局。

### 10.1 KV cache 与模型权重不是一回事

- 权重：模型参数，加载后通常长期占用；
- KV cache：随活跃请求、上下文和生成长度变化。

量化模型权重不意味着 KV cache 一定以同样位宽量化。

### 10.2 为什么长上下文会降低并发

每个活跃序列占用更多 KV cache，能同时容纳的序列可能减少。

因此“最大上下文长度”与“目标并发”必须一起压测。

### 10.3 PagedAttention 与连续批处理

PagedAttention 关注更灵活地管理 KV cache block，减少传统连续分配带来的浪费。

论文：[Efficient Memory Management for Large Language Model Serving with PagedAttention](https://arxiv.org/abs/2309.06180)

Continuous batching 允许调度器在运行中加入新请求、移除完成请求。

它可以提高资源利用率，但也带来排队和公平性取舍。

## 11. 先定位问题属于哪一层

| 需求 | 优先手段 | 原因 |
|---|---|---|
| 回答需要最新私有文档 | RAG / Tool | 知识需外部更新与引用 |
| 输出格式不稳定 | schema + 校验 + prompt | 先收紧接口契约 |
| 需要实时余额或库存 | Tool | 必须读取权威系统 |
| 固定术语和风格 | prompt，必要时 SFT | 先用低成本方案验证 |
| 大量稳定任务模式需要内化 | SFT/PEFT 候选 | 行为模式而非实时事实 |
| 偏好两种都正确回答中的一种 | preference optimization 候选 | 需要偏好数据 |
| 数学/代码结果可自动验证 | RLVR 候选 | 奖励可程序判定 |
| 请求慢 | 拆 TTFT、TPOT、工具、检索 | 不要直接换模型猜测 |

## 12. RAG 与微调不是二选一口号

### 12.1 RAG 擅长什么

- 注入频繁更新的知识；
- 使用企业私有资料；
- 给出来源；
- 删除或更正文档后可重建索引；
- 按权限过滤资料。

RAG 的主要风险：

- 召回不到；
- 召回错误；
- 切块破坏上下文；
- 检索内容 prompt injection；
- 模型忽略证据；
- 引用与结论不匹配。

### 12.2 微调擅长什么

- 稳定任务行为；
- 术语、风格和格式习惯；
- 某类输入到输出的映射；
- 在合适数据上适配特定工作模式。

微调不适合充当频繁更新事实的数据库。

它还会带来数据、训练、部署、回归和版本管理成本。

### 12.3 可以组合

常见组合：

```text
SFT/偏好优化后的模型
  + 运行时 RAG
  + 权威系统 Tool
  + 服务端校验
```

微调改善行为，RAG 提供资料，Tool 取得实时事实。

## 13. 两个正交维度

最容易混淆的地方是把“怎样更新参数”和“用什么训练目标”混成一张表。

### 13.1 参数更新方式

- Full fine-tuning：更新大量或全部参数；
- PEFT：只更新较小参数集合；
- LoRA：用低秩增量适配选定权重；
- QLoRA：在量化基座上训练 LoRA adapter 的一类方案。

### 13.2 训练目标/数据方式

- SFT：输入到目标输出的监督学习；
- DPO：用偏好对直接优化偏好目标；
- RLHF：从人类反馈建模奖励，再以强化学习优化的一类流程；
- RLVR：使用可验证奖励进行强化学习的范式。

可以出现：

- LoRA + SFT；
- LoRA + DPO；
- full fine-tuning + SFT；
- 某种参数高效方式 + 某种强化学习目标。

所以“LoRA 与 DPO 选哪个”通常是错误问题。

一个回答“改哪些参数”，另一个回答“用什么偏好目标训练”。

## 14. LoRA 与 QLoRA

LoRA 论文：[LoRA](https://arxiv.org/abs/2106.09685)

对某个权重更新，可用低秩矩阵表示增量：

```text
W' = W + ΔW
ΔW = BA
rank(BA) 远小于 W 的完整维度
```

训练时通常冻结基座权重，只优化低秩参数。

优点可能包括：

- 可训练参数更少；
- adapter 文件较小；
- 同一基座可管理多个 adapter。

但不代表：

- 训练不需要显存；
- 质量一定等同 full fine-tuning；
- 任意任务都适合；
- adapter 可以跨不同基座随意混用。

QLoRA 论文：[QLoRA](https://arxiv.org/abs/2305.14314)

QLoRA 的关键方向是在量化基座上训练 LoRA，从而降低训练资源需求。

实现时仍需理解量化格式、计算 dtype、optimizer state、序列长度和框架兼容性。

## 15. SFT

SFT 数据通常形如：

```json
{
  "input": "用户和上下文",
  "target": "期望输出"
}
```

数据质量重点：

- 任务定义一致；
- 输出真正确；
- 格式稳定；
- 难例和拒绝样例充分；
- 训练/验证/测试去重；
- 隐私和许可合规；
- 不把测试答案泄漏进训练。

更多低质量样例不一定优于更少高质量样例。

## 16. DPO

DPO 论文：[Direct Preference Optimization](https://arxiv.org/abs/2305.18290)

典型数据是同一 prompt 下的偏好对：

```json
{
  "prompt": "……",
  "chosen": "更符合偏好的回答",
  "rejected": "较差回答"
}
```

DPO 适合学习相对偏好，但数据必须说明“为什么 chosen 更好”。

若偏好标注本身不一致，训练会内化噪声。

DPO 不自动提供事实更新，也不能替代工具授权。

## 17. RLHF 与 RLVR

RLHF 不是单一固定算法。

典型流程可能包括：

1. 收集人类偏好；
2. 训练奖励模型或构造反馈信号；
3. 用强化学习更新策略；
4. 用约束抑制策略偏离过大；
5. 独立评估质量和安全。

参考：[Training language models to follow instructions with human feedback](https://arxiv.org/abs/2203.02155)

RLVR 的核心是奖励可被程序可靠验证，例如：

- 数学最终答案；
- 单元测试；
- 编译是否通过；
- 形式化约束；
- 游戏或模拟器结果。

边界：

- 可验证不等于奖励无漏洞；
- 模型可能利用 verifier 缺陷；
- 只奖励最终答案可能忽视过程风险；
- 主观写作或开放事实难以只靠二元 verifier；
- 训练成功不能替代分布外评估。

某个具体策略优化算法不等同于 RLVR 这个奖励范式。

## 18. 适配决策流程

```text
问题是否来自缺少实时/私有事实？
├─ 是 → 先做 Tool 或 RAG
└─ 否
   ↓
是否是输出 schema/流程约束问题？
├─ 是 → prompt + schema + validator + retry policy
└─ 否
   ↓
是否有稳定、大量、高质量示例？
├─ 否 → 先收集失败集，不训练
└─ 是
   ↓
是模仿目标输出，还是学习相对偏好？
├─ 模仿 → SFT 候选
└─ 偏好 → DPO/RLHF 候选
   ↓
结果能否可靠程序验证？
├─ 是 → 进一步评估 RLVR
└─ 否 → 人工偏好与离线评估
```

每一步都先建立 baseline。

否则无法证明训练是否真的改善。

## 19. 评估设计

### 19.1 分层评估

```text
Tokenizer / context
Provider API contract
结构化输出
Tool selection
Tool execution
Retrieval
Grounded answer
End-to-end task
安全与成本
```

### 19.2 数据集切分

- development：用于改 prompt 和代码；
- validation：用于选择方案；
- test：最后评估，避免反复偷看；
- regression：收集真实故障并永久保留；
- adversarial：注入、越权、乱码、超长和边界输入。

### 19.3 指标必须有口径

不要只写“准确率 90%”。

至少说明：

- 样本来源和数量；
- 标注规则；
- 指标公式；
- 模型和配置；
- prompt/schema 版本；
- 重试是否计入；
- 置信区间或波动；
- 失败样例。

学习项目没有真实测量，就写“待测”，不要编数字。

## 20. 验收题

1. Tokenizer 为什么影响 RAG 切块和 API 成本？
2. causal mask 在训练并行和自回归约束之间起什么作用？
3. prefill 与 decode 各自主要受什么影响？
4. KV cache 的占用与哪些变量近似成正比？
5. MQA/GQA 为什么可能降低 KV cache 压力？
6. TTFT 与 TPOT 为什么不能混成一个“速度”？
7. RAG、Tool、SFT 分别优先解决什么问题？
8. LoRA 与 DPO 为什么可以同时使用？
9. RLVR 的 verifier 有哪些被利用风险？
10. 模型切换前应保留哪些评估证据？

## 21. 本章结论

LLM 工程优化的第一步不是背模型名，而是定位层次。

输入长度主要影响 prefill 和上下文预算，输出长度主要增加 decode 步骤，KV cache 把长上下文与并发资源联系起来。

知识、行为、偏好和可验证推理属于不同问题；只有先定义失败和证据，RAG、LoRA、DPO 或 RLVR 的选择才有意义。
