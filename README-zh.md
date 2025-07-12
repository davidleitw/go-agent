# go-agent

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

一個輕量級的 Go AI 代理框架，用於建立智能對話和自動化工作流程，具有高效率。

## 特色功能

- 🚀 **輕量級與高效**: 專注於核心功能的最小化抽象
- ⚡ **函數式選項**: 使用 Go 的函數式選項模式提供清潔、直觀的 API
- 🔌 **可插拔架構**: 支援多種 LLM 提供商和儲存後端
- 🛠️ **工具整合**: 輕鬆整合自定義工具和函數呼叫
- 🔄 **流程控制**: 帶有條件規則的動態對話流程
- 📝 **結構化輸出**: 內建支援驗證的 JSON 輸出
- 💾 **會話管理**: 後端場景的持久對話歷史記錄
- 🧪 **測試支援**: 全面的模擬和測試工具

## 快速開始

### 安裝

```bash
go get github.com/davidleitw/go-agent
```

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/davidleitw/go-agent/pkg/agent"
    "github.com/davidleitw/go-agent/internal/storage"
)

func main() {
    // Create an agent with functional options
    assistant, err := agent.New(
        agent.WithName("helpful-assistant"),
        agent.WithDescription("A helpful AI assistant"),
        agent.WithInstructions("You are a helpful assistant. Be concise and friendly."),
        agent.WithOpenAI(os.Getenv("OPENAI_API_KEY")),
        agent.WithModel("gpt-4"),
        agent.WithModelSettings(&agent.ModelSettings{
            Temperature: floatPtr(0.7),
            MaxTokens:   intPtr(1000),
        }),
        agent.WithSessionStore(storage.NewInMemory()),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Have a conversation - much simpler!
    ctx := context.Background()
    response, _, err := assistant.Chat(ctx, "session-1", "Hello! How are you?")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Assistant:", response.Content)
}

func floatPtr(f float64) *float64 { return &f }
func intPtr(i int) *int { return &i }
```

### 搭配工具使用

```go
// Define a custom tool
type WeatherTool struct{}

func (t *WeatherTool) Name() string {
    return "get_weather"
}

func (t *WeatherTool) Description() string {
    return "Get current weather for a location"
}

func (t *WeatherTool) Schema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "location": map[string]interface{}{
                "type":        "string",
                "description": "The city and state/country",
            },
        },
        "required": []string{"location"},
    }
}

func (t *WeatherTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    location := args["location"].(string)
    // Simulate weather API call
    return map[string]interface{}{
        "location":    location,
        "temperature": "22°C",
        "condition":   "Sunny",
    }, nil
}

// Create agent with tool - much cleaner!
weatherAgent, err := agent.New(
    agent.WithName("weather-assistant"),
    agent.WithInstructions("You can help users get weather information."),
    agent.WithOpenAI(os.Getenv("OPENAI_API_KEY")),
    agent.WithModel("gpt-4"),
    agent.WithTools(&WeatherTool{}),
    agent.WithSessionStore(storage.NewInMemory()),
)
```

### 結構化輸出

```go
// Define output structure
type TaskResult struct {
    Title    string   `json:"title" validate:"required"`
    Priority string   `json:"priority" validate:"required,oneof=low medium high"`
    Tags     []string `json:"tags"`
}

// Create agent with structured output - much simpler!
taskAgent, err := agent.New(
    agent.WithName("task-creator"),
    agent.WithInstructions("Create tasks based on user input. Return structured JSON."),
    agent.WithOpenAI(os.Getenv("OPENAI_API_KEY")),
    agent.WithModel("gpt-4"),
    agent.WithStructuredOutput(&TaskResult{}), // Automatically generates schema
    agent.WithSessionStore(storage.NewInMemory()),
)

