package kademlia

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/mm-uh/rpc_udp/src/util"
	"github.com/sirupsen/logrus"
)

type Contacts []Kademlia

type RemoteKademlia struct {
	key  Key
	ip   string
	port int
}

func NewRemoteKademlia(key Key, ip string, port int) *RemoteKademlia {
	return &RemoteKademlia{
		key:  key,
		ip:   ip,
		port: port,
	}
}

func NewRemoteKademliaWithoutKey(ip string, port int) *RemoteKademlia {
	key := KeyNode(sha1.Sum([]byte(fmt.Sprintf("%s:%d", ip, port))))
	return &RemoteKademlia{
		key:  &key,
		ip:   ip,
		port: port,
	}
}
func (kc *RemoteKademlia) GetContactInformation() *ContactInformation {
	return &ContactInformation{
		node: kc,
		time: 0,
	}
}

func (kc *RemoteKademlia) GetNodeId() Key {
	return kc.key
}

func (kc *RemoteKademlia) GetIP() string {
	return kc.ip
}

func (kc *RemoteKademlia) GetPort() int {
	return kc.port
}

func (kc *RemoteKademlia) Ping(info *ContactInformation) bool {
	methodName := "Ping"
	rpcBase := &util.RPCBase{
		MethodName: methodName,
	}
	args := make([]string, 0)
	args = append(args, contactInfoToString(info))
	rpcBase.Args = args

	response, err := kc.MakeRequest(rpcBase)
	if err != nil {
		return false
	}
	if response.Response != "Pong" {
		return false
	}
	return true
}

func (kc *RemoteKademlia) StoreOnNetwork(info *ContactInformation, key Key, i string) error {
	methodName := "StoreOnNetwork"
	rpcBase := &util.RPCBase{
		MethodName: methodName,
	}
	args := make([]interface{}, 0)
	args = append(args, contactInfoToString(info))
	args = append(args, key.String())
	args = append(args, i)
	rpcBase.Args = args
	response, err := kc.MakeRequest(rpcBase)
	if err != nil {
		return err
	}
	if response.Response != "true" {
		return errors.New("error storing value")
	}
	return nil
}

func (kc *RemoteKademlia) ClosestNodes(cInfo *ContactInformation, k int, key Key) ([]Kademlia, error) {
	methodName := "ClosestNodes"
	rpcBase := &util.RPCBase{
		MethodName: methodName,
	}
	args := make([]interface{}, 0)
	args = append(args, contactInfoToString(cInfo))
	args = append(args, strconv.FormatInt(int64(k), 10))
	args = append(args, key.String())
	rpcBase.Args = args

	nodeResponse, err := kc.MakeRequest(rpcBase)
	if err != nil {
		return nil, err
	}

	returnedNodes, err := getKademliaNodes(nodeResponse.Response)
	if err != nil {
		return nil, err
	}
	return returnedNodes, nil
}

func (kc *RemoteKademlia) GetFromNetwork(info *ContactInformation, key Key) (string, error) {
	methodName := "GetFromNetwork"
	rpcBase := &util.RPCBase{
		MethodName: methodName,
	}
	args := make([]string, 0)
	args = append(args, contactInfoToString(info))
	args = append(args, key.String())
	rpcBase.Args = args

	response, err := kc.MakeRequest(rpcBase)
	if err != nil {
		return "", err
	}
	if response.Error != nil {
		return "", response.Error
	}
	return response.Response, nil
}

func (kc *RemoteKademlia) JoinNetwork(kademlia Kademlia) error {
	methodName := "JoinNetwork"
	rpcBase := &util.RPCBase{
		MethodName: methodName,
	}
	args := make([]interface{}, 0)
	args = append(args, kademlia.GetNodeId())
	args = append(args, kademlia.GetIP())
	args = append(args, kademlia.GetPort())
	rpcBase.Args = args
	response, err := kc.MakeRequest(rpcBase)
	if err != nil {
		return err
	}
	if response.Response != "true" {
		return errors.New("error storing value")
	}
	return nil

}

func (kc *RemoteKademlia) Store(info *ContactInformation, key Key, i string) error {
	methodName := "Store"
	rpcBase := &util.RPCBase{
		MethodName: methodName,
	}
	args := make([]interface{}, 0)
	args = append(args, contactInfoToString(info))
	args = append(args, key.String())
	args = append(args, i)
	rpcBase.Args = args
	response, err := kc.MakeRequest(rpcBase)
	if err != nil {
		return err
	}
	if response.Response != "true" {
		return errors.New("error storing value")
	}
	return nil
}

