# 多工具代理範例

此範例展示如何建立一個可以使用多種工具執行不同任務的 AI 代理。代理可以獲取天氣資訊、執行計算和取得時間資料。

## 你將學到

- 如何建立具有不同用途的多個工具
- 如何讓代理存取多個工具
- 代理如何智慧地為每個任務選擇正確的工具
- 工具如何在單一對話中協同工作

## 包含的工具

### 1. 天氣工具
- **用途**：獲取任何地點的天氣資訊
- **使用方式**：「東京的天氣如何？」
- **回傳**：溫度、天氣狀況、濕度、風速

### 2. 計算器工具  
- **用途**：執行數學計算
- **操作**：加法、減法、乘法、除法、平方根
- **使用方式**：「計算 15 + 7」或「144 的平方根是多少？」

### 3. 時間工具
- **用途**：獲取目前時間和日期資訊
- **功能**：不同時區、星期幾、unix 時間戳
- **使用方式**：「紐約現在幾點？」

## 執行範例

確保你已經設定好 OpenAI API key：

```bash
# 複製環境變數檔案
cp ../../.env.example .env

# 編輯 .env 並新增你的 OpenAI API key
```

執行範例：

```bash
go run main.go
```

## 展示的主要功能

### 1. 工具定義
每個工具都實作 `agent.Tool` 介面：

```go
type WeatherTool struct{}

func (w *WeatherTool) Name() string {
    return "get_weather"
}

func (w *WeatherTool) Description() string {
    return "Get current weather information for a specified location"
}

func (w *WeatherTool) Schema() map[string]any {
    // 參數的 JSON Schema 定義
}

func (w *WeatherTool) Execute(ctx context.Context, args map[string]any) (any, error) {
    // 工具實作
}
```

### 2. 具有多個工具的代理
```go
agent, err := agent.NewBasicAgent(agent.BasicAgentConfig{
    Name:        "multi-tool-assistant",
    Description: "具有天氣、計算器和時間工具存取權限的 AI 助手",
    Instructions: `你是一個有用的助手，可以存取多個工具...`,
    Model:       "gpt-4o-mini",
    Tools:       []agent.Tool{weatherTool, calculatorTool, timeTool},
    ChatModel:   chatModel,
})
```

### 3. 智慧工具選擇
代理會自動：
- 分析使用者請求
- 選擇適當的工具
- 在需要時使用多個工具
- 提供全面的回應

## 優勢

- **模組化**：每個工具都有單一、明確的用途
- **可重用性**：工具可以在不同代理之間使用
- **可擴展性**：容易新增新工具
- **智慧性**：代理根據上下文自動選擇工具

## 下一步

嘗試修改範例：
1. 新增一個新工具（例如：貨幣轉換器、單位轉換器）
2. 修改現有工具以新增更多功能
3. 建立具有不同工具組合的代理
4. 實驗不同的指示來改變行為

此範例展示了使用 go-agent 框架建立強大、多功能 AI 代理是多麼簡單！