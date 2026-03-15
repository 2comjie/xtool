package poolx

import "sync"

// Pool 泛型对象池，封装 sync.Pool 提供类型安全的 Get/Put
type Pool[T any] struct {
	pool sync.Pool
}

func NewPool[T any](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *Pool[T]) Put(v T) {
	p.pool.Put(v)
}
