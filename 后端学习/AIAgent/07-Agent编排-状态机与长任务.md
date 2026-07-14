# 07｜Agent 编排：显式状态机、Tool Call 轨迹与长任务恢复

> 贯穿工程：<code>Go工程/agentgo</code>
> 对外路由：<code>POST /api/agent/run</code>，请求体为 <code>{"message":"..."}</code>
> 核心原则：模型提出动作，代码校验和执行动作，状态机决定任务何时继续或停止
## 1. 学习目标

学完本章，你应当能够：

- 判断一个需求是否真的需要 Agent；用显式状态机替代不可控的无限循环；把模型输出限制为 final answer 或结构化 tool call；只记录允许的 Tool Call 轨迹，不保存原始 Chain of Thought；
- 设置 max steps、截止时间、token、费用和工具调用预算；为有副作用工具设计幂等键；使用 checkpoint、租约和心跳恢复长任务；正确处理取消、重试、人工审批和失败补偿；
- 为状态转换与崩溃恢复编写确定性测试。
## 2. 什么时候需要 Agent

适合 Agent 的任务通常具有：

- 需要根据中间结果选择下一步；会调用一个或多个受控工具；步骤数量不是完全固定，但有清晰上限；任务可以被拆成可观察的状态转换；
- 错误能够分类和恢复。
不需要 Agent 的情况：

- 固定 CRUD；已知三步的稳定工作流；只做一次模型结构化输出；单次 RAG 问答；
- 支付、授权等必须严格确定的核心决策。
确定性工作流能完成时，优先用普通代码。 Agent 适合处理有限的不确定选择，不适合替代所有业务逻辑。
## 3. agentgo 目录

~~~text
Go工程/agentgo/
├─ cmd/server/
├─ internal/
│  ├─ agent/
│  │  ├─ model.go
│  │  ├─ state.go
│  │  ├─ runner.go
│  │  ├─ budget.go
│  │  ├─ checkpoint.go
│  │  └─ memory_store.go
│  ├─ tool/
│  │  ├─ registry.go
│  │  ├─ schema.go
│  │  └─ executor.go
│  ├─ llm/
│  │  └─ agent_model.go
│  └─ httpapi/
│     └─ agent_handler.go
└─ testdata/
   └─ agent/
~~~

<code>internal/agent</code> 负责状态转换和预算。 <code>internal/tool</code> 负责注册、参数校验、授权与执行。 <code>internal/llm</code> 负责把供应商响应转换成内部 Decision。
## 4. 不使用隐藏循环

危险写法通常类似：

~~~go
for {
	response := callModel(history)
	if response.Done {
		return response.Text
	}
	callTools(response.ToolCalls)
}
~~~

问题包括：

- 没有步数上限；没有总超时；失败后不知道停在哪；工具可能重复执行；
- 无法安全取消；无法解释资源花在哪里；进程重启后全部丢失。
即使 V1 在单进程同步执行，也要把每轮表示成显式状态转换。
## 5. 状态机

### 5.1 建议状态

~~~text
accepted
  -> running_model
  -> running_tool
  -> running_model
  -> succeeded
任意运行态
  -> paused_approval
  -> running_tool
任意非终态
  -> failed
  -> cancelled
  -> budget_exhausted
~~~

状态名称可以调整，但必须有明确含义和允许的转换。
### 5.2 Go 类型

~~~go
package agent
type RunStatus string
const (
	StatusAccepted        RunStatus = "accepted"
	StatusRunningModel    RunStatus = "running_model"
	StatusRunningTool     RunStatus = "running_tool"
	StatusPausedApproval  RunStatus = "paused_approval"
	StatusSucceeded       RunStatus = "succeeded"
	StatusFailed          RunStatus = "failed"
	StatusCancelled       RunStatus = "cancelled"
	StatusBudgetExhausted RunStatus = "budget_exhausted"
)
func (s RunStatus) Terminal() bool {
	switch s {
	case StatusSucceeded,
		StatusFailed,
		StatusCancelled,
		StatusBudgetExhausted:
		return true
	default:
		return false
	}
}
~~~

### 5.3 允许的转换

不要让任意状态直接改成任意状态。