func (kc *RemoteKademlia) Get(info *ContactInformation, key Key) (*TimeStampedString, error) {
	methodName := "Get"
	rpcBase := &util.RPCBase{
		MethodName: methodName,
	}
	args := make([]string, 0)
	args = append(args, contactInfoToString(info))
	args = append(args, key.String())
	rpcBase.Args = args

	response, err := kc.MakeRequest(rpcBase)
	if err != nil {
		return nil, err
	}
	if response.Error != nil {
		return nil, response.Error
	}
	var data TimeStampedString = TimeStampedString{}
	err = json.Unmarshal([]byte(response.Response), &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (kc *RemoteKademlia) GetLock(ci *ContactInformation, id Key) error {
	return nil
}

func (kc *RemoteKademlia) LeaveLock(ci *ContactInformation, id Key) error {
	return nil
}

func (kc *RemoteKademlia) LockValue(ci *ContactInformation, id Key) (bool, error) {
	methodName := "LockValue"
	rpcBase := &util.RPCBase{
		MethodName: methodName,
	}
	args := make([]string, 0)
	args = append(args, contactInfoToString(ci))
	args = append(args, id.String())
	rpcBase.Args = args

	response, err := kc.MakeRequest(rpcBase)
	if err != nil {
		return false, err
	}
	if response.Error != nil {
		return false, response.Error
	}
	var data bool = false
	err = json.Unmarshal([]byte(response.Response), &data)
	if err != nil {
		return false, err
	}
	return data, nil
}

func (kc *RemoteKademlia) UpdateKey(ci *ContactInformation, id Key, data *TimeStampedString) error {
	methodName := "UpdateKey"
	rpcBase := &util.RPCBase{
		MethodName: methodName,
	}

	args := make([]string, 0)
	args = append(args, contactInfoToString(ci))
	args = append(args, id.String())
	dataForSend, err := json.Marshal(data)
	if err != nil {
		return err
	}
	args = append(args, string(dataForSend))

	rpcBase.Args = args

	response, err := kc.MakeRequest(rpcBase)
	if err != nil {
		return err
	}

	return response.Error

}

func (kc *RemoteKademlia) UnlockValue(ci *ContactInformation, id Key) error {
	methodName := "UnlockValue"
	rpcBase := &util.RPCBase{
		MethodName: methodName,
	}
	args := make([]string, 0)
	args = append(args, contactInfoToString(ci))
	args = append(args, id.String())
	rpcBase.Args = args

	response, err := kc.MakeRequest(rpcBase)
	if err != nil {
		return err
	}
	if response.Error != nil {
		return response.Error
	}
	return nil
}

func (kc *RemoteKademlia) MakeRequest(rpcBase *util.RPCBase) (*util.ResponseRPC, error) {

	service := kc.ip + ":" + strconv.Itoa(kc.port)

	RemoteAddr, err := net.ResolveUDPAddr("udp", service)

	conn, err := net.DialUDP("udp", nil, RemoteAddr)

	if err != nil {
		logrus.Warn(err, 1)
		return nil, err
	}

	defer conn.Close()

	// write a message to server
	toSend, err := json.Marshal(rpcBase)
	if err != nil {
		logrus.Warn(err, 2)
		return nil, err
	}

	message := []byte(string(toSend))

	_, err = conn.Write(message)

	if err != nil {
		logrus.Warn("Error: "+err.Error(), 3)
		return nil, err
	}

	timeout := 5 * time.Second
	deadline := time.Now().Add(timeout)
	err = conn.SetReadDeadline(deadline)
	if err != nil {
		return nil, err
	}
	// receive message from server
	buffer := make([]byte, 10000)
	n, _, err := conn.ReadFromUDP(buffer)
	if n == 0 {
		//logrus.Warn("from node with method: "+rpcBase.MethodName, " from: ", service)
		logrus.Warn(err.Error(), " HERE")
		return nil, err
	}
	//fmt.Println(string(buffer[:n]))
	var response util.ResponseRPC
	err = json.Unmarshal(buffer[:n], &response)
	if err != nil {
		logrus.Warn("Error Unmarshaling response 4 ")
		return nil, err
	}
	return &response, nil
}

// Get Kademlia nodes from string
func getKademliaNodes(kdNodes string) ([]Kademlia, error) {
	nodes := strings.Split(kdNodes, "$")
	response := make([]Kademlia, 0)
	for _, val := range nodes {
		node := strings.Split(val, ",")
		// node[0] =-> NodeId
		// node[1] =-> Ip
		// node[2] =-> Port
		if len(node) != 3 {
			return nil, errors.New("fail parsing nodes, mismatch params number")
		}
		var key KeyNode
		err := key.GetFromString(node[0])
		if err != nil {
			return nil, errors.New("fail parsing nodes, couldn't get key from string")
		}

		port, err := strconv.Atoi(node[2])
		if err != nil {
			return nil, errors.New("fail parsing nodes, couldn't get port from string")
		}
		response = append(response, NewRemoteKademlia(&key, node[1], port))
	}

	return response, nil
}

// Convert Contact Information into string
func contactInfoToString(cInfo *ContactInformation) string {
	return fmt.Sprintf("%v,%v,%v,%v", cInfo.node.GetNodeId().String(), cInfo.node.GetIP(), cInfo.node.GetPort(), cInfo.time)
}
