package utils

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// Custom type implementing Stringer interface
// 实现 Stringer 接口的自定义类型
type CustomStatus string

const (
	StatusActive   CustomStatus = "active"
	StatusInactive CustomStatus = "inactive"
)

func (s CustomStatus) String() string {
	return string(s)
}

// TestConvertValue_Time tests time.Time conversion
// TestConvertValue_Time 测试 time.Time 转换
func TestConvertValue_Time(t *testing.T) {
	testTime := time.Date(2025, 10, 4, 12, 30, 45, 0, time.UTC)

	result := convertValue(testTime)
	expected := "2025-10-04T12:30:45Z"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestConvertValue_TimePointer tests *time.Time conversion
// TestConvertValue_TimePointer 测试 *time.Time 转换
func TestConvertValue_TimePointer(t *testing.T) {
	testTime := time.Date(2025, 10, 4, 12, 30, 45, 0, time.UTC)

	result := convertValue(&testTime)
	expected := "2025-10-04T12:30:45Z"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestConvertValue_NilTimePointer tests nil *time.Time conversion
// TestConvertValue_NilTimePointer 测试 nil *time.Time 转换
func TestConvertValue_NilTimePointer(t *testing.T) {
	var nilTime *time.Time = nil

	result := convertValue(nilTime)

	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

// TestConvertValue_Map tests map conversion with nested time
// TestConvertValue_Map 测试包含嵌套时间的 map 转换
func TestConvertValue_Map(t *testing.T) {
	testTime := time.Date(2025, 10, 4, 12, 30, 45, 0, time.UTC)
	input := map[string]interface{}{
		"name":      "test",
		"timestamp": testTime,
		"count":     42,
	}

	result := convertValue(input)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}

	if resultMap["name"] != "test" {
		t.Errorf("Expected name=test, got %v", resultMap["name"])
	}

	if resultMap["timestamp"] != "2025-10-04T12:30:45Z" {
		t.Errorf("Expected ISO time, got %v", resultMap["timestamp"])
	}

	if resultMap["count"] != 42 {
		t.Errorf("Expected count=42, got %v", resultMap["count"])
	}
}

// TestConvertValue_Slice tests slice conversion with time values
// TestConvertValue_Slice 测试包含时间值的切片转换
func TestConvertValue_Slice(t *testing.T) {
	testTime := time.Date(2025, 10, 4, 12, 30, 45, 0, time.UTC)
	input := []interface{}{
		"test",
		testTime,
		123,
	}

	result := convertValue(input)
	resultSlice, ok := result.([]interface{})
	if !ok {
		t.Fatalf("Expected []interface{}, got %T", result)
	}

	if len(resultSlice) != 3 {
		t.Fatalf("Expected 3 elements, got %d", len(resultSlice))
	}

	if resultSlice[0] != "test" {
		t.Errorf("Expected first element=test, got %v", resultSlice[0])
	}

	if resultSlice[1] != "2025-10-04T12:30:45Z" {
		t.Errorf("Expected ISO time, got %v", resultSlice[1])
	}

	if resultSlice[2] != 123 {
		t.Errorf("Expected third element=123, got %v", resultSlice[2])
	}
}

// TestConvertValue_Stringer tests custom Stringer type conversion
// TestConvertValue_Stringer 测试自定义 Stringer 类型转换
func TestConvertValue_Stringer(t *testing.T) {
	status := StatusActive

	result := convertValue(status)
	expected := "active"

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestConvertValue_PrimitiveTypes tests primitive type pass-through
// TestConvertValue_PrimitiveTypes 测试基本类型的透传
func TestConvertValue_PrimitiveTypes(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"string", "hello"},
		{"int", 42},
		{"float", 3.14},
		{"bool", true},
		{"nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertValue(tt.input)
			if result != tt.input {
				t.Errorf("Expected %v, got %v", tt.input, result)
			}
		})
	}
}

// TestToJSON tests full JSON serialization
// TestToJSON 测试完整的 JSON 序列化
func TestToJSON(t *testing.T) {
	testTime := time.Date(2025, 10, 4, 12, 30, 45, 0, time.UTC)
	input := map[string]interface{}{
		"id":        "123",
		"timestamp": testTime,
		"status":    StatusActive,
		"count":     42,
	}

	jsonBytes, err := ToJSON(input)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Parse back to verify
	// 解析回来验证
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["id"] != "123" {
		t.Errorf("Expected id=123, got %v", result["id"])
	}

	if result["timestamp"] != "2025-10-04T12:30:45Z" {
		t.Errorf("Expected ISO timestamp, got %v", result["timestamp"])
	}

	if result["status"] != "active" {
		t.Errorf("Expected status=active, got %v", result["status"])
	}

	// JSON numbers are float64
	// JSON 数字是 float64 类型
	if result["count"] != float64(42) {
		t.Errorf("Expected count=42, got %v", result["count"])
	}
}

// TestToJSONString tests JSON string serialization
// TestToJSONString 测试 JSON 字符串序列化
func TestToJSONString(t *testing.T) {
	testTime := time.Date(2025, 10, 4, 12, 30, 45, 0, time.UTC)
	input := map[string]interface{}{
		"timestamp": testTime,
	}

	jsonStr, err := ToJSONString(input)
	if err != nil {
		t.Fatalf("ToJSONString failed: %v", err)
	}

	expected := `{"timestamp":"2025-10-04T12:30:45Z"}`
	if jsonStr != expected {
		t.Errorf("Expected %s, got %s", expected, jsonStr)
	}
}

