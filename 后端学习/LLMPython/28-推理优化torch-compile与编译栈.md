# 推理优化 torch.compile 与编译栈

> **文件编码**：UTF-8。  
> **前置**：[05 nn.Module](05-nn.Module与训练循环.md)、[08 AMP](08-GPU训练与混合精度AMP.md)、[20 vLLM Serving](20-vLLM-TGI与LMDeploy-Python侧.md)。  
> **对照**：[LLMInfra 07 推理架构](../LLMInfra/07-大模型推理引擎架构概览.md)、[15 FlashAttention](../LLMInfra/15-FlashAttention与算子融合.md)、[23 ONNX 导出](23-模型导出ONNX与TorchScript.md)。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**torch.compile = PyTorch 2 的「即时编译器」**：把 Python eager 小算子融合成更少的 GPU kernel，常获 1.2～2× 推理加速——与 vLLM 级系统优化 **互补**，不替代 PagedAttention。

### 0.2 你需要提前知道什么

| 情况 | 建议 |
|------|------|
| 只用 HF generate | 先读 [12 章](12-HuggingFace-Transformers入门.md) |
| 不懂 eager vs graph | 先读 [LLMInfra 07 计算图](../LLMInfra/07-大模型推理引擎架构概览.md) |
| 生产 LLM 部署 | 本章 + [20 章 vLLM](20-vLLM-TGI与LMDeploy-Python侧.md) + [LLMInfra 14](../LLMInfra/14-vLLM-TensorRT-LLM-llama.cpp架构导读.md) |
| 写 CUDA kernel | 编译栈是更高层；底层见 [LLMInfra 03～05](../LLMInfra/03-GPU架构与CUDA编程入门.md) |

### 0.3 本章知识地图

- [ ] 能解释 eager、TorchScript、torch.compile 三代关系
- [ ] 会用 `torch.compile(model)` 做推理 benchmark
- [ ] 知道 Inductor / Triton 在后端的位置
- [ ] 了解 dynamic shapes、fullgraph 等参数
- [ ] 知道 LLM 场景 compile 的局限与 workaround
- [ ] 能对比 compile vs ONNX vs TensorRT

### 0.4 建议学习时长

| 阶段 | 内容 | 时间 |
|------|------|------|
| 编译栈概览 | §0～§2 | 45 分钟 |
| 实战 | §3～§4 | 1.5 小时 |
| LLM 特化 | §5～§6 | 1 小时 |
| FAQ + 自测 | §7～§8 | 30 分钟 |

### 0.5 学完本章你能做什么

1. 对 [11 章](11-Transformer从零实现-PyTorch.md) 小模型 compile 前后测 latency。
2. 解释为何 vLLM 已很快仍有人研究 `torch.compile`。
3. 面试画 PyTorch 2 编译栈分层图。

---

## 1. 为什么需要编译

### 1.1 Eager 模式开销

PyTorch 默认 **逐 op 调度**：每个 `matmul`、`softmax` 一次 kernel launch + Python 开销。小 batch 推理 **GPU 吃不饱**。

### 1.2 编译收益

| 优化 | 说明 |
|------|------|
| 算子融合 | 例：Linear+GELU 合一 kernel，对照 [LLMInfra 15 融合](../LLMInfra/15-FlashAttention与算子融合.md) |
| 常量折叠 | 编译期算死 |
| 内存规划 | 减少中间 buffer |
| Triton 代码生成 | Inductor 自动生成 GPU kernel |

### 1.3 与专用引擎对比

```text
torch.compile     → 单进程 PyTorch 模型加速，改动小
ONNX/TensorRT     → 跨框架/深度 NVIDIA 优化（23 章）
vLLM/TRT-LLM      → LLM 系统级：KV、调度、PagedAttention（LLMInfra 14～16）
```

---

## 2. PyTorch 2 编译栈架构

```text
torch.compile(model)
       ↓
  TorchDynamo (前端)     ← 捕获 Python bytecode / FX Graph
       ↓
  AOT Autograd           ← 可选，训练用
       ↓
  TorchInductor (默认)   ← 生成 Triton (GPU) / C++ (CPU)
       ↓
  CUDA kernels
```

