package llm

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"goagent/pkg/config"
)

// Provider LLM 提供商
type Provider string

const (
	ProviderOpenAI Provider = "openai"
	ProviderArk    Provider = "ark" // 火山引擎
)

// LLM 大模型客户端
type LLM struct {
	provider Provider
	client   model.ToolCallingChatModel
	config   *config.LLMConfig
}

// NewLLM 创建 LLM 客户端
func NewLLM(cfg *config.LLMConfig) (*LLM, error) {
	var client model.ToolCallingChatModel
	var err error

	provider := Provider(cfg.Provider)

	switch provider {
	case ProviderOpenAI:
		client, err = createOpenAIClient(cfg)
	case ProviderArk:
		client, err = createArkClient(cfg)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	return &LLM{
		provider: provider,
		client:   client,
		config:   cfg,
	}, nil
}

// createOpenAIClient 创建 OpenAI 客户端
func createOpenAIClient(cfg *config.LLMConfig) (model.ToolCallingChatModel, error) {
	chatConfig := &openai.ChatModelConfig{
		Model:       cfg.Model,
		APIKey:      cfg.APIKey,
		BaseURL:     cfg.BaseURL,
		Temperature: float32Ptr(cfg.Temperature),
		MaxTokens:   &cfg.MaxTokens,
	}

	return openai.NewChatModel(context.Background(), chatConfig)
}

// createArkClient 创建火山引擎客户端
func createArkClient(cfg *config.LLMConfig) (model.ToolCallingChatModel, error) {
	// TODO: 实现火山引擎客户端
	return nil, fmt.Errorf("ark provider not implemented yet")
}

// Generate 生成回复
func (l *LLM) Generate(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	return l.client.Generate(ctx, messages)
}

// GenerateWithTools 使用工具生成回复
func (l *LLM) GenerateWithTools(ctx context.Context, messages []*schema.Message, tools []*schema.ToolInfo) (*schema.Message, error) {
	// 使用 WithTools 创建新的实例
	modelWithTools, err := l.client.WithTools(tools)
	if err != nil {
		return nil, fmt.Errorf("failed to bind tools: %w", err)
	}

	return modelWithTools.Generate(ctx, messages)
}

// Stream 流式生成回复
func (l *LLM) Stream(ctx context.Context, messages []*schema.Message) (*schema.StreamReader[*schema.Message], error) {
	return l.client.Stream(ctx, messages)
}

// BindTools 绑定工具 (已弃用，使用 WithTools)
func (l *LLM) BindTools(tools []*schema.ToolInfo) error {
	_, err := l.client.WithTools(tools)
	return err
}

// WithTools 绑定工具，返回新的实例
func (l *LLM) WithTools(tools []*schema.ToolInfo) (*LLM, error) {
	newClient, err := l.client.WithTools(tools)
	if err != nil {
		return nil, err
	}

	return &LLM{
		provider: l.provider,
		client:   newClient,
		config:   l.config,
	}, nil
}

// Close 关闭客户端
func (l *LLM) Close() error {
	// 如果有需要关闭的资源，在这里处理
	return nil
}

// float32Ptr 返回 float32 的指针
func float32Ptr(v float64) *float32 {
	f := float32(v)
	return &f
}
