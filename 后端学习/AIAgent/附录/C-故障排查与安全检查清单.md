# 附录 C：故障排查与安全检查清单

## 1. 排障总原则

先收集证据，再修改参数。

推荐顺序：

```text
请求入口
  → 认证与参数校验
  → Agent 编排
  → LLM provider
  → Tool / RAG
  → 数据库或外部服务
  → 响应组装 / SSE
```

每次只改变一个变量，并保留：

- 发生时间与时区；
- trace ID / request ID；
- 代码提交或构建版本；
- 配置摘要，密钥必须脱敏；
- 输入的安全摘要；
- provider、模型和 endpoint；
- 各阶段耗时；
- 重试次数与最终错误类型；
- 可复现步骤。

## 2. 第一轮快速检查

```powershell
go version
go env GOMOD
go mod verify
go test ./...
go run ./cmd/server
```

随后检查：

- [ ] `GET /health` 是否成功；
- [ ] 当前工作目录是否为 `Go工程/agentgo`；
- [ ] `AI_PROVIDER` 是否为预期值；
- [ ] 模型名是否来自当前环境；
- [ ] endpoint 是否多写或少写路径；
- [ ] timeout 是否过短；
- [ ] 系统时钟是否正确；
- [ ] 代理、DNS、证书和防火墙是否可用；
- [ ] 日志里是否有同一请求的完整阶段记录。

## 3. HTTP 与配置问题

| 症状 | 优先检查 | 不要立即做什么 |
|---|---|---|
| 401/403 | Key、认证头、账号权限、区域、资源授权 | 不要把完整 Key 打到日志 |
| 404 | base URL、API 路由、模型或 deployment 名称 | 不要无依据重复请求 |
| 408/504 | 总 deadline、DNS、连接、provider 延迟 | 不要无限增大 timeout |
| 429 | 限额、并发、token、重试提示 | 不要所有 goroutine 立即重试 |
| 400 | 请求 JSON、字段支持、schema、上下文长度 | 不要假设兼容端点支持全部字段 |
| 5xx | provider 状态、请求 ID、可重试分类 | 不要重放已执行的写工具 |
| 本地端口失败 | 监听地址、端口占用、容器映射 | 不要先关防火墙再说 |

配置日志只允许显示：

```text
provider=openai
model=<非敏感模型名>
base_url_host=<主机名>
timeout=30s
key_present=true
```

禁止显示：

- API Key；
- Authorization header；
- 数据库密码；
- Cookie、session token；
- 完整用户私密输入。

## 4. 流式与 SSE

### 有响应但客户端一直不显示

- [ ] handler 是否及时写入事件；
- [ ] 每个事件后是否 flush；
- [ ] 反向代理是否缓冲；
- [ ] `Content-Type` 是否为 `text/event-stream`；
- [ ] 客户端是否按事件边界解析，而不是按 TCP chunk；
- [ ] 是否把换行错误地当成消息结束；
- [ ] 首字节是否被中间层缓存。

### 流到一半中断

- [ ] 客户端是否取消 context；
- [ ] provider 是否发送错误事件；
- [ ] idle timeout 是否触发；
- [ ] 代理或负载均衡超时；
- [ ] goroutine 是否在 channel 写入时泄漏；
- [ ] 是否记录“正常结束”和“异常结束”的区别。

SSE 重连可能重复业务动作。

聊天流通常只允许重建输出，不应隐式重放有副作用的工具。

## 5. Tool Calling 与 Agent 循环

### 模型不调用工具

- 工具名和描述是否清楚；
- 参数 schema 是否过于复杂；
- 工具是否真的适用于问题；
- 当前模型和 API 是否通过工具能力探测；
- prompt 是否存在互相冲突的规则。

### 参数无法解析

- 每层是否设置 `additionalProperties: false`；
- 字段是否都列入 `required`；
- 时间、ID、枚举是否定义明确；
- 本地是否执行 schema 校验；
- 失败时是否把精简的校验错误交给下一轮，而非堆栈。

### 无限工具循环

- [ ] `MAX_AGENT_STEPS` 有硬上限；
- [ ] 对相同工具和相同参数做循环检测；
- [ ] context 有总 deadline；
- [ ] 工具失败被标记为失败；
- [ ] 模型不能伪造“工具已成功”；
- [ ] steps 仅记录可观察动作，不记录隐藏 Thought。

### 工具重复产生副作用

- 为写操作使用幂等键；
- 保存 operation 状态；
- 重试前判断结果未知还是确定失败；
- 把计划、确认、执行分开；
- 对资源执行服务端授权。

## 6. RAG 与 pgvector

### 完全召回不到

