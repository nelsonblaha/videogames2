package main

import (
	"log"
	"sync"
)

// GameCoordinator manages all game actors
type GameCoordinator struct {
	games map[string]*GameActor
	mu    sync.RWMutex
}

// NewGameCoordinator creates a new game coordinator
func NewGameCoordinator() *GameCoordinator {
	return &GameCoordinator{
		games: make(map[string]*GameActor),
	}
}

// GetOrCreateGame gets an existing game or creates a new one
func (gc *GameCoordinator) GetOrCreateGame(gameID string) *GameActor {
	gc.mu.RLock()
	game, exists := gc.games[gameID]
	gc.mu.RUnlock()

	if exists {
		return game
	}

	gc.mu.Lock()
	defer gc.mu.Unlock()

	// Double-check after acquiring write lock
	if game, exists := gc.games[gameID]; exists {
		return game
	}

	// Create new game actor
	game = NewGameActor(gameID)
	game.Start()
	gc.games[gameID] = game
	log.Printf("Created new game: %s", gameID)

	return game
}

// GetGame gets an existing game without creating it
func (gc *GameCoordinator) GetGame(gameID string) *GameActor {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	return gc.games[gameID]
}

// RemoveEmptyGames removes games with no players
func (gc *GameCoordinator) RemoveEmptyGames() {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	for id, game := range gc.games {
		responseChan := make(chan *GameState, 1)
		game.Send(GetGameStateMsg{ResponseChan: responseChan})
		state := <-responseChan

		if len(state.Players) == 0 {
			game.Stop()
			delete(gc.games, id)
			log.Printf("Removed empty game: %s", id)
		}
	}
}

// Stop stops all game actors
func (gc *GameCoordinator) Stop() {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	for id, game := range gc.games {
		game.Stop()
		log.Printf("Stopped game: %s", id)
	}
	gc.games = make(map[string]*GameActor)
}
