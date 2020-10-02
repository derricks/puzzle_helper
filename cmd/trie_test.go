package cmd

import (
	"testing"
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
