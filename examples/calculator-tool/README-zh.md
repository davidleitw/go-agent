# 計算器工具範例

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

此範例演示自定義工具實現和 OpenAI 函數呼叫整合，使用數學計算器工具。

## 概述

計算器工具範例展示了：
- **自定義工具實現**: 創建實現 `agent.Tool` 介面的工具
- **OpenAI 函數呼叫**: 與 OpenAI 原生函數呼叫機制整合
- **參數驗證**: 強健的輸入驗證和類型轉換
- **結構化結果**: 詳細的計算步驟和結果
- **錯誤處理**: 優雅處理數學錯誤（除零等）

## 場景：數學助手

此範例創建一個能使用自定義計算器工具執行各種計算的數學助手：
- **基本算術**: 加法、減法、乘法、除法
- **進階運算**: 乘方（指數）、開方
- **詳細步驟**: 每個計算都包含逐步分解
- **錯誤預防**: 對無效運算的驗證

## 程式碼結構

### 核心組件

1. **工具結果結構**
   ```go
   type CalculationResult struct {
       Expression    string    `json:"expression"`
       Result        float64   `json:"result"`
       Steps         []string  `json:"steps"`
       OperationType string    `json:"operation_type"`
       Timestamp     time.Time `json:"timestamp"`
   }
   ```
   - `Expression`: 人類可讀的數學表達式
   - `Result`: 計算的數值結果
   - `Steps`: 計算步驟陣列，提供透明度
   - `OperationType`: 執行的數學運算類型
   - `Timestamp`: 執行計算的時間

2. **自定義工具實現**
   ```go
   type CalculatorTool struct{}
   
   func (t *CalculatorTool) Name() string {
       return "calculator"
   }
   
   func (t *CalculatorTool) Description() string {
       return "Perform mathematical calculations including basic arithmetic, powers, and square roots"
   }
   
   func (t *CalculatorTool) Schema() map[string]any {
       return map[string]any{
           "type": "object",
           "properties": map[string]any{
               "operation": map[string]any{
                   "type": "string",
                   "enum": []string{"add", "subtract", "multiply", "divide", "power", "sqrt"},
               },
               "operand1": map[string]any{
                   "type": "number",
                   "description": "The first number",
               },
               "operand2": map[string]any{
                   "type": "number", 
                   "description": "The second number (not required for sqrt)",
               },
           },
           "required": []string{"operation", "operand1"},
       }
   }
   ```
   - 實現 `agent.Tool` 介面的四個必需方法
   - 為 OpenAI 函數呼叫提供 JSON schema
   - 定義支援的運算和參數

3. **工具執行邏輯**
   ```go
   func (t *CalculatorTool) Execute(ctx context.Context, args map[string]any) (any, error) {
       // 輸入驗證和類型轉換
       operation := args["operation"].(string)
       operand1, _ := convertToFloat64(args["operand1"])
       
       // 運算特定邏輯
       switch operation {
       case "add":
           result = operand1 + operand2
           expression = fmt.Sprintf("%.2f + %.2f", operand1, operand2)
           steps = []string{
               fmt.Sprintf("Addition: %.2f + %.2f", operand1, operand2),
               fmt.Sprintf("Result: %.2f", result),
           }
       // ... 其他運算
       }
       
       return CalculationResult{...}, nil
   }
   ```
   - 強健的類型轉換和驗證
   - 運算特定的計算邏輯
   - 詳細的逐步分解
   - 邊緣情況的錯誤處理

4. **代理配置**
   ```go
   mathAssistant, err := agent.New(
       agent.WithName("math-assistant"),
       agent.WithInstructions(`You are a helpful math assistant...`),
       agent.WithOpenAI(apiKey),
       agent.WithModel("gpt-4"),
       agent.WithModelSettings(&agent.ModelSettings{
           Temperature: floatPtr(0.1), // 低溫度以獲得精確計算
           MaxTokens:   intPtr(1000),
       }),
       agent.WithTools(calculatorTool),
       agent.WithSessionStore(agent.NewInMemorySessionStore()),
       agent.WithDebugLogging(),
   )
   ```
   - 非常低的溫度 (0.1) 以獲得數學精確性
   - 數學助手的特定指令
   - 計算器工具註冊

## 支援的運算

### 1. 加法 (`add`)
```go
// 輸入: 15 + 27
{
  "operation": "add",
  "operand1": 15,
  "operand2": 27
}

// 輸出:
{
  "expression": "15.00 + 27.00",
  "result": 42,
  "steps": [
    "Addition: 15.00 + 27.00",
    "Result: 42.00"
  ],
  "operation_type": "add"
}
```

### 2. 減法 (`subtract`)
```go
// 輸入: 125 - 47
{
  "operation": "subtract", 
  "operand1": 125,
  "operand2": 47
}
```

### 3. 乘法 (`multiply`)
```go
// 輸入: 13 × 7
{
  "operation": "multiply",
  "operand1": 13,
  "operand2": 7
}
```

### 4. 除法 (`divide`)
```go
// 輸入: 144 ÷ 12
{
  "operation": "divide",
  "operand1": 144,
  "operand2": 12
}
// 注意: 包含除零保護
```

### 5. 乘方 (`power`)
```go
// 輸入: 2^8
{
  "operation": "power",
  "operand1": 2,
  "operand2": 8
}
```

### 6. 開方 (`sqrt`)
```go
// 輸入: √64
{
  "operation": "sqrt",
  "operand1": 64
}
// 注意: 只需要 operand1，驗證非負輸入
```

## 函數呼叫流程

### 1. 用戶輸入處理
```
用戶: "計算 15 + 27"
↓
代理解釋自然語言
↓ 
代理決定使用計算器工具
↓
生成函數呼叫
```

