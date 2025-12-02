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
	HasTimer() bool
	GetTimeRemaining() int
	DecrementTimer()
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

// MinPlayersRequired returns the minimum number of players required for a game
func MinPlayersRequired(gameType string) int {
	switch gameType {
	case "imitations", "charades":
		return 2 // Needs 1 actor + at least 1 guesser
	default:
		return 1 // Most games work with 1 player
	}
}

// RandomGameTypeForPlayers returns a random game type appropriate for the player count
func RandomGameTypeForPlayers(playerCount int) string {
	validGames := []string{}
	for _, game := range AllGames {
		if MinPlayersRequired(game) <= playerCount {
			validGames = append(validGames, game)
		}
	}

	if len(validGames) == 0 {
		return "madlibs" // fallback
	}

	return validGames[rand.Intn(len(validGames))]
}

// Charades game
type Charades struct {
	topic       string
	actorID     string
	guessed     bool
	winnerID    string
	winnerName  string
	submissions map[string]string // Track all guesses
}

var charadeTopics = []string{
	"Titanic", "Star Wars", "treadmill", "sailing",
	"flying a drone", "sleeping in a hammock", "Superman",
	"cooking pasta", "riding a bicycle", "swimming",
}

func NewCharades() *Charades {
	return &Charades{
		topic:       charadeTopics[rand.Intn(len(charadeTopics))],
		guessed:     false,
		submissions: make(map[string]string),
	}
}

func (c *Charades) SetActor(actorID string) {
	c.actorID = actorID
}

func (c *Charades) GetActor() string {
	return c.actorID
}

func (c *Charades) GetTopic() string {
	return c.topic
}

func (c *Charades) GetName() string         { return "Charades" }
func (c *Charades) GetInstructions() string { return "Silently act out the topic" }
func (c *Charades) GetID() string           { return "charades" }
func (c *Charades) NeedsInput() bool        { return true }
func (c *Charades) GetPrompt() string       { return "Guess what's being acted out!" }
func (c *Charades) SubmitAnswer(playerID, answer string) bool {
	// Don't allow the actor to guess
	if playerID == c.actorID {
		return false
	}

	// Store the guess
	c.submissions[playerID] = answer

	// Check if answer matches topic using fuzzy matching
	if fuzzyMatch(answer, c.topic) {
		c.guessed = true
		c.winnerID = playerID
		return true
	}

	return false
}
func (c *Charades) IsComplete() bool { return c.guessed }
func (c *Charades) GetResult() string {
	if c.winnerName != "" {
		return c.winnerName + " guessed it! The topic was: " + c.topic
	}
	return "The topic was: " + c.topic
}
func (c *Charades) GetWinner() string {
	return c.winnerID
}
func (c *Charades) SetWinnerName(name string) {
	c.winnerName = name
}
func (c *Charades) HasTimer() bool        { return false }
func (c *Charades) GetTimeRemaining() int { return 0 }
func (c *Charades) DecrementTimer()       {}

// Claude's Game
type ClaudesGame struct {
	word1       string
	word2       string
	submissions map[string]string
	numPlayers  int
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
		numPlayers:  1, // will be updated when first player submits
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
	// Only accept one answer per player
	if _, exists := c.submissions[playerID]; !exists {
		c.submissions[playerID] = answer
	}
	// Complete when all players have submitted (need at least 1)
	return len(c.submissions) >= c.numPlayers && c.numPlayers > 0
}
func (c *ClaudesGame) IsComplete() bool {
	return len(c.submissions) >= c.numPlayers && c.numPlayers > 0
}
func (c *ClaudesGame) GetResult() string {
	result := "Pick the best connection between " + c.word1 + " and " + c.word2 + ":\n"
	for _, answer := range c.submissions {
		result += "- " + answer + "\n"
	}
	return result
}
func (c *ClaudesGame) HasTimer() bool        { return false }
func (c *ClaudesGame) GetTimeRemaining() int { return 0 }
func (c *ClaudesGame) DecrementTimer()       {}

// First to Find
type FirstToFind struct {
	item          string
	timeRemaining int
	timerActive   bool
}

