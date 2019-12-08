package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	core "github.com/mm-uh/go-agent-platform/src"
	kademlia "github.com/mm-uh/go-kademlia/src"
)

var Node *kademlia.LocalKademlia

func main() {
	ip := os.Args[1]
	portStr := os.Args[2]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic("Invalid port")
	}

	gateway := len(os.Args) == 3

	ln := kademlia.NewLocalKademlia(ip, port, 20, 3)
	Node = ln
	exited := make(chan bool)
	key := kademlia.KeyNode{}
	id := sha1.Sum([]byte("PORT"))
	err = key.GetFromString(hex.EncodeToString(id[:]))
	if err != nil {
		return
	}
	val := strconv.FormatInt(int64(port+1000), 10)
	err = ln.Store(ln.GetContactInformation(), &key, val)
	if err != nil {
		return
	}

	ln.RunServer(exited)

	db := core.DatabaseAndPexBasedOnKademlia{Kd: ln}
	if !gateway {
		if len(os.Args) == 5 {
			ipForJoin := os.Args[3]
			portForJoinStr := os.Args[4]
			portForJoin, err := strconv.Atoi(portForJoinStr)
			if err != nil {
				panic("Invalid port for join")
			}
			rn := kademlia.NewRemoteKademliaWithoutKey(ipForJoin, portForJoin)
			err = ln.JoinNetwork(rn)
			if err != nil {
				panic("Can't Join")
			}
		} else if len(os.Args) == 4 {

			addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+os.Args[3])
			if err != nil {
				panic("Invalid port for join")
			}

			myAddr, err := net.ResolveUDPAddr("udp", ip+":")
			if err != nil {
				panic("Invalid port for join")
			}

			recvConn, err := net.ListenUDP("udp", myAddr)
			if err != nil {
				panic("Can not  Listen")
			}
			for {

				_, err = recvConn.WriteToUDP([]byte("JOIN"), addr)
				if err != nil {
					continue
				}
				response := make([]byte, 1024)

				recvConn.SetDeadline(time.Now().Add(time.Second * 2))
				_, from, err := recvConn.ReadFromUDP(response)
				if err != nil {
					continue
				}

				ipForJoin := strings.Split(string(response), ":")[1]
				portForJoin, err := strconv.Atoi(strings.Split(string(response), ":")[2])
				if err != nil {
					continue
				}
				fmt.Println("JOINING TO ", from.IP)
				rn := kademlia.NewRemoteKademliaWithoutKey(ipForJoin, portForJoin)
				err = ln.JoinNetwork(rn)
				if err != nil {
					continue
				}
				break
			}

		}

	} else {

		pltId := fmt.Sprintf("%s:%d:%s", ip, port, time.Now().String())
		hash := sha1.Sum([]byte(pltId))
		pltIdHash := hex.EncodeToString(hash[:])
		err = db.Store(core.PlatformId, pltIdHash)
		if err != nil {
			return
		}
		fmt.Println("Platform Id: ", pltIdHash)
		trie := core.NewTrie()
		err = db.Store(core.Name, trie)
		if err != nil {
			return
		}
		err = db.Store(core.Function, trie)
		if err != nil {
			return
		}

	}

	platform := core.NewPlatform(core.Addr{Ip: ip, Port: port + 1000}, &db, &db)
	go platform.ListenBroadcast(6001, ln.GetPort())
	server := core.NewServer("", *platform, platform.Addr)
	go server.RunServer()
	if s := <-exited; s {
		// Handle Error in method
		fmt.Println("We get an error listen server")
		return
	}
}
