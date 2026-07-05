# vLLM、TGI 与 LMDeploy（Python 侧）

> **文件编码**：UTF-8。  
> **前置**：[12 HuggingFace](12-HuggingFace-Transformers入门.md)、[15 LoRA 合并](15-微调SFT与LoRA-PEFT.md)、[19 评估](19-模型评估与Benchmark.md)。  
> **定位**：用 Python 调用 **vLLM / TGI / LMDeploy** 部署推理，理解 OpenAI 兼容 API 与 batch infer；底层原理见 [LLMInfra 14～16](../LLMInfra/14-vLLM-TensorRT-LLM-llama.cpp架构导读.md)。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**Serving 引擎** = 把 HuggingFace `generate` 换成 **PagedAttention + continuous batching**，在相同 GPU 上获得更高 QPS 与更低延迟。

### 0.2 你需要提前知道什么

- `model.generate` 与 chat_template（12～13 章）
- HTTP / REST 基本概念
- 合并后的模型权重路径（15 章）

### 0.3 本章知识地图（☐→☑）

- [ ] 用 vLLM `LLM` 类 batch 推理
- [ ] 启动 OpenAI 兼容 `/v1/chat/completions`
- [ ] 了解 TGI Docker 启动参数
- [ ] 了解 LMDeploy pipeline 基本用法
- [ ] 对比 HF generate 与 vLLM 吞吐（概念）
- [ ] 完成 §14 闭卷自测 ≥8/10

### 0.4 建议学习时长

- **4～6 天**

---

## 1. 这份文档学什么

- 为何需要专用推理引擎
- vLLM：`LLM`、`SamplingParams`、OpenAI API server
- Text Generation Inference（TGI）：HF 官方栈
- LMDeploy：OpenMMLab、TurboMind
- OpenAI 兼容客户端调用
- batch infer、流式 SSE
- 量化模型加载（AWQ/GPTQ）
- 与 [Infra 08 PagedAttention](../LLMInfra/08-KVCache与PagedAttention原理.md)、[Infra 16 Continuous Batching](../LLMInfra/16-推理Batch调度与ContinuousBatching.md) 对照

---

## 2. 为何不用 HF generate 上线

| 问题 | HF generate | vLLM 等 |
|------|-------------|---------|
| KV Cache | 静态预分配 | PagedAttention 块管理 |
| Batch | 静态 padding | Continuous batching |
| 吞吐 | 低 | 高（尤其多用户） |
| 并发 | 差 | 请求级调度 |

Python 训练用 Transformers；**生产推理** 常用专用引擎（或云 API）。

---

## 3. vLLM Python API

**安装**：

```bash
pip install vllm
```

**离线 batch 推理**：

```python
from vllm import LLM, SamplingParams

llm = LLM(
    model="Qwen/Qwen2.5-0.5B-Instruct",
    dtype="bfloat16",
    max_model_len=4096,
    gpu_memory_utilization=0.9,
)

prompts = [
    "Hello, my name is",
    "The capital of France is",
    "解释一下 PagedAttention：",
]

params = SamplingParams(temperature=0.8, top_p=0.95, max_tokens=128)
outputs = llm.generate(prompts, params)

for out in outputs:
    print(out.prompt)
    print(out.outputs[0].text)
    print("---")
```

**Chat 格式**：传入已 `apply_chat_template` 的字符串，或 vLLM 新版 `chat` API：

```python
from vllm import LLM
from vllm.sampling_params import SamplingParams

llm = LLM(model="Qwen/Qwen2.5-0.5B-Instruct")
messages = [{"role": "user", "content": "什么是 KV Cache？"}]
# 部分版本支持 llm.chat(messages, sampling_params=...)
```

查阅当前 vLLM 文档确认 `chat` 接口版本差异。

---

## 4. vLLM OpenAI 兼容服务

**启动**：

```bash
python -m vllm.entrypoints.openai.api_server \
  --model Qwen/Qwen2.5-0.5B-Instruct \
  --dtype bfloat16 \
  --max-model-len 4096 \
  --port 8000
```

**Python 客户端**：

