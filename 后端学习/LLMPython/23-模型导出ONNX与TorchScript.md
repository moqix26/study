# 模型导出 ONNX 与 TorchScript

> **文件编码**：UTF-8。  
> **前置**：[05 nn.Module](05-nn.Module与训练循环.md)、[12 HuggingFace](12-HuggingFace-Transformers入门.md)、[15 LoRA](15-微调SFT与LoRA-PEFT.md)。  
> **对照**：[LLMInfra 09 量化](../LLMInfra/09-模型量化INT8-INT4-FP8与校准.md)、[14 推理引擎](../LLMInfra/14-vLLM-TensorRT-LLM-llama.cpp架构导读.md)；Serving 见 [20 章 vLLM](20-vLLM-TGI与LMDeploy-Python侧.md)。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**导出 = 把 PyTorch 动态训练图变成可部署的静态推理程序**：ONNX 跨框架交换；TorchScript 留在 PyTorch 生态；大 LLM 生产更常走 **TensorRT-LLM / vLLM**，本章打地基。

### 0.2 你需要提前知道什么

| 情况 | 建议 |
|------|------|
| 不懂 `model.eval()` | 先读 [05 章](05-nn.Module与训练循环.md) |
| 只用过 HF `generate` | 先读 [12 章](12-HuggingFace-Transformers入门.md) |
| 目标 C++ 推理 | 本章 + [LLMInfra 13 pybind](../LLMInfra/13-pybind11与Python-C++混合编程.md) |
| 7B+ 大模型部署 | 优先 [20 章 vLLM](20-vLLM-TGI与LMDeploy-Python侧.md)，导出作补充 |

### 0.3 本章知识地图

- [ ] 能区分 TorchScript trace vs script
- [ ] 会用 `torch.onnx.export` 导出小模型
- [ ] 知道 LLM 导出的难点（动态 shape、KV Cache）
- [ ] 会用 ONNX Runtime 跑推理对比 latency
- [ ] 能说明 ONNX vs TorchScript vs safetensors 选型
- [ ] 知道 TensorRT / OpenVINO 在栈中的位置

### 0.4 建议学习时长

| 阶段 | 内容 | 时间 |
|------|------|------|
| 概念 | §0～§2 | 40 分钟 |
| TorchScript | §3 | 1 小时 |
| ONNX | §4～§5 | 1.5 小时 |
| LLM 特化 | §6～§7 | 1 小时 |

### 0.5 学完本章你能做什么

1. 导出 BERT/小 CNN 的 ONNX 并在 ORT 上验证数值一致。
2. 解释为何完整 Decoder-only LLM 很少直接 ONNX 端到端部署。
3. 把导出流程写入 [24 章项目](24-项目实战微调小型语言模型.md) 可选交付。

### 0.6 部署栈位置

```text
PyTorch 训练（12～15）→ merge LoRA
  ↓
导出（本章 23）→ ONNX / TorchScript / TensorRT plan
  ↓
Python ORT / C++ TRT（LLMInfra 14）→ 生产 Serving（20 / AIAgent）
```

---

## 1. 为什么需要导出

| 需求 | 导出价值 |
|------|----------|
| 脱离 Python 训练栈 | C++/Java/移动端推理 |
| 图优化 | 算子融合、常量折叠 |
| 硬件后端 | TensorRT、OpenVINO、CoreML |
| 版本锁定 | 推理图固定，减少依赖 |

**注意**：HuggingFace `safetensors` 是 **权重格式**，不是计算图；vLLM 直接读 safetensors + 自家 kernel，见 [LLMInfra 12 mmap](../LLMInfra/12-Checkpoint加载与mmap权重IO.md)。

---

## 2. 推理模式基础

```python
model.eval()
with torch.inference_mode():  # 比 no_grad 更省开销
    output = model(input_ids)
```

| 模式 | 说明 |
|------|------|
| 训练 | dropout 开、autograd 开 |
| 推理 | dropout 关、固定 BN 统计 |
| 半精度 | `model.half()` 或 autocast，对照 [08 AMP](08-GPU训练与混合精度AMP.md) |

---

## 3. TorchScript

### 3.1 trace vs script

| 方法 | 原理 | 局限 |
|------|------|------|
| `torch.jit.trace` | 用样例输入录操作 | 控制流随输入变会错 |
| `torch.jit.script` | 解析 Python 子集编译 | 不支持任意 Python |

