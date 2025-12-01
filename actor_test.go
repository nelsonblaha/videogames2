package main

import (
	"sync"
	"testing"
	"time"
)

// Test message for actor tests
type TestMsg struct {
	Value string
}

func (m TestMsg) ActorMessage() {}

func TestActorCreation(t *testing.T) {
	received := make([]string, 0)
	var mu sync.Mutex

	handler := func(msg ActorMessage) {
		if testMsg, ok := msg.(TestMsg); ok {
			mu.Lock()
			received = append(received, testMsg.Value)
			mu.Unlock()
		}
	}

	actor := NewActor(handler, 10)
	actor.Start()
	defer actor.Stop()

	// Send messages
	actor.Send(TestMsg{Value: "hello"})
	actor.Send(TestMsg{Value: "world"})

	// Give actor time to process
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(received) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(received))
	}

	if received[0] != "hello" || received[1] != "world" {
		t.Errorf("Messages not received in order: %v", received)
	}
}

func TestActorStop(t *testing.T) {
	processed := 0
	var mu sync.Mutex

	handler := func(msg ActorMessage) {
		mu.Lock()
		processed++
		mu.Unlock()
		time.Sleep(10 * time.Millisecond) // Simulate work
	}

	actor := NewActor(handler, 10)
	actor.Start()

	// Send messages
	actor.Send(TestMsg{Value: "msg1"})
	actor.Send(TestMsg{Value: "msg2"})

	// Give actor time to start processing
	time.Sleep(20 * time.Millisecond)

	// Stop
	actor.Stop()

	// Try to send after stop (should not crash)
	actor.Send(TestMsg{Value: "msg3"})

	mu.Lock()
	defer mu.Unlock()

	// Should have processed at least some messages before stop
	if processed == 0 {
		t.Error("No messages processed before stop")
	}

	// Should not process message sent after stop
	if processed > 2 {
		t.Errorf("Expected at most 2 messages processed, got %d", processed)
	}
}

func TestActorConcurrency(t *testing.T) {
	counter := 0
	var mu sync.Mutex

	handler := func(msg ActorMessage) {
		mu.Lock()
		counter++
		mu.Unlock()
	}

	actor := NewActor(handler, 100)
	actor.Start()
	defer actor.Stop()

	// Send messages from multiple goroutines
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				actor.Send(TestMsg{Value: "msg"})
			}
		}()
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if counter != 100 {
		t.Errorf("Expected 100 messages processed, got %d", counter)
	}
}
