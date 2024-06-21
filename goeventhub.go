package goeventhub

import (
	"context"
	"log"
	"sort"
	"sync"
)

// EventHandler defines an interface for handling events.
type EventHandler interface {
	Handle(data interface{})
}

// PriorityEvent represents an event with its priority.
type PriorityEvent struct {
	event    string
	priority int
	data     interface{}
}

// EventEmitter represents an event emitter.
type EventEmitter struct {
	listeners map[string][]chan interface{}
	handlers  map[string][]EventHandler
	events    []PriorityEvent
	mutex     sync.Mutex
	logging   bool
	logger    *log.Logger
}

// NewEventEmitter creates a new instance of EventEmitter.
func NewEventEmitter(logging bool) *EventEmitter {
	emitter := &EventEmitter{
		listeners: make(map[string][]chan interface{}),
		handlers:  make(map[string][]EventHandler),
		events:    make([]PriorityEvent, 0),
		logging:   logging,
		logger:    log.New(log.Writer(), "[EventEmitter] ", log.Flags()), // Initialize logger
	}
	return emitter
}

// On registers a listener for the specified event with optional filters.
func (emitter *EventEmitter) On(event string, filters ...func(interface{}) bool) <-chan interface{} {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()

	ch := make(chan interface{}, 1)
	emitter.listeners[event] = append(emitter.listeners[event], ch)

	// Optionally apply filters to the channel
	go func(ch chan interface{}) {
		for data := range ch {
			if len(filters) == 0 {
				ch <- data
			} else {
				for _, filter := range filters {
					if filter(data) {
						ch <- data
						break
					}
				}
			}
		}
		close(ch)
	}(ch)

	return ch
}

// EmitWithContext emits an event with context for cancellation and timeout.
func (emitter *EventEmitter) EmitWithContext(ctx context.Context, event string, data interface{}, priority int) {
	emitter.mutex.Lock()
	if emitter.logging {
		emitter.logger.Printf("Emitting event %s with priority %d\n", event, priority)
	}
	emitter.events = append(emitter.events, PriorityEvent{event: event, priority: priority, data: data})
	emitter.mutex.Unlock()
}

// ProcessEvents processes events in priority order.
func (emitter *EventEmitter) ProcessEvents(ctx context.Context) {
	for {
		emitter.mutex.Lock()
		if len(emitter.events) == 0 {
			emitter.mutex.Unlock()
			return
		}
		sort.Slice(emitter.events, func(i, j int) bool {
			return emitter.events[i].priority > emitter.events[j].priority
		})
		eventToProcess := emitter.events[0]
		emitter.events = emitter.events[1:]
		emitter.mutex.Unlock()

		if emitter.logging {
			emitter.logger.Printf("Processing event %s with priority %d\n", eventToProcess.event, eventToProcess.priority)
		}

		if listeners, ok := emitter.listeners[eventToProcess.event]; ok {
			for _, ch := range listeners {
				select {
				case <-ctx.Done():
					return
				case ch <- eventToProcess.data:
				}
			}
		}

		emitter.dispatch(eventToProcess.event, eventToProcess.data)
	}
}

// OnEvent registers an event handler for the specified event.
func (emitter *EventEmitter) OnEvent(event string, handler EventHandler) {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()

	emitter.handlers[event] = append(emitter.handlers[event], handler)
}

// Off unregisters all listeners and handlers for the specified event.
func (emitter *EventEmitter) Off(event string) {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()

	delete(emitter.listeners, event)
	delete(emitter.handlers, event)
}

// Close closes all event channels and clears the listener map.
func (emitter *EventEmitter) Close() {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()

	for _, listeners := range emitter.listeners {
		for _, ch := range listeners {
			close(ch)
		}
	}
	emitter.listeners = make(map[string][]chan interface{})
	emitter.handlers = make(map[string][]EventHandler)
}

// Helper function to dispatch events to registered handlers
func (emitter *EventEmitter) dispatch(event string, data interface{}) {
	if eventHandlers, ok := emitter.handlers[event]; ok {
		for _, handler := range eventHandlers {
			handler.Handle(data)
		}
	}
}

// SetLogging enables or disables logging for the EventEmitter.
func (emitter *EventEmitter) SetLogging(enable bool) {
	emitter.logging = enable
}
