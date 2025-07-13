# 基本聊天範例

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

此範例演示了 go-agent 框架的基本用法，透過創建一個簡單的對話式 AI 代理。

## 概述

基本聊天範例展示了：
- **環境配置**: 從 `.env` 檔案載入 OpenAI API 金鑰
- **代理創建**: 使用函數式選項模式配置代理
- **簡單對話**: 執行多輪對話
- **會話管理**: 追蹤對話歷史記錄
- **詳細日誌**: 用於除錯和監控的全面日誌記錄

## 程式碼結構

### 核心組件

1. **環境設定**
   ```go
   if err := godotenv.Load("../../../.env"); err != nil {
       log.Printf("Warning: Could not load .env file: %v", err)
   }
   
   apiKey := os.Getenv("OPENAI_API_KEY")
   if apiKey == "" {
       log.Fatal("❌ OPENAI_API_KEY environment variable is required")
   }
   ```
   - 從 `.env` 檔案載入環境變數
   - 驗證 OpenAI API 金鑰是否存在
   - 提供清晰的錯誤訊息（如果配置遺失）

2. **代理創建**
   ```go
   assistant, err := agent.New(
       agent.WithName("helpful-assistant"),
       agent.WithDescription("A helpful AI assistant for general conversations"),
       agent.WithInstructions("You are a helpful, friendly AI assistant..."),
       agent.WithOpenAI(apiKey),
       agent.WithModel("gpt-4"),
       agent.WithModelSettings(&agent.ModelSettings{
           Temperature: floatPtr(0.7),
           MaxTokens:   intPtr(1000),
       }),
       agent.WithSessionStore(agent.NewInMemorySessionStore()),
       agent.WithDebugLogging(),
   )
   ```
   - 使用函數式選項模式進行乾淨的配置
   - 配置 OpenAI 使用 GPT-4 模型
   - 設定溫度為 0.7，平衡創造性和一致性
   - 啟用除錯日誌以獲得詳細的執行追蹤

3. **對話流程**
   ```go
   conversations := []struct {
       user     string
       expected string
   }{
       {
           user:     "Hello! How are you doing today?",
           expected: "greeting response",
       },
       // ... 更多範例
   }
   ```
   - 預定義的對話範例用於一致性測試
   - 每一輪展示不同類型的互動

4. **回應處理**
   ```go
   response, structuredOutput, err := assistant.Chat(ctx, sessionID, conv.user)
   if err != nil {
       log.Printf("❌ ERROR[%d]: Failed to get response: %v", i+1, err)
       continue
   }
   
   fmt.Printf("🤖 Assistant: %s\n", response.Content)
   ```
   - 優雅地處理錯誤
   - 記錄回應詳情以便除錯
   - 向用戶顯示格式化輸出

## 日誌系統

範例包含多層級的全面日誌記錄：

- **REQUEST**: 用戶輸入和請求參數
- **RESPONSE**: LLM 回應詳情，包括持續時間和內容長度
- **SESSION**: 會話狀態追蹤和訊息計數
- **ERROR**: 詳細的錯誤資訊

### 日誌輸出範例
```
✅ OpenAI API key loaded (length: 51)
📝 Creating AI agent...
✅ Agent 'helpful-assistant' created successfully
🆔 Session ID: basic-chat-1704067200
REQUEST[1]: Sending user input to agent
REQUEST[1]: Input: Hello! How are you doing today?
RESPONSE[1]: Duration: 1.234s
RESPONSE[1]: Content length: 87 characters
SESSION[1]: Total messages: 2
```

## 執行範例

### 前置需求
1. Go 1.22 或更高版本
2. OpenAI API 金鑰

### 設定
1. **配置 API 金鑰**:
   ```bash
   # 從根目錄
   cp .env.example .env
   # 編輯 .env 並添加你的 OPENAI_API_KEY
   ```

2. **安裝依賴項**:
   ```bash
   cd cmd/examples/basic-chat
   go mod tidy
   ```

3. **執行範例**:
   ```bash
   go run main.go
   ```

### 預期輸出
```
🤖 Basic Chat Agent Example
==================================================
✅ OpenAI API key loaded (length: 51)
📝 Creating AI agent...
✅ Agent 'helpful-assistant' created successfully

💬 Starting conversation...
==================================================

🔄 Turn 1/3
👤 User: Hello! How are you doing today?
🤖 Assistant: Hello! I'm doing great, thank you for asking...

🔄 Turn 2/3
👤 User: What's the weather like?
🤖 Assistant: I don't have access to current weather data...

🔄 Turn 3/3
👤 User: Can you help me write a simple Python function to add two numbers?
🤖 Assistant: Certainly! Here's a simple Python function...

==================================================
✅ Conversation completed successfully!
📊 Session Summary:
   • Session ID: basic-chat-1704067200
   • Total messages: 6
   • Created at: 2024-01-01 12:00:00
   • Updated at: 2024-01-01 12:00:15
```

## 重要學習要點

1. **函數式選項模式**: 乾淨且可擴展的配置
2. **錯誤處理**: 強健的錯誤檢查和優雅降級
3. **會話管理**: 自動對話歷史記錄追蹤
4. **日誌策略**: 用於除錯和監控的多層級日誌記錄
5. **環境配置**: 安全的 API 金鑰管理

## 問題排查

### 常見問題

1. **遺失 API 金鑰**
   ```
   ❌ OPENAI_API_KEY environment variable is required
   ```
   - 解決方案: 確保你的 `.env` 檔案包含有效的 OpenAI API 金鑰

2. **匯入錯誤**
   ```
   package basic-chat is not in GOROOT or GOPATH
   ```
   - 解決方案: 從範例目錄執行 `go mod tidy`

3. **網路問題**
   ```
   Failed to get response: connection timeout
   ```
   - 解決方案: 檢查網路連接和 OpenAI API 狀態

### 除錯技巧

1. **啟用除錯日誌**: 範例已包含 `agent.WithDebugLogging()`
2. **檢查會話狀態**: 檢查最後的會話摘要
3. **監控回應時間**: 在日誌中查找異常緩慢的回應
4. **驗證 API 金鑰**: 確保金鑰有足夠的額度和權限

## 下一步

成功執行此範例後：
1. 嘗試 **任務完成** 範例以了解進階條件處理
2. 探索 **計算器工具** 範例以了解函數呼叫
3. 修改對話範例以測試不同場景
4. 實驗不同的模型設定（溫度、最大 token 數）