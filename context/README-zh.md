# Context 模組

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

Context 模組提供統一的 data structure 和 provider system，用於管理需要傳遞給 LLM 的不同類型資訊在 Go Agent Framework 中。它充當各種資料來源（sessions、prompts、tools）與最終 prompt 建構之間的橋樑。

## 功能特色

- **統一的 `Context` 結構**：所有類型資訊的一致資料格式
- **Provider Pattern**：可擴展的不同 context sources 系統
- **Session 整合**：自動將 session history 轉換為 contexts
- **Metadata 支援**：豐富的 metadata 保存和增強
- **Type Safety**：適當處理 sessions 中的不同 entry types

## 核心概念

### Context Structure

```go
type Context struct {
    Type     string         // context type（例如 "user"、"assistant"、"system"、"tool_call"）
    Content  string         // context 的文字內容
    Metadata map[string]any // 額外的 metadata 和資訊
}
```

### Provider Interface

```go
type Provider interface {
    Provide(ctx context.Context, s session.Session) []Context
}
```

Providers 負責將 data sources 轉換為可被 Agent 使用的 `Context` objects。

## 內建提供器

### SystemPromptProvider

提供系統級別的指令和提示。

```go
provider := context.NewSystemPromptProvider("你是一個有用的助手。")
contexts := provider.Provide(ctx, session)
// 返回：[Context{Type: "system", Content: "你是一個有用的助手。", ...}]
```

### HistoryProvider

將會話歷史轉換為上下文物件，處理所有條目類型。

```go
provider := context.NewHistoryProvider(10) // 限制為最近 10 個條目
contexts := provider.Provide(ctx, session)
```

#### 條目類型轉換

1. **訊息條目** → 基於角色的類型上下文
   - `session.MessageContent{Role: "user", Text: "你好"}` → `Context{Type: "user", Content: "你好"}`

2. **工具呼叫條目** → 格式化工具資訊的上下文
   - `session.ToolCallContent{Tool: "search", Parameters: {...}}` → `Context{Type: "tool_call", Content: "Tool: search\nParameters: {...}"}`

3. **工具結果條目** → 包含成功/錯誤資訊的上下文
   - 成功：`Context{Type: "tool_result", Content: "Tool: search\nSuccess: true\nResult: [...]"}`
   - 錯誤：`Context{Type: "tool_result", Content: "Tool: search\nSuccess: false\nError: 連線失敗"}`

4. **思考條目** → 內部推理的上下文
   - `Context{Type: "thinking", Content: "我需要考慮..."}`

## 使用範例

### 基本用法

```go
import (
    "github.com/davidleitw/go-agent/context"
    "github.com/davidleitw/go-agent/session/memory"
)

// 創建有歷史的會話
store := memory.NewStore()
sess := store.Create(context.Background())

sess.AddEntry(session.NewMessageEntry("user", "預訂飛往東京的航班"))
sess.AddEntry(session.NewMessageEntry("assistant", "我來幫你預訂航班。"))

// 取得系統上下文
systemProvider := context.NewSystemPromptProvider("你是一個旅行助手。")
systemContexts := systemProvider.Provide(ctx, sess)

// 取得歷史上下文
historyProvider := context.NewHistoryProvider(5)
historyContexts := historyProvider.Provide(ctx, sess)

// 為 LLM 提示組合上下文
allContexts := append(systemContexts, historyContexts...)
```

### 進階中繼資料使用

```go
// HistoryProvider 自動新增中繼資料
contexts := historyProvider.Provide(ctx, sess)
for _, ctx := range contexts {
    entryID := ctx.Metadata["entry_id"].(string)
    timestamp := ctx.Metadata["timestamp"].(time.Time)
    
    if ctx.Type == "tool_call" {
        toolName := ctx.Metadata["tool_name"].(string)
        // 使用工具名稱進行特殊處理
    }
    
    if ctx.Type == "tool_result" {
        success := ctx.Metadata["success"].(bool)
        // 根據成功狀態處理
    }
}
```

## 上下文類型