~~~go
var allowed = map[RunStatus]map[RunStatus]bool{
	StatusAccepted: {
		StatusRunningModel: true,
		StatusCancelled:    true,
	},
	StatusRunningModel: {
		StatusRunningTool:     true,
		StatusSucceeded:       true,
		StatusFailed:          true,
		StatusBudgetExhausted: true,
		StatusCancelled:       true,
	},
	StatusRunningTool: {
		StatusRunningModel:    true,
		StatusPausedApproval:  true,
		StatusFailed:          true,
		StatusBudgetExhausted: true,
		StatusCancelled:       true,
	},
	StatusPausedApproval: {
		StatusRunningTool: true,
		StatusCancelled:   true,
		StatusFailed:      true,
	},
}
~~~

持久化状态时应使用 compare-and-set：

~~~text
UPDATE agent_runs
SET status = next_status, version = version + 1
WHERE id = run_id
  AND status = expected_status
  AND version = expected_version
~~~

受影响行数为零说明状态已被其他执行器修改。 此时不能强行覆盖。
## 6. 模型只返回 Decision

模型 adapter 把供应商响应转换为内部结构：

~~~go
package agent
type DecisionKind string
const (
	DecisionFinal DecisionKind = "final"
	DecisionTools DecisionKind = "tools"
)
type ToolCall struct {
	CallID    string
	Name      string
	Arguments map[string]any
}
type Decision struct {
	Kind      DecisionKind
	FinalText string
	ToolCalls []ToolCall
	Usage     Usage
}
~~~

Decision 中没有 Thought、Reasoning 或 ChainOfThought 字段。 系统不要求模型输出逐步思考，也不保存供应商可能返回的隐藏推理。 代码真正需要的是：

- 是否结束；要调用什么工具；参数是什么；本轮 token 用量；
- 响应标识等恢复所需元数据。
## 7. 不记录原始 Chain of Thought

原始思维链不适合作为运行轨迹：

- 可能包含敏感上下文；可能泄露系统提示和内部策略；它不是稳定、可验证的程序状态；文字看似合理，不代表动作正确；
- 长期保存增加隐私与合规风险。
agentgo 的 step 只记录：

1. tool 名称；
2. 经过 schema 校验和授权后的参数；
3. 结果摘要；
4. 状态；
5. 持续时间。
不记录：

- 模型原始 CoT；隐藏 reasoning token 内容；完整系统 Prompt；默认完整工具响应；
- 密钥、访问令牌和敏感字段。
### 7.1 Step 模型

~~~go
package agent
import "time"
type StepStatus string
const (
	StepStarted   StepStatus = "started"
	StepSucceeded StepStatus = "succeeded"
	StepFailed    StepStatus = "failed"
	StepSkipped   StepStatus = "skipped"
)
type Step struct {
	ToolName      string
	ValidatedArgs map[string]any
	ResultSummary string
	Status        StepStatus
	Duration      time.Duration
}
~~~

这五项就是允许进入 Step 轨迹的业务字段。run_id、ordinal 和持久化时间可以作为存储记录的外层索引与审计元数据，但不能把模型原始内容混入 Step。ValidatedArgs 仍需脱敏，例如发送邮件工具可以记录模板 ID 与收件人域名，但未必需要记录完整正文和完整地址。
## 8. Tool Call 执行边界

模型提出工具调用，不代表工具一定执行。 执行前顺序如下：

1. 工具名必须存在于 registry；
2. JSON 参数通过严格 schema；
3. 服务端注入 UserID、TenantID 等可信身份；
4. 检查调用者权限；
5. 检查运行预算；
6. 判断是否需要人工审批；
7. 生成或读取幂等键；
8. 设置工具级超时；
9. 执行；
10. 对返回值做大小限制、脱敏和摘要；
11. 持久化 Step；
12. 只把必要结果交回模型。
模型不能通过参数覆盖 TenantID。 即使参数 schema 中出现 tenant 字段，也应由服务端固定或直接禁止。
## 9. Runner 接口

~~~go
package agent
import "context"
type Model interface {
	Next(ctx context.Context, in ModelInput) (Decision, error)
}
type ToolExecutor interface {
	Execute(ctx context.Context, call AuthorizedCall) (ToolResult, error)
}
type Store interface {
	LoadRun(ctx context.Context, runID string) (Run, error)
	SaveCheckpoint(ctx context.Context, cp Checkpoint) error
	AppendStep(ctx context.Context, runID string, ordinal int, step Step) error
	Complete(ctx context.Context, result Completion) error
}
type Runner struct {
	model Model
	tools ToolExecutor
	store Store
	clock Clock
}
~~~

V1 可以使用 memory store。 V3 再用 PostgreSQL 实现 Store 和租约。
## 10. 单轮状态转换

伪代码：

