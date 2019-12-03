package core

type Agent struct {
	Name     string
	Function string

	EndpointServices []Addr
	IsAliveService   map[Addr]Addr
	Documentation    map[Addr]Addr
	Similar          []string
	TestCases        []TestCase
}

func NewAgent(name, functionality string, endpoints []Addr, alive, doc map[Addr]Addr, testCases []TestCase) *Agent {
	agent := &Agent{
		Name:             name,
		Function:         functionality,
		EndpointServices: endpoints,
		IsAliveService:   alive,
		Documentation:    doc,
		Similar:          nil,
		TestCases:        testCases,
	}
	go agent.UpdateSimilar()
	return agent
}

func (agent Agent) GetAliveService() Addr {
	return Addr{}
}

func (agent Agent) UpdateSimilar() {
	// TODO
}
