# Long Context 与稀疏注意力训练概念

> **文件编码**：UTF-8。  
> **前置**：[11 Transformer 实现](11-Transformer从零实现PyTorch.md)、[14 预训练原理](14-预训练与语言模型原理.md)、[15 LoRA/PEFT](15-微调SFT与LoRA-PEFT.md)。  
> **对照**：[LLMInfra 08 KV Cache](../LLMInfra/08-KVCache与PagedAttention原理.md)、[Infra 11 长上下文推理](../LLMInfra/11-长上下文与Ring-Attention推理优化.md)。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**长上下文 = 让模型看得更远**：训练侧用 RoPE 扩展、YaRN、Ring Attention、LongLoRA 突破预训练长度；推理侧靠 KV Cache 与稀疏注意力省显存——本章偏 **概念与 Python 配置**。

### 0.2 你需要提前知道什么

| 情况 | 建议 |
|------|------|
| 不懂 RoPE | 先读 [14 章](14-预训练与语言模型原理.md) |
| 只训 4K | 本章讲 32K～128K+ 扩展 |
| 做推理 | 结合 [20 章](20-vLLM-TGI与LMDeploy-Python侧.md) `max_model_len` |

### 0.3 本章知识地图（☐→☑）

- [ ] 解释 RoPE 外推问题
- [ ] 说清 YaRN / PI 思路
- [ ] 理解 Ring Attention 切 sequence
- [ ] 知道 LongLoRA S2-Attn
- [ ] 会改 HF `rope_scaling`
- [ ] 完成 §13 闭卷自测 ≥8/10

### 0.4 建议学习时长

长度瓶颈 1h · RoPE 扩展 2h · Ring/LongLoRA 2.5h

### 0.5 学完本章你能做什么

1. 读 model card 里 `rope_scaling` 不再懵。
2. 微调 7B 到 32K 时选对 PI vs YaRN。
3. 面试答「128K 怎么训/推」三层：位置编码、算力、显存。

---

## 1. 为什么需要 Long Context

| 场景 | 典型长度 |
|------|----------|
| RAG 多文档 | 8K～128K |
| 代码仓库 | 32K+ |
| 法律/论文 | 64K～ |

**三大瓶颈**：

```text
算力：Attention O(n²)
显存：KV Cache ∝ n × layers
外推：4K 预训练直接推 32K → perplexity 爆炸
```

---

## 2. 注意力复杂度与稀疏

| 组件 | 复杂度 |
|------|--------|
| Dense Self-Attention | O(n²) |
| KV Cache（推理） | O(n) 每层 |

**稀疏注意力**：局部窗口 + 少量全局 token，降为 O(n×w) 或 O(n log n)。

| 类型 | 模式 |
|------|------|
| Sliding window | 只看前后 w |
| Local + global | 部分 token 看全场 |

Mistral 等架构内置 sliding window；vLLM / FlashAttention 优化 **dense** 长序列。

---

## 3. RoPE 与长度外推

### 3.1 外推失败

预训练长度 \(L_{\text{train}}\)（如 4096）上设计 RoPE；推理 \(L_{\text{test}} > L_{\text{train}}\) 时：

- 未见位置角频率 → 注意力分布偏移
- **lost in the middle**：长文中间信息利用差

### 3.2 解决路线

```text
继续预训练长序列（最贵、最好）
RoPE 缩放：PI / YaRN（常用）
仅推理插值（提升有限）
```

---

## 4. Position Interpolation（PI）

把位置 \(p \in [0, L_{\text{new}})\) 线性压缩到 \([0, L_{\text{train}})\)：

\[
p' = p \times \frac{L_{\text{train}}}{L_{\text{new}}}
\]

通常需 **少量长文继续训练**；lr 偏小保短上下文能力。

```python
config.rope_scaling = {"type": "linear", "factor": 4.0}  # 4096→16384
config.max_position_embeddings = 16384
```

`factor = L_new / L_train`。

---

## 5. YaRN

PI 对所有频率统一缩放；YaRN **分区处理**高低频，并加 **mscale** 修正注意力熵。

```python
config.rope_scaling = {
    "type": "yarn",
    "factor": 8.0,
    "original_max_position_embeddings": 4096,
}
config.max_position_embeddings = 32768
```

| | PI | YaRN |
|---|----|----|
| 实现 | 简单 | 稍复杂 |
| 短上下文 | 一般 | 通常更好 |
| 采用 | 早期 Llama 扩展 | Qwen2.5 等 128K |

---

## 6. Ring Attention（训练长序列）

单卡装不下 n=128K 的 activation。**序列维切多 GPU 成环**，各卡持 Q/K/V 段，ring all-gather 遍历 K/V 块。

