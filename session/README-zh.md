# Session 模組

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

Session 模組為 Go Agent Framework 提供對話狀態和歷史管理。它設計簡潔且完整，支援記憶體內儲存，並具有可擴展的介面以支援未來的實作。

## 功能特色

- **狀態管理**：支援任何資料類型的鍵值儲存
- **歷史追蹤**：統一的 Entry 結構支援多種對話類型
- **生命週期管理**：TTL 支援和自動過期清理
- **執行緒安全**：所有操作都是並發安全的
- **JSON 序列化**：完整的 JSON 支援，包含適當的標籤
- **可擴展性**：為 Redis、資料庫和其他持久性實作保留介面

## 快速開始

```go
import (
    "github.com/davidleitw/go-agent/session"
    "github.com/davidleitw/go-agent/session/memory"
)

// 創建儲存
store := memory.NewStore()
defer store.Close() // 清理資源

// 創建會話
sess := store.Create(context.Background(),
    session.WithTTL(24 * time.Hour),
    session.WithID("custom-id"), // 可選
)

// 狀態管理
sess.Set("current_task", "booking_flight")
sess.Set("user_preference", map[string]string{
    "language": "zh-TW",
    "currency": "TWD",
})

// 新增對話歷史
sess.AddEntry(session.NewMessageEntry("user", "預訂飛往東京的航班"))
sess.AddEntry(session.NewToolCallEntry("search_flights", map[string]any{
    "destination": "Tokyo",
}))

// 儲存（記憶體實作中是 no-op，但保持介面一致性）
err := store.Save(context.Background(), sess)

// 檢索會話
retrieved, err := store.Get(context.Background(), sess.ID())
history := retrieved.GetHistory(10) // 取得最近 10 個條目

// JSON 序列化支援
entryJSON, _ := json.Marshal(history[0])
fmt.Println(string(entryJSON))
```

## API 參考

### Session 介面

```go
type Session interface {
    // 基本資訊
    ID() string
    CreatedAt() time.Time
    UpdatedAt() time.Time
    
    // 狀態管理
    Get(key string) (any, bool)
    Set(key string, value any)
    Delete(key string)
    
    // 歷史管理
    AddEntry(entry Entry) error
    GetHistory(limit int) []Entry
}
```

### SessionStore 介面

```go
type SessionStore interface {
    Create(ctx context.Context, opts ...CreateOption) Session
    Get(ctx context.Context, id string) (Session, error)
    Save(ctx context.Context, session Session) error
    Delete(ctx context.Context, id string) error
    DeleteExpired(ctx context.Context) error
    Close() error
}
```

### Entry 類型

支援四種對話記錄類型：

1. **Message**：用戶/助手/系統訊息
2. **ToolCall**：工具調用記錄
3. **ToolResult**：工具執行結果
4. **Thinking**：內部推理過程（保留）

每種類型都有對應的創建函數和類型安全的提取函數：

```go
// 創建
entry := session.NewMessageEntry("user", "你好")
entry := session.NewToolCallEntry("search", params)
entry := session.NewToolResultEntry("search", result, err)

// 提取
if content, ok := session.GetMessageContent(entry); ok {
    fmt.Printf("%s: %s\n", content.Role, content.Text)
}
```

## 設計決策

### 為什麼保留 Save() 方法？

雖然 `Save()` 在記憶體實作中是 no-op，但我們保留它是為了：
- 支援未來的持久性實作（Redis、資料庫）
- 允許批次更新優化
- 維持介面一致性

### 為什麼移除 Touch() 和 ExpiresAt()？

- `Touch()`：UpdatedAt 在每次修改時自動更新
- `ExpiresAt()`：內部實作細節，無需暴露給用戶

### 背景清理機制

- 每 5 分鐘自動清理過期會話
- 提供 `Close()` 方法用於優雅關閉
- 可透過 `DeleteExpired()` 手動清理

