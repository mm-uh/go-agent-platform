package kademlia

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	serverRpc "github.com/mm-uh/rpc_udp/src/server"
	"github.com/sirupsen/logrus"

	avl "github.com/mm-uh/go-avl/src"
)

type LocalKademlia struct {
	ft                *kademliaFingerTable
	ip                string
	port              int
	id                *KeyNode
	k                 int
	sm                StorageManager
	a                 int
	time              uint64
	mutex             sync.Mutex
	lockedObjects     map[string]LockIdentifier
	cacheUpdatedNodes map[string]time.Time
	mutexUpdatedNodes sync.Mutex
	cacheNodeLookups  map[string]NodesWithTime
	mutexNodeLookups  sync.Mutex
}

func NewLocalKademlia(ip string, port, k int, a int) *LocalKademlia {
	key := KeyNode(sha1.Sum([]byte(fmt.Sprintf("%s:%d", ip, port))))

	nn := &LocalKademlia{
		ft:                nil,
		id:                &key,
		ip:                ip,
		port:              port,
		k:                 k,
		a:                 a,
		time:              0,
		sm:                NewSimpleKeyValueStore(),
		lockedObjects:     make(map[string]LockIdentifier, 0),
		mutex:             sync.Mutex{},
		cacheUpdatedNodes: make(map[string]time.Time, 0),
		mutexUpdatedNodes: sync.Mutex{},
		cacheNodeLookups:  make(map[string]NodesWithTime, 0),
		mutexNodeLookups:  sync.Mutex{},
	}

	kBuckets := make([]KBucket, 0)
	for i := 0; i < 160; i++ {
		kBuckets = append(kBuckets, NewKademliaKBucket(k, nn))
	}
	ft := &kademliaFingerTable{
		id:       &key,
		kbuckets: kBuckets,
	}
	nn.ft = ft
	go nn.refreshData()
	return nn
}

func (lk *LocalKademlia) refreshData() {
	for {
		time.Sleep(30 * time.Second)
		iter := lk.sm.GetAllPairs()
		for iter.Next() {
			val := iter.Value()
			key := val.GetKey()

			data := val.GetValue()
			var timeStampedData TimeStampedString = TimeStampedString{}
			err := json.Unmarshal([]byte(data), &timeStampedData)
			if err != nil {
				break
			}
			nodes, err := lk.nodeLookup(key)
			if err != nil {
				break
			}
			for _, node := range nodes {
				node.UpdateKey(lk.GetContactInformation(), key, &timeStampedData)
			}
		}
	}

}

func (lk *LocalKademlia) GetContactInformation() *ContactInformation {
	return &ContactInformation{
		node: lk,
		time: lk.time,
	}
}

func (lk *LocalKademlia) Update(node Kademlia) error {
	lk.mutexUpdatedNodes.Lock()
	defer lk.mutexUpdatedNodes.Unlock()
	now := time.Now()
	tm, exist := lk.cacheUpdatedNodes[node.GetNodeId().String()]
	if !exist {
		go lk.ft.Update(node)
		lk.cacheUpdatedNodes[node.GetNodeId().String()] = time.Now()
		return nil
	}
	if tm.Add(20 * time.Second).Before(now) {
		go lk.ft.Update(node)
		lk.cacheUpdatedNodes[node.GetNodeId().String()] = now
	}
	return nil
}

func (lk *LocalKademlia) JoinNetwork(node Kademlia) error {
	logrus.Info("Joining")
	lk.Update(node)
	lk.nodeLookup(lk.GetNodeId())
	var index int = 0
	for ; index < lk.GetNodeId().Lenght(); index++ {
		kb, err := lk.ft.GetKBucket(index)
		if err != nil {
			return err
		}
		if len(kb.GetAllNodes()) != 0 {
			break
		}
	}
	index++
	for ; index < lk.GetNodeId().Lenght(); index++ {
		key := lk.ft.GetKeyFromKBucket(index)
		lk.nodeLookup(key)
	}
	logrus.Info("Joined")
	return nil
}

func (lk *LocalKademlia) Ping(ci *ContactInformation) bool {
	if ci != nil {
		lk.time = Max(lk.time, ci.time)
		lk.Update(ci.node)
	}

	return true
}

func (lk *LocalKademlia) GetIP() string {
	return lk.ip
}

func (lk *LocalKademlia) GetPort() int {
	return lk.port
}

func (lk *LocalKademlia) GetNodeId() Key {
	return lk.id
}

