package a2a

import (
	"fmt"
)

// ValidateRequest validates a JSON-RPC 2.0 request
// ValidateRequest 验证 JSON-RPC 2.0 请求
func ValidateRequest(req *JSONRPC2Request) error {
	// Check JSON-RPC version
	// 检查 JSON-RPC 版本
	if req.JSONRPC != "2.0" {
		return fmt.Errorf("invalid jsonrpc version: expected '2.0', got '%s'", req.JSONRPC)
	}

	// Check method
	// 检查方法
	if req.Method != "message/send" && req.Method != "message/stream" {
		return fmt.Errorf("invalid method: expected 'message/send' or 'message/stream', got '%s'", req.Method)
	}

	// Check request ID
	// 检查请求ID
	if req.ID == "" {
		return fmt.Errorf("request id is required")
	}

	// Validate message
	// 验证消息
	if err := ValidateMessage(&req.Params.Message); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	return nil
}

// ValidateMessage validates a message
// ValidateMessage 验证消息
func ValidateMessage(msg *Message) error {
	// Check message ID
	// 检查消息ID
	if msg.MessageID == "" {
		return fmt.Errorf("messageId is required")
	}

	// Check role
	// 检查角色
	if msg.Role != "user" && msg.Role != "agent" {
		return fmt.Errorf("invalid role: expected 'user' or 'agent', got '%s'", msg.Role)
	}

	// Check agent ID
	// 检查agent ID
	if msg.AgentID == "" {
		return fmt.Errorf("agentId is required")
	}

	// Check parts
	// 检查parts
	if len(msg.Parts) == 0 {
		return fmt.Errorf("at least one message part is required")
	}

	// Validate each part
	// 验证每个部分
	for i, part := range msg.Parts {
		if err := ValidatePart(&part); err != nil {
			return fmt.Errorf("invalid part at index %d: %w", i, err)
		}
	}

	return nil
}

// ValidatePart validates a message part
// ValidatePart 验证消息部分
func ValidatePart(part *Part) error {
	switch part.Kind {
	case "text":
		if part.Text == nil || *part.Text == "" {
			return fmt.Errorf("text part requires non-empty text field")
		}
	case "file":
		if part.File == nil {
			return fmt.Errorf("file part requires file field")
		}
		if part.File.URI == nil && part.File.Bytes == nil {
			return fmt.Errorf("file part requires either uri or bytes field")
		}
		if part.File.MimeType == "" {
			return fmt.Errorf("file part requires mimeType field")
		}
	case "data":
		if part.Data == nil {
			return fmt.Errorf("data part requires data field")
		}
		if part.Data.Content == "" {
			return fmt.Errorf("data part requires non-empty content field")
		}
		if part.Data.MimeType == "" {
			return fmt.Errorf("data part requires mimeType field")
		}
	default:
		return fmt.Errorf("invalid part kind: expected 'text', 'file', or 'data', got '%s'", part.Kind)
	}

	return nil
}
