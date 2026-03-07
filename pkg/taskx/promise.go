package taskx

import "sync"

type Promise[T any] struct {
	fn   func() (T, error)
	val  T
	err  error
	done chan struct{}
	once sync.Once
}

func NewPromise[T any](fn func() (T, error)) *Promise[T] {
	return &Promise[T]{
		fn:   fn,
		done: make(chan struct{}),
	}
}

func (p *Promise[T]) OnExecute() {
	p.val, p.err = p.fn()
}

func (p *Promise[T]) OnDone(err error) {
	if err != nil {
		p.err = err
	}
	p.once.Do(func() {
		close(p.done)
	})
}

func (p *Promise[T]) Get() (T, error) {
	<-p.done
	return p.val, p.err
}

func (p *Promise[T]) Done() <-chan struct{} {
	return p.done
}
