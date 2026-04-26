package pkgio

import (
	"sync"
	"sync/atomic"

	"github.com/samber/mo"
)

type IO[A any] struct {
	once   sync.Once
	f      func() mo.Result[A]
	cached mo.Result[A]
	hasRun atomic.Bool
}

func Lazy[A any](f func() mo.Result[A]) IO[A] {
	return IO[A]{f: f}
}

// Run executes f at most once and returns the cached result. Goroutine-safe.
func (io *IO[A]) Run() mo.Result[A] {
	io.once.Do(func() {
		io.cached = io.f()
		io.hasRun.Store(true)
	})
	return io.cached
}

// IsInitialized reports whether Run has been called at least once.
func (io *IO[A]) IsInitialized() bool {
	return io.hasRun.Load()
}
