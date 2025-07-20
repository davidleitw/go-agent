# LLM 模組

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

LLM 模組提供與語言模型互動的統一介面，初期支援 OpenAI 的 API。

## 功能特色

- **簡潔介面**：乾淨的 `Model` 介面提供同步完成功能
- **工具支援**：內建對函數呼叫/工具的支援
- **類型安全**：強類型的請求和回應
- **可擴展性**：易於新增其他 LLM 提供商的支援
- **配置**：靈活的配置，包含 API 金鑰和自訂端點

## 快速開始

```go
import (
    "github.com/davidleitw/go-agent/llm"
    "github.com/davidleitw/go-agent/llm/openai"
)

// 創建 OpenAI 客戶端
model := openai.New(llm.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
    Model:  "gpt-4",
})

// 發送完成請求
resp, err := model.Complete(ctx, llm.Request{
    Messages: []llm.Message{
        {Role: "system", Content: "你是一個有用的助手"},
        {Role: "user", Content: "天氣如何？"},
    },
})

fmt.Println(resp.Content)
```

## API 參考

### Model 介面

```go
type Model interface {
    Complete(ctx context.Context, request Request) (*Response, error)
    // TODO：串流完成的 Stream 方法
}
```

### 請求結構

```go
type Request struct {
    Messages     []Message         // 對話訊息
    Temperature  *float32          // 可選：0.0-2.0
    MaxTokens    *int             // 可選：生成的最大 token 數
    Tools        []tool.Definition // 可選：可用工具
}
```

### 回應結構

```go
type Response struct {
    Content      string      // 生成的文字
    ToolCalls    []tool.Call // 工具調用（如果有）
    Usage        Usage       // Token 使用統計
    FinishReason string      // stop/length/tool_calls
}
```

## 使用工具

```go
// 定義工具
weatherTool := tool.Definition{
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
            },
            Required: []string{"location"},
        },
    },
}

// 在請求中包含工具
resp, err := model.Complete(ctx, llm.Request{
    Messages: messages,
    Tools:    []tool.Definition{weatherTool},
})

// 處理工具呼叫
if len(resp.ToolCalls) > 0 {
    for _, call := range resp.ToolCalls {
        // 執行工具並繼續對話
        result := executeWeatherTool(call.Function.Arguments)
        
        // 將工具回應新增到訊息中
        messages = append(messages, llm.Message{
            Role:       "tool",
            Content:    result,
            Name:       call.Function.Name,
            ToolCallID: call.ID,
        })
    }
}
```

## 配置

### 基本配置

```go
config := llm.Config{
    APIKey: "your-api-key",
    Model:  "gpt-4",
}
```

### 使用自訂端點

```go
config := llm.Config{
    APIKey:  "your-api-key",
    Model:   "gpt-4",
    BaseURL: "https://your-proxy.com/v1", // 用於代理或自訂端點
}
```

## Token 使用量

追蹤 token 消耗以進行成本管理：

```go
resp, _ := model.Complete(ctx, request)

fmt.Printf("提示 tokens：%d\n", resp.Usage.PromptTokens)
fmt.Printf("完成 tokens：%d\n", resp.Usage.CompletionTokens)
fmt.Printf("總 tokens：%d\n", resp.Usage.TotalTokens)
```

## 錯誤處理

```go
resp, err := model.Complete(ctx, request)
if err != nil {
    // 處理 API 錯誤
    log.Printf("LLM 錯誤：%v", err)
    return
}

// 檢查完成原因
switch resp.FinishReason {
case "stop":
    // 正常完成
case "length":
    // 達到最大 tokens 限制
case "tool_calls":
    // 模型想要使用工具
default:
    // 意外的完成原因
}
```

## 未來增強功能

以下功能計劃在未來版本中實現：

- **串流支援**：即時 token 串流
- **額外參數**：TopP、停止序列、頻率懲罰
- **多模態支援**：圖像輸入
- **提供商擴展**：Anthropic、Google 和其他提供商
- **回應驗證**：回應的 schema 驗證
- **重試邏輯**：指數退避的自動重試
- **速率限制**：內建速率限制處理

## 測試

模組包含全面的測試：

```bash
go test ./llm/... -v
```

與實際 API 的整合測試：

```go
// 在測試中使用 mock
type mockModel struct{}

func (m *mockModel) Complete(ctx context.Context, req llm.Request) (*llm.Response, error) {
    return &llm.Response{
        Content: "Mock 回應",
        Usage: llm.Usage{
            TotalTokens: 10,
        },
    }, nil
}
```

## 最佳實踐

1. **API 金鑰安全性**：絕不硬編碼 API 金鑰，使用環境變數
2. **Context 使用**：總是傳遞 context 以支援取消
3. **錯誤處理**：檢查錯誤和完成原因
4. **Token 限制**：設定合理的 MaxTokens 以控制成本
5. **Temperature**：對確定性輸出使用較低的 temperature（0.0-0.3）

## 授權

MIT 授權 - 請參閱專案根目錄中的 LICENSE 檔案。