package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"

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
	http.HandleFunc("/", EndpointHandler)
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", port+1000), nil)
		if err != nil {
			return
		}
	}()

	db := core.DatabaseAndPexBasedOnKademlia{Kd: ln}
	if !gateway {
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
	} else {
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

	platform := core.NewPlatform(core.Addr{Ip: ip, Port: port + 2000}, &db, &db)
	server := core.NewServer("", *platform, platform.Addr)
	go server.RunServer()
	if s := <-exited; s {
		// Handle Error in method
		fmt.Println("We get an error listen server")
		return
	}
}

func EndpointHandler(w http.ResponseWriter, r *http.Request) {
	data := Node.GetInfo()
	_, err := fmt.Fprintf(w, "%s", data)
	if err != nil {
		return
	}
}
