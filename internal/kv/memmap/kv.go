package memmap

import (
	"context"
	"sync"

	"gitlab.com/Sh00ty/hootydb/internal/kv"
)

type InMemMapKv struct {
	kv map[kv.Key]kv.Value
	mu sync.RWMutex
}

func New(expectedKeys int) *InMemMapKv {
	return &InMemMapKv{
		kv: make(map[kv.Key]kv.Value, expectedKeys),
		mu: sync.RWMutex{},
	}
}

func (m *InMemMapKv) Get(ctx context.Context, k kv.Key) (kv.Value, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, exists := m.kv[k]
	if !exists {
		return kv.Value{}, kv.MakeNotFound(k)
	}
	return val, nil
}

func (m *InMemMapKv) Set(ctx context.Context, k kv.Key, v kv.Value) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	val, exists := m.kv[k]
	if exists && val.Ver.IsBigger(v.Ver) {
		return kv.MakeOldVerError(k, val.Ver, v.Ver)
	}
	m.kv[k] = v
	return nil
}
