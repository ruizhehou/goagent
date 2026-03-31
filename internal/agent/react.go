package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"

	"goagent/internal/llm"
	"goagent/internal/tools"
	"goagent/pkg/config"
)

// AgentState Agent 状态
type AgentState string

const (
	StateIdle       AgentState = "idle"
	StateThinking   AgentState = "thinking"
	StateActing     AgentState = "acting"
	StateObserving  AgentState = "observing"
	StateFinished   AgentState = "finished"
	StateError      AgentState = "error"
)

// StepResult 单步执行结果
type StepResult struct {
	Thought     string `json:"thought"`
	Action      string `json:"action,omitempty"`
	ActionInput string `json:"action_input,omitempty"`
	Observation string `json:"observation,omitempty"`
	FinalAnswer string `json:"final_answer,omitempty"`
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Steps       []StepResult `json:"steps"`
	FinalAnswer string       `json:"final_answer"`
	Iterations  int          `json:"iterations"`
	Success     bool         `json:"success"`
	Error       string       `json:"error,omitempty"`
}

// ReActAgent ReAct Agent 实现
type ReActAgent struct {
	llm       *llm.LLM
	registry  *tools.ToolRegistry
	config    *config.AgentConfig
	verbose   bool
	callbacks []Callback
}

// Callback Agent 回调接口
type Callback interface {
	// OnStep 开始执行一步
	OnStep(ctx context.Context, step int, state AgentState)
	// OnThought 思考完成
	OnThought(ctx context.Context, thought string)
	// OnAction 执行动作
	OnAction(ctx context.Context, action, input string)
	// OnObservation 观察结果
	OnObservation(ctx context.Context, observation string)
	// OnFinish 执行完成
	OnFinish(ctx context.Context, result *ExecutionResult)
	// OnError 发生错误
	OnError(ctx context.Context, err error)
}

// NewReActAgent 创建 ReAct Agent
func NewReActAgent(llmClient *llm.LLM, registry *tools.ToolRegistry, cfg *config.AgentConfig) *ReActAgent {
	return &ReActAgent{
		llm:      llmClient,
		registry: registry,
		config:   cfg,
		verbose:  cfg.Verbose,
	}
}

// AddCallback 添加回调
func (a *ReActAgent) AddCallback(cb Callback) {
	a.callbacks = append(a.callbacks, cb)
}

// Run 执行 Agent
func (a *ReActAgent) Run(ctx context.Context, input string) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Steps:   []StepResult{},
		Success: false,
	}

	// 构建初始消息
	messages := a.buildInitialMessages(input)

	// 获取工具信息并创建带工具的模型实例
	toolInfos := a.registry.ToSchemaToolInfos()
	llmWithTools, err := a.llm.WithTools(toolInfos)
	if err != nil {
		return nil, fmt.Errorf("failed to bind tools: %w", err)
	}

	// ReAct 循环
	for i := 0; i < a.config.MaxIterations; i++ {
		a.notifyStep(ctx, i+1, StateThinking)

		// 调用 LLM
		response, err := llmWithTools.Generate(ctx, messages)
		if err != nil {
			a.notifyError(ctx, err)
			result.Error = err.Error()
			return result, nil
		}

		// 解析响应
		stepResult, finished, err := a.parseResponse(response)
		if err != nil {
			a.notifyError(ctx, err)
			result.Error = err.Error()
			return result, nil
		}

		result.Steps = append(result.Steps, *stepResult)
		result.Iterations = i + 1

		// 通知回调
		if stepResult.Thought != "" {
			a.notifyThought(ctx, stepResult.Thought)
		}

		// 检查是否完成
		if finished {
			result.FinalAnswer = stepResult.FinalAnswer
			result.Success = true
			a.notifyFinish(ctx, result)
			return result, nil
		}

		// 执行工具
		a.notifyAction(ctx, stepResult.Action, stepResult.ActionInput)

		observation, err := a.executeTool(ctx, stepResult.Action, stepResult.ActionInput)
		if err != nil {
			observation = fmt.Sprintf("Error: %v", err)
		}

		stepResult.Observation = observation
		a.notifyObservation(ctx, observation)

		// 将结果添加到消息历史
		messages = append(messages,
			response,
			schema.ToolMessage(observation, stepResult.Action),
		)
	}

	// 达到最大迭代次数
	result.Error = "max iterations reached"
	a.notifyFinish(ctx, result)
	return result, nil
}

// buildInitialMessages 构建初始消息
func (a *ReActAgent) buildInitialMessages(input string) []*schema.Message {
	messages := []*schema.Message{
		schema.SystemMessage(a.config.SystemPrompt),
		schema.UserMessage(input),
	}
	return messages
}

