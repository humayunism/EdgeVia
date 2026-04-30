package queue

import (
	"context"
	"sync"
)

// WaitingRoom manages the virtual queue for a domain.
// Uses Redis List as a FIFO queue.
type WaitingRoom struct {
	// redisClient redis.Client
	mu      sync.Mutex
	entries map[string][]string
}

func NewWaitingRoom() *WaitingRoom {
	return &WaitingRoom{
		entries: make(map[string][]string),
	}
}

// Enqueue adds a visitor to the waiting room.
// Returns their position in the queue.
func (w *WaitingRoom) Enqueue(ctx context.Context, tenantID, domain, visitorID string) (int64, error) {
	_ = ctx
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.entries == nil {
		w.entries = make(map[string][]string)
	}

	key := tenantID + ":" + domain
	queue := w.entries[key]
	for i, existing := range queue {
		if existing == visitorID {
			return int64(i + 1), nil
		}
	}

	w.entries[key] = append(queue, visitorID)
	// TODO: RPUSH tenant:{tenantID}:queue:{domain} visitorID
	// Return LPOS to get position
	return int64(len(w.entries[key])), nil
}

// Dequeue releases the next N visitors from the queue.
func (w *WaitingRoom) Dequeue(ctx context.Context, tenantID, domain string, count int) ([]string, error) {
	_ = ctx
	w.mu.Lock()
	defer w.mu.Unlock()

	key := tenantID + ":" + domain
	queue := w.entries[key]
	if len(queue) == 0 || count <= 0 {
		return nil, nil
	}
	if count > len(queue) {
		count = len(queue)
	}

	released := append([]string(nil), queue[:count]...)
	w.entries[key] = append([]string(nil), queue[count:]...)
	// TODO: LPOP tenant:{tenantID}:queue:{domain} count
	return released, nil
}

// Position returns the current queue position for a visitor.
func (w *WaitingRoom) Position(ctx context.Context, tenantID, domain, visitorID string) (int64, error) {
	_ = ctx
	w.mu.Lock()
	defer w.mu.Unlock()

	key := tenantID + ":" + domain
	for i, existing := range w.entries[key] {
		if existing == visitorID {
			return int64(i + 1), nil
		}
	}

	return 0, nil
}

func (w *WaitingRoom) Depth(tenantID, domain string) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return len(w.entries[tenantID+":"+domain])
}
