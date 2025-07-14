# 計算器工具範例

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

此範例展示如何實現自定義工具並將其整合到 AI 助手中，使用數學計算器作為示例。

## 🎯 範例概述

此範例實現了一個**數學助手**，能夠：

- **🔧 自定義工具** - 實現 `agent.Tool` 介面創建計算器工具
- **➕ 數學運算** - 執行基本的數學計算操作
- **🛡️ 錯誤處理** - 安全處理除零等數學錯誤
- **🤖 智能整合** - AI 助手自動選擇並使用工具

## 🏗️ 核心功能

### 1. 計算器工具實現

範例實現了一個完整的計算器工具：

```go
type CalculatorTool struct{}

func (c *CalculatorTool) Name() string {
    return "calculator"
}

func (c *CalculatorTool) Description() string {
    return "Perform basic mathematical calculations"
}

func (c *CalculatorTool) Schema() map[string]any {
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "operation": map[string]any{
                "type": "string",
                "description": "The operation to perform",
                "enum": []string{"add", "subtract", "multiply", "divide", "sqrt"},
            },
            "a": map[string]any{
                "type": "number",
                "description": "First number",
            },
            "b": map[string]any{
                "type": "number", 
                "description": "Second number (not required for sqrt)",
            },
        },
        "required": []string{"operation", "a"},
    }
}
```

### 2. 支持的運算

計算器工具支持以下運算：

- **add** - 加法運算
- **subtract** - 減法運算  
- **multiply** - 乘法運算
- **divide** - 除法運算（包含除零檢查）
- **sqrt** - 平方根運算（包含負數檢查）

### 3. 錯誤處理

工具包含完整的錯誤處理機制：

```go
func (c *CalculatorTool) Execute(ctx context.Context, args map[string]any) (any, error) {
    // 參數驗證
    operation, ok := args["operation"].(string)
    if !ok {
        return nil, fmt.Errorf("operation must be a string")
    }

    // 除零檢查
    if b == 0 {
        return nil, fmt.Errorf("cannot divide by zero")
    }

    // 負數平方根檢查
    if a < 0 {
        return nil, fmt.Errorf("cannot calculate square root of negative number")
    }
    
    // ... 運算邏輯
}
```

## 🚀 執行範例

### 前置需求
1. Go 1.21 或更高版本
2. OpenAI API 金鑰

### 設定與執行
```bash
# 1. 設定環境變數
export OPENAI_API_KEY="your-openai-api-key"

# 2. 執行範例
cd examples/calculator-tool
go run main.go
```

## 📋 測試場景

範例展示三個數學計算場景：

### 1. 加法運算
- **用戶輸入**: "Calculate 15 + 27"
- **工具調用**: calculator(operation="add", a=15, b=27)
- **預期結果**: 42

### 2. 平方根運算
- **用戶輸入**: "What is the square root of 64?"
- **工具調用**: calculator(operation="sqrt", a=64)
- **預期結果**: 8

### 3. 除法運算
- **用戶輸入**: "Divide 144 by 12"
- **工具調用**: calculator(operation="divide", a=144, b=12)
- **預期結果**: 12

## 📊 範例輸出

```
🧮 Calculator Tool Example
===========================

🔄 Calculation 1
👤 User: Calculate 15 + 27
🤖 Assistant: I'll calculate 15 + 27 for you.

Using the calculator tool: 15 + 27 = 42

🔄 Calculation 2
👤 User: What is the square root of 64?
🤖 Assistant: I'll find the square root of 64.

Using the calculator tool: √64 = 8

🔄 Calculation 3
👤 User: Divide 144 by 12
🤖 Assistant: I'll divide 144 by 12 for you.

Using the calculator tool: 144 ÷ 12 = 12

✅ Calculator Tool Example Complete!
```

## 🎓 學習重點

完成此範例後，您將了解：

1. **工具介面** - 如何實現 `agent.Tool` 介面
2. **Schema 設計** - 如何定義工具的參數架構
3. **參數驗證** - 如何安全處理工具輸入
4. **錯誤處理** - 如何優雅處理運算錯誤
5. **工具整合** - 如何將工具整合到助手中

## 🔄 擴展建議

您可以進一步擴展此範例：

1. **更多運算** - 添加三角函數、對數等進階運算
2. **運算歷史** - 追蹤和顯示計算歷史
3. **複數運算** - 支持複數和矩陣運算
4. **單位轉換** - 整合單位轉換功能

## 💡 核心 API

### 工具實現
- `Tool` 介面 - 定義工具的標準介面
- `.Name()` - 工具名稱
- `.Description()` - 工具描述
- `.Schema()` - 參數架構定義
- `.Execute()` - 工具執行邏輯

### 助手整合
- `agent.New()` - 創建新助手
- `.WithTools()` - 添加工具到助手
- `.Build()` - 完成助手建構

### 對話執行
- `.Chat()` - 執行對話回合
- 自動工具選擇和調用
- 結果整合到回應中

此範例展示了如何創建**功能強大、安全可靠**的自定義工具，讓您的 AI 助手能夠執行特定的任務和運算。