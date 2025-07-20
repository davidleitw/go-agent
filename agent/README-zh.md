# Agent 模組

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

Agent 模組提供在 Go Agent Framework 中創建和使用 AI 代理的主要接口。它協調 LLM 模型、工具、會話管理和上下文提供器，創建強大且有狀態的 AI 互動。

## 功能特色

- **簡潔的 Agent 介面**：乾淨的 `Agent.Execute()` 方法執行代理任務
- **建造者模式**：使用流暢 API 靈活建構代理
- **會話管理**：具有持久記憶和 TTL 支援的有狀態對話
- **工具整合**：無縫的工具呼叫和執行
- **上下文提供器**：自動從各種來源收集資訊
- **便利函數**：常見使用案例的簡單模式
- **可擴展引擎**：可插拔的執行引擎以實現不同行為
- **效能優化**：預快取會話選項和高效的元件管理

## 快速開始

### 簡單代理

```go
import (
    "github.com/davidleitw/go-agent/agent"
    "github.com/davidleitw/go-agent/llm/openai"
)

// 創建簡單代理
model := openai.New(llm.Config{
    APIKey: "your-api-key",
    Model:  "gpt-4",
})

myAgent := agent.NewSimpleAgent(model)

// 使用代理
response, err := myAgent.Execute(ctx, agent.Request{
    Input: "法國的首都是什麼？",
})

fmt.Println(response.Output) // "法國的首都是巴黎。"
```

### 帶工具的代理

```go
// 定義自訂工具
type WeatherTool struct{}

func (w *WeatherTool) Definition() tool.Definition {
    return tool.Definition{
        Type: "function",
        Function: tool.Function{
            Name:        "get_weather",
            Description: "取得某地點的目前天氣",
            Parameters: tool.Parameters{
                Type: "object",
                Properties: map[string]tool.Property{
                    "location": {
                        Type:        "string",
                        Description: "城市名稱",
                    },
                },
                Required: []string{"location"},
            },
        },
    }
}

func (w *WeatherTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    location := params["location"].(string)
    return fmt.Sprintf("%s 的天氣：22°C，晴朗", location), nil
}

// 創建帶工具的代理
weatherTool := &WeatherTool{}
myAgent := agent.NewAgentWithTools(model, weatherTool)

response, _ := myAgent.Execute(ctx, agent.Request{
    Input: "東京的天氣如何？",
})
```

## 建造者模式

對於進階配置，使用建造者模式：

```go
agent, err := agent.NewBuilder().
    WithLLM(model).
    WithMemorySessionStore().
    WithTools(weatherTool, calculatorTool).
    WithSessionHistory(20).
    WithSessionTTL(6*time.Hour).         // 會話在 6 小時後過期
    WithMaxIterations(5).
    WithTemperature(0.7).
    Build()

if err != nil {
    log.Fatal(err)
}

response, _ := agent.Execute(ctx, agent.Request{
    Input:     "幫我規劃東京行程",
    SessionID: "user-123", // 可選：使用現有會話
})
```

## 便利函數

### 一次性聊天

```go
// 無會話管理的簡單聊天
response, err := agent.Chat(ctx, model, "你好，你好嗎？")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response)
```

### 對話介面

```go
// 自動會話管理的多輪對話
conv := agent.NewConversationWithModel(model)

response1, _ := conv.Say(ctx, "你好！")
response2, _ := conv.Say(ctx, "我剛剛說了什麼？")
fmt.Println(response2) // 代理記住先前的訊息

// 重置對話
conv.Reset()
```

### 不帶會話的多輪對話

```go
// 無會話持久性的簡單多輪對話
mt := agent.NewMultiTurn(model)

response1, _ := mt.Ask(ctx, "什麼是機器學習？")
response2, _ := mt.Ask(ctx, "你能給我一個例子嗎？")

// 取得對話歷史
history := mt.GetHistory()
```

## API 參考

### Agent 介面

```go
type Agent interface {
    Execute(ctx context.Context, request Request) (*Response, error)
}
```

### 請求結構

```go
type Request struct {
    Input     string            // 用戶輸入或指令
    SessionID string            // 可選的會話 ID
}
```

### 回應結構

```go
type Response struct {
    Output    string            // 代理的回應
    SessionID string            // 使用的會話 ID
    Session   session.Session   // 存取會話狀態
    Metadata  map[string]any    // 額外的回應資料
    Usage     Usage             // 資源使用資訊
}
```

### 使用量追蹤

```go
type Usage struct {
    LLMTokens     TokenUsage    // 語言模型 token 使用量
    ToolCalls     int           // 工具執行次數
    SessionWrites int           // 會話狀態修改次數
}

type TokenUsage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}
```