### 標準類型
- `"system"` - 系統提示和指令
- `"user"` - 用戶訊息和輸入
- `"assistant"` - 助手回應
- `"tool_call"` - 工具調用請求
- `"tool_result"` - 工具執行結果
- `"thinking"` - 內部推理（為未來使用保留）

### 自訂類型
提供器可以為特殊用例定義自訂類型。未知類型會使用 JSON 後備格式優雅處理。

## 中繼資料欄位

### 通用欄位（由 HistoryProvider 新增）
- `"entry_id"` - 來自會話條目的唯一識別符
- `"timestamp"` - 條目創建時間

### 工具特定欄位
- `"tool_name"` - 工具名稱（用於 tool_call 和 tool_result 類型）
- `"success"` - 布林成功狀態（用於 tool_result 類型）

### 自訂欄位
會話條目的原始中繼資料得以保留，允許應用程式特定的資訊。

## 測試

模組包含全面的測試覆蓋：

- 空歷史處理
- 所有條目類型轉換
- 中繼資料保存和增強
- 排序（來自會話歷史的最新優先）
- 限制功能
- 混合條目類型
- 未知類型的後備處理
- 錯誤處理

執行測試：
```bash
go test ./context -v
```

## 創建自訂提供器

```go
type CustomProvider struct {
    data string
}

func NewCustomProvider(data string) Provider {
    return &CustomProvider{data: data}
}

func (p *CustomProvider) Provide(ctx context.Context, s session.Session) []Context {
    // 自訂邏輯來產生上下文
    return []Context{
        {
            Type:     "custom",
            Content:  p.data,
            Metadata: map[string]any{"source": "custom_provider"},
        },
    }
}
```

## 與 Agent 系統整合

*待實作：此部分將在 Agent 核心功能完成後填入*

Context 模組設計為與 Agent 的提示建構系統無縫整合：

```go
// 預期的整合
agent := agent.New(
    agent.WithSystemPrompt("你很有用"),
    agent.WithContextProviders(
        context.NewHistoryProvider(10),
        // 其他提供器...
    ),
)
```

## 效能考量

- **記憶體效率**：上下文按需創建，不進行快取
- **排序**：歷史已按會話排序，無需額外排序
- **JSON 編組**：僅在複雜資料類型需要時執行
- **中繼資料複製**：為提升效能進行淺複製，同時保留原始資料

## 最佳實踐

1. **提供器選擇**：為你的用例使用適當的提供器
   - SystemPromptProvider 用於靜態指令
   - HistoryProvider 用於對話上下文

2. **限制設定**：為 HistoryProvider 設定合理限制以控制上下文大小
   - 考慮目標 LLM 的 token 限制
   - 在上下文豐富性和效能之間平衡

3. **中繼資料使用**：善用中繼資料用於：
   - 除錯和記錄
   - 基於條目類型的條件邏輯
   - 分析和監控

4. **錯誤處理**：提供器設計為具有彈性
   - 無效條目會優雅處理
   - 類型斷言失敗會後退到 JSON 格式化

## 架構

```
會話歷史 → HistoryProvider → Context[] → Agent → LLM
系統提示  → SystemPromptProvider → Context[] ↗
自訂資料    → CustomProvider → Context[] ↗
```

Context 模組作為資料轉換層，將各種輸入來源轉換為 Agent 可以處理並發送給 LLM 的統一格式。

## 未來擴展

1. **額外提供器**：
   - FileProvider（用於文件上下文）
   - DatabaseProvider（用於動態資料）
   - APIProvider（用於外部資料來源）

2. **上下文過濾**：
   - 基於內容的過濾
   - 基於時間的過濾
   - 相關性評分

3. **上下文壓縮**：
   - 大型上下文的智慧截斷
   - 舊歷史的摘要
   - 基於重要性的選擇

## 貢獻

在新增新提供器或擴展功能時：
- 遵循 Provider 介面
- 新增全面的測試
- 更新文檔
- 考慮中繼資料標準化
- 確保優雅的錯誤處理

## 授權

MIT 授權 - 請參閱專案根目錄中的 LICENSE 檔案。