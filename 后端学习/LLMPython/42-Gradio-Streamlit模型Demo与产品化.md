# Gradio / Streamlit 模型 Demo 与产品化

> **文件编码**：UTF-8。  
> **前置**：[24 项目实战](24-项目实战微调小型语言模型.md)、[20 vLLM Python](20-vLLM-TGI与LMDeploy-Python侧.md)、[38 OpenAI API](38-OpenAI-API兼容层与litellm.md)。  
> **定位**：把 **微调模型 / vLLM / RAG** 包装成 **可演示、可部署** 的 Web Demo；衔接算法与产品，是 **24 章项目** 的标准交付物。

---

## 0. 读前导读

### 0.1 用一句话弄懂本章

**Gradio / Streamlit** = 用几十行 Python 做出 **聊天/文生图/上传文档** 界面，无需写 Vue/React；生产环境再换 FastAPI + 静态前端或 [38 章] OpenAI 兼容 API。

### 0.2 你需要提前知道什么

- 24 章 SFT 模型或 vLLM 服务
- Python 基础 async（Streamlit 同步为主）
- 可选：35 章 RAG 检索链

### 0.3 本章知识地图（☐→☑）

- [ ] 用 Gradio `ChatInterface` 包装本地 HF 模型
- [ ] 用 Gradio 调用 vLLM OpenAI 兼容端点
- [ ] Streamlit 多页：聊天 + 文件上传 RAG
- [ ] 理解 Demo vs 生产（鉴权、限流、日志）
- [ ] 完成 §闭卷自测 ≥8/10

### 0.4 建议学习时长

- **3～5 天**（与 24 章项目合并完成）

---

## 1. 这份文档学什么

- Gradio vs Streamlit vs FastAPI 选型
- Gradio 5：`ChatInterface`、`Blocks`、流式输出
- Streamlit：`st.chat_message`、session state
- 挂载 LoRA adapter、多模型切换
- Docker 部署 Demo（配合 [Linux 11](../Linux/11-Docker入门与镜像构建.md)）
- HuggingFace Spaces 零成本托管
- 与 [AIAgent](../AIAgent/00-学习路线图与说明.md) Java 产品层的分工

---

## 2. 三者对比

| 框架 | 适合 | 缺点 |
|------|------|------|
| **Gradio** | ML Demo、HF Spaces | 复杂业务 UI 弱 |
| **Streamlit** | 数据/实验仪表盘 | 定制布局有限 |
| **FastAPI** | 生产 API | 需自写前端 |
| **Java Spring** | 企业 CRUD+Agent | 非 DL 主线 |

**学习路线建议**：24 章用 **Gradio** 交作业 → 求职作品集 → 生产用 **vLLM + FastAPI**（Python/04 备选）。

---

## 3. Gradio 包装 HuggingFace 模型

```python
import gradio as gr
import torch
from transformers import AutoModelForCausalLM, AutoTokenizer

MODEL_ID = "Qwen/Qwen2.5-0.5B-Instruct"
tokenizer = AutoTokenizer.from_pretrained(MODEL_ID)
model = AutoModelForCausalLM.from_pretrained(
    MODEL_ID, torch_dtype=torch.bfloat16, device_map="auto"
)

def chat(message, history):
    messages = [{"role": "user", "content": message}]
    text = tokenizer.apply_chat_template(messages, tokenize=False, add_generation_prompt=True)
    inputs = tokenizer(text, return_tensors="pt").to(model.device)
    out = model.generate(**inputs, max_new_tokens=256, do_sample=True, temperature=0.7)
    reply = tokenizer.decode(out[0][inputs["input_ids"].shape[1]:], skip_special_tokens=True)
    return reply

demo = gr.ChatInterface(fn=chat, title="SFT Demo")
if __name__ == "__main__":
    demo.launch(server_name="0.0.0.0", server_port=7860)
```

```bash
pip install gradio
python app.py
# 浏览器 http://127.0.0.1:7860
```

---

## 4. Gradio 流式输出

```python
def chat_stream(message, history):
    # ... 构造 inputs ...
    streamer = TextIteratorStreamer(tokenizer, skip_special_tokens=True)
    # 在单独线程 model.generate(..., streamer=streamer)
    for chunk in streamer:
        yield chunk

gr.ChatInterface(fn=chat_stream).launch()
```

流式体验接近 ChatGPT，**面试 Demo 加分**。

---