```text
GPU0: tokens [0, L)     ─┐
GPU1: tokens [L, 2L)    ─┼─ ring 通信
GPU2: tokens [2L, 3L)   ─┘
```

| 项目 | 说明 |
|------|------|
| ring-flash-attention | FlashAttention + ring |
| Megatron CP | NeMo context parallel |
| DeepSpeed Ulysses | sequence parallel |

与并行维关系（见 [40 章](40-Megatron-Core与3D并行Python侧入门.md)）：

```text
DP：切 batch · TP：切矩阵 · PP：切 layer · CP/Ring：切 sequence
```

---

## 7. FlashAttention 与 LongLoRA

### 7.1 FlashAttention

IO-aware 分块 attention，**dense 下数学等价**，使更长 n 可训：

```python
model = AutoModelForCausalLM.from_pretrained(
    "Qwen/Qwen2.5-7B-Instruct",
    attn_implementation="flash_attention_2",
    torch_dtype="bfloat16",
)
```

与 Ring 组合：Flash 降单卡峰值；Ring 分 sequence 到多卡。

### 7.2 LongLoRA

**LoRA + Shifted Short Attention（S2-Attn）**：head 分组，一组正常局部 window，一组 half-head shift 扩大感受野，低成本长文微调。

```text
1. sliding window base（Llama 等）
2. LoRA target: q_proj, v_proj（同 15 章）
3. 长文档 jsonl，max_seq_len=32K
```

社区实现：`dvlab-research/LongLoRA`。

---

## 8. 数据与推理实践

- **Packing**：短样本拼成长序列，减 padding 浪费
- **继续训**：更小 lr、gradient checkpointing 必选
- **监控短上下文 val**，防「只会长不会短」

```python
# vLLM 推理
LLM(model="...", max_model_len=32768)  # KV 显存 ∝ 长度
```

---

## 9. 方法对照表

| 方法 | 训/推 | 主要解决 |
|------|-------|----------|
| 继续预训练长序列 | 训 | 质量（最贵） |
| PI / YaRN | 训+推 | 位置外推 |
| FlashAttention | 训+推 | 单卡算力/显存 |
| Ring / CP | 训 | 多卡 sequence |
| Sliding Window | 推 | O(n²)→O(n×w) |
| LongLoRA | 训 | 低成本长文微调 |
| PagedAttention | 推 | KV 碎片（20 章） |

---

## 10. 练习建议

1. 读 Qwen2.5 model card 的 `max_position_embeddings`
2. 改 `rope_scaling` factor 对比 perplexity（小模型）
3. 画 Ring Attention 环状通信示意图
4. 调 vLLM `max_model_len` 观察 OOM 边界

---

## 11. FAQ

**Q1：只改 config 不训练能 128K 吗？** 能跑通，质量通常差。  
**Q2：RAG 8K 还要学吗？** FlashAttention 与排错仍有用。  
**Q3：Ring 微调 LoRA 需要吗？** 7B 32K 单卡或够；更长才需 SP/Ring。  
**Q4：lost in the middle？** 训练含中间检索样本；RAG 重排。  
**Q5：PI vs YaRN 选型？** 跟 model card；Qwen2.5 系多用 YaRN。

---

## 12. 学完标准

- [ ] 解释 RoPE 外推失败
- [ ] 对比 PI 与 YaRN
- [ ] 说 Ring Attention 切哪一维
- [ ] 改 HF `rope_scaling`

---

## 13. 闭卷自测

1. 长上下文三大瓶颈？
2. Self-Attention 对 n 复杂度？
3. RoPE 外推两条表现？
4. PI 核心直觉？
5. YaRN 相对 PI 差异？
6. Ring Attention 切哪个并行维？
7. FlashAttention 改变 dense 数学结果吗？
8. LongLoRA S2-Attn 做什么？
9. KV Cache 显存与 n 关系？
10. 继续预训练 vs 仅 rope_scaling，哪个质量更好？

<details>
<summary>参考答案</summary>

1. 算力 O(n²)、KV/激活显存、位置外推质量下降。
2. O(n²)（dense）。
3. perplexity 恶化；lost in the middle。
4. 将长位置线性压缩到训练长度内。
5. YaRN 分区缩放高低频；PI 统一线性缩放。
6. Sequence 维（context parallel）。
7. 不改变（IO 实现不同，数学等价）。
8. shifted 局部 attention 扩感受野且省算力。
9. 近似线性 O(n)。
10. 继续预训练（含长文）通常更好。

</details>

---

## 14. 下一章

[40 Megatron-Core 与 3D 并行 Python 侧入门](40-Megatron-Core与3D并行Python侧入门.md)

---

*KV：[Infra 08](../LLMInfra/08-KVCache与PagedAttention原理.md) · 分布式：[17 DDP/FSDP](17-分布式训练DDP-FSDP与DeepSpeed.md)*
