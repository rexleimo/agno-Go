package anthropic

import (
	"github.com/rexleimo/agno-go/pkg/agno/models"
)

// Detector 检测 Anthropic 推理模型
// Detector detects Anthropic reasoning models
type Detector struct{}

// IsReasoningModel 检查是否为 Anthropic 推理模型
// Checks if this is an Anthropic reasoning model
func (d *Detector) IsReasoningModel(model models.Model) bool {
	// 检查提供商 / Check provider
	provider := model.GetProvider()
	if provider != "anthropic" && provider != "claude" {
		return false
	}

	// 注意: Anthropic 推理需要显式配置 thinking 参数
	// Note: Anthropic reasoning requires explicit thinking configuration
	// 这里我们假设如果是 Claude 模型,可能支持推理
	// Here we assume if it's a Claude model, it may support reasoning
	// 实际检测需要访问模型配置(后续优化)
	// Actual detection needs access to model config (future enhancement)

	return false // 默认 false,需要用户显式启用
}

// Provider 返回提供商名称
// Returns the provider name
func (d *Detector) Provider() string {
	return "anthropic"
}
