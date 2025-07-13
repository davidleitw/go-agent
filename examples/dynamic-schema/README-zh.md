# 動態 Schema 選擇範例

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

這個範例展示了 go-agent 中最高級的基於 schema 的信息收集功能，包括智能意圖分類、動態 schema 選擇和多步驟工作流編排。

## 概述

現實世界的應用程序通常需要處理多種不同的對話流程，每種都需要不同的信息。這個範例展示如何：

- **分類** 從自然語言輸入中分類用戶意圖
- **選擇** 基於上下文動態選擇適當的 schema
- **編排** 多步驟信息收集工作流
- **適應** 實時調整對話策略
- **集成** 複雜業務邏輯與 schema 收集

## 快速開始

```bash
# 設置你的 OpenAI API key
export OPENAI_API_KEY="your-api-key-here"

# 運行範例
go run examples/dynamic-schema/main.go
```

## 代碼詳解

### 1. 意圖分類系統

```go
type IntentClassifier struct {
    intentKeywords map[string][]string
}

func NewIntentClassifier() *IntentClassifier {
    return &IntentClassifier{
        intentKeywords: map[string][]string{
            "technical_support": {"錯誤", "bug", "壞了", "無法工作", "技術", "登入", "密碼", "問題"},
            "billing_inquiry":   {"帳單", "付款", "收費", "發票", "退款", "訂閱", "費用", "價格"},
            "account_management": {"帳戶", "個人資料", "設定", "更改", "更新", "刪除", "隱私"},
            "product_inquiry":   {"功能", "如何", "幫助", "教程", "指南", "文檔", "使用"},
            "sales_inquiry":     {"購買", "採購", "定價", "計劃", "企業版", "演示", "報價", "聯絡銷售"},
            "general_inquiry":   {"你好", "嗨", "支持", "幫助", "問題", "信息"},
        },
    }
}
```

**關鍵要點：**
- 基於關鍵詞分析的意圖分類
- 可擴展到基於 ML 的分類
- 對未知意圖回退到一般查詢
- 處理多語言和同義詞

### 2. 意圖分類邏輯

```go
func (ic *IntentClassifier) ClassifyIntent(userInput string) string {
    input := strings.ToLower(userInput)
    intentScores := make(map[string]int)
    
    // 基於關鍵詞匹配為每個意圖評分
    for intent, keywords := range ic.intentKeywords {
        score := 0
        for _, keyword := range keywords {
            if strings.Contains(input, keyword) {
                score++
            }
        }
        if score > 0 {
            intentScores[intent] = score
        }
    }
    
    // 返回得分最高的意圖
    if len(intentScores) == 0 {
        return "general_inquiry"
    }
    
    maxScore := 0
    bestIntent := "general_inquiry"
    for intent, score := range intentScores {
        if score > maxScore {
            maxScore = score
            bestIntent = intent
        }
    }
    
    return bestIntent
}
```

**關鍵要點：**
- 準確分類的評分系統
- 多個關鍵詞匹配得更高分
- 健壯的回退處理
- 對大多數用例簡單而有效

### 3. 動態 Schema 選擇

```go
func getSchemaForIntent(intent string) []*schema.Field {
    switch intent {
    case "technical_support":
        return []*schema.Field{
            schema.Define("email", "請提供您的電子郵件地址以便技術後續聯絡"),
            schema.Define("error_description", "請描述您遇到的錯誤或問題"),
            schema.Define("steps_taken", "您已經嘗試了哪些故障排除步驟？"),
            schema.Define("environment", "您使用什麼瀏覽器/設備？").Optional(),
            schema.Define("urgency", "這個問題對您的工作有多關鍵？（低/中/高）").Optional(),
        }
    
    case "billing_inquiry":
        return []*schema.Field{
            schema.Define("email", "請提供與您帳戶關聯的電子郵件地址"),
            schema.Define("account_id", "您的帳戶 ID 或號碼是什麼？"),
            schema.Define("billing_question", "請描述您的帳單問題或疑慮"),
            schema.Define("amount", "如果涉及特定金額，請指定").Optional(),
            schema.Define("transaction_date", "交易發生時間？（如果適用）").Optional(),
        }
    
    case "sales_inquiry":
        return []*schema.Field{
            schema.Define("email", "請提供您的商務電子郵件"),
            schema.Define("company", "您代表哪家公司？"),
            schema.Define("team_size", "您的團隊有多少人？"),
            schema.Define("use_case", "您計劃如何使用我們的產品？"),
            schema.Define("timeline", "您希望什麼時候開始使用？").Optional(),
            schema.Define("budget", "您心中有預算範圍嗎？").Optional(),
        }
    
    // ... 更多意圖 schema
    
    default:
        return getGeneralInquirySchema()
    }
}
```

