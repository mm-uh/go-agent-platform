package kademlia

type Kademlia interface {
	Ping(*ContactInformation) bool
	Store(*ContactInformation, Key, string) error
	Get(*ContactInformation, Key) (*TimeStampedString, error)
	StoreOnNetwork(*ContactInformation, Key, string) error
	GetFromNetwork(*ContactInformation, Key) (string, error)
	//GetAndLock(*ContactInformation, Key) (string, error)
	//StoreAndUnlock(*ContactInformation, Key, string) error
	GetLock(*ContactInformation, Key) error
	LeaveLock(*ContactInformation, Key) error
	UpdateKey(*ContactInformation, Key, *TimeStampedString) error
	ClosestNodes(*ContactInformation, int, Key) ([]Kademlia, error)
	LockValue(*ContactInformation, Key) (bool, error)
	UnlockValue(*ContactInformation, Key) error
	GetNodeId() Key
	GetIP() string
	GetPort() int
	JoinNetwork(Kademlia) error
}

type TimeStampedString struct {
	Data string `json:"Data"`
	Time uint64 `json:"Time"`
}

type KBucket interface {
	Update(Kademlia)
	GetClosestNodes(int, Key) ([]Kademlia, error)
	GetAllNodes() []Kademlia
}

type FingerTable interface {
	GetClosestNodes(int, Key) ([]Kademlia, error)
	GetKBucket(int) (KBucket, error)
	Update(Kademlia) error
	GetKeyFromKBucket(k int) Key
}

type Key interface {
	XOR(other Key) (Key, error)
	IsActive(index int) bool
	Lenght() int
	Less(other interface{}) (bool, error)
	Equal(other interface{}) (bool, error)
	String() string
	GetFromString(string) error
}
