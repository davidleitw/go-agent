# 進階條件範例

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

此範例展示如何定義和創建 **go-agent 的條件系統**，演示條件和流程規則的基本概念。

## 🎯 此範例演示的內容

此範例展示了**條件系統的基礎概念**：

- **🔄 條件定義** - 如何創建不同類型的條件
- **💭 關鍵字檢測** - 使用 `conditions.Contains()` 識別特定文字
- **⏱️ 對話計數** - 使用 `conditions.Count()` 追蹤對話長度
- **🧠 組合邏輯** - 使用 `conditions.Or()` 組合多個條件

## 🏗️ 展示的核心功能

### 1. 基本條件使用

範例展示四種基本的流程規則：

```go
// 歡迎問候規則
welcomeRule := agent.NewFlowRule("welcome_new_users", 
    conditions.Or(
        conditions.Contains("hello"),
        conditions.Contains("hi"),
        conditions.Contains("hey"),
    )).
    WithDescription("歡迎說 hello 或問候的用戶").
    WithNewInstructions("用戶正在問候你。要格外歡迎和友好。詢問今天如何幫助他們。").
    WithPriority(10).
    Build()

// 緊急請求規則  
urgentRule := agent.NewFlowRule("urgent_requests",
    conditions.Or(
        conditions.Contains("urgent"),
        conditions.Contains("emergency"),
        conditions.Contains("asap"),
        conditions.Contains("immediately"),
    )).
    WithDescription("優先處理緊急或急迫請求").
    WithNewInstructions("這是緊急請求。快速回應並提供即時協助。").
    WithPriority(20).
    Build()

// 技術支援規則
techRule := agent.NewFlowRule("technical_support",
    conditions.Or(
        conditions.Contains("code"),
        conditions.Contains("programming"), 
        conditions.Contains("debug"),
        conditions.Contains("error"),
        conditions.Contains("technical"),
    )).
    WithDescription("切換到技術模式處理程式設計/技術問題").
    WithNewInstructions("用戶需要技術幫助。提供詳細的逐步技術指導。").
    WithPriority(15).
    Build()

// 長對話引導規則
countRule := agent.NewFlowRule("long_conversation", 
    conditions.Count(8)).
    WithDescription("為訊息較多的對話提供摘要").
    WithNewInstructions("這是一個長對話。考慮提供到目前為止討論內容的摘要。").
    WithPriority(5).
    Build()
```

### 2. 條件類型

範例使用的條件類型：

- **`conditions.Contains(text)`** - 檢查用戶輸入是否包含特定文字
- **`conditions.Count(n)`** - 檢查對話訊息數量是否達到閾值
- **`conditions.Or(...)`** - 組合多個條件，任一成立即觸發

## 🚀 執行範例

### 前置需求
1. Go 1.21 或更高版本
2. OpenAI API 金鑰

### 設定與執行
```bash
# 1. 設定環境變數
export OPENAI_API_KEY="your-openai-api-key"

# 2. 執行範例
cd examples/advanced-conditions
go run main.go
```

## 📋 測試場景

範例展示以下對話場景，說明不同條件如何被觸發：

1. **問候場景** - 演示歡迎條件的定義
   - 用戶: "Hello! I'm new here."
   - 展示: 如何檢測問候關鍵字

2. **技術問題** - 演示技術支援條件
   - 用戶: "I'm having trouble with my Python code."
   - 展示: 如何檢測技術相關關鍵字

3. **緊急請求** - 演示緊急處理條件
   - 用戶: "This is urgent! I need help immediately."
   - 展示: 如何檢測緊急關鍵字

4. **長對話** - 演示對話計數條件
   - 展示: 如何在對話達到一定長度時觸發

### 範例輸出
```
🎯 Advanced Conditions Example
==============================
This example demonstrates intelligent flow control using conditions.
The agent adapts its behavior based on conversation context.

✅ Created assistant with 4 conditional flow rules

📝 Scenario 1: Greeting scenario (should trigger welcome rule)
👤 User: Hello! I'm new here.
🎯 Expected: Welcoming response
🤖 Assistant: Hello! How can I help you today?

📝 Scenario 2: Technical question (should trigger technical support rule)  
👤 User: I'm having trouble with my Python code...
🎯 Expected: Technical assistance mode
🤖 Assistant: I'd be happy to help you with your Python code...
```

**注意**: 此範例主要展示條件定義的概念。在完整實現中，這些條件將會自動調整助手的行為。

## 🎓 學習重點

完成此範例後，您將了解：

1. **條件系統基礎** - 如何定義和使用流程規則
2. **關鍵字檢測** - 使用 `conditions.Contains()` 識別用戶意圖
3. **條件組合** - 使用 `conditions.Or()` 組合多個觸發條件
4. **對話管理** - 使用 `conditions.Count()` 管理長對話
5. **優先級控制** - 使用 `WithPriority()` 設定規則優先順序

## 🔄 下一步

掌握此範例後，您可以：
1. 探索 **[多工具代理](../multi-tool-agent/)** - 學習工具整合
2. 查看 **[簡單架構](../simple-schema/)** - 了解結構化資料收集
3. 創建自己的條件規則和對話流程

## 💡 展示的核心 API

### 條件函數
- `conditions.Contains(text)` - 文字內容檢測
- `conditions.Count(n)` - 訊息數量閾值
- `conditions.Or(...)` - 邏輯 OR 組合

### 流程規則建構
- `agent.NewFlowRule(name, condition)` - 創建新規則
- `.WithDescription(desc)` - 添加描述
- `.WithNewInstructions(prompt)` - 設定觸發時的指令
- `.WithPriority(level)` - 設定優先級
- `.Build()` - 完成規則建構

此範例展示了如何使用條件系統創建**智能、適應性**的對話代理，讓您的 AI 助手能夠根據不同情境提供合適的回應。