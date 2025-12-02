package main

import (
	"testing"
	"time"
)

func TestMadLibSlotReservation(t *testing.T) {
	madlib := NewMadLib()

	// Player 1 claims first slot
	claimed := madlib.ClaimSlotForPlayer("player1")
	if !claimed {
		t.Fatal("Expected player1 to claim a slot")
	}

	prompt1 := madlib.GetPromptForPlayer("player1")
	if prompt1 == "" {
		t.Fatal("Expected player1 to have a prompt")
	}

	// Player 2 claims second slot (should be different)
	claimed = madlib.ClaimSlotForPlayer("player2")
	if !claimed {
		t.Fatal("Expected player2 to claim a slot")
	}

	prompt2 := madlib.GetPromptForPlayer("player2")
	if prompt2 == "" {
		t.Fatal("Expected player2 to have a prompt")
	}

	// Players should have different slot indices
	idx1 := madlib.playerPrompts["player1"]
	idx2 := madlib.playerPrompts["player2"]
	if idx1 == idx2 {
		t.Errorf("Expected different slots, both got index %d", idx1)
	}

	// Both slots should be claimed
	if madlib.claimedBy[idx1] != "player1" {
		t.Errorf("Expected slot %d claimed by player1, got %s", idx1, madlib.claimedBy[idx1])
	}
	if madlib.claimedBy[idx2] != "player2" {
		t.Errorf("Expected slot %d claimed by player2, got %s", idx2, madlib.claimedBy[idx2])
	}
}

func TestMadLibQueryDoesNotMutate(t *testing.T) {
	madlib := NewMadLib()

	// Query before claiming - should return empty, not claim
	prompt := madlib.GetPromptForPlayer("player1")
	if prompt != "" {
		t.Errorf("Expected empty prompt before claiming, got '%s'", prompt)
	}

	// Verify no slot was claimed
	if _, exists := madlib.playerPrompts["player1"]; exists {
		t.Error("GetPromptForPlayer should not claim slots (query should be pure)")
	}
}

func TestMadLibSubmitAndAdvance(t *testing.T) {
	madlib := NewMadLib()

	// Player claims and fills first slot
	madlib.ClaimSlotForPlayer("player1")
	firstPrompt := madlib.GetPromptForPlayer("player1")
	firstIdx := madlib.playerPrompts["player1"]

	complete := madlib.AddWordForPlayer("player1", "big")
	if complete {
		t.Error("Expected game not complete after one word")
	}

	// Verify word was filled
	if madlib.Words[firstIdx] != "big" {
		t.Errorf("Expected word 'big' at index %d, got '%s'", firstIdx, madlib.Words[firstIdx])
	}

	// Player should have advanced to next slot
	newIdx := madlib.playerPrompts["player1"]
	if newIdx == firstIdx {
		t.Error("Expected player to advance to next slot after submission")
	}

	// New slot should be claimed by player1
	if madlib.claimedBy[newIdx] != "player1" {
		t.Errorf("Expected new slot claimed by player1, got '%s'", madlib.claimedBy[newIdx])
	}

	newPrompt := madlib.GetPromptForPlayer("player1")
	if newPrompt == "" {
		t.Error("Expected player to have new prompt after advancing")
	}
	if newPrompt == firstPrompt {
		t.Error("Expected different prompt after advancing")
	}
}

func TestMadLibAllSlotsClaimedNoMore(t *testing.T) {
	madlib := NewMadLib()
	numSlots := len(madlib.Prompts)

	// Claim all slots with different players
	for i := 0; i < numSlots; i++ {
		playerID := string(rune('A' + i))
		claimed := madlib.ClaimSlotForPlayer(playerID)
		if !claimed {
			t.Fatalf("Expected player %s to claim slot %d", playerID, i)
		}
	}

	// Try to claim one more - should fail
	claimed := madlib.ClaimSlotForPlayer("overflow")
	if claimed {
		t.Error("Expected no slots available for overflow player")
	}

	// Overflow player should get empty prompt
	prompt := madlib.GetPromptForPlayer("overflow")
	if prompt != "" {
		t.Errorf("Expected empty prompt for overflow player, got '%s'", prompt)
	}

	// Verify HasAvailableSlots returns false
	if madlib.HasAvailableSlots() {
		t.Error("Expected no available slots")
	}
}

func TestMadLibCompletionAllWordsFilled(t *testing.T) {
	madlib := NewMadLib()
	numSlots := len(madlib.Prompts)

	// Single player fills all slots
	madlib.ClaimSlotForPlayer("player1")

	for i := 0; i < numSlots; i++ {
		word := "word" + string(rune('A'+i))
		complete := madlib.AddWordForPlayer("player1", word)

		if i < numSlots-1 {
			if complete {
				t.Errorf("Expected not complete at slot %d/%d", i+1, numSlots)
			}
		} else {
			if !complete {
				t.Errorf("Expected complete at final slot %d/%d", i+1, numSlots)
			}
		}
	}

	// Verify IsComplete
	if !madlib.IsComplete() {
		t.Error("Expected game complete after all words filled")
	}
}

