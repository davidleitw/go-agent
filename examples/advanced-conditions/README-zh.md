# 進階條件範例

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

此範例展示了 go-agent 框架中**複雜的條件系統**，演示如何透過優雅的條件流程控制來創建智能、適應性的對話代理。透過全面的用戶入門場景，您將學習以簡潔、可讀的程式碼實現複雜的行為模式。

## 🎯 此範例演示的內容

此範例實現了一個**智能入門代理**，收集用戶資訊（姓名、電子郵件、電話、偏好），並展示：

- **🔄 簡單欄位條件** - 自動提示缺失資訊
- **🧠 組合邏輯** - AND/OR/NOT 條件組合  
- **⏱️ 對話管理** - 基於訊息數量和時間的觸發器
- **💭 內容識別** - 關鍵字和短語檢測
- **🕐 時間感知行為** - 營業時間和週末適應
- **😤 情緒智能** - 挫折感和困難檢測
- **📊 結構化輸出** - 完整的用戶檔案 JSON
- **🛠️ 函數式工具** - 資訊收集和驗證

## 🏗️ 展示的核心功能

### 1. 簡單欄位條件
```go
OnMissingInfo("name").Ask("您好！我很樂意幫助您開始。請問您的姓名是？").Build().
OnMissingInfo("email").Ask("很好！現在我需要您的電子郵件地址來設定帳戶。").Build().
```
自動檢測缺失的必填欄位並適當提示用戶。

### 2. 邏輯運算子組合條件
```go
When(agent.And(
    agent.WhenMissingFields("email"),
    agent.WhenMissingFields("phone"),
)).Ask("我需要您的電子郵件地址和電話號碼才能繼續。").Build().

When(agent.Or(
    agent.WhenContains("frustrated"),
    agent.WhenContains("difficult"),
    agent.WhenMessageCount(8),
)).Ask("讓我幫您簡化這個流程！").Build().
```
使用邏輯運算子組合多個條件以實現複雜行為。

### 3. 對話管理
```go
OnMessageCount(6).Summarize().Build().
OnMessageCount(10).Ask("我們已經聊了一段時間！讓我來幫您總結...").Build().
```
透過提供摘要和指導來管理長對話。

### 4. 基於內容的觸發器
```go
When(agent.WhenContains("help")).Ask("當然！我該如何幫助您？").Build().
When(agent.WhenContains("skip")).Ask("所有資訊都是必需的，但讓我們一起完成！").Build().
```
智能回應特定關鍵字和用戶情緒。

### 5. 時間感知行為
```go
When(agent.WhenFunc("business_hours", func(session agent.Session) bool {
    now := time.Now()
    hour := now.Hour()
    return hour >= 9 && hour <= 17
})).Ask("因為現在是營業時間，我可以提供即時協助！").Build().
```
基於時間、工作日或其他時間因素調整行為。

### 6. 行為模式檢測
```go
When(agent.WhenFunc("retry_attempt", func(session agent.Session) bool {
    messages := session.Messages()
    retryCount := 0
    for _, msg := range messages {
        if strings.Contains(strings.ToLower(msg.Content), "try again") {
            retryCount++
        }
    }
    return retryCount >= 2
})).Ask("我看到您已經重試了幾次。讓我提供額外的指導！").Build().
```
檢測複雜的行為模式並以同理心回應。

## 🛠️ 函數式工具

### 資訊收集工具
```go
collectTool := agent.NewTool("collect_info", 
    "收集和驗證用戶資訊",
    func(field, value string) map[string]any {
        // 自動驗證邏輯
        valid := len(value) > 0
        if field == "email" {
            valid = strings.Contains(value, "@")
        }
        if field == "phone" {
            valid = len(value) >= 10
        }
        
        return map[string]any{
            "field":     field,
            "value":     value,
            "is_valid":  valid,
            "timestamp": time.Now().Format(time.RFC3339),
        }
    })
```

### 檔案驗證工具
```go
validateTool := agent.NewTool("validate_profile",
    "檢查用戶檔案是否完整",
    func(userData map[string]any) map[string]any {
        required := []string{"name", "email", "phone", "preferences"}
        missing := []string{}
        
        for _, field := range required {
            if value, exists := userData[field]; !exists || value == "" {
                missing = append(missing, field)
            }
        }
        
        return map[string]any{
            "is_complete":   len(missing) == 0,
            "missing_fields": missing,
            "completion_rate": float64(len(required)-len(missing))/float64(len(required)),
        }
    })
```

## 📊 結構化輸出

代理產生完整的 `UserProfile` 結構作為 JSON 輸出：

```go
type UserProfile struct {
    Name         string   `json:"name"`
    Email        string   `json:"email"`
    Phone        string   `json:"phone"`
    Preferences  []string `json:"preferences"`
    CompletedAt  string   `json:"completed_at,omitempty"`
    IsComplete   bool     `json:"is_complete"`
    MissingInfo  []string `json:"missing_info,omitempty"`
    StatusText   string   `json:"status_text"`
}
```

## 🚀 執行範例

### 前置需求
1. Go 1.22 或更高版本
2. OpenAI API 金鑰

### 設定
```bash
# 1. 配置您的 API 金鑰
cp .env.example .env
# 編輯 .env 並添加您的 OPENAI_API_KEY

# 2. 執行範例
cd examples/advanced-conditions
go run main.go
```

## 📋 測試場景

範例執行全面的測試場景：

