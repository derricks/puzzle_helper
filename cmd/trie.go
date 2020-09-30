package cmd

import (
	"fmt"
)

// Implements a basic trie system, which ends up being by used by a number of word puzzles

type trieNode struct {
	letter   string
	children map[string]*trieNode
}

func newTrie() *trieNode {
	return newTrieWithLetter("")
}

func newTrieWithLetter(letter string) *trieNode {
	return &trieNode{letter, make(map[string]*trieNode)}
}

func (node *trieNode) addString(input string) {
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

func (node *trieNode) String() string {
	return fmt.Sprintf("%s: [%v]", node.letter, node.children)
}
