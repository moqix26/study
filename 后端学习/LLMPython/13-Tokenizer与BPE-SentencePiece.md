# Tokenizer 与 BPE / SentencePiece

> **文件编码**：UTF-8。  
> **前置**：[12 HuggingFace Transformers 入门](12-HuggingFace-Transformers入门.md)。  
> **定位**：理解 **分词算法、特殊 token、chat_template**，避免微调与推理时「能跑但输出格式错」。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**Tokenizer** = 文本 ↔ token id 的双向映射；现代 LLM 多用 **BPE 或 SentencePiece** 子词表，对话模型还需 **chat_template** 拼 prompt。

### 0.2 你需要提前知道什么

- 12 章 `AutoTokenizer.from_pretrained`
- 了解 Unicode 与字符串编码（不必精通）
- 知道 vocab_size 影响 embedding 与 lm_head 大小

### 0.3 本章知识地图（☐→☑）

- [ ] 解释 BPE merge 规则与 unk 处理
- [ ] 对比 GPT-2 BPE 与 Llama SentencePiece 差异
- [ ] 使用 `encode` / `decode` / `apply_chat_template`
- [ ] 处理 padding、truncation、attention_mask
- [ ] 训练小型 BPE 词表（可选）
- [ ] 完成 §14 闭卷自测 ≥8/10

### 0.4 建议学习时长

- **3～5 天**

---

## 1. 这份文档学什么

- 词表粒度：字符 / 词 / 子词
- BPE（Byte-Pair Encoding）训练与 merge
- SentencePiece：Unigram/BPE、空格用 `▁` 表示
- HF `PreTrainedTokenizer` API
- 特殊 token：`<|endoftext|>`、`<|im_start|>`、pad/bos/eos
- **chat_template**（Jinja2）与多轮对话格式
- `tokenizer.json` 与 Rust `tokenizers` 库
- 推理引擎侧 tokenizer 开销（Infra 07 章架构）

---

## 2. 为何需要子词

| 方案 | 优点 | 缺点 |
|------|------|------|
| 词级 | 语义整 | OOV、超大词表 |
| 字符级 | 词表小 | 序列过长 |
| **子词 BPE/SP** | OOV 少、长度适中 | 实现与空格规则需统一 |

LLM 词表常见 **32k～128k**；embedding 参数量 ≈ `vocab × hidden`（Infra 加载时需一并 mmap）。

---

## 3. BPE 原理（GPT-2 / RoBERTa 系）

**训练**（简化）：

1. 语料按字符 + 词尾标记初始化
2. 统计相邻 symbol 对频率，合并最高频对
3. 重复至词表大小达标

**编码**：

- 按 merge 优先级将词切为 subword
- 字节级 BPE（GPT-2）：任意 Unicode 可表示，极少 unk

```python
from transformers import AutoTokenizer

tok = AutoTokenizer.from_pretrained("gpt2")
text = "Hello world"
ids = tok.encode(text)
print(ids)
print(tok.decode(ids))
print(tok.convert_ids_to_tokens(ids))
# 可能 ['Hello', 'Ġworld'] — Ġ 表示词前空格
```

---

## 4. SentencePiece（Llama / Qwen / T5 系）

- 独立工具训练 `.model` 文件
- 将空格建模为 `▁`（U+2581）
- 常 **不区分** 是否加前缀空格——须与预训练一致

```python
tok = AutoTokenizer.from_pretrained("Qwen/Qwen2.5-0.5B-Instruct")
print(tok.tokenize("你好 world"))
# 中文常按字或子词切分
```

HF 封装 `LlamaTokenizer` / `Qwen2Tokenizer`，底层仍读 `tokenizer.model` 或合并进 `tokenizer.json`。

---

## 5. HuggingFace Tokenizer 核心 API

