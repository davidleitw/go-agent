# 簡單 Schema 收集範例

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

這個範例展示了 go-agent 中基於 schema 的信息收集基本概念。學習如何通過自然對話自動收集用戶的結構化信息。

## 概述

傳統的聊天機器人需要複雜的狀態管理來收集用戶信息。使用 go-agent 的 schema 系統，你只需定義需要的信息，框架會自動：

- **提取** 使用 LLM 語義理解從用戶消息中提取信息
- **識別** 缺失的必需字段
- **詢問** 使用自然提示詢問缺失信息
- **記憶** 跨對話輪次記住收集的信息

## 快速開始

```bash
# 設置你的 OpenAI API key
export OPENAI_API_KEY="your-api-key-here"

# 運行範例
go run examples/simple-schema/main.go
```

## 代碼詳解

### 1. 字段定義

```go
import "github.com/davidleitw/go-agent/pkg/schema"

// 定義必需字段
emailField := schema.Define("email", "請提供您的電子郵件地址")
issueField := schema.Define("issue", "請描述您的問題")

// 定義可選字段
phoneField := schema.Define("phone", "緊急聯絡電話").Optional()
```

**關鍵要點：**
- `schema.Define(name, prompt)` 預設創建必需字段
- `.Optional()` 使字段變為可選
- prompt 是當字段缺失時用戶看到的提示

### 2. Agent 創建

```go
assistant, err := agent.New("simple-assistant").
    WithOpenAI(os.Getenv("OPENAI_API_KEY")).
    WithModel("gpt-4o-mini").
    WithInstructions("您是一個幫助收集用戶信息的助手。").
    Build()
```

**關鍵要點：**
- 標準的 agent 創建，集成 OpenAI
- 指令幫助 agent 理解其角色
- schema 支持無需特殊配置

### 3. Schema 應用

```go
response, err := assistant.Chat(ctx, "我需要幫助",
    agent.WithSchema(
        schema.Define("email", "請提供您的電子郵件地址"),
        schema.Define("issue", "請描述您的問題"),
    ),
)
```

**關鍵要點：**
- `agent.WithSchema()` 將 schema 應用到特定對話
- 一次調用可以定義多個字段
- Schema 只適用於這次聊天互動

### 4. 信息提取

```go
userInput := "我的郵箱是 user@example.com，我有一個帳單問題"

response, err := assistant.Chat(ctx, userInput,
    agent.WithSchema(
        schema.Define("email", "請提供您的電子郵件地址"),
        schema.Define("question", "請詳細描述您的帳單問題"),
    ),
)

// 框架自動提取：
// - email: "user@example.com" ✓
// - question: 缺失，會詢問
```

**關鍵要點：**
- LLM 理解上下文和含義，不僅僅是精確匹配
- 部分信息被提取並記住
- 缺失信息自動識別

### 5. 多輪對話

```go
// 第一輪：用戶提供部分信息
response1, _ := assistant.Chat(ctx, "你好，我叫 John",
    agent.WithSchema(
        schema.Define("name", "請提供您的姓名"),
        schema.Define("email", "請提供您的電子郵件地址"),
        schema.Define("topic", "您想討論什麼主題？"),
    ),
)
// Agent 詢問缺失的 email 和 topic

// 第二輪：用戶提供更多信息
response2, _ := assistant.Chat(ctx, "我的郵箱是 john@example.com，我想談論價格",
    agent.WithSchema(/* 相同的 schema */),
)
// Agent 確認收集完成
```

**關鍵要點：**
- Session 自動維護對話上下文
- 先前收集的信息被記住
- Schema 持續運行直到所有必需字段收集完成

### 6. 檢查收集狀態

```go
// 檢查收集是否進行中
if schemaCollection, ok := response.Metadata["schema_collection"].(bool); ok && schemaCollection {
    fmt.Println("仍在收集信息...")
    if missingFields, ok := response.Metadata["missing_fields"].([]string); ok {
        fmt.Printf("缺失：%v\n", missingFields)
    }
} else {
    fmt.Println("所有必需信息收集完成！")
}
```

**關鍵要點：**
- `response.Metadata["schema_collection"]` 指示收集是否活躍
- `response.Metadata["missing_fields"]` 顯示仍需要什麼
- 使用這個來實現自定義收集邏輯

## 範例場景

### 場景 1：基本郵箱收集

```go
// 用戶："我需要幫助"
// Schema：email（必需）
// 結果：Agent 詢問郵箱地址

response, err := assistant.Chat(ctx, "我需要幫助",
    agent.WithSchema(
        schema.Define("email", "請提供您的電子郵件地址"),
    ),
)
// 回應："我很樂意幫助您！請提供您的電子郵件地址。"
```

### 場景 2：多個必需字段

```go
// 用戶："我想報告一個問題"
// Schema：email、issue（都是必需的）
// 結果：Agent 詢問缺失信息

response, err := assistant.Chat(ctx, "我想報告一個問題",
    agent.WithSchema(
        schema.Define("email", "請提供您的電子郵件地址"),
        schema.Define("issue", "請描述您的問題"),
    ),
)
// 回應：Agent 詢問郵箱（問題類型已從上下文理解）
```

