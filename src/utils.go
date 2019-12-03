package core

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

func AddWord(trie *Trie, word string) {
	if len(word) == 0 {
		trie.IsEnd = true
		return
	}
	act, rest := word[0], word[1:]
	child, exist := trie.Childrens[act]
	if !exist {
		newTrie := Trie{
			Childrens: make(map[byte]Trie),
			Value:     act,
			IsEnd:     false,
		}
		trie.Childrens[act] = newTrie
		child = newTrie
	}
	AddWord(&child, rest)
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

func GetAllWords(trie *Trie, prefix string, words []string) []string {
	prefix = append(prefix, trie.Value)
	if trie.IsEnd {
		words = append(words, prefix)
	}

	for _, child := range trie.Childrens {
		words = GetAllWords(&child, prefix, words)
	}

	return words

}