### 2. 工具調用
```go
// OpenAI 生成此函數呼叫:
{
  "name": "calculator",
  "arguments": {
    "operation": "add",
    "operand1": 15,
    "operand2": 27
  }
}
```

### 3. 執行和回應
```go
// 工具執行並返回:
CalculationResult{
  Expression: "15.00 + 27.00",
  Result: 42.0,
  Steps: ["Addition: 15.00 + 27.00", "Result: 42.00"],
  OperationType: "add",
}

// 代理格式化回應:
"15 + 27 的結果是 42。以下是我的計算過程:
1. 加法: 15.00 + 27.00  
2. 結果: 42.00"
```

## 日誌系統

範例為工具執行提供全面的日誌記錄：

### 日誌分類
- **TOOL**: 工具執行詳情和時間
- **TOOLCALL**: 函數呼叫參數和結果
- **REQUEST**: 用戶輸入和計算請求
- **RESPONSE**: 代理回應和工具整合
- **SESSION**: 對話和工具使用追蹤

### 範例日誌輸出
```
🧮 Calculator Tool Example
==================================================
✅ OpenAI API key loaded
🛠️  Creating calculator tool...
📝 Creating math assistant agent...
✅ Math assistant 'math-assistant' created with calculator tool

🧮 Starting calculator demonstrations...
==================================================

🔄 Calculation 1/6
👤 User: Calculate 15 + 27
REQUEST[1]: Processing calculation request
TOOL: Calculator tool execution started
TOOL: Input arguments: map[operand1:15 operand2:27 operation:add]
TOOL: Operation: add, Operand1: 15.000000, Operand2: 27.000000 (has: true)
TOOL: Calculation completed successfully
TOOL: Expression: 15.00 + 27.00
TOOL: Result: 42.00
RESPONSE[1]: Duration: 2.3s
RESPONSE[1]: Tool calls: 1
🤖 Assistant: I'll calculate 15 + 27 for you using the calculator.

The result is **42**.

Here's the calculation breakdown:
- Addition: 15.00 + 27.00
- Result: 42.00

🔧 Tool Calls:
   • Tool: calculator
   • Arguments: {"operation":"add","operand1":15,"operand2":27}
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
   cd cmd/examples/calculator-tool
   go mod tidy
   ```

3. **執行範例**:
   ```bash
   go run main.go
   ```

### 預期輸出
範例將執行 6 個預定義的計算：
1. 計算 15 + 27
2. 144 除以 12 是什麼？
3. 找出 64 的平方根
4. 計算 2 的 8 次方
5. 125 - 47 是什麼？
6. 13 乘以 7

每個計算將顯示：
- 用戶請求
- 代理回應與解釋
- 工具呼叫詳情
- 執行時間和會話統計

## 重要學習要點

### 1. 工具介面實現
- **方法需求**: 必須實現所有四個介面方法
- **Schema 設計**: JSON schema 為 OpenAI 定義參數
- **類型安全**: 強健的類型轉換和驗證

### 2. OpenAI 函數呼叫
- **自動整合**: 框架處理 OpenAI 函數呼叫協議
- **參數映射**: 參數自動從 JSON 映射到 Go 類型
- **回應格式化**: 工具結果整合到對話流程中

### 3. 錯誤處理策略
- **輸入驗證**: 檢查類型和必需參數
- **數學錯誤**: 處理除零、負數開方
- **優雅降級**: 當個別計算失敗時繼續運作

### 4. 結構化結果
- **詳細輸出**: 包含步驟、表達式和元資料
- **稽核軌跡**: 時間戳和運算追蹤
- **用戶體驗**: 清晰的計算解釋

## 問題排查

### 常見問題

1. **工具未被呼叫**
   - **原因**: 指令不清楚或 schema 問題
   - **解決方案**: 精煉工具描述並確保 schema 有效

2. **類型轉換錯誤**
   ```
   TOOL: ERROR - operand1 conversion failed
   ```
   - **原因**: OpenAI 發送意外的資料類型
   - **解決方案**: 實現強健的 `convertToFloat64` 函數

3. **數學錯誤**
   ```
   division by zero is not allowed
   ```
   - **原因**: 無效的數學運算
   - **解決方案**: 在工具執行中添加驗證邏輯

### 除錯技巧

1. **監控工具呼叫**: 檢查是否生成函數呼叫
2. **驗證參數**: 確保 OpenAI 發送正確參數
3. **測試邊緣情況**: 嘗試無效輸入和邊緣情況
4. **Schema 驗證**: 驗證 JSON schema 符合預期

## 客製化

### 新增新運算

1. **擴展 Schema**:
   ```go
   "enum": []string{"add", "subtract", "multiply", "divide", "power", "sqrt", "sin", "cos", "log"}
   ```

2. **實現運算**:
   ```go
   case "sin":
       result = math.Sin(operand1)
       expression = fmt.Sprintf("sin(%.2f)", operand1)
       steps = []string{
           fmt.Sprintf("Sine of %.2f radians", operand1),
           fmt.Sprintf("Result: %.6f", result),
       }
   ```

### 創建不同的工具

```go
type WeatherTool struct{}

func (t *WeatherTool) Name() string { return "get_weather" }
func (t *WeatherTool) Description() string { return "Get current weather for a location" }
func (t *WeatherTool) Schema() map[string]any { /* ... */ }
func (t *WeatherTool) Execute(ctx context.Context, args map[string]any) (any, error) {
    // 天氣 API 整合
}
```

## 下一步

理解此範例後：
1. 創建你自己的自定義工具
2. 實驗不同的參數類型
3. 與外部 API 整合
4. 添加更複雜的錯誤處理
5. 探索具有複雜工作流程的多工具場景