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
