package main

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Game interface for all games
type GameType interface {
	GetName() string
	GetInstructions() string
	GetID() string
	NeedsInput() bool
	GetPrompt() string
	SubmitAnswer(playerID, answer string) bool
	IsComplete() bool
	GetResult() string
}

// Game registry
var AllGames = []string{
	"madlibs",
	"charades",
	"claudesgame",
	"firsttofind",
	"imitations",
	"blankestblank",
	"youlaughyoulose",
}

func CreateGame(gameType string) GameType {
	switch gameType {
	case "madlibs":
		return NewMadLib()
	case "charades":
		return NewCharades()
	case "claudesgame":
		return NewClaudesGame()
	case "firsttofind":
		return NewFirstToFind()
	case "imitations":
		return NewImitations()
	case "blankestblank":
		return NewBlankestBlank()
	case "youlaughyoulose":
		return NewYouLaughYouLose()
	default:
		return NewMadLib()
	}
}

func RandomGameType() string {
	return AllGames[rand.Intn(len(AllGames))]
}

// Charades game
type Charades struct {
	topic   string
	actor   string
	guessed bool
}

var charadeTopics = []string{
	"Titanic", "Star Wars", "treadmill", "sailing",
	"flying a drone", "sleeping in a hammock", "Superman",
	"cooking pasta", "riding a bicycle", "swimming",
}

func NewCharades() *Charades {
	return &Charades{
		topic:   charadeTopics[rand.Intn(len(charadeTopics))],
		guessed: false,
	}
}

func (c *Charades) GetName() string         { return "Charades" }
func (c *Charades) GetInstructions() string { return "Silently act out the topic" }
func (c *Charades) GetID() string           { return "charades" }
func (c *Charades) NeedsInput() bool        { return true }
func (c *Charades) GetPrompt() string       { return "Guess what's being acted out!" }
func (c *Charades) SubmitAnswer(playerID, answer string) bool {
	// TODO: Check if answer matches topic
	c.guessed = true
	return true
}
func (c *Charades) IsComplete() bool { return c.guessed }
func (c *Charades) GetResult() string {
	return "The topic was: " + c.topic
}

// Claude's Game
type ClaudesGame struct {
	word1       string
	word2       string
	submissions map[string]string
	maxAnswers  int
}

var claudesGameWords = []string{
	"banana", "spaceship", "umbrella", "dinosaur", "piano",
	"volcano", "penguin", "telescope", "sandcastle", "lightning",
}

func NewClaudesGame() *ClaudesGame {
	shuffled := rand.Perm(len(claudesGameWords))
	return &ClaudesGame{
		word1:       claudesGameWords[shuffled[0]],
		word2:       claudesGameWords[shuffled[1]],
		submissions: make(map[string]string),
		maxAnswers:  3,
	}
}

func (c *ClaudesGame) GetName() string         { return "Claude's Game" }
func (c *ClaudesGame) GetInstructions() string { return "Connect two words creatively!" }
func (c *ClaudesGame) GetID() string           { return "claudesgame" }
func (c *ClaudesGame) NeedsInput() bool        { return true }
func (c *ClaudesGame) GetPrompt() string {
	return "How are " + c.word1 + " and " + c.word2 + " connected?"
}
func (c *ClaudesGame) SubmitAnswer(playerID, answer string) bool {
	c.submissions[playerID] = answer
	return len(c.submissions) >= c.maxAnswers
}
func (c *ClaudesGame) IsComplete() bool { return len(c.submissions) >= c.maxAnswers }
func (c *ClaudesGame) GetResult() string {
	result := "Connections:\n"
	for _, answer := range c.submissions {
		result += "- " + answer + "\n"
	}
	return result
}

// First to Find
type FirstToFind struct {
	item  string
	found bool
}

var itemsToFind = []string{
	"banana", "pen", "musical instrument", "ice cube",
	"compact disc", "something green", "a hat", "a book",
}

func NewFirstToFind() *FirstToFind {
	return &FirstToFind{
		item:  itemsToFind[rand.Intn(len(itemsToFind))],
		found: false,
	}
}

