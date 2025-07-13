# 任務完成範例

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

此範例演示進階的條件驗證和迭代式資訊收集，使用結構化輸出和 LLM 驅動的完成檢測。

## 概述

任務完成範例展示了：
- **結構化輸出**: 使用 JSON schema 追蹤缺失資訊
- **條件驗證**: LLM 驅動的必填欄位缺失檢測
- **迭代收集**: 多輪對話以收集完整資訊
- **完成檢測**: 當所有條件都滿足時自動設定標誌
- **安全控制**: 最大迭代限制以防止過度使用 token

## 場景：餐廳預訂系統

此範例模擬一個餐廳預訂助手，必須收集：
1. **客戶姓名** (`name`)
2. **電話號碼** (`phone`)
3. **日期** (`date`)
4. **時間** (`time`)
5. **人數** (`party_size`)

代理使用結構化輸出來追蹤進度並確定何時收集了所有必要資訊。

## 程式碼結構

### 核心組件

1. **結構化輸出定義**
   ```go
   type ReservationStatus struct {
       MissingFields  []string          `json:"missing_fields"`
       CollectedInfo  map[string]string `json:"collected_info"`
       CompletionFlag bool              `json:"completion_flag"`
       Message        string            `json:"message"`
       NextStep       string            `json:"next_step,omitempty"`
   }
   ```
   - `MissingFields`: 仍需資訊的陣列
   - `CollectedInfo`: 已收集資料的鍵值對
   - `CompletionFlag`: 指示任務完成的布林值
   - `Message`: 用戶友好的狀態訊息
   - `NextStep`: 下一步行動的可選指導

2. **代理配置**
   ```go
   reservationAgent, err := agent.New(
       agent.WithName("reservation-assistant"),
       agent.WithInstructions(`You are a restaurant reservation assistant...`),
       agent.WithOpenAI(apiKey),
       agent.WithModel("gpt-4"),
       agent.WithModelSettings(&agent.ModelSettings{
           Temperature: floatPtr(0.3), // 較低溫度以獲得一致的結構化輸出
           MaxTokens:   intPtr(800),
       }),
       agent.WithStructuredOutput(&ReservationStatus{}),
       agent.WithSessionStore(agent.NewInMemorySessionStore()),
       agent.WithDebugLogging(),
   )
   ```
   - 較低溫度 (0.3) 以獲得更一致的結構化輸出
   - 結構化輸出自動生成 JSON schema
   - 預訂收集的具體指令

3. **模擬用戶流程**
   ```go
   userInputs := []string{
       "我想要預訂餐廳，我是李先生",                    // 初始不完整請求
       "我的電話是0912345678，想要明天晚上7點",        // 部分資訊
       "4個人",                               // 最後缺失的部分
   }
   ```
   - 在多輪對話中逐步提供資訊
   - 測試代理追蹤部分進度的能力

4. **迭代控制**
   ```go
   maxTurns := 5
   for turn := 0; turn < len(userInputs) && turn < maxTurns; turn++ {
       // 處理每一輪
       response, structuredOutput, err := reservationAgent.Chat(ctx, sessionID, userInput)
       
       if reservationStatus, ok := structuredOutput.(*ReservationStatus); ok {
           if reservationStatus.CompletionFlag {
               log.Printf("COMPLETION[%d]: Task completed successfully!", turn+1)
               break
           }
       }
   }
   ```
   - 最多 5 輪的安全限制
   - 透過 `CompletionFlag` 自動完成檢測
   - 任務完成時優雅退出

## 結構化輸出處理

### 回應分析
```go
if structuredOutput != nil {
    if reservationStatus, ok := structuredOutput.(*ReservationStatus); ok {
        // 顯示狀態
        fmt.Printf("   • Missing fields: %s\n", strings.Join(reservationStatus.MissingFields, ", "))
        fmt.Printf("   • Collected info: %d items\n", len(reservationStatus.CollectedInfo))
        
        // 檢查完成狀態
        if reservationStatus.CompletionFlag {
            fmt.Println("\n🎉 Reservation completed successfully!")
            break
        }
    }
}
```

### 預期流程進展

**第 1 輪**: `"我想要預訂餐廳，我是李先生"`
```json
{
  "missing_fields": ["phone", "date", "time", "party_size"],
  "collected_info": {"name": "李先生"},
  "completion_flag": false,
  "message": "我已經記錄您的姓名。還需要您的電話號碼、用餐日期、時間和人數。"
}
```

**第 2 輪**: `"我的電話是0912345678，想要明天晚上7點"`
```json
{
  "missing_fields": ["party_size"],
  "collected_info": {
    "name": "李先生",
    "phone": "0912345678", 
    "date": "明天",
    "time": "晚上7點"
  },
  "completion_flag": false,
  "message": "很好！最後請告訴我有幾個人用餐？"
}
```

