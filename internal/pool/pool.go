package pool

import "sync"

// Typed sync.Pool wrapper
type Pool[T any] struct {
	pool sync.Pool
}

func New[T any](f func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return f()
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