**關鍵要點：**
- 每個意圖都有專業的 schema
- 按意圖優化必需與可選字段
- 業務上下文的專業字段提示
- 對未知意圖回退到一般 schema

### 4. 多步驟工作流系統

```go
func getWorkflowForIntent(intent string) [][]*schema.Field {
    switch intent {
    case "technical_support":
        return [][]*schema.Field{
            { // 步驟 1：基本聯絡和問題摘要
                schema.Define("email", "您的電子郵件地址"),
                schema.Define("issue_summary", "問題的簡要描述"),
            },
            { // 步驟 2：技術詳情
                schema.Define("error_message", "您看到的確切錯誤消息"),
                schema.Define("steps_to_reproduce", "我們如何重現這個問題？"),
                schema.Define("browser_version", "什麼瀏覽器和版本？").Optional(),
            },
            { // 步驟 3：影響評估
                schema.Define("when_started", "這個問題首次出現是什麼時候？"),
                schema.Define("frequency", "這種情況多久發生一次？"),
                schema.Define("workaround", "您找到任何臨時解決方案嗎？").Optional(),
            },
        }
    
    case "sales_inquiry":
        return [][]*schema.Field{
            { // 步驟 1：基本資格認定
                schema.Define("email", "商務電子郵件地址"),
                schema.Define("company", "公司名稱"),
                schema.Define("role", "您在公司的職位"),
            },
            { // 步驟 2：需求收集
                schema.Define("team_size", "您團隊的規模"),
                schema.Define("use_case", "我們產品的主要使用場景"),
                schema.Define("current_solution", "您現在使用什麼解決方案？").Optional(),
            },
            { // 步驟 3：時間表和預算
                schema.Define("timeline", "您需要什麼時候實施？"),
                schema.Define("decision_process", "還有誰參與決策？"),
                schema.Define("budget_range", "您考慮的預算範圍").Optional(),
            },
        }
    
    default:
        // 簡單意圖的單步工作流
        return [][]*schema.Field{getSchemaForIntent(intent)}
    }
}
```

**關鍵要點：**
- 複雜意圖分解為可管理的步驟
- 漸進式信息收集
- 上下文跨工作流步驟構建
- 靈活的單步回退

### 5. 適應性對話策略

```go
func runAdaptiveConversation(ctx context.Context, bot agent.Agent, scenarios []ConversationScenario) {
    classifier := NewIntentClassifier()
    
    for i, scenario := range scenarios {
        fmt.Printf("📝 場景 %d：%s\n", i+1, scenario.Description)
        fmt.Printf("👤 用戶：%s\n", scenario.UserInput)
        
        // 分類意圖
        intent := classifier.ClassifyIntent(scenario.UserInput)
        fmt.Printf("🎯 檢測到的意圖：%s\n", intent)
        
        // 獲取適當的 schema
        schema := getSchemaForIntent(intent)
        fmt.Printf("📋 選擇的 Schema（%d 個字段）：\n", len(schema))
        for _, field := range schema {
            requiredText := "必需"
            if !field.Required() {
                requiredText = "可選"
            }
            fmt.Printf("   - %s（%s）：%s\n", field.Name(), requiredText, field.Prompt())
        }
        
        // 使用選擇的 schema 執行對話
        response, err := bot.Chat(ctx, scenario.UserInput,
            agent.WithSchema(schema...),
        )
        
        if err != nil {
            fmt.Printf("❌ 錯誤：%v\n", err)
            continue
        }
        
        fmt.Printf("🤖 助手：%s\n", response.Message)
        fmt.Printf("⏱️  回應時間：%.3f秒\n", time.Since(startTime).Seconds())
    }
}
```

