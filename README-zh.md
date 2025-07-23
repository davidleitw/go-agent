# go-agent

<div align="center">
  <img src="docs/images/gopher.png" alt="Go Agent Mascot" width="300"/>
  
  [![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)
</div>

一個簡潔但功能完整的 Go 語言 AI Agent 框架。我們設計這個框架的目標是容易上手同時保持高度可擴充性，讓你能在 Go 專案中快速整合 AI agent 功能。

## 為什麼選擇 go-agent？

雖然市面上已經有很多優秀的 agent frameworks，但我們希望能創造一個專注於簡潔性和 Go 語言慣用設計的框架。我們的設計理念是「Context is Everything」+ **Easy to Start, Easy to Scale**：

**容易上手：**
- 一個 `Execute()` method 就能開始使用
- 清晰的 module 職責，不需要理解整個框架才能用
- 豐富的 examples 和文檔，看了就會用

**高度可擴充：**
- 模組化設計，可以只用需要的部分
- 清晰的 interface definitions，容易實作自訂功能
- 開放的 Provider pattern，可以整合任何 data sources

## 快速體驗

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/davidleitw/go-agent/agent"
    "github.com/davidleitw/go-agent/llm/openai"
)

func main() {
    // 建立 LLM model
    model := openai.New(llm.Config{
        APIKey: "your-openai-key",
        Model:  "gpt-4",
    })
    
    // 建立簡單的 Agent
    myAgent := agent.NewSimpleAgent(model)
    
    // 開始對話
    response, err := myAgent.Execute(context.Background(), agent.Request{
        Input: "幫我規劃一趟東京三日遊",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(response.Output)
    fmt.Printf("使用了 %d tokens\n", response.Usage.LLMTokens.TotalTokens)
}
```

## 框架架構

我們把複雜的 AI agent 功能拆解成幾個獨立但協調良好的 modules：

```
┌─────────────┐    ┌─────────────────────────────────────┐    ┌─────────────┐
│ User Input  │───▶│           Agent.Execute()            │───▶│   Response  │
└─────────────┘    └─────────────────┬───────────────────┘    └─────────────┘
                                     │
                        ┌────────────▼────────────┐
                        │  Step 1: Session Mgmt   │
                        │    (handleSession)      │
                        └────────────┬────────────┘
                                     │
                        ┌────────────▼────────────┐
                        │ Step 2: Context Gather  │
                        │   (gatherContexts)      │
                        └────────────┬────────────┘
                                     │
               ┌─────────────────────┼─────────────────────┐
               │                     │                     │
        ┌──────▼──────┐    ┌─────────▼──────┐    ┌─────────▼──────┐
        │System Prompt│    │    History     │    │    Custom      │
        │  Provider   │    │   Management   │    │  Providers     │
        └─────────────┘    └────────────────┘    └────────────────┘
                                     │
                        ┌────────────▼────────────┐
                        │ Step 3: Execute Loop    │
                        │  (executeIterations)    │
                        │                         │
                        │  ┌─────────────────┐    │
                        │  │ Build Messages  │    │
                        │  └─────────┬───────┘    │
                        │            │            │
                        │  ┌─────────▼───────┐    │
                        │  │  LLM Call       │◄───┼──── Tool Registry
                        │  └─────────┬───────┘    │
                        │            │            │
                        │  ┌─────────▼───────┐    │
                        │  │ Tool Execution  │    │
                        │  └─────────┬───────┘    │
                        │            │            │
                        │        Iterate until    │
                        │        completion       │
                        └─────────────────────────┘
                                     │
                              ┌──────▼──────┐
                              │   Session   │
                              │   Storage   │
                              │ (TTL mgmt)  │
                              └─────────────┘
```

### Context Provider 系統 - 我們的獨特方法

讓 go-agent 與眾不同的是我們的**統一 Context 管理系統**。我們不是簡單的字串拼接，而是將 context 視為結構化資料在整個系統中流動。

**Provider Pattern：**
不同的 providers 提供不同類型的 context 資訊，全部統一成 LLM 能理解的一致格式：

```go
// 系統指令
systemProvider := context.NewSystemPromptProvider("你是一個有用的助手")

// 注意：歷史記錄現在內建於 agent engine，不需要獨立的 provider

// 從 session 狀態讀取的自訂 provider
type TaskContextProvider struct{}

func (p *TaskContextProvider) Provide(ctx context.Context, s session.Session) []context.Context {
    // 從 session 狀態讀取當前任務
    if task, exists := s.Get("current_task"); exists {
        return []context.Context{{
            Type:    "task_context",
            Content: fmt.Sprintf("當前任務：%s", task),
            Metadata: map[string]any{
                "source": "session_state",
                "key":    "current_task",
            },
        }}
    }
    return nil
}

// 實際運作方式：
session.Set("current_task", "規劃東京行程")
session.AddEntry(session.NewMessageEntry("user", "天氣如何？"))
session.AddEntry(session.NewToolCallEntry("weather", map[string]any{"city": "Tokyo"}))
session.AddEntry(session.NewToolResultEntry("weather", "22°C, 晴朗", nil))

// 當 engine 收集 contexts 時，會自動將 session entries 轉換為 contexts：
// - Message entries → user/assistant contexts  
// - Tool call entries → "Tool: weather\nParameters: {city: Tokyo}"
// - Tool result entries → "Tool: weather\nSuccess: true\nResult: 22°C, 晴朗"
// - TaskContextProvider 讀取 session.Get("current_task") → "當前任務：規劃東京行程"

agent, _ := agent.NewBuilder().
    WithLLM(model).
    WithHistoryLimit(10).  // 內建歷史記錄管理
    WithContextProviders(systemProvider, &TaskContextProvider{}).
    Build()
```

**主要優勢：**
- **自動歷史管理**：Session 對話自動轉換為 context
- **豐富的 Metadata**：每個 context 都包含 metadata 用於除錯和分析
- **TTL 整合**：Context providers 與 session 過期機制無縫配合
- **可擴展性**：輕鬆新增新的 context sources（databases、APIs、files 等）

這個方法讓「Context is Everything」不只是理念，而是從簡單 chatbots 到複雜多模態 agents 都能擴展的實際實作。

### Context vs Session - 關鍵概念釐清

理解這兩個核心概念的區別很重要：

**Context** = 資訊食材（短暫的、無狀態的）
- 每次執行時重新組裝
- 用來建構 LLM prompts
- 例如：系統指令、最近訊息、當前用戶偏好

**Session** = 狀態冰箱（持久的、有狀態的）
- 跨多次執行持續存在
- 儲存對話歷史和變數
- 例如：用戶設定、對話歷史、TTL 管理

以下展示每次請求時 contexts 如何動態組裝：

```
┌─ Step 1: Session Management ─────────────────────────────────────────┐
│ 🚀 用戶輸入："什麼時候去東京最好？"                                    │
│ 💾 Session 查找：載入會話 "user-123"                                 │
│ 找到：current_task="規劃東京行程"，3 條之前的訊息                     │
└─────────────────────────────────────────────────────────────────────┘
                                   │
┌─ Step 2: Context 組裝 ───────────────────────────────────────────────┐
│ ⚡ 從所有 providers 收集：                                           │
│                                                                      │
│ 🎯 系統 Provider →                                                   │
│   Context："你是一個有用的旅行助手。"                                │
│                                                                      │
│ 📋 任務 Provider（從 session 狀態）→                                 │
│   Context："當前任務：規劃東京行程"                                  │
│                                                                      │
│ 📜 歷史記錄（從 session 條目）→                                     │
│   Context："user: 我想規劃東京行程"                                  │
│   Context："assistant: 太好了！我很樂意幫你規劃。"                    │
│   Context："user: 我的預算是 3000 美元"                              │
│                                                                      │
│ 🔗 結果：5 個 contexts 準備給 LLM                                   │
└─────────────────────────────────────────────────────────────────────┘
                                   │
┌─ Step 3: LLM Prompt 建構 ────────────────────────────────────────────┐
│ 🤖 組合成 LLM 訊息：                                                │
│                                                                      │
│ [                                                                    │
│   {role: "system", content: "你是一個有用的旅行助手。"}              │
│   {role: "system", content: "當前任務：規劃東京行程"}                │
│   {role: "user", content: "我想規劃東京行程"}                        │
│   {role: "assistant", content: "太好了！我很樂意幫你規劃。"}          │
│   {role: "user", content: "我的預算是 3000 美元"}                    │
│   {role: "user", content: "什麼時候去東京最好？"}                     │
│ ]                                                                    │
│                                                                      │
│ 💬 LLM 回應："去東京最好的時間是..."                                 │
└─────────────────────────────────────────────────────────────────────┘
                                   │
┌─ Step 4: Session 更新 ───────────────────────────────────────────────┐
│ 💾 儲存到 session 歷史：                                             │
│   - 新用戶訊息："什麼時候去東京最好？"                               │
│   - 新助手回應："去東京最好的時間是..."                              │
│ 🔄 Session 現在有 5 條訊息供下次互動使用                            │
└─────────────────────────────────────────────────────────────────────┘
```

精髓在於 **Context** 每次都從持久的 **Session** 狀態重新組裝，確保一致性和靈活性。

### 設計有效的 Context Providers

Context Providers 是我們框架靈活性的核心。它們決定了你的 agent 能存取什麼資訊，以及如何理解對話。讓我們探索不同的模式和實際場景：

**1. 靜態 Context Providers**
這些提供一致的資訊，不受 session 狀態影響：

```go
// 系統角色定義
type RoleProvider struct {
    role string
}

func (p *RoleProvider) Provide(ctx context.Context, s session.Session) []context.Context {
    return []context.Context{{
        Type: "system",
        Content: p.role,
        Metadata: map[string]any{"priority": "high"},
    }}
}

// 使用範例：客服 agent
roleProvider := &RoleProvider{
    role: "你是一位友善的客服專員。總是要理解客戶的困擾並提供解決方案。",
}
```

**2. 動態 Session-Based Providers**
這些根據 session 狀態和歷史進行調整：

```go
// 用戶偏好 provider
type UserPreferenceProvider struct {
    userDB UserDatabase
}

func (p *UserPreferenceProvider) Provide(ctx context.Context, s session.Session) []context.Context {
    userID, exists := s.Get("user_id")
    if !exists {
        return nil // 還沒有用戶 context
    }
    
    prefs := p.userDB.GetPreferences(userID.(string))
    return []context.Context{{
        Type: "user_preferences",
        Content: fmt.Sprintf("用戶偏好：語言=%s, 風格=%s, 專業程度=%s",
            prefs.Language, prefs.CommunicationStyle, prefs.ExpertiseLevel),
    }}
}
```

**3. 條件式 Providers**
這些根據條件提供不同的 context：

```go
// 營業時間 provider
type BusinessHoursProvider struct {
    timezone string
}

func (p *BusinessHoursProvider) Provide(ctx context.Context, s session.Session) []context.Context {
    loc, _ := time.LoadLocation(p.timezone)
    now := time.Now().In(loc)
    hour := now.Hour()
    
    if hour >= 9 && hour < 17 {
        return []context.Context{{
            Type: "availability",
            Content: "營業時間內。可以提供即時協助並安排電話會議。",
        }}
    }
    
    return []context.Context{{
        Type: "availability", 
        Content: "非營業時間。仍可協助但回電將安排在下個工作日。",
    }}
}
```

**4. 外部資料 Providers**
這些從外部來源獲取即時資訊：

```go
// 旅遊 agent 的天氣 context provider
type WeatherProvider struct {
    weatherAPI WeatherService
}

func (p *WeatherProvider) Provide(ctx context.Context, s session.Session) []context.Context {
    destination, exists := s.Get("travel_destination")
    if !exists {
        return nil
    }
    
    weather := p.weatherAPI.GetCurrent(ctx, destination.(string))
    return []context.Context{{
        Type: "environment_data",
        Content: fmt.Sprintf("%s 當前天氣：%s，%d°C", 
            destination, weather.Condition, weather.Temperature),
        Metadata: map[string]any{
            "source": "weather_api",
            "timestamp": time.Now(),
        },
    }}
}
```

**5. 對話階段 Providers**
這些追蹤並提供工作流程中的位置 context：

```go
// 銷售漏斗階段 provider
type SalesFunnelProvider struct{}

func (p *SalesFunnelProvider) Provide(ctx context.Context, s session.Session) []context.Context {
    history := s.GetHistory(20)
    
    // 分析對話以判斷階段
    stage := p.analyzeStage(history)
    
    stageGuidance := map[string]string{
        "discovery": "專注於理解需求。問開放式問題。",
        "qualification": "確定預算和決策流程。",
        "proposal": "提出符合其需求的解決方案。",
        "closing": "處理異議並引導做出決定。",
    }
    
    return []context.Context{{
        Type: "sales_guidance",
        Content: fmt.Sprintf("當前階段：%s。%s", stage, stageGuidance[stage]),
    }}
}
```

**實際應用場景：**

**客戶支援 Agent：**
```go
agent := NewBuilder().
    WithLLM(model).
    WithContextProviders(
        &RoleProvider{role: "客戶支援專員"},
        &UserPreferenceProvider{userDB: db},
        &TicketInfoProvider{ticketSystem: tickets},
        &BusinessHoursProvider{timezone: "Asia/Taipei"},
        &SentimentProvider{}, // 監控對話語氣
    ).
    Build()
```

**技術文件助理：**
```go
agent := NewBuilder().
    WithLLM(model).
    WithContextProviders(
        &RoleProvider{role: "技術文件專家"},
        &CodeContextProvider{}, // 分析對話中的程式碼片段
        &VersionProvider{docDB: docs}, // 提供版本特定資訊
        &ExpertiseProvider{}, // 根據用戶程度調整說明
    ).
    Build()
```

**電商購物助理：**
```go
agent := NewBuilder().
    WithLLM(model).
    WithContextProviders(
        &RoleProvider{role: "個人購物助理"},
        &CartProvider{cartService: carts}, // 當前購物車內容
        &ProductProvider{catalog: products}, // 產品推薦
        &PriceAlertProvider{}, // 優惠和折扣
        &OrderHistoryProvider{orderDB: orders},
    ).
    Build()
```

Context Providers 的強大之處在於關注點分離 - 每個 provider 專注於 context 的一個面向，讓你的系統模組化、可測試且易於擴展。你可以混合搭配 providers 來創建完美符合使用案例的 agents！

### [Agent 模組](./agent/) - 核心控制器
這是整個框架的大腦，負責協調其他所有 modules。提供了簡單的 `Execute()` interface 和靈活的 Builder pattern 讓你能輕鬆配置各種功能。

**Key Features：**
- 簡潔的 `Agent` interface，一個 method 搞定所有事情
- Builder pattern 讓配置變得很直觀
- 自動 session management，不用擔心 state 問題
- 內建的 convenience functions，常見用法一行搞定

### [Session 模組](./session/) - 記憶管理
負責管理對話的 state 和 history records。支援 TTL 自動過期、concurrent safety、還有完整的 JSON serialization。

**Key Features：**
- Key-Value state storage，什麼資料類型都能放
- 統一的 history record format，支援多種對話類型
- 自動 TTL management，過期 sessions 會自動 cleanup
- Thread-safe，多 goroutine 使用沒問題

### [Context 模組](./context/) - 資訊聚合
這個模組的工作是把各種來源的資訊（history conversations、system prompts、external data 等）統一打包成 LLM 能理解的格式。

**Key Features：**
- 統一的 `Context` data structure
- 可擴展的 `Provider` system
- 自動將 Session history 轉換成 contexts
- 豐富的 metadata 支援

### [Tool 模組](./tool/) - 工具整合
讓你的 AI agents 能夠呼叫外部功能，比如查詢資料庫、呼叫 API、執行計算等等。

**Key Features：**
- 簡單的 `Tool` interface，很容易實作 custom tools
- 基於 JSON Schema 的 parameter definitions
- Thread-safe 的 tool registry
- 完整的 error handling 機制

### [LLM 模組](./llm/) - 語言模型介面
提供統一的 language model interface，目前支援 OpenAI，未來會擴展到其他提供商。

**Key Features：**
- 清晰的 `Model` interface
- 內建 tool calling 支援
- 完整的 token usage tracking
- 支援 custom endpoints 和 proxies

## History Management（歷史記錄管理）

go-agent 框架提供靈活的對話歷史記錄管理，可以從簡單使用場景擴展到類似 Claude Code 等級的複雜實作。

### 基本使用

通過簡單的限制啟用歷史記錄追蹤：

```go
agent := agent.NewBuilder().
    WithLLM(model).
    WithHistoryLimit(20).  // 保留最近 20 輪對話
    Build()
```

### 進階歷史記錄處理

對於需要壓縮、過濾或自動摘要的複雜場景，可以實作 `HistoryInterceptor` 介面：

```go
type HistoryInterceptor interface {
    ProcessHistory(ctx context.Context, entries []session.Entry, llm llm.Model) ([]session.Entry, error)
}
```

### Claude Code 等級的實作範例

以下展示如何實作類似 Claude Code 的複雜歷史記錄管理：

```go
type AdvancedHistoryCompressor struct {
    maxTokens        int
    recentLimit      int
    compressionRatio float32
}

func (c *AdvancedHistoryCompressor) ProcessHistory(ctx context.Context, entries []session.Entry, llm llm.Model) ([]session.Entry, error) {
    if len(entries) <= c.recentLimit {
        return entries, nil
    }

    // 1. 保留最近的對話
    recent := entries[len(entries)-c.recentLimit:]
    older := entries[:len(entries)-c.recentLimit]

    // 2. 識別重要的條目
    important := c.filterImportant(older)
    
    // 3. 使用 LLM 生成壓縮摘要
    summary, err := c.generateSummary(ctx, older, llm)
    if err != nil {
        return entries, nil // 錯誤時回退到原始歷史記錄
    }

    // 4. 組合：摘要 + 重要條目 + 最近對話
    result := []session.Entry{summary}
    result = append(result, important...)
    result = append(result, recent...)
    
    return result, nil
}

func (c *AdvancedHistoryCompressor) generateSummary(ctx context.Context, entries []session.Entry, llm llm.Model) (session.Entry, error) {
    // 建構壓縮 prompt
    historyText := c.formatEntriesForSummary(entries)
    
    response, err := llm.Complete(ctx, llm.Request{
        Messages: []llm.Message{
            {
                Role: "system", 
                Content: "你是對話摘要器。保留關鍵資訊、決策和上下文。",
            },
            {
                Role: "user",
                Content: fmt.Sprintf("摘要這段對話歷史：\n\n%s", historyText),
            },
        },
    })
    
    if err != nil {
        return session.Entry{}, err
    }
    
    // 以 system message entry 形式返回
    return session.NewMessageEntry("system", 
        fmt.Sprintf("[壓縮歷史記錄摘要]\n%s", response.Content)), nil
}

func (c *AdvancedHistoryCompressor) filterImportant(entries []session.Entry) []session.Entry {
    var important []session.Entry
    
    for _, entry := range entries {
        // 自訂重要性評分邏輯
        if c.isImportant(entry) {
            important = append(important, entry)
        }
    }
    
    return important
}

func (c *AdvancedHistoryCompressor) isImportant(entry session.Entry) bool {
    // 重要性判斷標準範例：
    // - 錯誤訊息
    // - 成功的工具執行且有價值的結果
    // - 使用者偏好或設定
    // - 關鍵決策或確認
    
    if entry.Type == session.EntryTypeToolResult {
        if content, ok := session.GetToolResultContent(entry); ok {
            return !content.Success || c.hasValueableResult(content.Result)
        }
    }
    
    // 檢查錯誤關鍵字、偏好設定等
    return false
}

// 使用方式
compressor := &AdvancedHistoryCompressor{
    maxTokens:        4000,
    recentLimit:      10,
    compressionRatio: 0.3,
}

agent := agent.NewBuilder().
    WithLLM(model).
    WithHistoryLimit(100).
    WithHistoryInterceptor(compressor).
    Build()
```

### 主要特色

**Advanced Compression：**
- 基於 LLM 的摘要生成
- 基於重要性的 entry 保留
- Token 限制管理
- 可配置的壓縮比例

**Context 感知：**
- 在 system prompt 中自動加入歷史記錄提示
- 維持對話連續性
- 保留關鍵資訊

**Performance 優化：**
- 內部歷史記錄處理（無 ContextProvider 額外開銷）
- 支援 async 處理
- 高效的 entry 轉換

**Extensible Design：**
- 簡單的 interface 方便自訂實作
- 完整的 LLM 處理能力
- 與 session metadata 整合

### System Prompt 整合

當歷史記錄被處理時，系統會自動告知 LLM：

```
Note on Conversation History:
The conversation history provided may have been compressed or summarized to save space.
Key information and context have been preserved, but some details might be condensed.
Please use this history as reference for maintaining conversation continuity and context.
```

這種設計讓您能夠建構複雜的對話 agent，在長時間互動中維持上下文的同時，有效管理 token 成本和處理效率。

## 目前開發狀態

**Ready to Use：**
- 完整的 module interfaces 設計和實作
- Session management 和 TTL 支援
- Context provider system
- Tool registration 和 execution framework
- OpenAI 整合
- 豐富的 test coverage

**In Development：**
- Agent 的核心 execution logic（LLM calls、tool orchestration、iterative thinking 等）
- 更多 LLM providers 支援
- Streaming responses 支援
- 更多內建 tools 和 examples

**Future Plans：**
- Redis/Database 的 Session storage
- Async tool execution
- 更進階的 Context management 功能
- MCP (Model Context Protocol) tool 整合

## 設計哲學

### "Context is Everything"
我們相信 AI agents 的核心就是管理 context。不管是 conversation history、user preferences、external data，或是 tool execution results，都需要以一致的方式提供給 LLM。

我們計劃組織相關的 talks 並整理 Context Engineering 的資源，幫助社群更好地理解這個方法。

## 參與開發

這個專案還在積極開發中，我們非常歡迎各種形式的參與：

**Interface 設計討論（最重要！）：**
- 覺得某個 interface 設計不夠直觀嗎？
- 有更好的 API 設計想法嗎？
- 認為某些功能的抽象層次不對嗎？
- 希望某個 module 提供不同的使用方式嗎？

我們深信好的 interface design 是框架成功的關鍵，任何對 interfaces 有想法的朋友都非常歡迎提出討論！

**功能建議：**
- 希望增加什麼新功能？
- 遇到什麼使用上的困難？
- 有什麼實際 application scenarios 我們沒考慮到？

**程式碼貢獻：**
- 實作新的 LLM providers
- 建立更多實用的 tools
- 改善 performance 和 stability
- 增加更多 tests 和 examples

**文檔和範例：**
- 撰寫使用教學
- 建立實際的 application examples
- 翻譯文檔

隨時可以開 Issue 討論，或者直接發 PR。我們很樂意跟大家一起把這個框架做得更好用。

## 如何開始

1. **查看 module 文檔**：每個資料夾都有詳細的 README，建議先從 [Agent 模組](./agent/) 開始看
2. **執行測試**：`go test ./...` 看看所有功能是否正常
3. **加入討論**：有問題或想法就開 Issue 聊聊

## 授權

MIT License

---

期待看到你用這個框架做出什麼有趣的東西！