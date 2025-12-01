package main

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// GameActor manages a single game session using the actor model
type GameActor struct {
	id          string
	state       string // "lobby", "instructions", "playing", "finished"
	currentGame string
	players     map[string]*Player
	madlib      *MadLib
	mu          sync.RWMutex
	actor       *Actor
}

// Player represents a player in the game
type Player struct {
	ID    string
	Name  string
	Score int
	Ready bool
	Conn  *websocket.Conn
	mu    sync.Mutex
}

// NewGameActor creates a new game actor
func NewGameActor(gameID string) *GameActor {
	ga := &GameActor{
		id:      gameID,
		state:   "lobby",
		players: make(map[string]*Player),
	}

	// Create the actor with message handler
	ga.actor = NewActor(ga.handleMessage, 100)
	return ga
}

// Start starts the game actor
func (ga *GameActor) Start() {
	ga.actor.Start()
}

// Stop stops the game actor
func (ga *GameActor) Stop() {
	ga.actor.Stop()
}

// Send sends a message to the game actor
func (ga *GameActor) Send(msg ActorMessage) {
	ga.actor.Send(msg)
}

// handleMessage processes incoming messages
func (ga *GameActor) handleMessage(msg ActorMessage) {
	switch m := msg.(type) {
	case PlayerJoinMsg:
		ga.handlePlayerJoin(m)
	case PlayerLeaveMsg:
		ga.handlePlayerLeave(m)
	case NextGameMsg:
		ga.handleNextGame(m)
	case PingMsg:
		ga.handlePing(m)
	case SubmitWordMsg:
		ga.handleSubmitWord(m)
	case BroadcastStateMsg:
		ga.broadcastState()
	case GetGameStateMsg:
		ga.handleGetGameState(m)
	}
}

func (ga *GameActor) handlePlayerJoin(msg PlayerJoinMsg) {
	ga.mu.Lock()
	defer ga.mu.Unlock()

	player := &Player{
		ID:    msg.PlayerID,
		Name:  msg.PlayerName,
		Score: 0,
		Ready: false,
		Conn:  msg.Conn,
	}
	ga.players[msg.PlayerID] = player

	log.Printf("Player %s (%s) joined game %s", msg.PlayerName, msg.PlayerID, ga.id)
	ga.broadcastState()
}

func (ga *GameActor) handlePlayerLeave(msg PlayerLeaveMsg) {
	ga.mu.Lock()
	defer ga.mu.Unlock()

	if player, exists := ga.players[msg.PlayerID]; exists {
		player.mu.Lock()
		if player.Conn != nil {
			player.Conn.Close()
		}
		player.mu.Unlock()
		delete(ga.players, msg.PlayerID)
		log.Printf("Player %s left game %s", msg.PlayerID, ga.id)
		ga.broadcastState()
	}
}

func (ga *GameActor) handleNextGame(msg NextGameMsg) {
	ga.mu.Lock()
	defer ga.mu.Unlock()

	switch ga.state {
	case "lobby", "":
		// Move to instructions
		ga.state = "instructions"
		ga.currentGame = "madlibs"
		ga.broadcastState()

	case "instructions":
		// Mark player as ready
		if player, exists := ga.players[msg.PlayerID]; exists {
			player.Ready = true
		}

		// Check if all players are ready
		allReady := true
		for _, p := range ga.players {
			if !p.Ready {
				allReady = false
				break
			}
		}

		if allReady && len(ga.players) > 0 {
			ga.state = "playing"
			ga.madlib = NewMadLib()
			// Reset ready status
			for _, p := range ga.players {
				p.Ready = false
			}
		}
		ga.broadcastState()

	case "playing":
		// This shouldn't be called during playing - words are submitted via SubmitWordMsg
		ga.broadcastState()

	case "finished":
		// Go back to lobby
		ga.state = "lobby"
		ga.currentGame = ""
		ga.madlib = nil
		for _, p := range ga.players {
			p.Ready = false
		}
		ga.broadcastState()
	}
}

func (ga *GameActor) handleSubmitWord(msg SubmitWordMsg) {
	ga.mu.Lock()
	defer ga.mu.Unlock()

	if ga.state != "playing" || ga.madlib == nil {
		return
	}

	// Add word to mad lib
	isComplete := ga.madlib.AddWord(msg.Word)

	// Award point to player
	if player, exists := ga.players[msg.PlayerID]; exists {
		player.Score++
	}

	if isComplete {
		ga.state = "finished"
	}

	ga.broadcastState()
}

