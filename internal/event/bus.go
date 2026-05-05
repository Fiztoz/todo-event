package event

import "sync"

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]func(any)
}

func NewBus() *Bus {
	return &Bus{handlers: make(map[string][]func(any))}
}

func (b *Bus) Publish(name string, payload any) {
	b.mu.RLock()
	handlers := b.handlers[name]
	b.mu.RUnlock()
	for _, h := range handlers {
		h(payload)
	}
}

func (b *Bus) Subscribe(name string, handler func(any)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[name] = append(b.handlers[name], handler)
}
