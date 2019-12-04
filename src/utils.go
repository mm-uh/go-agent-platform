package core

import (
	"bufio"
	"fmt"
	"net"
	"time"
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
		Childrens: make(map[byte]Trie),
		Value:     0,
		IsEnd:     false,
	}
}

func AddWord(trie *Trie, word string) *Trie {
	if trie == nil {
		trie = NewTrie()
	}
	if len(word) == 0 {
		trie.IsEnd = true
		return trie
	}
	var newChild *Trie
	child, exist := trie.Childrens[word[0]]
	if !exist {
		newChild = AddWord(nil, word[1:])
		newChild.Value = word[0]
	} else {
		newChild = AddWord(&child, word[1:])
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

func RemoveWord(trie *Trie, word string) *Trie {
	if len(word) == 0 {
		trie.IsEnd = false
		return trie
	}

	child, exist := trie.Childrens[word[0]]
	if exist {
		newChild := RemoveWord(&child, word[1:])
		trie.Childrens[word[0]] = *newChild
	}

	return trie
}

func GetAllWords(trie *Trie) []string {
	return getAllWords(trie, "", make([]string, 0))
}

func getAllWords(trie *Trie, prefix string, words []string) []string {
	prefix += string(trie.Value)
	if trie.IsEnd {
		words = append(words, prefix[1:])
	}

	for _, child := range trie.Childrens {
		words = getAllWords(&child, prefix, words)
	}

	return words

}

func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// Check if agent is available
// Send over a tcp connection a message 'Alive?\n'
// Wait 5 seconds for response, that should be 'Yes\n'
func MakeRequest(endpoint, request string) (string, error) {

	conn, err := net.Dial("tcp", endpoint)
	if err != nil {
		return "", err
	}
	err = conn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return "", err
	}

	// send to socket
	_, err = fmt.Fprintf(conn, request+"\n")
	if err != nil {
		return "", err
	}

	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	return message, nil
}
