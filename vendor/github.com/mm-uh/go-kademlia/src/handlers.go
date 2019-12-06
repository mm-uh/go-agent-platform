package kademlia

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type HandlerRPC struct {
	kademlia Kademlia
}

func (handler *HandlerRPC) Ping(cInfo string) string {
	info := getContactInformationFromString(cInfo)
	pong := handler.kademlia.Ping(info)
	return strconv.FormatBool(pong)
}

func (handler *HandlerRPC) Store(cInfo string, keyAsString string, i string) string {
	var key KeyNode
	err := key.GetFromString(keyAsString)
	if err != nil {
		return "false"
	}
	err = handler.kademlia.Store(getContactInformationFromString(cInfo), &key, i)
	if err != nil {
		return "false"
	}
	return "true"
}

func (handler *HandlerRPC) Get(cInfo, keyAsString string) string {
	var key KeyNode
	err := key.GetFromString(keyAsString)
	if err != nil {
		return "false"
	}
	val, err := handler.kademlia.Get(getContactInformationFromString(cInfo), &key)
	if err != nil {
		return "false"
	}
	str, err := json.Marshal(*val)
	if err != nil {
		return "false"
	}
	return string(str)
}

func (handler *HandlerRPC) StoreOnNetwork(cInfo, keyAsString string, i string) string {
	var key KeyNode
	err := key.GetFromString(keyAsString)
	if err != nil {
		return "false"
	}
	err = handler.kademlia.StoreOnNetwork(getContactInformationFromString(cInfo), &key, i)
	if err != nil {
		return "false"
	}
	return "true"
}

func (handler *HandlerRPC) ClosestNodes(cInfo, k, keyAsString string) string {
	var key KeyNode
	err := key.GetFromString(keyAsString)

	if err != nil {
		return "false"
	}
	kVal, err := strconv.Atoi(k)
	if err != nil {
		return "false"
	}
	val, err := handler.kademlia.ClosestNodes(getContactInformationFromString(cInfo), kVal, &key)
	if err != nil {
		return "false"
	}
	return getStrFromKademliaNodes(val)
}

func (handler *HandlerRPC) UpdateKey(cInfo, keyAsString, data string) string {
	var key KeyNode
	//	fmt.Println("RECEIVING UPDATE ", keyAsString)
	err := key.GetFromString(keyAsString)
	if err != nil {
		//		fmt.Println("JODIO 1")
		return "false"
	}

	var timeStampedData *TimeStampedString = &TimeStampedString{}
	err = json.Unmarshal([]byte(data), timeStampedData)
	// fmt.Println("DATA AT HANDLER ", data)
	if err != nil {
		// fmt.Println("JODIO 2")
		// fmt.Println(err.Error())
		return "false"
	}
	err = handler.kademlia.UpdateKey(getContactInformationFromString(cInfo), &key, timeStampedData)
	if err != nil {
		// fmt.Println("JODIO 3")
		return "false"
	}

	return "true"
}
func (handler *HandlerRPC) LockValue(cInfo, keyAsString string) string {
	var key KeyNode
	err := key.GetFromString(keyAsString)
	if err != nil {
		return "false"
	}

	val, err := handler.kademlia.LockValue(getContactInformationFromString(cInfo), &key)
	if err != nil {
		return "false"
	}
	return strconv.FormatBool(val)
}

func (handler *HandlerRPC) UnlockValue(cInfo, keyAsString string) string {
	var key KeyNode
	err := key.GetFromString(keyAsString)
	if err != nil {
		return "false"
	}

	err = handler.kademlia.UnlockValue(getContactInformationFromString(cInfo), &key)
	if err != nil {
		return "false"
	}
	return ""
}

func (handler *HandlerRPC) GetFromNetwork(cInfo, keyAsString string) string {
	var key KeyNode
	err := key.GetFromString(keyAsString)
	if err != nil {
		return "false"
	}
	val, err := handler.kademlia.GetFromNetwork(getContactInformationFromString(cInfo), &key)
	if err != nil {
		return "false"
	}
	return val
}
func (handler *HandlerRPC) GetNodeId() string {
	return handler.kademlia.GetNodeId().String()
}
func (handler *HandlerRPC) GetIP() string {
	return handler.kademlia.GetIP()
}
func (handler *HandlerRPC) GetPort() string {
	return strconv.FormatInt(int64(handler.kademlia.GetPort()), 10)
}
func (handler *HandlerRPC) JoinNetwork(keyAsString, ip, portAsString string) string {
	var key KeyNode
	err := key.GetFromString(keyAsString)
	if err != nil {
		return "false"
	}
	port, err := strconv.Atoi(portAsString)
	if err != nil {
		return "false"
	}
	kdToJoin := NewRemoteKademlia(&key, ip, port)
	err = handler.kademlia.JoinNetwork(kdToJoin)
	if err != nil {
		return "false"
	}
	return "true"
}

func getContactInformationFromString(cInfo string) *ContactInformation {
	values := strings.Split(cInfo, ",")
	// values[0] =-> NodeId
	// values[1] =-> Ip
	// values[2] =-> Port
	// values[3] =-> time
	if len(values) != 4 {
		return &ContactInformation{}
	}
	var key KeyNode
	err := key.GetFromString(values[0])
	if err != nil {
		return &ContactInformation{}
	}

	port, err := strconv.Atoi(values[2])
	if err != nil {
		return &ContactInformation{}
	}

	tm, err := strconv.Atoi(values[3])
	if err != nil {
		return &ContactInformation{}
	}

	return &ContactInformation{
		node: NewRemoteKademlia(&key, values[1], port),
		time: uint64(tm),
	}
}

func getStrFromKademliaNodes(kds []Kademlia) string {
	var values []string
	for _, val := range kds {
		values = append(values, fmt.Sprintf("%v,%v,%v", val.GetNodeId(), val.GetIP(), val.GetPort()))
	}
	return strings.Join(values, "$")
}
