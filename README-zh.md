# go-agent

<div align="center">
  <img src="docs/images/gopher.png" alt="Go Agent" width="200" height="200">
</div>

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

一個輕量級的 Go AI 代理框架，專注於建構智能對話和自動化工作流程。

## 為什麼選擇 go-agent

說實話，現在市面上的 AI 框架大多都過度複雜了。我們想要的其實很簡單：給個 API key，建立一個代理，然後開始對話。就這樣。

go-agent 的設計哲學很簡單：讓常見的事情變得超級簡單，讓複雜的事情變得可能。你不需要寫 60 行代碼來建立一個基本的聊天機器人，你只需要 5 行。

## 快速開始

首先安裝 go-agent：

```bash
go get github.com/davidleitw/go-agent
```

然後寫你的第一個 AI 代理。真的，就這麼簡單：

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

看到了嗎？不需要手動建立 OpenAI 客戶端，不需要管理會話，不需要複雜的配置結構。框架會自動處理這些。

## 添加工具能力

當你的代理需要執行實際操作時，比如查天氣、做計算，你就需要工具了。以前要定義一個工具需要寫一堆接口實現，現在你只需要寫一個函數：

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

框架會自動從你的函數生成 JSON Schema，處理參數驗證，管理工具調用流程。你不需要手動處理 OpenAI 的 function calling 格式。

## 結構化輸出

有時候你希望 AI 回應特定格式的資料，比如 JSON。傳統做法是在 prompt 裡面祈禱 AI 會返回正確格式，然後手動解析。現在你只需要定義一個結構，框架會處理其他一切：

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

框架會自動生成 JSON Schema，驗證 AI 的輸出，並解析成你的 Go 結構。不用再手動處理 JSON 解析錯誤了。

## 智能流程控制

這是 go-agent 最強大的功能之一。你可以讓代理根據對話狀態自動調整行為。比如當用戶提供的資訊不完整時，自動引導他們補充：

```go
// 建立一個會自動收集缺失資訊的代理
onboardingAgent, err := agent.New("onboarding-specialist").
    WithOpenAI(apiKey).
    WithInstructions("你是一個入門指導專家，需要收集用戶的基本資訊。").
    
    // 當缺少姓名時，自動詢問
    OnMissingInfo("name").Ask("請問您的姓名是？").Build().
    
    // 當缺少電子郵件時，自動詢問
    OnMissingInfo("email").Ask("請提供您的電子郵件地址。").Build().
    
    // 當對話太長時，自動總結
    OnMessageCount(6).Summarize().Build().
    
    // 當用戶說"幫助"時，提供協助
    When(agent.WhenContains("幫助")).Ask("我該如何協助您？").Build().
    
    // 複雜條件組合：同時缺少多個資訊時
    When(agent.And(
        agent.WhenMissingFields("email"),
        agent.WhenMissingFields("phone"),
    )).Ask("我需要您的電子郵件和電話號碼才能繼續。").Build().
    
    Build()
```

這些條件會在每次對話時自動檢查，讓你的代理變得更智能、更人性化。不需要在程式碼裡寫一堆 if-else 判斷。

## 核心設計理念

我們在設計 go-agent 時有幾個核心原則：

**讓簡單的事情超級簡單**：建立一個基本的聊天機器人不應該需要讀文檔。API 應該直觀到你看一眼就知道怎麼用。

**讓複雜的事情變得可能**：當你需要進階功能時，比如多工具協調、條件流程、結構化輸出，框架應該提供強大的抽象，而不是讓你重新發明輪子。

**自動化的預設行為**：會話管理、工具調用循環、錯誤處理這些基礎設施應該默認就能正常工作，你不需要手動管理。

### 架構組成

框架主要由這幾個部分組成：

**Agent（代理）**：你的 AI 助手的大腦，負責處理對話邏輯。我們提供了 `agent.New()` 讓你快速建立，也保留了完整的介面讓你自定義。

**Session（會話）**：自動管理對話歷史。你不需要手動追蹤訊息，框架會處理。

**Tools（工具）**：讓代理能夠執行實際操作的能力。用 `agent.NewTool()` 可以快速把任何函數變成工具。

**Conditions（條件）**：智能流程控制的核心。用自然語言風格的 API 定義複雜的對話邏輯。

**Chat Models（聊天模型）**：抽象化不同的 LLM 提供商。目前支援 OpenAI，很快會支援更多。

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
