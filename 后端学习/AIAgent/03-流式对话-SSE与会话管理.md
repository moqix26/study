# 03 流式对话、SSE 与会话管理
> 适用环境：Go 1.26、Gin、Responses API。
>
> 本章重点：正确转发增量文本，并把会话状态安全地绑定到服务端身份。
>
> **与当前工程的边界**：`agentgo` V1 使用本地有界历史并向 Responses API 发送 `store:false`，不使用 `previous_response_id`。无状态推理续接会请求 `reasoning.encrypted_content`，并保留完整本地输入历史与全部 `response.output` item。本章后半的链尾、版本提交和乐观并发代码是进阶目标设计，不是当前 V1 已实现能力。
## 本章目标

学完本章，你应当能够：
1. 解释流式响应改善的是首字延迟，而不是模型总计算量。
2. 区分上游 Responses API SSE 与下游浏览器 SSE。
3. 用 Go 增量解析 SSE，而不是把整个响应读入内存。
4. 用 Gin 实现 POST /api/chat/stream。
5. 在浏览器断开后取消上游请求。
6. 用 conversation_id 安全管理 previous_response_id。
7. 正确处理同一会话的并发、失败、超时和状态提交。
## 1. 为什么需要流式响应

非流式调用的体验：
~~~text
发送请求 -> 等待模型生成完 -> 一次性看到完整结果
~~~
流式调用的体验：
~~~text
发送请求 -> 很快看到首段文字 -> 持续收到增量 -> 完成
~~~
流式响应主要降低“用户感知到的等待”。
它不意味着：
- 总 token 变少；
- 总费用必然降低；
- 模型生成速度必然更快；
- 中途收到的半句已经是最终答案；
- 断开后上游一定立刻停止计费。
你仍然需要总超时、取消传播、错误处理和用量监控。
## 2. SSE 协议基础

SSE 是基于 HTTP 的文本事件流，常见 Content-Type：
~~~http
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
~~~
一条事件可以写成：
~~~text
event: delta
data: {"text":"你"}
~~~
空行表示当前事件结束。
常见字段：
- event：事件名称；
- data：事件数据，可以出现多行；
- id：事件 ID；
- retry：客户端重连建议；
- 以冒号开头：注释，可用于心跳。
解析时不能只做 strings.Split(body, "\n\n")，因为：
- 网络分片不对应事件边界；
- data 可以有多行；
- 行结尾可能是 CRLF；
- 单个事件可能很大；
- 连接可能在任何位置中断。
## 3. 两段流，不是“一次透传”

系统中存在两套协议边界：
~~~text
浏览器
  |
  | POST /api/chat/stream
  | 你的 meta / delta / done / error 事件
  v
Gin 服务
  |
  | POST /v1/responses，stream=true
  | OpenAI response.* 事件
  v
模型供应商
~~~
建议不要把上游事件原封不动暴露给前端。
原因：
- 供应商事件可能变化；
- 兼容供应商事件名可能不同；
- 上游对象可能包含不应暴露的元数据；
- 前端只需要稳定的业务事件；
- 将来切换供应商时可保持前端协议不变。
本项目对前端定义四种事件：
| 事件 | data 示例 | 含义 |
|---|---|---|
| meta | {"conversation_id":"..."} | 建立或确认会话 |
| delta | {"text":"..." } | 用户可见文本增量 |
| done | {"response_id":"..."} | 当前 V1 收到上游 completed 后结束；conversation_id 已在 meta 中返回 |
| error | {"code":"UPSTREAM_ERROR"} | 流开始后的稳定错误 |
不要把 API Key、原始错误体或内部推理事件发给浏览器。
## 4. Responses API 的流式事件

请求增加：
~~~json
{
  "model": "YOUR_MODEL",
  "input": "用三点解释 Go channel。",
  "stream": true
}
~~~
你通常关心：
- response.created；
- response.output_text.delta；
- response.output_text.done；
- response.completed；
- error。
实际事件集合应以当前官方文档为准。
你的解析器应做到：
1. 根据 type 分支；
2. 只把 output_text.delta 的 delta 发给用户；
3. 从响应对象提取 response_id；
4. 收到 completed 后才提交会话状态；
5. 对未知事件忽略或记录指标；
6. 不把原始 Chain-of-Thought 或内部推理事件转发。
## 5. 会话状态的三种方式

