package cmd

import (
	"testing"
	"time"
)

func TestAdds(test *testing.T) {
	trie := newTrie()
	trie.addString("hello")

	childTrie, hPresent := trie.children["h"]
	if !hPresent {
		test.Errorf("h should have been present as a key but was not")
	}

	if _, ePresent := childTrie.children["e"]; !ePresent {
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
		addRetrieveTest{"thirsty", 123, true},
		addRetrieveTest{"thi", nil, true},
		addRetrieveTest{"this", nil, false},
	}
	for index, testCase := range tests {
		trie := newTrie()
		if testCase.shouldBePresent {
			trie.addStringWithValue(testCase.input, testCase.value)
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

func TestIterateWords(test *testing.T) {
	tests := map[string]int{
		"stringing": 123,
		"string":    456,
	}

	trie := newTrie()
	for testWord, testValue := range tests {
		trie.addStringWithValue(testWord, testValue)
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
	close(words)
}
