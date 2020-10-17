package cmd

import (
	"fmt"
	"regexp"
	"strings"
)

// Implements a basic trie system, which ends up being by used by a number of word puzzles
// this trie only accepts upper case, alphabetic strings

const ASCII_A = 65

type trieNode struct {
	letter         string
	atWordBoundary bool
	value          interface{}
	// each node's children is just a slice of childNodes. the position of each childNode represents its letter
	// i.e., A= 0 and so on
	children []*trieNode
}

func newTrie() *trieNode {
	return newTrieWithLetter("")
}

func newTrieWithLetter(letter string) *trieNode {
	return &trieNode{letter, false, nil, make([]*trieNode, 26, 26)}
}

var allUppercase = regexp.MustCompile("^[A-Z]+$")

func (node *trieNode) addValueForString(input string, value interface{}) error {

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

// getSize returns the number of items in the trie
func (node *trieNode) getSize() int {
	size := 0
	wordChannel := make(chan trieWord)
	go node.feedWordsToChannel(wordChannel)
	for _ = range wordChannel {
		size++
	}
	return size
}

// getValueForString retrieves the value set for the string. It does not assume
// the string is in the trie; it will return nil, false if the string wasn't there
func (node *trieNode) getValueForString(input string) (interface{}, bool) {
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

type trieWord struct {
	word  string
	value interface{}
}

func (node *trieNode) feedWordsToChannel(channel chan trieWord) {
	node.recursiveFindWords("", channel)
	close(channel)
}

func (node *trieNode) recursiveFindWords(currentWord string, channel chan trieWord) {
	if node.atWordBoundary {
		channel <- trieWord{currentWord, node.value}
	}

	for index, currentNode := range node.children {
		if currentNode != nil {
			currentNode.recursiveFindWords(currentWord+(string(index+ASCII_A)), channel)
		}
	}
}

func (node *trieNode) String() string {
	return fmt.Sprintf("%s (%s): [%v]", node.letter, node.value, node.children)
}
