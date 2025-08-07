# go-agent 框架完整技術架構分析

> **Context is Everything** - 一個以上下文為核心的 Go AI Agent 框架  
> 版本：v0.0.1 | 分析日期：2025-08-05

## 📋 目錄

- [1. 框架總覽與設計理念](#1-框架總覽與設計理念)
- [2. 核心接口與類型系統](#2-核心接口與類型系統)
- [3. 建構器模式系統](#3-建構器模式系統)
- [4. 會話管理系統](#4-會話管理系統)
- [5. 上下文提供者系統](#5-上下文提供者系統)
- [6. 工具系統架構](#6-工具系統架構)
- [7. LLM 抽象層](#7-llm-抽象層)
- [8. 提示模板系統](#8-提示模板系統)
- [9. 執行引擎深度分析](#9-執行引擎深度分析)
- [10. 實際應用範例](#10-實際應用範例)
- [11. 系統流程與生命週期](#11-系統流程與生命週期)
- [12. 擴展點與未來規劃](#12-擴展點與未來規劃)

---

## 1. 框架總覽與設計理念

### 🎯 核心設計理念

go-agent 框架基於 **"Context is Everything"** 的核心理念，將上下文管理作為 AI Agent 的核心能力。框架採用模組化設計，每個組件都有明確的職責分離：

```
┌─────────────────────────────────────────────────────────────┐
│                      Agent Interface                        │
│                   Execute(Request) -> Response              │
├─────────────────────────────────────────────────────────────┤
│                     Engine (執行引擎)                        │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐  │
│  │   Session   │   Context   │    Tool     │     LLM     │  │
│  │    會話管理   │   上下文系統  │   工具系統   │   語言模型   │  │
│  └─────────────┴─────────────┴─────────────┴─────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    Prompt Template                          │
│                      提示模板系統                             │
└─────────────────────────────────────────────────────────────┘
```

### 🏗️ 架構特點

1. **接口驅動設計**：所有核心組件都定義了清晰的接口
2. **建構器模式**：提供流暢的 API 構建體驗
3. **上下文中心化**：統一的上下文收集和管理機制
4. **工具系統集成**：完整的工具定義、註冊、執行流程
5. **會話持久化**：支援 TTL 的會話狀態管理
6. **模板化提示**：靈活的提示組織系統

---

## 2. 核心接口與類型系統

### 🔌 Agent 主接口

**文件位置**: `agent/agent.go:10-13`

```go
type Agent interface {
    Execute(ctx context.Context, request Request) (*Response, error)
}
```

#### Request 結構分析

**文件位置**: `agent/agent.go:15-22`

```go
type Request struct {
    Input     string  // 用戶輸入或指令
    SessionID string  // 可選 - 為空時代理創建新會話
}
```

**功能**：
- `Input`: 承載用戶的指令或查詢
- `SessionID`: 支援會話恢復，為空時自動創建新會話

#### Response 結構分析

**文件位置**: `agent/agent.go:24-40`

```go
type Response struct {
    Output    string             // 代理的回應內容
    SessionID string             // 用於此交互的會話 ID
    Session   session.Session    // 提供對更新後會話的訪問
    Metadata  map[string]any     // 額外的回應信息
    Usage     Usage              // 使用量統計信息
}
```

**功能詳解**：
- `Output`: 代理的最終回應文本
- `SessionID`: 會話標識符，用於追蹤對話
- `Session`: 完整的會話對象，包含狀態和歷史
- `Metadata`: 擴展性字段，存儲額外信息
- `Usage`: 詳細的資源使用統計

#### Usage 追蹤系統

**文件位置**: `agent/agent.go:42-64`

```go
type Usage struct {
    LLMTokens     TokenUsage  // 語言模型令牌使用量
    ToolCalls     int         // 工具執行次數
    SessionWrites int         // 會話狀態修改次數
}

type TokenUsage struct {
    PromptTokens     int  // 輸入使用的令牌
    CompletionTokens int  // 回應生成的令牌
    TotalTokens      int  // 提示和完成令牌的總和
}
```

**統計功能**：
- **LLM 令牌追蹤**：區分輸入和輸出令牌使用量
- **工具調用統計**：記錄每次執行中的工具使用次數
- **會話寫入統計**：追蹤會話狀態的修改頻率

---

## 3. 建構器模式系統

### 🏗️ Builder 核心結構

**文件位置**: `agent/builder.go:15-18`

```go
type Builder struct {
    config EngineConfig
}
```

### 📝 配置方法詳解

#### 基礎配置方法

1. **`WithLLM(model llm.Model)`** - `builder.go:31-35`
   - 設置語言模型
   - 必需配置，框架的核心依賴

2. **`WithSessionStore(store session.SessionStore)`** - `builder.go:37-41`
   - 設置會話存儲後端
   - 支援自定義持久化策略

3. **`WithMemorySessionStore()`** - `builder.go:43-48`
   - 設置內存會話存儲（開發/測試用）
   - 便捷方法，無需額外配置

#### 工具系統配置

4. **`WithTools(tools ...tool.Tool)`** - `builder.go:50-60`
   - 批量註冊工具
   - 自動創建工具註冊表

5. **`WithToolRegistry(registry *tool.Registry)`** - `builder.go:62-66`
   - 直接設置工具註冊表
   - 用於高級工具管理場景

#### 上下文系統配置

6. **`WithContextProviders(providers ...agentcontext.Provider)`** - `builder.go:68-72`
   - 添加上下文提供者
   - 支援多個提供者組合

7. **`WithPromptTemplate(template interface{})`** - `builder.go:74-88`
   - 設置自定義提示模板
   - 支援多種類型：字符串、Template、Builder

#### 歷史管理配置

8. **`WithSessionHistory(limit int)`** - `builder.go:90-96` ⚠️ 已棄用
   - 添加會話歷史提供者
   - 建議使用 `WithHistoryLimit` 替代

9. **`WithHistoryLimit(limit int)`** - `builder.go:98-102`
   - 設置包含的歷史條目數量（0 = 禁用）
   - 更高效的歷史管理方式

10. **`WithHistoryInterceptor(interceptor HistoryInterceptor)`** - `builder.go:104-108`
    - 設置自定義歷史處理器
    - 支援歷史壓縮、摘要等高級功能

#### 執行參數配置

11. **`WithMaxIterations(max int)`** - `builder.go:110-114`
    - 設置最大思考迭代次數
    - 防止無限循環，預設 5 次

12. **`WithTemperature(temp float32)`** - `builder.go:116-120`
    - 設置 LLM 溫度參數
    - 控制回應的隨機性

13. **`WithMaxTokens(tokens int)`** - `builder.go:122-126`
    - 設置 LLM 回應的最大令牌數
    - 控制回應長度

14. **`WithSessionTTL(ttl time.Duration)`** - `builder.go:128-132`
    - 設置會話生存時間
    - 自動清理過期會話

### 🚀 便捷建構函數

#### `NewSimpleAgent(model llm.Model)` - `builder.go:166-173`
```go
func NewSimpleAgent(model llm.Model) Agent {
    agent, _ := NewBuilder().
        WithLLM(model).
        WithMemorySessionStore().
        Build()
    return agent
}
```
**用途**：創建只有 LLM 的基礎代理

#### `NewAgentWithTools(model llm.Model, tools ...tool.Tool)` - `builder.go:175-183`
```go
func NewAgentWithTools(model llm.Model, tools ...tool.Tool) Agent {
    agent, _ := NewBuilder().
        WithLLM(model).
        WithMemorySessionStore().
        WithTools(tools...).
        Build()
    return agent
}
```
**用途**：創建具有工具能力的代理

#### `NewConversationalAgent(model llm.Model, historyLimit int)` - `builder.go:185-193`
**用途**：創建維護對話歷史的代理

#### `NewFullAgent(model llm.Model, tools []tool.Tool, historyLimit int)` - `builder.go:195-208`
**用途**：創建功能完整的代理，包含所有能力

---

## 4. 會話管理系統

### 🗂️ Session 接口定義

**文件位置**: `session/session.go:5-20`

```go
type Session interface {
    // 基本信息
    ID() string
    CreatedAt() time.Time
    UpdatedAt() time.Time
    
    // 狀態管理（鍵值對）
    Get(key string) (any, bool)
    Set(key string, value any)
    Delete(key string)
    
    // 歷史管理
    AddEntry(entry Entry) error
    GetHistory(limit int) []Entry
}
```

**功能分析**：
- **基本信息**：提供會話的標識和時間戳信息
- **狀態管理**：支援任意鍵值對存儲，用於會話上下文
- **歷史管理**：統一的歷史條目管理系統

### 📊 Entry 系統架構

**文件位置**: `session/entry.go:19-26`

```go
type Entry struct {
    ID        string         `json:"id"`
    Type      EntryType      `json:"type"`
    Timestamp time.Time      `json:"timestamp"`
    Content   any            `json:"content"`
    Metadata  map[string]any `json:"metadata"`
}
```

#### EntryType 類型系統

**文件位置**: `session/entry.go:12-17`

```go
const (
    EntryTypeMessage    EntryType = "message"     // 消息條目
    EntryTypeToolCall   EntryType = "tool_call"   // 工具調用
    EntryTypeToolResult EntryType = "tool_result" // 工具結果
    EntryTypeThinking   EntryType = "thinking"    // 思考過程
)
```

#### 內容結構體系統

**1. MessageContent** - `session/entry.go:28-32`
```go
type MessageContent struct {
    Role string `json:"role"` // user/assistant/system
    Text string `json:"text"`
}
```

**2. ToolCallContent** - `session/entry.go:34-38`
```go
type ToolCallContent struct {
    Tool       string         `json:"tool"`
    Parameters map[string]any `json:"parameters"`
}
```

**3. ToolResultContent** - `session/entry.go:40-46`
```go
type ToolResultContent struct {
    Tool    string `json:"tool"`
    Success bool   `json:"success"`
    Result  any    `json:"result"`
    Error   string `json:"error"`
}
```

### 🏭 Entry 創建工廠函數

#### `NewMessageEntry(role, text string)` - `session/entry.go:48-60`
- 創建消息條目
- 自動生成 UUID 和時間戳
- 支援 user/assistant/system 角色

#### `NewToolCallEntry(tool string, params map[string]any)` - `session/entry.go:62-74`
- 創建工具調用條目
- 記錄工具名稱和參數

#### `NewToolResultEntry(tool string, result any, err error)` - `session/entry.go:76-95`
- 創建工具結果條目
- 自動處理成功/失敗狀態

### 🔍 內容提取函數

#### `GetMessageContent(entry Entry)` - `session/entry.go:97-104`
- 安全提取消息內容
- 類型檢查和轉換

#### `GetToolCallContent(entry Entry)` - `session/entry.go:106-113`
- 提取工具調用內容

#### `GetToolResultContent(entry Entry)` - `session/entry.go:115-122`
- 提取工具結果內容

---

## 5. 上下文提供者系統

### 🔌 Provider 接口設計

**文件位置**: `context/provider.go:11-14`

```go
type Provider interface {
    Provide(ctx context.Context, s session.Session) []Context
    Type() string // 提供者類型，用於模板變量映射
}
```

**設計原理**：
- `Provide`: 根據會話狀態生成上下文列表
- `Type`: 提供者類型標識，用於模板系統映射

### 📋 Context 結構分析

**文件位置**: `context/context.go:20-24`

```go
type Context struct {
    Type     string
    Content  string
    Metadata map[string]any
}
```

#### Context 類型常量

**文件位置**: `context/context.go:4-18`

```go
const (
    // 消息類型（用於對話歷史）
    TypeUser      = "user"
    TypeAssistant = "assistant"
    TypeSystem    = "system"
    TypeTool      = "tool"
    
    // 工具相關類型（用於歷史追蹤）
    TypeToolCall   = "tool_call"
    TypeToolResult = "tool_result"
    
    // 特殊類型（用於高級用例）
    TypeThinking = "thinking"
    TypeSummary  = "summary"
)
```

### 🔧 內建 Provider 實現

#### SystemPromptProvider

**文件位置**: `context/provider.go:16-38`

```go
type SystemPromptProvider struct {
    systemPrompt string
}

func (p *SystemPromptProvider) Type() string {
    return "system"
}

func (p *SystemPromptProvider) Provide(ctx context.Context, s session.Session) []Context {
    return []Context{
        {
            Type:     "system",
            Content:  p.systemPrompt,
            Metadata: map[string]any{},
        },
    }
}
```

**功能**：提供系統提示上下文

#### HistoryProvider

**文件位置**: `context/provider.go:40-121`

```go
type HistoryProvider struct {
    limit int
}

func (p *HistoryProvider) Provide(ctx context.Context, s session.Session) []Context {
    history := s.GetHistory(p.limit)
    // ... 複雜的歷史條目轉換邏輯
}
```

**核心功能**：
1. **歷史條目獲取**：從會話中獲取指定數量的歷史記錄
2. **類型轉換**：將不同的 EntryType 轉換為 Context
3. **元數據保持**：保留原始條目的所有元數據
4. **角色映射**：正確映射消息角色到上下文類型

**轉換邏輯** - `context/provider.go:70-115`：
- **EntryTypeMessage** → 轉換為 "history" 類型，原始角色存儲在元數據中
- **EntryTypeToolCall** → 轉換為 TypeToolCall，包含工具名稱和參數
- **EntryTypeToolResult** → 轉換為 TypeToolResult，包含成功狀態和結果
- **EntryTypeThinking** → 轉換為 "thinking" 類型

---

## 6. 工具系統架構

### 🔧 Tool 接口設計

**文件位置**: `tool/tool.go:5-12`

```go
type Tool interface {
    Definition() Definition              // 返回工具的定義給 LLM
    Execute(ctx context.Context, params map[string]any) (any, error)  // 執行工具
}
```

**設計哲學**：
- **定義與執行分離**：Definition 用於 LLM 理解，Execute 用於實際執行
- **上下文支援**：所有工具執行都支援取消和超時
- **靈活參數**：使用 map[string]any 支援任意參數結構

### 📋 Definition 結構系統

**文件位置**: `tool/definition.go:3-7`

```go
type Definition struct {
    Type     string   `json:"type"`     // 始終為 "function"
    Function Function `json:"function"`
}
```

#### Function 描述結構

**文件位置**: `tool/definition.go:9-14`

```go
type Function struct {
    Name        string     `json:"name"`
    Description string     `json:"description"`
    Parameters  Parameters `json:"parameters"`
}
```

#### Parameters JSON Schema 系統

**文件位置**: `tool/definition.go:16-21`

```go
type Parameters struct {
    Type       string              `json:"type"`       // "object"
    Properties map[string]Property `json:"properties"`
    Required   []string            `json:"required,omitempty"`
}
```

#### Property 屬性定義

**文件位置**: `tool/definition.go:23-35`

```go
type Property struct {
    Type        string `json:"type"`        // string/number/boolean/array/object
    Description string `json:"description"`
    
    // TODO: 未來的 JSON Schema 功能
    // - Enum 用於有效值
    // - Pattern 用於正則驗證
    // - MinLength/MaxLength
    // - Minimum/Maximum 用於數字
    // - Items 用於數組類型
    // - Properties 用於嵌套對象
}
```

### 📞 Call 調用系統

**文件位置**: `tool/definition.go:37-47`

```go
type Call struct {
    ID       string       `json:"id"`
    Function FunctionCall `json:"function"`
}

type FunctionCall struct {
    Name      string `json:"name"`
    Arguments string `json:"arguments"` // JSON 字符串
}
```

### 🗂️ Registry 註冊表系統

**文件位置**: `tool/registry.go:10-14`

```go
type Registry struct {
    mu    sync.RWMutex
    tools map[string]Tool
}
```

#### 核心方法分析

**1. Register(tool Tool)** - `tool/registry.go:23-39`
```go
func (r *Registry) Register(tool Tool) error {
    def := tool.Definition()
    if def.Function.Name == "" {
        return fmt.Errorf("tool must have a name")
    }
    
    if _, exists := r.tools[def.Function.Name]; exists {
        return fmt.Errorf("tool %s already registered", def.Function.Name)
    }
    
    r.tools[def.Function.Name] = tool
    return nil
}
```

**功能**：
- 驗證工具名稱非空
- 檢查重複註冊
- 線程安全的工具存儲

**2. Execute(ctx context.Context, call Call)** - `tool/registry.go:41-69`
```go
func (r *Registry) Execute(ctx context.Context, call Call) (any, error) {
    // 1. 查找工具
    tool, exists := r.tools[call.Function.Name]
    
    // 2. 解析 JSON 參數
    var params map[string]any
    if err := json.Unmarshal([]byte(call.Function.Arguments), &params); err \!= nil {
        return nil, fmt.Errorf("failed to parse arguments: %w", err)
    }
    
    // 3. 執行工具
    result, err := tool.Execute(ctx, params)
    
    return result, nil
}
```

**執行流程**：
1. 工具查找和驗證
2. JSON 參數解析
3. 工具執行與錯誤處理

**3. GetDefinitions()** - `tool/registry.go:71-81`
- 返回所有已註冊工具的定義
- 用於向 LLM 提供可用工具列表

---

## 7. LLM 抽象層

### 🤖 Model 接口設計

**文件位置**: `llm/model.go:5-12`

```go
type Model interface {
    Complete(ctx context.Context, request Request) (*Response, error)
    
    // TODO: 未來實現流式處理
    // Stream(ctx context.Context, request Request) (<-chan StreamEvent, error)
}
```

**設計原理**：
- **同步完成**：當前實現同步調用模式
- **上下文支援**：支援取消和超時控制
- **流式預留**：為未來流式響應預留接口

### 📝 Request 結構分析

**文件位置**: `llm/types.go:5-21`

```go
type Request struct {
    Messages []Message
    
    // 可選模型參數
    Temperature *float32 `json:"temperature,omitempty"`
    MaxTokens   *int     `json:"max_tokens,omitempty"`
    
    // 可選工具定義
    Tools []tool.Definition `json:"tools,omitempty"`
}
```

**字段功能**：
- `Messages`: 對話消息列表
- `Temperature`: 控制回應隨機性（可選）
- `MaxTokens`: 最大回應長度（可選）
- `Tools`: 可用工具列表（可選）

### 💬 Message 結構系統

**文件位置**: `llm/types.go:23-32`

```go
type Message struct {
    Role    string `json:"role"`    // system/user/assistant/tool
    Content string `json:"content"`
    
    // 工具相關消息
    Name       string      `json:"name,omitempty"`         // 工具名稱
    ToolCallID string      `json:"tool_call_id,omitempty"` // 工具回應用
    ToolCalls  []tool.Call `json:"tool_calls,omitempty"`   // 助手消息中的工具調用
}
```

**角色系統**：
- **system**: 系統提示和指令
- **user**: 用戶輸入
- **assistant**: AI 助手回應
- **tool**: 工具執行結果

### 📊 Response 結構分析

**文件位置**: `llm/types.go:34-42`

```go
type Response struct {
    Content   string      `json:"content"`
    ToolCalls []tool.Call `json:"tool_calls,omitempty"`
    
    // 基本元數據
    Usage        Usage  `json:"usage"`
    FinishReason string `json:"finish_reason"` // stop/length/tool_calls
}
```

**完成原因**：
- **stop**: 自然結束
- **length**: 達到最大長度限制
- **tool_calls**: 需要工具調用

### 🔌 OpenAI 實現

**文件位置**: `llm/openai/client.go:12-16`

```go
type Client struct {
    client *openai.Client
    model  string
}
```

#### 核心方法：Complete

**文件位置**: `llm/openai/client.go:31-44`

```go
func (c *Client) Complete(ctx context.Context, request llm.Request) (*llm.Response, error) {
    // 1. 轉換請求格式
    openaiReq := c.toOpenAIRequest(request)
    
    // 2. 調用 API
    resp, err := c.client.CreateChatCompletion(ctx, openaiReq)
    
    // 3. 轉換回應格式
    return c.fromOpenAIResponse(resp), nil
}
```

**轉換邏輯**：
- 內部格式 → OpenAI API 格式
- OpenAI 回應 → 內部格式
- 工具調用的特殊處理

---

## 8. 提示模板系統

### 📝 Template 接口設計

**文件位置**: `prompt/template.go:14-27`

```go
type Template interface {
    // 將提供者的上下文轉換為 LLM 消息
    Render(ctx context.Context, providers []agentcontext.Provider, 
           session session.Session, userInput string) ([]llm.Message, error)
    
    // 返回模板中使用的變量列表
    Variables() []string
    
    // 返回模板結構的人類可讀描述
    Explain() string
    
    // 返回原始模板表示
    String() string
}
```

### 🔧 Builder 流暢 API

**文件位置**: `prompt/builder.go:4-24`

```go
type Builder interface {
    // 核心變量，用於常見用例
    System() Builder
    History() Builder
    UserInput() Builder
    
    // 自定義提供者變量
    Provider(providerType string) Builder
    NamedProvider(providerType, name string) Builder
    
    // 靜態文本內容
    Text(content string) Builder
    Line(content string) Builder // 帶換行的文本
    
    // 便捷方法
    DefaultFlow() Builder // 添加標準流程：system -> history -> user_input
    Separator() Builder   // 添加視覺分隔符
    
    // 構建模板
    Build() Template
}
```

#### 核心方法實現

**1. DefaultFlow()** - `prompt/builder.go:85-92`
```go
func (b *TemplateBuilder) DefaultFlow() Builder {
    return b.System().
        History().
        Text("Context information:\n").
        Provider("context_providers").
        UserInput()
}
```

**標準流程**：
1. 系統提示
2. 對話歷史
3. 上下文信息標題
4. 上下文提供者內容
5. 用戶輸入

**2. NamedProvider()** - `prompt/builder.go:58-67`
```go
func (b *TemplateBuilder) NamedProvider(providerType, name string) Builder {
    section := section{
        Type:     "variable",
        Content:  providerType,
        Metadata: map[string]string{"name": name},
    }
    b.sections = append(b.sections, section)
    return b
}
```

**命名提供者**：支援同類型多個提供者的區分

### 🔍 Parser 解析系統

**文件位置**: `prompt/parser.go:17-77`

支援兩種變量格式：
- `{{provider_type}}` - 類型引用
- `{{provider_type:name}}` - 命名引用

**解析流程**：
1. 正則表達式匹配變量
2. 提取類型和可選名稱
3. 構建 section 結構
4. 處理文本內容

### 🎯 渲染系統

#### 核心渲染邏輯

**文件位置**: `prompt/template.go:42-78`

```go
func (t *promptTemplate) Render(ctx context.Context, providers []agentcontext.Provider, 
                               s session.Session, userInput string) ([]llm.Message, error) {
    var messages []llm.Message
    
    for _, section := range t.sections {
        switch section.Type {
        case "variable":
            if section.Content == "user_input" {
                // 特殊處理用戶輸入
                messages = append(messages, llm.Message{
                    Role: "user", Content: userInput,
                })
            } else {
                // 收集此變量的上下文
                contexts := t.gatherContexts(section, providers, ctx, s)
                // 渲染上下文為消息
                msgs := t.renderContexts(section.Content, contexts)
                messages = append(messages, msgs...)
            }
            
        case "text":
            // 靜態文本成為系統消息
            messages = append(messages, llm.Message{
                Role: "system", Content: strings.TrimSpace(section.Content),
            })
        }
    }
    
    return messages, nil
}
```

#### 上下文收集邏輯

**文件位置**: `prompt/template.go:80-108`

```go
func (t *promptTemplate) gatherContexts(section section, providers []agentcontext.Provider, 
                                       ctx context.Context, s session.Session) []agentcontext.Context {
    var contexts []agentcontext.Context
    
    providerType := section.Content
    providerName := section.Metadata["name"]
    
    for _, provider := range providers {
        // 檢查提供者類型是否匹配
        if provider.Type() == providerType {
            // 如果需要特定名稱，檢查提供者是否支援
            if providerName \!= "" {
                if namedProvider, ok := provider.(NamedProvider); ok {
                    if namedProvider.Name() \!= providerName {
                        continue
                    }
                }
            }
            
            // 從此提供者收集上下文
            providerContexts := provider.Provide(ctx, s)
            contexts = append(contexts, providerContexts...)
        }
    }
    
    return contexts
}
```

#### 專門化渲染器

**1. renderSystemContexts** - `prompt/template.go:130-147`
- 將系統上下文組合為單個系統消息
- 使用 `\n\n` 連接多個內容

**2. renderHistoryContexts** - `prompt/template.go:149-171`
- 保持原始消息角色
- 從元數據中恢復 `original_role`
- 直接轉換為對話消息

**3. renderCustomContexts** - `prompt/template.go:173-190`
- 自定義變量預設為系統角色
- 使用 `\n` 連接多個內容

---

## 9. 執行引擎深度分析

### ⚙️ Engine 結構設計

**文件位置**: `agent/engine.go:21-44`

```go
type engine struct {
    // 核心組件
    model            llm.Model
    sessionStore     session.SessionStore
    toolRegistry     *tool.Registry
    contextProviders []agentcontext.Provider
    
    // 提示模板
    promptTemplate prompt.Template
    
    // 配置
    maxIterations int
    temperature   *float32
    maxTokens     *int
    
    // 歷史配置
    historyLimit       int
    historyInterceptor HistoryInterceptor
    
    // 會話配置
    sessionTTL       time.Duration
    cachedCreateOpts []session.CreateOption
}
```

### 🚀 執行流程深度解析

#### Execute 主流程

**文件位置**: `agent/engine.go:100-136`

```go
func (e *engine) Execute(ctx context.Context, request Request) (*Response, error) {
    // 步驟 1: 會話管理
    agentSession, err := e.handleSession(ctx, request)
    
    // 步驟 2: 上下文收集
    contexts, err := e.gatherContexts(ctx, request, agentSession)
    
    // 步驟 3: 主執行循環
    result, err := e.executeIterations(ctx, request, contexts, agentSession)
    
    // 步驟 4: 最終化回應
    response := &Response{
        Output: result.FinalOutput, SessionID: result.SessionID,
        Session: result.Session, Metadata: result.Metadata, Usage: result.Usage,
    }
    
    return response, nil
}
```

#### 會話處理邏輯

**文件位置**: `agent/engine.go:138-158`

```go
func (e *engine) handleSession(ctx context.Context, request Request) (session.Session, error) {
    if request.SessionID == "" {
        // 使用預緩存選項創建新會話
        newSession := e.sessionStore.Create(ctx, e.cachedCreateOpts...)
        
        // 添加基於請求的動態元數據
        newSession.Set("initial_input_length", len(request.Input))
        newSession.Set("session_start_time", time.Now().Format(time.RFC3339))
        
        return newSession, nil
    }
    
    // 加載現有會話
    existingSession, err := e.sessionStore.Get(ctx, request.SessionID)
    return existingSession, err
}
```

**優化特性**：
- **預緩存選項**：避免每次創建會話時的重複計算
- **動態元數據**：根據請求內容添加相關信息
- **錯誤處理**：會話不存在時返回特定錯誤

#### 上下文收集系統

**文件位置**: `agent/engine.go:160-188`

```go
func (e *engine) gatherContexts(ctx context.Context, request Request, agentSession session.Session) ([]agentcontext.Context, error) {
    var allContexts []agentcontext.Context
    
    // 1. 從提供者收集上下文（非歷史）
    for _, provider := range e.contextProviders {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }
        
        contexts := provider.Provide(ctx, agentSession)
        allContexts = append(allContexts, contexts...)
    }
    
    // 2. 如果啟用，添加歷史上下文
    if e.historyLimit > 0 {
        historyContexts, err := e.extractHistoryContexts(ctx, agentSession)
        allContexts = append(allContexts, historyContexts...)
    }
    
    return allContexts, nil
}
```

**特性**：
- **取消支援**：每個提供者調用前檢查上下文取消
- **歷史分離**：將歷史處理與普通上下文提供者分離
- **錯誤傳播**：任何步驟失敗都會終止收集過程

#### 歷史提取與處理

**文件位置**: `agent/engine.go:190-211`

```go
func (e *engine) extractHistoryContexts(ctx context.Context, agentSession session.Session) ([]agentcontext.Context, error) {
    // 1. 從會話獲取原始歷史條目
    entries := agentSession.GetHistory(e.historyLimit)
    
    // 2. 如果配置了歷史攔截器，應用它
    if e.historyInterceptor \!= nil {
        processedEntries, err := e.historyInterceptor.ProcessHistory(ctx, entries, e.model)
        entries = processedEntries
    }
    
    // 3. 將條目轉換為上下文
    contexts := e.convertEntriesToContexts(entries)
    
    return contexts, nil
}
```

**高級功能**：
- **歷史攔截器**：支援歷史壓縮、摘要、過濾等
- **模型感知**：攔截器可以使用相同的 LLM 進行處理
- **靈活轉換**：統一的條目到上下文轉換邏輯

### 🔄 迭代執行系統

#### 主迭代循環

**文件位置**: `agent/engine.go:278-414`

```go
func (e *engine) executeIterations(ctx context.Context, request Request, contexts []agentcontext.Context, agentSession session.Session) (*ExecutionResult, error) {
    var totalUsage Usage
    var conversationMessages []llm.Message
    var finalResponse string
    
    // 步驟 1: 從上下文和用戶輸入構建初始消息
    messages := e.buildLLMMessages(contexts, request)
    conversationMessages = append(conversationMessages, messages...)
    
    // 步驟 2: 主迭代循環
    for iteration := 0; iteration < e.maxIterations; iteration++ {
        // 詳細的執行日誌
        fmt.Printf("🤖 [Iteration %d] Agent thinking...\n", iteration+1)
        
        // 準備 LLM 請求
        tools := e.toolRegistry.GetDefinitions()
        llmRequest := llm.Request{
            Messages: conversationMessages, Tools: tools,
            Temperature: e.temperature, MaxTokens: e.maxTokens,
        }
        
        // 調用 LLM
        response, err := e.model.Complete(ctx, llmRequest)
        
        // 更新使用統計
        totalUsage.LLMTokens.PromptTokens += response.Usage.PromptTokens
        // ... 其他統計更新
        
        // 處理 LLM 回應
        if len(response.ToolCalls) > 0 {
            // 工具調用分支
            // ... 工具執行邏輯
            continue
        }
        
        // 完成分支
        if response.FinishReason == "stop" || response.FinishReason == "length" {
            finalResponse = response.Content
            break
        }
    }
    
    // 步驟 3: 保存對話到會話
    e.saveConversationToSession(agentSession, request.Input, finalResponse)
    
    return &ExecutionResult{...}, nil
}
```

**核心特性**：
1. **詳細日誌**：每個步驟都有清晰的日誌輸出
2. **使用統計**：精確追蹤令牌和工具使用
3. **工具循環**：支援多輪工具調用
4. **自然終止**：多種完成條件檢查
5. **會話保存**：自動保存對話歷史

#### 工具執行系統

**文件位置**: `agent/engine.go:416-450`

```go
func (e *engine) executeTools(ctx context.Context, toolCalls []tool.Call) []ToolResult {
    var results []ToolResult
    
    for i, call := range toolCalls {
        fmt.Printf("  🛠️  [Tool %d/%d] Calling: %s\n", i+1, len(toolCalls), call.Function.Name)
        fmt.Printf("  📋 Arguments: %s\n", call.Function.Arguments)
        
        // 使用註冊表執行工具
        result, err := e.toolRegistry.Execute(ctx, call)
        
        if err \!= nil {
            fmt.Printf("  ❌ Tool execution failed: %v\n", err)
        } else {
            // 截斷結果以便顯示
            resultStr := fmt.Sprintf("%v", result)
            if len(resultStr) > 200 {
                resultStr = resultStr[:200] + "..."
            }
            fmt.Printf("  ✅ Tool result: %s\n", resultStr)
        }
        
        results = append(results, ToolResult{Call: call, Result: result, Error: err})
    }
    
    return results
}
```

**執行特性**：
- **並發安全**：使用工具註冊表的線程安全執行
- **詳細日誌**：每個工具調用的完整追蹤
- **錯誤處理**：失敗不會終止其他工具執行
- **結果截斷**：避免過長輸出影響日誌可讀性

### 🏗️ 消息構建系統

#### 智能消息構建

**文件位置**: `agent/engine.go:452-489`

```go
func (e *engine) buildLLMMessages(contexts []agentcontext.Context, request Request) []llm.Message {
    // 如果可用，使用 PromptTemplate
    if e.promptTemplate \!= nil {
        // 將上下文轉換為提供者以供模板渲染
        providers := e.contextsToProviders(contexts)
        
        // 使用模板渲染消息
        messages, err := e.promptTemplate.Render(context.Background(), providers, nil, request.Input)
        if err == nil {
            return messages
        }
        
        // 如果模板失敗，回退到硬編碼格式
        fmt.Printf("Warning: PromptTemplate render failed: %v, falling back to hardcoded format\n", err)
    }
    
    // 回退：硬編碼格式（原始邏輯）
    var messages []llm.Message
    
    // 步驟 1: 添加系統消息（硬編碼格式）
    systemMessage := e.buildSystemMessage(contexts)
    messages = append(messages, llm.Message{Role: "system", Content: systemMessage})
    
    // 步驟 2: 從上下文添加對話歷史
    historyMessages := e.buildHistoryMessages(contexts)
    messages = append(messages, historyMessages...)
    
    // 步驟 3: 添加當前用戶輸入
    messages = append(messages, llm.Message{Role: "user", Content: request.Input})
    
    return messages
}
```

**智能特性**：
- **模板優先**：優先使用配置的提示模板
- **優雅回退**：模板失敗時自動使用內建格式
- **兼容性**：確保所有配置都能正常工作

#### 硬編碼系統提示構建

**文件位置**: `agent/engine.go:513-561`

```go
func (e *engine) buildSystemMessage(contexts []agentcontext.Context) string {
    systemPrompt := `You are a helpful AI agent. Follow these guidelines:

1. Be concise and helpful in your responses
2. Use available tools when needed to provide accurate information
3. If you need to use tools, explain what you're doing
4. Always strive to be accurate and truthful

`

    // 檢查是否有歷史上下文並添加歷史說明
    if e.hasHistoryContexts(contexts) {
        systemPrompt += `Note on Conversation History:
The conversation history provided may have been compressed or summarized to save space.
Key information and context have been preserved, but some details might be condensed.
Please use this history as reference for maintaining conversation continuity and context.

`
    }
    
    // 添加系統上下文（非歷史上下文）
    var systemContexts []string
    for _, ctx := range contexts {
        // 跳過歷史類型上下文，因為它們將作為單獨的消息添加
        if ctx.Type == agentcontext.TypeUser || ctx.Type == agentcontext.TypeAssistant ||
           ctx.Type == agentcontext.TypeToolCall || ctx.Type == agentcontext.TypeToolResult {
            continue
        }
        
        if ctx.Content \!= "" {
            systemContexts = append(systemContexts, ctx.Content)
        }
    }
    
    // 將額外的上下文信息附加到系統提示
    if len(systemContexts) > 0 {
        systemPrompt += "Additional Context:\n"
        for i, ctxContent := range systemContexts {
            systemPrompt += fmt.Sprintf("%d. %s\n", i+1, ctxContent)
        }
    }
    
    return systemPrompt
}
```

**構建邏輯**：
1. **基礎指導原則**：標準的 AI 助手行為指南
2. **歷史說明**：當檢測到歷史上下文時添加特殊說明
3. **上下文分離**：區分系統上下文和對話歷史
4. **結構化展示**：編號列出所有額外上下文

### 💾 會話保存系統

**文件位置**: `agent/engine.go:643-658`

```go
func (e *engine) saveConversationToSession(agentSession session.Session, userInput, agentResponse string) error {
    // 添加用戶消息條目
    userEntry := session.NewMessageEntry("user", userInput)
    agentSession.AddEntry(userEntry)
    
    // 添加助手回應條目
    assistantEntry := session.NewMessageEntry("assistant", agentResponse)
    agentSession.AddEntry(assistantEntry)
    
    // 更新會話元數據
    agentSession.Set("last_interaction", time.Now().Format(time.RFC3339))
    agentSession.Set("total_messages", len(agentSession.GetHistory(1000)))
    
    return nil
}
```

**保存特性**：
- **標準化條目**：使用統一的消息條目格式
- **元數據更新**：追蹤最後交互時間和消息總數
- **歷史積累**：每次交互都會增加會話歷史

---

## 10. 實際應用範例

### 🎓 學術研究助手 (academic-research-assistant)

**文件位置**: `examples/academic-research-assistant/main.go`

#### 核心功能分析

**1. ArXiv 搜索工具**
```go
type ArxivSearchTool struct{}

func (a *ArxivSearchTool) Definition() tool.Definition {
    return tool.Definition{
        Type: "function",
        Function: tool.Function{
            Name: "search_arxiv",
            Description: "Search academic papers on arXiv by keywords, author, or subject",
            Parameters: tool.Parameters{
                Type: "object",
                Properties: map[string]tool.Property{
                    "query":      {Type: "string", Description: "Search keywords or phrases"},
                    "max_results": {Type: "number", Description: "Maximum papers to return (1-20)"},
                    "category":   {Type: "string", Description: "arXiv category filter (optional)"},
                },
                Required: []string{"query"},
            },
        },
    }
}
```

**特性**：
- 支援關鍵詞、作者、主題搜索
- 可配置結果數量限制
- 支援 arXiv 分類篩選

**2. 論文詳情獲取工具**
```go
type ArxivDetailTool struct{}

func (a *ArxivDetailTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    arxivID := params["arxiv_id"].(string)
    
    // 從 arXiv API 獲取詳細信息
    url := fmt.Sprintf("http://export.arxiv.org/api/query?id_list=%s", arxivID)
    
    // ... HTTP 請求和 XML 解析邏輯
    
    return DetailedPaper{
        ID: paper.ID, Title: paper.Title, Authors: paper.Authors,
        Abstract: paper.Abstract, Published: paper.Published,
        Categories: paper.Categories, PDFLink: paper.PDFLink,
    }, nil
}
```

#### 工作流程設計

**探索工作流程** (`runExploreWorkflow`):
1. **初始搜索**：使用關鍵詞搜索相關論文
2. **結果分析**：分析搜索結果，識別重要論文
3. **深度研究**：獲取關鍵論文的詳細信息
4. **摘要生成**：生成研究領域的概述

**代理配置**：
```go
researchAgent, err := agent.NewBuilder().
    WithLLM(client).
    WithMemorySessionStore().
    WithTools(&ArxivSearchTool{}, &ArxivDetailTool{}).
    WithMaxIterations(10).
    WithPromptTemplate(researchTemplate).
    Build()
```

### 🛠️ 多工具演示 (multi-tool-demo)

**文件位置**: `examples/multi-tool-demo/main.go`

#### 工具生態系統

**1. 天氣查詢工具**
```go
type WeatherTool struct{}

func (w *WeatherTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    city := params["city"].(string)
    units, _ := params["units"].(string)
    
    // 模擬 API 延遲
    time.Sleep(500 * time.Millisecond)
    
    // 基於城市的模擬天氣數據
    switch strings.ToLower(city) {
    case "tokyo", "東京":
        return WeatherInfo{City: city, Temperature: 22, Condition: "Sunny", Humidity: 65}, nil
    // ... 更多城市
    }
}
```

**2. 計算器工具**
```go
type CalculatorTool struct{}

func (c *CalculatorTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    expression := params["expression"].(string)
    
    // 使用 Go 的運算能力進行計算
    switch {
    case strings.Contains(expression, "sqrt"):
        // 平方根計算
    case strings.Contains(expression, "^"):
        // 冪運算
    case strings.Contains(expression, "sin"), strings.Contains(expression, "cos"):
        // 三角函數
    default:
        // 基本算術運算
    }
}
```

**3. 時間工具**
```go
type TimeTool struct{}

func (t *TimeTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    timezone, _ := params["timezone"].(string)
    
    var loc *time.Location
    if timezone \!= "" {
        loc, _ = time.LoadLocation(timezone)
    } else {
        loc = time.UTC
    }
    
    now := time.Now().In(loc)
    return TimeInfo{
        CurrentTime: now.Format("2006-01-02 15:04:05"),
        Timezone: loc.String(),
        Weekday: now.Weekday().String(),
    }, nil
}
```

#### 交互式對話系統

**對話循環實現**：
```go
func runInteractiveMode(agent agent.Agent) {
    scanner := bufio.NewScanner(os.Stdin)
    var sessionID string
    
    fmt.Println("🤖 Multi-Tool Demo Agent")
    fmt.Println("Available tools: weather, calculator, time")
    
    for {
        fmt.Print("\n💬 You: ")
        if \!scanner.Scan() {
            break
        }
        
        input := strings.TrimSpace(scanner.Text())
        if input == "quit" || input == "exit" {
            break
        }
        
        // 執行代理請求
        response, err := agent.Execute(context.Background(), agent.Request{
            Input: input, SessionID: sessionID,
        })
        
        if err \!= nil {
            fmt.Printf("❌ Error: %v\n", err)
            continue
        }
        
        // 更新會話 ID 以維護上下文
        sessionID = response.SessionID
        
        // 顯示回應和統計
        fmt.Printf("\n🤖 Assistant: %s\n", response.Output)
        fmt.Printf("📊 Usage: %d tokens, %d tools called\n", 
                  response.Usage.LLMTokens.TotalTokens, response.Usage.ToolCalls)
    }
}
```

**特性**：
- **會話持續**：在整個對話中維護會話上下文
- **多工具集成**：無縫切換不同類型的工具
- **使用統計**：即時顯示資源使用情況
- **優雅退出**：支援 quit/exit 命令

### 🎯 使用模式分析

#### 模式 1：專門化代理
```go
// 專門用於特定領域的代理
agent := agent.NewBuilder().
    WithLLM(model).
    WithTools(domainSpecificTools...).
    WithContextProviders(domainContextProvider).
    WithPromptTemplate(specializedTemplate).
    Build()
```

#### 模式 2：通用多功能代理
```go
// 支援多種功能的通用代理
agent := agent.NewBuilder().
    WithLLM(model).
    WithTools(weatherTool, calculatorTool, timeTool, webTool).
    WithHistoryLimit(10).
    WithMaxIterations(5).
    Build()
```

#### 模式 3：對話式代理
```go
// 重點關注對話連續性的代理
agent := agent.NewBuilder().
    WithLLM(model).
    WithHistoryLimit(20).
    WithHistoryInterceptor(conversationSummarizer).
    WithSessionTTL(24 * time.Hour).
    Build()
```

---

## 11. 系統流程與生命週期

### 🔄 完整執行流程圖

```
用戶請求 (Request)
      ↓
┌─────────────────────┐
│   1. 會話管理         │
│   - 創建/加載會話      │
│   - 設置元數據        │
└─────────────────────┘
      ↓
┌─────────────────────┐
│   2. 上下文收集       │
│   - 調用所有提供者    │
│   - 提取歷史上下文    │
│   - 應用歷史攔截器    │
└─────────────────────┘
      ↓
┌─────────────────────┐
│   3. 消息構建        │
│   - 使用提示模板      │
│   - 或硬編碼格式      │
└─────────────────────┘
      ↓
┌─────────────────────┐
│   4. 迭代執行循環     │ ← ┐
│   - LLM 調用         │   │
│   - 工具執行         │   │
│   - 結果處理         │   │
└─────────────────────┘   │
      ↓                   │
    是否需要                │
    更多迭代？ ─────────────┘
      ↓ 否
┌─────────────────────┐
│   5. 會話保存        │
│   - 保存對話記錄      │
│   - 更新元數據        │
└─────────────────────┘
      ↓
   回應 (Response)
```

### 🏗️ 會話生命週期

#### 創建階段
```go
// 新會話創建流程
session := store.Create(ctx, 
    session.WithTTL(24*time.Hour),
    session.WithMetadata("created_by", "agent"),
    session.WithMetadata("agent_version", version),
)

// 添加初始元數據
session.Set("initial_input_length", len(request.Input))
session.Set("session_start_time", time.Now().Format(time.RFC3339))
```

#### 使用階段
```go
// 每次交互的會話更新
session.AddEntry(userEntry)          // 添加用戶輸入
session.AddEntry(assistantEntry)     // 添加助手回應
session.Set("last_interaction", now) // 更新最後交互時間
session.Set("total_messages", count) // 更新消息總數
```

#### 清理階段
- TTL 到期自動清理
- 存儲實現的垃圾回收機制
- 會話元數據的持久化選項

### 🎯 上下文收集流程

#### 階段 1：提供者調用
```go
for _, provider := range e.contextProviders {
    // 取消檢查
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // 調用提供者
    contexts := provider.Provide(ctx, agentSession)
    allContexts = append(allContexts, contexts...)
}
```

#### 階段 2：歷史處理
```go
if e.historyLimit > 0 {
    // 1. 獲取原始歷史
    entries := agentSession.GetHistory(e.historyLimit)
    
    // 2. 應用攔截器
    if e.historyInterceptor \!= nil {
        entries, err = e.historyInterceptor.ProcessHistory(ctx, entries, e.model)
    }
    
    // 3. 轉換為上下文
    historyContexts := e.convertEntriesToContexts(entries)
    allContexts = append(allContexts, historyContexts...)
}
```

#### 階段 3：上下文組織
- 按類型分組上下文
- 應用提示模板邏輯
- 生成 LLM 消息序列

### ⚙️ 工具調用循環

#### 循環結構
```
LLM 調用
    ↓
檢查 ToolCalls
    ↓ 有工具調用
執行所有工具
    ↓
格式化工具結果
    ↓
添加到對話歷史
    ↓
回到 LLM 調用
```

#### 詳細流程
1. **工具檢測**：LLM 回應包含 `tool_calls`
2. **並行執行**：所有工具調用並行處理
3. **結果格式化**：轉換為 LLM 可理解的格式
4. **對話更新**：添加助手消息和工具結果
5. **繼續迭代**：重新調用 LLM 處理工具結果

### 🔄 迭代終止條件

#### 自然終止
```go
if response.FinishReason == "stop" || response.FinishReason == "length" {
    fmt.Printf("✅ Agent completed task (reason: %s)\n", response.FinishReason)
    finalResponse = response.Content
    break
}
```

#### 最大迭代限制
```go
if finalResponse == "" {
    return nil, ErrMaxIterationsExceeded
}
```

#### 取消處理
```go
select {
case <-ctx.Done():
    return nil, ctx.Err()
default:
}
```

---

## 12. 擴展點與未來規劃

### 🔌 當前擴展點

#### 1. 自定義上下文提供者
```go
type CustomProvider struct {
    // 自定義字段
}

func (p *CustomProvider) Type() string {
    return "custom_type"
}

func (p *CustomProvider) Provide(ctx context.Context, s session.Session) []Context {
    // 自定義上下文收集邏輯
    return contexts
}
```

#### 2. 歷史攔截器
```go
type HistoryCompressor struct {
    // 配置字段
}

func (c *HistoryCompressor) ProcessHistory(ctx context.Context, entries []session.Entry, llm llm.Model) ([]session.Entry, error) {
    // 歷史壓縮、摘要或過濾邏輯
    return processedEntries, nil
}
```

#### 3. 自定義工具
```go
type CustomTool struct {
    // 工具特定配置
}

func (t *CustomTool) Definition() tool.Definition {
    // 工具定義邏輯
}

func (t *CustomTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    // 工具執行邏輯
}
```

#### 4. 自定義會話存儲
```go
type CustomSessionStore struct {
    // 存儲後端配置
}

func (s *CustomSessionStore) Create(ctx context.Context, opts ...session.CreateOption) session.Session {
    // 自定義會話創建邏輯
}

func (s *CustomSessionStore) Get(ctx context.Context, id string) (session.Session, error) {
    // 自定義會話檢索邏輯
}
```

#### 5. 自定義 LLM 提供者
```go
type CustomLLM struct {
    // LLM 特定配置
}

func (l *CustomLLM) Complete(ctx context.Context, request llm.Request) (*llm.Response, error) {
    // 自定義 LLM 調用邏輯
}
```

### 📋 TODO 項目分析

#### 工具系統增強 (tool/tool.go:14-18)
```go
// TODO: 未來增強功能
// - 添加輸入參數驗證
// - 支援輸出模式驗證（可選）
// - 異步執行支援
// - 工具中間件/攔截器
```

**影響分析**：
- **參數驗證**：提高工具調用的可靠性
- **輸出驗證**：確保工具回應符合預期格式
- **異步支援**：支援長時間運行的工具
- **中間件**：添加日誌、監控、權限檢查等

#### 工具定義增強 (tool/definition.go:28-35)
```go
// TODO: 未來的 JSON Schema 功能
// - Enum 用於有效值
// - Pattern 用於正則驗證
// - MinLength/MaxLength
// - Minimum/Maximum 用於數字
// - Items 用於數組類型
// - Properties 用於嵌套對象
```

**影響分析**：
- 更強大的參數驗證能力
- 更豐富的工具定義表達力
- 更好的 LLM 理解和調用準確性

#### LLM 系統增強 (llm/types.go:16-21, llm/model.go:10-12)
```go
// TODO: 未來增強功能
// - 模型覆蓋
// - TopP, StopSequences
// - ToolChoice 策略
// - 用戶標識用於速率限制

// TODO: 未來實現流式處理
// Stream(ctx context.Context, request Request) (<-chan StreamEvent, error)
```

**影響分析**：
- **流式處理**：實時回應用戶，改善用戶體驗
- **高級參數**：更精細的生成控制
- **速率限制**：多用戶支援和資源管理

#### 工具註冊表增強 (tool/registry.go:63-67)
```go
// TODO: 未來增強功能
// - 根據定義的模式驗證輸出
// - 添加執行指標/日誌
// - 支援中間件/攔截器
```

**影響分析**：
- **輸出驗證**：確保工具回應質量
- **監控指標**：性能分析和故障排除
- **中間件支援**：橫切關注點的統一處理

### 🚀 架構改進建議

#### 1. 性能優化
- **並發工具執行**：當前是順序執行，可以改為並發
- **上下文緩存**：緩存不變的上下文內容
- **模板編譯**：預編譯複雜的提示模板
- **連接池**：LLM API 調用的連接池管理

#### 2. 可觀測性
- **結構化日誌**：使用 structured logging 替代 printf
- **指標收集**：集成 Prometheus 指標
- **分佈式追蹤**：添加 OpenTelemetry 支援
- **健康檢查**：組件健康狀態監控

#### 3. 錯誤處理
- **錯誤分類**：區分可重試和不可重試錯誤
- **重試機制**：指數退避的重試策略
- **熔斷器**：防止依賴服務故障傳播
- **降級策略**：關鍵路徑的降級方案

#### 4. 安全性
- **輸入驗證**：所有外部輸入的嚴格驗證
- **權限控制**：工具執行的權限檢查
- **敏感數據**：PII 和敏感信息的處理
- **審計日誌**：所有操作的審計追蹤

#### 5. 擴展性
- **插件系統**：動態加載自定義組件
- **配置管理**：統一的配置管理系統
- **服務發現**：分佈式部署的服務發現
- **水平擴展**：支援多實例部署

---

## 📊 總結

### 🎯 框架優勢

1. **模組化設計**：清晰的接口定義和職責分離
2. **靈活配置**：建構器模式提供豐富的配置選項
3. **擴展友好**：所有核心組件都支援自定義實現
4. **會話管理**：完整的會話生命週期和狀態管理
5. **工具集成**：無縫的工具定義、註冊、執行流程
6. **模板系統**：靈活的提示組織和渲染機制

### 📈 成熟度評估

#### ✅ 已實現功能
- 核心代理接口和執行引擎
- 完整的建構器模式
- 會話管理和歷史追蹤
- 工具系統和註冊表
- LLM 抽象和 OpenAI 集成
- 提示模板系統
- 上下文提供者機制

#### 🚧 需要完善的領域
- 錯誤處理和重試機制
- 性能監控和指標收集
- 並發安全性驗證
- 更多 LLM 提供者支援
- 流式處理實現
- 完整的測試覆蓋

#### 🔮 未來發展方向
- 分佈式部署支援
- 插件生態系統
- 視覺化管理界面
- 多模態內容支援
- 高級 AI 功能集成

### 🎉 發布就緒性評估

**當前版本 (v0.0.1) 適合以下場景**：
- ✅ 原型開發和概念驗證
- ✅ 教育和學習目的
- ✅ 小規模應用開發
- ✅ 框架能力探索

**建議在以下方面完善後考慮生產使用**：
- 🔧 完善錯誤處理機制
- 🔧 添加全面的單元測試
- 🔧 實現性能監控
- 🔧 完善文檔和範例
- 🔧 安全性加固

---

**文檔版本**: 1.0  
**最後更新**: 2025-08-05  
**分析深度**: 函數級別完整分析  
**總代碼行數**: ~2000+ 行  
**核心文件數**: 25+ 個主要源文件  