1. **初始聯絡** - 觸發姓名收集條件
2. **求助請求** - 展示基於內容的條件  
3. **資訊收集** - 顯示漸進式欄位收集
4. **跳過嘗試** - 優雅處理用戶抗拒
5. **挫折檢測** - 回應用戶困難
6. **完成追蹤** - 產生結構化輸出

### 範例輸出
```
🎯 進階條件範例
展示優雅的條件使用和流程控制
============================================================

🚀 測試進階條件系統
============================================================

🔄 測試 1/8: 初始聯絡 - 應觸發缺少姓名條件
👤 用戶: 嗨！我想註冊。
🤖 助手: 您好！我很樂意幫助您開始。請問您的姓名是？

🔄 測試 2/8: 提供姓名 - 應詢問電子郵件
👤 用戶: 我的名字是 Alex Chen
🤖 助手: 很好！現在我需要您的電子郵件地址來設定帳戶。

🔄 測試 3/8: 包含 'help' - 應觸發幫助條件
👤 用戶: 您能幫我了解需要什麼資訊嗎？
🤖 助手: 當然！我在這裡幫助您完成檔案。您需要什麼特定資訊的協助？

📊 檔案狀態: 完成=false，缺失=[email, phone, preferences]
📈 會話訊息: 6

🎉 檔案成功完成！
```

## 🎓 條件類型參考

| 條件類型 | 語法 | 使用場景 | 範例 |
|----------|------|---------|------|
| **欄位條件** | `OnMissingInfo("field")` | 必填欄位收集 | 缺失電子郵件/電話 |
| **訊息計數** | `OnMessageCount(n)` | 對話管理 | 長對話 |
| **內容檢測** | `WhenContains("text")` | 關鍵字回應 | 求助請求 |
| **自訂函數** | `WhenFunc("name", fn)` | 複雜邏輯 | 營業時間 |
| **邏輯 AND** | `And(cond1, cond2)` | 多重需求 | 電子郵件 AND 電話 |
| **邏輯 OR** | `Or(cond1, cond2)` | 替代觸發器 | 挫折 OR 困難 |
| **邏輯 NOT** | `Not(condition)` | 否定邏輯 | 非週末期間 |

## 🏆 主要學習目標

學習此範例後，您應該了解：

1. **漸進條件複雜度** - 從簡單到複雜的建構
2. **優雅語法設計** - 可讀、可維護的條件定義
3. **行為智能** - 創建具有情緒感知的對話代理
4. **工具整合** - 結合條件與函數式工具
5. **結構化輸出** - 管理複雜的資料收集工作流程
6. **時間感知行為** - 創建情境敏感的回應
7. **模式識別** - 檢測和回應用戶行為模式

## 🔄 前後對比：條件系統

### 傳統方法（複雜且冗長）
```go
// 手動條件檢查
if len(session.Messages()) > 5 {
    if containsKeyword(lastMessage, "help") {
        if isBusinessHours() && userSeemsFrustrated(session) {
            // 複雜的嵌套邏輯...
        }
    }
}
```

### go-agent 方法（優雅且可讀）
```go
// 宣告式條件系統
When(agent.And(
    agent.WhenMessageCount(5),
    agent.WhenContains("help"),
    agent.WhenFunc("business_hours", isBusinessHours),
    agent.WhenFunc("frustrated", detectFrustration),
)).Ask("讓我提供即時協助！").Build()
```

## 🎯 展示的最佳實踐

1. **從簡單開始** - 從基本欄位條件開始，逐步增加複雜度
2. **深思熟慮地組合** - 使用 AND/OR 創建有意義的條件組合  
3. **處理邊緣情況** - 包含常見用戶行為的條件（幫助、跳過、挫折）
4. **維持可讀性** - 為自訂條件使用描述性名稱
5. **徹底測試** - 驗證條件在預期場景中觸發
6. **漸進增強** - 從簡單到複雜的條件層級

## 🔄 下一步

掌握此範例後：
1. **[多工具代理](../multi-tool-agent/)** - 協調多種能力
2. **[任務完成](../task-completion/)** - 進階結構化輸出處理
3. **自訂條件** - 創建您自己的條件類型
4. **整合模式** - 與外部系統結合

## 🐛 常見模式與解決方案

### 模式：漸進式資訊收集
```go
OnMissingInfo("name").Ask("您的姓名是什麼？").Build().
OnMissingInfo("email").Ask("您的電子郵件是什麼？").Build().
```

### 模式：挫折檢測與回應
```go
When(agent.Or(
    agent.WhenContains("frustrated"),
    agent.WhenMessageCount(8),
)).Ask("讓我幫您簡化這個流程！").Build().
```

### 模式：時間感知行為
```go
When(agent.WhenFunc("after_hours", func(session agent.Session) bool {
    return time.Now().Hour() > 17
})).Ask("我 24/7 都在這裡幫助您！").Build().
```

## 💡 展示的關鍵 API

### 條件創建
- `OnMissingInfo(fields...)` - 欄位條件
- `OnMessageCount(n)` - 訊息數量閾值
- `WhenContains(text)` - 內容檢測
- `WhenFunc(name, fn)` - 自訂函數條件

### 邏輯運算子
- `agent.And(conditions...)` - 所有條件必須為真
- `agent.Or(conditions...)` - 任一條件必須為真  
- `agent.Not(condition)` - 否定條件

### 流程動作
- `.Ask(message)` - 用特定訊息提示用戶
- `.Summarize()` - 請求對話摘要
- `.UseTemplate(template)` - 應用動態指令

此範例展示了 go-agent 的條件系統如何讓您創建**智能、適應性和富有同理心**的對話代理，同時保持**優雅、可維護的程式碼**。