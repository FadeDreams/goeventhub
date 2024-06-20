
## goeventhub

`goeventhub` is a Go package that provides an event emitter implementation with support for event handlers and filters.

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

	"github.com/fadedreams/goeventhub"
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
	emitter := goeventhub.NewEventEmitter()

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

	// Simulate sending notifications
	emitter.EmitWithContext(context.Background(), "email_notification", &Notification{ID: 1, Message: "New email received", CreatedAt: time.Now()})
	emitter.EmitWithContext(context.Background(), "sms_notification", &Notification{ID: 2, Message: "You have a new SMS", CreatedAt: time.Now()})

	// Handle notifications asynchronously
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	emitter.ListenAndServe(ctx, chEmail, chSMS, handler1, handler2)

	// Allow some time for notifications to be processed
	time.Sleep(2 * time.Second)

	// Unregister handlers
	emitter.Off("email_notification")
	emitter.Off("sms_notification")

	// Close the EventEmitter
	emitter.Close()
}

```

### Documentation
EventEmitter
func NewEventEmitter() *EventEmitter
Creates a new instance of EventEmitter.

func (*EventEmitter) On(event string, filters ...func(interface{}) bool) <-chan interface{}
Registers a listener for the specified event with optional filters.

func (*EventEmitter) EmitWithContext(ctx context.Context, event string, data interface{})
Emits an event with context for cancellation and timeout.

func (*EventEmitter) OnEvent(event string, handler EventHandler)
Registers an event handler for the specified event.

func (*EventEmitter) Off(event string)
Unregisters all listeners and handlers for the specified event.

func (*EventEmitter) Close()
Closes all event channels and clears the listener map.

func (*EventEmitter) ListenAndServe(ctx context.Context, chEmail, chSMS <-chan interface{}, handler1, handler2 EventHandler)
Listens for events on specified channels and dispatches them to handlers.
