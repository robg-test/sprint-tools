package internal

import "sync"

type SSEBroadcaster struct {
	mu        sync.Mutex
	listeners map[string][]chan struct{}
}

func NewSSEBroadcaster() *SSEBroadcaster {
	return &SSEBroadcaster{listeners: make(map[string][]chan struct{})}
}

var SessionEvents = NewSSEBroadcaster()

func (b *SSEBroadcaster) Subscribe(id string) chan struct{} {
	ch := make(chan struct{}, 1)
	b.mu.Lock()
	b.listeners[id] = append(b.listeners[id], ch)
	b.mu.Unlock()
	return ch
}

func (b *SSEBroadcaster) Unsubscribe(id string, ch chan struct{}) {
	b.mu.Lock()
	chs := b.listeners[id]
	for i, c := range chs {
		if c == ch {
			b.listeners[id] = append(chs[:i], chs[i+1:]...)
			break
		}
	}
	b.mu.Unlock()
}

func (b *SSEBroadcaster) Notify(id string) {
	b.mu.Lock()
	for _, ch := range b.listeners[id] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	b.mu.Unlock()
}
