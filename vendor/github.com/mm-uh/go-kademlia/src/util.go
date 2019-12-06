package kademlia

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

func XOR(a uint64, b uint64) uint64 {
	return a ^ b
}

//func log(a uint64) int {
//	return 1
//}

func Pow(a int, b int) int {
	if b == 0 {
		return 1
	}
	if b == 2 {
		return a * a
	}
	halfPow := Pow(a, b/2)
	if b%2 == 1 {
		return Pow(halfPow, 2) * a
	}
	return Pow(halfPow, 2)
}

func Min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a uint64, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

type StorageManager interface {
	Store(Key, string) error
	Get(Key) (string, error)
	GetAllPairs() KeyValueIterator
}

type KeyValueIterator interface {
	Next() bool
	Value() KeyValue
	HasNext() bool
}

type KeyValue interface {
	GetKey() Key
	GetValue() string
}

type SimpleKeyValue struct {
	key   *KeyNode
	value string
}

func (skv *SimpleKeyValue) GetKey() Key {
	return skv.key
}

func (skv *SimpleKeyValue) GetValue() string {
	return skv.value
}

func NewSimpleKeyValue(key *KeyNode, value string) *SimpleKeyValue {
	return &SimpleKeyValue{
		key:   key,
		value: value,
	}
}

type MyKeyValueIterator struct {
	data    []*SimpleKeyValue
	current int
}

func (kvi *MyKeyValueIterator) HasNext() bool {
	return kvi.current+1 < len(kvi.data)
}

func (kvi *MyKeyValueIterator) Next() bool {
	if kvi.HasNext() {
		kvi.current++
		return true
	}
	return false
}

func (kvi *MyKeyValueIterator) Value() KeyValue {
	return kvi.data[kvi.current]
}

func NewMyKeyValueIterator(data []*SimpleKeyValue) *MyKeyValueIterator {
	return &MyKeyValueIterator{
		data:    data,
		current: -1,
	}
}

type ContactInformation struct {
	node Kademlia
	time uint64
}

type NodeInformation struct {
	KBucketNodeInformation
	Kbuckets map[int][]KBucketNodeInformation `json:"KBuckets"`
}

type KBucketNodeInformation struct {
	Key  string `json:"Key"`
	Ip   string `json:"Ip"`
	Port int    `json:"Port"`
}

type SimpleKeyValueStore struct {
	data  map[string]string
	mutex sync.Mutex
}

func NewSimpleKeyValueStore() *SimpleKeyValueStore {
	return &SimpleKeyValueStore{
		data:  make(map[string]string),
		mutex: sync.Mutex{},
	}
}

func (kv *SimpleKeyValueStore) Store(id Key, data string) error {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	kv.data[id.String()] = data
	return nil
}

func (kv *SimpleKeyValueStore) Get(id Key) (string, error) {
	kv.mutex.Lock()
	defer kv.mutex.Unlock()
	val, ok := kv.data[id.String()]
	if !ok {
		return "", errors.New("There is no value for that key")
	}
	return val, nil
}

func (kv *SimpleKeyValueStore) GetAllPairs() KeyValueIterator {
	data := make([]*SimpleKeyValue, 0)
	for key, value := range kv.data {
		myKey := KeyNode{}
		myKey.GetFromString(key)
		nkv := NewSimpleKeyValue(&myKey, value)
		data = append(data, nkv)
	}
	return NewMyKeyValueIterator(data)
}

type LockIdentifier struct {
	Time time.Time
	Id   string
}

func NewLockIdentifier(time time.Time, id string) LockIdentifier {
	return LockIdentifier{
		Id:   id,
		Time: time,
	}
}

func printRemote(nodes []Kademlia) string {
	answ := ""
	for _, n := range nodes {
		answ = fmt.Sprintf("%s %s,", answ, n.GetIP()[7:])
	}
	return answ
}

type NodesWithTime struct {
	nodes []Kademlia
	time  time.Time
}
