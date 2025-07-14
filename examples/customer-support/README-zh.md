# 客戶支持機器人範例

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

這個範例展示如何使用 **schema 系統**創建一個簡單的客戶支持助手，能夠收集處理支持請求所需的基本資訊。

## 🎯 範例概述

此範例實現了一個**客戶支持助手**，能夠：

- **📋 結構化收集** - 使用 schema 收集支持請求的必要資訊
- **💼 專業回應** - 提供友好的客戶服務體驗  
- **⚡ 簡化流程** - 專注於基本資訊收集功能
- **🔧 實用設計** - 展示實際應用場景

## 🏗️ 核心功能

### 1. 支持 Schema 定義

範例定義了簡單的支持請求 schema：

```go
func supportSchema() []*schema.Field {
    return []*schema.Field{
        schema.Define("email", "Please provide your email address"),
        schema.Define("issue_type", "What type of issue? (technical/billing/general)"),
        schema.Define("description", "Please describe your issue"),
        schema.Define("urgency", "How urgent is this? (low/medium/high)").Optional(),
    }
}
```

### 2. 資訊收集欄位

支持 schema 包含：

- **email** (必填) - 客戶聯絡信箱
- **issue_type** (必填) - 問題類型分類  
- **description** (必填) - 問題詳細描述
- **urgency** (選填) - 緊急程度評估

## 🚀 執行範例

### 前置需求
1. Go 1.21 或更高版本
2. OpenAI API 金鑰

### 設定與執行
```bash
# 1. 設定環境變數
export OPENAI_API_KEY="your-openai-api-key"

# 2. 執行範例
cd examples/customer-support
go run main.go
```

## 📋 測試場景

範例展示單一支持請求場景：

### 基本支持請求
- **用戶輸入**: "I'm having trouble with my account"
- **收集欄位**: email, issue_type, description, urgency
- **助手行為**: 專業地收集必要的支持資訊

## 📊 範例輸出

```
🎧 Customer Support Bot Example
===============================

👤 Customer: I'm having trouble with my account
📋 Information to collect:
   - email (required)
   - issue_type (required)  
   - description (required)
   - urgency (optional)
🤖 Support Agent: I'm here to help you with your account issue. To get started, could you please provide your email address?
⏱️  Response time: 1.234s

✅ Customer Support Example Complete!
```

## 🎓 學習重點

完成此範例後，您將了解：

1. **Schema 設計** - 如何設計實用的資訊收集架構
2. **客戶服務** - 如何創建專業的支持助手
3. **必填與選填** - 如何區分必需和可選資訊
4. **用戶體驗** - 如何提供友好的支持互動

## 🔄 擴展建議

您可以進一步擴展此範例：

1. **多種 Schema** - 為不同問題類型創建專門的 schema
2. **工單系統** - 整合實際的工單創建和追蹤
3. **優先級邏輯** - 根據緊急程度自動路由
4. **知識庫** - 整合常見問題自動回答

## 💡 核心 API

### Schema 定義
- `schema.Define(name, prompt)` - 定義必填欄位
- `.Optional()` - 設定欄位為選填

### 助手配置
- `agent.New(name)` - 創建新助手
- `.WithDescription()` - 設定助手描述
- `.WithInstructions()` - 設定行為指令

### 支持互動
- `agent.WithSession()` - 管理對話狀態
- `agent.WithSchema()` - 指定收集架構
- `response.Data` - 獲取收集的結構化資料

此範例展示了如何創建**專業、高效**的客戶支持系統，提供結構化的問題收集和處理流程。