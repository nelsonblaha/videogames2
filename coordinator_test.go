package main

import (
	"testing"
	"time"
)

func TestCoordinatorCreation(t *testing.T) {
	gc := NewGameCoordinator()

	if gc == nil {
		t.Fatal("NewGameCoordinator returned nil")
	}

	if gc.games == nil {
		t.Fatal("Coordinator games map not initialized")
	}
}

func TestCoordinatorGetOrCreateGame(t *testing.T) {
	gc := NewGameCoordinator()
	defer gc.Stop()

	// Create a game
	game1 := gc.GetOrCreateGame("game1")
	if game1 == nil {
		t.Fatal("GetOrCreateGame returned nil")
	}

	if game1.id != "game1" {
		t.Errorf("Expected game ID 'game1', got '%s'", game1.id)
	}

	// Get same game again
	game2 := gc.GetOrCreateGame("game1")
	if game1 != game2 {
		t.Error("GetOrCreateGame returned different instance for same ID")
	}

	// Create different game
	game3 := gc.GetOrCreateGame("game2")
	if game1 == game3 {
		t.Error("GetOrCreateGame returned same instance for different ID")
	}
}

func TestCoordinatorGetGame(t *testing.T) {
	gc := NewGameCoordinator()
	defer gc.Stop()

	// Get non-existent game
	game := gc.GetGame("nonexistent")
	if game != nil {
		t.Error("GetGame returned non-nil for non-existent game")
	}

	// Create game
	gc.GetOrCreateGame("game1")

	// Get existing game
	game = gc.GetGame("game1")
	if game == nil {
		t.Error("GetGame returned nil for existing game")
	}
}

func TestCoordinatorRemoveEmptyGames(t *testing.T) {
	gc := NewGameCoordinator()
	defer gc.Stop()

	// Create games
	game1 := gc.GetOrCreateGame("game1")
	game2 := gc.GetOrCreateGame("game2")

	// Add player to game1
	game1.Send(PlayerJoinMsg{
		GameID:     "game1",
		PlayerID:   "player1",
		PlayerName: "Alice",
		Conn:       nil,
	})

	time.Sleep(50 * time.Millisecond)

	// Remove empty games
	gc.RemoveEmptyGames()

	time.Sleep(50 * time.Millisecond)

	// game1 should still exist (has player)
	if gc.GetGame("game1") == nil {
		t.Error("Game with players was removed")
	}

	// game2 should be removed (empty)
	if gc.GetGame("game2") != nil {
		t.Error("Empty game was not removed")
	}
}

func TestCoordinatorConcurrency(t *testing.T) {
	gc := NewGameCoordinator()
	defer gc.Stop()

	// Create games concurrently
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				gc.GetOrCreateGame("concurrent-game")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should only have one game
	game := gc.GetGame("concurrent-game")
	if game == nil {
		t.Fatal("Concurrent game not created")
	}

	// Verify only one instance exists
	gc.mu.RLock()
	count := len(gc.games)
	gc.mu.RUnlock()

	if count != 1 {
		t.Errorf("Expected 1 game, got %d", count)
	}
}

func TestCoordinatorStop(t *testing.T) {
	gc := NewGameCoordinator()

	// Create games
	gc.GetOrCreateGame("game1")
	gc.GetOrCreateGame("game2")
	gc.GetOrCreateGame("game3")

	gc.mu.RLock()
	count := len(gc.games)
	gc.mu.RUnlock()

	if count != 3 {
		t.Fatalf("Expected 3 games, got %d", count)
	}

	// Stop all games
	gc.Stop()

	gc.mu.RLock()
	count = len(gc.games)
	gc.mu.RUnlock()

	if count != 0 {
		t.Errorf("Expected 0 games after stop, got %d", count)
	}
}
