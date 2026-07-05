# PyTorch Lightning 工程化训练

> **文件编码**：UTF-8。  
> **前置**：[05 nn.Module 训练循环](05-nn.Module与训练循环.md)、[08 GPU/AMP](08-GPU训练与混合精度AMP.md)、[17 分布式](17-分布式训练DDP-FSDP与DeepSpeed.md)。  
> **对照**：[22 MLOps wandb/mlflow](22-MLOps与实验跟踪wandb-mlflow.md)；大规模训练见 [LLMInfra 10 并行](../LLMInfra/10-分布式训练并行策略与NCCL入门.md)。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**Lightning = 把训练样板代码从业务逻辑里抽走**：你写 `LightningModule`（模型+step），`Trainer` 管 GPU、DDP、AMP、checkpoint、logger——少写 200 行 for 循环。

### 0.2 你需要提前知道什么

| 情况 | 建议 |
|------|------|
| 手写训练循环不熟 | 先读 [05 章](05-nn.Module与训练循环.md) |
| 只会 HF Trainer | ✅ 本章学「原生 PyTorch 工程化」 |
| 要改 11 章 Transformer | 本章可 refactor 11 章代码 |
| 超大模型 | HF Trainer + DeepSpeed 仍常用；Lightning 适合中小自定义模型 |

### 0.3 本章知识地图

- [ ] 能解释 LightningModule 与 nn.Module 关系
- [ ] 能实现 training_step / validation_step / configure_optimizers
- [ ] 会用 Trainer 开启 GPU、DDP、混合精度
- [ ] 会接 WandbLogger / ModelCheckpoint callback
- [ ] 知道与 HF Trainer、DeepSpeed 的分工
- [ ] 能把 [11 章 Mini-Transformer](11-Transformer从零实现-PyTorch.md) 迁到 Lightning

### 0.4 建议学习时长

| 阶段 | 内容 | 时间 |
|------|------|------|
| 概念 | §0～§2 | 40 分钟 |
| 最小示例 | §3 | 1 小时 |
| 进阶 | §4～§6 | 1.5 小时 |
| 与项目结合 | §7 | 1 小时 |

### 0.5 学完本章你能做什么

1. 用 Lightning 重写 [05 章](05-nn.Module与训练循环.md) MNIST/CIFAR 训练，代码减半。
2. 双卡 DDP 一条 `Trainer(accelerator="gpu", devices=2, strategy="ddp")`。
3. 面试说清 Lightning 与 HF Trainer 选型。

---

## 1. 为什么用 Lightning

### 1.1 手写循环的重复劳动

```text
for epoch:
  for batch in train_loader:
    optimizer.zero_grad()
    loss = ...
    loss.backward()
    optimizer.step()
  # val、save、log、scheduler、AMP、DDP sync...
```

每新项目复制粘贴，易漏 `model.train()` / `eval()`、DDP `set_epoch` 等。

### 1.2 Lightning 分工

| 你写 | Trainer 管 |
|------|------------|
| forward、loss | 设备转移 |
| training_step | backward、optimizer_step |
| val_step | 验证循环 |
| configure_optimizers | DDP、AMP、进度条 |
| dataloaders（可选） | checkpoint、early stop |

---

## 2. 核心类

### 2.1 LightningModule

继承 `pl.LightningModule`，封装原 `nn.Module`：

```python
import pytorch_lightning as pl
import torch
import torch.nn.functional as F
from torchmetrics.classification import Accuracy

class LitClassifier(pl.LightningModule):
    def __init__(self, lr=1e-3, num_classes=10):
        super().__init__()
        self.save_hyperparameters()  # 自动进 checkpoint
        self.model = torch.nn.Sequential(
            torch.nn.Flatten(),
            torch.nn.Linear(28 * 28, 128),
            torch.nn.ReLU(),
            torch.nn.Linear(128, num_classes),
        )
        self.acc = Accuracy(task="multiclass", num_classes=num_classes)

    def forward(self, x):
        return self.model(x)

    def training_step(self, batch, batch_idx):
        x, y = batch
        logits = self(x)
        loss = F.cross_entropy(logits, y)
        self.log("train/loss", loss, prog_bar=True)
        return loss

    def validation_step(self, batch, batch_idx):
        x, y = batch
        logits = self(x)
        loss = F.cross_entropy(logits, y)
        self.acc.update(logits, y)
        self.log("val/loss", loss, prog_bar=True)

    def on_validation_epoch_end(self):
        self.log("val/acc", self.acc.compute())
        self.acc.reset()

    def configure_optimizers(self):
        optimizer = torch.optim.AdamW(self.parameters(), lr=self.hparams.lr)
        scheduler = torch.optim.lr_scheduler.CosineAnnealingLR(optimizer, T_max=10)
        return [optimizer], [scheduler]
```

