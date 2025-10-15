# Google Sheets 工具包示例 / Google Sheets Toolkit Example

本示例演示如何在 Agno-Go Agent 中使用 Google Sheets 工具包来操作 Google 电子表格。

This example demonstrates how to use the Google Sheets toolkit in an Agno-Go Agent to manipulate Google Spreadsheets.

## 前置要求 / Prerequisites

### 1. 创建 Google Cloud 服务账号 / Create Google Cloud Service Account

1. 访问 [Google Cloud Console](https://console.cloud.google.com/)
2. 创建新项目或选择现有项目 / Create a new project or select an existing one
3. 启用 Google Sheets API:
   - 在左侧菜单中，选择 "APIs & Services" > "Library"
   - 搜索 "Google Sheets API" 并启用
4. 创建服务账号凭证:
   - 进入 "APIs & Services" > "Credentials"
   - 点击 "Create Credentials" > "Service Account"
   - 填写服务账号名称和描述
   - 点击 "Create and Continue"
   - 跳过授予角色步骤（可选）
   - 点击 "Done"
5. 生成 JSON 密钥:
   - 点击刚创建的服务账号
   - 进入 "Keys" 标签
   - 点击 "Add Key" > "Create new key"
   - 选择 "JSON" 格式
   - 下载 JSON 文件

### 2. 授予服务账号访问权限 / Grant Service Account Access

1. 打开你的 Google 电子表格
2. 点击右上角的 "Share" 按钮
3. 添加服务账号的邮箱地址（在 JSON 文件的 `client_email` 字段中）
4. 授予 "Editor" 权限
5. 点击 "Send"

## 设置环境变量 / Environment Variables

```bash
# OpenAI API Key (用于 Agent)
# OpenAI API Key (for the Agent)
export OPENAI_API_KEY=sk-...

# Google Sheets 凭证 (JSON 文件路径或 JSON 字符串)
# Google Sheets credentials (JSON file path or JSON string)
export GOOGLE_SHEETS_CREDENTIALS=/path/to/your/credentials.json

# 电子表格 ID (从 URL 中获取)
# Spreadsheet ID (get from URL)
# 例如: https://docs.google.com/spreadsheets/d/SPREADSHEET_ID/edit
export SPREADSHEET_ID=your-spreadsheet-id
```

## 运行示例 / Run the Example

### 演示模式 (仅显示可用工具) / Demo Mode (Show Available Tools Only)

```bash
# 不设置 SPREADSHEET_ID，仅显示工具列表
# Without SPREADSHEET_ID, only show available tools
export OPENAI_API_KEY=sk-...
export GOOGLE_SHEETS_CREDENTIALS=/path/to/credentials.json

go run main.go
```

### 完整模式 (实际操作电子表格) / Full Mode (Actual Spreadsheet Operations)

```bash
# 设置所有必需的环境变量
# Set all required environment variables
export OPENAI_API_KEY=sk-...
export GOOGLE_SHEETS_CREDENTIALS=/path/to/credentials.json
export SPREADSHEET_ID=your-spreadsheet-id

go run main.go
```

## 功能演示 / Features Demonstrated

### 1. 读取数据 / Read Data

```go
// 读取 Sheet1!A1:B5 范围的数据
// Read data from Sheet1!A1:B5 range
query := "请读取电子表格的 Sheet1!A1:B5 范围的数据"
output, err := agent.Run(ctx, query)
```

### 2. 写入数据 / Write Data

```go
// 写入数据到 Sheet1!A1 位置
// Write data to Sheet1!A1
query := `请将以下数据写入电子表格:
第一行: Name, Score
第二行: Alice, 95
第三行: Bob, 87`
output, err := agent.Run(ctx, query)
```

### 3. 追加数据 / Append Data

```go
// 追加新行到 Sheet1
// Append new row to Sheet1
query := "请在电子表格中追加一行: Charlie, 92"
output, err := agent.Run(ctx, query)
```

## 可用工具函数 / Available Tool Functions

### read_range

读取指定范围的数据 / Read data from a specified range

**参数 / Parameters:**
- `spreadsheet_id` (string, 必需): 电子表格 ID
- `range` (string, 必需): 要读取的范围，例如 'Sheet1!A1:D10'

**返回 / Returns:**
```json
{
  "range": "Sheet1!A1:B2",
  "values": [
    ["Name", "Score"],
    ["Alice", "95"]
  ],
  "rows": 2
}
```

### write_range

写入数据到指定范围 / Write data to a specified range

**参数 / Parameters:**
- `spreadsheet_id` (string, 必需): 电子表格 ID
- `range` (string, 必需): 要写入的范围，例如 'Sheet1!A1'
- `values` (array, 必需): 要写入的数据（二维数组）

**返回 / Returns:**
```json
{
  "updated_range": "Sheet1!A1:B3",
  "updated_rows": 3,
  "updated_cells": 6,
  "updated_columns": 2
}
```

### append_rows

追加行到表格 / Append rows to the sheet

**参数 / Parameters:**
- `spreadsheet_id` (string, 必需): 电子表格 ID
- `range` (string, 必需): 追加范围，例如 'Sheet1!A:D'
- `values` (array, 必需): 要追加的数据（二维数组）

**返回 / Returns:**
```json
{
  "updated_range": "Sheet1!A4:B4",
  "updated_rows": 1,
  "updated_cells": 2,
  "updated_columns": 2
}
```

## 使用 JSON 字符串凭证 / Using JSON String Credentials

除了使用文件路径，你也可以直接使用 JSON 字符串:

Instead of using a file path, you can also use a JSON string directly:

```bash
export GOOGLE_SHEETS_CREDENTIALS='{"type":"service_account","project_id":"your-project",...}'
```

## 故障排除 / Troubleshooting

### 错误: "The caller does not have permission"

- 确保已将服务账号邮箱添加到电子表格的共享列表中
- 确保授予了 "Editor" 权限

### 错误: "Requested entity was not found"

- 检查 SPREADSHEET_ID 是否正确
- 检查 range 参数格式是否正确（例如 'Sheet1!A1:B5'）

### 错误: "Invalid credentials"

- 检查 credentials JSON 文件内容是否完整
- 确保 JSON 格式正确
- 确认服务账号是否已启用 Google Sheets API

## 进一步了解 / Learn More

- [Google Sheets API 文档](https://developers.google.com/sheets/api)
- [服务账号认证指南](https://cloud.google.com/iam/docs/service-accounts)
- [Agno-Go 工具包文档](https://github.com/rexleimo/agno-go)
