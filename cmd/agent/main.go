package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"goagent/internal/agent"
	"goagent/internal/llm"
	"goagent/internal/tools"
	"goagent/pkg/config"
)

func main() {
	// 命令行参数
	apiKey := flag.String("api-key", "", "OpenAI API Key")
	baseURL := flag.String("base-url", "", "OpenAI API Base URL")
	model := flag.String("model", "gpt-4o-mini", "Model name")
	verbose := flag.Bool("verbose", true, "Verbose output")
	maxIter := flag.Int("max-iter", 10, "Max iterations")
	flag.Parse()

	fmt.Println("🤖 Eino Agent - 基于 Eino 框架的智能助手")
	fmt.Println("============================================")
	fmt.Println()

	// 加载配置
	cfg := config.DefaultConfig()

	// 从命令行参数覆盖配置
	if *apiKey != "" {
		cfg.LLM.APIKey = *apiKey
	}
	if *baseURL != "" {
		cfg.LLM.BaseURL = *baseURL
	}
	if *model != "" {
		cfg.LLM.Model = *model
	}
	cfg.Agent.Verbose = *verbose
	cfg.Agent.MaxIterations = *maxIter

	// 从环境变量加载配置
	config.LoadFromEnv(cfg)

	// 创建 LLM 客户端
	llmClient, err := llm.NewLLM(&cfg.LLM)
	if err != nil {
		fmt.Printf("❌ 创建 LLM 客户端失败: %v\n", err)
		printHelp()
		os.Exit(1)
	}
	defer llmClient.Close()

	// 创建工具注册表
	registry := tools.NewToolRegistry()
	for _, tool := range tools.DefaultTools() {
		registry.Register(tool)
	}

	// 创建 Agent
	reactAgent := agent.NewReActAgent(llmClient, registry, &cfg.Agent)
	reactAgent.AddCallback(&agent.PrintCallback{})

	// 打印可用工具
	printAvailableTools(registry)

	// 交互式命令行
	fmt.Println("输入问题开始对话，输入 'quit' 或 'exit' 退出")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	ctx := context.Background()

	for {
		fmt.Print("👤 You: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		if input == "quit" || input == "exit" {
			fmt.Println("👋 再见！")
			break
		}

		// 特殊命令
		if strings.HasPrefix(input, "/") {
			handleCommand(input, registry)
			continue
		}

		fmt.Println()
		result, err := reactAgent.Run(ctx, input)
		if err != nil {
			fmt.Printf("❌ 执行失败: %v\n", err)
			continue
		}

		if !result.Success {
			fmt.Printf("❌ 任务执行失败: %s\n", result.Error)
		}

		fmt.Println()
		fmt.Printf("📝 执行结果:\n%s\n", result.ToJSON())
		fmt.Println()
	}
}

func printHelp() {
	fmt.Println()
	fmt.Println("请确保设置了正确的环境变量：")
	fmt.Println("  - OPENAI_API_KEY: OpenAI API 密钥")
	fmt.Println("  - OPENAI_BASE_URL: (可选) API 基础 URL，用于自定义端点")
	fmt.Println("  - LLM_MODEL: (可选) 模型名称，默认 gpt-4o-mini")
	fmt.Println()
	fmt.Println("或者使用命令行参数：")
	fmt.Println("  -api-key string")
	fmt.Println("        OpenAI API Key")
	fmt.Println("  -base-url string")
	fmt.Println("        OpenAI API Base URL")
	fmt.Println("  -model string")
	fmt.Println("        Model name (default \"gpt-4o-mini\")")
}

func printAvailableTools(registry *tools.ToolRegistry) {
	fmt.Println("可用工具：")
	for _, tool := range registry.List() {
		info := tool.Info()
		fmt.Printf("  - %s: %s\n", info.Name, info.Description)
	}
	fmt.Println()
	fmt.Println("特殊命令：")
	fmt.Println("  /tools - 列出所有可用工具")
	fmt.Println("  /help  - 显示帮助信息")
	fmt.Println()
}

func handleCommand(input string, registry *tools.ToolRegistry) {
	cmd := strings.TrimSpace(strings.TrimPrefix(input, "/"))

	switch cmd {
	case "tools":
		printAvailableTools(registry)
	case "help":
		printHelp()
	default:
		fmt.Printf("未知命令: %s\n", cmd)
		fmt.Println("输入 /help 查看可用命令")
	}
}
