package core

import "fmt"

type Platform struct {
	Addr     Addr
	DataBase DataBase
	Pex      Pex
}

const (
	Name     = "Name"
	Function = "Function"
)

func NewPlatform(addr Addr, db DataBase, pex Pex) *Platform {
	return &Platform{
		Addr:     addr,
		DataBase: db,
		Pex:      pex,
	}
}

func (pl Platform) Register(agent *Agent) bool {
	names := &Trie{}
	unlockKey := func(key string) {
		go func() {
			for {
				err := pl.DataBase.Unlock(key)
				if err != nil {
					continue
				}
				break
			}
		}()
	}
	err := pl.DataBase.Lock(Name)
	if err != nil {
		return false
	}
	defer unlockKey(Name)
	err = pl.DataBase.Get(Name, names)
	if err != nil {
		return false
	}

	taken := CheckWord(names, agent.Name)
	if taken {

		return false
	}
	names = AddWord(names, agent.Name)
	err = pl.DataBase.Store(Name, names)
	if err != nil {
		return false
	}
	eraseName := func() {
		for {
			err := pl.DataBase.Lock(Name)
			defer unlockKey(Name)
			if err != nil {
				continue
			}
			RemoveWord(names, agent.Name)
			err = pl.DataBase.Store(Name, names)
			if err != nil {
				continue
			}
			break
		}

	}

	functions := &Trie{}
	err = pl.DataBase.Lock(Function)
	if err != nil {
		go eraseName()
		return false
	}
	defer unlockKey(Function)
	exist := CheckWord(functions, agent.Function)
	if !exist {
		functions = AddWord(functions, agent.Function)
	}
	err = pl.DataBase.Store(Function, functions)
	if err != nil {
		go eraseName()
		return false
	}
	err = pl.DataBase.Store(fmt.Sprintf("%s:%s", Name, agent.Name), agent)
	if err != nil {
		go eraseName()
		return false
	}

	agentsByFunction := make([]string, 0)
	err = pl.DataBase.Lock(fmt.Sprintf("%s:%s", Function, agent.Function))
	if err != nil {
		go eraseName()
		return false
	}
	defer unlockKey(fmt.Sprintf("%s:%s", Function, agent.Function))
	err = pl.DataBase.Get(fmt.Sprintf("%s:%s", Function, agent.Function), &agentsByFunction)
	if err != nil {
		go eraseName()
		return false
	}
	agentsByFunction = append(agentsByFunction, agent.Name)
	err = pl.DataBase.Store(fmt.Sprintf("%s:%s", Function, agent.Function), agentsByFunction)
	if err != nil {
		go eraseName()
		return false
	}

	return true
}

// Get all agents location Matching a criteria, Should be one of next's:
// criteria:
//	ByName: Only 0 or 1 Agent should exits if we have this criteria
//	ByFunction: As many as agents in our platform
func (pl Platform) GetAllAgentsNames() ([]string, error) {
	var agentsNames Trie
	// Should return a []string in agentsNames
	// Represent all agents names
	err := pl.DataBase.Get(Name, &agentsNames)
	if err != nil {
		return nil, err
	}
	return GetAllWords(&agentsNames), nil
}

// Get a specific agents matching a criteria, Should be one of next's:
// criteria:
//	ByName
//	ByFunction
// Only one Agent
func (pl Platform) LocateAgent(name string) (Agent, error) {
	var agent Agent
	// Here we follow the indexation criteria:
	// [keys] : [Value] -> [criteria:AgentName] : [Agent]
	err := pl.DataBase.Get(Name+":"+name, &agent)
	if err != nil {
		return Agent{}, err
	}
	return agent, nil
}

// Get a specific agents matching a criteria, Should be one of next's:
// criteria:
//	ByName
//	ByFunction
// Only one Agent
func (pl Platform) GetAgentsByFunctions(name string) ([]Agent, error) {
	var agents []string
	// Here we follow the indexation criteria:
	// [keys] : [Value] -> [criteria:AgentName] : [Agent]
	err := pl.DataBase.Get(Function+":"+name, &agents)
	if err != nil {
		return nil, nil
	}
	response := make([]Agent, 0)
	for _, val := range agents {
		agent, err := pl.LocateAgent(val)
		if err != nil {
			continue
		}
		response = append(response, agent)
	}
	return response, nil
}

// Return the name of the agents that are similar to this agent name
func (pl Platform) GetSimilarToAgent(agentName string) []string {
	agent, err := pl.LocateAgent(agentName)
	if err != nil {
		return nil
	}
	// agent.UpdateSimilar()
	return agent.Similar
}
