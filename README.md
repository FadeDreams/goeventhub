
## goeventhub
`goeventhub` is a Go package that provides an event emitter implementation with support for event handlers, filters, asynchronous event processing, contextual event handling, and event prioritization.


### Installation

To use `goeventhub` in your Go module, run:

```bash
go get -u github.com/fadedreams/goeventhub
```

### Usage
```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fadedreams/goeventhub/goeventhub"
	"log"
)

// Notification represents a notification entity
type Notification struct {
	ID        int
	Message   string
	CreatedAt time.Time
}

// NotificationHandler handles notifications
type NotificationHandler struct {
	Name string
}

// Handle implements EventHandler interface for NotificationHandler
func (nh *NotificationHandler) Handle(data interface{}) {
	if notification, ok := data.(*Notification); ok {
		fmt.Printf("[%s] Received notification: ID %d, Message: %s\n", nh.Name, notification.ID, notification.Message)
	}
}

func main() {
	emitter := goeventhub.NewEventEmitter(true)

	// Register notification handlers
	handler1 := &NotificationHandler{Name: "EmailHandler"}
	handler2 := &NotificationHandler{Name: "SMSHandler"}

	emitter.OnEvent("email_notification", handler1)
	emitter.OnEvent("sms_notification", handler2)

	// Registering listeners with filters
	emailFilter := func(data interface{}) bool {
		if notification, ok := data.(*Notification); ok {
			return notification.ID > 0 // Example filter: process only notifications with ID > 0
		}
		return false
	}
	chEmail := emitter.On("email_notification", emailFilter)
	chSMS := emitter.On("sms_notification")

	// Simulate sending notifications with priority
	go func() {
		emitter.EmitWithContext(context.Background(), "email_notification", &Notification{ID: 1, Message: "New email received", CreatedAt: time.Now()}, 2) // Higher priority
		emitter.EmitWithContext(context.Background(), "sms_notification", &Notification{ID: 2, Message: "You have a new SMS", CreatedAt: time.Now()}, 1)   // Lower priority
	}()

	// Handle notifications asynchronously
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// go emitter.ListenAndServe(ctx, chEmail, chSMS, handler1, handler2)
	go func() {
		for {
			select {
			case notification := <-chEmail:
				log.Println("Received email notification")
				handler1.Handle(notification)
			case notification := <-chSMS:
				log.Println("Received SMS notification")
				handler2.Handle(notification)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Allow some time for notifications to be processed
	time.Sleep(2 * time.Second)

}
```

### Documentation
#### NewEventEmitter(logging bool) *EventEmitter

Creates a new instance of EventEmitter with optional logging enabled.

- **Parameters:**
  - `logging bool`: Set to `true` to enable logging; `false` to disable.

#### (*EventEmitter) On(event string, filters ...func(interface{}) bool) <-chan interface{}

Registers a listener for the specified event with optional filters.

- **Parameters:**
  - `event string`: The name of the event to listen for.
  - `filters ...func(interface{}) bool`: Optional filters to apply to incoming data.

- **Returns:**
  - `<-chan interface{}`: A channel to receive filtered events.

#### (*EventEmitter) EmitWithContext(ctx context.Context, event string, data interface{}, priority int)

Emits an event with context for cancellation and timeout.

- **Parameters:**
  - `ctx context.Context`: Context for cancellation and timeout.
  - `event string`: The name of the event to emit.
  - `data interface{}`: Data associated with the event.
  - `priority int`: Priority of the event for processing.

#### (*EventEmitter) OnEvent(event string, handler EventHandler)

Registers an event handler for the specified event.

- **Parameters:**
  - `event string`: The name of the event to handle.
  - `handler EventHandler`: Handler function that implements the Handle(data interface{}) method.

#### (*EventEmitter) Off(event string)

Unregisters all listeners and handlers for the specified event.

- **Parameters:**
  - `event string`: The name of the event to unregister.

#### (*EventEmitter) Close()

Closes all event channels and clears the listener map.

#### (*EventEmitter) ListenAndServe(ctx context.Context, chEmail, chSMS <-chan interface{}, handler1, handler2 EventHandler)

Listens for events on specified channels and dispatches them to handlers.

- **Parameters:**
  - `ctx context.Context`: Context for cancellation.
  - `chEmail <-chan interface{}`: Channel for receiving email notifications.
  - `chSMS <-chan interface{}`: Channel for receiving SMS notifications.
  - `handler1, handler2 EventHandler`: Event handlers for email and SMS notifications respectively.