### 2.2 Trainer

```python
from pytorch_lightning import Trainer
from pytorch_lightning.loggers import WandbLogger
from pytorch_lightning.callbacks import ModelCheckpoint, EarlyStopping

wandb_logger = WandbLogger(project="lightning-lab", name="mnist-mlp")
checkpoint_cb = ModelCheckpoint(
    monitor="val/acc", mode="max", save_top_k=1, filename="best-{epoch}-{val/acc:.3f}"
)

trainer = Trainer(
    max_epochs=10,
    accelerator="gpu",
    devices=1,
    precision="16-mixed",  # AMP，对照 08 章
    logger=wandb_logger,
    callbacks=[checkpoint_cb, EarlyStopping(monitor="val/loss", patience=3)],
)

trainer.fit(model, train_dataloaders=train_loader, val_dataloaders=val_loader)
```

### 2.3 推理与加载

```python
best = LitClassifier.load_from_checkpoint("path/to/best.ckpt")
best.eval()
with torch.inference_mode():
    pred = best(x)
```

---

## 3. DataModule（可选但推荐）

```python
from pytorch_lightning import LightningDataModule

class MNISTDataModule(LightningDataModule):
    def __init__(self, batch_size=64):
        super().__init__()
        self.batch_size = batch_size

    def setup(self, stage=None):
        self.mnist_train = ...
        self.mnist_val = ...

    def train_dataloader(self):
        return DataLoader(self.mnist_train, batch_size=self.batch_size, shuffle=True)

    def val_dataloader(self):
        return DataLoader(self.mnist_val, batch_size=self.batch_size)

dm = MNISTDataModule()
trainer.fit(model, datamodule=dm)
```

`setup` 在 DDP 各 rank 调用，注意下载只在 rank 0，见 [17 章](17-分布式训练DDP-FSDP与DeepSpeed.md)。

---

## 4. 分布式与混合精度

### 4.1 DDP 一行

```python
Trainer(accelerator="gpu", devices=2, strategy="ddp")
```

等价于手写 `torchrun --nproc_per_node=2` + DDP wrapper；通信原理见 [LLMInfra 10](../LLMInfra/10-分布式训练并行策略与NCCL入门.md)。

### 4.2 精度选项

| precision | 说明 |
|-----------|------|
| 32 | 默认 fp32 |
| 16-mixed | AMP fp16 + fp32 master |
| bf16-mixed | Ampere+ 推荐，对照 [08 章](08-GPU训练与混合精度AMP.md) |
| 64 | 双精度，少见 |

### 4.3 梯度累积

```python
Trainer(accumulate_grad_batches=4)  # 等效 batch×4
```

---

## 5. Callback 与 Logger

### 5.1 常用 Callback

| Callback | 作用 |
|----------|------|
| ModelCheckpoint | 存 best / last |
| EarlyStopping | val 不降则停 |
| LearningRateMonitor | 打 lr 曲线 |
| RichProgressBar | 美化进度 |

### 5.2 Logger 对接 MLOps

```python
from pytorch_lightning.loggers import MLFlowLogger

mlf = MLFlowLogger(experiment_name="lit-transformer", tracking_uri="http://localhost:5000")
Trainer(logger=mlf)
```

与 [22 章](22-MLOps与实验跟踪wandb-mlflow.md) 一致；HF 项目常用 `report_to="wandb"` 而非 Lightning。

---

## 6. 语言模型 Lightning 示例骨架

