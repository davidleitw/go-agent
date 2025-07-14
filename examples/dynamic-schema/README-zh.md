# 動態 Schema 選擇範例

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

這個範例展示如何根據用戶輸入的關鍵字動態選擇不同的 schema 來收集相應的資訊。

## 🎯 範例概述

此範例實現了一個**智能客服助手**，能夠：

- **🔍 關鍵字檢測** - 分析用戶輸入中的關鍵字
- **📋 動態 Schema 選擇** - 根據關鍵字選擇合適的資訊收集架構
- **💼 情境適應** - 為不同類型的查詢提供專門的回應

## 🏗️ 核心功能

### 1. 關鍵字映射

範例定義了簡單的關鍵字映射系統：

```go
func getSchemaForKeywords(input string) []*schema.Field {
    inputLower := strings.ToLower(input)
    
    // 技術支援 schema
    if strings.Contains(inputLower, "technical") || 
       strings.Contains(inputLower, "error") || 
       strings.Contains(inputLower, "bug") {
        return []*schema.Field{
            schema.Define("email", "Please provide your email for follow-up"),
            schema.Define("error_description", "Please describe the error you're experiencing"),
            schema.Define("browser", "What browser are you using?").Optional(),
        }
    }
    
    // 帳務查詢 schema
    if strings.Contains(inputLower, "billing") || 
       strings.Contains(inputLower, "payment") || 
       strings.Contains(inputLower, "charge") {
        return []*schema.Field{
            schema.Define("email", "Please provide your account email"),
            schema.Define("account_id", "What is your account ID?"),
            schema.Define("billing_question", "Please describe your billing question"),
        }
    }
    
    // 預設的一般查詢 schema
    return []*schema.Field{
        schema.Define("email", "Please provide your email address"),
        schema.Define("topic", "What would you like to know about?"),
    }
}
```

### 2. Schema 類型

範例支援三種 schema 類型：

- **技術支援** - 收集技術問題相關資訊
- **帳務查詢** - 收集帳單和付款相關資訊  
- **一般查詢** - 收集基本聯絡資訊

## 🚀 執行範例

### 前置需求
1. Go 1.21 或更高版本
2. OpenAI API 金鑰

### 設定與執行
```bash
# 1. 設定環境變數
export OPENAI_API_KEY="your-openai-api-key"

# 2. 執行範例
cd examples/dynamic-schema
go run main.go
```

## 📋 測試場景

範例執行三個測試場景：

### 1. 技術錯誤查詢
- **用戶輸入**: "I'm getting a technical error when logging in"
- **觸發**: 技術支援 schema (關鍵字: "technical", "error")
- **收集欄位**: email, error_description, browser

### 2. 帳務問題查詢
- **用戶輸入**: "I have a billing question about my invoice"
- **觸發**: 帳務查詢 schema (關鍵字: "billing")
- **收集欄位**: email, account_id, billing_question

### 3. 一般查詢
- **用戶輸入**: "Hello, I have some general questions"
- **觸發**: 預設 schema (沒有特定關鍵字)
- **收集欄位**: email, topic

## 📊 範例輸出

```
🎯 Dynamic Schema Selection Example
===================================

📝 Scenario 1
👤 User: I'm getting a technical error when logging in
📋 Selected Schema (3 fields):
   - email (required)
   - error_description (required)
   - browser (optional)
🤖 Assistant: I'll help you resolve this technical error...

📝 Scenario 2  
👤 User: I have a billing question about my invoice
📋 Selected Schema (3 fields):
   - email (required)
   - account_id (required)
   - billing_question (required)
🤖 Assistant: I'll help you with your billing inquiry...

📝 Scenario 3
👤 User: Hello, I have some general questions
📋 Selected Schema (2 fields):
   - email (required)
   - topic (required)
🤖 Assistant: Hello! I'd be happy to help with your questions...

✅ Dynamic Schema Selection Example Complete!
```

## 🎓 學習重點

完成此範例後，您將了解：

1. **關鍵字檢測** - 如何分析用戶輸入中的特定關鍵字
2. **動態選擇** - 如何根據內容動態選擇不同的 schema
3. **條件邏輯** - 如何實現簡單的規則引擎
4. **彈性設計** - 如何創建可擴展的 schema 選擇系統

## 🔄 擴展建議

您可以進一步擴展此範例：

1. **增加更多 schema 類型** - 如退款、技術規格查詢等
2. **改進關鍵字檢測** - 使用更複雜的文字分析
3. **添加優先級系統** - 當多個關鍵字匹配時的處理策略
4. **整合機器學習** - 使用 NLP 模型進行意圖分類

## 💡 核心 API

### Schema 定義
- `schema.Define(name, prompt)` - 定義必填欄位
- `.Optional()` - 設定欄位為選填

### 動態選擇
- `strings.Contains()` - 檢查關鍵字存在
- `agent.WithSchema()` - 為對話指定 schema

### 助手創建
- `agent.New()` - 創建新助手
- `.WithOpenAI()` - 設定 OpenAI 提供者
- `.Build()` - 完成助手建構

此範例展示了如何創建**智能、適應性**的客服系統，能夠根據用戶查詢類型自動調整資訊收集策略。