## 5. 对接 vLLM OpenAI API

```python
from openai import OpenAI

client = OpenAI(base_url="http://127.0.0.1:8000/v1", api_key="EMPTY")

def chat_vllm(message, history):
    resp = client.chat.completions.create(
        model="your-model",
        messages=[{"role": "user", "content": message}],
        max_tokens=512,
    )
    return resp.choices[0].message.content
```

Gradio 前端不变，**后端从 HF 换 vLLM**，吞吐提升（见 20 章）。

---

## 6. Streamlit RAG 多页 Demo

```python
import streamlit as st

st.set_page_config(page_title="RAG Demo", layout="wide")
st.title("文档问答")

if "messages" not in st.session_state:
    st.session_state.messages = []

uploaded = st.file_uploader("上传 PDF/txt", type=["pdf", "txt"])
if uploaded:
    # 35 章：切分、embedding、写入 FAISS
    st.success(f"已索引 {uploaded.name}")

for msg in st.session_state.messages:
    with st.chat_message(msg["role"]):
        st.write(msg["content"])

if prompt := st.chat_input("提问"):
    st.session_state.messages.append({"role": "user", "content": prompt})
    # retrieve + llm.generate
    answer = "..."  # 你的 RAG 链
    st.session_state.messages.append({"role": "assistant", "content": answer})
    st.rerun()
```

---

## 7. 24 章项目交付 Checklist

```text
✅ train/ 脚本 + configs/
✅ wandb 链接或截图
✅ eval 报告（19 章 metrics）
✅ Gradio Demo（本章）+ 启动命令
✅ README：环境、数据、复现步骤
✅ （可选）HF Space 或 Docker 一键跑
```

---

## 8. Docker 化（简要）

```dockerfile
FROM python:3.10-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .
EXPOSE 7860
CMD ["python", "app.py"]
```

GPU 版需 `nvidia/cuda` 基础镜像 + `--gpus all`（见 Linux Docker 章）。

---

## 9. HuggingFace Spaces

| 类型 | 说明 |
|------|------|
| Gradio SDK | 推送 `app.py` 自动托管 |
| ZeroGPU | 免费 GPU 额度（见 HF 文档） |
| 私有 Space | 求职作品集链接写进简历 |

---

## 10. Demo → 生产差距

| 维度 | Demo | 生产 |
|------|------|------|
| 鉴权 | 无 | JWT / API Key |
| 限流 | 无 | 令牌桶 |
| 日志 | print | 结构化 + 追踪 |
| 模型 | 单进程 HF | vLLM / Triton |
| 前端 | Gradio | 自研或 B 端 |

**Infra 岗** 仍建议会做 Demo——证明 **端到端闭环**。

---

## 闭卷自测

1. Gradio 与 Streamlit 各适合什么场景？
2. `ChatInterface` 最少需要哪两个参数？
3. 如何让 Gradio 调用 vLLM？
4. Streamlit 刷新聊天列表常用什么 API？
5. 24 章项目 Demo 最低交付物？
6. 流式输出对用户体验的意义？
7. HF Spaces 对简历的价值？
8. Demo 与生产在鉴权上的典型差异？
9. RAG Demo 需要哪几个模块（35 章）？
10. 为何 Infra 学生也要会做 Demo？

<details>
<summary>参考答案</summary>

1. Gradio 偏 ML 交互 Demo；Streamlit 偏数据/dashboard 与快速原型。
2. `fn`（推理函数）和界面由 ChatInterface 封装；历史由框架管理。
3. OpenAI SDK 指 `base_url` 到 vLLM `/v1`，Gradio fn 内 `chat.completions.create`。
4. `st.chat_message` + `st.session_state` + `st.rerun()`。
5. 训练脚本、评估、Gradio、README 复现命令。
6. 降低首 token 等待感知，体验更流畅。
7. 可公开链接展示项目，无需面试官本地装环境。
8. Demo 常无鉴权；生产必须 API Key/OAuth 与 HTTPS。
9. 切分、embedding、向量库、检索、LLM 生成。
10. 证明能打通数据→训练→评估→展示全链路，沟通协作更顺畅。

</details>

---

## 下一章

本系列 **42 章完结**。复习：[25 面试总表](25-面试专题与知识点总表.md) · [00 路线图](00-学习路线图与说明.md)

实验代码：[examples/src/train/sft_lora.py](../examples/src/train/sft_lora.py) + 自建 `app.py`
