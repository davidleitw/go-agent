# 簡單 Schema 收集範例

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

這個範例展示 **go-agent 的 schema 系統**基礎概念，學習如何定義資訊收集架構並自動收集用戶資料。

## 🎯 範例概述

此範例實現了一個**基礎資訊收集助手**，展示：

- **📋 Schema 定義** - 如何定義需要收集的資訊欄位
- **🔄 自動收集** - 助手自動識別和收集缺失資訊
- **💬 自然對話** - 透過對話方式收集結構化資料
- **📊 結果展示** - 顯示收集到的結構化資料

## 🏗️ 核心功能

### 1. 聯絡資訊 Schema

範例定義了簡單的聯絡資訊收集架構：

```go
func main() {
    // 定義聯絡資訊 schema
    contactSchema := []*schema.Field{
        schema.Define("name", "What's your name?"),
        schema.Define("email", "Please provide your email address"),
        schema.Define("phone", "Contact number").Optional(),
    }
}
```

### 2. Schema 欄位說明

聯絡 schema 包含：

- **name** (必填) - 用戶姓名
- **email** (必填) - 電子郵件地址  
- **phone** (選填) - 聯絡電話號碼

### 3. 欄位類型

Schema 支援兩種欄位類型：

- **必填欄位** - 使用 `schema.Define(name, prompt)`
- **選填欄位** - 使用 `.Optional()` 方法標記

## 🚀 執行範例

### 前置需求
1. Go 1.21 或更高版本
2. OpenAI API 金鑰

### 設定與執行
```bash
# 1. 設定環境變數
export OPENAI_API_KEY="your-openai-api-key"

# 2. 執行範例
cd examples/simple-schema
go run main.go
```

## 📋 測試場景

範例展示單一資訊收集場景：

### 基本聯絡資訊收集
- **用戶輸入**: "Hi, I need some help"
- **收集目標**: name, email, phone (選填)
- **助手行為**: 自動引導用戶提供缺失的資訊

## 📊 範例輸出

```
🔧 Simple Schema Collection Example
===================================

📋 Schema fields:
   - name (required)
   - email (required)
   - phone (optional)

👤 User: Hi, I need some help
🤖 Assistant: Hello! I'd be happy to help you. To get started, could you please tell me your name?

📊 Collected Data: {...}

✅ Simple Schema Example Complete!
```

## 🎓 學習重點

完成此範例後，您將了解：

1. **Schema 基礎** - 如何定義資訊收集架構
2. **欄位定義** - 必填與選填欄位的區別
3. **自動收集** - 助手如何自動識別和收集資訊
4. **結構化輸出** - 如何獲取收集到的結構化資料

## 🔄 進階應用

此基礎模式可擴展到更複雜的場景：

1. **用戶註冊** - 收集完整的用戶註冊資訊
2. **問卷調查** - 結構化問卷資料收集
3. **客戶資料** - CRM 系統的客戶資訊收集
4. **表單自動化** - 各種表單資料的智能收集

## 💡 核心 API

### Schema 定義
- `schema.Define(name, prompt)` - 定義必填欄位
- `.Optional()` - 標記欄位為選填
- 欄位陣列組合

### 助手配置
- `agent.New(name)` - 創建新助手
- `.WithOpenAI(key)` - 設定 OpenAI 提供者
- `.WithInstructions(text)` - 設定助手指令

### 對話執行
- `.Chat(ctx, input, options...)` - 執行對話
- `agent.WithSession(id)` - 指定會話ID
- `agent.WithSchema(fields...)` - 指定收集架構

### 資料存取
- `response.Message` - 助手回應訊息
- `response.Data` - 收集到的結構化資料

此範例是學習 go-agent schema 系統的**最佳起點**，展示了如何用最少的程式碼實現智能的資訊收集功能。