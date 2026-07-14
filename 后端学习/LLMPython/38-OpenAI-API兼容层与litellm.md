# OpenAI API 兼容层与 litellm

> **文件编码**：UTF-8。  
> **前置**：[12 HuggingFace](12-HuggingFace-Transformers入门.md)、[20 vLLM/TGI](20-vLLM-TGI与LMDeploy-Python侧.md)、[21 LangChain](21-LangChain与LlamaIndex应用层.md)。  
> **对照**：[AIAgent 03 网关](../AIAgent/02-Go模型调用与结构化输出.md)；[LLMInfra 12 OpenAI 兼容](../LLMInfra/12-OpenAI兼容API与推理协议.md)。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**OpenAI 兼容层 = 同一套 SDK / HTTP 调 OpenAI、vLLM、Azure、Ollama**；litellm 在其上做路由、fallback、成本统计与 function calling 统一。

### 0.2 你需要提前知道什么

| 情况 | 建议 |
|------|------|
| 只会 `model.generate` | 先读 [20 章](20-vLLM-TGI与LMDeploy-Python侧.md) |
| 做 Agent | 结合 [21 章](21-LangChain与LlamaIndex应用层.md) |
| Java 后端 | 对照 AIAgent 网关 |

### 0.3 本章知识地图（☐→☑）

- [ ] OpenAI SDK chat completions + streaming
- [ ] vLLM server + 改 `base_url`
- [ ] litellm 路由与 fallback
- [ ] function calling 两轮对话
- [ ] 完成 §12 闭卷自测 ≥8/10

### 0.4 建议学习时长

OpenAI SDK 1.5h · vLLM 1h · litellm 2h · function calling 1.5h

### 0.5 学完本章你能做什么

1. 本地 vLLM + OpenAI SDK，零改业务切换云 API。
2. litellm proxy 接 GPT-4 与本地 Qwen 兜底。
3. 说清「兼容层 vs 网关 vs 推理引擎」三层分工。

---

## 1. OpenAI API 协议概览

| 端点 | 用途 |
|------|------|
| `POST /v1/chat/completions` | 对话补全（最常用） |
| `POST /v1/embeddings` | 向量嵌入 |
| `GET /v1/models` | 列出模型 |

**finish_reason**：`stop`、`length`（触达 max_tokens）、`tool_calls`。

---

## 2. OpenAI Python SDK

```bash
pip install openai
```

```python
from openai import OpenAI

client = OpenAI(api_key="sk-...")  # 或 OPENAI_API_KEY

resp = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "1+1=?"}],
    temperature=0,
)
print(resp.choices[0].message.content)
print(resp.usage.total_tokens)
```

**流式**：

```python
stream = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "写短诗"}],
    stream=True,
)
for chunk in stream:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="", flush=True)
```

高并发 Agent 用 `AsyncOpenAI` + `asyncio.gather`。

---

## 3. 指向自建服务：base_url

```python
client = OpenAI(
    base_url="http://localhost:8000/v1",
    api_key="EMPTY",
)
resp = client.chat.completions.create(
    model="Qwen/Qwen2.5-0.5B-Instruct",
    messages=[{"role": "user", "content": "你好"}],
)
```

| 参数 | 说明 |
|------|------|
| `base_url` | 含 `/v1` 前缀 |
| `model` | 与 server `--served-model-name` 一致 |

```bash
export OPENAI_BASE_URL=http://localhost:8000/v1
export OPENAI_API_KEY=EMPTY
```

---

## 4. vLLM OpenAI Compatible Server

```bash
python -m vllm.entrypoints.openai.api_server \
  --model Qwen/Qwen2.5-0.5B-Instruct \
  --dtype bfloat16 \
  --max-model-len 4096 \
  --port 8000
```

| 维度 | HF `generate` | vLLM OpenAI server |
|------|---------------|---------------------|
| 调用 | 进程内 Python | HTTP / SDK |
| 并发 | 差 | continuous batching |
| 跨语言 | 需自建 | 直接 HTTP |

其他兼容：**Ollama** `http://localhost:11434/v1`、**TGI**、**Azure OpenAI**（`AzureOpenAI` 类）。

---

## 5. litellm 入门

```text
业务（OpenAI 格式）→ litellm.completion() → OpenAI / vLLM / Azure / ...
```

```bash
pip install litellm
```

```python
from litellm import completion

# 云 API
resp = completion(model="gpt-4o-mini", messages=[{"role": "user", "content": "hi"}])

# 本地 vLLM
resp = completion(
    model="openai/Qwen/Qwen2.5-0.5B-Instruct",
    api_base="http://localhost:8000/v1",
    api_key="EMPTY",
    messages=[{"role": "user", "content": "hi"}],
)
```

model 格式：`provider/model_name`（如 `ollama/qwen2.5`）。

---

## 6. litellm 路由与 Proxy

```python
from litellm import Router

router = Router(
    model_list=[
        {"model_name": "smart-chat", "litellm_params": {"model": "gpt-4o-mini", "api_key": "sk-..."}},
        {"model_name": "smart-chat", "litellm_params": {
            "model": "openai/Qwen/Qwen2.5-0.5B-Instruct",
            "api_base": "http://localhost:8000/v1", "api_key": "EMPTY",
        }},
    ],
    fallbacks=[{"smart-chat": ["gpt-4o-mini", "openai/Qwen/Qwen2.5-0.5B-Instruct"]}],
)
resp = router.completion(model="smart-chat", messages=[{"role": "user", "content": "hello"}])
```