```python
from transformers import AutoTokenizer

tokenizer = AutoTokenizer.from_pretrained("meta-llama/Llama-3.2-1B-Instruct")

batch = tokenizer(
    ["first sentence", "second one"],
    padding=True,
    truncation=True,
    max_length=512,
    return_tensors="pt",
)

# batch: input_ids, attention_mask
# attention_mask: 1=有效 token，0=pad
```

| 方法 | 作用 |
|------|------|
| `encode` / `decode` | 单条 id 列表 ↔ 文本 |
| `tokenize` | 只看 subword 字符串 |
| `__call__` | batch + tensor + padding |
| `apply_chat_template` | 对话 → 单条 prompt 字符串 |
| `add_special_tokens` | 扩展词表（LoRA 新 special） |

---

## 6. 特殊 Token

```python
print(tokenizer.special_tokens_map)
# bos_token, eos_token, pad_token, unk_token ...
print(tokenizer.eos_token_id)
```

| Token | 典型用途 |
|-------|----------|
| `bos` | 序列开始（部分模型无） |
| `eos` | 结束；也可作 pad（GPT-2） |
| `pad` | batch 对齐 |
| `unk` | 未登录词（BPE 字节级很少用） |

**微调新增 token**（如工具名）：

```python
tokenizer.add_special_tokens({"additional_special_tokens": ["<tool>"]})
model.resize_token_embeddings(len(tokenizer))
```

新 embedding 行随机初始化，需足够数据学习。

---

## 7. chat_template 详解

Instruct 模型依赖 **固定对话格式**，否则 SFT 分布偏移。

```python
messages = [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "什么是 KV Cache？"},
]

prompt = tokenizer.apply_chat_template(
    messages,
    tokenize=False,
    add_generation_prompt=True,
)
print(prompt)

inputs = tokenizer(prompt, return_tensors="pt")
```

- `tokenize=False`：先看字符串是否符合预期
- `add_generation_prompt=True`：末尾加 assistant 开头，供 `generate` 续写
- 模板存在 `tokenizer.chat_template`（Jinja2），各模型不同

**常见格式**：

- ChatML：`<|im_start|>user\n...`
- Llama 3：`<|start_header_id|>user<|end_header_id|>`
- 切勿手写字符串替代——与官方 template 差一个换行都会降质

---

## 8. 训练新 Tokenizer（了解）

```python
from tokenizers import Tokenizer, models, trainers, pre_tokenizers

tokenizer = Tokenizer(models.BPE())
tokenizer.pre_tokenizer = pre_tokenizers.ByteLevel(add_prefix_space=False)
trainer = trainers.BpeTrainer(vocab_size=8000, special_tokens=["<|endoftext|>"])
# trainer.train(files=["corpus.txt"], tokenizer=tokenizer)
# tokenizer.save("my-bpe.json")
```

领域预训练或极小语种可考虑；多数场景 **直接用基座 tokenizer**。

---

## 9. 与训练 / 推理的衔接

**SFT 数据**（15 章）：每条样本应先 `apply_chat_template` 再 tokenize，labels 对 **user 部分 mask 为 -100**（不算 loss）。

```python
labels = input_ids.clone()
# 伪代码：labels[:, :user_len] = -100
```

**推理**（20 章）：OpenAI API 的 messages 由服务端 apply template；自建服务须与训练一致。

**Infra**：C++ 引擎常用相同 `tokenizer.json` 或 SentencePiece C++ API——见 [LLMInfra 13 pybind11](../LLMInfra/13-pybind11与Python-C++混合编程.md) 混合调用场景。

---

## 10. 常见坑

| 现象 | 原因 |
|------|------|
| 中文乱码 | 混用 tokenizer / 错误 decode `skip_special_tokens` |
| 无限生成 | 未设 `eos_token_id` |
| loss 不降 | template 与预训练不一致 |
| 词表变大 OOM | `resize_token_embeddings` 后 lm_head 变大 |
| 空格多/少 token | GPT `Ġ` vs SP `▁` 规则不同 |

---

## 11. 练习建议

