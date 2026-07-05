# 序列模型与 Embedding 入门

> **文件编码**：UTF-8。  
> **前置**：[03 PyTorch 张量](03-PyTorch入门与张量操作.md)、[05 训练循环](05-nn.Module与训练循环.md)、[LLMInfra 02 Transformer](../LLMInfra/02-Transformer与注意力机制原理.md)（可并行）。  
> **定位**：掌握 **Embedding、位置编码、RNN/LSTM** 基础——理解 LLM 输入层与序列建模的历史脉络，通向 11 章 Transformer。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**Embedding** = 把离散 token id 映射为稠密向量；**序列模型** = 按时间步聚合上下文（RNN/LSTM → Self-Attention）。

### 0.2 你需要提前知道什么

| 背景 | 建议 |
|------|------|
| 03 章 tensor shape | 必须 |
| 06 章 padding/mask | 建议 |
| 词袋/one-hot 概念 | 本节会对比 |

### 0.3 本章知识地图（☐→☑）

- [ ] 使用 `nn.Embedding` 与 padding_idx
- [ ] 实现 sinusoidal 位置编码
- [ ] 理解 RNN/LSTM 的 `(batch, seq, hidden)` 与 hidden state
- [ ] 对比 RNN 与 Self-Attention 长依赖
- [ ] 完成 §14 闭卷自测 ≥8/10

### 0.4 建议学习时长

- **4～5 天**

### 0.5 学完你能做什么

构建简单字符级语言模型；读懂 GPT 输入层（token + position embed）；与 [LLMInfra 02](../LLMInfra/02-Transformer与注意力机制原理.md) 公式对照。

---

## 1. 从 one-hot 到 Embedding

词表大小 V=10000，one-hot 维度过高且稀疏。**Embedding** 是可学习的查表矩阵 `E ∈ R^{V×d}`。

```python
import torch
import torch.nn as nn

vocab_size = 100
embed_dim = 32
emb = nn.Embedding(vocab_size, embed_dim, padding_idx=0)

ids = torch.tensor([[2, 5, 0], [7, 3, 9]])   # 0 为 PAD
vectors = emb(ids)
print(vectors.shape)  # torch.Size([2, 3, 32])
print(emb.weight[0])  # padding 行初始为 0，且通常不更新
```

**预期**：输出 `(batch, seq_len, embed_dim)`。

```mermaid
flowchart LR
  ID["token ids"] --> E["Embedding lookup"]
  E --> VEC["(B, S, D)"]
  VEC --> PE["+ Pos Encoding"]
```

---

## 2. Embedding 与 nn.Linear 关系

Embedding 等价于 one-hot 乘矩阵，但 **O(1) 查表** 而非 O(V)：

```python
idx = torch.tensor([3, 7])
one_hot = torch.zeros(2, vocab_size)
one_hot[0, 3] = 1
one_hot[1, 7] = 1
manual = one_hot @ emb.weight
print(torch.allclose(manual, emb(idx), atol=1e-6))
```

**True** — 理解即可，实现永远用 Embedding。

---

## 3. 位置编码（Sinusoidal）

Transformer 无 recurrence，需注入位置信息。原始 Transformer 用固定 sin/cos：

```python
import math

def sinusoidal_pe(seq_len, d_model):
    pe = torch.zeros(seq_len, d_model)
    position = torch.arange(seq_len, dtype=torch.float).unsqueeze(1)
    div_term = torch.exp(torch.arange(0, d_model, 2).float() * (-math.log(10000.0) / d_model))
    pe[:, 0::2] = torch.sin(position * div_term)
    pe[:, 1::2] = torch.cos(position * div_term)
    return pe.unsqueeze(0)   # (1, seq_len, d_model)

pe = sinusoidal_pe(10, 32)
x = emb(torch.randint(1, 100, (2, 10)))
x = x + pe[:, : x.size(1), :]
print(x.shape)
```