```python
from openai import OpenAI

client = OpenAI(base_url="http://localhost:8000/v1", api_key="EMPTY")

resp = client.chat.completions.create(
    model="Qwen/Qwen2.5-0.5B-Instruct",
    messages=[
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "用三句话介绍 vLLM。"},
    ],
    temperature=0.7,
    max_tokens=256,
)
print(resp.choices[0].message.content)
```

**流式**：

```python
stream = client.chat.completions.create(..., stream=True)
for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

与 [AIAgent 03 SSE](../AIAgent/03-流式对话与SSE实战.md) 产品层衔接。

---

## 5. 量化与 LoRA

```python
llm = LLM(
    model="path/to/awq-model",
    quantization="awq",
)
```

- AWQ/GPTQ 权重见 [Infra 09](../LLMInfra/09-模型量化INT8-INT4-FP8与校准.md)
- LoRA：`enable_lora=True` + `--lora-modules name=path`（视 vLLM 版本）
- **部署前** 15 章 `merge_and_unload` 最省事

---

## 6. Text Generation Inference（TGI）

**Docker 启动**：

```bash
docker run --gpus all -p 8080:80 \
  -v $PWD/models:/data \
  ghcr.io/huggingface/text-generation-inference:latest \
  --model-id Qwen/Qwen2.5-0.5B-Instruct \
  --dtype bfloat16 \
  --max-input-length 2048 \
  --max-total-tokens 4096
```

**Python 请求**：

```python
import requests

resp = requests.post(
    "http://localhost:8080/generate",
    json={
        "inputs": "My name is",
        "parameters": {"max_new_tokens": 50, "temperature": 0.7},
    },
)
print(resp.json()["generated_text"])
```

TGI 与 HF 生态集成深；多模型生产见官方 `router` 文档。

---

## 7. LMDeploy（了解）

**Pipeline API**：

```python
from lmdeploy import pipeline, TurbomindEngineConfig, GenerationConfig

pipe = pipeline(
    "Qwen/Qwen2.5-0.5B-Instruct",
    backend_config=TurbomindEngineConfig(tp=1, session_len=4096),
)

