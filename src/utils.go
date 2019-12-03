package core

import (
	"fmt"
	"strconv"
)

type Addr struct {
	Ip   string
	Port int
}

type Trie struct {
	Childrens map[byte]Trie
	Value     byte
	IsEnd     bool
}

func NewTrie() *Trie {
	return &Trie{
		Childrens: make(map[byte]Trie, 0),
		Value:     0,
		IsEnd:     false,
	}
}

func AddWord(trie *Trie, word string) *Trie {
	if trie == nil {
		trie = NewTrie()
	}
	trie.Value, word = word[0], word[1:]
	if len(word) == 0 {
		trie.IsEnd = true
		return trie
	}
	var newChild *Trie
	child, exist := trie.Childrens[word[0]]
	if !exist {
		newChild = AddWord(nil, word)
	} else {
		newChild = AddWord(&child, word)
	}
	trie.Childrens[word[0]] = *newChild
	return trie

}

func CheckWord(trie *Trie, word string) bool {
	if len(word) == 0 {
		return trie.IsEnd
	}

	act, rest := word[0], word[1:]
	child, exist := trie.Childrens[act]
	if !exist {
		return false
	}
	return CheckWord(&child, rest)
}

func RemoveWord(trie *Trie, word string) {
	if len(word) == 0 {
		trie.IsEnd = false
	}
	act, rest := word[0], word[1:]
	child, exist := trie.Childrens[act]
	if !exist {
		return
	}
	RemoveWord(&child, rest)
}

func GetAllWords(trie *Trie) []string {
	return getAllWords(trie, "", make([]string, 0))
}
func getAllWords(trie *Trie, prefix string, words []string) []string {
	prefix += string(trie.Value)
	fmt.Println(string(trie.Value), " ", strconv.FormatBool(trie.IsEnd), " ", trie)
	if trie.IsEnd {
		fmt.Println("ADDING")
		words = append(words, prefix)
	}

	for _, child := range trie.Childrens {
		words = getAllWords(&child, prefix, words)
	}

	return words

}
