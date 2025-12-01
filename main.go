package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

var coordinator *GameCoordinator

func main() {
	coordinator = NewGameCoordinator()

	// Cleanup empty games periodically
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			coordinator.RemoveEmptyGames()
		}
	}()

	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/api/user", handleUser)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	// Check for X-Remote-User header from nginx forward auth
	remoteUser := r.Header.Get("X-Remote-User")

	response := map[string]interface{}{
		"authenticated": remoteUser != "",
		"name":          remoteUser,
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

	var gameActor *GameActor
	var playerID string

	defer func() {
		if gameActor != nil && playerID != "" {
			gameActor.Send(PlayerLeaveMsg{PlayerID: playerID})
		}
		conn.Close()
	}()

	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		action, ok := msg["action"].(string)
		if !ok {
			continue
		}

		data, _ := msg["data"].(map[string]interface{})

		switch action {
		case "join":
			gameID, _ := data["group"].(string)
			playerName, _ := data["name"].(string)
			playerID = generatePlayerID()

			if gameID == "" {
				gameID = "default"
			}

			gameActor = coordinator.GetOrCreateGame(gameID)
			gameActor.Send(PlayerJoinMsg{
				GameID:     gameID,
				PlayerID:   playerID,
				PlayerName: playerName,
				Conn:       conn,
			})

		case "next-game":
			if gameActor != nil {
				gameActor.Send(NextGameMsg{PlayerID: playerID})
			}

		case "ping":
			if gameActor != nil {
				gameActor.Send(PingMsg{PlayerID: playerID})
			}

		case "submit-word":
			if gameActor != nil {
				word, _ := data["word"].(string)
				gameActor.Send(SubmitWordMsg{
					PlayerID: playerID,
					Word:     word,
				})
			}

		case "vote":
			if gameActor != nil {
				votedForID, _ := data["player_id"].(string)
				gameActor.Send(VoteMsg{
					PlayerID:   playerID,
					VotedForID: votedForID,
				})
			}
		}
	}
}

func generatePlayerID() string {
	return time.Now().Format("20060102150405.000000")
}
