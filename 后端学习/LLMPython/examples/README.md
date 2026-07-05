# LLMPython 实验工程说明

> 路径：`f:\study\后端学习\LLMPython\examples\`  
> 与章节配套的 **可运行脚本 / Notebook**；详细讲解见上级目录各章 Markdown。

---

## 1. 推荐本地结构

```text
examples/
├── .venv/
├── requirements.txt          ← 依赖清单
├── notebooks/                ← 01～11 实验（待建）
├── src/
│   ├── models/
│   │   └── mini_gpt.py       ← 11 章 字符级 Mini-GPT
│   └── train/
│       └── sft_lora.py       ← 15 / 29 章 LoRA 微调
├── configs/                  ← yaml 超参（待建）
├── scripts/
│   └── eval_ppl.py           ← 19 章 Perplexity 评估
└── README.md
```

---

## 2. 依赖安装

`requirements.txt` 已包含核心库：

| 包 | 用途 |
|----|------|
| torch | 张量与 nn.Module |
| transformers | AutoModel、Trainer |
| peft | LoRA 注入 |
| datasets | 数据集加载 |
| accelerate | 多卡 / device_map |
| trl | SFTTrainer、DPOTrainer（29 章） |
| wandb | 实验跟踪（22 章） |
| bitsandbytes | QLoRA 4bit（15 章） |

```bash
cd f:\study\后端学习\LLMPython\examples
python -m venv .venv
# Windows: .venv\Scripts\activate
pip install -r requirements.txt
python -c "import torch; print(torch.__version__, torch.cuda.is_available())"
```

---

## 3. 核心文件说明

### 3.1 `src/models/mini_gpt.py`

字符级 Mini-GPT（~150 行），对应 [11 Transformer 从零实现](../11-Transformer从零实现PyTorch.md) 与 [14 预训练原理](../14-预训练与语言模型原理.md)。

| 符号 | 说明 |
|------|------|
| `GPTBlock` | Pre-LN Transformer Block（Causal MHA + FFN） |
| `MiniGPT` | 完整 GPT；weight tying；`generate()` 自回归采样 |
| `train_step` | 单步前向 + 反向 + 梯度裁剪 |
| `CharVocab` | 字符词表编解码 |

```bash
python src/models/mini_gpt.py
# 预期：loss 下降，打印生成样例
```

### 3.2 `src/train/sft_lora.py`

peft LoRA 微调脚本（~120 行），对应 [15 LoRA/PEFT](../15-微调SFT与LoRA-PEFT.md) 与 [29 TRL 实战](../29-HuggingFace-TRL与SFTTrainer实战.md)（待发布）。

```bash
# 准备 data/alpaca_sample.jsonl（instruction/output 或 messages 格式）
python src/train/sft_lora.py \
  --model_id distilgpt2 \
  --data_path data/alpaca_sample.jsonl \
  --output_dir ./lora-output \
  --epochs 1 --use_wandb
```

| 参数 | 说明 |
|------|------|
| `--model_id` | HF 模型 ID 或本地路径 |
| `--data_path` | `.json` 数组或 `.jsonl` |
| `--output_dir` | LoRA adapter 输出目录 |
| `--use_wandb` | 可选 wandb 日志 |

### 3.3 `scripts/eval_ppl.py`

Perplexity 评估（~80 行），对应 [19 模型评估](../19-模型评估与Benchmark.md)。

```bash
# 默认 WikiText-2 test
python scripts/eval_ppl.py --model_id distilgpt2

