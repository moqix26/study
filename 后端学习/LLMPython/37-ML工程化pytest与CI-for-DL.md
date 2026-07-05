# ML 工程化：pytest 与 CI for DL

> **文件编码**：UTF-8。  
> **前置**：[05 nn.Module 训练循环](05-nn.Module与训练循环.md)、[22 MLOps wandb/mlflow](22-MLOps与实验跟踪wandb-mlflow.md)、[26 PyTorch Lightning](26-PyTorch-Lightning工程化训练.md)。  
> **对照**：[LLMInfra 18 K8s GPU](../LLMInfra/18-容器化与Kubernetes-GPU推理部署.md)。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**ML 工程化 = 让训练代码像后端服务一样可测、可 lint、可 CI**：pytest 覆盖数据管道与模型 smoke test，Hydra 管超参，pre-commit 拦坏提交，GitHub Actions 在 GPU runner 上跑 nightly 训练。

### 0.2 你需要提前知道什么

| 情况 | 建议 |
|------|------|
| 只会 `python train.py` | 先读 [05 章](05-nn.Module与训练循环.md) |
| 没用过 pytest | 本章 §2 从零讲 fixture |
| 熟悉 Java CI | GitHub Actions ≈ Jenkins Pipeline |
| 要做 24 章项目 | 本章为 **交付加分项** |

### 0.3 本章知识地图（☐→☑）

- [ ] 能为 Dataset / collator 写单元测试
- [ ] 会用 `@pytest.mark.parametrize` 测多种 shape
- [ ] 能写 Hydra `config.yaml` 并与训练脚本集成
- [ ] 会配置 pre-commit（ruff）
- [ ] 能写 GitHub Actions（CPU lint + GPU smoke）
- [ ] 完成 §12 闭卷自测 ≥8/10

### 0.4 建议学习时长

pytest 2h · Hydra 1.5h · pre-commit 45min · GitHub Actions 2h

### 0.5 学完本章你能做什么

1. LoRA 项目加 `tests/test_dataloader.py`，PR 自动跑测试。
2. Hydra 切换 `small.yaml` / `large.yaml` 无需改代码。
3. 简历写「pre-commit + GitHub Actions GPU nightly」。

---

## 1. 为什么 DL 需要测试与 CI

| 痛点 | 解法 |
|------|------|
| collator shape bug | pytest 断言 tensor shape |
| 超参散落 | Hydra yaml |
| 风格不统一 | pre-commit + ruff |
| 「我机器能跑」 | CI 复现环境 |

```text
        ┌─────────────┐
        │ E2E 1-epoch │  ← nightly
        ├─────────────┤
        │ 集成 smoke  │  ← GPU：forward + 1 backward
        ├─────────────┤
        │ 单元测试    │  ← PR：dataset、loss
        └─────────────┘
```

**原则**：PR 门禁 <5 min（CPU）；重 GPU 训练放 schedule。

---

## 2. pytest 入门（ML 场景）

### 2.1 目录结构

```text
project/
├── src/llm_lab/{data,model}.py
├── tests/{conftest,test_data,test_model}.py
├── configs/
└── pyproject.toml
```

### 2.2 测试 Dataset 与 Collator

```python
# tests/test_data.py
from llm_lab.data import JsonlSFTDataset, sft_collator

def test_collator_output_shape(sample_jsonl_path):
    ds = JsonlSFTDataset(sample_jsonl_path, max_len=128)
    batch = sft_collator([ds[0], ds[1]])
    assert batch["input_ids"].shape == (2, 128)
    assert batch["labels"].shape == batch["input_ids"].shape
```

### 2.3 conftest.py 与 mark

```python
# tests/conftest.py
import pytest

@pytest.fixture
def sample_jsonl_path(tmp_path):
    p = tmp_path / "train.jsonl"
    p.write_text(
        '{"instruction":"hi","output":"hello"}\n',
        encoding="utf-8",
    )
    return p

@pytest.mark.parametrize("max_len", [64, 128])
def test_padding(max_len, sample_jsonl_path):
    ...

@pytest.mark.gpu
def test_forward_cuda():
    if not torch.cuda.is_available():
        pytest.skip("no GPU")
```

```bash
pytest -m "not gpu"              # PR 默认
pytest -m gpu                    # GPU runner
pytest --cov=src/llm_lab
```

---

## 3. 模型 Smoke Test（不下载大权重）

| 策略 | 适用 |
|------|------|
| `AutoConfig` + `from_config` | 测 forward shape |
| 缓存 `Qwen2.5-0.5B` | CI artifact cache |
| `fixture(scope="session")` | 避免重复 load |

```python
def test_lora_forward_shape():
    from transformers import AutoConfig, AutoModelForCausalLM
    from peft import LoraConfig, get_peft_model

    config = AutoConfig.from_pretrained("Qwen/Qwen2.5-0.5B-Instruct")
    model = get_peft_model(
        AutoModelForCausalLM.from_config(config),
        LoraConfig(r=8, target_modules=["q_proj", "v_proj"]),
    )
    x = torch.randint(0, config.vocab_size, (2, 16))
    out = model(input_ids=x, labels=x)
    assert torch.isfinite(out.loss)
```

固定 `seed`；断言 **shape / finite loss**，不断言 loss 精确值（防 flaky）。

---

## 4. Hydra 配置管理

### 4.1 目录与入口

```yaml
# configs/config.yaml
defaults:
  - model: qwen_0.5b
  - data: sft_jsonl
  - train: lora
  - _self_

seed: 42
output_dir: outputs/${now:%Y-%m-%d}/${hydra.job.name}
```

