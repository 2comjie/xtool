package syncx

import "sync"

type SafeMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{m: make(map[K]V)}
}

func (s *SafeMap[K, V]) Store(key K, val V) {
	s.mu.Lock()
	s.m[key] = val
	s.mu.Unlock()
}

func (s *SafeMap[K, V]) Load(key K) (V, bool) {
	s.mu.RLock()
	v, ok := s.m[key]
	s.mu.RUnlock()
	return v, ok
}

func (s *SafeMap[K, V]) Delete(key K) {
	s.mu.Lock()
	delete(s.m, key)
	s.mu.Unlock()
}

func (s *SafeMap[K, V]) LoadAndDelete(key K) (V, bool) {
	s.mu.Lock()
	v, ok := s.m[key]
	if ok {
		delete(s.m, key)
	}
	s.mu.Unlock()
	return v, ok
}

func (s *SafeMap[K, V]) LoadOrStore(key K, val V) (V, bool) {
	s.mu.Lock()
	if v, ok := s.m[key]; ok {
		s.mu.Unlock()
		return v, true
	}
	s.m[key] = val
	s.mu.Unlock()
	return val, false
}

func (s *SafeMap[K, V]) Range(fn func(K, V) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.m {
		if !fn(k, v) {
			break
		}
	}
}

func (s *SafeMap[K, V]) Len() int {
	s.mu.RLock()
	n := len(s.m)
	s.mu.RUnlock()
	return n
}
