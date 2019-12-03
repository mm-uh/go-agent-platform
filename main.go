package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/golang/go/src/pkg/encoding/hex"
	core "github.com/mm-uh/go-agent-platform/src"
	kademlia "github.com/mm-uh/go-kademlia/src"
	"net/http"
	"os"
	"strconv"
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
	key.GetFromString(hex.EncodeToString(id[:]))
	val:= strconv.FormatInt(int64(port+1000), 10)
	ln.Store(ln.GetContactInformation(), &key, val)
	ln.RunServer(exited)
	http.HandleFunc("/", EndpointHandler)
	go http.ListenAndServe(fmt.Sprintf(":%d", port+1000), nil)

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
	}
	db := core.DatabaseAndPexBasedOnKademlia{Kd: ln}

	platform := core.NewPlatform(core.Addr{Ip:ip, Port:port+2000}, &db, &db)
	server := core.NewServer("", *platform, platform.Addr )
	go server.RunServer()
	if s := <-exited; s {
		// Handle Error in method
		fmt.Println("We get an error listen server")
		return
	}
}

func EndpointHandler(w http.ResponseWriter, r *http.Request) {
	data := Node.GetInfo()
	fmt.Fprintf(w, "%s", data)
}