## 建造者選項

### 核心元件

```go
builder := agent.NewBuilder()

// 必要：設定語言模型
builder.WithLLM(model)

// 可選：設定會話儲存
builder.WithMemorySessionStore()        // 記憶體內（預設）
builder.WithSessionStore(customStore)   // 自訂儲存

// 可選：新增工具
builder.WithTools(tool1, tool2)
builder.WithToolRegistry(registry)

// 可選：新增上下文提供器
builder.WithContextProviders(provider1, provider2)
builder.WithSessionHistory(20)          // 包含對話歷史
```

### 配置選項

```go
// 會話管理
builder.WithSessionTTL(24*time.Hour)    // 會話過期時間（預設：24小時）

// 執行限制
builder.WithMaxIterations(5)            // 最大思考迴圈次數

// LLM 參數
builder.WithTemperature(0.7)            // 回應創意度
builder.WithMaxTokens(1000)             // 回應長度限制
```

## 上下文提供器

上下文提供器為代理收集資訊：

```go
import (
    agentcontext "github.com/davidleitw/go-agent/context"
)

// 系統提示提供器
systemProvider := agentcontext.NewSystemPromptProvider("你是一個有用的助手")

// 歷史提供器（最後 10 條訊息）
historyProvider := agentcontext.NewHistoryProvider(10)

// 動態資訊的自訂用戶上下文提供器
type UserContextProvider struct {
    userPreferences map[string]any
}

func (p *UserContextProvider) Provide(ctx context.Context, s session.Session) []agentcontext.Context {
    return []agentcontext.Context{{
        Type:    "user_info",
        Content: fmt.Sprintf("用戶偏好：%v", p.userPreferences),
    }}
}

userProvider := &UserContextProvider{
    userPreferences: map[string]any{
        "language": "中文",
        "location": "東京",
    },
}

agent, _ := agent.NewBuilder().
    WithLLM(model).
    WithContextProviders(systemProvider, historyProvider, userProvider).
    Build()
```

## 會話管理

會話提供有狀態的對話，具有自動過期和中繼資料追蹤功能。

### 自動會話創建

```go
// 代理自動創建新會話，預設 24 小時 TTL
response, _ := agent.Execute(ctx, agent.Request{
    Input: "你好！",
    // SessionID 留空 - 自動創建新會話
})

sessionID := response.SessionID // 用於後續互動

// 會話包含中繼資料和狀態：
// - 中繼資料：created_by="agent", agent_version="v1.0"
// - 狀態：initial_input_length, session_start_time 等
```

### 明確會話管理

```go
// 使用特定會話
response, _ := agent.Execute(ctx, agent.Request{
    Input:     "繼續我們的對話", 
    SessionID: "existing-session-id",
})

// 存取會話狀態和中繼資料
session := response.Session
userPrefs := session.Get("user_preferences")      // 用戶定義的狀態
startTime := session.Get("session_start_time")    // 系統新增的狀態

// 會話根據 TTL 自動過期
// 過期的會話返回 ErrSessionNotFound
```

### 自訂會話 TTL

```go
// 臨時互動的短期會話
agent, _ := agent.NewBuilder().
    WithLLM(model).
    WithSessionTTL(30*time.Minute).    // 30 分鐘
    Build()

// 持久對話的長期會話
agent, _ := agent.NewBuilder().
    WithLLM(model).
    WithSessionTTL(7*24*time.Hour).    // 7 天
    Build()

// 無過期（謹慎使用）
agent, _ := agent.NewBuilder().
    WithLLM(model).
    WithSessionTTL(0).                 // 永不過期
    Build()
```

### 長時間運行的對話

```go
// 範例：旅行規劃對話
conv := agent.NewConversationWithModel(model)

// 第一次互動建立上下文
response1, _ := conv.Say(ctx, "我計劃 3 天東京行程，預算 50,000 日圓")

// 後續互動自動維護上下文
response2, _ := conv.Say(ctx, "第二天我應該參觀哪些博物館？")
// 代理記住：東京、3 天、50,000 日圓預算

response3, _ := conv.Say(ctx, "我比較喜歡現代藝術而不是傳統藝術")
// 代理現在知道：東京、3 天、預算、現代藝術偏好
```

## 錯誤處理

