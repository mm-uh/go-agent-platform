package kademlia

import "errors"

type kademliaFingerTable struct {
	kbuckets []KBucket
	id       *KeyNode
}

func (ft *kademliaFingerTable) GetKBucket(index int) (KBucket, error) {
	if index < 0 || index > len(ft.kbuckets) {
		return nil, errors.New("Invalid index")
	}

	return ft.kbuckets[index], nil
}

func (ft *kademliaFingerTable) GetClosestNodes(k int, key Key) ([]Kademlia, error) {

	closestNodes := make([]Kademlia, 0)

	dist, err := ft.id.XOR(key)
	if err != nil {
		return nil, err
	}

	for i := dist.Lenght() - 1; i > 0; i-- {
		if dist.IsActive(i) {
			newClosestNodes, err := ft.kbuckets[i].GetClosestNodes(k-len(closestNodes), key)
			if err != nil {
				return nil, err
			}
			if newClosestNodes != nil {
				closestNodes = append(closestNodes, newClosestNodes...)
			}

		}
		if len(closestNodes) == k {
			return closestNodes, nil
		}
	}

	for i := 0; i < dist.Lenght(); i++ {
		if !dist.IsActive(i) {
			newClosestNodes, err := ft.kbuckets[i].GetClosestNodes(k-len(closestNodes), key)
			if err != nil {
				return nil, err
			}
			if newClosestNodes != nil {
				closestNodes = append(closestNodes, newClosestNodes...)
			}
		}
		if len(closestNodes) == k {
			return closestNodes, nil
		}
	}
	return closestNodes, nil

}

func (ft *kademliaFingerTable) Update(node Kademlia) error {
	key := node.GetNodeId()
	equal, err := key.Equal(ft.id)
	if err != nil {
		return err
	}
	if equal {
		return nil
	}
	dist, _ := ft.id.XOR(node.GetNodeId())
	var index int = ft.id.Lenght() - 1
	for ; index >= 0; index-- {
		if dist.IsActive(index) {
			break
		}

	}

	kbucket, err := ft.GetKBucket(index)
	if err != nil {
		return err
	}

	kbucket.Update(node)
	return nil
}

func (ft *kademliaFingerTable) GetKeyFromKBucket(k int) Key {
	myKey := ft.id
	artKey := KeyNode{}
	indexByte := k / 8
	indexBit := uint8(k % 8)
	for i, bt := range myKey {
		artKey[i] = bt
		if i == indexByte {
			artKey[i] = bt ^ 1<<(indexBit)
		}
	}

	return &artKey
}