// TestMustToJSONString tests panic-on-error JSON serialization
// TestMustToJSONString 测试出错时 panic 的 JSON 序列化
func TestMustToJSONString(t *testing.T) {
	testTime := time.Date(2025, 10, 4, 12, 30, 45, 0, time.UTC)
	input := map[string]interface{}{
		"timestamp": testTime,
	}

	jsonStr := MustToJSONString(input)
	expected := `{"timestamp":"2025-10-04T12:30:45Z"}`

	if jsonStr != expected {
		t.Errorf("Expected %s, got %s", expected, jsonStr)
	}
}

// TestMustToJSONString_ValidInput tests normal case without panic
// TestMustToJSONString_ValidInput 测试不会 panic 的正常情况
func TestMustToJSONString_ValidInput(t *testing.T) {
	// This should not panic
	// 这不应该 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustToJSONString panicked unexpectedly: %v", r)
		}
	}()

	result := MustToJSONString(map[string]interface{}{"key": "value"})
	if result == "" {
		t.Error("Expected non-empty string")
	}
}

// TestMarshalJSON tests the MarshalJSON function
// TestMarshalJSON 测试 MarshalJSON 函数
func TestMarshalJSON(t *testing.T) {
	testTime := time.Date(2025, 10, 4, 12, 30, 45, 0, time.UTC)

	jsonBytes, err := MarshalJSON(testTime)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var result string
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	expected := "2025-10-04T12:30:45Z"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

// TestComplexNestedStructure tests deeply nested structure with mixed types
// TestComplexNestedStructure 测试包含混合类型的深层嵌套结构
func TestComplexNestedStructure(t *testing.T) {
	testTime := time.Date(2025, 10, 4, 12, 30, 45, 0, time.UTC)
	input := map[string]interface{}{
		"metadata": map[string]interface{}{
			"created_at": testTime,
			"status":     StatusActive,
			"tags": []interface{}{
				"tag1",
				testTime,
				StatusInactive,
			},
		},
		"data": []interface{}{
			map[string]interface{}{
				"id":   1,
				"time": testTime,
			},
			map[string]interface{}{
				"id":   2,
				"time": testTime,
			},
		},
	}

	jsonBytes, err := ToJSON(input)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify it's valid JSON
	// 验证它是有效的 JSON
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Verify nested values
	// 验证嵌套值
	metadata := result["metadata"].(map[string]interface{})
	if metadata["created_at"] != "2025-10-04T12:30:45Z" {
		t.Errorf("Expected ISO timestamp in metadata, got %v", metadata["created_at"])
	}

	if metadata["status"] != "active" {
		t.Errorf("Expected status=active, got %v", metadata["status"])
	}
}

// BenchmarkToJSON benchmarks JSON serialization performance
// BenchmarkToJSON 基准测试 JSON 序列化性能
func BenchmarkToJSON(b *testing.B) {
	testTime := time.Date(2025, 10, 4, 12, 30, 45, 0, time.UTC)
	input := map[string]interface{}{
		"id":        "123",
		"timestamp": testTime,
		"status":    StatusActive,
		"count":     42,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := ToJSON(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConvertValue benchmarks value conversion performance
// BenchmarkConvertValue 基准测试值转换性能
func BenchmarkConvertValue(b *testing.B) {
	testTime := time.Date(2025, 10, 4, 12, 30, 45, 0, time.UTC)
	input := map[string]interface{}{
		"id":        "123",
		"timestamp": testTime,
		"status":    StatusActive,
		"count":     42,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = convertValue(input)
	}
}

// ExampleToJSON demonstrates basic usage of ToJSON
// ExampleToJSON 演示 ToJSON 的基本用法
func ExampleToJSON() {
	testTime := time.Date(2025, 10, 4, 12, 30, 45, 0, time.UTC)
	data := map[string]interface{}{
		"event":     "user_login",
		"timestamp": testTime,
		"user_id":   "12345",
	}

	jsonBytes, err := ToJSON(data)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(string(jsonBytes))
	// Output: {"event":"user_login","timestamp":"2025-10-04T12:30:45Z","user_id":"12345"}
}

// ExampleToJSONString demonstrates ToJSONString usage
// ExampleToJSONString 演示 ToJSONString 的用法
func ExampleToJSONString() {
	testTime := time.Date(2025, 10, 4, 12, 30, 45, 0, time.UTC)
	data := map[string]interface{}{
		"timestamp": testTime,
	}

	jsonStr, err := ToJSONString(data)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(jsonStr)
	// Output: {"timestamp":"2025-10-04T12:30:45Z"}
}

// TestToJSONString_Error tests error handling in ToJSONString
// TestToJSONString_Error 测试 ToJSONString 中的错误处理
func TestToJSONString_Error(t *testing.T) {
	// Create an unserialiable type (channel)
	// 创建一个不可序列化的类型（channel）
	ch := make(chan int)
	input := map[string]interface{}{
		"channel": ch,
	}

	_, err := ToJSONString(input)
	if err == nil {
		t.Error("Expected error for unserializable type, got nil")
	}
}

// TestToJSON_Error tests error handling in ToJSON
// TestToJSON_Error 测试 ToJSON 中的错误处理
func TestToJSON_Error(t *testing.T) {
	// Create an unserialiable type (channel)
	// 创建一个不可序列化的类型（channel）
	ch := make(chan int)
	input := map[string]interface{}{
		"channel": ch,
	}

	_, err := ToJSON(input)
	if err == nil {
		t.Error("Expected error for unserializable type, got nil")
	}
}

// TestMustToJSONString_Panic tests panic behavior
// TestMustToJSONString_Panic 测试 panic 行为
func TestMustToJSONString_Panic(t *testing.T) {
	// This should panic
	// 这应该 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for unserializable type, but didn't panic")
		}
	}()

	// Create an unserialiable type (channel)
	// 创建一个不可序列化的类型（channel）
	ch := make(chan int)
	input := map[string]interface{}{
		"channel": ch,
	}

	_ = MustToJSONString(input)
}
