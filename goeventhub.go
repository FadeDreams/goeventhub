package goeventhub

import (
	"context"
	"sync"
)

// EventHandler defines an interface for handling events.
type EventHandler interface {
	Handle(data interface{})
}

// EventEmitter represents an event emitter.
type EventEmitter struct {
	listeners map[string][]chan interface{}
	handlers  map[string][]EventHandler // Map of event handlers
	mutex     sync.Mutex
}

// NewEventEmitter creates a new instance of EventEmitter.
func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		listeners: make(map[string][]chan interface{}),
		handlers:  make(map[string][]EventHandler),
	}
}

// On registers a listener for the specified event with optional filters.
func (emitter *EventEmitter) On(event string, filters ...func(interface{}) bool) <-chan interface{} {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()

	ch := make(chan interface{}, 1)
	emitter.listeners[event] = append(emitter.listeners[event], ch)

	// Optionally apply filters to the channel
	go func(ch chan interface{}) {
		for {
			select {
			case data := <-ch:
				if len(filters) == 0 {
					// No filters, pass data directly
					ch <- data
				} else {
					// Apply filters
					for _, filter := range filters {
						if filter(data) {
							ch <- data
							break
						}
					}
				}
			}
		}
	}(ch)

	return ch
}

// EmitWithContext emits an event with context for cancellation and timeout.
func (emitter *EventEmitter) EmitWithContext(ctx context.Context, event string, data interface{}) {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()

	if listeners, ok := emitter.listeners[event]; ok {
		for _, ch := range listeners {
			go func(ch chan interface{}, data interface{}) {
				select {
				case <-ctx.Done():
					// Context cancelled, do not send data
				default:
					ch <- data
				}
			}(ch, data)
		}
	}

	// Dispatch to registered handlers
	emitter.dispatch(event, data)
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

// ListenAndServe listens for events on specified channels and dispatches them to handlers.
func (emitter *EventEmitter) ListenAndServe(ctx context.Context, chEmail, chSMS <-chan interface{}, handler1, handler2 EventHandler) {
	go func() {
		for {
			select {
			case notification := <-chEmail:
				if h, ok := notification.(EventHandler); ok {
					h.Handle(notification)
				}
				handler1.Handle(notification)
			case notification := <-chSMS:
				if h, ok := notification.(EventHandler); ok {
					h.Handle(notification)
				}
				handler2.Handle(notification)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Helper function to dispatch events to registered handlers
func (emitter *EventEmitter) dispatch(event string, data interface{}) {
	if eventHandlers, ok := emitter.handlers[event]; ok {
		for _, handler := range eventHandlers {
			handler.Handle(data)
		}
	}
}