| 方式 | 优点 | 代价 |
|---|---|---|
| 应用保存完整历史 | 跨供应商、可裁剪和审计 | 自己管理顺序、脱敏和 token |
| previous_response_id | 只保存当前链尾 | 依赖供应商状态和保留策略 |
| Conversations API | 适合持久、跨设备状态 | 仍要维护用户映射、权限与删除 |
previous_response_id 的下一轮请求：
~~~json
{
  "model": "YOUR_MODEL",
  "previous_response_id": "resp_xxx",
  "input": "再举一个例子。"
}
~~~
服务端必须把链尾绑定用户与会话；应用级 instructions、成本、上下文和供应商兼容仍需单独管理。客户端不能直接指定任意上游响应 ID。
当前 `agentgo` V1 实际使用：
~~~text
客户端 conversation_id
        |
        v
服务端 WindowStore（身份键 + conversation_id）
        |
        v
完整的本地消息窗口；上游请求 store=false
~~~
下面介绍的 `owner_id + previous_response_id + version` 是另一种可选的服务端状态方案。若采用它，必须显式核验供应商存储/保留策略；不能在 `store:false` 时只发送 `previous_response_id`。
## 6. conversation_id 的安全边界

本节描述生产目标。当前 V1 的 Header 只是演示身份；完成真实 JWT/Session/网关认证前，不能把它当成安全鉴权。

### 6.1 客户端请求
POST /api/chat 与 POST /api/chat/stream 统一接收：
~~~json
{
  "message": "继续解释刚才的例子",
  "conversation_id": "可选"
}
~~~
身份演示使用 Header：
~~~http
X-User-ID: user-123
~~~
生产环境中，X-User-ID 不能由公网客户端随意伪造。
它应由：
- 已验证 JWT 的中间件；
- API Gateway；
- 内部身份代理；
- 服务端 session；
解析后写入请求上下文。
### 6.2 服务端必须检查
如果 conversation_id 已存在：
1. 查询会话；
2. 验证 owner_id 等于当前身份；
3. 检查是否过期或删除；
4. 获取链尾 previous_response_id；
5. 对本轮并发进行控制。
如果 conversation_id 不存在：
- 生成高熵随机 ID；
- owner_id 使用当前已认证用户；
- 不接受客户端指定 owner_id；
- 初始 previous_response_id 为空。
### 6.3 不把上游响应 ID直接当公开会话 ID
公开上游 response_id 会造成：
- 供应商实现泄漏；
- 更难做用户绑定；
- 切换供应商困难；
- 客户端可能尝试枚举或串接不属于自己的响应。
公开的 conversation_id 应是你的资源标识。
## 7. Go 工程对应路径

当前统一工程：Go工程/agentgo
~~~text
Go工程/agentgo/
├─ cmd/server/main.go
├─ internal/llm/types.go
├─ internal/llm/openai_responses.go
├─ internal/memory/store.go
├─ internal/httpapi/server.go
└─ internal/httpapi/server_test.go
~~~
后文按职责拆出的 `stream.go`、`chat_stream.go` 等文件名是教学拆分建议，不代表磁盘上已经存在同名文件。
配置使用：
- AI_PROVIDER，默认 mock；
- AI_MODEL；
- OPENAI_API_KEY；
- OPENAI_BASE_URL；
- REQUEST_TIMEOUT；
- SERVER_ADDR。
## 8. 手把手实验一：解析上游 SSE

