package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

type Player struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Score    int             `json:"score"`
	Ready    bool            `json:"ready"`
	Conn     *websocket.Conn `json:"-"`
	mu       sync.Mutex
}

type Game struct {
	ID            string
	Players       map[string]*Player
	State         string // "lobby", "instructions", "playing"
	CurrentGame   string
	mu            sync.RWMutex
}

type Message struct {
	Action string                 `json:"action"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type StateMessage struct {
	State struct {
		GameTitle        string              `json:"game_title"`
		GameInstructions string              `json:"game_instructions"`
		RoundInstructions string             `json:"round_instructions"`
		Players          []map[string]interface{} `json:"players"`
		Message          string              `json:"message,omitempty"`
	} `json:"state"`
}

var games = make(map[string]*Game)
var gamesMu sync.RWMutex

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/api/user", handleUserAPI)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleUserAPI(w http.ResponseWriter, r *http.Request) {
	// Check for X-Remote-User header (from homepage forward auth)
	username := r.Header.Get("X-Remote-User")

	response := map[string]interface{}{
		"authenticated": username != "",
		"name":         username,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	var player *Player
	var game *Game

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			if player != nil && game != nil {
				removePlayer(game, player)
			}
			break
		}

		switch msg.Action {
		case "join":
			groupName := msg.Data["group"].(string)
			playerName := msg.Data["name"].(string)

			game = getOrCreateGame(groupName)
			player = &Player{
				ID:    generateID(),
				Name:  playerName,
				Score: 0,
				Ready: false,
				Conn:  conn,
			}

			game.mu.Lock()
			game.Players[player.ID] = player
			game.mu.Unlock()

			broadcastState(game)

		case "next-game":
			if game != nil && player != nil {
				handleNextGame(game, player)
			}

		case "ping":
			// Respond to keep-alive
			continue
		}
	}
}

func handleNextGame(game *Game, player *Player) {
	game.mu.Lock()
	defer game.mu.Unlock()

	switch game.State {
	case "lobby", "":
		// Start game - go to instructions
		game.State = "instructions"
		game.CurrentGame = "madlibs" // For now, always Madlibs
		broadcastState(game)

	case "instructions":
		// Mark player as ready
		player.Ready = true

		// Check if all players are ready
		allReady := true
		for _, p := range game.Players {
			if !p.Ready {
				allReady = false
				break
			}
		}

		if allReady {
			// Start the actual game
			game.State = "playing"
			// Reset ready status
			for _, p := range game.Players {
				p.Ready = false
			}
			broadcastState(game)
		} else {
			broadcastState(game)
		}
	}
}

func broadcastState(game *Game) {
	game.mu.RLock()
	defer game.mu.RUnlock()

	var state StateMessage

	playerCount := len(game.Players)

	switch game.State {
	case "lobby", "":
		if playerCount == 0 {
			state.State.GameTitle = "Waiting for players to join"
			state.State.GameInstructions = "Wait for players to join"
			state.State.RoundInstructions = "Share this URL with friends to play together"
		} else if playerCount == 1 {
			state.State.GameTitle = "Ready to play!"
			state.State.GameInstructions = "Click 'Next' to start a game"
			state.State.RoundInstructions = "Most games work solo, but are more fun with friends!"
		} else {
			state.State.GameTitle = "Waiting for more players"
			state.State.GameInstructions = "Wait for other players or click 'Next' to start"
			state.State.RoundInstructions = "Ready to play?"
		}

	case "instructions":
		state.State.GameTitle = "Madlibs"
		state.State.GameInstructions = "Supply the requested parts of speech until the Madlib is complete"
		state.State.RoundInstructions = "Waiting for all players to hit ready."

	case "playing":
		state.State.GameTitle = "Madlibs"
		state.State.GameInstructions = "Fill in the blanks!"
		state.State.RoundInstructions = "Please enter: [adjective]"
	}

	// Add players to state
	state.State.Players = make([]map[string]interface{}, 0, len(game.Players))
	for _, p := range game.Players {
		state.State.Players = append(state.State.Players, map[string]interface{}{
			"id":    p.ID,
			"name":  p.Name,
			"score": p.Score,
			"ready": p.Ready,
		})
	}

	// Broadcast to all players
	for _, p := range game.Players {
		p.mu.Lock()
		if p.Conn != nil {
			p.Conn.WriteJSON(state)
		}
		p.mu.Unlock()
	}
}

func removePlayer(game *Game, player *Player) {
	game.mu.Lock()
	defer game.mu.Unlock()

	delete(game.Players, player.ID)
	broadcastState(game)
}

func getOrCreateGame(groupName string) *Game {
	gamesMu.Lock()
	defer gamesMu.Unlock()

	if game, exists := games[groupName]; exists {
		return game
	}

	game := &Game{
		ID:      groupName,
		Players: make(map[string]*Player),
		State:   "lobby",
	}
	games[groupName] = game
	return game
}

func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