**關鍵要點：**
- 實時意圖分類
- 基於檢測意圖的 schema 選擇
- 性能指標和調試信息
- 錯誤處理和優雅降級

### 6. 多步驟工作流執行

```go
func runMultiStepWorkflow(ctx context.Context, bot agent.Agent, intent string, userInput string) {
    workflow := getWorkflowForIntent(intent)
    sessionID := fmt.Sprintf("workflow-%s", intent)
    
    fmt.Printf("🔄 多步驟工作流範例\n")
    fmt.Printf("👤 用戶：%s\n", userInput)
    fmt.Printf("🎯 意圖：%s\n", intent)
    fmt.Printf("📊 工作流步驟：%d\n\n", len(workflow))
    
    for stepIndex, stepSchema := range workflow {
        fmt.Printf("📋 步驟 %d/%d - 收集中：\n", stepIndex+1, len(workflow))
        for _, field := range stepSchema {
            requiredText := "必需"
            if !field.Required() {
                requiredText = "可選"
            }
            fmt.Printf("   - %s（%s）\n", field.Name(), requiredText)
        }
        
        // 執行此步驟
        response, err := bot.Chat(ctx, userInput,
            agent.WithSession(sessionID),
            agent.WithSchema(stepSchema...),
        )
        
        if err != nil {
            fmt.Printf("❌ 步驟 %d 失敗：%v\n", stepIndex+1, err)
            break
        }
        
        fmt.Printf("👤 用戶：%s\n", userInput)
        fmt.Printf("🤖 助手：%s\n", response.Message)
        
        // 檢查此步驟是否完成
        if schemaCollection, ok := response.Metadata["schema_collection"].(bool); ok && schemaCollection {
            if missingFields, ok := response.Metadata["missing_fields"].([]string); ok {
                fmt.Printf("📊 仍需要：%v\n", missingFields)
            }
        } else {
            fmt.Printf("✅ 步驟 %d 完成！\n", stepIndex+1)
        }
        
        // 為下一步模擬用戶提供信息
        userInput = generateSimulatedUserResponse(stepIndex, intent)
        fmt.Printf("\n")
    }
}
```

**關鍵要點：**
- 帶進度跟踪的逐步執行
- 跨工作流步驟的 session 連續性
- 缺失字段識別和處理
- 用於演示的模擬用戶回應

## 範例場景

### 場景 1：技術支持分類

```go
// 用戶輸入："嘗試訪問系統時遇到登入錯誤"
// 預期：檢測到技術支持意圖
// Schema：技術支持字段（email、error_description、steps_taken 等）
// 工作流：3 步技術故障排除流程

userInput := "嘗試訪問系統時遇到登入錯誤"
intent := classifier.ClassifyIntent(userInput)
// 返回："technical_support"

schema := getSchemaForIntent(intent)
// 返回：技術支持 schema，5 個字段（3 個必需，2 個可選）

workflow := getWorkflowForIntent(intent)
// 返回：全面技術支持的 3 步工作流
```

### 場景 2：銷售查詢處理

```go
// 用戶輸入："我有興趣為我的公司購買您的企業版計劃"
// 預期：檢測到銷售查詢意圖
// Schema：銷售資格字段（email、company、team_size、use_case 等）
// 工作流：3 步銷售資格認定流程

userInput := "我有興趣為我的公司購買您的企業版計劃"
intent := classifier.ClassifyIntent(userInput)
// 返回："sales_inquiry"

schema := getSchemaForIntent(intent)
// 返回：銷售 schema，6 個字段（4 個必需，2 個可選）

workflow := getWorkflowForIntent(intent)
// 返回：銷售資格認定的 3 步工作流
```

### 場景 3：帳單問題處理

```go
// 用戶輸入："我對最新發票上的收費有疑問"
// 預期：檢測到帳單查詢意圖
// Schema：帳單專用字段（email、account_id、billing_question 等）
// 工作流：單步帳單信息收集

userInput := "我對最新發票上的收費有疑問"
intent := classifier.ClassifyIntent(userInput)
// 返回："billing_inquiry"

schema := getSchemaForIntent(intent)
// 返回：帳單 schema，5 個字段（3 個必需，2 個可選）
```

## 高級功能

### 1. 意圖信心度評分