| 组件 | 作用 |
|------|------|
| Dynamo | `frame evaluation` _hook，图断裂处 fallback eager |
| Inductor | 图级优化 + codegen |
| Triton | GPU kernel DSL，OpenAI 开源 |
| cudagraphs | 可选，减 launch 开销 |

后端选项：`backend="inductor"`（默认）、`aot_eager`（调试）、`cudagraphs` 等。

---

## 3. 最小使用

### 3.1 一行 compile

```python
import torch

model = MyModel().eval().cuda()
compiled = torch.compile(model)  # 或 model = torch.compile(model)

x = torch.randn(8, 128, device="cuda")
# 首次运行含编译 warmup，较慢
with torch.inference_mode():
    for _ in range(warmup):
        compiled(x)
    torch.cuda.synchronize()
    t0 = torch.cuda.Event(enable_timing=True)
    t1 = torch.cuda.Event(enable_timing=True)
    t0.record()
    for _ in range(100):
        compiled(x)
    t1.record()
    torch.cuda.synchronize()
    print(f"avg ms: {t0.elapsed_time(t1)/100:.3f}")
```

### 3.2 常用参数

```python
torch.compile(
    model,
    mode="reduce-overhead",   # default | reduce-overhead | max-autotune
    fullgraph=False,          # True 则不允许 graph break
    dynamic=True,             # 支持变长输入
)
```

| mode | 说明 |
|------|------|
| default | 平衡 |
| reduce-overhead | 小 batch 推理，偏 cudagraph |
| max-autotune | 慢编译，最快运行 |

### 3.3 HuggingFace 模型

```python
from transformers import AutoModelForCausalLM
import torch

model = AutoModelForCausalLM.from_pretrained(
    "Qwen/Qwen2.5-0.5B-Instruct", torch_dtype=torch.bfloat16, device_map="cuda"
).eval()

# 仅 compile 单次 forward（不含 generate 循环）
model = torch.compile(model, dynamic=True)

input_ids = torch.randint(0, 32000, (1, 128), device="cuda")
with torch.inference_mode():
    out = model(input_ids=input_ids)
```

`generate()` 含 Python 循环与 cache 逻辑，**整链 compile 较复杂**；可对 `model.forward` 或 decode step 单独 compile。

---

## 4. Graph Break 与调试

### 4.1 何谓 graph break

Dynamo 无法静态捕获的分支（Python if、某些自定义 op）会 **断图**，断点前后分别 compile/eager。

```python
import torch._dynamo as dynamo
dynamo.config.verbose = True  # 打印 break 原因
```

### 4.2 调试工具

```bash
# 查看生成的代码
TORCH_LOGS="+dynamo,inductor" python bench.py
```

### 4.3 与 23 章 TorchScript 对比

| | TorchScript | torch.compile |
|--|-------------|---------------|
| 时代 | PyTorch 1.x | PyTorch 2.x |
| 捕获 | trace/script | Dynamo bytecode |
| 维护 | 维护模式 | **主推** |
| LLM | 少 | 活跃研究 |

---

## 5. LLM 推理场景

### 5.1 Prefill vs Decode

见 [LLMInfra 02](../LLMInfra/02-Transformer与注意力机制原理.md)、[16 调度](../LLMInfra/16-推理Batch调度与ContinuousBatching.md)：

| 阶段 | 特点 | compile 收益 |
|------|------|--------------|
| Prefill | 长 seq 一次 forward | 高（大 matmul） |
| Decode | batch 小、逐 token | 中（launch 开销） |

### 5.2 与 vLLM 关系

vLLM 自有 CUDA kernel + PagedAttention；**不依赖** torch.compile。研究路线：`torch.compile` 优化 attention 子图 + vLLM 调度并存。

### 5.3 实践建议

| 场景 | 建议 |
|------|------|
| 本地 HF generate 0.5B～7B | 可试 compile forward，测首 token 与吞吐 |
| 生产 API | [20 章 vLLM](20-vLLM-TGI与LMDeploy-Python侧.md) |
| Embedding batch | compile 收益明显 |
| 训练 | `torch.compile` 亦可加速，但需测 numerics |

### 5.4 cudagraphs 注意

