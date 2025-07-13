# Go-Agent 範例集

此目錄包含展示 go-agent 框架功能的完整範例。每個範例都展示了構建不同複雜度 AI 代理的各個方面。

## 🚀 快速開始

### 前置需求

1. **Go 1.21 或更高版本** 已安裝在您的系統上
2. **OpenAI API 金鑰** - 從 [OpenAI Platform](https://platform.openai.com/) 取得

### 設定

1. **複製儲存庫**:
   ```bash
   git clone <repository-url>
   cd go-agent
   ```

2. **安裝依賴**:
   ```bash
   go mod download
   ```

3. **設定環境變數**:
   ```bash
   export OPENAI_API_KEY="your-openai-api-key-here"
   ```

### 執行範例

每個範例都可以直接從專案根目錄執行：

```bash
# 執行基本聊天範例
go run examples/basic-chat/main.go

# 執行計算器工具範例
go run examples/calculator-tool/main.go

# 執行任務完成範例
go run examples/task-completion/main.go

# 執行多工具代理範例
go run examples/multi-tool-agent/main.go

# 執行條件測試範例
go run examples/condition-testing/main.go

# 執行簡單 Schema 範例
go run examples/simple-schema/main.go

# 執行客戶支持範例
go run examples/customer-support/main.go

# 執行動態 Schema 範例
go run examples/dynamic-schema/main.go
```

## 📚 範例概覽

### 1. 基本聊天 (`basic-chat/`)
**目的**: 展示最簡單的 AI 代理實現

**展示內容**:
- 基本對話流程
- 使用 `BasicAgent` 處理簡單場景
- 會話管理
- 訊息處理

**核心 API**:
- `agent.NewBasicAgent()`
- `agent.NewSession()`
- `agent.Chat()`

**使用案例**: 理解框架基礎的完美起點。

**詳細文檔**: [README](basic-chat/README-zh.md) | [English](basic-chat/README.md)

---

### 2. 計算器工具 (`calculator-tool/`)
**目的**: 展示如何將外部工具與 AI 代理整合

**展示內容**:
- 工具實現和註冊
- 工具執行和錯誤處理
- 數學運算整合
- 結構化工具回應

**核心 API**:
- `agent.Tool` 介面
- 工具架構定義
- 工具執行上下文
- 工具中的錯誤處理

**使用案例**: 學習如何用自訂工具擴展代理功能。

**詳細文檔**: [README](calculator-tool/README-zh.md) | [English](calculator-tool/README.md)

---

### 3. 任務完成 (`task-completion/`)
**目的**: 展示結構化輸出和資料收集工作流程

**展示內容**:
- 結構化輸出類型
- 資料驗證
- 任務導向對話
- 進度追蹤

**核心 API**:
- `agent.OutputType` 介面
- `agent.NewStructuredOutputType()`
- JSON 架構生成
- 輸出驗證

**使用案例**: 建立系統性收集和結構化用戶資料的代理。

**詳細文檔**: [README](task-completion/README-zh.md) | [English](task-completion/README.md)

---

### 4. 多工具代理 (`multi-tool-agent/`)
**目的**: 展示具有多個工具的自訂代理實現的高級範例

**展示內容**:
- 自訂代理實現
- 多工具協調
- 工具使用統計
- 動態指令增強
- 高級狀態管理

**核心 API**:
- `agent.Agent` 介面實現
- 自訂聊天邏輯
- 工具編排
- 狀態追蹤

**使用案例**: 建立智能協調多種功能的複雜代理。

**詳細文檔**: [README](multi-tool-agent/README-zh.md) | [English](multi-tool-agent/README.md)

---

### 5. 條件測試 (`condition-testing/`)
**目的**: 對話中的高級流程控制和條件邏輯

**展示內容**:
- 流程規則和條件
- 動態對話流程
- 用戶入職流程
- 條件工具執行
- 高級狀態管理

**核心 API**:
- `agent.FlowRule` 介面
- `agent.Condition` 介面
- 動態流程控制
- 條件動作

**使用案例**: 建立具有複雜、適應性對話流程的代理。

**詳細文檔**: [README](condition-testing/README-zh.md) | [English](condition-testing/README.md)

---

### 6. 簡單 Schema (`simple-schema/`)
**目的**: 基於 schema 的基本信息收集

**展示內容**:
- 使用 `schema.Define()` 定義字段
- 必需與可選字段
- 自動信息提取
- 自然對話流程
- 基於會話的信息持久化

**核心 API**:
- `schema.Define()` - 字段定義
- `schema.Field.Optional()` - 可選字段標記
- `agent.WithSchema()` - Schema 應用
- Schema 收集元數據

**使用案例**: 學習智能信息收集的基本概念。

**詳細文檔**: [README](simple-schema/README-zh.md) | [English](simple-schema/README.md)

---

### 7. 客戶支持 (`customer-support/`)
**目的**: 具有智能信息收集的真實客戶支持機器人

**展示內容**:
- 專業支持工作流程
- 不同支持類型的專業 schema
- 多輪對話處理
- 上下文信息提取
- 支持工單信息收集

**核心 API**:
- 動態 schema 選擇
- 支持專用字段定義
- 多 schema 工作流程
- 專業提示設計

**使用案例**: 建立生產就緒的客戶支持系統。

**詳細文檔**: [README](customer-support/README-zh.md) | [English](customer-support/README.md)

---

### 8. 動態 Schema (`dynamic-schema/`)
**目的**: 高級 schema 選擇和多步驟工作流程

**展示內容**:
- 意圖分類系統
- 基於上下文的動態 schema 選擇
- 多步驟信息收集工作流程
- 複雜對話管理
- 實時 schema 適應

**核心 API**:
- 基於意圖的 schema 選擇
- 多步驟工作流程編排
- 高級對話分析
- 複雜業務邏輯集成

**使用案例**: 建立具有適應性數據收集的複雜對話系統。

**詳細文檔**: [README](dynamic-schema/README-zh.md) | [English](dynamic-schema/README.md)

## 🏗️ 架構模式

### BasicAgent vs 自訂 Agent

**使用 BasicAgent 當**:
- 簡單、直接的對話
- 標準工具使用模式
- 最少的狀態管理需求
- 快速原型開發

**使用自訂 Agent 當**:
- 需要複雜狀態管理
- 需要高級工具協調
- 自訂對話流程邏輯
- 複雜的錯誤處理

### 工具整合模式

**簡單工具**: 單一目的、無狀態操作
```go
type CalculatorTool struct{}

func (t *CalculatorTool) Execute(ctx context.Context, args map[string]any) (any, error) {
    // Implementation
}
```

**複雜工具**: 多操作、有狀態工具
```go
type WeatherTool struct {
    apiKey string
    cache  map[string]WeatherData
}
```

### 流程控制模式

**基於條件**: 對對話狀態做出反應
```go
type MissingFieldsCondition struct {
    requiredFields []string
}

func (c *MissingFieldsCondition) Evaluate(ctx context.Context, session agent.Session, data map[string]any) (bool, error) {
    // Check if fields are missing
}
```

**基於規則**: 當條件滿足時執行動作
```go
type FlowRule struct {
    Name      string
    Condition agent.Condition
    Action    agent.Action
}
```

## 🔧 開發指南

### 新增範例

1. 在 `cmd/examples/` 下建立新目錄
2. 實現帶有完整日誌記錄的 `main.go`
3. 新增包含程式碼說明的詳細 README
4. 如需要，包含 `.env.example`
5. 用各種場景徹底測試

### 程式碼風格

- 使用描述性變數名稱
- 新增完整的除錯日誌記錄
- 優雅地處理錯誤
- 在相關地方包含效能指標
- 用註解記錄複雜邏輯

### 測試

- 用各種輸入場景測試
- 驗證錯誤處理
- 檢查資源清理
- 驗證輸出格式
- 測試邊界情況

## 🐛 疑難排解

### 常見問題

1. **OpenAI API 金鑰問題**:
   - 確保金鑰設定正確
   - 檢查速率限制
   - 驗證金鑰權限

2. **工具執行錯誤**:
   - 檢查工具參數驗證
   - 驗證工具架構符合使用方式
   - 檢查超時設定

3. **流程規則問題**:
   - 除錯條件評估
   - 檢查動作實現
   - 驗證規則順序

### 除錯技巧

- 啟用詳細日誌記錄
- 使用小型測試案例
- 檢查 API 回應
- 驗證輸入/輸出格式
- 監控資源使用

## 📖 延伸閱讀

- [Go-Agent 文件](../../README.md)
- [API 參考](../../docs/api.md)
- [架構指南](../../docs/architecture.md)
- [最佳實踐](../../docs/best-practices.md)

## 🤝 貢獻

我們歡迎貢獻！請：

1. 遵循現有的程式碼風格
2. 新增完整的測試
3. 更新文件
4. 包含使用範例
5. 用多種場景測試

## 📄 授權

此專案採用 MIT 授權 - 詳見 [LICENSE](../../LICENSE) 檔案。 