```go
func (ic *IntentClassifier) ClassifyWithConfidence(userInput string) (string, float64) {
    input := strings.ToLower(userInput)
    intentScores := make(map[string]int)
    totalKeywords := 0
    
    for intent, keywords := range ic.intentKeywords {
        score := 0
        for _, keyword := range keywords {
            if strings.Contains(input, keyword) {
                score++
            }
        }
        intentScores[intent] = score
        totalKeywords += len(keywords)
    }
    
    maxScore := 0
    bestIntent := "general_inquiry"
    for intent, score := range intentScores {
        if score > maxScore {
            maxScore = score
            bestIntent = intent
        }
    }
    
    confidence := float64(maxScore) / float64(len(strings.Fields(input)))
    return bestIntent, confidence
}
```

### 2. 上下文感知的 Schema 適應

```go
func adaptSchemaToContext(baseSchema []*schema.Field, conversationHistory []agent.Message) []*schema.Field {
    adaptedSchema := make([]*schema.Field, 0, len(baseSchema))
    
    // 分析對話歷史的上下文
    hasUrgencyIndicators := false
    hasCompanyMention := false
    
    for _, msg := range conversationHistory {
        content := strings.ToLower(msg.Content)
        if strings.Contains(content, "緊急") || strings.Contains(content, "關鍵") {
            hasUrgencyIndicators = true
        }
        if strings.Contains(content, "公司") || strings.Contains(content, "企業") {
            hasCompanyMention = true
        }
    }
    
    // 基於上下文適應 schema
    for _, field := range baseSchema {
        adaptedField := field
        
        // 如果存在緊急指標，使緊急度成為必需
        if field.Name() == "urgency" && hasUrgencyIndicators {
            adaptedField = schema.Define(field.Name(), field.Prompt())
        }
        
        // 如果檢測到企業上下文，添加公司專用字段
        if field.Name() == "email" && hasCompanyMention {
            adaptedSchema = append(adaptedSchema, 
                schema.Define("company_name", "您代表哪家公司？"))
        }
        
        adaptedSchema = append(adaptedSchema, adaptedField)
    }
    
    return adaptedSchema
}
```

### 3. 工作流狀態管理

```go
type WorkflowState struct {
    Intent         string                 `json:"intent"`
    CurrentStep    int                    `json:"current_step"`
    TotalSteps     int                    `json:"total_steps"`
    CollectedData  map[string]interface{} `json:"collected_data"`
    CompletedSteps []int                  `json:"completed_steps"`
    StartTime      time.Time              `json:"start_time"`
}

func (ws *WorkflowState) IsComplete() bool {
    return ws.CurrentStep >= ws.TotalSteps
}

func (ws *WorkflowState) Progress() float64 {
    return float64(len(ws.CompletedSteps)) / float64(ws.TotalSteps)
}

func (ws *WorkflowState) NextStep() int {
    if ws.CurrentStep < ws.TotalSteps-1 {
        return ws.CurrentStep + 1
    }
    return ws.CurrentStep
}
```

### 4. 分析和報告

```go
type ConversationAnalytics struct {
    TotalMessages     int                    `json:"total_messages"`
    IntentChanges     int                    `json:"intent_changes"`
    WorkflowComplete  bool                   `json:"workflow_complete"`
    CollectionRate    float64                `json:"collection_rate"`
    AverageResponse   time.Duration          `json:"average_response_time"`
    FieldsCollected   map[string]interface{} `json:"fields_collected"`
}

func generateAnalytics(session agent.Session, workflow [][]*schema.Field) *ConversationAnalytics {
    messages := session.Messages()
    
    // 計算收集率
    totalFields := 0
    collectedFields := 0
    
    for _, step := range workflow {
        for _, field := range step {
            totalFields++
            if field.Required() {
                // 檢查字段是否被收集
                if isFieldCollected(field.Name(), messages) {
                    collectedFields++
                }
            }
        }
    }
    
    collectionRate := float64(collectedFields) / float64(totalFields)
    
    return &ConversationAnalytics{
        TotalMessages:    len(messages),
        WorkflowComplete: collectionRate >= 0.8, // 80% 收集率閾值
        CollectionRate:   collectionRate,
        FieldsCollected:  extractCollectedFields(messages),
    }
}
```

## 測試