```python
import hydra
from omegaconf import DictConfig, OmegaConf

@hydra.main(version_base=None, config_path="configs", config_name="config")
def main(cfg: DictConfig):
    # cfg.train.lr, cfg.model.name ...
    if cfg.get("wandb", {}).get("enabled"):
        wandb.init(config=OmegaConf.to_container(cfg))
```

```bash
python train.py train.lr=1e-4 train.epochs=1
python train.py --multirun train.lr=1e-4,2e-4,3e-4
```

Hydra config 即 wandb 单一来源（[22 章](22-MLOps与实验跟踪wandb-mlflow.md)）。

### 4.2 在 pytest 里测配置

```python
from hydra import compose, initialize

def test_config_composes():
    with initialize(version_base=None, config_path="../configs"):
        cfg = compose(config_name="config", overrides=["train.epochs=1"])
    assert cfg.train.epochs == 1
```

---

## 5. pre-commit

```bash
pip install pre-commit ruff
pre-commit install
```

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.8.0
    hooks:
      - id: ruff
        args: [--fix]
      - id: ruff-format
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v5.0.0
    hooks:
      - id: check-yaml
      - id: check-added-large-files
        args: ["--maxkb=10240"]
```

- 不要 pre-commit 跑完整训练；`.safetensors` 用 `.gitignore` 或 artifact。

---

## 6. GitHub Actions

### 6.1 CPU 门禁

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  lint-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with: {python-version: "3.11", cache: pip}
      - run: pip install -e ".[dev]"
      - run: ruff check .
      - run: pytest -m "not gpu" --cov=src --cov-fail-under=60
```

### 6.2 GPU smoke（nightly / self-hosted）

| 方案 | 说明 |
|------|------|
| self-hosted runner | 自有 GPU 机器 |
| PR 仅 CPU；merge 后 nightly | 成本友好 |

```yaml
# .github/workflows/gpu-smoke.yml
name: GPU Smoke
on:
  schedule: [{cron: "0 2 * * *"}]
  workflow_dispatch:
jobs:
  smoke:
    runs-on: [self-hosted, gpu, linux]
    steps:
      - uses: actions/checkout@v4
      - run: pip install -e ".[dev]"
      - run: pytest -m gpu --maxfail=1
      - run: python train.py train.max_steps=5
        env: {CUDA_VISIBLE_DEVICES: "0", WANDB_MODE: offline}
```

Secrets：`WANDB_API_KEY`、`HF_TOKEN` 用 `${{ secrets.* }}`，勿写明文。

---

## 7. 与 MLOps 衔接

```text
pre-commit → PR CI（CPU）→ merge → nightly GPU smoke
  → wandb artifact（22 章）→ vLLM 验证（20 章）
```

---

## 8. 练习建议

1. 为 [15 章](15-微调SFT与LoRA-PEFT.md) LoRA 加 dataset / collator / loss 三个 pytest。
2. Hydra 拆 `model/data/train` 三组 yaml。
3. 写 CPU-only Actions，push 看绿勾。

---

## 9. FAQ

**Q1：CI flaky？** 固定 seed；断言 finite loss 非精确值。  
**Q2：HF 下载慢？** Actions cache；`from_config` 免下载。  
**Q3：GPU CI 太贵？** PR 只 CPU；GPU nightly。  
**Q4：coverage 多少？** 数据管道 80%+；训练脚本 smoke 40% 即可。  
**Q5：测 DDP？** self-hosted 双卡 `torchrun --nproc_per_node=2 pytest ...`。

---

## 10. 学完标准

- [ ] 写出带 fixture 的 dataset 测试
- [ ] 配置 Hydra defaults 三层 yaml
- [ ] 解释 PR CI 与 nightly GPU 分工

---

## 11. 闭卷自测

1. ML 测试金字塔三层分别测什么？
2. pytest `conftest.py` 作用？
3. `@pytest.mark.parametrize` 解决什么问题？
4. Hydra `defaults` 与 `_self_` 含义？
5. CLI `train.lr=2e-4` 对应什么机制？
6. pre-commit 与 GitHub Actions 分工？
7. PR CI 为何不跑 10 epoch？
8. GPU smoke 最少验证什么？
9. 如何避免 CI 因 HF 下载失败？
10. `@pytest.mark.gpu` 在 CPU runner 如何处理？

<details>
<summary>参考答案</summary>

1. 单元（dataset/collator）、集成 smoke（forward+1 step）、E2E 长训练（nightly）。
2. 目录级共享 fixture，自动被测试发现。
3. 同一逻辑多组输入，避免重复 test 函数。
4. defaults 组合 yaml 组；`_self_` 表示当前文件覆盖 defaults。
5. Hydra CLI override，运行时合并进 DictConfig。
6. pre-commit 本地 lint；Actions 远端门禁与 schedule。
7. 耗时、费 GPU、阻塞 merge。
8. CUDA 可用、forward/backward 不 OOM、loss 有限。
9. cache huggingface；小模型；`from_config` 免下载。
10. `pytest.skip` 或 CI 用 `-m "not gpu"` 排除。

</details>

---

## 12. 下一章

[38 OpenAI API 兼容层与 litellm](38-OpenAI-API兼容层与litellm.md)

---

*Serving：[20 vLLM](20-vLLM-TGI与LMDeploy-Python侧.md) · 实验：[22 MLOps](22-MLOps与实验跟踪wandb-mlflow.md)*
