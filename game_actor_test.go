package main

import (
	"testing"
	"time"
)

func TestGameActorCreation(t *testing.T) {
	ga := NewGameActor("test-game")
	ga.Start()
	defer ga.Stop()

	// Query game state
	responseChan := make(chan *GameState, 1)
	ga.Send(GetGameStateMsg{ResponseChan: responseChan})

	state := <-responseChan

	if state.ID != "test-game" {
		t.Errorf("Expected game ID 'test-game', got '%s'", state.ID)
	}

	if state.State != "lobby" {
		t.Errorf("Expected initial state 'lobby', got '%s'", state.State)
	}

	if len(state.Players) != 0 {
		t.Errorf("Expected no players initially, got %d", len(state.Players))
	}
}

func TestGameActorPlayerJoin(t *testing.T) {
	ga := NewGameActor("test-game")
	ga.Start()
	defer ga.Stop()

	// Add a player
	ga.Send(PlayerJoinMsg{
		GameID:     "test-game",
		PlayerID:   "player1",
		PlayerName: "Alice",
		Conn:       nil, // No actual connection for test
	})

	time.Sleep(50 * time.Millisecond)

	// Query state
	responseChan := make(chan *GameState, 1)
	ga.Send(GetGameStateMsg{ResponseChan: responseChan})
	state := <-responseChan

	if len(state.Players) != 1 {
		t.Fatalf("Expected 1 player, got %d", len(state.Players))
	}

	player, exists := state.Players["player1"]
	if !exists {
		t.Fatal("Player 'player1' not found")
	}

	if player.Name != "Alice" {
		t.Errorf("Expected player name 'Alice', got '%s'", player.Name)
	}

	if player.Score != 0 {
		t.Errorf("Expected initial score 0, got %d", player.Score)
	}

	if player.Ready {
		t.Error("Expected player not ready initially")
	}
}

func TestGameActorPlayerLeave(t *testing.T) {
	ga := NewGameActor("test-game")
	ga.Start()
	defer ga.Stop()

	// Add and remove a player
	ga.Send(PlayerJoinMsg{
		GameID:     "test-game",
		PlayerID:   "player1",
		PlayerName: "Alice",
		Conn:       nil,
	})

	time.Sleep(50 * time.Millisecond)

	ga.Send(PlayerLeaveMsg{PlayerID: "player1"})

	time.Sleep(50 * time.Millisecond)

	// Query state
	responseChan := make(chan *GameState, 1)
	ga.Send(GetGameStateMsg{ResponseChan: responseChan})
	state := <-responseChan

	if len(state.Players) != 0 {
		t.Errorf("Expected no players after leave, got %d", len(state.Players))
	}
}

func TestGameActorStateTransitions(t *testing.T) {
	ga := NewGameActor("test-game")
	ga.Start()
	defer ga.Stop()

	// Add a player
	ga.Send(PlayerJoinMsg{
		GameID:     "test-game",
		PlayerID:   "player1",
		PlayerName: "Alice",
		Conn:       nil,
	})

	time.Sleep(50 * time.Millisecond)

	// Check initial state
	responseChan := make(chan *GameState, 1)
	ga.Send(GetGameStateMsg{ResponseChan: responseChan})
	state := <-responseChan

	if state.State != "lobby" {
		t.Errorf("Expected state 'lobby', got '%s'", state.State)
	}

	// Start game (move to instructions)
	ga.Send(NextGameMsg{PlayerID: "player1"})
	time.Sleep(50 * time.Millisecond)

	responseChan = make(chan *GameState, 1)
	ga.Send(GetGameStateMsg{ResponseChan: responseChan})
	state = <-responseChan

	if state.State != "instructions" {
		t.Errorf("Expected state 'instructions', got '%s'", state.State)
	}

	if state.CurrentGame != "madlibs" {
		t.Errorf("Expected current game 'madlibs', got '%s'", state.CurrentGame)
	}

	// Player clicks ready
	ga.Send(NextGameMsg{PlayerID: "player1"})
	time.Sleep(50 * time.Millisecond)

	responseChan = make(chan *GameState, 1)
	ga.Send(GetGameStateMsg{ResponseChan: responseChan})
	state = <-responseChan

	// Single player should be ready and game should start
	if state.State != "playing" {
		t.Errorf("Expected state 'playing', got '%s'", state.State)
	}
}

func TestGameActorMultiplePlayersReady(t *testing.T) {
	ga := NewGameActor("test-game")
	ga.Start()
	defer ga.Stop()

	// Add two players
	ga.Send(PlayerJoinMsg{
		GameID:     "test-game",
		PlayerID:   "player1",
		PlayerName: "Alice",
		Conn:       nil,
	})

	ga.Send(PlayerJoinMsg{
		GameID:     "test-game",
		PlayerID:   "player2",
		PlayerName: "Bob",
		Conn:       nil,
	})

	time.Sleep(50 * time.Millisecond)

	// Start game
	ga.Send(NextGameMsg{PlayerID: "player1"})
	time.Sleep(50 * time.Millisecond)

	// Player 1 ready
	ga.Send(NextGameMsg{PlayerID: "player1"})
	time.Sleep(50 * time.Millisecond)

	responseChan := make(chan *GameState, 1)
	ga.Send(GetGameStateMsg{ResponseChan: responseChan})
	state := <-responseChan

	// Should still be in instructions (not all ready)
	if state.State != "instructions" {
		t.Errorf("Expected state 'instructions', got '%s'", state.State)
	}

	// Player 2 ready
	ga.Send(NextGameMsg{PlayerID: "player2"})
	time.Sleep(50 * time.Millisecond)

	responseChan = make(chan *GameState, 1)
	ga.Send(GetGameStateMsg{ResponseChan: responseChan})
	state = <-responseChan

	// Now should be playing
	if state.State != "playing" {
		t.Errorf("Expected state 'playing', got '%s'", state.State)
	}
}

func TestGameActorPing(t *testing.T) {
	ga := NewGameActor("test-game")
	ga.Start()
	defer ga.Stop()

	// Add a player
	ga.Send(PlayerJoinMsg{
		GameID:     "test-game",
		PlayerID:   "player1",
		PlayerName: "Alice",
		Conn:       nil,
	})

	time.Sleep(50 * time.Millisecond)

	// Send ping (should not crash)
	ga.Send(PingMsg{PlayerID: "player1"})
	time.Sleep(50 * time.Millisecond)

	// Verify player still exists
	responseChan := make(chan *GameState, 1)
	ga.Send(GetGameStateMsg{ResponseChan: responseChan})
	state := <-responseChan

	if len(state.Players) != 1 {
		t.Errorf("Expected 1 player after ping, got %d", len(state.Players))
	}
}