~~~go
func (r *Runner) Advance(ctx context.Context, runID string) error {
	run, err := r.store.LoadRun(ctx, runID)
	if err != nil {
		return err
	}
	if run.Status.Terminal() {
		return nil
	}
	if err := run.Budget.CheckBeforeStep(r.clock.Now()); err != nil {
		return r.finishBudgetExceeded(ctx, run, err)
	}
	decision, err := r.model.Next(ctx, modelInputFrom(run))
	if err != nil {
		return r.handleModelError(ctx, run, err)
	}
	run.Budget.AddUsage(decision.Usage)
	switch decision.Kind {
	case DecisionFinal:
		return r.complete(ctx, run, decision.FinalText)
	case DecisionTools:
		return r.executeCalls(ctx, run, decision.ToolCalls)
	default:
		return r.fail(ctx, run, "invalid model decision")
	}
}
~~~

一个 Advance 是否执行一个还是多个 tool call，需要项目明确。 无论哪种策略，每个持久化边界都必须可恢复。
## 11. 预算不是只有 max steps

### 11.1 Budget

~~~go
package agent
import "time"
type Budget struct {
	MaxSteps       int
	MaxToolCalls   int
	MaxInputTokens int
	MaxOutputTokens int
	MaxCostMicros  int64
	Deadline       time.Time
	UsedSteps       int
	UsedToolCalls   int
	UsedInputTokens int
	UsedOutputTokens int
	UsedCostMicros  int64
}
~~~

每次模型调用前和工具调用前都检查预算。 只在循环顶部检查，会允许一轮中的多个并行工具突破预算。
### 11.2 常见预算

- max steps：防止无限决策；max tool calls：限制工具扇出；deadline：限制墙钟时间；input/output tokens：限制上下文增长；
- cost：限制模型和付费工具费用；per-tool timeout：阻止单工具卡死；max result bytes：防止工具结果塞爆上下文；max parallelism：限制并行资源。
达到预算是预期终态，不应伪装成内部 500。 响应可标记 <code>budget_exhausted</code> 并给出安全摘要。
## 12. 上下文增长控制

如果每一步都把完整历史和完整工具结果再次发送，token 会快速增长。 可采用：

- 只保留必要对话消息；工具结果保存到对象存储或数据库，模型只读摘要和引用；对旧步骤生成受控摘要；保留关键事实和未完成任务；
- 限制每个结果的最大字符或 token；不把二进制、HTML 页面和巨型 JSON 直接塞回模型。
摘要可能丢信息，所以关键业务字段必须结构化保存，不能只存在摘要文本里。
## 13. 幂等：有副作用工具的生命线

读工具重复执行通常只浪费资源。 写工具重复执行可能导致：

- 重复发邮件；重复创建工单；重复扣费；重复发布内容；
- 重复修改外部系统。
### 13.1 幂等键

一种内部幂等键：

~~~text
idempotency_key = run_id + step_ordinal + tool_call_id
~~~

如果供应商支持 idempotency key，应传递稳定值。 如果不支持，项目需保存执行记录并在重试前查询。
### 13.2 不能承诺神奇的 exactly once

跨网络系统中，调用方超时并不知道对方是否已经成功。 没有对方幂等支持或可查询结果时，无法仅靠本地事务保证 exactly once。 正确表达是：

- at-least-once 执行；幂等效果；可查询操作状态；必要时人工消歧；
- 对可逆动作提供补偿。
### 13.3 工具执行记录

~~~text
tool_executions
  idempotency_key UNIQUE
  run_id
  step_ordinal
  tool_name
  args_hash
  status
  external_operation_id
  result_summary
  started_at
  completed_at
~~~

参数哈希应基于规范化、校验后的参数。 同一个幂等键出现不同参数时必须报警并拒绝。
## 14. 长任务与 Checkpoint

长任务不能只保存在 goroutine 内存中。 进程部署、崩溃或扩容都会丢失它。 Checkpoint 至少包含：

~~~go
package agent
type Checkpoint struct {
	RunID            string
	Version          int64
	Status           RunStatus
	NextStepOrdinal  int
	ModelCursor      string
	Budget           Budget
	WorkingState     map[string]any
	LeaseOwner       string
	LeaseUntilUnixMs int64
}
~~~

WorkingState 只保存恢复所需的结构化事实。 不要把模型原始 CoT 当 checkpoint。
### 14.1 保存时机

推荐在以下边界保存：

- 运行创建后；模型 Decision 被解析和验证后；工具调用开始前；工具结果落库后；
- 预算更新后；等待人工审批前；每个终态。
保存顺序应避免出现“副作用已经发生，但没有任何执行记录”。
## 15. 租约、心跳与恢复