### 8.1 定义上游事件
internal/llm/stream.go：
~~~go
package llm
type streamEvent struct {
	Type     string          `json:"type"`
	Delta    string          `json:"delta"`
	Response json.RawMessage `json:"response"`
	Code     string          `json:"code"`
	Message  string          `json:"message"`
}
type responseMeta struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
~~~
这里只解析程序真正使用的字段。
### 8.2 编写通用 SSE 行解析
~~~go
func scanSSE(
	body io.Reader,
	handle func(eventName string, data []byte) error,
) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	var eventName string
	var dataLines []string
	dispatch := func() error {
		if len(dataLines) == 0 {
			eventName = ""
			return nil
		}
		data := []byte(strings.Join(dataLines, "\n"))
		if err := handle(eventName, data); err != nil {
			return err
		}
		eventName = ""
		dataLines = dataLines[:0]
		return nil
	}
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == "" {
			if err := dispatch(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		field, value, found := strings.Cut(line, ":")
		if found && strings.HasPrefix(value, " ") {
			value = value[1:]
		}
		switch field {
		case "event":
			eventName = value
		case "data":
			dataLines = append(dataLines, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取 SSE: %w", err)
	}
	return dispatch()
}
~~~
关键点：
- Scanner 默认 token 上限偏小，所以显式扩大；
- 设置硬上限，防止异常事件无限占内存；
- 多行 data 用换行拼回；
- 忽略注释心跳；
- EOF 前仍尝试派发最后一个完整数据事件。
### 8.3 调用 Responses API
~~~go
type StreamResult struct {
	ResponseID string
}
func (c *Client) StreamText(
	ctx context.Context,
	input string,
	previousResponseID string,
	onDelta func(string) error,
) (StreamResult, error) {
	payload := map[string]any{
		"model":        c.model,
		"instructions": "你是一名 Go 后端助教。不要输出内部思维过程。",
		"input":        input,
		"store":        true,
		"stream":       true,
	}
	if previousResponseID != "" {
		payload["previous_response_id"] = previousResponseID
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return StreamResult{}, fmt.Errorf("编码流请求: %w", err)
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/v1/responses",
		bytes.NewReader(raw),
	)
	if err != nil {
		return StreamResult{}, fmt.Errorf("创建流请求: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.http.Do(req)
	if err != nil {
		return StreamResult{}, fmt.Errorf("打开模型流: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512<<10))
		return StreamResult{}, fmt.Errorf(
			"模型状态=%d request_id=%q body=%s",
			resp.StatusCode,
			resp.Header.Get("x-request-id"),
			truncate(string(b), 500),
		)
	}
	var result StreamResult
	var completed bool
	err = scanSSE(resp.Body, func(_ string, data []byte) error {
		if bytes.Equal(bytes.TrimSpace(data), []byte("[DONE]")) {
			// 某些兼容实现使用 [DONE]，不能替代 completed 校验。
			return nil
		}
		var event streamEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("解析流事件: %w", err)
		}
		switch event.Type {
		case "response.created", "response.completed":
			var meta responseMeta
			if len(event.Response) > 0 {
				if err := json.Unmarshal(event.Response, &meta); err != nil {
					return fmt.Errorf("解析 response 元数据: %w", err)
				}
				if meta.ID != "" {
					result.ResponseID = meta.ID
				}
			}
			if event.Type == "response.completed" {
				completed = true
			}
		case "response.output_text.delta":
			if event.Delta != "" {
				if err := onDelta(event.Delta); err != nil {
					return err
				}
			}
		case "error":
			return fmt.Errorf(
				"模型流错误 code=%s message=%s",
				event.Code,
				event.Message,
			)
		default:
			// 未识别事件不向前端透传。
		}
		return nil
	})
	if err != nil {
		return StreamResult{}, err
	}
	if !completed || result.ResponseID == "" {
		return StreamResult{}, errors.New("模型流未完整结束")
	}
	return result, nil
}
~~~
## 9. 手把手实验二：内存会话存储

### 9.1 定义状态
internal/memory/store.go：
~~~go
package memory
type Conversation struct {
	ID                 string
	OwnerID            string
	PreviousResponseID string
	Version            uint64
	UpdatedAt          time.Time
}
var (
	ErrNotFound  = errors.New("conversation not found")
	ErrForbidden = errors.New("conversation forbidden")
	ErrConflict  = errors.New("conversation conflict")
)
type Store interface {
	Create(ctx context.Context, ownerID string) (Conversation, error)
	Get(ctx context.Context, ownerID, conversationID string) (Conversation, error)
	Commit(
		ctx context.Context,
		ownerID string,
		conversationID string,
		expectedVersion uint64,
		previousResponseID string,
	) (Conversation, error)
}
~~~
### 9.2 为什么 Commit 需要 expectedVersion
假设同一会话同时发送 A、B：
~~~text
版本 3 -> A 读取版本 3
版本 3 -> B 读取版本 3
A 完成 -> 写入 resp_A，版本 4
B 完成 -> 若无检查，会覆盖为 resp_B
~~~
这会丢失分支或让历史顺序不确定。
乐观并发控制：
- A 提交版本 3，成功变成 4；
- B 再提交期望版本 3，得到 ErrConflict；
- B 不应静默覆盖；
- 前端可提示“会话已有更新，请重试”。
另一种方案是为每个会话加锁并串行化整轮模型调用。
权衡：
| 方案 | 优点 | 缺点 |
|---|---|---|
| 整轮加锁 | 顺序直观 | 慢请求长期占锁 |
| 乐观版本 | 不长期占锁 | 冲突请求可能已经消耗模型费用 |
| 显式分支 | 支持并行探索 | 产品与数据模型更复杂 |
V1 可以选择整轮串行；生产系统应结合产品语义决定。
## 10. 手把手实验三：Gin 流式接口

internal/httpapi/chat_stream.go：
~~~go
type chatRequest struct {
	Message        string `json:"message" binding:"required"`
	ConversationID string `json:"conversation_id"`
}
type sseDelta struct {
	Text string `json:"text"`
}
func writeSSE(w gin.ResponseWriter, event string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return err
	}
	w.Flush()
	return nil
}
func (h *Handler) ChatStream(c *gin.Context) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_REQUEST"})
		return
	}
	userID := strings.TrimSpace(c.GetHeader("X-User-ID"))
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHENTICATED"})
		return
	}
	message := strings.TrimSpace(req.Message)
	if message == "" || utf8.RuneCountInString(message) > 4000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_MESSAGE"})
		return
	}
	ctx, cancel := context.WithTimeout(
		c.Request.Context(),
		h.requestTimeout,
	)
	defer cancel()
	conv, err := h.loadOrCreateConversation(
		ctx,
		userID,
		req.ConversationID,
	)
	if err != nil {
		h.writeConversationError(c, err)
		return
	}
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	if err := writeSSE(c.Writer, "meta", gin.H{
		"conversation_id": conv.ID,
	}); err != nil {
		return
	}
	result, err := h.llm.StreamText(
		ctx,
		message,
		conv.PreviousResponseID,
		func(delta string) error {
			return writeSSE(c.Writer, "delta", sseDelta{Text: delta})
		},
	)
	if err != nil {
		_ = writeSSE(c.Writer, "error", gin.H{
			"code": "UPSTREAM_ERROR",
		})
		return
	}
	updated, err := h.memory.Commit(
		ctx,
		userID,
		conv.ID,
		conv.Version,
		result.ResponseID,
	)
	if err != nil {
		_ = writeSSE(c.Writer, "error", gin.H{
			"code": "CONVERSATION_CONFLICT",
		})
		return
	}
	_ = writeSSE(c.Writer, "done", gin.H{
		"conversation_id": updated.ID,
	})
}
~~~
这段代码体现了一个重要事务边界：
> 只有上游 response.completed 且本地 Commit 成功，本轮才发送 done。
如果流中断：
- 不更新 previous_response_id；
- 不把半段答案当成完整会话轮次；
- 客户端可发起新请求；
- 日志记录中断原因。
## 11. 浏览器为何不用 EventSource 直接发请求

