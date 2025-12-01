package main

import (
	"sync"
)

// Message types for actor communication
type ActorMessage interface {
	ActorMessage()
}

// Actor represents a concurrent entity that processes messages
type Actor struct {
	inbox   chan ActorMessage
	handler func(ActorMessage)
	wg      *sync.WaitGroup
	stop    chan struct{}
}

// NewActor creates a new actor with a message handler
func NewActor(handler func(ActorMessage), bufferSize int) *Actor {
	return &Actor{
		inbox:   make(chan ActorMessage, bufferSize),
		handler: handler,
		wg:      &sync.WaitGroup{},
		stop:    make(chan struct{}),
	}
}

// Start begins processing messages
func (a *Actor) Start() {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case msg := <-a.inbox:
				a.handler(msg)
			case <-a.stop:
				return
			}
		}
	}()
}

// Send sends a message to the actor
func (a *Actor) Send(msg ActorMessage) {
	select {
	case a.inbox <- msg:
	case <-a.stop:
		// Actor stopped, drop message
	}
}

// Stop gracefully stops the actor
func (a *Actor) Stop() {
	close(a.stop)
	a.wg.Wait()
}

// ActorRef is a reference to an actor for sending messages
type ActorRef struct {
	actor *Actor
}

func (ref *ActorRef) Tell(msg ActorMessage) {
	if ref.actor != nil {
		ref.actor.Send(msg)
	}
}
