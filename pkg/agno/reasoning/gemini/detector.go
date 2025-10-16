package gemini

import (
	"strings"

	"github.com/rexleimo/agno-go/pkg/agno/models"
)

// Detector 检测 Gemini 推理模型
// Detector detects Gemini reasoning models
type Detector struct{}

// IsReasoningModel 检查是否为 Gemini 推理模型
// Checks if this is a Gemini reasoning model
func (d *Detector) IsReasoningModel(model models.Model) bool {
	// 检查提供商 / Check provider
	if model.GetProvider() != "gemini" {
		return false
	}

	// 检查模型 ID 是否为 2.5+ / Check if model ID is 2.5+
	modelID := strings.ToLower(model.GetID())
	if strings.Contains(modelID, "2.5") ||
		strings.Contains(modelID, "thinking") ||
		strings.Contains(modelID, "flash-thinking") {
		return true
	}

	return false
}

// Provider 返回提供商名称
// Returns the provider name
func (d *Detector) Provider() string {
	return "gemini"
}
