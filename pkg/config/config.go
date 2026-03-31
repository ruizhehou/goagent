package config

import (
	"os"
)

// Config 应用配置
type Config struct {
	// LLM 配置
	LLM LLMConfig `yaml:"llm"`

	// Agent 配置
	Agent AgentConfig `yaml:"agent"`
}

// LLMConfig 大模型配置
type LLMConfig struct {
	// Provider 模型提供商: openai, ark (火山引擎)
	Provider string `yaml:"provider"`

	// APIKey API 密钥
	APIKey string `yaml:"api_key"`

	// BaseURL API 基础 URL (可选)
	BaseURL string `yaml:"base_url"`

	// Model 模型名称
	Model string `yaml:"model"`

	// Temperature 温度参数
	Temperature float64 `yaml:"temperature"`

	// MaxTokens 最大 token 数
	MaxTokens int `yaml:"max_tokens"`
}

// AgentConfig Agent 配置
type AgentConfig struct {
	// MaxIterations 最大迭代次数
	MaxIterations int `yaml:"max_iterations"`

	// Verbose 是否输出详细日志
	Verbose bool `yaml:"verbose"`

	// SystemPrompt 系统提示词
	SystemPrompt string `yaml:"system_prompt"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		LLM: LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o-mini",
			Temperature: 0.7,
			MaxTokens:   4096,
		},
		Agent: AgentConfig{
			MaxIterations: 10,
			Verbose:       true,
			SystemPrompt: `你是一个智能助手，具备以下能力：
1. 通过工具执行各种任务
2. 分析问题并制定解决方案
3. 验证执行结果

请根据用户的需求，选择合适的工具执行任务，并在执行完成后验证结果。

输出格式：
- Thought: 思考过程
- Action: 要执行的动作
- Action Input: 动作的输入参数
- Observation: 观察到的结果
- Final Answer: 最终答案`,
		},
	}
}

// LoadFromEnv 从环境变量加载配置
func LoadFromEnv(cfg *Config) {
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		cfg.LLM.APIKey = apiKey
	}
	if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
		cfg.LLM.BaseURL = baseURL
	}
	if model := os.Getenv("LLM_MODEL"); model != "" {
		cfg.LLM.Model = model
	}
}