var itemsToFind = []string{
	"banana",
	"working electrical device over 50 years old",
	"blade of grass",
	"pen",
	"musical instrument",
	"ice cube",
	"compact disc",
}

func NewFirstToFind() *FirstToFind {
	return &FirstToFind{
		item:          itemsToFind[rand.Intn(len(itemsToFind))],
		timeRemaining: 30,
		timerActive:   false, // Timer starts when players click "Start"
	}
}

func (f *FirstToFind) GetName() string { return "First to Find" }
func (f *FirstToFind) GetInstructions() string {
	return "First to show a " + f.item + " wins!"
}
func (f *FirstToFind) GetID() string     { return "firsttofind" }
func (f *FirstToFind) NeedsInput() bool  { return false }
func (f *FirstToFind) GetPrompt() string { return "First to show a " + f.item + " wins!" }
func (f *FirstToFind) SubmitAnswer(playerID, answer string) bool {
	// Timer completion triggers voting
	if answer == "timer_complete" {
		f.timerActive = false
		return true
	}
	return false
}
func (f *FirstToFind) IsComplete() bool  { return !f.timerActive }
func (f *FirstToFind) GetResult() string { return "Time's up! Vote for who showed the best " + f.item }
func (f *FirstToFind) HasTimer() bool    { return true }
func (f *FirstToFind) GetTimeRemaining() int {
	return f.timeRemaining
}
func (f *FirstToFind) DecrementTimer() {
	if f.timeRemaining > 0 {
		f.timeRemaining--
	}
	if f.timeRemaining == 0 {
		f.timerActive = false
	}
}

// Imitations
type Imitations struct {
	person      string
	actorID     string
	guessed     bool
	winnerID    string
	winnerName  string
	submissions map[string]string // Track all guesses
}

var peopleToImitate = []string{
	"George Washington", "Clint Eastwood", "Captain Kirk",
	"Barack Obama", "Morgan Freeman", "Donald Duck",
}

func NewImitations() *Imitations {
	return &Imitations{
		person:      peopleToImitate[rand.Intn(len(peopleToImitate))],
		guessed:     false,
		submissions: make(map[string]string),
	}
}

func (i *Imitations) SetActor(actorID string) {
	i.actorID = actorID
}

func (i *Imitations) GetActor() string {
	return i.actorID
}

func (i *Imitations) GetPerson() string {
	return i.person
}

func (i *Imitations) GetName() string { return "Imitations" }
func (i *Imitations) GetInstructions() string {
	return "Imitate the person without saying their name!"
}
func (i *Imitations) GetID() string    { return "imitations" }
func (i *Imitations) NeedsInput() bool { return true }
func (i *Imitations) GetPrompt() string {
	return "Guess who's being imitated!"
}

// Fuzzy match helper - checks if guess is close to target
func fuzzyMatch(guess, target string) bool {
	// Normalize both strings
	g := normalizeString(guess)
	t := normalizeString(target)

	// Exact match
	if g == t {
		return true
	}

	// Split target into words (first name, last name, etc)
	words := []string{}
	currentWord := ""
	for _, c := range t {
		if c == ' ' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
		} else {
			currentWord += string(c)
		}
	}
	if currentWord != "" {
		words = append(words, currentWord)
	}

	// Check if guess matches any individual word (first or last name)
	for _, word := range words {
		if g == word {
			return true
		}
	}

	// Check if target contains guess or vice versa
	if len(g) >= 3 && (contains(t, g) || contains(g, t)) {
		return true
	}

	// Simple Levenshtein distance for typos (allow 1-2 character differences)
	if len(g) > 3 && len(t) > 3 && levenshteinDistance(g, t) <= 2 {
		return true
	}

	return false
}

func normalizeString(s string) string {
	result := ""
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			if c >= 'A' && c <= 'Z' {
				result += string(c + 32) // lowercase
			} else {
				result += string(c)
			}
		}
	}
	return result
}

func contains(haystack, needle string) bool {
	if len(needle) > len(haystack) {
		return false
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}

			min := matrix[i-1][j] + 1 // deletion
			if matrix[i][j-1]+1 < min {
				min = matrix[i][j-1] + 1 // insertion
			}
			if matrix[i-1][j-1]+cost < min {
				min = matrix[i-1][j-1] + cost // substitution
			}

			matrix[i][j] = min
		}
	}

	return matrix[len(s1)][len(s2)]
}

