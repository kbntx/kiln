package stream

import (
	"sync"
)

// Event represents a single event in a run's stream.
type Event struct {
	Type string // "log", "status", "projects"
	Data string
}

// runChannel holds a replay buffer and a set of subscribed clients for a single run.
type runChannel struct {
	mu      sync.Mutex
	buffer  []Event
	clients []chan Event
}

// Broker manages per-run event streams with replay buffer.
type Broker struct {
	mu       sync.RWMutex
	channels map[string]*runChannel
}

// NewBroker creates a new Broker.
func NewBroker() *Broker {
	return &Broker{
		channels: make(map[string]*runChannel),
	}
}

// getOrCreate returns the runChannel for the given runID, creating it if needed.
func (b *Broker) getOrCreate(runID string) *runChannel {
	b.mu.RLock()
	ch, ok := b.channels[runID]
	b.mu.RUnlock()
	if ok {
		return ch
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	// Double-check after acquiring write lock.
	if ch, ok = b.channels[runID]; ok {
		return ch
	}
	ch = &runChannel{}
	b.channels[runID] = ch
	return ch
}

// Publish sends an event to all subscribers of the given run and appends it to the replay buffer.
func (b *Broker) Publish(runID string, event Event) {
	rc := b.getOrCreate(runID)
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.buffer = append(rc.buffer, event)
	for _, c := range rc.clients {
		// Non-blocking send; drop if client is slow.
		select {
		case c <- event:
		default:
		}
	}
}

// Subscribe returns a channel that receives events for the given run and an
// unsubscribe function. All buffered events are replayed immediately into the channel.
func (b *Broker) Subscribe(runID string) (<-chan Event, func()) {
	rc := b.getOrCreate(runID)
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Use a buffered channel large enough to hold the replay plus some headroom.
	ch := make(chan Event, len(rc.buffer)+64)

	// Replay existing buffer.
	for _, e := range rc.buffer {
		ch <- e
	}

	rc.clients = append(rc.clients, ch)

	unsubscribe := func() {
		rc.mu.Lock()
		defer rc.mu.Unlock()
		for i, c := range rc.clients {
			if c == ch {
				rc.clients = append(rc.clients[:i], rc.clients[i+1:]...)
				break
			}
		}
	}

	return ch, unsubscribe
}

// Cleanup removes the channel and replay buffer for the given run.
func (b *Broker) Cleanup(runID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if rc, ok := b.channels[runID]; ok {
		rc.mu.Lock()
		for _, c := range rc.clients {
			close(c)
		}
		rc.clients = nil
		rc.buffer = nil
		rc.mu.Unlock()
		delete(b.channels, runID)
	}
}
