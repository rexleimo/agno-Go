// Package utils provides utility functions for the agno framework
// utils 包为 agno 框架提供工具函数
package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// JSONSerializer is a custom JSON serializer for handling special types
// JSONSerializer 是一个自定义 JSON 序列化器，用于处理特殊类型
type JSONSerializer struct{}

// MarshalJSON provides custom JSON serialization for objects not serializable by default
// MarshalJSON 为默认情况下不可序列化的对象提供自定义 JSON 序列化
//
// Handles:
// - time.Time objects -> ISO 8601 format strings
// - Custom types implementing Stringer interface -> string representation
//
// 处理:
// - time.Time 对象 -> ISO 8601 格式字符串
// - 实现了 Stringer 接口的自定义类型 -> 字符串表示
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(convertValue(v))
}

// convertValue converts special types to JSON-serializable values
// convertValue 将特殊类型转换为可 JSON 序列化的值
func convertValue(v interface{}) interface{} {
	switch val := v.(type) {
	case time.Time:
		// Convert time.Time to ISO 8601 format
		// 将 time.Time 转换为 ISO 8601 格式
		return val.Format(time.RFC3339)

	case *time.Time:
		if val != nil {
			return val.Format(time.RFC3339)
		}
		return nil

	case map[string]interface{}:
		// Recursively convert map values
		// 递归转换 map 中的值
		result := make(map[string]interface{})
		for k, v := range val {
			result[k] = convertValue(v)
		}
		return result

	case []interface{}:
		// Recursively convert slice values
		// 递归转换切片中的值
		result := make([]interface{}, len(val))
		for i, v := range val {
			result[i] = convertValue(v)
		}
		return result

	case fmt.Stringer:
		// Use String() method for types implementing Stringer
		// 对实现了 Stringer 接口的类型使用 String() 方法
		return val.String()

	default:
		return v
	}
}

// ToJSON converts any value to JSON bytes with custom serialization
// ToJSON 使用自定义序列化将任意值转换为 JSON 字节
func ToJSON(v interface{}) ([]byte, error) {
	converted := convertValue(v)
	return json.Marshal(converted)
}

// ToJSONString converts any value to JSON string with custom serialization
// ToJSONString 使用自定义序列化将任意值转换为 JSON 字符串
func ToJSONString(v interface{}) (string, error) {
	bytes, err := ToJSON(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// MustToJSONString converts any value to JSON string, panics on error
// MustToJSONString 将任意值转换为 JSON 字符串，出错时 panic
func MustToJSONString(v interface{}) string {
	str, err := ToJSONString(v)
	if err != nil {
		panic(fmt.Sprintf("failed to serialize to JSON: %v", err))
	}
	return str
}