運行綜合測試套件：

```bash
go test ./examples/dynamic-schema/
```

### 測試覆蓋

#### 意圖分類測試
```go
func TestIntentClassifier(t *testing.T) {
    classifier := NewIntentClassifier()
    
    testCases := []struct {
        input    string
        expected string
    }{
        {"遇到登入錯誤", "technical_support"},
        {"對發票有帳單問題", "billing_inquiry"},
        {"需要更改密碼", "account_management"},
        {"如何使用分析功能？", "product_inquiry"},
        {"想購買企業版計劃", "sales_inquiry"},
        {"你好，我有一些問題", "general_inquiry"},
    }
    
    for _, tc := range testCases {
        result := classifier.ClassifyIntent(tc.input)
        assert.Equal(t, tc.expected, result)
    }
}
```

#### Schema 選擇測試
```go
func TestGetSchemaForIntent(t *testing.T) {
    testCases := []struct {
        intent           string
        expectedRequired []string
        expectedOptional []string
    }{
        {
            intent:           "technical_support",
            expectedRequired: []string{"email", "error_description", "steps_taken"},
            expectedOptional: []string{"environment", "urgency"},
        },
        // ... 更多測試案例
    }
    
    for _, tc := range testCases {
        schema := getSchemaForIntent(tc.intent)
        validateSchemaFields(t, schema, tc.expectedRequired, tc.expectedOptional)
    }
}
```

#### 工作流測試
```go
func TestWorkflowExecution(t *testing.T) {
    mockModel := NewMockChatModel(
        `{"email": null, "error_description": null, "steps_taken": null}`,
        "我理解您遇到技術問題。請提供您的電子郵件並描述錯誤。",
    )
    
    bot, err := agent.New("test-bot").
        WithChatModel(mockModel).
        WithInstructions("您是測試助手。").
        Build()
    
    require.NoError(t, err)
    
    workflow := getWorkflowForIntent("technical_support")
    require.Equal(t, 3, len(workflow))
    
    // 測試每個工作流步驟
    for i, step := range workflow {
        response, err := bot.Chat(context.Background(), 
            fmt.Sprintf("步驟 %d 的測試輸入", i+1),
            agent.WithSchema(step...),
        )
        require.NoError(t, err)
        require.NotEmpty(t, response.Message)
    }
}
```

## 性能優化

### 1. Schema 緩存

```go
var (
    schemaCache     = make(map[string][]*schema.Field)
    workflowCache   = make(map[string][][]*schema.Field)
    cacheMutex      sync.RWMutex
)

func getCachedSchema(intent string) []*schema.Field {
    cacheMutex.RLock()
    defer cacheMutex.RUnlock()
    
    if schema, exists := schemaCache[intent]; exists {
        return schema
    }
    
    cacheMutex.RUnlock()
    cacheMutex.Lock()
    defer cacheMutex.Unlock()
    
    // 獲取寫鎖後雙重檢查
    if schema, exists := schemaCache[intent]; exists {
        return schema
    }
    
    schema := getSchemaForIntent(intent)
    schemaCache[intent] = schema
    return schema
}
```

### 2. 意圖分類優化

```go
type OptimizedIntentClassifier struct {
    keywordTrie  *Trie
    intentScorer *IntentScorer
}

func (oic *OptimizedIntentClassifier) FastClassify(input string) string {
    // 使用 trie 進行 O(1) 關鍵詞查找
    keywords := oic.keywordTrie.FindKeywords(input)
    
    // 使用優化的評分算法
    return oic.intentScorer.CalculateBestIntent(keywords)
}
```

### 3. 工作流狀態持久化

```go
type WorkflowStateManager struct {
    states map[string]*WorkflowState
    mutex  sync.RWMutex
}

func (wsm *WorkflowStateManager) SaveState(sessionID string, state *WorkflowState) {
    wsm.mutex.Lock()
    defer wsm.mutex.Unlock()
    wsm.states[sessionID] = state
}

func (wsm *WorkflowStateManager) LoadState(sessionID string) (*WorkflowState, bool) {
    wsm.mutex.RLock()
    defer wsm.mutex.RUnlock()
    state, exists := wsm.states[sessionID]
    return state, exists
}
```

## 集成範例

### 與機器學習分類