```go
response, err := agent.Execute(ctx, request)
if err != nil {
    switch {
    case errors.Is(err, agent.ErrInvalidInput):
        log.Println("提供的輸入無效")
    case errors.Is(err, agent.ErrSessionNotFound):
        log.Println("找不到會話")
    case errors.Is(err, agent.ErrMaxIterationsExceeded):
        log.Println("代理思考迴圈超過限制")
    case errors.Is(err, agent.ErrToolExecutionFailed):
        log.Println("工具執行失敗")
    case errors.Is(err, agent.ErrLLMCallFailed):
        log.Println("LLM 請求失敗")
    default:
        log.Printf("意外錯誤：%v", err)
    }
    return
}

// 檢查資源使用量
if response.Usage.LLMTokens.TotalTokens > 10000 {
    log.Println("偵測到高 token 使用量")
}
```

## 自訂引擎

對於進階用例，實作自訂執行引擎：

```go
type CustomEngine struct {
    // 自訂欄位
}

func (e *CustomEngine) Execute(ctx context.Context, request agent.Request, config agent.ExecutionConfig) (*agent.Response, error) {
    // 自訂執行邏輯
    // - 處理會話管理
    // - 收集上下文
    // - 使用工具呼叫 LLM
    // - 執行工具呼叫
    // - 返回結構化回應
}

// 使用自訂引擎
agent, _ := agent.NewBuilder().
    WithLLM(model).
    WithEngine(&CustomEngine{}).
    Build()
```

## 最佳實踐

### 1. 資源管理

```go
// 設定合理限制
agent, _ := agent.NewBuilder().
    WithLLM(model).
    WithMaxIterations(3).           // 防止無限迴圈
    WithMaxTokens(500).             // 控制成本
    WithTemperature(0.3).           // 更確定性
    Build()

// 監控使用量
response, _ := agent.Execute(ctx, request)
log.Printf("使用了 %d tokens", response.Usage.LLMTokens.TotalTokens)
```

### 2. 錯誤處理

```go
// 總是處理上下文取消
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := agent.Execute(ctx, request)
if err != nil {
    // 處理特定錯誤類型
    return
}
```

### 3. 會話管理

```go
// 為對話體驗重用會話
sessionID := ""

for {
    input := getUserInput()
    
    response, err := agent.Execute(ctx, agent.Request{
        Input:     input,
        SessionID: sessionID,
    })
    
    if err != nil {
        break
    }
    
    sessionID = response.SessionID // 記住下次互動
    fmt.Println(response.Output)
}
```

### 4. 工具設計

```go
// 設計工具要是冪等的並優雅地處理錯誤
type SafeTool struct{}

func (t *SafeTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    // 驗證輸入
    input, ok := params["input"].(string)
    if !ok {
        return nil, fmt.Errorf("輸入必須是字串")
    }
    
    // 尊重上下文取消
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // 安全地執行操作
    result, err := safeOperation(input)
    if err != nil {
        return nil, fmt.Errorf("操作失敗：%w", err)
    }
    
    return result, nil
}
```

## 開發狀態

**目前狀態**：核心介面、建造者模式和會話管理已實作。

**已完成功能**：
- ✅ Agent 介面和建造者模式
- ✅ 具有 TTL 和中繼資料的會話管理
- ✅ 元件配置和快取
- ✅ 上下文提供器框架
- ✅ 工具註冊表整合
- ✅ 便利函數和多輪對話
- ✅ 全面的測試覆蓋

**下一步**（引擎執行邏輯）：
1. 實作從提供器收集上下文
2. 實作 LLM 訊息建構管線
3. 實作帶工具呼叫的主執行迴圈
4. 新增工具執行編排
5. 新增使用量追蹤和錯誤處理

**測試**：所有介面和會話管理測試都通過。核心邏輯實作時會新增引擎執行測試。

## 架構

```
Agent 模組
├── 核心介面
│   ├── Agent.Execute() - 主要入口點
│   └── Engine.Execute() - 核心執行邏輯
├── 建造者模式
│   ├── 元件配置
│   ├── 會話 TTL 設定
│   └── 效能優化
├── ConfiguredEngine (✅ 已實作)
│   ├── 會話管理 (✅ 完成)
│   ├── 上下文收集 (🚧 框架就緒)
│   ├── LLM 編排 (🚧 佔位符)
│   └── 工具執行 (🚧 佔位符)
├── 便利函數 (✅ 完成)
│   ├── Chat (一次性)
│   ├── Conversation (有狀態)
│   └── MultiTurn (簡單)
└── 會話功能 (✅ 完成)
    ├── 自動 TTL 管理
    ├── 中繼資料追蹤
    └── 狀態持久性
```

**圖例**：✅ 完成、🚧 進行中、❌ 未開始

## 授權

MIT 授權 - 請參閱專案根目錄中的 LICENSE 檔案。