### 15.1 为什么需要租约

多个 Worker 可能同时看见同一个可运行任务。 租约表示某个 Worker 在有限时间内拥有执行权。 字段：

~~~text
lease_owner
lease_until
heartbeat_at
run_version
~~~

抢占使用短事务。 远程模型和工具调用不能放在持有数据库锁的长事务中。
### 15.2 心跳

对于长工具调用，Worker 周期性延长租约。 心跳失败时不要立刻启动第二份有副作用调用。 应结合：

- 当前外部 operation ID；idempotency key；工具可查询状态；租约宽限期；
- 人工介入策略。
### 15.3 租约过期

新 Worker 恢复时：

1. 读取最新 checkpoint 与版本；
2. 检查上一步 execution 状态；
3. 如果有 external operation ID，先查询远端；
4. 已成功则补写结果；
5. 未开始或明确失败才重试；
6. 状态不确定且工具不可幂等时暂停人工处理；
7. compare-and-set 保存下一状态。
恢复不是简单“从头再跑”。
## 16. 人工审批

高风险工具在执行前进入 <code>paused_approval</code>：

- 发起付款；删除数据；对外发布；批量发送消息；
- 修改权限；高额付费调用。
审批记录应包含：

- run 与 step；工具名称；脱敏后的校验参数；风险摘要；
- 请求人与审批人；到期时间；批准或拒绝；审计时间。
审批后仍需重新检查权限、预算和参数版本。 不能假设等待期间环境没有变化。
## 17. 并行 Tool Call

模型可能一次提出多个工具。 只有满足以下条件才适合并行：

- 工具之间无依赖；预算允许；权限均已校验；并发上限明确；
- 结果顺序可稳定恢复；一个失败时的策略明确。
如果 B 依赖 A 的结果，则必须串行。 不要为了展示并发而破坏业务顺序。 使用 <code>errgroup</code> 时仍应：

- 给每个工具独立 timeout；限制并发；给结果按 CallID 排序；分别记录 Step；
- 明确部分成功如何处理。
## 18. 错误分类

| 类别 | 示例 | 默认动作 |
|---|---|---|
| validation | 参数不符合 schema | 不执行，反馈模型或失败 |
| authorization | 用户无权限 | 立即失败并审计 |
| transient | 429、临时 5xx | 有限重试 |
| permanent | 工具不存在、账号禁用 | 失败 |
| ambiguous | 超时但外部状态未知 | 查询状态或人工处理 |
| budget | 达到步数、费用、截止时间 | budget_exhausted |
| cancelled | 用户取消 | 停止新动作并收尾 |
| conflict | checkpoint 版本冲突 | 重新加载，不覆盖 |
模型不应看到内部堆栈、连接串或密钥。 它只需要安全、可操作的错误摘要。
## 19. 取消

取消不是只改一个布尔值。 Runner 在这些位置检查：

- 调模型前；调工具前；重试前；checkpoint 后；
- 人工审批恢复时。
已经发出的外部副作用未必能取消。 此时记录 external operation ID，并按工具能力取消、等待或补偿。 终态后不能再追加普通步骤。 审计和补偿事件应走独立记录。
## 20. POST /api/agent/run

V1 可以同步运行一个很短的 mock Agent：

~~~json
{
  "message": "查询北京天气并给出穿衣建议"
}
~~~

响应至少包含：

~~~json
{
  "run_id": "run_123",
  "status": "succeeded",
  "answer": "今天有雨，建议携带雨具。",
  "steps": [
    {
      "tool": "get_weather",
      "status": "succeeded",
      "result_summary": "北京，小雨，18°C",
      "duration_ms": 12
    }
  ]
}
~~~

不要在响应 steps 中加入 thought 或 reasoning 字段。 生产长任务更适合创建 run 后返回 202，再提供状态查询和事件流；该扩展可放入 V3。
## 21. 测试

### 21.1 状态转换测试

- accepted 只能进入允许状态；succeeded 后不能继续执行；非法转换返回错误；compare-and-set 冲突不覆盖新状态；
- budget_exhausted 是终态。
### 21.2 预算测试

- max steps 正好用完；工具调用前已无预算；deadline 已过；并行调用总数不越界；
- token 与费用在每轮后累计；超预算不再调用模型。
### 21.3 幂等测试

- 相同幂等键相同参数只执行一次；相同键不同参数被拒绝；工具成功后进程崩溃，恢复不重复副作用；ambiguous 状态进入查询或人工路径。
### 21.4 租约测试

