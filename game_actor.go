package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"sync"

	"github.com/gorilla/websocket"
)

// GameActor manages a single game session using the actor model
type GameActor struct {
	id          string
	state       string // "lobby", "instructions", "playing", "voting", "finished"
	currentGame string
	players     map[string]*Player
	game        GameType
	votes       map[string]string // playerID -> votedForPlayerID
	winners     []string          // names of winners from last vote
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
	case VoteMsg:
		ga.handleVote(m)
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
		// Pick a random game appropriate for player count and move to instructions
		ga.state = "instructions"
		ga.currentGame = RandomGameTypeForPlayers(len(ga.players))
		ga.broadcastState()

	case "instructions":
		// Mark player as ready
		if player, exists := ga.players[msg.PlayerID]; exists {
			player.Ready = true
			log.Printf("Player %s marked ready in game %s", player.Name, ga.id)
		}

		// Check if all players are ready
		allReady := true
		readyCount := 0
		for _, p := range ga.players {
			if p.Ready {
				readyCount++
			} else {
				allReady = false
			}
		}
		log.Printf("Ready check: %d/%d players ready in game %s", readyCount, len(ga.players), ga.id)

		if allReady && len(ga.players) > 0 {
			log.Printf("All players ready! Starting %s in game %s", ga.currentGame, ga.id)
			ga.state = "playing"
			ga.game = CreateGame(ga.currentGame)

			// Set number of players for Claude's Game
			if cg, ok := ga.game.(*ClaudesGame); ok {
				cg.numPlayers = len(ga.players)
			}

			// Set random actor for Imitations
			if im, ok := ga.game.(*Imitations); ok && len(ga.players) > 0 {
				// Pick random player as actor
				playerIDs := make([]string, 0, len(ga.players))
				for id := range ga.players {
					playerIDs = append(playerIDs, id)
				}
				actorID := playerIDs[rand.Intn(len(playerIDs))]
				im.SetActor(actorID)
				log.Printf("Set %s as actor for Imitations in game %s", actorID, ga.id)
			}

			// Set random actor for Charades
			if ch, ok := ga.game.(*Charades); ok && len(ga.players) > 0 {
				// Pick random player as actor
				playerIDs := make([]string, 0, len(ga.players))
				for id := range ga.players {
					playerIDs = append(playerIDs, id)
				}
				actorID := playerIDs[rand.Intn(len(playerIDs))]
				ch.SetActor(actorID)
				log.Printf("Set %s as actor for Charades in game %s", actorID, ga.id)
			}

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
		// Pick next random game appropriate for player count and move to instructions
		log.Printf("Game finished in %s, picking next game", ga.id)
		ga.state = "instructions"
		ga.currentGame = RandomGameTypeForPlayers(len(ga.players))
		log.Printf("Next game will be: %s", ga.currentGame)
		ga.game = nil
		ga.winners = nil
		for _, p := range ga.players {
			p.Ready = false
		}
		ga.broadcastState()
	}
}

func (ga *GameActor) handleSubmitWord(msg SubmitWordMsg) {
	ga.mu.Lock()
	defer ga.mu.Unlock()

	if ga.state != "playing" || ga.game == nil {
		return
	}

	// Submit word/answer to current game
	isComplete := ga.game.SubmitAnswer(msg.PlayerID, msg.Word)

	// Award points based on game type
	if msg.Word != "timer_complete" && msg.Word != "video_complete" {
		// For Imitations, only award points to the winner
		if im, ok := ga.game.(*Imitations); ok {
			if isComplete && im.GetWinner() == msg.PlayerID {
				if player, exists := ga.players[msg.PlayerID]; exists {
					player.Score += 3 // Award 3 points for guessing correctly
					im.SetWinnerName(player.Name)
				}
			}
		} else if ch, ok := ga.game.(*Charades); ok {
			// For Charades, only award points to the winner
			if isComplete && ch.GetWinner() == msg.PlayerID {
				if player, exists := ga.players[msg.PlayerID]; exists {
					player.Score += 3 // Award 3 points for guessing correctly
					ch.SetWinnerName(player.Name)
				}
			}
		} else {
			// For other games, award 1 point per submission
			if player, exists := ga.players[msg.PlayerID]; exists {
				player.Score++
			}
		}
	}

	if isComplete {
		// Games that need voting: youlaughyoulose, firsttofind, blankestblank, claudesgame
		needsVoting := ga.currentGame == "youlaughyoulose" ||
			ga.currentGame == "firsttofind" ||
			ga.currentGame == "blankestblank" ||
			ga.currentGame == "claudesgame"

		if needsVoting {
			ga.state = "voting"
			ga.votes = make(map[string]string)
		} else {
			ga.state = "finished"
		}
	}

	ga.broadcastState()
}

