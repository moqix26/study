# Megatron-Core 与 3D 并行 Python 侧入门

> **文件编码**：UTF-8。  
> **前置**：[17 DDP/FSDP/DeepSpeed](17-分布式训练DDP-FSDP与DeepSpeed.md)、[14 预训练原理](14-预训练与语言模型原理.md)、[39 长上下文](39-Long-Context与稀疏注意力训练概念.md)。  
> **对照**：[LLMInfra 10 并行与 NCCL](../LLMInfra/10-分布式训练并行策略与NCCL入门.md)。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**Megatron 系 = 大模型预训练的 3D/4D 并行栈**：TP 切矩阵、PP 切层、DP 切 batch；Megatron-LM 是脚本，Megatron-Core 是库，NeMo 是 NVIDIA 上层框架——与 DeepSpeed ZeRO 互补。

### 0.2 你需要提前知道什么

| 情况 | 建议 |
|------|------|
| 只会 DDP | 先读 [17 章](17-分布式训练DDP-FSDP与DeepSpeed.md) |
| 只做 7B LoRA | 本章偏预训练/超大模型视野 |
| 读 Infra 10 | 本章 Python 入口与框架名对照 |

### 0.3 本章知识地图（☐→☑）

- [ ] 区分 DP、TP、PP、CP
- [ ] Megatron-LM vs Core vs NeMo
- [ ] 读懂 TP/PP/DP rank 划分
- [ ] NeMo recipe 概念
- [ ] 对比 Megatron+TP 与 DeepSpeed ZeRO-3
- [ ] 完成 §13 闭卷自测 ≥8/10

### 0.4 建议学习时长

并行维 2h · Megatron 栈 2h · NeMo 1.5h · DeepSpeed 对比 1.5h

### 0.5 学完本章你能做什么

1. 读 NVIDIA 文档不混淆 Megatron-LM / Core / NeMo。
2. 面试画 8 卡 TP=2 PP=2 DP=2 分组图。
3. 判断 70B 预训练该调研 Megatron 还是 DeepSpeed。

---

## 1. 为什么需要 3D 并行

7B bf16 约 14GB；70B 约 140GB——超单卡 80GB。DDP 每卡完整副本，无法训 70B+。

```text
TP（切 column/row 线性层）
  × PP（切 transformer layer 段）
  × DP（切 micro-batch）
  = 3D Parallelism
```

可选第四维 **CP**（context parallel）切长序列（[39 章](39-Long-Context与稀疏注意力训练概念.md)）。

| 规模 | 常见策略 |
|------|----------|
| ≤7B | 单卡 / DDP / FSDP |
| 7B～30B | FSDP、ZeRO-3、TP=2 |
| 30B～100B+ | TP + PP + DP |
| 100B+ 长序列 | + CP |

---

## 2. 张量并行（TP）

Attention / MLP 线性层按 column / row 切分；每层 forward/backward 有 All-Reduce 或 All-Gather。

- TP 组内需 **高带宽 NVLink**；尽量少跨机 TP
- vLLM `tensor_parallel_size=2`（[20 章](20-vLLM-TGI与LMDeploy-Python侧.md)）同源；训练 TP 还涉及 optimizer 分片

```bash
# Megatron-LM CLI（示意）
--tensor-model-parallel-size 2
```

---

## 3. 流水线并行（PP）

按 **layer 段** 分到不同 GPU：

```text
GPU0: Layer 0-7  →  GPU1: Layer 8-15  →  GPU2: Layer 16-23  →  GPU3: Layer 24-31
```

- 朴素 pipeline 有 **气泡 idle**；1F1B schedule 减小（Infra 10）
- 约束：`num_layers % pp_size == 0` 常见

```bash
--pipeline-model-parallel-size 4
```

---

## 4. 数据并行（DP）与组合

完整模型副本 = 一组 `(TP, PP, CP)`；多副本间 **DP**，不同 batch，梯度 All-Reduce。

**8 卡示例** `TP=2, PP=2, DP=2`：

```text
DP group 0: GPU0-3（内部分 TP×PP）  batch A
DP group 1: GPU4-7                  batch B
```

\[
B_{\text{global}} = B_{\text{micro}} \times \text{grad\_accum} \times DP
\]

TP/PP **不**乘 batch；它们分模型。

\[
N_{\text{GPU}} = TP \times PP \times CP \times DP
\]

---

## 5. Megatron-LM 与 Megatron-Core

| 名称 | 是什么 |
|------|--------|
| **Megatron-LM** | 开源训练仓库（`pretrain_gpt.py`） |
| **Megatron-Core** | 库（`megatron.core`，parallel layers） |
| **Transformer Engine** | FP8 / fused kernel |

```text
Megatron-LM（脚本）→ Megatron-Core（parallel_state, ColumnParallelLinear）
  → PyTorch + NCCL
```

```python
# 概念性 import
from megatron.core import parallel_state

parallel_state.initialize_model_parallel(
    tensor_model_parallel_size=2,
    pipeline_model_parallel_size=2,
)
```

实际项目多通过 Megatron-LM 或 NeMo 封装。预训练权重常 Megatron 格式，需转 HF 做 PEFT 微调（[15 章](15-微调SFT与LoRA-PEFT.md)）。

---

## 6. Megatron-LM 训练入口（了解）

