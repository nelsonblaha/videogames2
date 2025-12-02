package main

import (
	"math/rand"
	"time"
)

type MadLib struct {
	Template      string
	Prompts       []string
	Words         []string
	playerPrompts map[string]int // tracks which prompt index each player is currently on
}

var madLibTemplates = []MadLib{
	{
		Template: "Once upon a time, there was a {adjective} {noun} who loved to {verb} in the {place}. One day, they met a {adjective} {animal} who said '{exclamation}!'",
		Prompts:  []string{"adjective", "noun", "verb", "place", "adjective", "animal", "exclamation"},
	},
	{
		Template: "I went to the {place} to buy a {adjective} {noun}. The {occupation} behind the counter said it would cost {number} {plural_noun}! I said '{exclamation}' and {verb_past_tense} away.",
		Prompts:  []string{"place", "adjective", "noun", "occupation", "number", "plural_noun", "exclamation", "verb_past_tense"},
	},
	{
		Template: "My {adjective} {family_member} always told me to {verb} every day. They said if I didn't, a {adjective} {animal} would come and {verb} my {noun}. What {adjective} advice!",
		Prompts:  []string{"adjective", "family_member", "verb", "adjective", "animal", "verb", "noun", "adjective"},
	},
	{
		Template: "Breaking news: A {adjective} {noun} was spotted {verb_ing} through {place} today. Witnesses described it as '{adjective}' and '{adjective}'. The local {occupation} had no comment.",
		Prompts:  []string{"adjective", "noun", "verb_ing", "place", "adjective", "adjective", "occupation"},
	},
	{
		Template: "Dear {person_name}, I am writing to inform you that your {adjective} {noun} has been {verb_past_tense}. Please come to the {place} and bring {number} {plural_noun}. Sincerely, The {adjective} {occupation}",
		Prompts:  []string{"person_name", "adjective", "noun", "verb_past_tense", "place", "number", "plural_noun", "adjective", "occupation"},
	},
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewMadLib() *MadLib {
	template := madLibTemplates[rand.Intn(len(madLibTemplates))]
	return &MadLib{
		Template:      template.Template,
		Prompts:       template.Prompts,
		Words:         make([]string, len(template.Prompts)),
		playerPrompts: make(map[string]int),
	}
}

// AddWordForPlayer adds a word from a specific player to the next available slot
func (m *MadLib) AddWordForPlayer(playerID, word string) bool {
	// Find the next empty slot in the Words array
	for i, w := range m.Words {
		if w == "" {
			m.Words[i] = word
			// Update this player's position to the next empty slot
			m.updatePlayerPrompt(playerID)
			// Return true if all words are now filled
			return m.IsComplete()
		}
	}
	return true
}

// AddWord is kept for backward compatibility with the GameType interface
func (m *MadLib) AddWord(word string) bool {
	return m.AddWordForPlayer("", word)
}

// GetPromptForPlayer returns the current prompt for a specific player
func (m *MadLib) GetPromptForPlayer(playerID string) string {
	// Get or initialize this player's current position
	idx, exists := m.playerPrompts[playerID]
	if !exists {
		idx = m.findNextEmptySlot()
		m.playerPrompts[playerID] = idx
	}

	if idx < len(m.Prompts) {
		return m.Prompts[idx]
	}
	return ""
}

// findNextEmptySlot finds the next unfilled word slot
func (m *MadLib) findNextEmptySlot() int {
	for i, w := range m.Words {
		if w == "" {
			return i
		}
	}
	return len(m.Words) // All filled
}

// updatePlayerPrompt moves player to the next empty slot
func (m *MadLib) updatePlayerPrompt(playerID string) {
	nextSlot := m.findNextEmptySlot()
	m.playerPrompts[playerID] = nextSlot
}

// CurrentPrompt returns the next unfilled prompt (used for interface compatibility)
func (m *MadLib) CurrentPrompt() string {
	for i, w := range m.Words {
		if w == "" {
			return m.Prompts[i]
		}
	}
	return ""
}

func (m *MadLib) IsComplete() bool {
	for _, w := range m.Words {
		if w == "" {
			return false
		}
	}
	return true
}

func (m *MadLib) GetStory() string {
	story := m.Template
	for i, word := range m.Words {
		// Replace {prompt} with actual word
		prompt := "{" + m.Prompts[i] + "}"
		story = replaceFirst(story, prompt, word)
	}
	return story
}

func replaceFirst(s, old, new string) string {
	i := 0
	for j := 0; j <= len(s)-len(old); j++ {
		if s[j:j+len(old)] == old {
			return s[:j] + new + s[j+len(old):]
		}
		i++
	}
	return s
}

// Implement GameType interface
func (m *MadLib) GetName() string         { return "Mad Libs" }
func (m *MadLib) GetInstructions() string { return "Fill in the blanks with words!" }
func (m *MadLib) GetID() string           { return "madlibs" }
func (m *MadLib) NeedsInput() bool        { return true }
func (m *MadLib) GetPrompt() string       { return m.CurrentPrompt() }
func (m *MadLib) SubmitAnswer(playerID, answer string) bool {
	return m.AddWord(answer)
}
func (m *MadLib) GetResult() string       { return m.GetStory() }
func (m *MadLib) HasTimer() bool          { return false }
func (m *MadLib) GetTimeRemaining() int   { return 0 }
func (m *MadLib) DecrementTimer()         {}