func (ga *GameActor) handleVote(msg VoteMsg) {
	ga.mu.Lock()
	defer ga.mu.Unlock()

	if ga.state != "voting" {
		return
	}

	// Record vote
	ga.votes[msg.PlayerID] = msg.VotedForID

	// Check if all players have voted
	if len(ga.votes) >= len(ga.players) {
		// Count votes
		voteCounts := make(map[string]int)
		for _, votedFor := range ga.votes {
			voteCounts[votedFor]++
		}

		// Find winner(s)
		maxVotes := 0
		for _, count := range voteCounts {
			if count > maxVotes {
				maxVotes = count
			}
		}

		// Award points to winner(s)
		ga.winners = []string{}
		for playerID, count := range voteCounts {
			if count == maxVotes {
				if player, exists := ga.players[playerID]; exists {
					player.Score += 3
					ga.winners = append(ga.winners, player.Name)
				}
			}
		}

		ga.state = "finished"
		ga.broadcastState()
	} else {
		ga.broadcastState()
	}
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
		if ga.currentGame != "" {
			tempGame := CreateGame(ga.currentGame)
			// For BlankestBlank, use the prompt as the title instead of the generic game name
			if ga.currentGame == "blankestblank" {
				gameTitle = tempGame.GetPrompt()
			} else {
				gameTitle = tempGame.GetName()
			}
			gameInstructions = tempGame.GetInstructions()
		} else {
			gameTitle = "Get Ready!"
			gameInstructions = "Prepare for the next game"
		}
		if numPlayers == 1 {
			roundInstructions = "Click 'Next' when ready to play"
		} else {
			roundInstructions = "Everyone click 'Next' when ready"
		}

	case "playing":
		if ga.game != nil {
			// Use the prompt as the big title text
			gameTitle = ga.game.GetPrompt()
			if ga.game.NeedsInput() {
				gameInstructions = "Enter your answer:"
				roundInstructions = ""
			} else {
				gameInstructions = ""
				roundInstructions = ""
			}
		} else {
			gameTitle = "Playing..."
			gameInstructions = "Loading..."
			roundInstructions = ""
		}

	case "voting":
		gameTitle = "Time to Vote!"
		if ga.game != nil {
			gameInstructions = ga.game.GetResult()
			roundInstructions = "Vote for the winner!"
		} else {
			gameInstructions = "Who kept the straightest face?"
			roundInstructions = "Vote for the person who didn't laugh!"
		}

	case "finished":
		gameTitle = "Game Complete!"
		if ga.game != nil {
			gameInstructions = ga.game.GetName() + " finished!"
			if len(ga.winners) > 0 {
				if len(ga.winners) == 1 {
					roundInstructions = ga.winners[0] + " wins! " + ga.game.GetResult()
				} else {
					// Tie
					winnersList := ""
					for i, w := range ga.winners {
						if i > 0 {
							winnersList += ", "
						}
						winnersList += w
					}
					roundInstructions = "Tie! " + winnersList + " win! " + ga.game.GetResult()
				}
			} else {
				roundInstructions = ga.game.GetResult()
			}
		} else {
			gameInstructions = "Click Next for another game"
			roundInstructions = ""
		}
	}

	stateData := map[string]interface{}{
		"game_title":         gameTitle,
		"game_instructions":  gameInstructions,
		"round_instructions": roundInstructions,
		"players":            playersList,
		"game_state":         ga.state,
		"game_type":          ga.currentGame,
		"needs_input":        ga.game != nil && ga.game.NeedsInput(),
	}

	// Add timer data if game has a timer
	if ga.state == "playing" && ga.game != nil && ga.game.HasTimer() {
		stateData["has_timer"] = true
		stateData["time_remaining"] = ga.game.GetTimeRemaining()
	}

	// Add YouTube video ID for You Laugh You Lose
	if ga.currentGame == "youlaughyoulose" && ga.game != nil {
		if ylyl, ok := ga.game.(*YouLaughYouLose); ok {
			stateData["youtube_video_id"] = ylyl.videoID
		}
	}

	// Add voting data
	if ga.state == "voting" {
		votedPlayers := []string{}
		for playerID := range ga.votes {
			votedPlayers = append(votedPlayers, playerID)
		}
		stateData["voted_players"] = votedPlayers
		stateData["total_votes"] = len(ga.votes)
		stateData["expected_votes"] = len(ga.players)
	}

	// Add game-specific data for Mad Libs
	if ga.state == "playing" && ga.currentGame == "madlibs" && ga.game != nil {
		if madlib, ok := ga.game.(*MadLib); ok {
			stateData["current_prompt"] = madlib.CurrentPrompt()
			stateData["words_collected"] = len(madlib.Words) - countEmpty(madlib.Words)
			stateData["total_words"] = len(madlib.Words)
		}
	}

	// Add completed result if finished
	if ga.state == "finished" && ga.game != nil {
		stateData["story"] = ga.game.GetResult()
	}

	stateMsg := map[string]interface{}{
		"state": stateData,
	}

	// Send to all players (personalized for Imitations)
	for _, player := range ga.players {
		player.mu.Lock()
		if player.Conn != nil {
			// Personalize message for Imitations and Charades games
			playerStateMsg := stateMsg
			if ga.state == "playing" && ga.currentGame == "imitations" {
				if im, ok := ga.game.(*Imitations); ok {
					playerStateData := make(map[string]interface{})
					for k, v := range stateData {
						playerStateData[k] = v
					}

					// Actor gets told who to imitate
					if player.ID == im.GetActor() {
						playerStateData["game_title"] = "Imitate " + im.GetPerson() + "!"
						playerStateData["game_instructions"] = ""
						playerStateData["round_instructions"] = ""
						playerStateData["needs_input"] = false // Actor doesn't submit anything
					} else {
						// Guessers get the normal prompt
						playerStateData["game_title"] = "Guess who's being imitated!"
						playerStateData["game_instructions"] = "Enter your answer:"
						playerStateData["round_instructions"] = ""
						playerStateData["needs_input"] = true // Guesser needs to submit
					}

					playerStateMsg = map[string]interface{}{
						"state": playerStateData,
					}
				}
			} else if ga.state == "playing" && ga.currentGame == "charades" {
				if ch, ok := ga.game.(*Charades); ok {
					playerStateData := make(map[string]interface{})
					for k, v := range stateData {
						playerStateData[k] = v
					}

					// Actor gets told what to act out
					if player.ID == ch.GetActor() {
						playerStateData["game_title"] = "Act out: " + ch.GetTopic() + "!"
						playerStateData["game_instructions"] = ""
						playerStateData["round_instructions"] = ""
						playerStateData["needs_input"] = false // Actor doesn't submit anything
					} else {
						// Guessers get the normal prompt
						playerStateData["game_title"] = "Guess what's being acted out!"
						playerStateData["game_instructions"] = "Enter your answer:"
						playerStateData["round_instructions"] = ""
						playerStateData["needs_input"] = true // Guesser needs to submit
					}

					playerStateMsg = map[string]interface{}{
						"state": playerStateData,
					}
				}
			}

			playerJsonData, err := json.Marshal(playerStateMsg)
			if err != nil {
				log.Printf("Error marshaling player state: %v", err)
				player.mu.Unlock()
				continue
			}

			err = player.Conn.WriteMessage(websocket.TextMessage, playerJsonData)
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
