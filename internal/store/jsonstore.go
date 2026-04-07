package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/channinghe/ems2sns/internal/model"
)

// JSONStore implements Store using a JSON file
type JSONStore struct {
	path string
	mu   sync.Mutex
}

func NewJSONStore(path string) *JSONStore {
	return &JSONStore{path: path}
}

func (s *JSONStore) Load() (map[string]*model.Subscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*model.Subscription), nil
		}
		return nil, fmt.Errorf("reading %s: %w", s.path, err)
	}

	subs := make(map[string]*model.Subscription)
	if err := json.Unmarshal(data, &subs); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", s.path, err)
	}
	return subs, nil
}

func (s *JSONStore) Save(subs map[string]*model.Subscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return fmt.Errorf("creating directory for %s: %w", s.path, err)
	}

	data, err := json.MarshalIndent(subs, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling subscriptions: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", s.path, err)
	}
	return nil
}