func TestMadLibActorRequestPrompt(t *testing.T) {
	ga := NewGameActor("madlibs-test")
	ga.Start()
	defer ga.Stop()

	// Add player and start Mad Libs
	ga.Send(PlayerJoinMsg{
		GameID:     "madlibs-test",
		PlayerID:   "player1",
		PlayerName: "Alice",
		Conn:       nil,
	})
	time.Sleep(50 * time.Millisecond)

	// Start game and navigate to playing state
	ga.Send(NextGameMsg{PlayerID: "player1"})
	time.Sleep(50 * time.Millisecond)
	ga.Send(NextGameMsg{PlayerID: "player1"})
	time.Sleep(50 * time.Millisecond)

	// Verify we're in playing state
	responseChan := make(chan *GameState, 1)
	ga.Send(GetGameStateMsg{ResponseChan: responseChan})
	state := <-responseChan

	if state.State != "playing" {
		t.Fatalf("Expected 'playing' state, got '%s'", state.State)
	}

	// Force Mad Libs as the game (since it's random)
	ga.mu.Lock()
	ga.currentGame = "madlibs"
	ga.game = NewMadLib()
	ga.mu.Unlock()

	// Send RequestPromptMsg
	ga.Send(RequestPromptMsg{PlayerID: "player1"})
	time.Sleep(50 * time.Millisecond)

	// Verify slot was claimed
	ga.mu.RLock()
	madlib, ok := ga.game.(*MadLib)
	ga.mu.RUnlock()

	if !ok {
		t.Fatal("Expected game to be MadLib")
	}

	prompt := madlib.GetPromptForPlayer("player1")
	if prompt == "" {
		t.Error("Expected player to have a prompt after RequestPromptMsg")
	}

	idx, exists := madlib.playerPrompts["player1"]
	if !exists {
		t.Fatal("Expected player to have claimed a slot")
	}

	if madlib.claimedBy[idx] != "player1" {
		t.Errorf("Expected slot %d claimed by player1, got '%s'", idx, madlib.claimedBy[idx])
	}
}

func TestMadLibActorMultiplePlayersNoDuplicateSlots(t *testing.T) {
	ga := NewGameActor("madlibs-multi")
	ga.Start()
	defer ga.Stop()

	// Add two players
	ga.Send(PlayerJoinMsg{
		GameID:     "madlibs-multi",
		PlayerID:   "player1",
		PlayerName: "Alice",
		Conn:       nil,
	})
	ga.Send(PlayerJoinMsg{
		GameID:     "madlibs-multi",
		PlayerID:   "player2",
		PlayerName: "Bob",
		Conn:       nil,
	})
	time.Sleep(50 * time.Millisecond)

	// Force Mad Libs and playing state
	ga.mu.Lock()
	ga.currentGame = "madlibs"
	ga.state = "playing"
	ga.game = NewMadLib()
	ga.mu.Unlock()

	// Both players request prompts
	ga.Send(RequestPromptMsg{PlayerID: "player1"})
	ga.Send(RequestPromptMsg{PlayerID: "player2"})
	time.Sleep(100 * time.Millisecond)

	// Verify they got different slots
	ga.mu.RLock()
	madlib, _ := ga.game.(*MadLib)
	idx1 := madlib.playerPrompts["player1"]
	idx2 := madlib.playerPrompts["player2"]
	ga.mu.RUnlock()

	if idx1 == idx2 {
		t.Errorf("Expected different slots, both got %d", idx1)
	}

	// Verify both are claimed
	if madlib.claimedBy[idx1] != "player1" {
		t.Errorf("Expected slot %d claimed by player1", idx1)
	}
	if madlib.claimedBy[idx2] != "player2" {
		t.Errorf("Expected slot %d claimed by player2", idx2)
	}
}

func TestMadLibActorSubmitWord(t *testing.T) {
	ga := NewGameActor("madlibs-submit")
	ga.Start()
	defer ga.Stop()

	// Setup player and Mad Libs game
	ga.Send(PlayerJoinMsg{
		GameID:     "madlibs-submit",
		PlayerID:   "player1",
		PlayerName: "Alice",
		Conn:       nil,
	})
	time.Sleep(50 * time.Millisecond)

	ga.mu.Lock()
	ga.currentGame = "madlibs"
	ga.state = "playing"
	ga.game = NewMadLib()
	ga.mu.Unlock()

	// Request prompt and submit word
	ga.Send(RequestPromptMsg{PlayerID: "player1"})
	time.Sleep(50 * time.Millisecond)

	ga.mu.RLock()
	madlib := ga.game.(*MadLib)
	firstIdx := madlib.playerPrompts["player1"]
	ga.mu.RUnlock()

	ga.Send(SubmitWordMsg{
		PlayerID: "player1",
		Word:     "fantastic",
	})
	time.Sleep(50 * time.Millisecond)

	// Verify word was filled
	ga.mu.RLock()
	madlib = ga.game.(*MadLib)
	filledWord := madlib.Words[firstIdx]
	newIdx := madlib.playerPrompts["player1"]
	ga.mu.RUnlock()

	if filledWord != "fantastic" {
		t.Errorf("Expected word 'fantastic', got '%s'", filledWord)
	}

	if newIdx == firstIdx {
		t.Error("Expected player to advance to next slot")
	}
}
