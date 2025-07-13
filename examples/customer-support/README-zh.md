# 客戶支持機器人範例

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

這個範例展示了一個真實世界的客戶支持機器人，使用智能基於 schema 的信息收集來專業高效地處理不同類型的支持請求。

## 概述

客戶支持需要根據請求類型收集不同的信息。這個範例展示如何：

- **分類** 支持請求為不同類別
- **應用** 每種支持類型的專業 schema
- **收集** 跨多個對話輪次自然地收集信息
- **處理** 部分信息和上下文切換
- **維持** 專業支持代理行為

## 快速開始

```bash
# 設置你的 OpenAI API key
export OPENAI_API_KEY="your-api-key-here"

# 運行範例
go run examples/customer-support/main.go
```

## 代碼詳解

### 1. 支持類別定義

```go
// 定義不同類型的支持請求
const (
    GeneralSupport    = "general"
    BillingSupport    = "billing"
    TechnicalSupport  = "technical"
)
```

**關鍵要點：**
- 不同支持類型需要不同信息
- 每個類別都有自己的專業 schema
- 專業分類改善用戶體驗

### 2. 每種支持類型的 Schema 定義

#### 一般支持 Schema
```go
func getGeneralSupportSchema() []*schema.Field {
    return []*schema.Field{
        schema.Define("email", "請提供您的電子郵件地址以便後續聯絡"),
        schema.Define("issue_category", "您遇到的是什麼類型的問題？"),
        schema.Define("description", "請詳細描述問題"),
        schema.Define("order_id", "如果與購買相關，請提供訂單號").Optional(),
        schema.Define("urgency", "這個問題有多緊急？").Optional(),
        schema.Define("previous_contact", "您之前是否聯絡過我們關於此問題？").Optional(),
    }
}
```

#### 帳單支持 Schema
```go
func getBillingSupportSchema() []*schema.Field {
    return []*schema.Field{
        schema.Define("email", "請提供與您帳戶關聯的電子郵件地址"),
        schema.Define("account_number", "您的帳戶 ID 或號碼是什麼？"),
        schema.Define("billing_question", "請描述您的帳單問題或疑慮"),
        schema.Define("amount_disputed", "如果對收費有爭議，金額是多少？").Optional(),
        schema.Define("payment_method", "使用了什麼付款方式？").Optional(),
        schema.Define("billing_period", "這涉及哪個帳單週期？").Optional(),
    }
}
```

#### 技術支持 Schema
```go
func getTechnicalSupportSchema() []*schema.Field {
    return []*schema.Field{
        schema.Define("email", "請提供您的電子郵件地址以便技術後續聯絡"),
        schema.Define("error_message", "您看到什麼錯誤消息？"),
        schema.Define("steps_taken", "您已經嘗試了哪些故障排除步驟？"),
        schema.Define("browser", "您使用什麼瀏覽器？").Optional(),
        schema.Define("device_type", "您使用什麼設備？（桌面/手機/平板）").Optional(),
        schema.Define("operating_system", "什麼操作系統？").Optional(),
    }
}
```

**關鍵要點：**
- 每個 schema 都針對收集相關信息進行定制
- 必需字段專注於重要支持數據
- 可選字段收集有用的上下文
- 字段提示專業且具體

### 3. 支持機器人創建

```go
supportBot, err := agent.New("customer-support-bot").
    WithOpenAI(apiKey).
    WithModel("gpt-4o-mini").
    WithInstructions(`您是一名專業的客戶支持助手。
    您的目標是通過收集必要信息來高效地幫助客戶解決問題。
    要有同理心、清晰且有幫助。`).
    Build()
```

**關鍵要點：**
- 專業指令設定正確的語調
- 強調效率和同理心
- 模型選擇平衡成本和能力

### 4. 支持類型檢測

```go
func detectSupportType(userInput string) string {
    input := strings.ToLower(userInput)
    
    // 檢查帳單關鍵詞
    billingKeywords := []string{"帳單", "付款", "收費", "發票", "退款", "訂閱"}
    for _, keyword := range billingKeywords {
        if strings.Contains(input, keyword) {
            return BillingSupport
        }
    }
    
    // 檢查技術關鍵詞
    technicalKeywords := []string{"錯誤", "bug", "登入", "密碼", "技術", "壞了", "無法運作"}
    for _, keyword := range technicalKeywords {
        if strings.Contains(input, keyword) {
            return TechnicalSupport
        }
    }
    
    return GeneralSupport
}
```

