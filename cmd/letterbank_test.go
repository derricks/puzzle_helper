package cmd

import (
	"testing"
)

func TestCreateLetterSet(t *testing.T) {
	input := "NEEDLESS"
	expected := map[string]bool{
		"N": true,
		"E": true,
		"D": true,
		"L": true,
		"S": true,
	}
	actual := CreateLetterSet(input)

	if len(actual) != len(expected) {
		t.Errorf("Expected %d unique letters, got %d", len(expected), len(actual))
	}

	for letter := range expected {
		if !actual[letter] {
			t.Errorf("Expected letter %s to be in set", letter)
		}
	}

	// Verify no extra letters
	for letter := range actual {
		if !expected[letter] {
			t.Errorf("Unexpected letter %s in set", letter)
		}
	}
}

func TestCreateLetterSetIgnoresSpaces(t *testing.T) {
	input := "A B C"
	actual := CreateLetterSet(input)
	if len(actual) != 3 {
		t.Errorf("Expected 3 unique letters, got %d", len(actual))
	}
	if actual[" "] {
		t.Error("Space should not be in letter set")
	}
}

func TestLetterBankSolve(t *testing.T) {
	// Build a small trie with known words
	trie := newTrie()
	trie.addValueForString("LENDS", nil)
	trie.addValueForString("NEEDLESS", nil)
	trie.addValueForString("NEEDLES", nil)
	trie.addValueForString("DELL", nil) // missing N and S, should NOT match
	trie.addValueForString("SEND", nil) // missing L, should NOT match

	results := PerformLetterBankSolve("LENDS", trie, 1, 100, 1, 1)

	// LENDS {D,E,L,N,S} should match NEEDLESS {D,E,L,N,S} and NEEDLES {D,E,L,N,S} and itself
	foundWords := make(map[string]bool)
	for _, r := range results {
		if len(r.Words) == 1 {
			foundWords[r.Words[0]] = true
		}
	}

	if !foundWords["LENDS"] {
		t.Error("Expected LENDS to be a letter bank match")
	}
	if !foundWords["NEEDLESS"] {
		t.Error("Expected NEEDLESS to be a letter bank match")
	}
	if !foundWords["NEEDLES"] {
		t.Error("Expected NEEDLES to be a letter bank match")
	}
	if foundWords["DELL"] {
		t.Error("DELL should NOT be a letter bank match (missing N and S)")
	}
	if foundWords["SEND"] {
		t.Error("SEND should NOT be a letter bank match (missing L)")
	}
}

func TestLetterBankSolveMinWordLength(t *testing.T) {
	trie := newTrie()
	trie.addValueForString("LENDS", nil)
	trie.addValueForString("NEEDLESS", nil)

	results := PerformLetterBankSolve("LENDS", trie, 6, 100, 1, 1)

	// Only NEEDLESS (8 letters) should pass, LENDS (5 letters) should be filtered
	foundWords := make(map[string]bool)
	for _, r := range results {
		if len(r.Words) == 1 {
			foundWords[r.Words[0]] = true
		}
	}

	if foundWords["LENDS"] {
		t.Error("LENDS should be filtered out by min word length of 6")
	}
	if !foundWords["NEEDLESS"] {
		t.Error("Expected NEEDLESS to pass min word length filter")
	}
}
