# Tool 模組

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

Tool 模組為在 Go Agent Framework 中定義和執行可被語言模型使用的工具提供框架。

## 功能特色

- **簡潔的工具介面**：易於實作自訂工具
- **類型安全定義**：基於 JSON Schema 的參數定義
- **註冊表管理**：工具註冊和執行的中央註冊表
- **執行緒安全**：支援並發存取
- **可擴展性**：為未來增強功能（如驗證和 MCP 支援）做好準備

## 快速開始

```go
import (
    "github.com/davidleitw/go-agent/tool"
)

// 定義自訂工具
type WeatherTool struct{}

func (w *WeatherTool) Definition() tool.Definition {
    return tool.Definition{
        Type: "function",
        Function: tool.Function{
            Name:        "get_weather",
            Description: "取得某地點的目前天氣",
            Parameters: tool.Parameters{
                Type: "object",
                Properties: map[string]tool.Property{
                    "location": {
                        Type:        "string",
                        Description: "城市名稱",
                    },
                    "unit": {
                        Type:        "string",
                        Description: "溫度單位（攝氏/華氏）",
                    },
                },
                Required: []string{"location"},
            },
        },
    }
}

func (w *WeatherTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    location, _ := params["location"].(string)
    unit, _ := params["unit"].(string)
    if unit == "" {
        unit = "celsius"
    }
    
    // 實作天氣取得邏輯
    return map[string]any{
        "temperature": 22,
        "unit":        unit,
        "description": "晴朗",
    }, nil
}

// 使用工具
registry := tool.NewRegistry()
registry.Register(&WeatherTool{})
```

## API 參考

### 工具介面

```go
type Tool interface {
    Definition() Definition
    Execute(ctx context.Context, params map[string]any) (any, error)
}
```

### 定義結構

```go
type Definition struct {
    Type     string   // 目前總是 "function"
    Function Function
}

type Function struct {
    Name        string
    Description string
    Parameters  Parameters
}
```

### 參數（JSON Schema 子集）

```go
type Parameters struct {
    Type       string              // "object"
    Properties map[string]Property
    Required   []string
}

type Property struct {
    Type        string // string/number/boolean/array/object
    Description string
}
```

## 註冊表使用

### 創建和管理註冊表

```go
// 創建註冊表
registry := tool.NewRegistry()

// 註冊工具
err := registry.Register(tool1)
err := registry.Register(tool2)

// 取得所有定義（用於 LLM）
definitions := registry.GetDefinitions()

// 取得特定工具
weatherTool, exists := registry.Get("get_weather")

// 清除所有工具
registry.Clear()
```

### 執行工具

```go
// 執行來自 LLM 的工具呼叫
call := tool.Call{
    ID: "call_123",
    Function: tool.FunctionCall{
        Name:      "get_weather",
        Arguments: `{"location": "東京", "unit": "celsius"}`,
    },
}

result, err := registry.Execute(ctx, call)
if err != nil {
    log.Printf("工具執行失敗：%v", err)
}
```

## 創建自訂工具

### 簡單計算機工具

```go
type CalculatorTool struct{}

func (c *CalculatorTool) Definition() tool.Definition {
    return tool.Definition{
        Type: "function",
        Function: tool.Function{
            Name:        "calculator",
            Description: "執行基本算術運算",
            Parameters: tool.Parameters{
                Type: "object",
                Properties: map[string]tool.Property{
                    "operation": {
                        Type:        "string",
                        Description: "要執行的運算（加/減/乘/除）",
                    },
                    "a": {
                        Type:        "number",
                        Description: "第一個數字",
                    },
                    "b": {
                        Type:        "number",
                        Description: "第二個數字",
                    },
                },
                Required: []string{"operation", "a", "b"},
            },
        },
    }
}

func (c *CalculatorTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    op, _ := params["operation"].(string)
    a, _ := params["a"].(float64)
    b, _ := params["b"].(float64)
    
    switch op {
    case "add":
        return a + b, nil
    case "subtract":
        return a - b, nil
    case "multiply":
        return a * b, nil
    case "divide":
        if b == 0 {
            return nil, errors.New("除零錯誤")
        }
        return a / b, nil
    default:
        return nil, fmt.Errorf("未知運算：%s", op)
    }
}
```

