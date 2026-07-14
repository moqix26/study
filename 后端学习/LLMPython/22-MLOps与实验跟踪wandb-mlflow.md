# MLOps 与实验跟踪 wandb / mlflow

> **文件编码**：UTF-8。  
> **前置**：[05 nn.Module 训练循环](05-nn.Module与训练循环.md)、[15 LoRA/PEFT](15-微调SFT与LoRA-PEFT.md)、[17 分布式训练](17-分布式训练DDP-FSDP与DeepSpeed.md)。  
> **对照**：[AIAgent 08 可观测性](../AIAgent/08-评估可观测安全与成本.md)；模型部署见 [LLMInfra 18 K8s GPU](../LLMInfra/18-容器化与Kubernetes-GPU推理部署.md)。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**MLOps = 让训练实验可复现、可对比、可交付**：用 wandb / MLflow 记录超参、指标曲线、权重 artifact，避免「上周那个 lr=3e-4 的 checkpoint 找不到了」。

### 0.2 你需要提前知道什么

| 情况 | 建议 |
|------|------|
| 只会 `print(loss)` | 先读 [05 章](05-nn.Module与训练循环.md) 标准训练循环 |
| 没跑过微调 | 先读 [15 章](15-微调SFT与LoRA-PEFT.md) |
| 熟悉 Java CI/CD | 把 MLOps 类比为「训练流水线的 Jenkins + 制品库」 |
| 要做 24 章项目 | 本章为 **必做** 交付项 |

### 0.3 本章知识地图

- [ ] 能说出 MLOps 与 DevOps 的差异
- [ ] 会用 wandb 记录 config / log / artifact
- [ ] 会用 MLflow 做 experiment / run / model registry
- [ ] 能设计 yaml 超参与代码分离
- [ ] 知道 checkpoint 与 artifact 版本策略
- [ ] 能在 24 章项目中接入其一并写 README 复现命令

### 0.4 建议学习时长

| 阶段 | 内容 | 时间 |
|------|------|------|
| 概念 | §0～§2 | 40 分钟 |
| wandb 实战 | §3～§4 | 1.5 小时 |
| MLflow 实战 | §5～§6 | 1.5 小时 |
| 选型与项目 | §7～§8 | 1 小时 |

### 0.5 学完本章你能做什么

1. 一次 LoRA 微调全程指标在浏览器可对比。
2. 从 artifact 下载 `best.ckpt` 并在 [20 章 vLLM](20-vLLM-TGI与LMDeploy-Python侧.md) 加载。
3. 简历写「wandb 跟踪 12 组超参，验证集 perplexity 从 18→12」。

### 0.6 与 AIAgent / LLMInfra 关系

```text
训练实验（本章 22）→ 产出 checkpoint + 评估报告
  ↓
Python Serving（20 章）/ Java API（AIAgent 02）→ 在线指标
  ↓
GPU 部署（LLMInfra 18）→ 延迟、吞吐监控（Prometheus 等）
```

---

## 1. MLOps 核心概念

### 1.1 为什么需要实验跟踪

| 痛点 | MLOps 解法 |
|------|------------|
| 超参散落笔记本 | `config` 结构化存储 |
| 无法对比多次 run | Dashboard 曲线叠加 |
| checkpoint 混乱 | artifact 版本 + tag |
| 复现困难 | git commit + config + seed 绑定 |
| 团队协作 | 共享 project、权限 |

### 1.2 实验生命周期

```text
Hypothesis → Config → Train → Log metrics → Eval → Register model → Deploy
                ↑__________________________________|
                        反馈调参
```

### 1.3 应记录什么

| 类别 | 示例 |
|------|------|
| 超参 | lr, batch_size, lora_r, warmup_steps |
| 环境 | torch 版本, GPU 型号, git sha |
| 指标 | train/val loss, perplexity, BLEU |
| 产物 | adapter.safetensors, tokenizer, eval.json |
| 数据 | 数据集版本 hash、样本数 |

---

## 2. wandb 入门

### 2.1 安装与登录

