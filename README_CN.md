# Eino Agent

[English Documentation](./README.md)

基于字节跳动 [Eino](https://github.com/cloudwego/eino) 框架实现的 ReAct (Reasoning + Acting) Agent。

## 功能特性

- **ReAct 范式**: 实现思考-行动-观察循环，智能执行任务
- **工具系统**: 可扩展的工具注册表，内置计算器、天气、时间、搜索等工具
- **LLM 集成**: 支持 OpenAI 兼容的 LLM 客户端，具备工具调用能力
- **回调系统**: 观察者模式监控 Agent 执行过程
- **灵活配置**: 支持环境变量和命令行参数配置

## 项目结构

```
goagent/
├── cmd/
│   └── agent/
│       └── main.go          # 应用入口
├── internal/
│   ├── agent/
│   │   └── react.go         # ReAct Agent 实现
│   ├── llm/
│   │   └── llm.go           # LLM 客户端封装
│   └── tools/
│       ├── tool.go          # 工具接口和注册表
│       └── builtin.go       # 内置工具
├── pkg/
│   └── config/
│       └── config.go        # 配置管理
├── go.mod
└── go.sum
```

## 快速开始

### 环境要求

- Go 1.21+
- OpenAI API Key（或兼容的端点）

### 安装

```bash
# 克隆仓库
git clone <repository-url>
cd goagent

# 安装依赖
go mod tidy
```

### 配置

设置环境变量：

```bash
export OPENAI_API_KEY="your-api-key"
export OPENAI_BASE_URL="https://api.openai.com/v1"  # 可选，用于自定义端点
export LLM_MODEL="gpt-4o-mini"                       # 可选，默认: gpt-4o-mini
```

或使用命令行参数：

```bash
go run ./cmd/agent -api-key="your-api-key" -model="gpt-4o-mini"
```

### 运行

```bash
go run ./cmd/agent
```

## 使用示例

```
🤖 Eino Agent - 基于 Eino 框架的智能助手
============================================

可用工具：
  - calculator: 执行数学计算，支持加减乘除运算
  - get_weather: 获取指定城市的天气信息
  - get_current_time: 获取当前时间
  - search: 搜索互联网信息

特殊命令：
  /tools - 列出所有可用工具
  /help  - 显示帮助信息

输入问题开始对话，输入 'quit' 或 'exit' 退出

👤 You: 
```

## 创建自定义工具

实现 `Tool` 接口：

```go
type Tool interface {
    Info() *ToolInfo
    Execute(ctx context.Context, input string) (string, error)
}
```

示例：

```go
package main

import (
    "context"
    "github.com/cloudwego/eino/schema"
    "goagent/internal/tools"
)

func NewMyCustomTool() *tools.BaseTool {
    return tools.NewBaseTool(
        "my_tool",
        "工具功能描述",
        map[string]*schema.ParameterInfo{
            "param1": {
                Type:     schema.String,
                Desc:     "参数描述",
                Required: true,
            },
        },
        func(ctx context.Context, input string) (string, error) {
            // 解析输入并执行工具逻辑
            return "result", nil
        },
    )
}

// 注册工具
registry := tools.NewToolRegistry()
registry.Register(NewMyCustomTool())
```

## ReAct Agent 架构

ReAct Agent 遵循思考-行动-观察循环：

1. **思考 (Thought)**: Agent 分析输入并规划下一步
2. **行动 (Action)**: Agent 选择并执行工具
3. **观察 (Observation)**: Agent 观察工具输出
4. **循环**: 重复上述步骤直到任务完成或达到最大迭代次数

```
┌─────────────────────────────────────────────────────┐
│                    用户输入                          │
└─────────────────────┬───────────────────────────────┘
                      ▼
┌─────────────────────────────────────────────────────┐
│                  LLM (带工具)                        │
└─────────────────────┬───────────────────────────────┘
                      ▼
            ┌─────────────────┐
            │   有工具调用?    │
            └────────┬────────┘
                     │
        ┌────────────┼────────────┐
        ▼            ▼            ▼
      是           否          错误
        │            │            │
        ▼            ▼            ▼
┌──────────────┐ ┌────────┐ ┌──────────┐
│  执行工具     │ │  完成  │ │  处理    │
└──────┬───────┘ └────────┘ │  错误    │
       │                    └──────────┘
       ▼
┌──────────────┐
│   观察结果    │
└──────┬───────┘
       │
       └──────────────▶ 回到 LLM 继续循环
```

## 命令行参数

```
  -api-key string
        OpenAI API Key
  -base-url string
        OpenAI API Base URL
  -model string
        模型名称 (默认 "gpt-4o-mini")
  -verbose
        详细输出 (默认 true)
  -max-iter int
        最大迭代次数 (默认 10)
```

## 依赖

- [github.com/cloudwego/eino](https://github.com/cloudwego/eino) - LLM 应用开发框架
- [github.com/cloudwego/eino-ext/components/model/openai](https://github.com/cloudwego/eino-ext) - OpenAI 模型集成

## 许可证

MIT License

## 致谢

- [CloudWeGo Eino](https://github.com/cloudwego/eino) - 底层 LLM 应用框架
- [MOA AI Agent](https://km.sankuai.com/collabpage/2742907745) - ReAct 实现的参考灵感