func (f *FirstToFind) GetName() string { return "First to Find" }
func (f *FirstToFind) GetInstructions() string {
	return "First to find and show the item on camera wins!"
}
func (f *FirstToFind) GetID() string          { return "firsttofind" }
func (f *FirstToFind) NeedsInput() bool       { return false }
func (f *FirstToFind) GetPrompt() string      { return "Find: " + f.item }
func (f *FirstToFind) SubmitAnswer(string, string) bool { f.found = true; return true }
func (f *FirstToFind) IsComplete() bool       { return f.found }
func (f *FirstToFind) GetResult() string      { return "Item found: " + f.item }

// Imitations
type Imitations struct {
	person  string
	guessed bool
}

var peopleToImitate = []string{
	"George Washington", "Clint Eastwood", "Captain Kirk",
	"Barack Obama", "Morgan Freeman", "Donald Duck",
}

func NewImitations() *Imitations {
	return &Imitations{
		person:  peopleToImitate[rand.Intn(len(peopleToImitate))],
		guessed: false,
	}
}

func (i *Imitations) GetName() string { return "Imitations" }
func (i *Imitations) GetInstructions() string {
	return "Imitate the person without saying their name!"
}
func (i *Imitations) GetID() string     { return "imitations" }
func (i *Imitations) NeedsInput() bool  { return true }
func (i *Imitations) GetPrompt() string { return "Guess who's being imitated!" }
func (i *Imitations) SubmitAnswer(playerID, answer string) bool {
	i.guessed = true
	return true
}
func (i *Imitations) IsComplete() bool { return i.guessed }
func (i *Imitations) GetResult() string {
	return "The person was: " + i.person
}

// Find the Blankest Blank
type BlankestBlank struct {
	adjective string
	noun      string
	found     bool
}

var adjectives = []string{
	"oldest", "biggest", "fanciest", "most bizarre",
	"smallest", "newest", "trendiest", "weirdest", "pinkest",
}

var nouns = []string{
	"thing", "food", "kitchen utensil", "costume",
	"coin", "book", "hat", "toy",
}

func NewBlankestBlank() *BlankestBlank {
	return &BlankestBlank{
		adjective: adjectives[rand.Intn(len(adjectives))],
		noun:      nouns[rand.Intn(len(nouns))],
		found:     false,
	}
}

func (b *BlankestBlank) GetName() string { return "Find the Blankest Blank" }
func (b *BlankestBlank) GetInstructions() string {
	return "Find the " + b.adjective + " " + b.noun + " and show it!"
}
func (b *BlankestBlank) GetID() string     { return "blankestblank" }
func (b *BlankestBlank) NeedsInput() bool  { return false }
func (b *BlankestBlank) GetPrompt() string { return "Find: " + b.adjective + " " + b.noun }
func (b *BlankestBlank) SubmitAnswer(string, string) bool {
	b.found = true
	return true
}
func (b *BlankestBlank) IsComplete() bool  { return b.found }
func (b *BlankestBlank) GetResult() string { return "Found the " + b.adjective + " " + b.noun }

// You Laugh You Lose
type YouLaughYouLose struct {
	videoID  string
	duration int
	elapsed  int
}

var funnyVideos = []string{
	"XCPj4JPbKtA", "nFAK8Vj62WM", "0H25ve3qts4", "Q9zvgcOrTtw",
	"Veg63B8ofnQ", "tjiouAv0-Gk", "oaTxUeZWC4M", "BKInDainD5M",
}

func NewYouLaughYouLose() *YouLaughYouLose {
	return &YouLaughYouLose{
		videoID:  funnyVideos[rand.Intn(len(funnyVideos))],
		duration: 90,
		elapsed:  0,
	}
}

func (y *YouLaughYouLose) GetName() string { return "You Laugh You Lose" }
func (y *YouLaughYouLose) GetInstructions() string {
	return "Last person to keep a straight face wins!"
}
func (y *YouLaughYouLose) GetID() string          { return "youlaughyoulose" }
func (y *YouLaughYouLose) NeedsInput() bool       { return false }
func (y *YouLaughYouLose) GetPrompt() string      { return "Watch and don't laugh!" }
func (y *YouLaughYouLose) SubmitAnswer(string, string) bool { return false }
func (y *YouLaughYouLose) IsComplete() bool       { return y.elapsed >= y.duration }
func (y *YouLaughYouLose) GetResult() string      { return "Video ID: " + y.videoID }