**第 3 輪**: `"4個人"`
```json
{
  "missing_fields": [],
  "collected_info": {
    "name": "李先生",
    "phone": "0912345678",
    "date": "明天", 
    "time": "晚上7點",
    "party_size": "4個人"
  },
  "completion_flag": true,
  "message": "完美！預訂已完成。"
}
```

## 日誌系統

範例提供多層級的詳細日誌記錄：

### 日誌分類
- **REQUEST**: 輸入處理和輪次追蹤
- **RESPONSE**: LLM 回應詳情和時間
- **STRUCTURED**: 結構化輸出解析和驗證
- **PROGRESS**: 缺失欄位追蹤和完成狀態
- **COMPLETION**: 任務完成檢測
- **SESSION**: 對話狀態管理

### 範例日誌輸出
```
🏪 Task Completion Example - Restaurant Reservation
============================================================
✅ OpenAI API key loaded
📝 Creating reservation agent with structured output...
✅ Reservation agent 'reservation-assistant' created successfully

💬 Starting reservation collection process...
============================================================

🔄 Turn 1/3
👤 User: 我想要預訂餐廳，我是李先生
REQUEST[1]: Processing user input
RESPONSE[1]: Duration: 2.1s
STRUCTURED[1]: Parsed reservation status successfully
STRUCTURED[1]: Missing fields: [phone date time party_size]
STRUCTURED[1]: Completion flag: false
PROGRESS[1]: Still missing: phone, date, time, party_size

🔄 Turn 2/3
👤 User: 我的電話是0912345678，想要明天晚上7點
STRUCTURED[2]: Missing fields: [party_size]
PROGRESS[2]: Still missing: party_size

🔄 Turn 3/3
👤 User: 4個人
COMPLETION[3]: Task completed successfully!
🎉 Reservation completed successfully!
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
   cd cmd/examples/task-completion
   go mod tidy
   ```

3. **執行範例**:
   ```bash
   go run main.go
   ```

## 重要學習要點

### 1. 結構化輸出設計
- **清晰的 Schema**: 定義良好的 JSON 結構用於追蹤狀態
- **進度追蹤**: 缺失欄位陣列提供透明度
- **完成檢測**: 布林標誌用於自動終止

### 2. LLM 驅動邏輯
- **條件評估**: LLM 確定缺少哪些資訊
- **動態指令**: 代理根據當前狀態調整提示
- **自然語言處理**: 從對話輸入中提取相關資料

### 3. 安全機制
- **迭代限制**: 防止無限循環和過度使用 token
- **錯誤處理**: 解析失敗時的優雅降級
- **狀態驗證**: 確保結構化輸出符合預期格式

### 4. 後端友好設計
- **會話持久化**: 對話狀態在輪次間保持
- **結構化資料**: 易於與資料庫和 API 整合
- **稽核軌跡**: 完整的日誌記錄用於除錯和分析

## 問題排查

### 常見問題

1. **不一致的結構化輸出**
   - **原因**: 溫度過高或指令不清楚
   - **解決方案**: 降低溫度 (0.1-0.3) 並精煉提示

2. **完成標誌從未設定**
   - **原因**: LLM 未識別完成標準
   - **解決方案**: 在指令中添加明確的完成範例

3. **缺失欄位檢測問題**
   - **原因**: 欄位名稱或需求不明確
   - **解決方案**: 使用清晰、具體的欄位名稱和驗證規則

### 除錯技巧

1. **監控結構化輸出**: 檢查 JSON 解析是否成功
2. **追蹤欄位變化**: 觀察 `missing_fields` 陣列如何演變
3. **驗證指令**: 確保 LLM 理解完成標準
4. **測試邊緣情況**: 嘗試不完整或模糊的用戶輸入

## 客製化

### 適應不同場景

1. **更改必填欄位**:
   ```go
   // 用於酒店預訂
   type BookingStatus struct {
       MissingFields  []string `json:"missing_fields"`
       CollectedInfo  map[string]string `json:"collected_info"`
       CompletionFlag bool `json:"completion_flag"`
       // 添加酒店特定欄位
       RoomType      string `json:"room_type,omitempty"`
       CheckIn       string `json:"check_in,omitempty"`
       CheckOut      string `json:"check_out,omitempty"`
   }
   ```

2. **修改指令**:
   ```go
   agent.WithInstructions(`你是一個酒店預訂助手。收集：
   1. 客人姓名，2. 電話號碼，3. 入住日期，
   4. 退房日期，5. 房間偏好...`)
   ```

3. **調整迭代限制**:
   ```go
   maxTurns := 10 // 用於更複雜的場景
   ```

## 下一步

理解此範例後：
1. 實現你自己的結構化輸出類型
2. 實驗不同的完成標準
3. 為收集的資訊添加驗證邏輯
4. 與外部 API 整合以進行真實預訂
5. 探索 **計算器工具** 範例以了解函數呼叫