// Package a2a implements the Agent-to-Agent (A2A) communication protocol
// based on JSON-RPC 2.0 standard.
//
// A2A包实现基于JSON-RPC 2.0标准的Agent间通信协议。
package a2a

import "time"

// JSONRPC2Request represents a JSON-RPC 2.0 request
// JSONRPC2Request 表示 JSON-RPC 2.0 请求
type JSONRPC2Request struct {
	JSONRPC string        `json:"jsonrpc"` // Must be "2.0"
	Method  string        `json:"method"`  // "message/send" or "message/stream"
	ID      string        `json:"id"`      // Request ID
	Params  RequestParams `json:"params"`  // Request parameters
}

// JSONRPC2Response represents a JSON-RPC 2.0 response
// JSONRPC2Response 表示 JSON-RPC 2.0 响应
type JSONRPC2Response struct {
	JSONRPC string  `json:"jsonrpc"` // Must be "2.0"
	ID      string  `json:"id"`      // Request ID
	Result  *Result `json:"result,omitempty"`
	Error   *Error  `json:"error,omitempty"`
}

// RequestParams contains the message parameter
// RequestParams 包含消息参数
type RequestParams struct {
	Message Message `json:"message"`
}

// Result contains the task result
// Result 包含任务结果
type Result struct {
	Task Task `json:"task"`
}

// Error represents a JSON-RPC 2.0 error
// Error 表示 JSON-RPC 2.0 错误
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Standard JSON-RPC 2.0 error codes
// 标准 JSON-RPC 2.0 错误码
const (
	ParseError     = -32700 // Invalid JSON
	InvalidRequest = -32600 // Invalid request object
	MethodNotFound = -32601 // Method not found
	InvalidParams  = -32602 // Invalid parameters
	InternalError  = -32603 // Internal error
	ServerError    = -32000 // Generic server error
)

// Message represents an A2A message
// Message 表示 A2A 消息
type Message struct {
	MessageID string `json:"messageId"`           // Unique message ID
	Role      string `json:"role"`                // "user" or "agent"
	AgentID   string `json:"agentId"`             // Target agent/team/workflow ID
	ContextID string `json:"contextId"`           // Session context ID
	Parts     []Part `json:"parts"`               // Message parts
	Timestamp string `json:"timestamp,omitempty"` // ISO 8601 timestamp
}

// Part represents a message part (text, file, or data)
// Part 表示消息部分（文本、文件或数据）
type Part struct {
	Kind string `json:"kind"` // "text", "file", or "data"

	// Text part fields
	Text *string `json:"text,omitempty"`

	// File part fields
	File *FilePart `json:"file,omitempty"`

	// Data part fields
	Data *DataPart `json:"data,omitempty"`
}

// FilePart represents a file in the message
// FilePart 表示消息中的文件
type FilePart struct {
	URI      *string `json:"uri,omitempty"`   // File URI
	Bytes    *string `json:"bytes,omitempty"` // Base64-encoded bytes
	MimeType string  `json:"mimeType"`        // MIME type
	Name     *string `json:"name,omitempty"`  // File name
}

// DataPart represents structured data in the message
// DataPart 表示消息中的结构化数据
type DataPart struct {
	Content  string  `json:"content"`        // Data content
	MimeType string  `json:"mimeType"`       // MIME type
	Name     *string `json:"name,omitempty"` // Data name
}

// Task represents the response task
// Task 表示响应任务
type Task struct {
	ID        string     `json:"id"`                  // Task ID
	ContextID string     `json:"context_id"`          // Session context
	Status    TaskStatus `json:"status"`              // Task status
	History   []Message  `json:"history"`             // Message history
	Artifacts []Artifact `json:"artifacts,omitempty"` // Output artifacts
	Error     *TaskError `json:"error,omitempty"`     // Error if failed
}

// TaskStatus represents task completion status
// TaskStatus 表示任务完成状态
type TaskStatus string

const (
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// Artifact represents an output artifact
// Artifact 表示输出工件
type Artifact struct {
	ArtifactID string    `json:"artifactId"`
	Name       string    `json:"name"`
	Parts      []Part    `json:"parts"`
	CreatedAt  time.Time `json:"createdAt"`
}

// TaskError represents a task error
// TaskError 表示任务错误
type TaskError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// StreamEvent represents a server-sent event
// StreamEvent 表示服务器发送事件
type StreamEvent struct {
	Event string `json:"event"` // Event type
	Data  any    `json:"data"`  // Event data
}

// TaskStatusUpdate represents a task status update event
// TaskStatusUpdate 表示任务状态更新事件
type TaskStatusUpdate struct {
	TaskID    string     `json:"taskId"`
	ContextID string     `json:"contextId"`
	Status    TaskStatus `json:"status"`
	Message   *Message   `json:"message,omitempty"`
	Error     *TaskError `json:"error,omitempty"`
}
