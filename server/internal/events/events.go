package events

import (
	"fmt"
	"sync"

	console "kori/internal/utils/logger"
)

var log = console.New("EVENTS")

type EventHandler func(interface{})

type EventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

var defaultBus = NewEventBus()

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// On registers a handler for an event
func (bus *EventBus) On(event string, handler EventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.handlers[event] = append(bus.handlers[event], handler)
	log.Info("Registered handler for event: %s", event)
}

// Emit triggers an event with the given data
func (bus *EventBus) Emit(event string, data interface{}) {
	bus.mu.RLock()
	handlers, exists := bus.handlers[event]
	bus.mu.RUnlock()

	if !exists {
		return
	}

	log.Info("Emitting event: %s", event)

	for _, handler := range handlers {
		go func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					err := log.Error("Panic in event handler: %v", fmt.Errorf("panic: %v", r))
					if err != nil {
						return
					}
				}
			}()
			h(data)
		}(handler)
	}
}

// On Global event functions that use the default event bus
func On(event string, handler EventHandler) {
	defaultBus.On(event, handler)
}

func Emit(event string, data interface{}) {
	defaultBus.Emit(event, data)
}
