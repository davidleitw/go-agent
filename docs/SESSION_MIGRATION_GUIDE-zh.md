# Session API 遷移指南

此指南協助您從舊的 Session 介面遷移到 v0.0.2 引入的新簡化 Session 介面。

## 概覽

Session 介面已大幅簡化，以提升效能、易用性和可維護性。新介面將方法數量從 12 個減少到 5 個核心方法，同時保持所有必要功能。

## API 變更摘要

### 舊介面 (v0.0.1)
```go
type Session interface {
    ID() string
    Messages() []Message
    AddMessage(msg Message)
    AddUserMessage(content string) Message
    AddAssistantMessage(content string) Message
    AddSystemMessage(content string) Message
    AddToolMessage(toolCallID, toolName, content string) Message
    GetLastMessage() *Message
    GetMessagesByRole(role string) []Message
    MessageCount() int
    Clear()
    Clone() Session
    // ... 加上資料儲存方法
}
```

### 新介面 (v0.0.2+)
```go
type Session interface {
    ID() string
    Messages() []Message
    AddMessage(role, content string) Message
    GetData(key string) interface{}
    SetData(key string, value interface{})
}
```

## 遷移步驟

### 1. 更新訊息新增

**舊 API：**
```go
// 角色特定方法
session.AddUserMessage("你好！")
session.AddAssistantMessage("嗨！")
session.AddSystemMessage("使用者已驗證")
session.AddToolMessage("call_123", "calculator", `{"result": 42}`)

// 使用 Message 物件的通用方法
userMsg := agent.NewUserMessage("你好！")
session.AddMessage(userMsg)
```

**新 API：**
```go
// 單一方法使用角色參數
session.AddMessage(agent.RoleUser, "你好！")
session.AddMessage(agent.RoleAssistant, "嗨！")
session.AddMessage(agent.RoleSystem, "使用者已驗證")
session.AddMessage(agent.RoleTool, `{"result": 42}`)
```

### 2. 更新訊息取得

**舊 API：**
```go
// 取得最後一則訊息
lastMsg := session.GetLastMessage()

// 依角色取得訊息
userMessages := session.GetMessagesByRole(agent.RoleUser)

// 取得訊息數量
count := session.MessageCount()
```

**新 API：**
```go
// 取得所有訊息並處理
messages := session.Messages()

// 取得最後一則訊息
var lastMsg *agent.Message
if len(messages) > 0 {
    lastMsg = &messages[len(messages)-1]
}

// 依角色篩選訊息
var userMessages []agent.Message
for _, msg := range messages {
    if msg.Role == agent.RoleUser {
        userMessages = append(userMessages, msg)
    }
}

// 取得訊息數量
count := len(session.Messages())
```

### 3. 更新 Session 操作

**舊 API：**
```go
// 清除 session
session.Clear()

// 複製 session
clonedSession := session.Clone()
```

**新 API：**
```go
// 清除 session（使用型別斷言存取進階功能）
if clearable, ok := session.(interface{ ClearMessages() }); ok {
    clearable.ClearMessages()
}

// 複製 session（使用型別斷言存取進階功能）
if cloneable, ok := session.(interface{ Clone() Session }); ok {
    clonedSession := cloneable.Clone()
}
```

### 4. 更新資料儲存

**舊 API：**
```go
// 資料儲存通常是自訂或大型 session 物件的一部分
// 沒有標準化的 session 資料儲存方式
```

**新 API：**
```go
// 內建資料儲存
session.SetData("user_id", "user_12345")
session.SetData("context", map[string]string{"theme": "dark"})

// 取得資料
userID := session.GetData("user_id").(string)
context := session.GetData("context").(map[string]string)
```

## 遷移範例

### 範例 1：基本聊天應用程式

**舊程式碼：**
```go
func handleUserInput(session agent.Session, input string) error {
    // 新增使用者訊息
    session.AddUserMessage(input)
    
    // 使用 AI 處理
    response := generateAIResponse(session.Messages())
    
    // 新增 AI 回應
    session.AddAssistantMessage(response)
    
    // 檢查訊息數量
    if session.MessageCount() > 10 {
        session.Clear()
    }
    
    return nil
}
```

**新程式碼：**
```go
func handleUserInput(session agent.Session, input string) error {
    // 新增使用者訊息
    session.AddMessage(agent.RoleUser, input)
    
    // 使用 AI 處理
    response := generateAIResponse(session.Messages())
    
    // 新增 AI 回應
    session.AddMessage(agent.RoleAssistant, response)
    
    // 檢查訊息數量
    if len(session.Messages()) > 10 {
        if clearable, ok := session.(interface{ ClearMessages() }); ok {
            clearable.ClearMessages()
        }
    }
    
    return nil
}
```

### 範例 2：工具整合

