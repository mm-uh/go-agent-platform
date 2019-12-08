package core

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	kademlia "github.com/mm-uh/go-kademlia/src"
)

type DatabaseAndPexBasedOnKademlia struct {
	Kd *kademlia.LocalKademlia
}

func (db *DatabaseAndPexBasedOnKademlia) Get(key string, receiver interface{}) error {
	id := kademlia.KeyNode{}
	hash := sha1.Sum([]byte(key))
	err := id.GetFromString(hex.EncodeToString(hash[:]))
	if err != nil {
		return err
	}
	val, err := db.Kd.GetFromNetwork(db.Kd.GetContactInformation(), &id)
	if err != nil {
		return err
	}
	if val == "" && key[:Min(9, len(key))] == Function+":" {
		receiver = []string{}//nolint
		return nil
	}
	return json.Unmarshal([]byte(val), receiver)
}

func (db *DatabaseAndPexBasedOnKademlia) Store(key string, value interface{}) error {
	id := kademlia.KeyNode{}
	hash := sha1.Sum([]byte(key))
	err := id.GetFromString(hex.EncodeToString(hash[:]))
	if err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return db.Kd.StoreOnNetwork(db.Kd.GetContactInformation(), &id, string(data))
}

func (db *DatabaseAndPexBasedOnKademlia) Lock(key string) error {
	id := kademlia.KeyNode{}
	hash := sha1.Sum([]byte(key))
	err := id.GetFromString(hex.EncodeToString(hash[:]))
	if err != nil {
		return err
	}
	return db.Kd.GetLock(db.Kd.GetContactInformation(), &id)

}

func (db *DatabaseAndPexBasedOnKademlia) Unlock(key string) error {
	id := kademlia.KeyNode{}
	hash := sha1.Sum([]byte(key))
	err := id.GetFromString(hex.EncodeToString(hash[:]))
	if err != nil {
		return err
	}
	return db.Kd.LeaveLock(db.Kd.GetContactInformation(), &id)
}

func (db *DatabaseAndPexBasedOnKademlia) GetPeers() []Addr {
	nodes, err := db.Kd.ClosestNodes(db.Kd.GetContactInformation(), 10, db.Kd.GetNodeId())
	if err != nil {
		return nil
	}
	id := kademlia.KeyNode{}
	hash := sha1.Sum([]byte("PORT"))
	err = id.GetFromString(hex.EncodeToString(hash[:]))
	if err != nil {
		return nil
	}

	answ := make([]Addr, 0)
	fmt.Println("NODES ", nodes)
	for _, node := range nodes {
		data, err := node.Get(db.Kd.GetContactInformation(), &id)
		if err != nil {
			continue
		}
		port, err := strconv.ParseInt(data.Data, 10, 32)
		if err != nil {
			continue
		}
		answ = append(answ, Addr{
			Ip:   node.GetIP(),
			Port: int(port),
		})
	}

	return answ
}