**關鍵要點：**
- 基於關鍵詞的簡單分類
- 對不清楚的情況回退到一般支持
- 可以用 ML 分類增強

### 5. 動態 Schema 應用

```go
func handleSupportRequest(ctx context.Context, bot agent.Agent, userInput string) {
    // 檢測需要的支持類型
    supportType := detectSupportType(userInput)
    
    // 獲取適當的 schema
    var supportSchema []*schema.Field
    switch supportType {
    case BillingSupport:
        supportSchema = getBillingSupportSchema()
    case TechnicalSupport:
        supportSchema = getTechnicalSupportSchema()
    default:
        supportSchema = getGeneralSupportSchema()
    }
    
    // 將 schema 應用到對話
    response, err := bot.Chat(ctx, userInput,
        agent.WithSchema(supportSchema...),
    )
}
```

**關鍵要點：**
- 基於檢測意圖的動態 schema 選擇
- 不同支持類型間的無縫切換
- 始終維持對話上下文

### 6. 多輪對話處理

```go
// 對話流程範例
func runSupportConversation(ctx context.Context, bot agent.Agent) {
    sessionID := "support-conversation"
    
    // 第一次互動
    response1, _ := bot.Chat(ctx, "我需要某些幫助",
        agent.WithSession(sessionID),
        agent.WithSchema(getGeneralSupportSchema()...),
    )
    
    // 客戶提供更多信息
    response2, _ := bot.Chat(ctx, "我的郵箱是 john.doe@example.com，這是技術問題",
        agent.WithSession(sessionID),
        agent.WithSchema(getTechnicalSupportSchema()...),  // 切換到技術 schema
    )
    
    // 繼續直到收集所有信息
    response3, _ := bot.Chat(ctx, "我無法訪問我的儀表板。顯示'連接超時'錯誤",
        agent.WithSession(sessionID),
        agent.WithSchema(getTechnicalSupportSchema()...),
    )
}
```

**關鍵要點：**
- Session 維持對話上下文
- Schema 可以根據新信息變更
- 信息跨輪次累積

## 範例場景

### 場景 1：一般支持請求

```go
// 用戶："我的帳戶有問題"
// 檢測：一般支持（沒有特定關鍵詞）
// Schema：應用一般支持 schema
// 結果：收集郵箱、問題類別、描述

userInput := "我的帳戶有問題"
supportType := detectSupportType(userInput)  // 返回 "general"
schema := getGeneralSupportSchema()

response, err := supportBot.Chat(ctx, userInput,
    agent.WithSchema(schema...),
)
// 機器人詢問郵箱和帳戶問題的更多詳情
```

### 場景 2：帳單查詢

```go
// 用戶："我對我的上一張發票有問題"
// 檢測：帳單支持（包含"發票"）
// Schema：應用帳單專用 schema
// 結果：收集郵箱、帳戶號碼、帳單問題

userInput := "我對我的上一張發票有問題"
supportType := detectSupportType(userInput)  // 返回 "billing"
schema := getBillingSupportSchema()

response, err := supportBot.Chat(ctx, userInput,
    agent.WithSchema(schema...),
)
// 機器人詢問郵箱、帳戶號碼和具體帳單問題
```

### 場景 3：技術支持

```go
// 用戶："嘗試登入時遇到錯誤"
// 檢測：技術支持（包含"錯誤"和"登入"）
// Schema：應用技術支持 schema
// 結果：收集郵箱、錯誤消息、故障排除步驟

userInput := "嘗試登入時遇到錯誤"
supportType := detectSupportType(userInput)  // 返回 "technical"
schema := getTechnicalSupportSchema()

response, err := supportBot.Chat(ctx, userInput,
    agent.WithSchema(schema...),
)
// 機器人詢問郵箱、具體錯誤消息和已嘗試的步驟
```

### 場景 4：客戶提供部分信息

