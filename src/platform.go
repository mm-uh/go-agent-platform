package core

type Platform struct {
	Db   DataBase
	Addr Addr
	Pex  Pex
}

func (pl Platform) Register(agent Agent) bool {
	return true
}

func (pl Platform) LocateAgentByName(name string) Addr {
	return Addr{}
}

func (pl Platform) GetAllAgentLocationsByName(name string) []Addr {
	return nil
}

func (pl Platform) LocateAgentByFunction(name string) Addr {
	return Addr{}
}

func (pl Platform) GetAllAgentLocationsByFunction(name string) []Addr {
	return nil
}

func (pl Platform) GetSimilarToAgent(name string) []Addr {
	return nil
}