# 本地文本
python scripts/eval_ppl.py --model_id ./lora-output --text_path eval.txt
```

---

## 4. 章节 ↔ 实验对照（00～28）

| 章节 | 实验 | 验收 |
|------|------|------|
| 01 | env-check | CUDA True |
| 03 | tensor matmul GPU | 无报错 |
| 05 | train_mnist.py | loss 下降 |
| 08 | amp_train.py | bf16 开启 |
| 11 | `mini_gpt.py` | 生成可读字符 |
| 12 | hf_pipeline.py | summarization |
| 15 | `sft_lora.py` | loss 收敛 |
| 17 | torchrun 2 GPU | 双卡 log |
| 19 | `eval_ppl.py` | PPL 数值合理 |
| 24 | 完整 SFT 项目 | Gradio + wandb |

---

## 5. 章节 ↔ 实验对照（29～40 进阶）

| 章节 | 主题 | 本目录实验 / 扩展方向 | 验收 |
|------|------|----------------------|------|
| 29 | [TRL 与 SFTTrainer](../29-HuggingFace-TRL与SFTTrainer实战.md) | `sft_lora.py` → 换 TRL `SFTTrainer` | chat template mask 正确 |
| 30 | [Unsloth / Axolotl / LLaMA-Factory](../30-Unsloth-Axolotl与LLaMA-Factory工具链.md) | `configs/` yaml 驱动微调 | yaml 一键复现 |
| 31 | [Ray Train 弹性分布式](../31-Ray-Train与弹性分布式训练.md) | `train/ddp_train.py`（待建） | 多机容错 resume |
| 32 | [多模态 LLaVA](../32-多模态LLaVA与视觉语言模型.md) | VLM demo notebook（待建） | 图文 QA |
| 33 | [合成数据 Self-Instruct](../33-合成数据Self-Instruct与知识蒸馏.md) | `data/self_instruct.py`（待建） | 生成 ≥1k 条 |
| 34 | [mergekit 与 MoE](../34-模型合并mergekit与MoE入门.md) | merge 脚本（待建） | SLERP 合并可推理 |
| 35 | [RAG FAISS / Milvus](../35-RAG向量检索FAISS与Milvus深度实战.md) | `scripts/rag_demo.py`（待建） | Top-k 检索命中 |
| 36 | [Guardrails 与安全](../36-LLM安全Guardrails与Red-Team.md) | red-team prompt 集（待建） | 拦截注入样例 |
| 37 | [pytest 与 CI for DL](../37-ML工程化pytest与CI-for-DL.md) | `.github/workflows/`（待建） | CI 跑通 unit test |
| 38 | [OpenAI API / litellm](../38-OpenAI-API兼容层与litellm.md) | `scripts/openai_client.py`（待建） | 兼容 vLLM server |
| 39 | [Long Context 训练概念](../39-Long-Context与稀疏注意力训练概念.md) | RoPE 扩展 notebook（待建） | 长序列 loss 稳定 |
| 40 | [Megatron 3D 并行入门](../40-Megatron-Core与3D并行Python侧入门.md) | TP/PP 概念对照笔记 | 说出 TP/PP/DP 分工 |

> 29～40 章 Markdown 由同级目录文档提供；本目录优先落地 **29（SFT）**、**19（PPL）**、**11（Mini-GPT）** 三个可运行脚本，其余章节实验标注「待建」供后续扩展。

---

## 6. 与 C++ MiniInfer 联调

| Python（本系列） | C++（LLMInfra 19） |
|------------------|---------------------|
| 15 章 LoRA 导出权重 | 12 章 mmap 加载 |
| 20 章 vLLM 部署 | 19 章 gRPC 客户端压测 |
| 13 章 pybind 调 C++ 算子 | LLMInfra 13 + C++ 20 |

---

## 7. 云 GPU 建议

| 章节 | 最低显存 | 推荐 |
|------|----------|------|
| 03～08 | 4G | 本地 3060 |
| 11 `mini_gpt.py` | 6G | 3060 12G / CPU 可跑 demo |
| 15 `sft_lora.py` 7B | 16G | A10 24G |
| 17 DDP | 2×16G | 2×A10 |
| 29～40 进阶 | 视任务 | A100 40G+ |

---

详见 [LLMPython 00 路线图](../00-学习路线图与说明.md) · [24 项目实战](../24-项目实战微调小型语言模型.md)