```python
import torch

class TinyNet(torch.nn.Module):
    def __init__(self):
        super().__init__()
        self.fc = torch.nn.Linear(768, 10)

    def forward(self, x):
        return self.fc(x)

model = TinyNet().eval()
example = torch.randn(1, 768)

# Trace
traced = torch.jit.trace(model, example)
traced.save("tinynet_traced.pt")

# Script（含简单 if）
scripted = torch.jit.script(model)
scripted.save("tinynet_scripted.pt")

# 加载（C++ 可读 traced/scripted）
loaded = torch.jit.load("tinynet_traced.pt")
```

### 3.2 C++ 推理（LibTorch）

```cpp
#include <torch/script.h>
torch::jit::script::Module module = torch::jit::load("tinynet_traced.pt");
auto input = torch::randn({1, 768});
auto output = module.forward({input}).toTensor();
```

与 [LLMInfra 13 pybind](../LLMInfra/13-pybind11与Python-C++混合编程.md) 组合可包成 Python 扩展。

### 3.3 HF 模型导出 TorchScript

小 Encoder（如 BERT）可 trace；GPT 类 **自回归 + KV** 通常不整模 trace，只导出 embedding 或单层做实验。

---

## 4. ONNX 导出

### 4.1 基本流程

```python
import torch
from transformers import AutoModelForSequenceClassification

model = AutoModelForSequenceClassification.from_pretrained(
    "bert-base-uncased", num_labels=2
).eval()
dummy_input = {
    "input_ids": torch.randint(0, 30000, (1, 128)),
    "attention_mask": torch.ones(1, 128, dtype=torch.long),
}

torch.onnx.export(
    model,
    (dummy_input["input_ids"], dummy_input["attention_mask"]),
    "bert_cls.onnx",
    input_names=["input_ids", "attention_mask"],
    output_names=["logits"],
    dynamic_axes={
        "input_ids": {0: "batch", 1: "seq"},
        "attention_mask": {0: "batch", 1: "seq"},
        "logits": {0: "batch"},
    },
    opset_version=17,
    do_constant_folding=True,
)
```

### 4.2 dynamic_axes 含义

LLM 的 seq_len、batch 变化大，必须声明动态维，否则只能固定 128 token。

### 4.3 验证数值一致

```python
import onnxruntime as ort
import numpy as np

sess = ort.InferenceSession("bert_cls.onnx", providers=["CPUExecutionProvider"])
ort_out = sess.run(None, {
    "input_ids": dummy_input["input_ids"].numpy(),
    "attention_mask": dummy_input["attention_mask"].numpy(),
})
with torch.inference_mode():
    pt_out = model(**dummy_input).logits.numpy()
np.testing.assert_allclose(pt_out, ort_out[0], rtol=1e-3, atol=1e-4)
```

---

## 5. ONNX Runtime 推理

### 5.1 Provider 选型

| Provider | 场景 |
|----------|------|
| CPUExecutionProvider | 开发机、无 GPU |
| CUDAExecutionProvider | NVIDIA GPU |
| TensorrtExecutionProvider | ORT + TRT 融合 |

```python
providers = ["CUDAExecutionProvider", "CPUExecutionProvider"]
sess = ort.InferenceSession("model.onnx", providers=providers)
```

### 5.2 性能对比模板

```python
import time
for _ in range(warmup):
    sess.run(None, feeds)
t0 = time.perf_counter()
for _ in range(n):
    sess.run(None, feeds)
print(f"ORT avg ms: {(time.perf_counter()-t0)/n*1000:.2f}")
```

与 PyTorch eager 对比；大 LLM 见 [28 章 torch.compile](28-推理优化torch-compile与编译栈.md)。

---

## 6. 大语言模型导出的现实

### 6.1 难点

| 难点 | 说明 |
|------|------|
| 自回归循环 | 导出通常是单步 forward，外层循环在外部 |
| KV Cache | 动态增长 cache 形状，ONNX 支持有限 |
| 变长 prompt | 需 dynamic_axes 或 padding 策略 |
| 自定义算子 | RoPE、GQA 可能无 ONNX op |
| 量化 | 需 QDQ 或 PTQ 流程，见 [LLMInfra 09](../LLMInfra/09-模型量化INT8-INT4-FP8与校准.md) |

### 6.2 工业界常见路径

```text
HF safetensors
  → TensorRT-LLM / vLLM / llama.cpp（不经过通用 ONNX）
  → 或 Optimum + ONNX 导出 Encoder 部分（Embedding/Reranker）
```

