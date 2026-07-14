package memory

import (
	"errors"
	"sync"

	"study.local/agentgo/internal/llm"
)

var ErrInvalidKey = errors.New("tenantID、userID 和 conversationID 不能为空")

type Store interface {
	History(tenantID, userID, conversationID string) ([]llm.InputItem, error)
	Append(tenantID, userID, conversationID string, item llm.InputItem) error
	AppendMany(tenantID, userID, conversationID string, items ...llm.InputItem) error
	Delete(tenantID, userID, conversationID string) error
}

type WindowStore struct {
	mu          sync.RWMutex
	maxMessages int
	items       map[string][]llm.InputItem
}

func NewWindowStore(maxMessages int) *WindowStore {
	if maxMessages < 2 {
		maxMessages = 2
	}
	return &WindowStore{
		maxMessages: maxMessages,
		items:       make(map[string][]llm.InputItem),
	}
}

func (s *WindowStore) History(tenantID, userID, conversationID string) ([]llm.InputItem, error) {
	key, err := makeKey(tenantID, userID, conversationID)
	if err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	history := s.items[key]
	result := make([]llm.InputItem, 0, len(history))
	for _, item := range history {
		result = append(result, cloneItem(item))
	}
	return result, nil
}

func (s *WindowStore) Append(tenantID, userID, conversationID string, item llm.InputItem) error {
	return s.AppendMany(tenantID, userID, conversationID, item)
}

func (s *WindowStore) AppendMany(tenantID, userID, conversationID string, items ...llm.InputItem) error {
	key, err := makeKey(tenantID, userID, conversationID)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	history := append([]llm.InputItem(nil), s.items[key]...)
	for _, item := range items {
		history = append(history, cloneItem(item))
	}
	if len(history) > s.maxMessages {
		history = append([]llm.InputItem(nil), history[len(history)-s.maxMessages:]...)
	}
	s.items[key] = history
	return nil
}

func (s *WindowStore) Delete(tenantID, userID, conversationID string) error {
	key, err := makeKey(tenantID, userID, conversationID)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
	return nil
}

func makeKey(tenantID, userID, conversationID string) (string, error) {
	if tenantID == "" || userID == "" || conversationID == "" {
		return "", ErrInvalidKey
	}
	return tenantID + "\x00" + userID + "\x00" + conversationID, nil
}

func cloneItem(item llm.InputItem) llm.InputItem {
	copyItem := make(llm.InputItem, len(item))
	for key, value := range item {
		copyItem[key] = value
	}
	return copyItem
}
