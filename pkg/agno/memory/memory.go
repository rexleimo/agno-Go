package memory

import (
	"sync"

	"github.com/yourusername/agno-go/pkg/agno/types"
)

// Memory manages conversation history
type Memory interface {
	// Add appends a message to memory
	Add(message *types.Message)

	// GetMessages returns all messages
	GetMessages() []*types.Message

	// Clear removes all messages
	Clear()

	// Size returns the number of messages
	Size() int
}

// InMemory provides simple in-memory message storage
type InMemory struct {
	messages []*types.Message
	maxSize  int
	mu       sync.RWMutex
}

// NewInMemory creates a new in-memory storage
func NewInMemory(maxSize int) *InMemory {
	if maxSize <= 0 {
		maxSize = 100 // default
	}
	return &InMemory{
		messages: make([]*types.Message, 0, maxSize),
		maxSize:  maxSize,
	}
}

// Add appends a message to memory
func (m *InMemory) Add(message *types.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, message)

	// Trim if exceeds max size (keep recent messages)
	if len(m.messages) > m.maxSize {
		// Keep system messages and recent messages
		systemMsgs := make([]*types.Message, 0)
		for _, msg := range m.messages {
			if msg.Role == types.RoleSystem {
				systemMsgs = append(systemMsgs, msg)
			}
		}

		// Keep system messages + most recent messages
		keepCount := m.maxSize - len(systemMsgs)
		if keepCount > 0 {
			recentMsgs := m.messages[len(m.messages)-keepCount:]
			m.messages = append(systemMsgs, recentMsgs...)
		} else {
			m.messages = systemMsgs
		}
	}
}

// GetMessages returns all messages
func (m *InMemory) GetMessages() []*types.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a deep copy to prevent external modification
	messages := make([]*types.Message, len(m.messages))
	for i, msg := range m.messages {
		// Create a copy of the message
		msgCopy := *msg
		messages[i] = &msgCopy
	}
	return messages
}

// Clear removes all messages
func (m *InMemory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = make([]*types.Message, 0, m.maxSize)
}

// Size returns the number of messages
func (m *InMemory) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.messages)
}
