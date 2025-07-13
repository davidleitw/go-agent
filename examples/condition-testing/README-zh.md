# 條件測試範例

[![English](https://img.shields.io/badge/README-English-blue.svg)](README.md) [![繁體中文](https://img.shields.io/badge/README-繁體中文-red.svg)](README-zh.md)

這個範例展示了進階的條件驗證和流程規則實現，使用用戶入職場景。它測試各種條件類型並驗證流程控制系統是否按預期工作。

## 概述

條件測試範例展示了：
- **多種條件類型**：測試不同種類的條件與各種觸發場景
- **流程規則編排**：演示流程規則如何動態修改代理行為
- **自定義條件實現**：創建超越內建類型的領域特定條件
- **結構化輸出整合**：使用條件與結構化數據驗證
- **動態指令更新**：測試條件如何即時改變代理指令

## 測試的條件類型

### 1. 🎯 缺失欄位條件
測試特定數據欄位是否從收集的信息中缺失。

**實現**：
```go
type MissingFieldsCondition struct {
    name   string
    fields []string  // 要檢查缺失的欄位
}
```

**使用案例**：
- 缺失聯絡信息（電子郵件、電話）
- 缺失用戶偏好
- 不完整的個人資料數據

### 2. 📋 完成階段條件
檢查入職流程的當前階段。

**實現**：
```go
type CompletionStageCondition struct {
    name  string
    stage string  // 要匹配的目標階段
}
```

**階段**：
- `basic_info`：初始信息收集
- `contact_details`：電子郵件和電話收集
- `preferences`：興趣和偏好收集
- `completed`：入職完成

### 3. 💬 訊息計數條件
基於對話中的訊息數量觸發。

**實現**：
```go
type MessageCountCondition struct {
    name     string
    minCount int  // 最小訊息閾值
}
```

**使用案例**：
- 長對話優化
- 進度摘要
- 升級觸發

### 4. 🔍 數據鍵存在條件（內建）
使用框架的內建條件檢查數據存在。

**用法**：
```go
condition := agent.NewDataKeyExistsCondition("check_missing", "missing_fields")
```

## 流程規則配置

### 規則 1：聯絡信息收集
**條件**：缺失電子郵件或電話
**動作**：
- 更新指令以特別詢問聯絡詳情
- 推薦 `collect_user_info` 工具
- 提供鼓勵性訊息

### 規則 2：偏好收集
**條件**：缺失偏好數據
**動作**：
- 詢問用戶興趣和愛好
- 推薦收集和驗證工具
- 指導偏好格式（1-5 項，逗號分隔）

### 規則 3：基本信息處理
**條件**：用戶處於 basic_info 階段
**動作**：
- 首先專注於姓名收集
- 設定下一步期望
- 提供結構化入職流程

### 規則 4：長對話優化
**條件**：交換超過 6 條訊息
**動作**：
- 提供進度摘要
- 清楚說明剩餘要求
- 推薦完成驗證

## 工具整合

### CollectInfoTool
收集和驗證用戶信息，具有欄位特定的驗證規則。

**驗證邏輯**：
- **姓名**：最少 2 個字符
- **電子郵件**：必須包含 @ 和 . 符號
- **電話**：最少 10 位數字，必須以 0 開頭
- **偏好**：1-5 項逗號分隔項目

### ValidationTool
評估完成狀態並確定下一步。

**功能**：
- 計算完成百分比
- 識別缺失欄位
- 確定當前階段
- 提供完成建議

## 結構化輸出

代理返回包含以下內容的 `UserStatusOutput` 結構：

```json
{
  "user_id": "user-12345",
  "name": "John Smith",
  "email": "john.smith@email.com", 
  "phone": "0123456789",
  "preferences": ["reading", "cooking", "traveling"],
  "completion_stage": "completed",
  "missing_fields": [],
  "completion_flag": true,
  "message": "入職完成！歡迎加入。"
}
```

## 測試場景

範例運行全面的測試套件：

### 場景 1：初始聯絡
**輸入**："Hi! I want to sign up for your service."
**預期**：基本信息階段條件觸發
**驗證**：基於階段的流程控制

### 場景 2：姓名收集
**輸入**："My name is John Smith"
**預期**：進展到聯絡詳情階段
**驗證**：信息收集和階段進展

### 場景 3：抗拒處理
**輸入**："I don't want to give my contact details yet"
**預期**：缺失聯絡條件觸發
**驗證**：條件指令更新

### 場景 4-5：部分聯絡信息
**輸入**：先電子郵件，然後電話
**預期**：漸進式完成追蹤
**驗證**：增量進度條件

### 場景 6：偏好收集
**輸入**：興趣和愛好信息
**預期**：偏好條件處理
**驗證**：自定義欄位驗證

### 場景 7：延長對話
**輸入**：額外的對話輪次
**預期**：訊息計數條件觸發
**驗證**：長對話優化

### 場景 8：完成驗證
**輸入**："Is that everything? Am I done now?"
**預期**：最終驗證和完成
**驗證**：端到端條件流程

## 運行範例

### 前置條件
1. Go 1.22 或更高版本
2. OpenAI API 金鑰

### 設定
1. **配置環境**：
   ```bash
   cp .env.example .env
   # 編輯 .env 並添加您的 OpenAI API 金鑰
   ```

2. **安裝依賴**：
   ```bash
   go mod tidy
   ```

3. **運行範例**：
   ```bash
   go run main.go
   ```

## 預期輸出

```
🧪 條件測試與流程規則演示
測試各種條件類型和流程規則觸發...
============================================================

🔄 測試 1/8
👤 用戶：Hi! I want to sign up for your service.
🎯 CONDITION[at_basic_info_stage]: Stage 'basic_info' == 'basic_info' ? true
🤖 代理：歡迎！我將幫助您註冊。讓我們從您的姓名開始...
📊 狀態：Stage=basic_info, Missing=[name email phone preferences], Complete=false

🔄 測試 3/8  
👤 用戶：I don't want to give my contact details yet
🎯 CONDITION[missing_contact_info]: Field 'email' is missing
🎯 CONDITION[missing_contact_info]: Field 'phone' is missing
🤖 代理：我理解您對隱私的擔憂。我們需要您的聯絡信息以便...
📊 狀態：Stage=contact_details, Missing=[email phone], Complete=false

🔄 測試 7/8
👤 用戶：Actually, let me add that I also enjoy photography and hiking
🎯 CONDITION[long_conversation]: Message count 7 >= 6 ? true
🤖 代理：很棒的額外興趣！讓我總結一下我們迄今為止收集的信息...
📊 狀態：Stage=completed, Missing=[], Complete=true
```

## 學習成果

這個範例演示了：

1. **條件多樣性**：不同類型的條件用於各種使用案例
2. **流程控制**：條件如何動態修改代理行為
3. **自定義邏輯**：實現領域特定的條件邏輯
4. **狀態管理**：通過結構化輸出追蹤複雜狀態
5. **用戶體驗**：創建自然、適應性的對話流程

## 關鍵實現模式

### 自定義條件實現
```go
func (c *MissingFieldsCondition) Evaluate(ctx context.Context, session agent.Session, data map[string]any) (bool, error) {
    // 從結構化輸出中提取缺失欄位
    missingFields, exists := data["missing_fields"]
    if !exists {
        return false, nil
    }
    
    // 檢查是否有任何目標欄位缺失
    for _, targetField := range c.fields {
        if contains(missing, targetField) {
            return true, nil  // 條件滿足
        }
    }
    
    return false, nil  // 條件不滿足
}
```

### 流程規則創建
```go
rule, err := agent.NewFlowRule("collect-contact-info", missingContactCondition).
    WithDescription("提示用戶提供缺失的聯絡信息").
    WithNewInstructions("專注於收集電子郵件和電話...").
    WithRecommendedTools("collect_user_info").
    WithSystemMessage("需要聯絡信息").
    Build()
```

### 結構化輸出整合
```go
agent.New(
    // ... 其他選項
    agent.WithStructuredOutput(&UserStatusOutput{}),
    agent.WithFlowRules(flowRules...),
)
```

## 架構優勢

1. **模組化**：條件是獨立且可重用的
2. **可測試性**：每個條件都可以單獨測試
3. **靈活性**：易於添加新的條件類型
4. **可維護性**：邏輯和配置的清晰分離
5. **可擴展性**：框架處理複雜的條件編排

## 常見用例

這種模式適用於：
- **用戶入職**：多步驟註冊流程
- **表單驗證**：基於用戶輸入的動態表單行為
- **工作流管理**：條件流程流
- **客戶支援**：上下文感知回應處理
- **電子商務**：適應性結帳和推薦流程