原生 EventSource 主要用于 GET，不适合携带本章的 POST JSON body。
使用 fetch + ReadableStream，生产解析器仍要支持 CRLF、多行 data 和半帧缓存：
~~~javascript
async function streamChat(message, conversationId) {
  const response = await fetch("/api/chat/stream", {
    method: "POST",
    headers: {"Content-Type": "application/json", "X-User-ID": "user-123"},
    body: JSON.stringify({message, conversation_id: conversationId || ""})
  });
  if (!response.ok || !response.body) throw new Error("request failed");
  const reader = response.body.getReader();
  // 循环 reader.read()，用 TextDecoder 累积并解析完整 SSE 帧。
}
~~~
展示 delta 必须使用 textContent；模型文本直接写 innerHTML 会产生 XSS 风险。
## 12. 取消、背压与代理缓冲

### 12.1 客户端断开
c.Request.Context() 会在断开时取消。把它持续传给模型、memory、数据库和工具，下游才能尽快停止；Handler 不要换成 context.Background()。
### 12.2 背压
客户端读取慢时 Write 会阻塞，连接和缓冲占用增加。设置总超时、单连接输出上限、并发限制；写失败立即取消并监控活跃流。
### 12.3 代理缓冲
本地流式、线上整段出现时，检查 Flush、text/event-stream、X-Accel-Buffering、路由压缩/缓存、网关 read timeout、CDN 长连接支持和注释心跳。
## 13. 会话生命周期与隐私

内存存储只适合单进程学习。生产环境要明确共享存储、TTL、删除与导出、账号注销、数据地域、脱敏、供应商保留策略、链尾失效降级和审计访问。不要无限保存聊天原文。
## 14. 测试流式接口