**Proxy Server**：

```bash
pip install 'litellm[proxy]'
litellm --config config.yaml --port 4000
```

```yaml
model_list:
  - model_name: gpt-4o-mini
    litellm_params: {model: gpt-4o-mini, api_key: os.environ/OPENAI_API_KEY}
  - model_name: local-qwen
    litellm_params:
      model: openai/Qwen/Qwen2.5-0.5B-Instruct
      api_base: http://localhost:8000/v1
      api_key: EMPTY
router_settings:
  fallbacks: [{"gpt-4o-mini": ["local-qwen"]}]
```

客户端 `base_url=http://localhost:4000` 即 OpenAI 兼容——类似 [AIAgent 03](../AIAgent/02-Go模型调用与结构化输出.md)。

---

## 7. LangChain 集成

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    model="Qwen/Qwen2.5-0.5B-Instruct",
    openai_api_base="http://localhost:8000/v1",
    openai_api_key="EMPTY",
)
print(llm.invoke("什么是 RoPE？"))
```

---

## 8. Function Calling

### 8.1 流程

```text
User 提问 → model 返回 tool_calls → 宿主执行函数
  → tool 结果写入 messages → model 生成最终回答
```

### 8.2 示例

```python
tools = [{
    "type": "function",
    "function": {
        "name": "get_weather",
        "description": "查询城市天气",
        "parameters": {
            "type": "object",
            "properties": {"city": {"type": "string"}},
            "required": ["city"],
        },
    },
}]

resp = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "北京天气？"}],
    tools=tools,
    tool_choice="auto",
)

msg = resp.choices[0].message
if msg.tool_calls:
    call = msg.tool_calls[0]
    args = json.loads(call.function.arguments)
    tool_result = get_weather(**args)  # mock

    messages = [
        {"role": "user", "content": "北京天气？"},
        msg.model_dump(),
        {"role": "tool", "tool_call_id": call.id,
         "content": json.dumps(tool_result, ensure_ascii=False)},
    ]
    final = client.chat.completions.create(model="gpt-4o-mini", messages=messages, tools=tools)
    print(final.choices[0].message.content)
```

| 本地模型注意 | 建议 |
|--------------|------|
| JSON 格式差 | few-shot；强模型 |
| vLLM | 查 `--enable-auto-tool-choice` |
| litellm | `completion(..., tools=tools)` 透传 |

**JSON mode** 约束输出格式；**tools** 是模型选函数、宿主执行——二者不同。

---

## 9. 生产与架构

超时重试（`timeout=60` + tenacity）、429 backoff、Key 仅环境变量。三层：**应用**（LangChain）→ **路由**（litellm/网关）→ **推理**（vLLM/云）。

---

## 10. FAQ 与练习

**练习**：vLLM + SDK 流式；litellm fallback；mock function calling 两轮对话。

**Q1：base_url 要 `/v1` 吗？** 通常 `http://host:8000/v1`。  
**Q2：model 404？** 查 `--served-model-name`。  
**Q3：litellm vs LangChain？** 前者偏路由 proxy；后者偏链式应用。  
**Q4：function calling vs JSON mode？** 前者调外部函数；后者约束输出 JSON。

---

## 11. 学完标准

- [ ] SDK 非流式 + 流式各写一次
- [ ] vLLM + base_url 调通
- [ ] litellm 接两个 provider
- [ ] 解释 tool_calls 两轮流程

---

## 12. 闭卷自测

1. `messages` 常见 role？
2. `finish_reason=length` 含义？
3. SDK 如何指向本地 vLLM？
4. 流式内容在哪个字段增量返回？
5. litellm `openai/...` 前缀含义？
6. Router `fallbacks` 何时触发？
7. `tool` role message 作用？
8. vLLM HTTP 对跨语言集成的好处？
9. litellm proxy vs 直连 OpenAI？
10. `tool_choice="auto"` 含义？

<details>
<summary>参考答案</summary>

1. system、user、assistant、tool（及可选 developer）。
2. 达到 max_tokens 被截断。
3. `OpenAI(base_url="http://localhost:8000/v1", api_key="EMPTY")`。
4. `chunk.choices[0].delta.content`。
5. 用 OpenAI 兼容协议访问非 OpenAI 后端。
6. 主模型超时/5xx/429 时切备用。
7. 承载函数执行结果供第二轮生成。
8. Java/Go 等任意语言 HTTP 调用，无需嵌入 Python。
9. proxy 集中多模型、鉴权、fallback；客户端只认一个 base_url。
10. 模型自行决定是否调用 tools。

</details>

---

## 13. 下一章

[39 Long Context 与稀疏注意力训练概念](39-Long-Context与稀疏注意力训练概念.md)

---

*推理：[20 vLLM](20-vLLM-TGI与LMDeploy-Python侧.md) · 网关：[AIAgent 03](../AIAgent/02-Go模型调用与结构化输出.md)*