// The agent will automatically validate and parse the output
response, structuredOutput, err := taskAgent.Chat(ctx, "session-1", "Create a high priority task for code review")
if taskResult, ok := structuredOutput.(*TaskResult); ok {
    fmt.Printf("Created task: %s (Priority: %s)\n", taskResult.Title, taskResult.Priority)
}
```

### 流程規則

```go
// Create conditional flow rules
missingInfoCondition := agent.NewDataKeyExistsCondition("missing_info_check", "missing_fields")

flowRule, err := agent.NewFlowRule("collect-missing-info", missingInfoCondition).
    WithDescription("Prompt user for missing information").
    WithNewInstructions("Please ask the user for the following missing information: {{missing_fields}}").
    WithRecommendedTools("collect_info").
    WithSystemMessage("The user needs to provide additional information.").
    Build()

// Create agent with flow rules
smartAgent, err := agent.New(
    agent.WithName("smart-assistant"),
    agent.WithInstructions("You are a smart assistant that adapts based on context."),
    agent.WithOpenAI(os.Getenv("OPENAI_API_KEY")),
    agent.WithModel("gpt-4"),
    agent.WithFlowRules(flowRule),
    agent.WithSessionStore(storage.NewInMemory()),
)
```

## 架構

該框架採用清晰的關注點分離設計：

- **`pkg/agent/`**: 核心介面和公共 API
- **`internal/base/`**: 預設實作
- **`internal/llm/`**: LLM 提供商實作
- **`internal/storage/`**: 會話儲存實作

### 核心組件

1. **Agent**: 完整的 AI 代理，具有配置和執行功能
2. **Session**: 對話歷史記錄和狀態管理（用於後端場景）
3. **Tools**: 代理可以使用的外部功能
4. **Flow Rules**: 基於條件的動態行為控制
5. **Chat Models**: 不同 LLM 提供商的抽象化
6. **Storage**: 可插拔的會話持久化後端

## 支援的 LLM 提供商

- ✅ **OpenAI** (GPT-4, GPT-3.5-turbo, etc.)
- 🔜 **Anthropic** (Claude 3.5 Sonnet, etc.)
- 🔜 **Google** (Gemini)
- 🔜 **本地模型** (透過 Ollama)

## 儲存後端

- ✅ **記憶體**: 用於開發和測試
- ✅ **檔案系統**: 簡單的檔案持久化
- 🔜 **Redis**: 用於生產環境的分散式系統
- 🔜 **PostgreSQL**: 用於進階查詢和分析

## 範例

查看 [`cmd/examples/`](./cmd/examples/) 目錄獲得完整的工作範例：

- **基本聊天代理**: 簡單的對話式 AI
- **任務自動化代理**: 具有工具和結構化輸出的進階功能
- **多代理工作流程**: 協調的多代理互動

## 開發

### 前置需求

- Go 1.21 或更高版本
- (可選) golangci-lint 用於代碼檢查

### 建置

```bash
make build
```

### 測試

```bash
# Run all tests
make test

# Run only unit tests
make unit-test

# Run with coverage
make coverage
```

### 代碼檢查

```bash
make lint
```

## API 文件

詳細的 API 文件請參閱：

- [開始指南](./docs/getting-started.md)
- [API 參考](./docs/api-reference.md)
- [架構概述](./docs/architecture.md)
- [範例](./docs/examples.md)

## 貢獻

1. Fork 這個儲存庫
2. 創建您的功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交您的變更 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 開啟一個 Pull Request

## 許可證

本專案採用 MIT 許可證 - 詳情請參閱 [LICENSE](LICENSE) 檔案。

## 路線圖

- [ ] 額外的 LLM 提供商 (Anthropic, Google, etc.)
- [ ] 進階儲存後端 (Redis, PostgreSQL)
- [ ] 串流回應支援
- [ ] 多代理協調
- [ ] 可觀測性和指標
- [ ] 代理管理的 Web UI
- [ ] 自定義擴展的插件系統

## 支援

- 📖 [文件](./docs/)
- 🐛 [問題追蹤](https://github.com/davidleitw/go-agent/issues)
- 💬 [討論](https://github.com/davidleitw/go-agent/discussions) 