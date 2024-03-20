package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	
	"sync"

	"github.com/eiannone/keyboard"
)

// TrieNode represents a node in the trie data structure.
type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
	meaning  string 
}

// Trie represents the trie data structure.
type Trie struct {
	root *TrieNode
	mu   sync.Mutex
}

// NewTrie creates a new Trie.
func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			children: make(map[rune]*TrieNode),
		},
	}
}

// Insert inserts a word into the trie.
func (t *Trie) Insert(word, meaning string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	node := t.root
	for _, char := range word {
		if node.children[char] == nil {
			node.children[char] = &TrieNode{
				children: make(map[rune]*TrieNode),
			}
		}
		node = node.children[char]
	}
	node.isEnd = true
	node.meaning = meaning
}

// AutoSuggest returns a list of suggestions based on the given prefix.
func (t *Trie) AutoSuggest(prefix string) ([]string, []string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	node := t.root
	for _, char := range prefix {
		if node.children[char] == nil {
			return nil, nil
		}
		node = node.children[char]
	}

	var suggestions []string
	var meanings []string
	t.autoSuggestRecursive(node, prefix, &suggestions, &meanings)
	return suggestions, meanings
}

// helper function for AutoSuggest.
func (t *Trie) autoSuggestRecursive(node *TrieNode, currentPrefix string, suggestions *[]string, meanings *[]string) {
	if node.isEnd {
		*suggestions = append(*suggestions, currentPrefix)
		*meanings = append(*meanings, node.meaning)
	}

	for char, childNode := range node.children {
		t.autoSuggestRecursive(childNode, currentPrefix+string(char), suggestions, meanings)
	}
}

func loadDictionaryFromFile(filePath string) (map[string]string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var dictionary map[string]string
	err = json.Unmarshal(data, &dictionary)
	if err != nil {
		return nil, err
	}

	return dictionary, nil
}

func main() {
	// Load dictionary from JSON file
	dictionaryFilePath := "dictionary.json"
	dictionary, err := loadDictionaryFromFile(dictionaryFilePath)
	if err != nil {
		panic(err)
	}

	trie := NewTrie()

	// Populate trie with dictionary words and meanings
	for word, meaning := range dictionary {
		trie.Insert(word, meaning)
	}

	// Start listening for keyboard events
	err = keyboard.Open()
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()

	var currentInput string

	fmt.Println("Type a word (press 'Enter' to exit):")

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}

		if key == keyboard.KeyEnter {
			break
		}

		if key == keyboard.KeySpace {
			currentInput += " "
		} else if key == keyboard.KeyBackspace {
			if len(currentInput) > 0 {
				currentInput = currentInput[:len(currentInput)-1]
			}
		} else if char != 0 {
			currentInput += string(char)
		}

		// Clear the current line and print the current input
		fmt.Print("\033[K") // ANSI escape code to clear the line
		fmt.Printf("\r%s", currentInput)

		// Getting suggestions with meanings
		suggestions, meanings := trie.AutoSuggest(currentInput)

		// Sorting suggestions alphabetically
		sort.Strings(suggestions)

		// Display the top two suggestions with meanings
		if len(suggestions) > 0 {
			fmt.Printf("\nTop Suggestions:")
			for i := 0; i < 2 && i < len(suggestions); i++ {
				fmt.Printf("\n- %s: %s", suggestions[i], meanings[i])
			}
		}
	}
}