现代 LLM 常用 **可学习位置 embed** 或 **RoPE**（11/16 章）；本节掌握 sin/cos 即可读论文原版。

---

## 4. 简单 RNN

```python
rnn = nn.RNN(input_size=32, hidden_size=64, batch_first=True)
out, h_n = rnn(x)
print(out.shape, h_n.shape)
```

**预期**：

```text
torch.Size([2, 3, 64]) torch.Size([1, 2, 64])
```

- `out[:, t]`：时刻 t 的 hidden（含当前输入）
- `h_n`：最后层最终 hidden

```python
# 展开计算直觉（单步）
rnn_cell = nn.RNNCell(32, 64)
h = torch.zeros(2, 64)
for t in range(x.size(1)):
    h = rnn_cell(x[:, t, :], h)
print(h.shape)  # torch.Size([2, 64])
```

---

## 5. LSTM 与门控

```python
lstm = nn.LSTM(input_size=32, hidden_size=64, batch_first=True, num_layers=1)
out, (h_n, c_n) = lstm(x)
print(out.shape, h_n.shape, c_n.shape)
```

LSTM 用 **forget/input/output 门** 缓解 RNN 梯度消失，更长依赖（仍弱于 Transformer 并行与直接 attention）。

| 模型 | 并行度 | 长依赖 | LLM 现状 |
|------|--------|--------|----------|
| RNN/LSTM | 低（逐步） | 中 | 基本淘汰 |
| Transformer | 高 | 强（O(n²)） | 主流 |

---

## 6. 字符级语言模型 micro 示例

```python
class CharLM(nn.Module):
    def __init__(self, vocab_size, embed_dim, hidden_dim):
        super().__init__()
        self.emb = nn.Embedding(vocab_size, embed_dim)
        self.lstm = nn.LSTM(embed_dim, hidden_dim, batch_first=True)
        self.head = nn.Linear(hidden_dim, vocab_size)

    def forward(self, ids):
        e = self.emb(ids)
        out, _ = self.lstm(e)
        return self.head(out)

vocab_size = 50
model = CharLM(vocab_size, 32, 64)
ids = torch.randint(0, vocab_size, (4, 20))
logits = model(ids)
print(logits.shape)  # torch.Size([4, 20, 50])
```

训练：输入 `ids[:, :-1]`，预测 `ids[:, 1:]`（next-token prediction，与 GPT 相同目标）。

---

## 7. 处理 padding 与 pack（了解）

```python
from torch.nn.utils.rnn import pack_padded_sequence, pad_packed_sequence

lengths = torch.tensor([3, 2])
padded = torch.randn(2, 3, 32)
packed = pack_padded_sequence(padded, lengths.cpu(), batch_first=True, enforce_sorted=False)
lstm = nn.LSTM(32, 64, batch_first=True)
packed_out, _ = lstm(packed)
out, lens = pad_packed_sequence(packed_out, batch_first=True)
print(out.shape)
```

跳过 PAD 位置计算，RNN 时代常用；Transformer 用 **attention mask** 代替（11 章）。

---

## 8. Embedding 初始化与缩放

```python
emb = nn.Embedding(1000, 512)
nn.init.normal_(emb.weight, mean=0, std=0.02)
# GPT 类：embed_dim ** 0.5 缩放有时见论文
```

大词表 Embedding 占显存：`V × d × 4` 字节（fp32）；LLM 用 tying：`lm_head.weight` 与 `embed.weight` 共享（11 章）。

---

## 9. 与 Transformer 输入层对照

| 组件 | 本章 | Transformer（11 章） |
|------|------|----------------------|
| Token | nn.Embedding | 同 |
| Position | sin/cos 或可学习 | 可学习 / RoPE |
| 混合 | concat 或相加 | 通常 **相加** |
| 序列混合 | LSTM | Multi-Head Attention |

