package memory

import (
	"testing"

	"github.com/rexleimo/agno-go/pkg/agno/types"
)

func TestInMemory_Add(t *testing.T) {
	mem := NewInMemory(10)

	msg1 := types.NewUserMessage("hello")
	mem.Add(msg1)

	if mem.Size() != 1 {
		t.Errorf("expected size 1, got %d", mem.Size())
	}

	messages := mem.GetMessages()
	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}

	if messages[0].Content != "hello" {
		t.Errorf("expected content 'hello', got '%s'", messages[0].Content)
	}
}

func TestInMemory_MaxSize(t *testing.T) {
	maxSize := 5
	mem := NewInMemory(maxSize)

	// Add system message first
	mem.Add(types.NewSystemMessage("system"))

	// Add more messages than max size
	for i := 0; i < 10; i++ {
		mem.Add(types.NewUserMessage("message"))
	}

	size := mem.Size()
	if size > maxSize {
		t.Errorf("expected size <= %d, got %d", maxSize, size)
	}

	// Check that system message is preserved
	messages := mem.GetMessages()
	hasSystem := false
	for _, msg := range messages {
		if msg.Role == types.RoleSystem {
			hasSystem = true
			break
		}
	}
	if !hasSystem {
		t.Error("expected system message to be preserved")
	}
}

func TestInMemory_Clear(t *testing.T) {
	mem := NewInMemory(10)

	mem.Add(types.NewUserMessage("hello"))
	mem.Add(types.NewUserMessage("world"))

	if mem.Size() != 2 {
		t.Errorf("expected size 2 before clear, got %d", mem.Size())
	}

	mem.Clear()

	if mem.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", mem.Size())
	}
}

func TestInMemory_GetMessages_Copy(t *testing.T) {
	mem := NewInMemory(10)
	mem.Add(types.NewUserMessage("original"))

	messages := mem.GetMessages()
	messages[0].Content = "modified"

	// Original should not be modified
	original := mem.GetMessages()
	if original[0].Content != "original" {
		t.Error("GetMessages should return a copy, not the original slice")
	}
}
