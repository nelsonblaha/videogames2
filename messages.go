package main

import "github.com/gorilla/websocket"

// Player actor messages
type PlayerJoinMsg struct {
	GameID     string
	PlayerID   string
	PlayerName string
	Conn       *websocket.Conn
}

func (m PlayerJoinMsg) ActorMessage() {}

type PlayerLeaveMsg struct {
	PlayerID string
}

func (m PlayerLeaveMsg) ActorMessage() {}

type NextGameMsg struct {
	PlayerID string
}

func (m NextGameMsg) ActorMessage() {}

type PingMsg struct {
	PlayerID string
}

func (m PingMsg) ActorMessage() {}

type PlayerReadyMsg struct {
	PlayerID string
}

func (m PlayerReadyMsg) ActorMessage() {}

type SubmitWordMsg struct {
	PlayerID string
	Word     string
}

func (m SubmitWordMsg) ActorMessage() {}

type RequestPromptMsg struct {
	PlayerID string
}

func (m RequestPromptMsg) ActorMessage() {}

type VoteMsg struct {
	PlayerID     string
	VotedForID   string
}

func (m VoteMsg) ActorMessage() {}

type TimerTickMsg struct{}

func (m TimerTickMsg) ActorMessage() {}

// Game state broadcast message
type BroadcastStateMsg struct{}

func (m BroadcastStateMsg) ActorMessage() {}

// Get game state message (for queries)
type GetGameStateMsg struct {
	ResponseChan chan *GameState
}

func (m GetGameStateMsg) ActorMessage() {}

// GameState represents the current state of a game
type GameState struct {
	ID          string
	State       string // "lobby", "instructions", "playing"
	CurrentGame string
	Players     map[string]*PlayerInfo
}

type PlayerInfo struct {
	ID    string
	Name  string
	Score int
	Ready bool
}