```python
class LitGPT(pl.LightningModule):
    def __init__(self, model, lr=3e-4):
        super().__init__()
        self.model = model
        self.lr = lr

    def training_step(self, batch, batch_idx):
        input_ids, labels = batch["input_ids"], batch["labels"]
        out = self.model(input_ids=input_ids, labels=labels)
        self.log("train/loss", out.loss, prog_bar=True)
        return out.loss

    def validation_step(self, batch, batch_idx):
        out = self.model(**batch)
        self.log("val/loss", out.loss, sync_dist=True)  # DDP 聚合

    def configure_optimizers(self):
        return torch.optim.AdamW(self.parameters(), lr=self.lr)
```

**大模型微调**：7B+ 更常用 HF `Trainer` + DeepSpeed ZeRO；Lightning 适合 [11 章](11-Transformer从零实现-PyTorch.md) 小模型或研究代码。

---

## 7. Lightning vs HF Trainer vs 手写

| 维度 | Lightning | HF Trainer | 手写 |
|------|-----------|------------|------|
| 自定义模型 | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ |
| HF 模型微调 | ⭐⭐ | ⭐⭐⭐ | ⭐ |
| DeepSpeed 集成 | 插件 | 原生 | 自己接 |
| 学习曲线 | 中 | 低（微调） | 高 |
| 24 章项目 | 可选 | **推荐** | 不推荐 |

---

## 8. 与 11 / 24 章结合

### 8.1 重构 11 章 Mini-Transformer

1. 把 `TransformerLM` 包进 `LitGPT`。
2. `DataModule` 读字符级 txt。
3. `Trainer(max_epochs=20, precision="bf16-mixed")`。
4. checkpoint 接 [22 章 wandb artifact](22-MLOps与实验跟踪wandb-mlflow.md)。

### 8.2 24 章项目

默认 `train_lora.py` + HF Trainer；若面试强调 PyTorch 功底，可写 `LitLoRA` 包装 PEFT 模型（进阶）。

---

## 9. FAQ

**Q1：Lightning 与 PyTorch 版本？**  
查 [Lightning 兼容矩阵](https://lightning.ai/docs/pytorch/stable)；通常 torch 2.x + lightning 2.x。

**Q2：`self.log` 第一个参数为何用 `/`？**  
`train/loss` 分组，wandb 面板自动分栏。

**Q3：DDP 下 log 要 sync_dist 吗？**  
`validation_step` 建议 `sync_dist=True` 聚合各卡 metric。

**Q4：能否用 Lightning 训 LLaMA 70B？**  
理论可以 + DeepSpeed strategy；工程上 HF+DeepSpeed 资料更多。

**Q5：checkpoint 含 optimizer 吗？**  
默认含；纯推理用 `load_from_checkpoint` 或只导出 `state_dict`。

**Q6：与 torch.compile 一起用？**  
可在 `configure_model` 或 `setup` 中 `torch.compile(self.model)`，见 [28 章](28-推理优化torch-compile与编译栈.md)。

**Q7：Java 岗要会 Lightning 吗？**  
不必须；体现 Python 训练工程能力加分，主栈仍 AIAgent。

---

## 10. 闭卷自测

1. LightningModule 必须实现哪三个方法（最简）？
2. Trainer 的 `precision="16-mixed"` 对应什么？
3. `save_hyperparameters()` 作用？
4. DDP 时 `sync_dist=True` 何时需要？
5. Lightning 与 HF Trainer 各适合什么场景？
6. ModelCheckpoint 的 monitor 常用什么指标？
7. DataModule 的 setup 在 DDP 下调用几次？
8. 11 章 Mini-Transformer 迁 Lightning 的好处？

<details>
<summary>参考答案</summary>

1. training_step、configure_optimizers；（验证需 validation_step）。
2. 自动混合精度 AMP fp16 训练。
3. 把 __init__ 参数存入 hparams，checkpoint 可复现。
4. validation/test 在多卡上分别算 metric 需聚合时。
5. Lightning 自定义研究模型；HF Trainer 标准 LM 微调。
6. val/loss 或 val/acc，mode min/max。
7. 每个 rank 各一次；下载数据需 rank0 先行。
8. 少样板代码、规范 log/checkpoint、易扩展 DDP/AMP。

</details>

---

## 11. 下一章

[27 多模态 CLIP 入门](27-多模态CLIP入门.md)
