package core

type DataBase interface {
	Get(string, interface{}) error
	Store(string, interface{}) error
	GetLock(string, interface{}) error
	StoreLock(string, interface{}) error
}

type Pex interface {
	GetPeers() []Addr
}