```bash
pip install wandb
wandb login   # 或 WANDB_API_KEY 环境变量
export WANDB_PROJECT=llm-finetune-lab
```

### 2.2 最小集成

```python
import wandb

wandb.init(
    project="llm-finetune-lab",
    config={
        "model_name": "Qwen/Qwen2.5-0.5B-Instruct",
        "lr": 2e-4,
        "epochs": 3,
        "lora_r": 16,
        "seed": 42,
    },
    tags=["lora", "domain-sft"],
)

for step, batch in enumerate(dataloader):
    loss = train_step(batch)
    if step % 10 == 0:
        wandb.log({"train/loss": loss, "lr": scheduler.get_last_lr()[0]}, step=step)

val_ppl = evaluate(model, val_loader)
wandb.log({"val/perplexity": val_ppl}, step=step)
wandb.finish()
```

### 2.3 HuggingFace Trainer 集成

```python
from transformers import TrainingArguments, Trainer

training_args = TrainingArguments(
    output_dir="./outputs",
    report_to="wandb",
    logging_steps=10,
    evaluation_strategy="steps",
    eval_steps=200,
    save_strategy="steps",
    load_best_model_at_end=True,
    metric_for_best_model="eval_loss",
)
trainer = Trainer(model=model, args=training_args, ...)
trainer.train()
```

### 2.4 Artifact 管理

```python
artifact = wandb.Artifact("lora-adapter-v1", type="model")
artifact.add_dir("./outputs/checkpoint-best")
wandb.log_artifact(artifact)
# 下载：run.use_artifact("lora-adapter-v1:latest").download()
```

### 2.5 Sweep 超参搜索

```yaml
# sweep.yaml
program: train_lora.py
method: bayes
metric:
  name: val/perplexity
  goal: minimize
parameters:
  lr:
    min: 1e-5
    max: 5e-4
    distribution: log_uniform
  lora_r:
    values: [8, 16, 32]
```

```bash
wandb sweep sweep.yaml
wandb agent <sweep_id>
```

---

## 3. MLflow 入门

安装 `pip install mlflow`；`mlflow ui --port 5000` 启动 UI。核心：`mlflow.set_experiment` → `start_run` → `log_params` / `log_metrics` / `log_artifact` → `transformers.log_model(..., registered_model_name=...)`。Registry 流程：None → Staging → Production → Archived。

与 PyTorch Lightning 集成见 [26 章 MLFlowLogger](26-PyTorch-Lightning工程化训练.md)。

---

## 4. wandb vs MLflow 选型

| 维度 | wandb | MLflow |
|------|-------|--------|
| UI / 曲线 | 强，实时友好 | 可用，偏企业 |
| HF 集成 | `report_to="wandb"` 一行 | transformers 插件 |
| 模型注册 | Artifact | **Registry 更成熟** |
| 超参搜索 | Sweep 内置 | 需配合 Optuna 等 |

**建议**：个人/小团队研究用 wandb；企业已有 MLflow 平台则统一 MLflow；24 章项目 **至少选一个**。

---

## 5. 超参与配置工程

### 5.1 yaml 驱动训练

```yaml
# configs/lora_qwen.yaml
model_name: Qwen/Qwen2.5-0.5B-Instruct
data:
  train_path: data/train.jsonl
  max_seq_len: 2048
train:
  lr: 2.0e-4
  batch_size: 4
  grad_accum: 8
  epochs: 3
lora:
  r: 16
  alpha: 32
  dropout: 0.05
logging:
  backend: wandb
  project: llm-finetune-lab
seed: 42
```

```python
import yaml
from dataclasses import dataclass

with open("configs/lora_qwen.yaml") as f:
    cfg = yaml.safe_load(f)
wandb.config.update(cfg)  # 完整 config 入库
```

### 5.2 可复现清单

```text
☑ git commit hash（wandb.log_code 或 mlflow log git）
☑ random seed（torch, numpy, python）
☑ 数据文件 md5
☑ requirements.txt / conda-lock
☑ CUDA / torch 版本
```

### 5.3 与 [22→24 章项目](24-项目实战微调小型语言模型.md) 绑定

