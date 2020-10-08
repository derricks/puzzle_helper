package cmd

import (
	"fmt"
)

// Implements a basic trie system, which ends up being by used by a number of word puzzles

type trieNode struct {
	letter   string
	value    interface{}
	children map[string]*trieNode
}

func newTrie() *trieNode {
	return newTrieWithLetter("")
}

func newTrieWithLetter(letter string) *trieNode {
	return &trieNode{letter, nil, make(map[string]*trieNode)}
}

func (node *trieNode) addString(input string) {

	// if the value is there already, you're done
	if input == "" {
		// mark a word boundary
		node.children[""] = newTrie()
		return
	}

	head := input[0:1]
	tail := input[1:]

	childNode, present := node.children[head]
	if !present {
		// make a new Node, insert it into the existing node. and recurse
		childNode = newTrieWithLetter(head)
		node.children[head] = childNode
	}
	childNode.addString(tail)
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

func (node *trieNode) addStringWithValue(input string, value interface{}) {
	node.addString(input)
	node.setValueForString(input, value)
}

// setValueForString puts the given value at the correct spot in the trie. It assumes
// the string is already in the trie! if a string is not in the trie, use addStringWithValue
func (node *trieNode) setValueForString(input string, value interface{}) {
	if input == "" {
		node.value = value
		return
	}
	head := input[0:1]
	node.children[head].setValueForString(input[1:], value)
}

// getValueForString retrieves the value set for the string. It does not assume
// the string is in the trie; it will return nil, false if the string wasn't there
func (node *trieNode) getValueForString(input string) (interface{}, bool) {
	if input == "" {
		return node.value, true
	}

	head := input[0:1]
	tail := input[1:]

	childNode, isPresent := node.children[head]
	if !isPresent {
		// can't continue this string in the trie
		return nil, false
	}

	return childNode.getValueForString(tail)
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
	for letter, child := range node.children {
		if letter == "" {
			channel <- trieWord{currentWord, node.value}
			continue
		}

		child.recursiveFindWords(currentWord+letter, channel)
	}
}

func (node *trieNode) String() string {
	return fmt.Sprintf("%s (%s): [%v]", node.letter, node.value, node.children)
}
