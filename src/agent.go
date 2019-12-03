package core

type Agent struct {
	Name     string
	Function string

	EndpointServices []Addr
	IsAliveService   []Addr
	Documentation    map[Addr]Addr
	Similar          []string
	TestCases        []TestCase
}

func NewCreate(name, functionality string, endpoints []Addr, doc map[Addr]Addr, testCases []TestCase) *Agent {
	agent := &Agent{
		Name:             name,
		Function:         functionality,
		EndpointServices: endpoints,
		IsAliveService:   nil,
		Documentation:    doc,
		Similar:          nil,
		TestCases:        testCases,
	}
	go agent.UpdateAll()
	return agent
}

func (agent Agent) GetAliveService() Addr {
	return Addr{}
}

func (agent Agent) UpdateAliveServices() {
	// TODO
}

func (agent Agent) UpdateSimilar() {
	// TODO
}

func (agent Agent) UpdateAll() {
	agent.UpdateAliveServices()
	agent.UpdateSimilar()
}
