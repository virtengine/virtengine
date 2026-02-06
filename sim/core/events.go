package core

import (
	"sync"

	"github.com/virtengine/virtengine/sim/model"
)

// EventQueue stores pending events.
type EventQueue struct {
	mu     sync.Mutex
	events []model.Event
}

// NewEventQueue creates an event queue.
func NewEventQueue() *EventQueue {
	return &EventQueue{events: make([]model.Event, 0)}
}

// Push appends events to the queue.
func (q *EventQueue) Push(events ...model.Event) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.events = append(q.events, events...)
}

// Drain returns all events and clears the queue.
func (q *EventQueue) Drain() []model.Event {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.events) == 0 {
		return nil
	}
	drained := q.events
	q.events = nil
	return drained
}