**舊程式碼：**
```go
func executeTool(session agent.Session, toolCall agent.ToolCall) error {
    result, err := callExternalAPI(toolCall.Function.Arguments)
    if err != nil {
        session.AddToolMessage(toolCall.ID, toolCall.Function.Name, 
            fmt.Sprintf("Error: %v", err))
        return err
    }
    
    resultJSON, _ := json.Marshal(result)
    session.AddToolMessage(toolCall.ID, toolCall.Function.Name, string(resultJSON))
    return nil
}
```

**新程式碼：**
```go
func executeTool(session agent.Session, toolCall agent.ToolCall) error {
    result, err := callExternalAPI(toolCall.Function.Arguments)
    if err != nil {
        session.AddMessage(agent.RoleTool, fmt.Sprintf("Error: %v", err))
        return err
    }
    
    resultJSON, _ := json.Marshal(result)
    session.AddMessage(agent.RoleTool, string(resultJSON))
    return nil
}
```

## 輔助函數

為了簡化遷移，您可以建立提供舊 API 語意的輔助函數：

```go
// 向後相容的輔助函數
func AddUserMessage(session agent.Session, content string) agent.Message {
    return session.AddMessage(agent.RoleUser, content)
}

func AddAssistantMessage(session agent.Session, content string) agent.Message {
    return session.AddMessage(agent.RoleAssistant, content)
}

func AddSystemMessage(session agent.Session, content string) agent.Message {
    return session.AddMessage(agent.RoleSystem, content)
}

func GetLastMessage(session agent.Session) *agent.Message {
    messages := session.Messages()
    if len(messages) == 0 {
        return nil
    }
    return &messages[len(messages)-1]
}

func GetMessagesByRole(session agent.Session, role string) []agent.Message {
    var filtered []agent.Message
    for _, msg := range session.Messages() {
        if msg.Role == role {
            filtered = append(filtered, msg)
        }
    }
    return filtered
}

func MessageCount(session agent.Session) int {
    return len(session.Messages())
}
```

## 新 API 的優勢

### 1. 簡化的介面
- 從 12 個方法減少到 5 個核心方法
- 更容易理解和實作
- 減少認知負擔

### 2. 更好的效能
- 優化的訊息儲存和取得
- 減少記憶體配置
- 執行緒安全操作

### 3. 提升可測試性
- 更簡潔的 mock 實作
- 更容易建立測試場景
- 與測試框架更好的整合

### 4. 增強靈活性
- 內建 session 資料儲存
- 支援 session 複製
- 透過型別斷言擴充

## 重大變更

### 移除的方法
- `AddUserMessage()` → 使用 `AddMessage(agent.RoleUser, content)`
- `AddAssistantMessage()` → 使用 `AddMessage(agent.RoleAssistant, content)`
- `AddSystemMessage()` → 使用 `AddMessage(agent.RoleSystem, content)`
- `AddToolMessage()` → 使用 `AddMessage(agent.RoleTool, content)`
- `GetLastMessage()` → 使用 `Messages()[len(Messages())-1]`
- `GetMessagesByRole()` → 手動篩選 `Messages()`
- `MessageCount()` → 使用 `len(Messages())`
- `Clear()` → 使用型別斷言存取 `ClearMessages()`
- `Clone()` → 使用型別斷言存取 `Clone()`

### 行為變更
- 訊息新增現在需要明確指定角色
- 進階功能需要型別斷言
- 資料儲存現在透過 `SetData()`/`GetData()` 內建

## 遷移檢查清單

- [ ] 將所有 `AddUserMessage()` 呼叫更新為 `AddMessage(agent.RoleUser, content)`
- [ ] 將所有 `AddAssistantMessage()` 呼叫更新為 `AddMessage(agent.RoleAssistant, content)`
- [ ] 將所有 `AddSystemMessage()` 呼叫更新為 `AddMessage(agent.RoleSystem, content)`
- [ ] 將所有 `AddToolMessage()` 呼叫更新為 `AddMessage(agent.RoleTool, content)`
- [ ] 將 `GetLastMessage()` 替換為手動索引 `Messages()`
- [ ] 將 `GetMessagesByRole()` 替換為手動篩選 `Messages()`
- [ ] 將 `MessageCount()` 替換為 `len(Messages())`
- [ ] 更新 `Clear()` 呼叫使用型別斷言
- [ ] 更新 `Clone()` 呼叫使用型別斷言
- [ ] 將 session 資料儲存遷移到 `SetData()`/`GetData()`
- [ ] 更新測試案例使用新 API
- [ ] 執行全面測試確保相容性

## 需要協助？

如果您在遷移過程中遇到問題：

1. 查看 [Session 管理範例](../examples/session-management/) 了解完整使用模式
2. 檢閱 [API 文檔](../README-zh.md#session-管理) 了解詳細介面資訊
3. 如需協助請在 [GitHub](https://github.com/davidleitw/go-agent/issues) 開啟 issue

新的 Session API 為建構對話式 AI 應用程式提供了更簡潔、高效能的基礎，同時保持所有必要功能。