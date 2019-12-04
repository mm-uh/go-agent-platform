package core

type DataBase interface {
	Get(string, interface{}) error
	Store(string, interface{}) error
	Lock(string) error
	Unlock(string) error
}

type Pex interface {
	GetPeers() []Addr
}