```go
// 用戶："嗨，我的郵箱是 customer@example.com，我對訂單 #12345 有帳單問題"
// 檢測：帳單支持（包含"帳單"）
// Schema：帳單 schema，但郵箱和訂單信息已提取
// 結果：詢問剩餘必需信息

userInput := "嗨，我的郵箱是 customer@example.com，我對訂單 #12345 有帳單問題"
response, err := supportBot.Chat(ctx, userInput,
    agent.WithSchema(getBillingSupportSchema()...),
)
// 機器人確認郵箱和訂單，詢問帳戶號碼和帳單問題詳情
```

## 高級功能

### 1. 支持優先級檢測

```go
func detectPriority(userInput string) string {
    input := strings.ToLower(userInput)
    
    urgentKeywords := []string{"緊急", "關鍵", "緊急情況", "立即", "馬上"}
    for _, keyword := range urgentKeywords {
        if strings.Contains(input, keyword) {
            return "high"
        }
    }
    
    return "normal"
}

// 將優先級應用到 schema
priority := detectPriority(userInput)
if priority == "high" {
    // 為高優先級請求添加緊急性字段作為必需
    schema = append(schema, schema.Define("urgency_reason", "為什麼這很緊急？"))
}
```

### 2. 升級邏輯

```go
func shouldEscalate(session agent.Session, supportType string) bool {
    messages := session.Messages()
    
    // 如果對話變得很長則升級
    if len(messages) > 10 {
        return true
    }
    
    // 多次失敗嘗試的技術問題升級
    if supportType == TechnicalSupport {
        for _, msg := range messages {
            if strings.Contains(strings.ToLower(msg.Content), "仍然無法運作") {
                return true
            }
        }
    }
    
    return false
}
```

### 3. 信息驗證

```go
func validateSupportInfo(extractedData map[string]interface{}) []string {
    var errors []string
    
    // 驗證郵箱格式
    if email, ok := extractedData["email"].(string); ok {
        if !isValidEmail(email) {
            errors = append(errors, "無效的郵箱格式")
        }
    }
    
    // 驗證帳戶號碼格式
    if accountNum, ok := extractedData["account_number"].(string); ok {
        if len(accountNum) < 6 {
            errors = append(errors, "帳戶號碼太短")
        }
    }
    
    return errors
}
```

### 4. 支持工單創建

```go
type SupportTicket struct {
    ID           string                 `json:"id"`
    Type         string                 `json:"type"`
    Priority     string                 `json:"priority"`
    CustomerInfo map[string]interface{} `json:"customer_info"`
    CreatedAt    time.Time              `json:"created_at"`
    Status       string                 `json:"status"`
}

func createSupportTicket(session agent.Session, supportType string) (*SupportTicket, error) {
    // 從 session 提取信息
    extractedInfo := extractAllInformation(session)
    
    ticket := &SupportTicket{
        ID:           generateTicketID(),
        Type:         supportType,
        Priority:     determinePriority(extractedInfo),
        CustomerInfo: extractedInfo,
        CreatedAt:    time.Now(),
        Status:       "open",
    }
    
    return ticket, nil
}
```

## 測試

運行範例測試：

```bash
go test ./examples/customer-support/
```

測試驗證：
- 支持類型檢測準確性
- 不同支持類型的 schema 選擇
- 跨對話輪次的信息收集
- 專業回應生成

### 測試案例

```go
func TestSupportTypeDetection(t *testing.T) {
    testCases := []struct {
        input    string
        expected string
    }{
        {"我有帳單問題", "billing"},
        {"遇到登入錯誤", "technical"},
        {"需要帳戶幫助", "general"},
        {"退款我的付款", "billing"},
        {"應用無法運作", "technical"},
    }
    
    for _, tc := range testCases {
        result := detectSupportType(tc.input)
        assert.Equal(t, tc.expected, result)
    }
}
```

## 集成範例

### 與 CRM 系統

```go
func saveToCRM(ticket *SupportTicket) error {
    crmData := map[string]interface{}{
        "customer_email": ticket.CustomerInfo["email"],
        "issue_type":     ticket.Type,
        "description":    ticket.CustomerInfo["description"],
        "priority":       ticket.Priority,
        "created_at":     ticket.CreatedAt,
    }
    
    return crmClient.CreateTicket(crmData)
}
```

