package kademlia

import (
	"sort"
	"sync"
)

type kademliaKBucket struct {
	start *linkedList
	last  *linkedList
	k     int
	lk    *LocalKademlia
	mutex sync.Mutex
}

func (kB *kademliaKBucket) Update(c Kademlia) {
	kB.mutex.Lock()
	defer kB.mutex.Unlock()
	nn := newLinkedListNode(c)
	if kB.start == nil {
		kB.start, kB.last = nn, nn
		return
	}
	// if the contact is already in the kBucket
	first := kB.start
	var prev *linkedList = nil
	for true {
		equal, _ := first.value.GetNodeId().Equal(c.GetNodeId())
		if equal {
			kB.last.next = first
			kB.last = first
			if prev == nil {
				kB.start = first.next
			} else {
				prev.next = first.next
			}
			first.next = nil
			return
		}
		if first == kB.last {
			break
		}
		prev = first
		first = first.next
	}
	// if the kBucket is not full
	if kB.start.len() < kB.k {
		kB.last.next = nn
		kB.last = nn
		return
	}
	//if the kBucket is full
	head := kB.start
	if head.value.Ping(kB.lk.GetContactInformation()) {
		kB.start = head.next
		head.next = nil
		kB.last.next = head
		kB.last = head
		return
	}
	kB.start = head.next
	kB.last.next = nn
	kB.last = nn
	return
}

func (kB *kademliaKBucket) GetClosestNodes(k int, nodeId Key) ([]Kademlia, error) {
	//kB.mutex.Lock()
	//defer kB.mutex.Unlock()
	if kB.start == nil {
		return nil, nil
	}

	unorderedScl, err := sortableContactListFromLinkedList(kB.start, nodeId)
	if err != nil {
		return nil, err
	}
	sort.Sort(unorderedScl)
	scl := (*unorderedScl)[:Min(k, len(*unorderedScl))]
	contacts := make([]Kademlia, 0)
	for _, cd := range scl {
		contacts = append(contacts, cd.c)
	}

	return contacts, nil
}

func (kB *kademliaKBucket) GetAllNodes() []Kademlia {
	// kB.mutex.Lock()
	// defer kB.mutex.Unlock()
	start := kB.start
	if start == nil {
		return nil
	}
	result := make([]Kademlia, 0)
	for start != kB.last {
		result = append(result, start.value)
		start = start.next
	}
	result = append(result, start.value)
	return result
}

func NewKademliaKBucket(n int, k *LocalKademlia) *kademliaKBucket {
	return &kademliaKBucket{
		k:     n,
		mutex: sync.Mutex{},
		lk:    k,
	}
}

type distanceToContact struct {
	distance Key
	c        Kademlia
}

func sortableContactListFromLinkedList(start *linkedList, nodeId Key) (*sortableContactList, error) {
	scl := new(sortableContactList)
	for true {
		if start == nil {
			break
		}

		dist, err := nodeId.XOR(start.value.GetNodeId())
		if err != nil {
			return nil, err
		}
		sclv := sclAppend(scl, &distanceToContact{
			c:        start.value,
			distance: dist,
		})
		scl = &sclv
		start = start.next
	}
	return scl, nil
}

type sortableContactList []*distanceToContact

func sclAppend(scl *sortableContactList, dc *distanceToContact) sortableContactList {
	return append(*scl, dc)
}

func (scl *sortableContactList) Len() int {
	return len(*scl)
}

func (scl *sortableContactList) Swap(i, j int) {
	(*scl)[i], (*scl)[j] = (*scl)[j], (*scl)[i]
}

func (scl *sortableContactList) Less(i, j int) bool {
	//ToDo Handel Error
	val, _ := (*scl)[i].distance.Less((*scl)[j].distance)
	return val
}

func newLinkedListNode(c Kademlia) *linkedList {
	return &linkedList{
		next:  nil,
		value: c,
	}
}

type linkedList struct {
	next  *linkedList
	value Kademlia
}

func (n *linkedList) len() int {
	if n.next == nil {
		return 1
	}
	return n.next.len() + 1
}