固定 shape 时 `reduce-overhead` 用 cuda graph；变长 batch 需 `dynamic=True` 或禁用。

---

## 6. CPU 与其他后端

### 6.1 CPU Inductor

```python
model = torch.compile(model.cpu(), backend="inductor")
```

适合 [27 章 CLIP](27-多模态CLIP入门.md) 小模型 CPU 推理。

### 6.2 OpenVINO / XPU

Intel / 其他硬件有专用 backend；查 PyTorch 文档 `backend=` 列表。

---

## 7. 性能 Checklist

```text
☑ warmup ≥ 3 次（含编译）
☑ torch.cuda.synchronize() 再计时
☑ 对比 eager vs compile 同 dtype（bf16）
☑ 扫 batch / seq_len 多维
☑ 检查 numerics：max diff < 1e-2（fp16）
☑ 生产前长时间 soak 测 OOM / 泄漏
```

与 [LLMInfra 17 Nsight](../LLMInfra/17-GPU性能剖析Nsight与perf.md) 结合找剩余瓶颈。

---

## 8. FAQ

**Q1：首次运行很慢？**  
JIT 编译正常；部署需 **warmup** 或 **预编译** save（高级 API 在演进）。

**Q2：compile 后 OOM？**  
max-autotune 编译期占显存；换 default 或减 batch。

**Q3：训练能用吗？**  
可以 `torch.compile` 包 training_step；注意梯度数值与 checkpoint 兼容。

**Q4：和 FlashAttention 关系？**  
FA 是手写融合 attention kernel；Inductor 可能自动融合部分 pattern，FA 仍常更快。

**Q5：Dynamic shape 编译次数？**  
每个新 shape 可能重编译；过多 shape 反而慢，可 bucketing。

**Q6：Java Agent 需要懂吗？**  
Servicing Python 侧或面试 Infra 岗需要；纯 Spring 开发了解概念即可。

**Q7：compile vs quantization？**  
正交；[LLMInfra 09 量化](../LLMInfra/09-模型量化INT8-INT4-FP8与校准.md) 减内存带宽，compile 减 kernel 数。

**Q8：24 章项目要加吗？**  
可选 benchmark 小节：Gradio 后端 compile 前后 latency 表。

---

## 9. 闭卷自测

1. torch.compile 默认后端是什么？
2. Dynamo 的作用？
3. 为何首次 inference 慢？
4. `mode="max-autotune"`  tradeoff？
5. graph break 是什么？
6. LLM 生产为何仍首选 vLLM？
7. Prefill 与 Decode 哪个更易从 compile 受益？
8. torch.compile 与 [23 章 ONNX](23-模型导出ONNX与TorchScript.md) 选型差异？

<details>
<summary>参考答案</summary>

1. inductor（TorchInductor + Triton on GPU）。
2. 捕获 Python 执行中的 tensor 操作为 FX 图供后端优化。
3. 含 JIT 编译与 autotune 搜索最优 kernel。
4. 编译更慢、占资源，运行更快。
5. 图捕获中断，回退 eager，优化范围缩小。
6. vLLM 有 PagedAttention、continuous batching 等系统级 LLM 优化。
7. Prefill（大矩阵乘）；Decode 受 launch 开销限制但 reduce-overhead 仍有帮助。
8. compile 改一行 PyTorch 代码；ONNX 导出静态图跨运行时、适合 TRT 部署链。

</details>

---

## 10. 系列总结

```text
01～08  PyTorch 基础
11～16  Transformer 与微调
17～20  分布式与 Serving
21～23  应用 / MLOps / 导出
24～25  项目与面试
26～28  Lightning / CLIP / compile  ← 进阶工程与扩展
```

回到 [00 学习路线图](00-学习路线图与说明.md)；Infra 深化 [LLMInfra 00](../LLMInfra/00-学习路线图与说明.md)；产品化 [AIAgent 00](../AIAgent/00-学习路线图与说明.md)。

---

## 11. 下一章

[29 HuggingFace TRL 与 SFTTrainer 实战](29-HuggingFace-TRL与SFTTrainer实战.md) — 进阶微调与对齐工具链。

复习：[25 面试专题与知识点总表](25-面试专题与知识点总表.md) · [00 路线图](00-学习路线图与说明.md)
