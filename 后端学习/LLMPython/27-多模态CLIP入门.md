# 多模态 CLIP 入门

> **文件编码**：UTF-8。  
> **前置**：[09 CNN 入门](09-视觉CNN入门-可选.md)、[10 Embedding](10-序列模型与Embedding入门.md)、[11 Transformer](11-Transformer从零实现-PyTorch.md)。  
> **对照**：[AIAgent 22 生态选型](../AIAgent/22-大模型生态选型与前沿推理范式.md)；视觉算子见 [LLMInfra 05 GEMM](../LLMInfra/05-矩阵运算cuBLAS与GEMM优化入门.md)。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**CLIP = 用对比学习把图像和文本映射到同一向量空间**：「狗」的图片向量靠近「a photo of a dog」的文本向量——零样本分类、跨模态检索、多模态 LLM 的地基。

### 0.2 你需要提前知道什么

| 情况 | 建议 |
|------|------|
| 不懂 Embedding | 先读 [10 章](10-序列模型与Embedding入门.md) |
| 没跑过 CNN | 先读 [09 章](09-视觉CNN入门-可选.md) 或了解 ResNet/ViT 概念 |
| 只做大语言模型 | ✅ CLIP 是通向 LLaVA、GPT-4V 的桥梁 |
| 要做 RAG 图文 | 本章 + [21 章 LangChain](21-LangChain与LlamaIndex应用层.md) |

### 0.3 本章知识地图

- [ ] 能画 CLIP 双塔结构与对比损失
- [ ] 理解 zero-shot 分类 prompt ensemble
- [ ] 会用 `openai/clip-vit-base-patch32` 做图文相似度
- [ ] 知道 CLIP 与 SigLIP、ALIGN 的关系
- [ ] 了解 CLIP 在 LLaVA 等多模态 LLM 中的位置
- [ ] 能写简单图文检索 Demo

### 0.4 建议学习时长

| 阶段 | 内容 | 时间 |
|------|------|------|
| 原理 | §0～§3 | 1 小时 |
| 代码 | §4～§5 | 1.5 小时 |
| 应用 | §6～§7 | 1 小时 |
| FAQ + 自测 | §8～§9 | 30 分钟 |

### 0.5 学完本章你能做什么

1. 输入一张图 + 多个文本标签，用 CLIP 做 zero-shot 分类。
2. 构建「以文搜图 / 以图搜文」FAISS 索引。
3. 面试解释多模态 RAG 与纯文本 RAG 差异。

---

## 1. 为什么需要 CLIP

| 传统做法 | CLIP 做法 |
|----------|-----------|
| ImageNet 固定类别分类 | 开放词汇，任意文本标签 |
| 图像标签需人工标注 | 400M 图文对弱监督 |
| 视觉与 NLP 模型分离 | **共享语义空间** |

下游：图文检索、VQA、Grounding、作为 LLM 的视觉编码器（LLaVA）。

---

## 2. 模型架构

### 2.1 双塔结构

```text
Image ──→ Image Encoder (ViT / ResNet) ──→ v_img  ─┐
                                                    ├→ cosine similarity matrix
Text  ──→ Text Encoder (Transformer)   ──→ v_txt  ─┘
```

- 图像编码器：ViT 将 patch 当 token；或 ResNet 全局池化。
- 文本编码器：类似 GPT 的 Transformer，取 [EOS] 位置向量。
- 投影层：映射到同一维度 d（如 512）。

### 2.2 对比学习损失（InfoNCE）

一个 batch 含 N 个图文对：

- 正样本：配对的 (I_i, T_i)
- 负样本：batch 内其余 N-1 对

```text
L = -1/N Σ log( exp(sim(v_i^I, v_i^T)/τ) / Σ_j exp(sim(v_i^I, v_j^T)/τ) )
```

τ 为可学习 temperature；对称地对 Text→Image 也算一遍。

**直觉**：拉近配对、推远不配对的，类似 [06 RAG](../AIAgent/06-RAG检索增强生成基础.md) 里 Embedding 的「意思近则向量近」，但 CLIP 同时学 **两种模态**。

---

## 3. Zero-Shot 分类

### 3.1 Prompt 模板

```python
labels = ["dog", "cat", "car"]
texts = [f"a photo of a {label}" for label in labels]
# 对 image 算与每条 text 的相似度，argmax 为预测类
```

OpenAI 论文使用 **prompt ensemble**：多种模板取平均 embedding 更稳。

### 3.2 与监督分类对比

| 维度 | 监督 ResNet | CLIP zero-shot |
|------|-------------|----------------|
| 训练标签 | 固定 1000 类 | 开放文本 |
| 新类 | 需 retrain | 改 prompt 即可 |
| 精度 | 专用集更高 | 通用略低 |

---

## 4. HuggingFace 实战

### 4.1 安装

```bash
pip install transformers torch pillow
```

### 4.2 图文相似度

```python
import torch
from PIL import Image
from transformers import CLIPProcessor, CLIPModel

model = CLIPModel.from_pretrained("openai/clip-vit-base-patch32")
processor = CLIPProcessor.from_pretrained("openai/clip-vit-base-patch32")

image = Image.open("cat.jpg")
texts = ["a photo of a cat", "a photo of a dog", "a photo of a car"]

inputs = processor(text=texts, images=image, return_tensors="pt", padding=True)
with torch.inference_mode():
    outputs = model(**inputs)
    logits_per_image = outputs.logits_per_image  # [1, 3]
    probs = logits_per_image.softmax(dim=1)
print(dict(zip(texts, probs[0].tolist())))
```

