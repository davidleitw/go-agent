# 任務完成範例

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

此範例展示如何使用 **schema 系統**實現多輪對話的資訊收集，模擬餐廳預訂的完整流程。

## 🎯 範例概述

此範例實現了一個**餐廳預訂助手**，能夠：

- **📋 結構化收集** - 使用 schema 定義預訂所需的所有欄位
- **💬 多輪對話** - 透過多次互動逐步收集完整資訊
- **🔄 狀態追蹤** - 追蹤收集進度和缺失資訊
- **✅ 完成檢測** - 確認所有必要資訊都已收集

## 🏗️ 核心功能

### 1. 預訂 Schema 定義

範例定義了餐廳預訂所需的完整資訊：

```go
func reservationSchema() []*schema.Field {
    return []*schema.Field{
        schema.Define("name", "Please provide your name"),
        schema.Define("phone", "Please provide your phone number"),
        schema.Define("date", "What date would you like to reserve?"),
        schema.Define("time", "What time would you prefer?"),
        schema.Define("party_size", "How many people will be dining?"),
    }
}
```

### 2. 資訊收集欄位

預訂 schema 包含五個必填欄位：

- **name** - 預訂人姓名
- **phone** - 聯絡電話
- **date** - 用餐日期
- **time** - 用餐時間  
- **party_size** - 用餐人數

## 🚀 執行範例

### 前置需求
1. Go 1.21 或更高版本
2. OpenAI API 金鑰

### 設定與執行
```bash
# 1. 設定環境變數
export OPENAI_API_KEY="your-openai-api-key"

# 2. 執行範例
cd examples/task-completion
go run main.go
```

## 📋 測試場景

範例模擬完整的預訂對話流程：

### 多輪對話收集
1. **第一輪**: "I want to make a reservation for dinner"
   - 助手開始收集基本資訊

2. **第二輪**: "My name is John Smith, phone is 555-1234" 
   - 用戶提供姓名和電話

3. **第三輪**: "Tomorrow at 7pm for 4 people"
   - 用戶提供日期、時間和人數

### 進度追蹤
每輪對話後，助手會：
- 識別已收集的資訊
- 確定還缺少哪些欄位
- 引導用戶提供缺失的資訊

## 📊 範例輸出

```
🏪 Task Completion Example
=========================

📋 Collecting information for reservation:
   - name
   - phone
   - date
   - time
   - party_size

🔄 Turn 1
👤 Customer: I want to make a reservation for dinner
🤖 Assistant: I'd be happy to help you make a reservation. To get started, could you please provide your name?
⏱️  Response time: 1.234s

🔄 Turn 2
👤 Customer: My name is John Smith, phone is 555-1234
🤖 Assistant: Thank you John! I have your name and phone number. What date would you like to make the reservation for?
⏱️  Response time: 1.456s
📊 Collected Data: {...}

🔄 Turn 3
👤 Customer: Tomorrow at 7pm for 4 people
🤖 Assistant: Perfect! I have all the information for your reservation...
⏱️  Response time: 1.789s
📊 Collected Data: {...}

✅ Task Completion Example Finished!
```

## 🎓 學習重點

完成此範例後，您將了解：

1. **多輪對話** - 如何透過多次互動收集完整資訊
2. **進度追蹤** - 如何追蹤資訊收集的進度
3. **智能引導** - 如何引導用戶提供缺失的資訊  
4. **完成檢測** - 如何確認任務完成狀態

## 🔄 實際應用場景

此模式適用於許多實際場景：

1. **預訂系統** - 餐廳、飯店、服務預約
2. **申請表單** - 帳戶註冊、貸款申請
3. **客戶入門** - 收集用戶偏好和設定
4. **問卷調查** - 結構化資料收集

## 💡 核心 API

### Schema 定義
- `schema.Define(name, prompt)` - 定義收集欄位
- 多欄位 schema 組合

### 對話管理
- `agent.WithSession()` - 保持對話狀態
- `agent.WithSchema()` - 指定收集架構
- 多輪狀態追蹤

### 資料存取
- `response.Data` - 獲取已收集的結構化資料
- `response.Message` - 獲取助手回應訊息

此範例展示了如何創建**智能、持續**的資訊收集系統，能夠透過自然對話完成複雜的任務流程。