func (ga *GameActor) handlePing(msg PingMsg) {
	ga.mu.RLock()
	defer ga.mu.RUnlock()

	if player, exists := ga.players[msg.PlayerID]; exists {
		player.mu.Lock()
		defer player.mu.Unlock()

		pong := map[string]interface{}{
			"action": "pong",
		}
		if player.Conn != nil {
			player.Conn.WriteJSON(pong)
		}
	}
}

func (ga *GameActor) handleGetGameState(msg GetGameStateMsg) {
	ga.mu.RLock()
	defer ga.mu.RUnlock()

	state := &GameState{
		ID:          ga.id,
		State:       ga.state,
		CurrentGame: ga.currentGame,
		Players:     make(map[string]*PlayerInfo),
	}

	for id, p := range ga.players {
		state.Players[id] = &PlayerInfo{
			ID:    p.ID,
			Name:  p.Name,
			Score: p.Score,
			Ready: p.Ready,
		}
	}

	msg.ResponseChan <- state
}

func (ga *GameActor) broadcastState() {
	// Build state message
	playersList := make([]map[string]interface{}, 0, len(ga.players))
	for _, p := range ga.players {
		playersList = append(playersList, map[string]interface{}{
			"id":    p.ID,
			"name":  p.Name,
			"score": p.Score,
			"ready": p.Ready,
		})
	}

	var gameTitle, gameInstructions, roundInstructions string
	numPlayers := len(ga.players)

	switch ga.state {
	case "lobby", "":
		if numPlayers == 0 {
			gameTitle = "Waiting for players to join"
			gameInstructions = "Wait for players to join"
			roundInstructions = "Share this URL with friends to play together!"
		} else if numPlayers == 1 {
			gameTitle = "Ready to play!"
			gameInstructions = "Click 'Next' to start a game"
			roundInstructions = "Most games work solo! Or share the URL to play with friends."
		} else {
			gameTitle = "Waiting for more players"
			gameInstructions = "Wait for more players or click 'Next' to start"
			roundInstructions = "Ready to play?"
		}

	case "instructions":
		gameTitle = "Mad Libs"
		gameInstructions = "Fill in the blanks with funny words!"
		if numPlayers == 1 {
			roundInstructions = "Click 'Next' when ready to play"
		} else {
			roundInstructions = "Everyone click 'Next' when ready"
		}

	case "playing":
		if ga.madlib != nil {
			currentPrompt := ga.madlib.CurrentPrompt()
			gameTitle = "Mad Libs"
			gameInstructions = "Enter a word for:"
			roundInstructions = currentPrompt
		} else {
			gameTitle = "Mad Libs"
			gameInstructions = "Loading..."
			roundInstructions = ""
		}

	case "finished":
		gameTitle = "Story Complete!"
		gameInstructions = "Here's your Mad Lib:"
		if ga.madlib != nil {
			roundInstructions = ga.madlib.GetStory()
		} else {
			roundInstructions = ""
		}
	}

	stateData := map[string]interface{}{
		"game_title":         gameTitle,
		"game_instructions":  gameInstructions,
		"round_instructions": roundInstructions,
		"players":            playersList,
		"game_state":         ga.state,
	}

	// Add mad lib data if in playing state
	if ga.state == "playing" && ga.madlib != nil {
		stateData["current_prompt"] = ga.madlib.CurrentPrompt()
		stateData["words_collected"] = len(ga.madlib.Words) - countEmpty(ga.madlib.Words)
		stateData["total_words"] = len(ga.madlib.Words)
	}

	// Add completed story if finished
	if ga.state == "finished" && ga.madlib != nil {
		stateData["story"] = ga.madlib.GetStory()
	}

	stateMsg := map[string]interface{}{
		"state": stateData,
	}

	jsonData, err := json.Marshal(stateMsg)
	if err != nil {
		log.Printf("Error marshaling state: %v", err)
		return
	}

	// Send to all players
	for _, player := range ga.players {
		player.mu.Lock()
		if player.Conn != nil {
			err := player.Conn.WriteMessage(websocket.TextMessage, jsonData)
			if err != nil {
				log.Printf("Error sending to player %s: %v", player.ID, err)
			}
		}
		player.mu.Unlock()
	}
}

func countEmpty(words []string) int {
	count := 0
	for _, w := range words {
		if w == "" {
			count++
		}
	}
	return count
}
