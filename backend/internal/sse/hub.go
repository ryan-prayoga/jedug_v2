// Package sse provides a minimal in-process SSE hub that manages per-follower
// channels. It is intentionally simple: no external broker, no persistence.
// The hub is safe for concurrent use from multiple goroutines.
package sse

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

// Hub maps each follower to the set of channels currently listening for events.
// A follower can have multiple open connections (e.g. two browser tabs).
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[string]chan string // followerID → connID → ch
	dropped atomic.Uint64
}

// NewHub creates an empty Hub.
func NewHub() *Hub {
	return &Hub{clients: make(map[string]map[string]chan string)}
}

// Default is the process-wide SSE hub used by both the notification dispatcher
// (repository layer) and the SSE HTTP handler.
var Default = NewHub()

// Subscribe registers a new SSE connection for followerID.
// Returns a receive channel and a cleanup function that MUST be deferred.
func (h *Hub) Subscribe(followerID string) (<-chan string, func()) {
	connID := uuid.NewString()
	ch := make(chan string, 16) // buffered so Push never blocks the caller

	h.mu.Lock()
	if h.clients[followerID] == nil {
		h.clients[followerID] = make(map[string]chan string)
	}
	h.clients[followerID][connID] = ch
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if conns, ok := h.clients[followerID]; ok {
			delete(conns, connID)
			if len(conns) == 0 {
				delete(h.clients, followerID)
			}
		}
	}
}

// Push sends msg to every active connection of followerID.
// Non-blocking: if a channel's buffer is full the message is silently dropped.
func (h *Hub) Push(followerID, msg string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.clients[followerID] {
		select {
		case ch <- msg:
		default:
			h.dropped.Add(1)
		}
	}
}

func (h *Hub) ConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	total := 0
	for _, conns := range h.clients {
		total += len(conns)
	}
	return total
}

func (h *Hub) FollowerCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) DroppedCount() uint64 {
	return h.dropped.Load()
}

// FormatEvent renders a single SSE frame with optional event id.
func FormatEvent(eventName string, payload []byte, eventID string) string {
	frame := ""
	if eventID != "" {
		frame += fmt.Sprintf("id: %s\n", eventID)
	}
	frame += fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, payload)
	return frame
}
