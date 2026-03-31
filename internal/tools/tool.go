package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/schema"
)

// Tool 工具接口
type Tool interface {
	// Info 返回工具信息
	Info() *ToolInfo

	// Execute 执行工具
	Execute(ctx context.Context, input string) (string, error)
}

// ToolInfo 工具信息
type ToolInfo struct {
	// Name 工具名称
	Name string `json:"name"`

	// Description 工具描述
	Description string `json:"description"`

	// Parameters 参数定义
	Parameters map[string]*schema.ParameterInfo
}

// ToolRegistry 工具注册表
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry 创建工具注册表
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具
func (r *ToolRegistry) Register(tool Tool) {
	r.tools[tool.Info().Name] = tool
}

// Get 获取工具
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// List 列出所有工具
func (r *ToolRegistry) List() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ToSchemaToolInfos 转换为 Eino 框架的 ToolInfo 格式
func (r *ToolRegistry) ToSchemaToolInfos() []*schema.ToolInfo {
	var result []*schema.ToolInfo
	for _, tool := range r.tools {
		info := tool.Info()
		toolInfo := &schema.ToolInfo{
			Name:  info.Name,
			Desc:  info.Description,
		}
		if len(info.Parameters) > 0 {
			toolInfo.ParamsOneOf = schema.NewParamsOneOfByParams(info.Parameters)
		}
		result = append(result, toolInfo)
	}
	return result
}

// Execute 执行工具
func (r *ToolRegistry) Execute(ctx context.Context, name string, input string) (string, error) {
	tool, ok := r.Get(name)
	if !ok {
		return "", fmt.Errorf("tool not found: %s", name)
	}
	return tool.Execute(ctx, input)
}

// BaseTool 基础工具实现
type BaseTool struct {
	name        string
	description string
	parameters  map[string]*schema.ParameterInfo
	execute     func(ctx context.Context, input string) (string, error)
}

// Info 返回工具信息
func (t *BaseTool) Info() *ToolInfo {
	return &ToolInfo{
		Name:        t.name,
		Description: t.description,
		Parameters:  t.parameters,
	}
}

// Execute 执行工具
func (t *BaseTool) Execute(ctx context.Context, input string) (string, error) {
	if t.execute == nil {
		return "", fmt.Errorf("execute function not implemented")
	}
	return t.execute(ctx, input)
}

// NewBaseTool 创建基础工具
func NewBaseTool(name, description string, parameters map[string]*schema.ParameterInfo, execute func(ctx context.Context, input string) (string, error)) *BaseTool {
	return &BaseTool{
		name:        name,
		description: description,
		parameters:  parameters,
		execute:     execute,
	}
}

// ParseInput 解析工具输入
func ParseInput[T any](input string) (T, error) {
	var result T
	if err := json.Unmarshal([]byte(input), &result); err != nil {
		return result, fmt.Errorf("failed to parse input: %w", err)
	}
	return result, nil
}