24 章 Phase 3 要求：每次实验 run 有唯一 id；`best` checkpoint 打 tag；README 写 `wandb artifact get ...` 或 `mlflow models serve`。

---

## 6. 评估与 Artifact 规范

### 6.1 应上传的制品

| 文件 | 说明 |
|------|------|
| `adapter_config.json` | LoRA 结构 |
| `adapter_model.safetensors` | 权重 |
| `eval_metrics.json` | perplexity、人工评分 |
| `samples.jsonl` | 生成样例对比 |
| `confusion_notes.md` | 失败 case 分析 |

### 6.2 指标命名约定

```text
train/loss, train/lr
val/loss, val/perplexity
test/bleu, test/rouge_l
infra/gpu_mem_gb, infra/tokens_per_sec
```

与 [19 章评估](19-模型评估与Benchmark.md) 指标一致，便于跨 run 对比。

---

## 7. 生产衔接

### 7.1 从实验到部署

```text
MLflow Production model / wandb artifact
  → 导出 merged 权重（15 章 merge_and_unload）
  → [23 章 ONNX](23-模型导出ONNX与TorchScript.md) 或 [20 章 vLLM](20-vLLM-TGI与LMDeploy-Python侧.md)
  → [LLMInfra 18](../LLMInfra/18-容器化与Kubernetes-GPU推理部署.md) K8s 部署
```

### 7.2 在线 vs 离线监控

| 层级 | 工具 | 关注点 |
|------|------|--------|
| 训练 | wandb/mlflow | loss、超参 |
| 推理 | Prometheus + Grafana | QPS、P99 延迟 |
| LLM 质量 | [AIAgent 08](../AIAgent/08-评估可观测安全与成本.md) | 幻觉率、用户反馈 |

---

## 8. FAQ

**Q1：wandb 必须联网吗？**  
可 `wandb offline` 后 `wandb sync`；或部署 wandb local server。

**Q2：MLflow 数据库用什么？**  
默认本地 SQLite；团队用 Postgres + 对象存储（S3/MinIO）存 artifact。

**Q3：只训练一次也要 MLOps 吗？**  
24 章项目、面试作品集 **需要**；单次 demo 至少保存 config.json + git sha。

**Q4：如何防止 log 爆磁盘？**  
限制 `logging_steps`；大文件用 artifact 而非每 step log；定期清理 stale run。

**Q5：多机 DDP 会重复 log 吗？**  
只在 rank 0 调用 `wandb.log` / `mlflow.log_metric`。

**Q6：敏感数据能上传 wandb 吗？**  
不要 log 原始训练文本；只 log 聚合指标；私有化部署或离线模式。

---

## 9. 闭卷自测

1. MLOps 解决训练中的哪三类问题？
2. wandb 中 `config` 与 `log` 区别？
3. MLflow Model Registry 的 Staging / Production 含义？
4. 为什么 DDP 只在 rank 0 打 log？
5. 24 章项目至少应记录哪些 artifact？
6. Sweep 与手动 grid search 优劣？
7. `report_to="wandb"` 在 Trainer 里做什么？
8. 训练监控与 [LLMInfra 17 Nsight](../LLMInfra/17-GPU性能剖析Nsight与perf.md) Profiling 分工？

<details>
<summary>参考答案</summary>

1. 超参/环境不可追溯、实验不可对比、模型制品不可版本化交付。
2. config 是 run 级静态超参；log 是训练过程动态指标（按 step/epoch）。
3. Staging 待验证版本；Production 线上Serving 默认引用版本。
4. 避免多进程重复写入同一 run 导致指标翻倍或冲突。
5. adapter 权重、eval 指标、生成样例、config/yaml、复现命令。
6. Sweep 可贝叶斯自适应采样，省 trial；grid 穷举适合低维离散超参。
7. 将 Trainer 的 loss、eval 指标自动同步到 wandb project。
8. MLOps 跟踪算法指标与实验；Nsight 剖析 GPU kernel 级性能瓶颈。

</details>

---

## 10. 下一章

[23 模型导出 ONNX 与 TorchScript](23-模型导出ONNX与TorchScript.md)
