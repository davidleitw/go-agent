# go-agent

<div align="center">
  <img src="docs/images/gopher.png" alt="Go Agent" width="200" height="200">
</div>

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

一個輕量級的 Go AI 代理框架，專注於建構智能對話和自動化工作流程。

> ⚠️ **早期開發階段**：此框架目前正處於早期開發階段。API 可能會因應使用者反饋而頻繁調整。我們計劃在 v0.1.0 版本發布後穩定 API。生產環境中請謹慎使用。

## 開發狀態

**目前階段**：早期開發階段（Pre-v0.1.0）
- ✅ 核心功能已實作並測試
- ⚠️ API 可能會根據使用者反饋進行調整
- 🔄 介面優化持續進行中
- 📋 計劃在 v0.1.0 後穩定 API

歡迎在我們部署穩定版本的過程中提供反饋和建議。

## 為什麼選擇 go-agent

go-agent 提供直觀的介面來建構 AI 應用程式。框架專注於最少配置：提供 API key，建立代理，開始對話。

設計優先考慮常見用例的簡潔性，同時為複雜場景保持靈活性。建立基本聊天機器人需要最少的程式碼。

## 快速開始

首先安裝 go-agent：

```bash
go get github.com/davidleitw/go-agent
```

建立你的第一個 AI 代理：

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/davidleitw/go-agent/pkg/agent"
)

func main() {
    // 建立一個 AI 代理，就這一行
    assistant, err := agent.New("helpful-assistant").
        WithOpenAI(os.Getenv("OPENAI_API_KEY")).
        WithModel("gpt-4o-mini").
        WithInstructions("你是一個有用的助手，請簡潔友善地回應。").
        Build()
    if err != nil {
        panic(err)
    }

    // 開始對話
    response, err := assistant.Chat(context.Background(), "你好！今天過得如何？")
    if err != nil {
        panic(err)
    }

    fmt.Println("助手：", response.Message)
}
```

框架會自動處理 OpenAI 客戶端建立、session 管理和配置設定。

## 核心功能

### 工具整合

**使用時機**：當代理需要執行外部操作，如 API 呼叫、計算或資料處理。

工具讓代理能與外部系統互動。使用簡單的函數語法定義工具：

```go
// 建立一個天氣查詢工具，直接用函數定義
weatherTool := agent.NewTool("get_weather", 
    "查詢指定地點的天氣資訊",
    func(location string) map[string]any {
        // 模擬天氣 API 調用
        return map[string]any{
            "location":    location,
            "temperature": "22°C",
            "condition":   "晴天",
        }
    })

// 建立有工具能力的代理
weatherAgent, err := agent.New("weather-assistant").
    WithOpenAI(apiKey).
    WithInstructions("你可以幫用戶查詢天氣資訊。").
    WithTools(weatherTool).
    Build()
```

框架會自動生成 JSON Schema、處理參數驗證和管理工具執行流程。

**完整範例**：[Calculator Tool Example](./examples/calculator-tool/)

### 結構化輸出

**使用時機**：當你需要代理回傳特定格式的資料，供後續處理使用。

使用 Go 結構定義結構化輸出：

```go
// 定義你想要的輸出格式
type TaskResult struct {
    Title    string   `json:"title"`
    Priority string   `json:"priority"`
    Tags     []string `json:"tags"`
}

// 建立會返回結構化資料的代理
taskAgent, err := agent.New("task-creator").
    WithOpenAI(apiKey).
    WithInstructions("根據用戶輸入建立任務，返回結構化的 JSON 資料。").
    WithOutputType(&TaskResult{}).
    Build()

// 對話會自動返回解析好的結構
response, err := taskAgent.Chat(ctx, "建立一個高優先級的程式碼審查任務")
if taskResult, ok := response.Data.(*TaskResult); ok {
    fmt.Printf("建立任務：%s (優先級：%s)\n", taskResult.Title, taskResult.Priority)
}
```

框架會自動生成 JSON Schema、驗證 AI 輸出，並解析成 Go 結構。

**完整範例**：[Task Completion Example](./examples/task-completion/)

### Schema 式資訊收集

**使用時機**：當你需要跨對話輪次收集結構化資料，如表單填寫、使用者引導或支援工單建立。

Schema 系統會自動從使用者訊息中提取資訊並管理收集狀態。這消除了手動狀態管理的需求，提供自然的對話流程。

#### 基本 Schema 定義

```go
import "github.com/davidleitw/go-agent/pkg/schema"

// 必需欄位（預設）
emailField := schema.Define("email", "請提供您的電子郵件地址")
issueField := schema.Define("issue", "請描述您的問題")

