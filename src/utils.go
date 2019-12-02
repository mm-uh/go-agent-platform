package core

type Addr struct {
	ip   string
	port int
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

func AddWord(trie *Trie, word string) {
	if len(word) == 0 {
		trie.IsEnd = true
		return
	}
	act, rest := word[0], word[1:]
	child, exist := trie.Childrens[act]
	if !exist {
		newTrie := Trie{
			Childrens: make(map[byte]Trie, 0),
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