| 组件 | 适合 ONNX |
|------|-----------|
| Embedding 模型 | ✅ 常导出 |
| Cross-Encoder Reranker | ✅ |
| 7B Decoder 全序列 | ⚠️ 实验性 |
| 70B+ | ❌ 用专用引擎 |

### 6.3 Optimum 示例（Embedding）

```bash
pip install optimum[onnxruntime]
optimum-cli export onnx --model BAAI/bge-small-zh-v1.5 onnx/bge-small/
```

用于 [21 章 RAG](21-LangChain与LlamaIndex应用层.md) 检索侧加速。

---

## 7. LoRA 合并后导出

```python
from peft import PeftModel

base = AutoModelForCausalLM.from_pretrained("Qwen/Qwen2.5-0.5B-Instruct")
model = PeftModel.from_pretrained(base, "./lora_adapter")
merged = model.merge_and_unload()
merged.save_pretrained("./merged_model")
# 再尝试 export 单步 forward 或交给 vLLM
```

[22 章 artifact](22-MLOps与实验跟踪wandb-mlflow.md) 应同时保留 adapter 与 merged 版本。

---

## 8. 格式选型表

| 格式 | 跨语言 | 典型用途 |
|------|--------|----------|
| safetensors | 需框架加载 | HF 训练、vLLM |
| TorchScript | C++ LibTorch | 传统 CV、小模型 |
| ONNX | 最广 | 移动端、ORT、TRT 前置 |
| TensorRT engine | NVIDIA | 生产低延迟 |
| GGUF | llama.cpp | CPU/边缘，见 LLMInfra 09 |

---

## 9. 与 Java / Agent 栈

[AIAgent](../AIAgent/00-学习路线图与说明.md) 通常 **不直接加载 ONNX**；Java 调 Python gRPC 服务或云端 API。若 Embedding  ONNX 化，可由独立 Python 微服务暴露 REST，Spring AI 仍用 HTTP EmbeddingClient。

---

## 10. FAQ

**Q1：ONNX opset 选多少？**  
越新支持算子越多；ORT 1.17+ 常用 opset 17；导出失败时尝试降低或升级 ORT。

**Q2：trace 结果不对怎么办？**  
换 script；或检查 `training` 模式、随机 op、Python 侧控制流。

**Q3：LLM 能用 ONNX 部署吗？**  
Reranker/Embedding 可以；完整生成推荐 vLLM/TensorRT-LLM，见 [LLMInfra 14](../LLMInfra/14-vLLM-TensorRT-LLM-llama.cpp架构导读.md)。

**Q4：导出后体积变大？**  
ONNX 未压缩；可用 onnx simplifier；权重量化另走 PTQ。

**Q5：TorchScript 和 torch.compile 关系？**  
compile 是 PyTorch 2 新编译栈，见 [28 章](28-推理优化torch-compile与编译栈.md)；与 JIT 不同路线。

**Q6：如何测导出是否成功？**  
数值对齐 + 延迟 benchmark + 边界 shape（batch=1/8, seq=32/512）。

**Q7：24 章项目必须导出吗？**  
可选；必做 HF 权重 + Gradio。

---

## 11. 闭卷自测

1. safetensors 与 ONNX 区别？
2. trace 与 script 各适合什么模型？
3. `dynamic_axes` 解决什么问题？
4. 为何 LLM 全模型 ONNX 少见？
5. LoRA 导出前为什么要 merge？
6. ORT 中 CUDAExecutionProvider 作用？
7. 工业界 7B LLM 更常走哪条部署路径？
8. Embedding ONNX 化对 RAG 有何好处？

<details>
<summary>参考答案</summary>

1. safetensors 是权重存储；ONNX 是含计算图的跨框架格式。
2. trace 适合固定控制流；script 适合可编译的 Python 子集与控制流。
3. 声明 batch/seq 等可变维度，避免仅支持固定输入大小。
4. 自回归+KV 动态 shape、自定义算子、专用引擎性能更好。
5. 合并后得到完整权重，便于单文件导出与 vLLM 加载。
6. 在 NVIDIA GPU 上跑 ONNX 图加速推理。
7. safetensors + vLLM / TensorRT-LLM / llama.cpp。
8. 检索侧 batch 编码更快、可脱离 PyTorch 训练依赖部署。

</details>

---

## 12. 下一章

[24 项目实战：微调小型语言模型](24-项目实战微调小型语言模型.md)

并行：[LLMInfra 09 量化](../LLMInfra/09-模型量化INT8-INT4-FP8与校准.md)、[28 torch.compile](28-推理优化torch-compile与编译栈.md)。
