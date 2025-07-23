# Multi-Tool Agent Demo

這個例子展示了 go-agent 框架的多工具使用能力，包含了詳細的 logging 來追蹤整個執行過程。

## 功能特色

### 🛠️ 包含的工具

1. **WeatherTool** 🌤️
   - 查詢城市天氣資訊
   - 支援攝氏/華氏溫度
   - 模擬真實 API 延遲

2. **CalculatorTool** 🧮
   - 基本算術運算 (+, -, *, /)
   - 數學函數 (sqrt, sin, cos)
   - 表達式計算

3. **TimeTool** ⏰
   - 當前時間查詢
   - 多時區轉換
   - 時間格式化

4. **FileWriteTool** 📝
   - 檔案寫入
   - 支援附加模式
   - 自動建立檔案

### 🔍 詳細 Logging

- LLM 調用追蹤
- 工具執行過程
- 錯誤處理記錄
- 性能指標統計
- Session 狀態管理

## 如何執行

```bash
# 設定 OpenAI API Key (必須)
export OPENAI_API_KEY="your-api-key-here"

# 在 multi-tool-demo 目錄下
go mod tidy

# 方法 1: 直接執行 main.go （推薦）
go run main.go

# 方法 2: 執行整個模組
go run .

# 方法 3: 先建構再執行
go build -o multi-tool-demo
./multi-tool-demo

# 或使用提供的腳本
./run_demo.sh
```

## ✅ 單檔案設計

**簡化設計：** 所有代碼都整合在 `main.go` 中，包括：

- 4 個工具實作 (WeatherTool, CalculatorTool, TimeTool, FileWriteTool)
- OpenAI API 整合與 LoggingModel 包裝
- 主程式邏輯
- 詳細 logging

現在可以直接使用 `go run main.go` 執行！

## 測試範例

啟動後，可以試試以下指令：

### 天氣查詢
```
What's the weather in Tokyo?
Show me the temperature in Taipei
How's the weather in London in fahrenheit?
```

### 數學計算
```
Calculate 15 + 25
What is 5 * 7?
Calculate sqrt(16)
What's sqrt(25)?
```

### 時間資訊
```
What time is it?
Show me different timezones
What time is it in Tokyo?
```

### 檔案操作
```
Write a file with some content
Create a file for me
```

### 對話測試
```
Hello, what can you do?
Help me with calculations
```

## Logging 說明

### 日誌級別標識

- 🤖 `[OpenAI]` - OpenAI API 調用操作
- 🌤️ `[WeatherTool]` - 天氣工具
- 🧮 `[CalculatorTool]` - 計算工具  
- ⏰ `[TimeTool]` - 時間工具
- 📝 `[FileWriteTool]` - 檔案工具
- 👤 用戶輸入
- ✅ 成功操作
- ❌ 錯誤訊息
- ⚠️ 警告訊息

### 追蹤資訊

1. **LLM 調用詳情**
   ```
   🤖 [OpenAI] === LLM CALL #1 ===
   🤖 [OpenAI] Request contains 2 messages
   🤖 [OpenAI] Available tools: 4
   ```

2. **工具執行過程**
   ```
   🌤️ [WeatherTool] Starting weather lookup...
   🌤️ [WeatherTool] Looking up weather for Tokyo (units: celsius)
   ✅ [WeatherTool] Successfully retrieved weather
   ```

3. **性能統計**
   ```
   📊 Usage stats:
      - LLM tokens: 170 (prompt: 120, completion: 50)
      - Tool calls: 1
      - Session writes: 1
   ```

## 程式架構

```
main.go
├── OpenAI Integration (真實 LLM)
│   ├── API 調用管理
│   ├── LoggingModel 包裝
│   └── 詳細請求/回應記錄
├── WeatherTool
├── CalculatorTool  
├── TimeTool
├── FileWriteTool
└── Agent 整合
    ├── Builder Pattern
    ├── Context Providers
    └── Session Management
```

## 開發用途

這個例子可以用來：

1. **測試工具整合** - 驗證新工具是否正確集成
2. **調試 Agent 行為** - 透過詳細 log 理解執行流程
3. **性能測試** - 監控 token 使用量和執行時間
4. **功能演示** - 展示 go-agent 的完整能力
5. **學習參考** - 理解如何建構多工具 Agent

## 注意事項

- **需要 OpenAI API Key** - 會消耗實際 API 配額，請注意使用量
- 產生的檔案會儲存在當前目錄
- 所有 API 調用和工具操作都有詳細 logging
- 支援持續對話和 session 管理
- 建議使用 GPT-3.5 或 GPT-4 進行測試

## 擴展建議

可以嘗試：

1. 加入更多實用工具
2. 測試不同的 OpenAI 模型 (GPT-3.5, GPT-4)
3. 增加錯誤處理測試
4. 實作更複雜的工具鏈
5. 加入記憶體和知識庫整合
6. 實作 token 使用量限制和成本控制