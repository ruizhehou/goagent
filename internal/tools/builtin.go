package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
)

// CalculatorTool 计算器工具
type CalculatorTool struct {
	*BaseTool
}

// NewCalculatorTool 创建计算器工具
func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{
		BaseTool: NewBaseTool(
			"calculator",
			"执行数学计算，支持加减乘除运算",
			map[string]*schema.ParameterInfo{
				"expression": {
					Type:     schema.String,
					Desc:     "数学表达式，例如: 2+3*4",
					Required: true,
				},
			},
			func(ctx context.Context, input string) (string, error) {
				type CalcInput struct {
					Expression string `json:"expression"`
				}
				parsed, err := ParseInput[CalcInput](input)
				if err != nil {
					return "", err
				}
				// 简单实现，实际应用中应该使用更安全的表达式解析
				// 这里只做演示
				return fmt.Sprintf("计算结果: %s = [需要实现表达式解析]", parsed.Expression), nil
			},
		),
	}
}

// WeatherTool 天气查询工具
type WeatherTool struct {
	*BaseTool
}

// NewWeatherTool 创建天气查询工具
func NewWeatherTool() *WeatherTool {
	return &WeatherTool{
		BaseTool: NewBaseTool(
			"get_weather",
			"获取指定城市的天气信息",
			map[string]*schema.ParameterInfo{
				"city": {
					Type:     schema.String,
					Desc:     "城市名称",
					Required: true,
				},
			},
			func(ctx context.Context, input string) (string, error) {
				type WeatherInput struct {
					City string `json:"city"`
				}
				parsed, err := ParseInput[WeatherInput](input)
				if err != nil {
					return "", err
				}
				// 模拟天气查询
				return fmt.Sprintf("%s 的天气: 晴，温度 25°C，空气质量良好", parsed.City), nil
			},
		),
	}
}

// TimeTool 时间查询工具
type TimeTool struct {
	*BaseTool
}

// NewTimeTool 创建时间查询工具
func NewTimeTool() *TimeTool {
	return &TimeTool{
		BaseTool: NewBaseTool(
			"get_current_time",
			"获取当前时间",
			map[string]*schema.ParameterInfo{}, // 无参数
			func(ctx context.Context, input string) (string, error) {
				now := time.Now()
				return fmt.Sprintf("当前时间: %s", now.Format("2006-01-02 15:04:05")), nil
			},
		),
	}
}

// SearchTool 搜索工具
type SearchTool struct {
	*BaseTool
}

// NewSearchTool 创建搜索工具
func NewSearchTool() *SearchTool {
	return &SearchTool{
		BaseTool: NewBaseTool(
			"search",
			"搜索互联网信息",
			map[string]*schema.ParameterInfo{
				"query": {
					Type:     schema.String,
					Desc:     "搜索关键词",
					Required: true,
				},
			},
			func(ctx context.Context, input string) (string, error) {
				type SearchInput struct {
					Query string `json:"query"`
				}
				parsed, err := ParseInput[SearchInput](input)
				if err != nil {
					return "", err
				}
				// 模拟搜索结果
				return fmt.Sprintf("搜索 '%s' 的结果:\n1. 相关文章1\n2. 相关文章2\n3. 相关文章3", parsed.Query), nil
			},
		),
	}
}

// DefaultTools 返回默认工具列表
func DefaultTools() []Tool {
	return []Tool{
		NewCalculatorTool(),
		NewWeatherTool(),
		NewTimeTool(),
		NewSearchTool(),
	}
}