### 14.1 上游分片测试
httptest.Server 应故意把一个事件拆成多次 Write：
~~~go
fmt.Fprint(w, "event: response.output_text.delta\n")
fmt.Fprint(w, "data: {\"type\":\"response.output_text.delta\",")
w.(http.Flusher).Flush()
fmt.Fprint(w, "\"delta\":\"Go\"}\n\n")
w.(http.Flusher).Flush()
~~~
解析器不能依赖一次 Read 就得到完整 JSON。
### 14.2 取消测试
上游发送 delta 后取消 context；断言 StreamText 很快退出且 memory.Commit 未调用。
### 14.3 会话越权测试
user-A 创建会话，user-B 复用 ID；断言 403、ErrForbidden 且模型未调用。
### 14.4 并发冲突测试
两个请求读取同一 version；断言首个提交成功，第二个 ErrConflict，链尾不被覆盖。
若你把本章的进阶会话 Store 和 handler 真正实现并补齐对应测试，再运行：
~~~powershell
cd F:\study\后端学习\AIAgent\Go工程\agentgo
go test ./...
go run ./cmd/server
~~~
## 15. 常见故障表

| 现象 | 常见原因 | 处理 |
|---|---|---|
| 一直不出字 | 未 stream=true、代理缓冲、未 Flush | 检查上游与下游两段流 |
| 事件 JSON 偶发解析失败 | 把网络 Read 当事件边界 | 按 SSE 空行组帧 |
| 长回答报 Scanner too long | Scanner 默认上限 | 设置合理 Buffer 和硬上限 |
| 浏览器主动停止但上游继续 | 使用 Background 或未传 ctx | 全链路传播 Request.Context |
| 同会话回答顺序错乱 | 并发请求覆盖链尾 | 串行化、版本控制或显式分支 |
| 用户读到别人的历史 | conversation 未绑定 owner | 查询时强制校验身份 |
| 返回半段后仍记录成功 | 在 completed 前 Commit | 只在完整结束后提交 |
| 本地流式，线上整段返回 | 网关、压缩或 CDN 缓冲 | 调整路由配置与心跳 |
| 前端出现 XSS | 把 delta 写入 innerHTML | 使用 textContent |
| 兼容供应商只发 [DONE] | 事件语义不同 | 单独适配并做完成性测试 |
| previous_response_id 失效 | 上游保留或资源变化 | 返回可恢复错误或用本地历史重建 |
## 16. 练习

1. 为什么只让客户端提交 conversation_id，而不直接提交 previous_response_id？
2. 已输出三段 delta 后上游断开，是否更新链尾？
3. 每 15 秒一次的 SSE 注释心跳是什么格式？
4. 比较整轮加锁和乐观版本提交。
5. 为什么 delta 不能直接写 innerHTML？
6. 如何测试多行 data 和 CRLF？
## 17. 参考答案

1. conversation_id 是可绑定用户、TTL 和权限的业务资源；直接接收供应商 ID 会造成越权串接与迁移困难。
2. 不更新。只有 completed、有效 response_id 和本地 Commit 全部成功才推进。
3. 写“: ping”并跟一个空行。
4. 加锁顺序明确但长期占锁；乐观提交不占长锁，但冲突可能已产生模型费用。
5. 模型文本不可信，innerHTML 可能执行脚本；使用 textContent。
6. httptest.Server 输出多行 data、CRLF 并拆成多个 Write，断言最终正确组帧。
## 18. 学完标准

- [ ] 能画出浏览器、Gin、模型供应商之间的两段流。
- [ ] 能按 SSE 规则处理半帧、多行 data 与 CRLF。
- [ ] 能只转发用户可见 output_text.delta。
- [ ] 能让浏览器断开取消上游请求。
- [ ] 能安全绑定 user_id、conversation_id 与 previous_response_id。
- [ ] 能解释同会话并发的三种处理方式。
- [ ] 能保证未 completed 的流不推进会话状态。
- [ ] 能用 httptest 覆盖分片、取消、越权和冲突。
## 19. 下一章衔接

到这里，模型只能生成文本或结构化数据。
下一章让模型请求调用工具，但权限仍牢牢掌握在 Go 服务端：
- strict 工具参数；
- 工具白名单；
- 服务端身份注入；
- 超时、幂等和审计；
- 多轮 function_call 与 function_call_output；
- 防止提示注入把工具变成越权入口。
## 20. 官方参考

- [Streaming API responses](https://developers.openai.com/api/docs/guides/streaming-responses)
- [Conversation state](https://developers.openai.com/api/docs/guides/conversation-state)
- [迁移到 Responses API](https://developers.openai.com/api/docs/guides/migrate-to-responses)
- [Function calling](https://developers.openai.com/api/docs/guides/function-calling)
- [Structured Outputs](https://developers.openai.com/api/docs/guides/structured-outputs)
> SSE 事件类型和会话能力应以当前官方文档为准。兼容供应商必须分别验证事件名、结束语义、错误事件和状态延续能力。