// parseResponse 解析 LLM 响应
func (a *ReActAgent) parseResponse(msg *schema.Message) (*StepResult, bool, error) {
	result := &StepResult{}

	// 检查是否有工具调用
	if len(msg.ToolCalls) > 0 {
		tc := msg.ToolCalls[0]
		result.Action = tc.Function.Name
		result.ActionInput = tc.Function.Arguments

		// 尝试从内容中提取思考过程
		if msg.Content != "" {
			result.Thought = a.extractThought(msg.Content)
		}

		return result, false, nil
	}

	// 没有工具调用，认为已完成
	result.FinalAnswer = msg.Content
	return result, true, nil
}

// extractThought 从内容中提取思考过程
func (a *ReActAgent) extractThought(content string) string {
	// 尝试提取 Thought: 后面的内容
	if idx := strings.Index(content, "Thought:"); idx != -1 {
		thought := content[idx+8:]
		if endIdx := strings.Index(thought, "\n"); endIdx != -1 {
			thought = thought[:endIdx]
		}
		return strings.TrimSpace(thought)
	}
	return content
}

// executeTool 执行工具
func (a *ReActAgent) executeTool(ctx context.Context, name, input string) (string, error) {
	return a.registry.Execute(ctx, name, input)
}

// Stream 流式执行 Agent
func (a *ReActAgent) Stream(ctx context.Context, input string) (<-chan StreamEvent, error) {
	eventChan := make(chan StreamEvent, 100)

	go func() {
		defer close(eventChan)

		result, err := a.Run(ctx, input)
		if err != nil {
			eventChan <- StreamEvent{Type: EventTypeError, Error: err}
			return
		}

		eventChan <- StreamEvent{Type: EventTypeFinish, Result: result}
	}()

	return eventChan, nil
}

// StreamEvent 流式事件
type StreamEvent struct {
	Type   EventType        `json:"type"`
	Step   *StepResult      `json:"step,omitempty"`
	Result *ExecutionResult `json:"result,omitempty"`
	Error  error            `json:"error,omitempty"`
}

// EventType 事件类型
type EventType string

const (
	EventTypeStep        EventType = "step"
	EventTypeThought     EventType = "thought"
	EventTypeAction      EventType = "action"
	EventTypeObservation EventType = "observation"
	EventTypeFinish      EventType = "finish"
	EventTypeError       EventType = "error"
)

// 回调通知方法

func (a *ReActAgent) notifyStep(ctx context.Context, step int, state AgentState) {
	for _, cb := range a.callbacks {
		cb.OnStep(ctx, step, state)
	}
}

func (a *ReActAgent) notifyThought(ctx context.Context, thought string) {
	for _, cb := range a.callbacks {
		cb.OnThought(ctx, thought)
	}
}

func (a *ReActAgent) notifyAction(ctx context.Context, action, input string) {
	for _, cb := range a.callbacks {
		cb.OnAction(ctx, action, input)
	}
}

func (a *ReActAgent) notifyObservation(ctx context.Context, observation string) {
	for _, cb := range a.callbacks {
		cb.OnObservation(ctx, observation)
	}
}

func (a *ReActAgent) notifyFinish(ctx context.Context, result *ExecutionResult) {
	for _, cb := range a.callbacks {
		cb.OnFinish(ctx, result)
	}
}

func (a *ReActAgent) notifyError(ctx context.Context, err error) {
	for _, cb := range a.callbacks {
		cb.OnError(ctx, err)
	}
}

// PrintCallback 打印回调实现
type PrintCallback struct{}

func (c *PrintCallback) OnStep(ctx context.Context, step int, state AgentState) {
	fmt.Printf("\n=== Step %d: %s ===\n", step, state)
}

func (c *PrintCallback) OnThought(ctx context.Context, thought string) {
	fmt.Printf("💭 Thought: %s\n", thought)
}

func (c *PrintCallback) OnAction(ctx context.Context, action, input string) {
	fmt.Printf("🎬 Action: %s(%s)\n", action, input)
}

func (c *PrintCallback) OnObservation(ctx context.Context, observation string) {
	fmt.Printf("👁 Observation: %s\n", observation)
}

func (c *PrintCallback) OnFinish(ctx context.Context, result *ExecutionResult) {
	fmt.Printf("\n✅ Finished: %s\n", result.FinalAnswer)
	fmt.Printf("Iterations: %d, Success: %v\n", result.Iterations, result.Success)
}

func (c *PrintCallback) OnError(ctx context.Context, err error) {
	fmt.Printf("❌ Error: %v\n", err)
}

// ToJSON 转换为 JSON
func (r *ExecutionResult) ToJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}