```go
type MLIntentClassifier struct {
    modelEndpoint string
    fallbackClassifier *IntentClassifier
}

func (ml *MLIntentClassifier) ClassifyIntent(input string) string {
    // 首先嘗試 ML 分類
    result, confidence, err := ml.callMLModel(input)
    if err != nil || confidence < 0.8 {
        // 回退到基於關鍵詞的分類
        return ml.fallbackClassifier.ClassifyIntent(input)
    }
    return result
}
```

### 與外部 CRM 系統

```go
func syncWithCRM(collectedData map[string]interface{}, intent string) error {
    switch intent {
    case "sales_inquiry":
        return createCRMLead(collectedData)
    case "technical_support":
        return createSupportTicket(collectedData)
    case "billing_inquiry":
        return createBillingInquiry(collectedData)
    default:
        return createGeneralInquiry(collectedData)
    }
}
```

### 與分析平台

```go
func trackConversationMetrics(analytics *ConversationAnalytics) {
    metrics := map[string]interface{}{
        "intent":              analytics.Intent,
        "workflow_completed":  analytics.WorkflowComplete,
        "collection_rate":     analytics.CollectionRate,
        "response_time":       analytics.AverageResponse.Milliseconds(),
        "total_messages":      analytics.TotalMessages,
    }
    
    analyticsClient.Track("conversation_completed", metrics)
}
```

## 最佳實踐

### 1. 意圖設計

**好的意圖類別：**
- 足夠具體以指導 schema 選擇
- 足夠通用以處理變化
- 不重疊以避免分類衝突
- 與您的業務領域相關

**範例結構：**
```go
intents := map[string][]string{
    "technical_support": {"錯誤", "bug", "壞了", "無法工作"},
    "account_help":     {"帳戶", "個人資料", "設定", "登入"},
    "billing_support":  {"帳單", "付款", "發票", "收費"},
}
```

### 2. Schema 設計

**漸進式揭露：**
- 從重要信息開始
- 在後續步驟中添加詳細信息
- 對額外數據使用可選字段

**上下文感知：**
- 基於對話歷史適應字段
- 考慮用戶角色和意圖
- 優化用戶體驗

### 3. 工作流設計

**步驟組織：**
- 邏輯信息進展
- 合理的步驟大小（最多 3-5 個字段）
- 清晰的完成標準
- 複雜情況的退出路線

### 4. 錯誤處理

```go
func handleWorkflowError(err error, step int, intent string) *agent.Response {
    switch {
    case errors.Is(err, ErrTooManyRetries):
        return &agent.Response{
            Message: "我收集這些信息時遇到困難。讓我為您聯繫人工代理。",
        }
    case errors.Is(err, ErrInvalidInput):
        return &agent.Response{
            Message: "我不太理解。您能換個說法嗎？",
        }
    default:
        return &agent.Response{
            Message: "出了點問題。讓我們重新開始或聯繫我們的支持團隊。",
        }
    }
}
```

## 相關範例

- **[簡單 Schema](../simple-schema/)**：基於 schema 收集的基礎概念
- **[客戶支持](../customer-support/)**：帶專業 schema 的現實世界應用
- **[基本聊天](../basic-chat/)**：核心對話概念
- **[多工具 Agent](../multi-tool-agent/)**：添加外部能力

## 下一步

1. **實驗意圖分類**：嘗試不同的關鍵詞集和評分算法
2. **設計自定義工作流**：為您的特定用例創建多步驟流程
3. **實施 ML 分類**：從基於關鍵詞升級到基於 ML 的意圖檢測
4. **添加分析**：跟踪工作流性能和用戶行為
5. **為生產擴展**：實施緩存、持久化和監控

## 故障排除

**問題**：意圖分類準確性差
**解決方案**：檢查並擴展關鍵詞列表，考慮基於 ML 的分類

**問題**：工作流卡住
**解決方案**：實施超時處理和升級路徑

**問題**：Schema 選擇不符合用戶需求
**解決方案**：添加上下文感知和對話歷史分析

**問題**：複雜工作流的性能問題
**解決方案**：實施緩存並優化 schema/工作流創建

要獲得全面指導，請參閱 [schema 收集文檔](../../docs/schema-collection.md) 和 [範例概述](../README.md)。