func (lk *LocalKademlia) ClosestNodes(ci *ContactInformation, k int, id Key) ([]Kademlia, error) {
	if ci != nil {
		lk.time = Max(lk.time, ci.time)
		lk.Update(ci.node)
	}
	return lk.ft.GetClosestNodes(k, id)
}

func (lk *LocalKademlia) Store(ci *ContactInformation, key Key, data string) error {
	lk.time = Max(lk.time, ci.time)
	lk.Update(ci.node)
	lk.time++
	dataForSave := TimeStampedString{
		Data: data,
		Time: lk.time,
	}
	str, err := json.Marshal(&dataForSave)
	if err != nil {
		return err
	}

	return lk.sm.Store(key, string(str))
}

func (lk *LocalKademlia) Get(ci *ContactInformation, id Key) (*TimeStampedString, error) {
	if ci != nil {
		lk.time = Max(lk.time, ci.time)
		lk.Update(ci.node)
	}
	//logrus.Infof("Getting %s key", id.String())
	var tmStr TimeStampedString = TimeStampedString{}
	data, err := lk.sm.Get(id)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(data), &tmStr)
	return &tmStr, err
}

func (lk *LocalKademlia) StoreOnNetwork(ci *ContactInformation, id Key, data string) error {
	if ci != nil {
		lk.time = Max(lk.time, ci.time)
		lk.Update(ci.node)
	}

	nodes, err := lk.nodeLookup(id)
	if err != nil {
		return err
	}
	//fmt.Println(nodes)
	for _, node := range nodes {
		err = node.Store(lk.GetContactInformation(), id, data)
		//if err != nil {
		//	logrus.WithError(err).Warnf("Error storing data in %s:%d", node.GetIP(), node.GetPort())
		//}
	}
	return nil
}

func (lk *LocalKademlia) GetFromNetwork(ci *ContactInformation, id Key) (string, error) {
	if ci != nil {
		lk.time = Max(lk.time, ci.time)
		lk.Update(ci.node)
	}
	// fmt.Println("GETTING 1")
	nodes, err := lk.nodeLookup(id)
	if err != nil {
		return "", err
	}
	// fmt.Println(nodes)
	// fmt.Println("GETTING 2")
	var val string = ""
	var time uint64 = 0
	// fmt.Println("GETTING 3")
	for _, node := range nodes {
		timeStampedData, err := node.Get(lk.GetContactInformation(), id)
		if err != nil {
			// fmt.Println("GETTING 4")
			// fmt.Println(err.Error())
			logrus.WithError(err).Warnf("value from %s:%d", node.GetIP(), node.GetPort())
		} else {
			// fmt.Println("GETTING 5")
			if timeStampedData.Time >= time {
				time = timeStampedData.Time
				val = timeStampedData.Data
			}
		}
	}
	// fmt.Println("GETTING 6")
	return val, nil
}

func (lk *LocalKademlia) LockValue(ci *ContactInformation, id Key) (bool, error) {
	if ci != nil {
		lk.time = Max(lk.time, ci.time)
		lk.Update(ci.node)
	}
	lk.mutex.Lock()
	defer lk.mutex.Unlock()
	key := id.String()
	tm, exist := lk.lockedObjects[key]
	if !exist {
		lk.lockedObjects[key] = NewLockIdentifier(time.Now(), ci.node.GetNodeId().String())
		return true, nil
	}
	top := tm.Time.Add(2 * time.Minute)
	if top.Before(time.Now()) {
		lk.lockedObjects[key] = NewLockIdentifier(time.Now(), ci.node.GetNodeId().String())
		return true, nil
	}
	return false, nil

}

func (lk *LocalKademlia) UnlockValue(ci *ContactInformation, id Key) error {
	if ci != nil {
		lk.time = Max(lk.time, ci.time)
		lk.Update(ci.node)
	}

	lk.mutex.Lock()
	defer lk.mutex.Unlock()
	locker, exist := lk.lockedObjects[id.String()]
	if !exist {
		return nil
	}
	if locker.Id == ci.node.GetNodeId().String() {
		delete(lk.lockedObjects, id.String())
	}

	return nil
}

