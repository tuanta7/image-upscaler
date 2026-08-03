package sse

import "sync"

type Hub struct {
	mu   sync.Mutex
	subs map[string][]chan string
}

func NewHub() *Hub {
	return &Hub{
		subs: make(map[string][]chan string),
	}
}

func (h *Hub) Subscribe(taskID string) (<-chan string, func()) {
	ch := make(chan string, 1)

	h.mu.Lock()
	h.subs[taskID] = append(h.subs[taskID], ch)
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		subs := h.subs[taskID]
		for i, c := range subs {
			if c == ch {
				h.subs[taskID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		if len(h.subs[taskID]) == 0 {
			delete(h.subs, taskID)
		}
	}
	return ch, unsubscribe
}

func (h *Hub) Publish(taskID, status string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, ch := range h.subs[taskID] {
		select {
		case ch <- status:
		default:
			// ponytail: buffer of 1, a slow client drops updates.
			// The client re-reads status from the DB on reconnect,
			// so widen the buffer only if that stops being enough.
		}
	}
}
