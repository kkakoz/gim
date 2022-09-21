package mapx

import (
	"golang.org/x/exp/constraints"
	"sync"
)

type SyncMap[K constraints.Ordered, V any] struct {
	m map[K]V
	sync.RWMutex
}

func NewSyncMap[K constraints.Ordered, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{m: map[K]V{}}
}

func (m *SyncMap[K, V]) Add(k K, v V) {
	m.Lock()
	defer m.Unlock()
	m.m[k] = v
}

func (m *SyncMap[K, V]) Get(k K) (V, bool) {
	m.RLock()
	defer m.RUnlock()
	v, ok := m.m[k]
	return v, ok
}

func (m *SyncMap[K, V]) Delete(k K) {
	m.Lock()
	defer m.Unlock()
	delete(m.m, k)
}

func (m *SyncMap[K, V]) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.m)
}

func (m *SyncMap[K, V]) Values() []V {
	m.RLock()
	defer m.RUnlock()
	res := make([]V, 0, len(m.m))
	for _, v := range m.m {
		res = append(res, v)
	}
	return res
}