// 可選欄位
phoneField := schema.Define("phone", "緊急聯絡電話").Optional()
```

#### 在對話中應用 Schema

```go
supportBot, err := agent.New("support-agent").
    WithOpenAI(apiKey).
    WithInstructions("您是客戶支援助手。").
    Build()

response, err := supportBot.Chat(ctx, "我的帳戶需要幫助",
    agent.WithSchema(
        schema.Define("email", "請提供您的電子郵件地址"),
        schema.Define("issue", "請詳細描述您的問題"),
        schema.Define("urgency", "這有多緊急？").Optional(),
    ),
)
```

框架會智能地：
- **提取** 使用 LLM 語義理解從使用者訊息中提取資訊
- **識別** 自動識別缺失的必需欄位
- **詢問** 使用自然、符合上下文的提示詢問缺失資訊
- **記憶** 跨對話輪次記住收集的資訊
- **適應** 不同的對話風格和使用者輸入模式

#### 動態 Schema 選擇

**使用時機**：當不同對話類型需要不同資訊時（例如支援請求與銷售查詢）。

```go
func getSchemaForIntent(intent string) []*schema.Field {
    switch intent {
    case "technical_support":
        return []*schema.Field{
            schema.Define("email", "技術追蹤用電子郵件"),
            schema.Define("error_message", "您看到什麼錯誤？"),
            schema.Define("steps_taken", "您嘗試了什麼？"),
        }
    case "billing_inquiry":
        return []*schema.Field{
            schema.Define("email", "帳戶電子郵件地址"),
            schema.Define("account_id", "您的帳戶號碼"),
            schema.Define("billing_question", "帳單問題詳情"),
        }
    }
}

// 根據檢測的意圖應用 schema
intent := detectIntent(userInput)
schema := getSchemaForIntent(intent)
response, err := agent.Chat(ctx, userInput, agent.WithSchema(schema...))
```

#### 多步驟工作流程

**使用時機**：對於應該分解為邏輯步驟的複雜表單或流程。

```go
func getTechnicalSupportWorkflow() [][]*schema.Field {
    return [][]*schema.Field{
        { // 步驟 1：聯絡資訊
            schema.Define("email", "您的電子郵件地址"),
            schema.Define("issue_summary", "問題簡要描述"),
        },
        { // 步驟 2：技術詳情
            schema.Define("error_message", "確切的錯誤訊息"),
            schema.Define("browser", "瀏覽器和版本"),
        },
        { // 步驟 3：影響評估
            schema.Define("urgency", "這有多重要？"),
            schema.Define("affected_users", "有多少使用者受影響？"),
        },
    }
}
```

**完整範例**：
- [Simple Schema Example](./examples/simple-schema/) - 基礎用法
- [Customer Support Example](./examples/customer-support/) - 真實情境
- [Dynamic Schema Example](./examples/dynamic-schema/) - 進階工作流程

### 條件式流程控制

**使用時機**：當你需要代理根據對話上下文、使用者狀態或外部條件做出不同回應。

流程控制透過條件和規則實現動態代理行為。這對建立智能、感知上下文的對話至關重要。

#### 內建條件

常見對話情境的通用條件：

```go
import "github.com/davidleitw/go-agent/pkg/conditions"

// 文字型條件
conditions.Contains("help")       // 使用者訊息包含 "help"
conditions.Count(5)               // 對話有 5+ 則訊息
conditions.Missing("email", "name") // 必需欄位缺失
conditions.DataEquals("status", "urgent") // 資料欄位具有特定值

// 自訂函數條件
conditions.Func("custom_check", func(session conditions.Session) bool {
    // 自訂邏輯
    return len(session.Messages()) > 3
})
```

#### 自訂條件

實作 `Condition` 介面處理複雜邏輯：

```go
type BusinessHoursCondition struct{}

func (c *BusinessHoursCondition) Name() string {
    return "business_hours"
}

func (c *BusinessHoursCondition) Evaluate(ctx context.Context, session agent.Session, data map[string]interface{}) (bool, error) {
    now := time.Now()
    hour := now.Hour()
    return hour >= 9 && hour <= 17, nil
}

// 使用自訂條件
businessRule := agent.FlowRule{
    Name:      "office_hours_response",
    Condition: &BusinessHoursCondition{},
    Action: agent.FlowAction{
        NewInstructionsTemplate: "您可以在營業時間獲得完整支援。",
    },
}
```

#### 組合條件

```go
// 邏輯運算子
conditions.And(conditions.Contains("urgent"), conditions.Missing("phone"))
conditions.Or(conditions.Contains("help"), conditions.Contains("support"))
conditions.Not(conditions.Missing("email"))

// 複雜條件組合
complexCondition := conditions.And(
    conditions.Or(conditions.Contains("billing"), conditions.Contains("payment")),
    conditions.Missing("account_id"),
    conditions.Count(2),
)