### 4.3 批量图文检索

```python
# 1. 编码图库
image_embs = []
for path in image_paths:
    img = processor(images=Image.open(path), return_tensors="pt")
    emb = model.get_image_features(**img)
    emb = emb / emb.norm(dim=-1, keepdim=True)
    image_embs.append(emb)
image_matrix = torch.cat(image_embs, dim=0)  # [N, d]

# 2. 文本查询
text_inputs = processor(text=["a red backpack"], return_tensors="pt")
text_emb = model.get_text_features(**text_inputs)
text_emb = text_emb / text_emb.norm(dim=-1, keepdim=True)

# 3. 相似度 TopK
sims = (text_emb @ image_matrix.T).squeeze(0)
topk = sims.topk(5)
```

可接 FAISS，流程同 [21 章 RAG](21-LangChain与LlamaIndex应用层.md) 向量检索。

---

## 5. 训练 CLIP 要点（了解）

### 5.1 数据

WebImageText (WIT) 等 4 亿对；需过滤、去重、NSFW 过滤。

### 5.2 训练技巧

| 技巧 | 说明 |
|------|------|
| 大 batch | 负样本多，对比学习效果好 |
| 多卡 global batch | 见 [17 章 DDP](17-分布式训练DDP-FSDP与DeepSpeed.md) |
| mixed precision | fp16/bf16 |
| 数据增强 | 仅图像侧 RandomResizedCrop 等 |

### 5.3 变体

| 模型 | 特点 |
|------|------|
| OpenCLIP | 开源复现，多规模 |
| SigLIP | Sigmoid loss，batch 可更小 |
| Chinese-CLIP | 中文图文 |

---

## 6. CLIP 与多模态 LLM

### 6.1 LLaVA 类架构

```text
Image → CLIP ViT → projector → LLM token 空间 → 与 text token 拼接 → causal LM
```

LLM 部分即 [12～15 章](12-HuggingFace-Transformers入门.md) 微调对象；CLIP 常 **冻结** 只训 projector + LLM。

### 6.2 与纯文本 Agent

[AIAgent 05 Agent](../AIAgent/05-Agent架构与ReAct模式.md) 若需「看 UI 截图操作」，视觉编码器多为 CLIP 或 SigLIP 衍生；文本 RAG 无法直接索引像素。

### 6.3 多模态 RAG

```text
Document: 图片 + 说明文字
  → CLIP 编图 + text embedding 编文
  → 联合索引（或仅 caption 进文本 RAG）
Query: 文本或图像
  → 检索 → 多模态 LLM 生成
```

---

## 7. 工程注意事项

### 7.1 分辨率与 patch

`clip-vit-base-patch32`：224×224，patch 32 → 7×7=49 tokens + class token。

### 7.2 延迟

ViT-L 比 Base 准但慢；边缘部署考虑 [23 章 ONNX](23-模型导出ONNX与TorchScript.md) 导出图像塔。

### 7.3 偏见与安全

CLIP 继承网络数据偏见；生产需过滤与人工审核，对照 [AIAgent 11 安全](../AIAgent/11-生产化与安全.md)。

---

## 8. FAQ

**Q1：CLIP 会生成 caption 吗？**  
不会；只算相似度。生成用 BLIP、LLaVA 等。

**Q2：中文怎么办？**  
`OFA-Sys/chinese-clip-vit-base-patch16` 或多语言 SigLIP。

**Q3：CLIP 向量维数？**  
base 通常 512；large 768+。

**Q4：与 BGE 文本 Embedding 能混吗？**  
不建议；空间未对齐。统一用 CLIP text tower 或分别检索再融合。

**Q5：fine-tune CLIP 值得吗？**  
领域图像（医疗、工业）常 fine-tune 图像塔或加 adapter。

**Q6：temperature τ 作用？**  
缩放 logits，控制分布锐度；推理时通常用训练好值。

**Q7：算力需求？**  
推理：单图 ms 级 GPU；训练：需大 batch 多卡。

**Q8：和 27 章路线图关系？**  
本系列 27 章为多模态扩展；LLM 主线 11～16 仍优先。

---

## 9. 闭卷自测

1. CLIP 双塔各自输出什么？
2. 对比损失为何需要大 batch？
3. zero-shot 分类如何用 prompt？
4. `logits_per_image` 形状含义？
5. CLIP 与 LLaVA 分工？
6. 图文检索为何要 L2 normalize？
7. SigLIP 与 CLIP 损失有何不同？
8. 多模态 RAG 与 [AIAgent 06](../AIAgent/06-RAG检索增强生成基础.md) 文本 RAG 差在哪？

<details>
<summary>参考答案</summary>

1. 同维 L2 归一化 embedding 向量。
2. in-batch 负样本越多，对比信号越强。
3. 把类别名填入模板如 "a photo of a {label}"，算与图像相似度 argmax。
4. [batch_image, num_texts] 未归一化相似度 logits。
5. CLIP 编码图像；projector 对齐 LLM；LLM 负责生成。
6. 归一化后点积等于 cosine similarity。
7. SigLIP 用 sigmoid  pairwise，不依赖 softmax 全局归一化。
8. 需编码图像、处理图文对索引；查询可为图像；纯文本 RAG 仅文本向量。

</details>

---

## 10. 下一章

[28 推理优化 torch.compile 与编译栈](28-推理优化torch-compile与编译栈.md)

并行：[09 CNN](09-视觉CNN入门-可选.md)、[AIAgent 22 多模态趋势](../AIAgent/22-大模型生态选型与前沿推理范式.md)。