func (i *Imitations) SubmitAnswer(playerID, answer string) bool {
	// Don't allow the actor to guess
	if playerID == i.actorID {
		return false
	}

	// Store the guess
	i.submissions[playerID] = answer

	// Check if answer is correct using fuzzy matching
	if fuzzyMatch(answer, i.person) {
		i.guessed = true
		i.winnerID = playerID
		return true
	}

	return false
}

func (i *Imitations) IsComplete() bool { return i.guessed }
func (i *Imitations) GetResult() string {
	if i.winnerName != "" {
		return i.winnerName + " guessed it! The person was: " + i.person
	}
	return "The person was: " + i.person
}
func (i *Imitations) GetWinner() string {
	return i.winnerID
}
func (i *Imitations) SetWinnerName(name string) {
	i.winnerName = name
}
func (i *Imitations) HasTimer() bool        { return false }
func (i *Imitations) GetTimeRemaining() int { return 0 }
func (i *Imitations) DecrementTimer()       {}

// Find the Blankest Blank
type BlankestBlank struct {
	adjective     string
	noun          string
	timeRemaining int
	timerActive   bool
}

var adjectives = []string{
	"oldest", "biggest", "fanciest", "most bizarre",
	"smallest", "newest", "trendiest", "weirdest", "pinkest", "best",
}

var nouns = []string{
	"thing", "food", "kitchen utensil", "costume",
	"coin", "book", "hat",
}

func NewBlankestBlank() *BlankestBlank {
	return &BlankestBlank{
		adjective:     adjectives[rand.Intn(len(adjectives))],
		noun:          nouns[rand.Intn(len(nouns))],
		timeRemaining: 30,
		timerActive:   false, // Timer starts when players click "Start"
	}
}

func (b *BlankestBlank) GetName() string { return "Find the Blankest Blank" }
func (b *BlankestBlank) GetInstructions() string {
	return "Click next when everyone has their " + b.noun
}
func (b *BlankestBlank) GetID() string     { return "blankestblank" }
func (b *BlankestBlank) NeedsInput() bool  { return false }
func (b *BlankestBlank) GetPrompt() string { return "Find the " + b.adjective + " " + b.noun + "!" }
func (b *BlankestBlank) GetNoun() string   { return b.noun }
func (b *BlankestBlank) SubmitAnswer(playerID, answer string) bool {
	// Timer completion triggers voting
	if answer == "timer_complete" {
		b.timerActive = false
		return true
	}
	return false
}
func (b *BlankestBlank) IsComplete() bool { return !b.timerActive }
func (b *BlankestBlank) GetResult() string {
	return "Who brought the " + b.adjective + " " + b.noun + "?"
}
func (b *BlankestBlank) HasTimer() bool { return true }
func (b *BlankestBlank) GetTimeRemaining() int {
	return b.timeRemaining
}
func (b *BlankestBlank) DecrementTimer() {
	if b.timeRemaining > 0 {
		b.timeRemaining--
	}
	if b.timeRemaining == 0 {
		b.timerActive = false
	}
}

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
func (y *YouLaughYouLose) GetID() string    { return "youlaughyoulose" }
func (y *YouLaughYouLose) NeedsInput() bool { return false }
func (y *YouLaughYouLose) GetPrompt() string {
	return "Watch and don't laugh!"
}
func (y *YouLaughYouLose) SubmitAnswer(playerID, answer string) bool {
	// Video end triggers voting
	if answer == "video_complete" {
		return true
	}
	return false
}
func (y *YouLaughYouLose) IsComplete() bool  { return y.elapsed >= y.duration }
func (y *YouLaughYouLose) GetResult() string { return "Who kept the straightest face?" }
func (y *YouLaughYouLose) HasTimer() bool    { return false }
func (y *YouLaughYouLose) GetTimeRemaining() int {
	return 0
}
func (y *YouLaughYouLose) DecrementTimer() {}