// 流暢介面建構複雜條件
complexCondition := conditions.Contains("support").
    And(conditions.Missing("email")).
    Or(conditions.Count(5)).
    Build()
```

**完整範例**：
- [Condition Testing Example](./examples/condition-testing/) - 基礎流程控制
- [Advanced Conditions Example](./examples/advanced-conditions/) - 複雜情境

## 核心設計理念

框架設計遵循以下原則：

**常見用例的簡潔性**：基本功能需要最少的配置。建立代理和管理對話等重要操作使用直接的 API。

**複雜情境的靈活性**：進階功能包括多工具協調、條件流程和結構化輸出，透過可組合的介面提供。

**自動基礎設施管理**：Session 管理、工具執行和錯誤處理無需手動干預即可運作。

### 架構組成

框架主要由這幾個部分組成：

**Agent**：對話處理的核心介面。使用 `agent.New()` 建立或透過 `Agent` 介面實作自訂邏輯。

**Session**：管理對話歷史和狀態。跨對話輪次的自動持久化和檢索。

**Tools**：透過 `Tool` 介面啟用外部操作。使用 `agent.NewTool()` 將函數轉換為工具。

**Conditions**：透過 `conditions` 套件進行流程控制。提供常見情境的內建條件，包括文字匹配、欄位驗證和訊息計數。

**Schema**：透過 `schema` 套件進行資訊收集。自動提取和驗證結構化資料。

**Chat Models**：LLM 供應商抽象。支援 OpenAI，其他供應商開發中。

## 支援的 LLM 提供商

目前主要支援 OpenAI 的模型，包括 GPT-4、GPT-4o、GPT-3.5-turbo 等。我們正在積極開發對其他提供商的支援：

**已支援**：OpenAI（完整支援，包括 function calling 和結構化輸出）

**開發中**：Anthropic Claude、Google Gemini、本地模型（透過 Ollama）

## 會話儲存

框架自帶記憶體會話儲存，適合開發和測試。生產環境的話，我們正在開發 Redis 和 PostgreSQL 後端支援。

不過老實說，對於大部分應用來說，記憶體儲存已經足夠了。你可以隨時實現自己的儲存後端。

## 範例程式

我們在 [`examples/`](./examples/) 目錄裡準備了完整的範例，每個都是可以直接執行的 Go 程式。

### 快速設定

先設定你的 OpenAI API key：

```bash
# 複製範例環境檔案
cp .env.example .env

# 編輯 .env，加入你的 OpenAI API key
```

### 主要範例

**基本聊天（basic-chat）**：最簡單的起點，展示如何用幾行代碼建立聊天機器人。

**計算器工具（calculator-tool）**：展示如何讓代理使用工具，這個例子會建立一個會做數學運算的助手。

**進階條件（advanced-conditions）**：展示智能流程控制，代理會根據對話狀態自動調整行為。這是我們最推薦的範例，展示了框架的強大功能。

**多工具代理（multi-tool-agent）**：展示如何讓一個代理同時使用多個工具，智能選擇合適的工具來完成任務。

**任務完成（task-completion）**：展示結構化輸出和條件驗證，模擬餐廳預訂系統。

每個範例都有詳細的 README 說明如何執行和重點學習內容。建議從 basic-chat 開始，然後嘗試 advanced-conditions。

## 常見問題

如果遇到問題，先檢查這幾個：

**API Key 設定錯誤**：確保 `.env` 檔案裡有正確的 `OPENAI_API_KEY`

**匯入錯誤**：確保你在正確的目錄執行，並且使用 `github.com/davidleitw/go-agent/pkg/agent`

**模組問題**：在範例目錄執行 `go mod tidy`

所有範例都有詳細的日誌輸出，可以幫你追蹤執行流程和錯誤。

## 開發相關

如果你想參與開發或者客製化框架：

```bash
# 執行測試
make test

# 程式碼檢查
make lint

# 建構專案
make build
```

需要 Go 1.22 或更新版本。

## 未來計畫

我們正在開發這些功能：

更多 LLM 提供商支援（Anthropic、Google 等）、生產級儲存後端（Redis、PostgreSQL）、串流回應、多代理協調、監控和觀測功能。

如果你有特定需求或想法，歡迎在 [GitHub Issues](https://github.com/davidleitw/go-agent/issues) 提出討論。

## 總結

go-agent 的目標是讓 Go 開發者能夠快速建構 AI 應用，而不需要深入了解各種 LLM API 的細節。我們相信好的框架應該讓常見任務變得簡單，讓複雜任務變得可能。

如果你正在考慮為你的 Go 專案添加 AI 功能，試試 go-agent 吧。從一個簡單的聊天機器人開始，當你需要更多功能時，框架會跟著你的需求成長。