- 未过期租约不能被第二 Worker 抢占；过期租约可恢复；旧 Worker 不能用旧 version 覆盖新 checkpoint；心跳能延长租约；
- 取消状态不会被恢复 Worker 重新运行。
### 21.5 不保存 CoT

构造模型 adapter 返回包含供应商内部 reasoning 的测试数据。 断言 Store、日志、HTTP 响应和 Step 中都没有保存该内容。
## 22. 常见故障表

| 现象 | 原因 | 诊断 | 修复 |
|---|---|---|---|
| Agent 一直循环 | 无 max steps 或终态 | run 状态与预算 | 显式终态 |
| 同一邮件发两次 | 重试无幂等 | execution 记录 | 稳定幂等键 |
| 重启后任务消失 | 只存在 goroutine | 存储与部署日志 | 持久 checkpoint |
| 两个 Worker 同跑 | 无租约或 CAS | lease 与 version | 短租约 + CAS |
| 恢复后从头执行 | checkpoint 太少 | step 和 operation ID | 边界持久化 |
| 日志泄露思维链 | 直接保存模型响应 | trace payload | 白名单字段 |
| 工具越权 | 模型参数带身份 | 授权日志 | 服务端注入身份 |
| 成本失控 | 只有步数限制 | usage 与工具计费 | 多维预算 |
| 工具结果塞爆上下文 | 保存完整大响应 | token 轨迹 | 摘要与外部引用 |
| 取消后仍执行 | 只在入口检查 | step 时间线 | 每个边界检查 |
| 长事务占满连接 | 锁内调模型或工具 | DB 活跃事务 | 租约短事务 |
| 超时后状态不明 | 外部动作无查询能力 | operation ID | 幂等或人工消歧 |
## 23. 练习

### 练习 1

为“查询库存，库存足够时创建预订单”画出状态转换，并指出哪个工具有副作用。
### 练习 2

为什么只设置 max steps 仍不足以控制 Agent 成本？
### 练习 3

发送邮件超时，供应商没有返回成功，但可能已经发送。 系统下一步应该做什么？
### 练习 4

两个 Worker 同时加载 version=7 的 run。 怎样保证只有一个能保存 version=8？
### 练习 5

列出允许写入 Agent Step 的字段，并说明为何不保存原始 CoT。
## 24. 参考答案

### 答案 1

~~~text
accepted
 -> running_model
 -> running_tool: check_inventory
 -> running_model
 -> paused_approval 或 running_tool: create_reservation
 -> running_model
 -> succeeded
~~~

check_inventory 是读操作。 create_reservation 会改变外部状态，必须使用幂等键，必要时审批。
### 答案 2

单步可能包含巨大 Prompt、昂贵模型、多个并行付费工具或长时间等待。 还需限制 token、费用、工具数、deadline、结果大小和并发。
### 答案 3

不能盲目重发。 先用幂等键或 external operation ID 查询供应商状态。 若不可查询且无法保证幂等，应标记 ambiguous 并人工处理。
### 答案 4

保存语句同时匹配 run ID 和 version=7。 第一个更新成功并把 version 改为 8，第二个受影响行数为零，必须重新加载。
### 答案 5

只记录工具名、校验和脱敏后的参数、结果摘要、状态和持续时间。 原始 CoT 不是可靠程序状态，且可能泄露敏感上下文、系统策略或内部提示。
## 25. 学完标准

- [ ] 能判断 Agent 与固定工作流的边界；
- [ ] 能画出显式状态机和终态；
- [ ] Decision 中没有 Thought 字段；
- [ ] Step 仅含工具、校验参数、结果摘要、状态和耗时；
- [ ] 每轮前后检查多维预算；
- [ ] 有副作用工具使用幂等键；
- [ ] 能解释跨系统 exactly once 的限制；
- [ ] 长任务有 checkpoint、租约、心跳和 CAS；
- [ ] 能恢复崩溃后的明确与不明确状态；
- [ ] 能实现取消、审批和并行限制；
- [ ] 测试能证明不保存原始 CoT。
## 26. 与下一章衔接

状态机让 Agent 可以运行，但不能证明它运行得好。 下一章建立评估、可观测、安全与成本体系：

- retrieval、answer、tool、task 四层指标；LLM judge 的能力边界；固定人工评估集；trace 与隐私最小化；
- token 和工具成本公式；限流、熔断与降级放在哪一层；Prompt Injection、越权、数据泄露与供应链风险。