- [ ] 文档是否真的写入当前 store；
- [ ] namespace / tenant 是否一致；
- [ ] query 与 document 是否使用同一 embedding 模型；
- [ ] 向量维度是否一致；
- [ ] 距离函数和排序方向是否正确；
- [ ] filter 是否把结果全部过滤；
- [ ] `top_k` 是否为合法正数。

### 召回结果相关性差

- 先人工查看 top-k 文本，不要先改 prompt；
- 检查切块边界、标题和元数据；
- 建立带相关性标注的小评估集；
- 分开测 retrieval 与 generation；
- 比较关键词、向量和混合检索；
- 需要时加入 reranker，但要单独测成本和延迟。

### pgvector 报错

```sql
SELECT version();
SELECT extversion FROM pg_extension WHERE extname = 'vector';
```

继续核对：

- extension 是否已创建；
- 列维度与 embedding 维度；
- operator 与索引类型；
- migration 是否执行；
- `DATABASE_URL` 指向哪个实例；
- 连接池是否耗尽；
- 查询计划是否使用预期索引。

不要因为“有索引”就假定查询一定使用索引。

## 7. 本地推理

### Ollama

- 服务是否启动；
- 模型是否已拉取；
- 模型名与 tag 是否精确；
- 内存或显存是否足够；
- 首次加载是否被误认为永久延迟；
- 当前模型是否支持所需工具或 JSON 行为。

### vLLM

- GPU 驱动、CUDA 与镜像是否兼容；
- dtype/quantization 是否受硬件支持；
- chat template 是否存在且正确；
- 最大上下文是否挤占过多 KV cache；
- 并发与批处理设置是否导致排队；
- OOM 前后的日志与显存曲线；
- 兼容 API 是否支持项目使用的字段。

任何性能结论都记录：模型、硬件、精度、输入长度、输出长度、并发、预热和采样方法。

## 8. MCP 排障

- [ ] 初始化是否完成；
- [ ] 双方记录的协议版本是否一致；
- [ ] capability 是否经过协商；
- [ ] stdio server 是否把协议外日志写入 stdout；
- [ ] 子进程路径、工作目录和环境变量是否正确；
- [ ] Streamable HTTP session 标识是否正确处理；
- [ ] Origin 和认证是否校验；
- [ ] 工具列表是否发生变更；
- [ ] 参数与结果是否符合声明 schema；
- [ ] 取消、超时和进度事件是否被正确处理。

stdio 场景中，协议消息与调试日志混在 stdout 是常见故障。

调试日志应写 stderr，并避免敏感数据。

## 9. A2A 排障

- [ ] Agent Card 是否从预期来源获取；
- [ ] endpoint 是否可信且允许访问；
- [ ] 认证要求是否满足；
- [ ] skill/capability 是否仍存在；
- [ ] task ID 和 context 是否正确关联；
- [ ] message、artifact 和状态是否按当前规范解析；
- [ ] streaming 或异步更新是否能断线恢复；
- [ ] 远端返回内容是否按不可信输入处理。

不要根据远端 Agent 自报能力直接授予本地权限。

## 10. 安全基线

### 身份与授权

- [ ] 生产环境使用真实认证，不信任任意 `X-User-ID`；
- [ ] 每个工具做资源级授权；
- [ ] tenant 条件进入查询；
- [ ] 管理工具与普通工具分离；
- [ ] 默认拒绝未知工具。

### 输入与输出

- [ ] 限制请求体、prompt、文件和工具输出大小；
- [ ] 检索文本与工具结果视为不可信；
- [ ] 防止路径遍历、SSRF、SQL 注入和命令注入；
- [ ] URL 工具使用域名/IP allowlist；
- [ ] 输出经过 schema、引用和敏感信息校验；
- [ ] Markdown/HTML 在前端安全渲染。

### 密钥与数据

- [ ] 密钥来自 secret manager 或环境，不入库；
- [ ] 日志、trace、评估集已脱敏；
- [ ] 定义会话、prompt 和 provider 数据保留策略；
- [ ] 备份同样加密并控制访问；
- [ ] 支持密钥轮换和撤销。

### 资源与成本

- [ ] 请求、用户和工具维度限流；
- [ ] context 总超时；
- [ ] token、工具轮数和并发上限；
- [ ] 队列有界；
- [ ] 异常成本有告警和熔断；
- [ ] 大输入先估算或拒绝。

## 11. 故障复盘模板

```text
标题：
影响窗口：
用户影响：
检测方式：
直接原因：
促成因素：
为何监控未更早发现：
证据：
临时缓解：
永久修复：
防回归测试：
责任人与期限：
```

复盘不写“模型偶发抽风”作为根因。

要继续追问：哪个输入、哪个 provider 版本、哪个超时或哪个未覆盖分支，让随机性变成了用户可见故障。