## 擴展指南

### 實作自訂 SessionStore

```go
type MyStore struct {
    // 你的實作
}

func (s *MyStore) Create(ctx context.Context, opts ...session.CreateOption) session.Session {
    // 實作創建邏輯
}

func (s *MyStore) Get(ctx context.Context, id string) (session.Session, error) {
    // 實作檢索邏輯
}

func (s *MyStore) Save(ctx context.Context, sess session.Session) error {
    // 實作儲存邏輯
}

// ... 其他方法
```

### 計劃中的擴展

1. **Redis Store**
   - 分散式會話支援
   - 自動過期（使用 Redis TTL）
   - 批次操作優化

2. **Database Store**
   - SQL/NoSQL 支援
   - 交易保證
   - 查詢和分析能力

3. **混合儲存**
   - 記憶體作為 L1 快取
   - Redis/DB 作為持久層
   - 自動同步

## 與 Agent 系統整合

*待實作：此部分將在 Agent 核心功能完成後填入*

```go
// 預期的整合方式
agent := agent.New(
    agent.WithSessionStore(store),
    // 其他配置...
)

// Agent 自動在內部管理會話
response := agent.Process(ctx, request)
```

## 效能考量

- 記憶體實作適用於單機應用程式
- 建議限制歷史大小（例如最多 1000 個條目）
- 定期清理過期會話以防止記憶體洩漏
- 大規模應用程式應使用 Redis/DB 實作

## JSON 序列化

所有會話資料結構都包含適當的 JSON 標籤用於序列化：

```go
// 序列化完整會話歷史
history := session.GetHistory(0)
jsonData, err := json.Marshal(history)

// 序列化個別條目
entry := session.NewMessageEntry("user", "你好")
entryJSON, err := json.Marshal(entry)

// 反序列化條目
var deserializedEntry session.Entry
err = json.Unmarshal(entryJSON, &deserializedEntry)
```

### JSON 結構範例

**訊息條目：**
```json
{
  "id": "uuid-string",
  "type": "message", 
  "timestamp": "2024-01-01T12:00:00Z",
  "content": {
    "role": "user",
    "text": "你好世界"
  },
  "metadata": {}
}
```

**工具呼叫條目：**
```json
{
  "id": "uuid-string",
  "type": "tool_call",
  "timestamp": "2024-01-01T12:00:00Z", 
  "content": {
    "tool": "search",
    "parameters": {"query": "航班"}
  },
  "metadata": {}
}
```

**工具結果條目：**
```json
{
  "id": "uuid-string",
  "type": "tool_result",
  "timestamp": "2024-01-01T12:00:00Z",
  "content": {
    "tool": "search", 
    "success": true,
    "result": ["結果1", "結果2"],
    "error": ""
  },
  "metadata": {}
}
```

## 最佳實踐

1. **狀態管理**
   - 使用結構化的鍵命名（例如 `task.current`、`user.preference`）
   - 避免儲存過大的物件
   - 儲存前加密敏感資訊

2. **歷史記錄**
   - 適當使用 limit 參數避免載入過多歷史
   - 定期封存舊記錄
   - 使用 Metadata 新增額外資訊

3. **生命週期**
   - 根據使用情境設定合理的 TTL
   - 記得呼叫 `Close()` 清理資源
   - 監控會話數量防止洩漏

## 錯誤處理

```go
sess, err := store.Get(context.Background(), sessionID)
if errors.Is(err, session.ErrSessionNotFound) {
    // 創建新會話
    sess = store.Create(context.Background())
}
```

## 範例專案

請查看 `/examples/session/` 目錄以獲得完整範例。

## 貢獻

歡迎提交 Issues 和 Pull Requests！請確保：
- 新增適當的測試
- 更新相關文檔
- 遵循 Go 編碼慣例

## 授權

MIT 授權 - 請參閱專案根目錄中的 LICENSE 檔案。