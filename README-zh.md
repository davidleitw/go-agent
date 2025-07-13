# go-agent

<div align="center">
  <img src="docs/images/gopher.png" alt="Go Agent" width="200" height="200">
</div>

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
    "github.com/davidleitw/go-agent/pkg/openai"
)

func main() {
    // 建立 OpenAI chat model
    chatModel, err := openai.NewChatModel(os.Getenv("OPENAI_API_KEY"), nil)
    if err != nil {
        log.Fatal(err)
    }

    // 使用函數選項建立 agent
    assistant, err := agent.New(
        agent.WithName("helpful-assistant"),
        agent.WithDescription("A helpful AI assistant"),
        agent.WithInstructions("You are a helpful assistant. Be concise and friendly."),
        agent.WithChatModel(chatModel),
        agent.WithModel("gpt-4"),
        agent.WithModelSettings(&agent.ModelSettings{
            Temperature: floatPtr(0.7),
            MaxTokens:   intPtr(1000),
        }),
        agent.WithSessionStore(agent.NewInMemorySessionStore()),
    )
    if err != nil {
        log.Fatal(err)
    }

    // 開始對話 - 簡單多了！
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

// 建立 OpenAI chat model
chatModel, err := openai.NewChatModel(os.Getenv("OPENAI_API_KEY"), nil)
if err != nil {
    log.Fatal(err)
}

// 建立搭配工具的 agent - 更簡潔！
weatherAgent, err := agent.New(
    agent.WithName("weather-assistant"),
    agent.WithInstructions("You can help users get weather information."),
    agent.WithChatModel(chatModel),
    agent.WithModel("gpt-4"),
    agent.WithTools(&WeatherTool{}),
    agent.WithSessionStore(agent.NewInMemorySessionStore()),
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

// 建立 OpenAI chat model
chatModel, err := openai.NewChatModel(os.Getenv("OPENAI_API_KEY"), nil)
if err != nil {
    log.Fatal(err)
}

// 建立具有結構化輸出的 agent - 更簡單！
taskAgent, err := agent.New(
    agent.WithName("task-creator"),
    agent.WithInstructions("Create tasks based on user input. Return structured JSON."),
    agent.WithChatModel(chatModel),
    agent.WithModel("gpt-4"),
    agent.WithStructuredOutput(&TaskResult{}), // 自動生成 schema
    agent.WithSessionStore(agent.NewInMemorySessionStore()),
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

// 建立 OpenAI chat model
chatModel, err := openai.NewChatModel(os.Getenv("OPENAI_API_KEY"), nil)
if err != nil {
    log.Fatal(err)
}

// 建立具有流程規則的 agent
smartAgent, err := agent.New(
    agent.WithName("smart-assistant"),
    agent.WithInstructions("You are a smart assistant that adapts based on context."),
    agent.WithChatModel(chatModel),
    agent.WithModel("gpt-4"),
    agent.WithFlowRules(flowRule),
    agent.WithSessionStore(agent.NewInMemorySessionStore()),
)
```

## 架構

該框架採用清晰的關注點分離設計：

- **`pkg/agent/`**: 核心介面、實作和公共 API
- **`pkg/openai/`**: OpenAI ChatModel 實作

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
- 🔜 **Redis**: 用於生產環境的分散式系統
- 🔜 **PostgreSQL**: 用於進階查詢和分析

## 範例

查看 [`cmd/examples/`](./cmd/examples/) 目錄獲得完整的工作範例。每個範例都是一個獨立的 Go 程式，演示 go-agent 框架的特定功能。

### 🚀 快速設定

1. **配置你的 OpenAI API 金鑰**:
   ```bash
   # 複製範例環境檔案
   cp .env.example .env
   
   # 編輯 .env 並添加你的 OpenAI API 金鑰
   # OPENAI_API_KEY=your_openai_api_key_here
   ```

2. **安裝依賴項** (針對範例):
   ```bash
   go mod download
   ```

### 📋 可用範例

#### 1. **基本聊天** (`cmd/examples/basic-chat/`)
簡單的對話式 AI，演示核心框架使用。

**特色**:
- 環境變量配置 (.env 支援)
- 使用函數式選項的基本代理創建
- 簡單的對話流程
- 詳細的日誌記錄用於排查問題

**運行範例**:
```bash
cd cmd/examples/basic-chat
go run main.go
```

**展示內容**:
- 使用 `agent.New()` 創建代理
- OpenAI 整合
- 會話管理
- 基本對話處理

---

#### 2. **任務完成** (`cmd/examples/task-completion/`)
進階範例，展示條件驗證和迭代式資訊收集。

**特色**:
- **條件式流程**: 演示缺失資訊檢測
- **結構化輸出**: 使用 JSON schema 進行狀態追蹤
- **迭代收集**: 模擬餐廳預訂系統
- **完成檢測**: LLM 在所有條件滿足時設定完成標誌
- **安全限制**: 最多 5 次迭代以防止過度使用 token

**運行範例**:
```bash
cd cmd/examples/task-completion
go run main.go
```

**展示內容**:
- 使用自定義類型的結構化輸出 (`ReservationStatus`)
- 條件驗證邏輯
- 多輪對話管理
- LLM 驅動的完成標誌檢測
- 詳細的流程日誌記錄

**模擬流程**:
1. 用戶: "我想要預訂餐廳，我是李先生" → 缺少: 電話、日期、時間、人數
2. 用戶: "我的電話是0912345678，想要明天晚上7點" → 缺少: 人數
3. 用戶: "4個人" → 所有條件滿足，completion_flag = true

---

#### 3. **計算器工具** (`cmd/examples/calculator-tool/`)
演示工具整合和 OpenAI 函數呼叫。

**特色**:
- **自定義工具實現**: 數學計算器
- **函數呼叫**: OpenAI 工具整合
- **多種運算**: 加、減、乘、除、乘方、開方
- **結構化結果**: 工具返回詳細的計算步驟
- **錯誤處理**: 除零、無效運算等

**運行範例**:
```bash
cd cmd/examples/calculator-tool
go run main.go
```

**展示內容**:
- 自定義工具實現 (`agent.Tool` 介面)
- OpenAI 函數呼叫機制
- 工具參數驗證
- 結構化工具回應
- 工具執行日誌記錄

**支援的運算**:
- `add`: 加法 (15 + 27)
- `subtract`: 減法 (125 - 47)
- `multiply`: 乘法 (13 × 7)
- `divide`: 除法 (144 ÷ 12)
- `power`: 乘方 (2^8)
- `sqrt`: 開方 (√64)

---

#### 4. **多工具智能代理** (`cmd/examples/multi-tool-agent/`)
進階 AI 助手，展示智能工具選擇和多工具協調。

**特色**:
- **上下文感知工具選擇**: 代理自動選擇適當的工具
- **多工具整合**: 天氣、計算器、時間和通知工具
- **順序工具使用**: 代理可以對複雜請求按順序使用多個工具
- **真實場景**: 多工具互動的實際範例

**可用工具**:
- 🌤️ **天氣工具**: 獲取任何地點的天氣信息
- 🧮 **計算器工具**: 執行數學計算
- ⏰ **時間工具**: 獲取不同時區的當前時間
- 📢 **通知工具**: 發送通知和提醒

**運行範例**:
```bash
cd cmd/examples/multi-tool-agent
go run main.go
```

**展示內容**:
- 基於用戶輸入的上下文感知工具選擇
- 複雜請求的多工具協調
- 工具組合場景（例如："倫敦的天氣如何，現在幾點？"）
- 多工具間的錯誤處理
- 工具編排的全面日誌記錄

**測試場景**:
- 單工具使用：天氣查詢、計算、時間請求
- 多工具組合：天氣 + 時間、計算 + 天氣
- 複雜工作流：時間查詢與預定通知

---

#### 5. **條件測試** (`cmd/examples/condition-testing/`)
使用用戶入職場景全面測試條件類型和流程規則實現。

**特色**:
- **多種條件類型**: 缺失欄位、完成階段、訊息計數
- **流程規則編排**: 動態代理行為修改
- **自定義條件實現**: 領域特定的條件邏輯
- **結構化輸出整合**: 條件與結構化數據工作
- **動態指令更新**: 基於條件的實時指令修改

**測試的條件類型**:
- 🎯 **缺失欄位條件**: 檢查缺失的數據欄位
- 📋 **完成階段條件**: 驗證當前流程階段
- 💬 **訊息計數條件**: 基於對話長度觸發
- 🔍 **數據鍵存在條件**: 內建框架條件測試

**運行範例**:
```bash
cd cmd/examples/condition-testing
go run main.go
```

**展示內容**:
- 自定義條件實現和評估邏輯
- 流程規則配置和觸發場景
- 帶條件驗證的結構化輸出
- 基於用戶狀態的動態對話流程
- 多場景的全面條件測試

**入職流程**:
- 基本信息收集（姓名）
- 聯絡詳情收集（電子郵件、電話）
- 偏好收集（興趣、愛好）
- 完成驗證和確認

### 🔧 問題排查

所有範例都包含詳細的日誌記錄，幫助你理解執行流程：

- **REQUEST**: 用戶輸入和請求參數
- **AGENT**: 代理處理和決策過程
- **TOOL**: 工具執行詳情和結果
- **RESPONSE**: LLM 回應和解析結果
- **SESSION**: 會話狀態變化
- **STRUCTURED**: 結構化輸出解析
- **ERROR**: 錯誤詳情和恢復過程

**常見問題**:

1. **缺少 API 金鑰**: 確保在 `.env` 檔案中設定了 `OPENAI_API_KEY`
2. **匯入錯誤**: 確保從範例目錄運行
3. **模組問題**: 在範例目錄中運行 `go mod tidy`

**範例日誌**:
```
✅ OpenAI API key loaded (length: 51)
📝 Creating AI agent...
✅ Agent 'helpful-assistant' created successfully
REQUEST[1]: Sending user input to agent
RESPONSE[1]: Duration: 1.234s
SESSION[1]: Total messages: 2
```

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