gen_config = GenerationConfig(top_p=0.8, temperature=0.7, max_new_tokens=128)
response = pipe(["你好，介绍一下 LMDeploy。"], gen_config=gen_config)
print(response[0].text)
```

**OpenAI 服务**：

```bash
lmdeploy serve api_server Qwen/Qwen2.5-0.5B-Instruct --server-port 23333
```

适合 OpenMMLab 栈与国内文档；原理同为 C++/CUDA 引擎 + Python 封装。

---

## 8. Batch Infer 模式

**离线评测**（19 章）：大量 prompt 一次 `llm.generate(prompts)`，vLLM 内部 continuous batch。

**在线服务**：请求异步到达，scheduler 合并 prefill/decode——见 [Infra 16](../LLMInfra/16-推理Batch调度与ContinuousBatching.md)。

```python
# 批量大小由引擎调度，非简单 Python for 循环
prompts = [f"问题 {i}" for i in range(1000)]
outputs = llm.generate(prompts, SamplingParams(max_tokens=64))
```

**注意**：`max_num_seqs` 限制并发序列数；OOM 时调低 `gpu_memory_utilization`。

---

## 9. 选型对照

| 引擎 | 优势 | 场景 |
|------|------|------|
| vLLM | 社区活跃、OpenAI API | 通用 LLM 服务 |
| TGI | HF 官方、稳定 | Hub 模型一键部署 |
| LMDeploy | TurboMind 性能 | 国产模型优化 |
| TensorRT-LLM | 极致延迟 | Infra 14 C++ 栈 |

本章 Python 侧以 **vLLM 为主**；C++ 内核读 Infra 14。

---

## 10. 监控与调参

| 参数 | 作用 |
|------|------|
| `max_model_len` | 最大 context，影响 KV 显存 |
| `gpu_memory_utilization` | 预占显存比例 |
| `tensor_parallel_size` | 多卡切模型（TP） |
| `enforce_eager` | 调试禁用 CUDA graph |

**指标**：tokens/s、TTFT（首 token 延迟）、P99 latency、GPU util。

生产部署延伸：[Infra 18 K8s GPU](../LLMInfra/18-容器化与Kubernetes-GPU推理部署.md)。

---

## 11. 练习建议

1. 本地起 vLLM API，用 OpenAI SDK 发 10 条 chat
2. 同一模型对比 HF `generate` 与 vLLM 100 条 prompt 总耗时
3. 开 stream=True 接简单 CLI 聊天
4. 调 `max_model_len` 观察 OOM 边界
5. 合并 15 章 LoRA 后部署 vLLM
6. 读 Infra 08 一节 PagedAttention，对应 vLLM 文档一句话

---

## 12. 学完标准

- [ ] 写出 vLLM batch 推理最小脚本
- [ ] 启动 OpenAI 兼容 server 并用 client 调用
- [ ] 解释 continuous batching 相对静态 batch 优势
- [ ] 知道 TGI 与 LMDeploy 各一条启动命令
- [ ] 说明 merge LoRA 后再部署的原因

---

## 13. FAQ

**Q1：vLLM 支持 CPU 吗？**  
主要 GPU；CPU 推理看 llama.cpp（Infra 14）。

**Q2：和 Ollama 关系？**  
Ollama 是本地产品封装；底层也可调 llama.cpp 等。

**Q3：OpenAI API 的 model 名填什么？**  
通常填 HF model id 或启动时 `--served-model-name`。

**Q4：多卡推理怎么设？**  
vLLM `tensor_parallel_size=2`；与训练 DDP 不同（17 章）。

**Q5：chat_template 谁做？**  
server 端或 client 先 template；须与训练一致（13 章）。

**Q6：BF16 与 FP16？**  
A100/H100 优先 bf16；老卡 fp16。

**Q7：能否 batch 不同 max_tokens？**  
引擎支持 per-request `max_tokens`；调度器统一 batch decode。

**Q8：量化后精度掉多少？**  
用 19 章 benchmark 实测；AWQ 常小幅下降。

**Q9：Python 绑定层性能？**  
瓶颈在 GPU kernel；Python 仅调度（Infra 13 pybind 对照）。

**Q10：生产还要什么？**  
鉴权、限流、日志、K8s HPA、模型版本回滚（Infra 18、AIAgent 11）。

---

## 14. 闭卷自测

1. PagedAttention 主要解决什么问题？
2. vLLM 相对 HF generate 两大吞吐优势？
3. OpenAI 兼容 API 常见路径？
4. `SamplingParams.max_tokens` 含义？
5. TGI 常见容器端口映射？
6. merge LoRA 的目的？
7. TTFT 指什么？
8. `tensor_parallel_size` 用于训练还是推理切分？
9. continuous batching 指什么？
10. 本章三引擎 Python 侧各一个类/命令？

<details>
<summary>参考答案</summary>

1. KV Cache 显存碎片与过度预分配，块式管理显存。
2. PagedAttention + continuous batching（及 fused kernel 等）。
3. `/v1/chat/completions`（OpenAI 格式）。
4. 单次生成最多新 token 数。
5. 常 `-p 8080:80` 映射到容器 80。
6. 合并 adapter 为单一权重，便于引擎加载与量化。
7. Time To First Token，首 token 延迟。
8. 推理时模型张量并行切分（也可训练 TP，本章语境为推理）。
9. 请求动态加入/离开 batch，提高 GPU 利用率。
10. vLLM `LLM`；TGI `docker run ... text-generation-inference`；LMDeploy `pipeline()` / `lmdeploy serve api_server`。

</details>

---

## 15. 下一章预告

20 章完成 **训练→评估→Serving** 闭环——可继续 21 章 LangChain 应用层，或 24 章完整项目；Infra 双栈读 [LLMInfra 19 MiniInfer](../LLMInfra/19-项目实战简易推理引擎.md)。

---

*Infra 深入：[14 架构导读](../LLMInfra/14-vLLM-TensorRT-LLM-llama.cpp架构导读.md) · [16 Batch 调度](../LLMInfra/16-推理Batch调度与ContinuousBatching.md)*  
*项目实战：[24 微调小型语言模型](24-项目实战微调小型语言模型.md)（路线图）*
