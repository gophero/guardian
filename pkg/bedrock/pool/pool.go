package pool

import "sync"

// Pool is a generic [sync.Pool] wrapper.
type Pool[T any] struct {
	pool sync.Pool
}

// New constructs new [Pool].
func New[T any](f func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return f()
			},
		},
	}
}

// Get selects an arbitrary item from the pool, removes it from the pool, and returns it to the caller.
//
// If pool is empty Get create a new instance.
func (p *Pool[T]) Get() T {
	//nolint:errcheck
	return p.pool.Get().(T)
}

// Put adds x to the pool.
func (p *Pool[T]) Put(x T) {
	p.pool.Put(x)
}
