package queue

import "context"

// WaitingRoom manages the virtual queue for a domain.
// Uses Redis List as a FIFO queue.
type WaitingRoom struct {
	// redisClient redis.Client
}

// Enqueue adds a visitor to the waiting room.
// Returns their position in the queue.
func (w *WaitingRoom) Enqueue(ctx context.Context, tenantID, domain, visitorID string) (int64, error) {
	// TODO: RPUSH tenant:{tenantID}:queue:{domain} visitorID
	// Return LPOS to get position
	return 0, nil
}

// Dequeue releases the next N visitors from the queue.
func (w *WaitingRoom) Dequeue(ctx context.Context, tenantID, domain string, count int) ([]string, error) {
	// TODO: LPOP tenant:{tenantID}:queue:{domain} count
	return nil, nil
}

// Position returns the current queue position for a visitor.
func (w *WaitingRoom) Position(ctx context.Context, tenantID, domain, visitorID string) (int64, error) {
	return 0, nil
}