### 場景 3：必需和可選字段混合

```go
// 用戶："我的帳戶需要幫助"
// Schema：email（必需）、issue（必需）、phone（可選）
// 結果：Agent 首先關注必需字段

response, err := assistant.Chat(ctx, "我的帳戶需要幫助",
    agent.WithSchema(
        schema.Define("email", "請提供您的電子郵件地址"),
        schema.Define("issue", "請描述您遇到的問題"),
        schema.Define("phone", "緊急聯絡電話").Optional(),
    ),
)
// 回應：Agent 詢問郵箱和問題描述
```

### 場景 4：提供部分信息

```go
// 用戶："我的郵箱是 user@example.com，我有帳單問題"
// Schema：email、question、account_type（提取郵箱，其他缺失）
// 結果：Agent 詢問剩餘必需信息

response, err := assistant.Chat(ctx, 
    "我的郵箱是 user@example.com，我有帳單問題",
    agent.WithSchema(
        schema.Define("email", "請提供您的電子郵件地址"),
        schema.Define("question", "請詳細描述您的帳單問題"),
        schema.Define("account_type", "您是什麼類型的帳戶？").Optional(),
    ),
)
// 回應：Agent 確認郵箱，詢問詳細問題
```

## 最佳實踐

### 1. 字段設計

**好的字段提示：**
```go
schema.Define("email", "請提供您的電子郵件地址")
schema.Define("issue", "請詳細描述您的問題")
schema.Define("urgency", "這有多緊急？（低/中/高）")
```

**避免：**
```go
schema.Define("email", "郵箱")  // 太簡短
schema.Define("data", "輸入數據")  // 太模糊
schema.Define("field1", "需要輸入")  // 不描述性
```

### 2. 必需 vs 可選

**使用必需字段適用於：**
- 重要聯絡信息（郵箱）
- 核心業務數據（問題描述）
- 繼續進行所需的信息

**使用可選字段適用於：**
- 額外的詳細信息（電話號碼）
- 偏好設置（聯絡方式）
- 附加上下文（之前的嘗試）

### 3. 字段命名

```go
// 好：描述性和清晰
schema.Define("customer_email", "您的電子郵件地址")
schema.Define("issue_description", "描述您遇到的問題")
schema.Define("preferred_contact_method", "您希望我們如何聯絡您？")

// 避免：通用或令人困惑的名稱
schema.Define("data1", "郵箱")
schema.Define("info", "詳細信息")
schema.Define("field", "信息")
```

### 4. 錯誤處理

```go
response, err := assistant.Chat(ctx, userInput, agent.WithSchema(schema...))
if err != nil {
    log.Printf("聊天失敗：%v", err)
    return
}

// 檢查收集是否完成
if schemaActive := response.Metadata["schema_collection"]; schemaActive == true {
    // 繼續收集
    fmt.Println("收集更多信息中...")
} else {
    // 處理收集的數據
    fmt.Println("信息收集完成！")
    processCollectedData(response.Session)
}
```

## 測試

運行範例測試：

```bash
go test ./examples/simple-schema/
```

測試驗證：
- 字段定義和屬性
- Schema 應用到對話
- 信息提取準確性
- 缺失字段識別

## 集成範例

### 與 Web 表單

```go
// 將 web 表單轉換為 schema
func webFormToSchema(formFields []FormField) []*schema.Field {
    var schemaFields []*schema.Field
    for _, field := range formFields {
        schemaField := schema.Define(field.Name, field.Label)
        if !field.Required {
            schemaField = schemaField.Optional()
        }
        schemaFields = append(schemaFields, schemaField)
    }
    return schemaFields
}
```

### 與驗證

```go
// 添加自定義驗證
func validateEmail(email string) bool {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}

// 在收集流程中使用
if email := extractFieldFromSession("email", session); email != "" {
    if !validateEmail(email) {
        // 重新詢問郵箱，提供具體指導
        return agent.WithSchema(
            schema.Define("email", "請提供有效的電子郵件地址（例如：user@example.com）"),
        )
    }
}
```

## 相關範例

- **[客戶支持](../customer-support/)**：使用專業 schema 的真實世界支持機器人
- **[動態 Schema](../dynamic-schema/)**：基於意圖的高級 schema 選擇
- **[基本聊天](../basic-chat/)**：不使用 schema 的基礎概念
- **[任務完成](../task-completion/)**：帶驗證的結構化輸出

## 下一步

1. **試用範例**：運行代碼並嘗試不同的輸入
2. **修改字段**：添加您自己的字段和提示
3. **測試邊緣情況**：嘗試部分信息、拼寫錯誤、不同措辭
4. **探索高級功能**：進入 customer-support 或 dynamic-schema 範例

## 故障排除

**問題**：信息未被提取
**解決方案**：檢查字段名稱是否描述性強，提示是否清晰

**問題**：Agent 詢問已提供的信息
**解決方案**：驗證字段名稱是否與用戶輸入中的語義含義匹配

**問題**：收集永不完成
**解決方案**：確保所有必需字段都是合理和可實現的

要獲得更多幫助，請參閱[主要範例文檔](../README.md)或[schema 收集指南](../../docs/schema-collection.md)。