func (lk *LocalKademlia) GetLock(ci *ContactInformation, id Key) error {
	if ci != nil {
		lk.time = Max(lk.time, ci.time)
		lk.Update(ci.node)
	}

	nodes, err := lk.nodeLookup(id)
	if err != nil {
		return err
	}

	for {
		votes := make([]Kademlia, 0)
		for _, node := range nodes {
			acquire, _ := node.LockValue(lk.GetContactInformation(), id)
			if err != nil {
				//logrus.WithError(err).Warnf("Error getting lock of value with id %s in host %s:%d", id.String(), node.GetIP(), node.GetPort())
			} else if acquire {
				votes = append(votes, node)
			}
		}
		if len(votes) > (len(nodes) / 2) {
			break
		}
		for _, node := range votes {
			_ = node.UnlockValue(lk.GetContactInformation(), id)
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func (lk *LocalKademlia) LeaveLock(ci *ContactInformation, id Key) error {
	if ci != nil {
		lk.time = Max(lk.time, ci.time)
		lk.Update(ci.node)
	}

	nodes, err := lk.nodeLookup(id)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		_ = node.UnlockValue(lk.GetContactInformation(), id)
	}
	return nil
}

// func (lk *LocalKademlia) GetAndLock(ci *ContactInformation, id Key) (string, error) {
// 	if ci != nil {
// 		lk.time = Max(lk.time, ci.time)
// 		lk.ft.Update(ci.node)
// 	}

// 	nodes, err := lk.nodeLookup(id)
// 	if err != nil {
// 		return "", err
// 	}

// 	for {
// 		votes := make([]Kademlia, 0)
// 		for _, node := range nodes {
// 			acquire, err := node.LockValue(lk.GetContactInformation(), id)
// 			if err != nil {
// 				logrus.WithError(err).Warnf("Error getting lock of value with id %s in host %s:%d", id.String(), node.GetIP(), node.GetPort())
// 			} else if acquire {
// 				votes = append(votes, node)
// 			}
// 		}
// 		if len(votes) > (len(nodes) / 2) {
// 			break
// 		}
// 		for _, node := range votes {
// 			_ = node.UnlockValue(lk.GetContactInformation(), id)
// 		}
// 		time.Sleep(500 * time.Millisecond)
// 	}

// 	var val string = ""
// 	var time uint64 = 0

// 	for _, node := range nodes {
// 		timeStampedData, err := node.Get(lk.GetContactInformation(), id)
// 		if err != nil {
// 			logrus.WithError(err).Warnf("Could not get the value from %s:%d", node.GetIP(), node.GetPort())
// 		} else {
// 			if timeStampedData.Time >= time {
// 				time = timeStampedData.Time
// 				val = timeStampedData.Data
// 			}
// 		}
// 	}

// 	return val, nil

// }

func (lk *LocalKademlia) UpdateKey(ci *ContactInformation, key Key, data *TimeStampedString) error {
	if ci != nil {
		lk.time = Max(lk.time, ci.time)
		lk.Update(ci.node)
	}
	// fmt.Println("UPDATING ", key.String())
	// fmt.Println("DATA ", data.Data)
	storedData, err := lk.sm.Get(key)
	if err != nil {
		dataForStore, err := json.Marshal(data)
		if err != nil {
			return err
		}
		return lk.sm.Store(key, string(dataForStore))
	}
	var timeStamptedString TimeStampedString
	err = json.Unmarshal([]byte(storedData), &timeStamptedString)
	if err != nil {
		return err
	}
	if timeStamptedString.Time < data.Time {
		timeStamptedString.Time = data.Time
		timeStamptedString.Data = data.Data
		// fmt.Println("UPDATE ", key.String(), " WITH ", timeStamptedString.Data)
	}
	dataForStore, err := json.Marshal(timeStamptedString)
	if err != nil {
		// fmt.Println("SE JODIO")
		return err
	}
	return lk.sm.Store(key, string(dataForStore))
}

// func (lk *LocalKademlia) StoreAndUnlock(ci *ContactInformation, id Key, data string) error {
// 	if ci != nil {
// 		lk.time = Max(lk.time, ci.time)
// 		lk.ft.Update(ci.node)
// 	}

// 	nodes, err := lk.nodeLookup(id)
// 	if err != nil {
// 		return err
// 	}
// 	for _, node := range nodes {
// 		err = node.Store(lk.GetContactInformation(), id, data)
// 		if err != nil {
// 			logrus.WithError(err).Warnf("Error storing data in %s:%d", node.GetIP(), node.GetPort())
// 		}
// 	}

// 	for _, node := range nodes {
// 		_ = node.UnlockValue(lk.GetContactInformation(), id)
// 	}
// 	return nil
// }

func (lk *LocalKademlia) GetInfo() string {
	ln := KBucketNodeInformation{
		Ip:   lk.GetIP(),
		Port: lk.GetPort(),
		Key:  fmt.Sprintf("%s", lk.GetNodeId()),
	}
	kbuckets := make(map[int][]KBucketNodeInformation, 0)
	for i := 0; i < lk.GetNodeId().Lenght(); i++ {
		kb, _ := lk.ft.GetKBucket(i)
		nodes := kb.GetAllNodes()
		if nodes != nil {
			for _, node := range nodes {
				n := KBucketNodeInformation{
					Ip:   node.GetIP(),
					Port: node.GetPort(),
					Key:  fmt.Sprintf("%s", node.GetNodeId()),
				}
				kbuckets[i] = append(kbuckets[i], n)
			}
		}
	}
	info := NodeInformation{
		KBucketNodeInformation: ln,
		Kbuckets:               kbuckets,
	}

	bytes, _ := json.Marshal(info)

	return string(bytes)
}

func (lk *LocalKademlia) nodeLookup(id Key) ([]Kademlia, error) {
	var round int = 1
	// fmt.Println("LOOKUP ", id.String())
	//Create structure to keep ordered nodes
	//startNodes, err := lk.ft.GetClosestNodes(lk.a, id)
	//if err != nil {
	//	return nil, err
	//}
	//if len(startNodes) == 0 {
	//	return nil, nil
	//}

	var candidates *avl.Node = nil
	avlLock := new(sync.Mutex)
	var nodesAnsw *avl.Node = nil
	mu := sync.Mutex{}
	cond := sync.NewCond(&mu)

	//myDist, err := lk.GetNodeId().XOR(id)
	//if err != nil {
	//	return nil, err
	//}
	//nodesAnsw = avl.Insert(nodesAnsw, avl.NewNode(myDist, lk))
	//for _, node := range startNodes {
	//	dist, err := node.GetNodeId().XOR(id)
	//	if err != nil {
	//		return nil, err
	//	}
	//	candidates = avl.Insert(candidates, avl.NewNode(dist, node))
	//}

	myDist, err := lk.GetNodeId().XOR(id)
	if err != nil {
		return nil, err
	}
	candidates = avl.Insert(candidates, avl.NewNode(myDist, lk))
	//Nodes, startNodes := avl.NewNode(dist, newNodeLookup(startNodes[0])), startNodes[1:]
	//node, _ := Nodes.Value.(nodeLookup)

	//node.queried = true
	nextRoundMain := make(chan bool)
	nextRoundReceiver := make(chan bool)
	allNodesComplete := make(chan int)
	receivFromWorkers := make(chan nodesPackage)
	receivFromStorage := make(chan nodesPackage)
	endStorage := make(chan bool)
	endGuard := make(chan bool)

	go startRoundGuard(nextRoundMain, nextRoundReceiver, endGuard, allNodesComplete)
	go replyReceiver(receivFromStorage, receivFromWorkers, nextRoundReceiver, endStorage, allNodesComplete, lk.a)

	for {
		if candidates.GetSize() == 0 {
			answ := make([]Kademlia, 0)
			(*avlLock).Lock()
			for _, node := range nodesAnsw.GetKMins(lk.k) {
				n, ok := node.Value.(Kademlia)
				if !ok {
					panic("Incorrect type")
				}
				answ = append(answ, n)
			}
			(*avlLock).Unlock()
			endGuard <- true
			endStorage <- true
			cond.L.Lock()
			cond.Broadcast()
			cond.L.Unlock()
			// fmt.Println("2 ", printRemote(answ))
			return answ, nil
		}
		(*avlLock).Lock()
		size := nodesAnsw.GetSize()
		(*avlLock).Unlock()
		if size >= lk.k {
			candidatesMinTemp := candidates.GetKMins(1)[0]
			nodesAnswMaxTemp := nodesAnsw.GetMax()
			candidatesMin, ok := candidatesMinTemp.Value.(Kademlia)
			if !ok {
				panic("Incorrect type")
			}
			nodesAnswMax, ok := nodesAnswMaxTemp.Value.(Kademlia)
			if !ok {
				panic("Incorrect type")
			}

			dist1, err := candidatesMin.GetNodeId().XOR(id)
			if err != nil {
				return nil, err
			}
			dist2, err := nodesAnswMax.GetNodeId().XOR(id)
			if err != nil {
				return nil, err
			}
			less, err := dist1.Less(dist2)
			if err != nil {
				return nil, err
			}

			if !less {
				answ := make([]Kademlia, 0)
				(*avlLock).Lock()
				for _, node := range nodesAnsw.GetKMins(lk.k) {
					n, ok := node.Value.(Kademlia)
					if !ok {
						panic("Incorrect type")
					}
					answ = append(answ, n)
				}
				(*avlLock).Unlock()
				endGuard <- true
				endStorage <- true
				cond.L.Lock()
				cond.Broadcast()
				cond.L.Unlock()
				// fmt.Println("1 ", printRemote(answ))
				return answ, nil
			}
		}
		top := Min(lk.a, candidates.GetSize())
		for i := 0; i < top; i++ {
			n := candidates.GetKMins(1)[0]
			nk, ok := n.Value.(Kademlia)
			if !ok {
				return nil, err
			}
			dist, err := nk.GetNodeId().XOR(id)
			if err != nil {
				return nil, err
			}
			candidates = avl.Delete(candidates, dist)
			update := func(nq Kademlia) error {
				dist, err := nq.GetNodeId().XOR(id)
				if err != nil {
					return err
				}
				(*avlLock).Lock()
				nodesAnsw = avl.Insert(nodesAnsw, avl.NewNode(dist, nq))
				(*avlLock).Unlock()
				return lk.Update(nq)
			}
			go queryNode(nk, id, round, lk.k, lk.GetContactInformation(), receivFromWorkers, cond, update)
		}
		_ = <-nextRoundMain
		round++
		np := <-receivFromStorage
		nodes := np.receivNodes
		for node := range nodes {
			//add nodes to ordered structure
			dist, err := node.GetNodeId().XOR(id)
			if err != nil {
				return nil, err
			}
			(*avlLock).Lock()
			exist := nodesAnsw.HasKey(dist)
			(*avlLock).Unlock()
			if !exist {
				newNode := avl.NewNode(dist, node)
				candidates = avl.Insert(candidates, newNode)
			}

		}

	}
}

func (lk *LocalKademlia) RunServer(exited chan bool) {
	var h HandlerRPC
	h.kademlia = lk

	server := serverRpc.NewServer(h, lk.ip+":"+strconv.FormatInt(int64(lk.port), 10))

	// listen to incoming udp packets
	go server.ListenServer(exited)

}

func startRoundGuard(nextRoundMain, nextRoundReceiver, lookupEnd chan bool, allNodesComplete chan int) {
	var actRound int = 1
	for {
		timeout := make(chan int)
		go func(roundN int) { //timeout goroutine
			time.Sleep(1 * time.Second)
			timeout <- actRound
		}(actRound)

		select {
		case timeRound := <-timeout:
			{
				if timeRound != actRound {
					// fmt.Println("ROUND TIMEOUT")
					continue
				}
			}
		case round := <-allNodesComplete:
			{
				if round == actRound {
					break
				}
			}
		case _ = <-lookupEnd:
			{
				return
			}
		}

		actRound++
		nextRoundMain <- true
		nextRoundReceiver <- true
	}

}

func queryNode(node Kademlia, id Key, round, k int, ci *ContactInformation, send chan nodesPackage, finish *sync.Cond, update func(Kademlia) error) {
	lookupFinished := make(chan bool)
	go func() {
		finish.L.Lock()
		finish.Wait()
		finish.L.Unlock()
		lookupFinished <- true
	}()
	// fmt.Printf("ASKING TO %s\n", printRemote([]Kademlia{node}))
	nodes, err := node.ClosestNodes(ci, k, id)
	if err == nil {
		update(node)
		// fmt.Printf("RESPONSE FROM: %s\n", printRemote([]Kademlia{node}))
	}

	channel := make(chan Kademlia)
	np := nodesPackage{
		round:       round,
		receivNodes: channel,
	}
	select {
	case send <- np:
		{
			for _, n := range nodes {
				channel <- n
			}
			close(channel)
		}
	case _ = <-lookupFinished:
		{
			return
		}
	}

}

func replyReceiver(sendToMain, receivFromWorkers chan nodesPackage, nextRound, lookupEnd chan bool, allNodesComplete chan int, a int) {
	var actRound int = 1
	var finished int = 0
	var nodesForSend []Kademlia

	for {

		select {
		case np := <-receivFromWorkers:
			{
				channel := np.receivNodes
				for node := range channel {
					nodesForSend = append(nodesForSend, node)
				}

				if np.round == actRound {
					finished++
				}
				if finished == a {
					allNodesComplete <- actRound
				}
			}

		case _ = <-nextRound:
			{
				actRound++
				channel := make(chan Kademlia)
				np := nodesPackage{
					round:       actRound,
					receivNodes: channel,
				}
				sendToMain <- np
				for _, node := range nodesForSend {
					channel <- node
				}
				close(channel)
				nodesForSend = make([]Kademlia, 0)
				finished = 0

			}

		case _ = <-lookupEnd:
			{
				return
			}
		}

	}
}

type nodesPackage struct {
	round       int
	receivNodes chan Kademlia
}

type dataSaved struct {
	time uint64
	data interface{}
}