1. 对同一英文句比较 `gpt2` 与 `llama3` 的 token 数
2. 打印 Qwen Instruct 的 `apply_chat_template` 原始字符串
3. 实现：只对 assistant 段计算 loss 的 label mask
4. 用 `tokenizer.get_vocab()` 查 rare token id
5. 阅读 `tokenizer_config.json` 中 `chat_template` 字段
6. 统计 10 万条语料的平均 token/字符比（18 章数据）

---

## 12. 学完标准

- [ ] 解释 BPE 一次 merge 在做什么
- [ ] 说出 `attention_mask` 含义
- [ ] 正确调用 `apply_chat_template` + `generate`
- [ ] 知道何时 `resize_token_embeddings`
- [ ] 区分 `encode` 与 `__call__` 返回值

---

## 13. FAQ

**Q1：BPE 和 WordPiece 区别？**  
WordPiece 用似然选 merge；BPE 用频率。BERT 常用 WordPiece；GPT 用 BPE。

**Q2：字节级 BPE 还有 unk 吗？**  
理论上任意 UTF-8 可分解为字节 token，极少 unk。

**Q3：为什么要 `add_prefix_space`？**  
GPT-2 词首空格编码进 subword（`Ġ`）；与预训练必须一致。

**Q4：chat_template 能改吗？**  
可以 `tokenizer.chat_template = "..."`，但微调数据与推理须同步改，否则错位。

**Q5：`truncation_side` 左还是右？**  
对话常保留 **最近** 轮次，设 `truncation_side="left"` 删旧消息。

**Q6：max_length 超过模型 context？**  
超过 `max_position_embeddings` 会报错或截断——与 RoPE 外推不同（Infra 02）。

**Q7：tokenizer 会成为推理瓶颈吗？**  
极长 batch prefill 时 CPU tokenize 可并行；主要瓶颈仍在 GPU（Infra 07）。

**Q8：`slow` vs `fast` tokenizer？**  
`use_fast=True`（默认）走 Rust；fast 支持 offset mapping（NER）。

**Q9：多模态 image token 呢？**  
另有 `processor` 拼 image placeholder——27 章多模态扩展。

**Q10：如何验证 template 正确？**  
与官方推理 demo 同 messages 对比 token ids 前 20 个是否一致。

---

## 14. 闭卷自测

1. BPE 训练循环合并的是什么？
2. GPT-2 中 `Ġ` 表示什么？
3. SentencePiece 如何表示空格？
4. `attention_mask` 中 0 表示什么？
5. `apply_chat_template` 中 `add_generation_prompt` 作用？
6. SFT 为何要对 user 段 label 置 -100？
7. 增加 special token 后必须对 model 做什么？
8. `return_tensors="pt"` 返回什么类型？
9. decode 时 `skip_special_tokens=True` 会怎样？
10. 词表大小直接影响哪两个矩阵参数量？

<details>
<summary>参考答案</summary>

1. 语料中出现频率最高的相邻 symbol 对。
2. 词首前的空格（下一 subword 属于新词）。
3. 用 `▁`（spiace marker）前缀标记。
4. padding 位置，不参与 attention。
5. 在 prompt 末尾添加 assistant 起始标记，供模型续写。
6. 只对 assistant 回复算 LM loss，不训练模型复述 user。
7. `model.resize_token_embeddings(len(tokenizer))`。
8. PyTorch 张量字典（input_ids、attention_mask 等）。
9. 输出文本中省略 bos/eos 等特殊符号。
10. embedding 层与 lm_head（若 weight tying 则共享）。

</details>

---

## 15. 下一章预告

Tokenizer 把文本变成 id 后，**预训练目标** 是在海量 id 序列上做因果建模——14 章讲 CLM、perplexity 与训练数据格式。

---

*下一章：[14 预训练与语言模型原理](14-预训练与语言模型原理.md)*  
*推理架构：[LLMInfra 07 大模型推理引擎架构概览](../LLMInfra/07-大模型推理引擎架构概览.md)*