### 與服務台軟件

```go
func createHelpDeskTicket(ticket *SupportTicket) error {
    helpdeskTicket := helpdesk.Ticket{
        Subject:     generateSubject(ticket),
        Description: formatDescription(ticket),
        Priority:    mapPriority(ticket.Priority),
        Customer:    ticket.CustomerInfo["email"].(string),
        Category:    ticket.Type,
    }
    
    return helpdeskClient.Create(helpdeskTicket)
}
```

### 與知識庫

```go
func suggestKnowledgeBaseArticles(supportType string, description string) []KBArticle {
    switch supportType {
    case TechnicalSupport:
        return searchTechnicalArticles(description)
    case BillingSupport:
        return searchBillingArticles(description)
    default:
        return searchGeneralArticles(description)
    }
}
```

## 最佳實踐

### 1. 專業溝通

```go
// 好：專業且有同理心
"很抱歉聽到您遇到這個問題。為了更好地幫助您，請您提供..."

// 避免：太隨意或機械化
"給我您的郵箱" 或 "請輸入必需的數據字段"
```

### 2. 信息優先級

```go
// 優先考慮重要信息
requiredFields := []*schema.Field{
    schema.Define("email", "後續聯絡的郵箱"),
    schema.Define("issue_description", "問題描述"),
}

// 稍後收集額外信息
optionalFields := []*schema.Field{
    schema.Define("phone", "緊急聯絡電話").Optional(),
    schema.Define("preferred_contact_time", "最佳聯絡時間").Optional(),
}
```

### 3. 上下文保持

```go
// 始終使用 session 維持上下文
response, err := supportBot.Chat(ctx, userInput,
    agent.WithSession(sessionID),
    agent.WithSchema(schema...),
)

// 不要失去先前的對話上下文
// 每輪都基於先前信息構建
```

### 4. 錯誤恢復

```go
if err != nil {
    // 在支持上下文中優雅處理錯誤
    fallbackResponse := "很抱歉，我在處理您的請求時遇到問題。" +
                        "請重試或直接聯絡我們的支持團隊。"
    return &agent.Response{Message: fallbackResponse}
}
```

## 性能考慮

### 1. Schema 緩存

```go
var (
    generalSchema   []*schema.Field
    billingSchema   []*schema.Field
    technicalSchema []*schema.Field
)

func init() {
    // 預創建 schema 以避免重複分配
    generalSchema = getGeneralSupportSchema()
    billingSchema = getBillingSupportSchema()
    technicalSchema = getTechnicalSupportSchema()
}
```

### 2. 對話限制

```go
const (
    MaxConversationTurns = 20
    MaxResponseTime      = 30 * time.Second
)

func enforceConversationLimits(session agent.Session) error {
    if len(session.Messages()) > MaxConversationTurns {
        return errors.New("對話太長，升級到人工代理")
    }
    return nil
}
```

## 相關範例

- **[簡單 Schema](../simple-schema/)**：基於 schema 收集的基礎概念
- **[動態 Schema](../dynamic-schema/)**：高級意圖分類和工作流管理
- **[基本聊天](../basic-chat/)**：核心對話概念
- **[多工具 Agent](../multi-tool-agent/)**：為支持機器人添加外部能力

## 下一步

1. **嘗試不同支持類型**：用各種支持場景測試
2. **自定義 Schema**：根據您的具體業務需求修改字段
3. **添加集成**：連接到您的 CRM 或服務台系統
4. **增強分類**：用 ML 模型改善支持類型檢測
5. **測試邊緣情況**：處理異常請求和錯誤場景

## 故障排除

**問題**：檢測到錯誤的支持類型
**解決方案**：改善關鍵詞列表或實現基於 ML 的分類

**問題**：信息未被正確提取
**解決方案**：檢查字段名稱和提示的清晰度

**問題**：對話變得太長
**解決方案**：實現升級邏輯和對話限制

**問題**：Schema 切換不起作用
**解決方案**：確保 session 連續性和正確的 schema 應用

要獲得更多幫助，請參閱[主要範例文檔](../README.md)或[schema 收集指南](../../docs/schema-collection.md)。