阅读 [LLMInfra 02](../LLMInfra/02-Transformer与注意力机制原理.md) 时对照本节 embedding 输出 shape。

---

## 10. 练习

1. 实现 `position = token_embed + pe` 的模块 `TokenWithPE(nn.Module)`。
2. 用 3 层 LSTM CharLM 在固定字符串上过拟合（loss→0）。
3. 对比 RNN vs LSTM 同数据 50 step 的 loss（随机数据即可）。
4. 计算 vocab=50257, d=4096 的 Embedding 参数量（MB，fp32）。
5. 画 RNN 与 Self-Attention 信息 flow 的 Mermaid 对比图。

---

## 11. 学完标准

- [ ] 闭卷写出 Embedding 输出 shape
- [ ] 实现 sin 位置编码
- [ ] 解释 padding_idx 作用
- [ ] 说出 LSTM 相对 RNN 改进点
- [ ] 说明 next-token prediction 训练目标

---

## 12. FAQ

**Q1：Embedding 会更新吗？**  
会，除非 freeze；padding_idx 行不更新。

**Q2：token id 从 0 还是 1 开始？**  
自定义；0 常作 PAD。HF tokenizer 有固定 special ids（12 章）。

**Q3：位置编码为何 sin/cos？**  
相对位置线性组合性质；现代 LLM 多用 RoPE 等。

**Q4：batch_first 是什么？**  
`(N, S, D)` vs `(S, N, D)`；推荐 True，与 Transformer 一致。

**Q5：双向 LSTM 用于 GPT 吗？**  
不；GPT 因果单向。BERT 用双向（12 章）。

**Q6：embed_dim 与 hidden_dim 必须相等吗？**  
不必；RNN 可投影；Transformer 常 d_model 统一。

**Q7：为何 LLM 不用 LSTM？**  
难并行、长程仍弱；Attention 可扩展（配合 FlashAttention，LLMInfra 15）。

**Q8：字符级 vs 子词？**  
字符 seq 长；BPE/SentencePiece 折中（12 章 tokenizer）。

**Q9：weight tying？**  
输入 embed 与输出 lm_head 共享，减参、有时提升 perplexity。

**Q10：OOV 词？**  
子词拆分 `<unk>` 减少；Embedding 行数 = 词表大小。

---

## 13. 闭卷自测

1. Embedding(1000, 512) 参数量？
2. padding_idx=0 时 weight[0] 梯度？
3. batch_first=True 时 RNN 输入 shape？
4. LSTM 比 RNN 多哪个状态？
5. sin 位置编码输入依赖什么？
6. next-token 对齐方式（输入/标签）？
7. pack_padded_sequence 目的？
8. 为何 one-hot 不如 Embedding？
9. 因果 LM 能否用 bidirectional LSTM？
10. LLM 主流序列混合机制？

<details>
<summary>参考答案</summary>

1. 1000×512 = 512,000。
2. 通常为 0（不更新 padding 行）。
3. (batch, seq, input_size)。
4. cell state c_t（及 h_t）。
5. 位置 index 与 d_model 维度。
6. 输入 ids[:,:-1]，标签 ids[:,1:]。
7. 忽略 PAD 步，省计算。
8. 稀疏高维、无参数共享、效率低。
9. 训练因果 LM 不用；会泄露未来 token。
10. Self-Attention（Multi-Head）。

</details>

---

## 14. 下一章预告

**11 Transformer PyTorch 从零实现**（见 [00 路线图](00-学习路线图与说明.md)）：Multi-Head Attention、FFN、Decoder-only 堆叠——本章 Embedding + 位置编码将直接接入。

---

*上一章：[09 视觉 CNN](09-视觉CNN入门.md)*  
*路线图：[00 学习路线图与说明](00-学习路线图与说明.md)*  
*数学原理：[LLMInfra 02 Transformer](../LLMInfra/02-Transformer与注意力机制原理.md)*