### 資料庫查詢工具

```go
type DatabaseTool struct {
    db *sql.DB
}

func (d *DatabaseTool) Definition() tool.Definition {
    return tool.Definition{
        Type: "function",
        Function: tool.Function{
            Name:        "query_database",
            Description: "從資料庫查詢用戶資訊",
            Parameters: tool.Parameters{
                Type: "object",
                Properties: map[string]tool.Property{
                    "user_id": {
                        Type:        "string",
                        Description: "要查詢的用戶 ID",
                    },
                },
                Required: []string{"user_id"},
            },
        },
    }
}

func (d *DatabaseTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    userID, _ := params["user_id"].(string)
    
    var user struct {
        ID    string
        Name  string
        Email string
    }
    
    err := d.db.QueryRowContext(ctx, 
        "SELECT id, name, email FROM users WHERE id = ?", 
        userID,
    ).Scan(&user.ID, &user.Name, &user.Email)
    
    if err != nil {
        return nil, err
    }
    
    return user, nil
}
```

## 錯誤處理

工具應該返回有意義的錯誤：

```go
func (t *MyTool) Execute(ctx context.Context, params map[string]any) (any, error) {
    // 驗證必要參數
    value, ok := params["required_param"].(string)
    if !ok {
        return nil, fmt.Errorf("required_param 必須是字串")
    }
    
    // 處理上下文取消
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // 業務邏輯錯誤
    result, err := doSomething(value)
    if err != nil {
        return nil, fmt.Errorf("操作失敗：%w", err)
    }
    
    return result, nil
}
```

## 執行緒安全

註冊表對並發存取是執行緒安全的：

```go
// 從多個 goroutine 安全使用
go func() {
    registry.Register(tool1)
}()

go func() {
    result, _ := registry.Execute(ctx, call)
}()

go func() {
    definitions := registry.GetDefinitions()
}()
```

## 未來增強功能

以下功能正在計劃中：

- **輸入驗證**：針對 schema 的自動參數驗證
- **輸出 Schema**：可選的輸出驗證
- **MCP 支援**：與 Model Context Protocol 整合
- **非同步執行**：支援長時間運行的工具
- **中介軟體**：用於記錄、指標等的攔截器
- **工具組合**：組合多個工具
- **速率限制**：每個工具的速率限制

## 測試

### 單元測試工具

```go
func TestMyTool(t *testing.T) {
    tool := &MyTool{}
    
    // 測試定義
    def := tool.Definition()
    if def.Function.Name != "my_tool" {
        t.Errorf("預期名稱 my_tool，得到 %s", def.Function.Name)
    }
    
    // 測試執行
    params := map[string]any{
        "param1": "value1",
    }
    
    result, err := tool.Execute(context.Background(), params)
    if err != nil {
        t.Errorf("預期無錯誤，得到 %v", err)
    }
    
    // 驗證結果
    if result != expectedResult {
        t.Errorf("預期 %v，得到 %v", expectedResult, result)
    }
}
```

### 與註冊表測試

```go
func TestToolIntegration(t *testing.T) {
    registry := tool.NewRegistry()
    registry.Register(&MyTool{})
    
    call := tool.Call{
        ID: "test-call",
        Function: tool.FunctionCall{
            Name:      "my_tool",
            Arguments: `{"param1": "value1"}`,
        },
    }
    
    result, err := registry.Execute(context.Background(), call)
    // 斷言結果
}
```

## 最佳實踐

1. **清晰描述**：為 LLM 理解編寫清晰、簡潔的工具描述
2. **參數類型**：使用適當的 JSON Schema 類型
3. **錯誤訊息**：返回有用的錯誤訊息
4. **上下文尊重**：總是尊重上下文取消
5. **冪等性**：儘可能讓工具具有冪等性
6. **記錄**：為除錯新增適當的記錄
7. **文檔**：記錄預期的輸入和輸出

## 授權

MIT 授權 - 請參閱專案根目錄中的 LICENSE 檔案。