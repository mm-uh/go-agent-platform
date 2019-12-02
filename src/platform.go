package core

type Platform struct {
	Addr     Addr
	DataBase DataBase
	Pex      Pex
}

const (
	Name     = "Name"
	Function = "Function"
)

func (pl Platform) NameAvailable(name string) bool {
	return true
}

func (pl Platform) Register(agent Agent) bool {
	return true
}

// Get all agents location Matching a criteria, Should be one of next's:
// criteria:
//	ByName: Only 0 or 1 Agent should exits if we have this criteria
//	ByFunction: As many as agents in our platform
func (pl Platform) GetAllAgentsNames() ([]string, error) {
	var agentsNames []string
	// Should return a []string in agentsNames
	// Represent all agents names
	err := pl.DataBase.Get(Name, &agentsNames)
	if err != nil {
		return nil, err
	}
	if agentsNames == nil {
		return make([]string, 0), nil
	}

	return agentsNames, nil
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
func (pl Platform) getAgentsByFunctions(name string) ([]Agent, error) {
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