```bash
torchrun --nproc_per_node=8 pretrain_gpt.py \
  --tensor-model-parallel-size 2 \
  --pipeline-model-parallel-size 2 \
  --num-layers 32 \
  --hidden-size 4096 \
  --seq-length 4096 \
  --micro-batch-size 1 \
  --global-batch-size 256 \
  ...
```

- 数据：**indexed binary**（`.bin` + `.idx`），非原始 jsonl
- checkpoint：每 TP/PP rank 一份；合并需官方工具或 NeMo export

---

## 7. NVIDIA NeMo 框架

**NeMo** = LLM 训练与应用框架，封装 Megatron-Core、recipe、export、NeMo Aligner。

```text
NeMo Framework
├── nemo.collections.llm
├── Recipe YAML（模型+数据+并行+优化器）
└── Export → TensorRT-LLM / vLLM
```

```yaml
# 概念性 recipe
model: {config: llama3_8b}
trainer: {num_nodes: 4, devices: 8}
strategy:
  tensor_model_parallel_size: 2
  pipeline_model_parallel_size: 2
data: {path: /data/megatron_bins}
```

| | Megatron-LM | NeMo |
|---|-------------|------|
| 灵活 hack | 高 | 中（recipe 驱动） |
| 企业支持 | 社区 | NVIDIA NGC |
| 学习曲线 | 陡 | 中等 |

CLI：`nemo llm pretrain` + recipe yaml。

---

## 8. 与 DeepSpeed 对比

| 维度 | Megatron 3D | DeepSpeed ZeRO |
|------|-------------|----------------|
| 切分 | 层（PP）+ 矩阵（TP） | 优化器/梯度/参数分片 |
| 场景 | 100B+ 预训练 | 7B～70B 微调、中等规模 |
| HF 集成 | 需转换 | Trainer + deepspeed json |

**可组合**：Megatron TP+PP + ZeRO-1 optimizer 等混合。

### 选型决策树

```text
7B LoRA？        → HF + PEFT + DDP（15、17 章）
13B 全参 SFT？   → FSDP / ZeRO-2/3
70B+ 预训练？    → Megatron TP+PP+DP 或 NeMo
不愿转 HF 格式？ → DeepSpeed ZeRO-3 + TP（accelerate）
```

### 与 17 章对照

| 17 章 | 本章 |
|-------|------|
| DDP、FSDP、ZeRO json | TP、PP、Megatron 栈 |
| HF Trainer | `pretrain_gpt.py` + indexed data |
| 单机多卡常见 | 多机多卡预训练 |

---

## 9. 练习建议

1. 32 层 PP=4，每 stage 几层？
2. 画 TP=2 PP=2 四卡示意图
3. 查 NeMo NGC 一个 Llama recipe 的 TP/PP/DP
4. 对比 ZeRO-3 与 Megatron TP=2 各解决什么瓶颈

---

## 10. FAQ

**Q1：微调必须 Megatron？** 否；7B LoRA 用 HF 即可。  
**Q2：权重给 vLLM？** NeMo export 或 megatron-to-huggingface。  
**Q3：TP 跨机？** 能但慢；尽量单机 NVLink 内 TP。  
**Q4：PP 气泡？** 1F1B 减小，无法为零。  
**Q5：只写 Java 后端要学吗？** ML 系统岗需要概念；纯 API 集成 17 章够用。

---

## 11. 学完标准

- [ ] 写出 TP/PP/DP 各切什么
- [ ] 解释 Megatron-LM 与 Core 关系
- [ ] 说 NeMo recipe 作用
- [ ] 对比 Megatron TP 与 ZeRO-3

---

## 12. 闭卷自测

1. 3D 并行三个维度各切什么？
2. TP 为何偏好 NVLink 组内？
3. PP 气泡是什么？
4. Megatron-Core 相对 Megatron-LM 角色？
5. NeMo 与 Megatron-LM 关系？
6. global batch 公式？
7. ZeRO-3 分片什么？
8. 70B 预训练选 Megatron 还是 DDP？
9. Context Parallel 解决什么？
10. 7B LoRA 推荐 Megatron 还是 HF+PEFT？

<details>
<summary>参考答案</summary>

1. TP 切矩阵；PP 切 layer 段；DP 切 batch（副本间）。
2. TP 每层多次 all-reduce，跨机带宽瓶颈。
3. pipeline 阶段 GPU 空闲等待。
4. Core 是可复用并行库；LM 是训练脚本仓库。
5. NeMo 上层封装 Core/LM，提供 recipe、export。
6. B_global = B_micro × grad_accum × DP。
7. 参数、梯度、优化器状态分片。
8. Megatron TP+PP+DP；DDP 装不下 70B。
9. 超长 sequence 的 activation/KV，切序列维。
10. HF + PEFT + DDP/FSDP。

</details>

---

## 13. 路线延伸

本章为 LLMPython **并行训练进阶**之一；可回 [24 章项目](24-项目实战微调小型语言模型.md)，或读 [Infra 10](../LLMInfra/10-分布式训练并行策略与NCCL入门.md) 补 NCCL。

---

*分布式：[17 DDP/FSDP](17-分布式训练DDP-FSDP与DeepSpeed.md) · 长序列：[39 Long Context](39-Long-Context与稀疏注意力训练概念.md) · 面试：[25 总表](25-面试专题与知识点总表.md)*
