package main

import (
	"math/rand"
	"time"
)

type MadLib struct {
	Template      string
	Prompts       []string
	Words         []string
	claimedBy     []string          // tracks which player has claimed each slot ("" = unclaimed)
	playerPrompts map[string]int    // tracks which prompt index each player is currently on
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
	numPrompts := len(template.Prompts)
	return &MadLib{
		Template:      template.Template,
		Prompts:       template.Prompts,
		Words:         make([]string, numPrompts),
		claimedBy:     make([]string, numPrompts), // all start as ""
		playerPrompts: make(map[string]int),
	}
}

// AddWordForPlayer fills the slot claimed by this player
func (m *MadLib) AddWordForPlayer(playerID, word string) bool {
	// Find the slot this player has claimed
	idx, exists := m.playerPrompts[playerID]
	if !exists {
		return m.IsComplete()
	}

	// Verify this slot is actually claimed by this player
	if idx >= len(m.claimedBy) || m.claimedBy[idx] != playerID {
		return m.IsComplete()
	}

	// Fill the slot
	m.Words[idx] = word

	// Clear this player's claim and prompt them for next slot
	m.claimNextSlotForPlayer(playerID)

	// Return true if all words are now filled
	return m.IsComplete()
}

// AddWord is kept for backward compatibility with the GameType interface
func (m *MadLib) AddWord(word string) bool {
	return m.AddWordForPlayer("", word)
}

// GetPromptForPlayer returns the current prompt for a specific player (query only, no mutation)
func (m *MadLib) GetPromptForPlayer(playerID string) string {
	// Check if player already has a claimed slot
	idx, exists := m.playerPrompts[playerID]
	if !exists {
		return "" // Player hasn't claimed a slot yet
	}

	// Return the prompt for their claimed slot, or empty if no slots available
	if idx >= 0 && idx < len(m.Prompts) {
		return m.Prompts[idx]
	}
	return ""
}

// ClaimSlotForPlayer explicitly claims the next available slot for a player
// This is a command (mutation) and should be called through the actor message handler
func (m *MadLib) ClaimSlotForPlayer(playerID string) bool {
	// Check if player already has a claimed slot
	if _, exists := m.playerPrompts[playerID]; exists {
		return true // Already has a slot
	}

	// Claim the next available slot
	idx := m.claimNextSlotForPlayer(playerID)
	return idx >= 0 // Returns true if slot was claimed, false if none available
}

// claimNextSlotForPlayer finds and claims the next unclaimed slot for a player
// Returns the slot index, or -1 if no slots available
func (m *MadLib) claimNextSlotForPlayer(playerID string) int {
	// Find first unclaimed slot
	for i, claimer := range m.claimedBy {
		if claimer == "" {
			// Claim this slot for the player
			m.claimedBy[i] = playerID
			m.playerPrompts[playerID] = i
			return i
		}
	}
	// No slots available - all are claimed or filled
	m.playerPrompts[playerID] = -1
	return -1
}

// findNextEmptySlot finds the next unfilled word slot (for backward compatibility)
func (m *MadLib) findNextEmptySlot() int {
	for i, w := range m.Words {
		if w == "" {
			return i
		}
	}
	return len(m.Words) // All filled
}

// HasAvailableSlots returns true if there are unclaimed slots
func (m *MadLib) HasAvailableSlots() bool {
	for _, claimer := range m.claimedBy {
		if claimer == "" {
			return true
		}
	}
	return false
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
