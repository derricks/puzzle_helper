package cmd

import (
	"testing"
	"time"
)

func TestAdds(test *testing.T) {
	trie := newTrie()

	err := trie.addValueForString("hello", nil)
	if err == nil {
		test.Errorf("Trie should have rejected 'hello'")
	}

	trie.addValueForString("HELLO", nil)

	childTrie := trie.children['H'-ASCII_A]
	if childTrie == nil {
		test.Errorf("H should have been present as a key but was not")
	}

	childTrie = childTrie.children['E'-ASCII_A]
	if childTrie == nil {
		test.Errorf("e should have been present within h but was not")
	}
}

type addRetrieveTest struct {
	input           string
	value           interface{}
	shouldBePresent bool
}

func TestAddingRetrieving(test *testing.T) {
	tests := []addRetrieveTest{
		addRetrieveTest{"THIRSTY", 123, true},
		addRetrieveTest{"THI", nil, true},
		addRetrieveTest{"THIS", nil, false},
	}
	for index, testCase := range tests {
		trie := newTrie()
		if testCase.shouldBePresent {
			trie.addValueForString(testCase.input, testCase.value)
		}

		value, stringWasPresent := trie.getValueForString(testCase.input)
		if stringWasPresent != testCase.shouldBePresent {
			test.Errorf("Test case %d: expected %v for string's presence, got %v", index, testCase.shouldBePresent, stringWasPresent)
		}

		if value != testCase.value {
			test.Errorf("Test case %d: Expected value of %v but got %v", index, testCase.value, value)
		}
	}
}

func TestGetSize(test *testing.T) {
	trie := newTrie()
	trie.addValueForString("HELLO", nil)
	trie.addValueForString("HELL", nil)
	trie.addValueForString("HE", nil)
	trie.addValueForString("GOODBYE", nil)
	actualSize := trie.getSize()
	if actualSize != 4 {
		test.Errorf("Expected trie size of 4 but got %d", actualSize)
	}
}

func TestIterateWords(test *testing.T) {
	tests := map[string]int{
		"STRINGING": 123,
		"STRING":    456,
	}

	trie := newTrie()
	for testWord, testValue := range tests {
		trie.addValueForString(testWord, testValue)
	}

	words := make(chan trieWord)
	timer := time.NewTimer(1 * time.Second)

	go trie.feedWordsToChannel(words)
	select {
	case foundTrieWord := <-words:
		testCount, wasPresent := tests[foundTrieWord.word]
		if !wasPresent {
			test.Errorf("Channel put out a word that's not in test case: %s", foundTrieWord.word)
		}

		if testCount != foundTrieWord.value {
			test.Errorf("Expected count of %d for %s but got %d", testCount, foundTrieWord.word, foundTrieWord.value)
		}
		delete(tests, foundTrieWord.word)
	case _ = <-timer.C:
		if len(tests) != 0 {
			test.Errorf("Tests should be empty but had %d items in it", len(tests))
		}
		break
	}
}
