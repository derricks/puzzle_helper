package cmd

import (
	"fmt"
	"regexp"
	"strings"
)

// Implements a basic trie system, which ends up being by used by a number of word puzzles
// this trie only accepts upper case, alphabetic strings

const ASCII_A = 65

type TrieNode struct {
	letter         string
	atWordBoundary bool
	value          interface{}
	// each node's children is just a slice of childNodes. the position of each childNode represents its letter
	// i.e., A= 0 and so on
	children [27]*TrieNode
}

func newTrie() *TrieNode {
	return newTrieWithLetter("")
}

func newTrieWithLetter(letter string) *TrieNode {
	var children [27]*TrieNode
	trie := &TrieNode{letter, false, nil, children}
	// a special character at the end so that transposals can check if they're at a word boundary _and_ traverse the children
	trie.children[26] = &TrieNode{"", false, nil, children}
	return trie
}

var allUppercase = regexp.MustCompile("^[A-Z]+$")

func (node *TrieNode) addValueForString(input string, value interface{}) error {

	if !allUppercase.MatchString(input) {
		return fmt.Errorf("This trie only accepts upper case. String %s is invalid", input)
	}

	curChild := node
	for _, curLetter := range strings.Split(input, "") {
		// because the trie only accepts uppercase letters, we can just subtract 65 to find the index
		childIndex := []byte(curLetter)[0] - ASCII_A
		nextChild := curChild.children[childIndex]
		if nextChild == nil {
			nextChild = newTrieWithLetter(curLetter)
			curChild.children[childIndex] = nextChild
		}
		curChild = nextChild
	}
	curChild.atWordBoundary = true
	curChild.value = value
	return nil
}

// GetSize returns the number of items in the trie
func (node *TrieNode) GetSize() int {
	size := 0
	wordChannel := make(chan TrieWord) // Renamed trieWord to TrieWord
	go node.FeedWordsToChannel(wordChannel) // Renamed feedWordsToChannel to FeedWordsToChannel
	for _ = range wordChannel {
		size++
	}
	return size
}

// GetValueForString retrieves the value set for the string. It does not assume
// the string is in the trie; it will return nil, false if the string wasn't there
func (node *TrieNode) GetValueForString(input string) (interface{}, bool) {
	currentNode := node

	for _, curChar := range strings.Split(input, "") {
		childIndex := []byte(curChar)[0] - ASCII_A
		nextNode := currentNode.children[childIndex]
		if nextNode == nil {
			return nil, false
		}
		currentNode = nextNode
	}
	// you could be at the end of a requested key but not actually at a word boundary
	if currentNode.atWordBoundary {
		return currentNode.value, true
	} else {
		return nil, false
	}
}

type TrieWord struct { // Renamed trieWord to TrieWord
	Word  string // Renamed word to Word
	Value interface{}
}

func (node *TrieNode) FeedWordsToChannel(channel chan TrieWord) { // Renamed feedWordsToChannel to FeedWordsToChannel
	node.recursiveFindWords("", channel)
	close(channel)
}

func (node *TrieNode) recursiveFindWords(currentWord string, channel chan TrieWord) { // Renamed trieWord to TrieWord
	if node.atWordBoundary {
		channel <- TrieWord{currentWord, node.value} // Renamed trieWord to TrieWord
	}

	for index, currentNode := range node.children {
		if currentNode != nil {
			currentNode.recursiveFindWords(currentWord+(string(index+ASCII_A)), channel)
		}
	}
}

func (node *TrieNode) String() string {
	return fmt.Sprintf("%s (%s): [%v]", node.letter, node.value, node.children)
}
