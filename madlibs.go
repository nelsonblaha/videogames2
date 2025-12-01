package main

import (
	"math/rand"
	"time"
)

type MadLib struct {
	Template string
	Prompts  []string
	Words    []string
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
		Template: template.Template,
		Prompts:  template.Prompts,
		Words:    make([]string, len(template.Prompts)),
	}
}

func (m *MadLib) AddWord(word string) bool {
	for i, w := range m.Words {
		if w == "" {
			m.Words[i] = word
			return i == len(m.Words)-1 // Return true if this was the last word
		}
	}
	return true
}

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
