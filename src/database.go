package core

import (
	"crypto/sha1"
	"encoding/json"
	"strconv"

	kademlia "github.com/mm-uh/go-kademlia/src"
)

type DatabaseAndPexBasedOnKademlia struct {
	kd kademlia.LocalKademlia
}

func (db *DatabaseAndPexBasedOnKademlia) Get(key string, receiver interface{}) error {
	id := kademlia.KeyNode{}
	hash := sha1.Sum([]byte(key))
	id.GetFromString(string(hash[:]))
	val, err := db.kd.GetFromNetwork(db.kd.GetContactInformation(), &id)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), receiver)
}

func (db *DatabaseAndPexBasedOnKademlia) Store(key string, value interface{}) error {
	id := kademlia.KeyNode{}
	hash := sha1.Sum([]byte(key))
	id.GetFromString(string(hash[:]))
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return db.kd.StoreOnNetwork(db.kd.GetContactInformation(), &id, string(data))
}

func (db *DatabaseAndPexBasedOnKademlia) GetLock(key string, receiver interface{}) error {
	id := kademlia.KeyNode{}
	hash := sha1.Sum([]byte(key))
	id.GetFromString(string(hash[:]))
	val, err := db.kd.GetAndLock(db.kd.GetContactInformation(), &id)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), receiver)
}

func (db *DatabaseAndPexBasedOnKademlia) StoreLock(key string, value interface{}) error {
	id := kademlia.KeyNode{}
	hash := sha1.Sum([]byte(key))
	id.GetFromString(string(hash[:]))
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return db.kd.StoreAndUnlock(db.kd.GetContactInformation(), &id, string(data))
}

func (db *DatabaseAndPexBasedOnKademlia) GetPeers() []Addr {
	nodes, err := db.kd.ClosestNodes(db.kd.GetContactInformation(), 10, db.kd.GetNodeId())
	if err != nil {
		return nil
	}
	id := kademlia.KeyNode{}
	hash := sha1.Sum([]byte("PORT"))
	id.GetFromString(string(hash[:]))

	answ := make([]Addr, 0)
	for _, node := range nodes {
		data, err := node.Get(db.kd.GetContactInformation(), &id)
		if err != nil {
			continue
		}
		port, err := strconv.ParseInt(data.Data, 10, 32)
		if err != nil {
			continue
		}
		answ = append(answ, Addr{
			ip:   node.GetIP(),
			port: int(port),
		})
	}

	return answ
}
