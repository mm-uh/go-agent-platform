package core

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Platform struct {
	Addr     Addr
	DataBase DataBase
	Pex      Pex
}

const (
	Name       = "Name"
	Function   = "Function"
	IsAlive    = "IsAlive?"
	PlatformId = "PlatformId"
)

func NewPlatform(addr Addr, db DataBase, pex Pex) *Platform {
	return &Platform{
		Addr:     addr,
		DataBase: db,
		Pex:      pex,
	}
}

func (pl Platform) ListenBroadcast(port, portForConnect int) {
	var response string = ""
	for {

		err := pl.DataBase.Get(PlatformId, &response)
		if err != nil {
			continue
		}
		break
	}
	response = fmt.Sprintf("%s:%s:%d:", response, pl.Addr.Ip, portForConnect)
	p := strconv.Itoa(port)
	addr, err := net.ResolveUDPAddr("udp", ":"+p)
	if err != nil {
		return
	}

	conn, err := net.ListenUDP("udp", addr)

	if err != nil {
		return
	}

	for {
		request := make([]byte, 1024)
		_, from, err := conn.ReadFromUDP(request)
		if err != nil {
			continue
		}

		if string(request)[:4] == "JOIN" {
			_, err := conn.WriteToUDP([]byte(response), from)
			if err != nil {
				continue
			}

		}

	}

}

func (pl Platform) EditAgent(agent *Agent) bool {
	var tmpAgent AgentWithUncheckedSimilars = AgentWithUncheckedSimilars{}
	// Here we follow the indexation criteria:
	// [keys] : [Value] -> [criteria:AgentName] : [Agent]
	err := pl.DataBase.Get(Name+":"+agent.Name, &tmpAgent)
	if err != nil {
		return false
	}
	if agent.Password != tmpAgent.Agent.Password {
		return false
	}
	newAg := AgentWithUncheckedSimilars{
		Agent:      *agent,
		Uncheckeds: tmpAgent.Uncheckeds,
	}
	err = pl.DataBase.Store(fmt.Sprintf("%s:%s", Name, agent.Name), &newAg)
	if err != nil {
		return false
	}
	return true
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
	err = pl.DataBase.Get(Function, functions)
	if err != nil {
		go eraseName()
		return false
	}
	exist := CheckWord(functions, agent.Function)
	if !exist {
		functions = AddWord(functions, agent.Function)
	}
	err = pl.DataBase.Store(Function, functions)
	if err != nil {
		go eraseName()
		return false
	}

	regAgent := &AgentWithUncheckedSimilars{
		Agent:      *agent,
		Uncheckeds: make([]UncheckedSimilar, 0),
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
	fmt.Println(agentsByFunction)
	for _, a := range agentsByFunction {
		regAgent.Uncheckeds = append(regAgent.Uncheckeds, UncheckedSimilar{
			Name:      a,
			MyPeer:    false,
			OtherPeer: false,
		})
	}

	err = pl.DataBase.Store(fmt.Sprintf("%s:%s", Name, agent.Name), regAgent)
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
	go UpdateSimilarToAgent(regAgent.Agent.Name, &pl)
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

// Get all agents location Matching a criteria, Should be one of next's:
// criteria:
//	ByName: Only 0 or 1 Agent should exits if we have this criteria
//	ByFunction: As many as agents in our platform
func (pl Platform) GetAllFunctionNames() ([]string, error) {
	var functionsNames Trie
	// Should return a []string in agentsNames
	// Represent all agents names
	err := pl.DataBase.Get(Function, &functionsNames)
	if err != nil {
		return nil, err
	}
	return GetAllWords(&functionsNames), nil
}

// Get a specific agents matching a criteria, Should be one of next's:
// criteria:
//	ByName
// Only one Agent
// Response reference:
// Response Should contain 3 Addr
// Response[0] Agent Addr
// Response[1] Agent Is Alive endpoint Addr
// Response[2] Agent Documentation Addr
func (pl Platform) LocateAgent(name string) ([3]Addr, error) {
	var agentWithUnchecked AgentWithUncheckedSimilars
	// Here we follow the indexation criteria:
	// [keys] : [Value] -> [criteria:AgentName] : [Agent]
	err := pl.DataBase.Get(Name+":"+name, &agentWithUnchecked)
	if err != nil {
		return [3]Addr{}, err
	}
	agent := agentWithUnchecked.Agent
	addr := [3]Addr{}
	for key, val := range agent.IsAliveService {
		if NodeIsAlive(val.Ip + ":" + strconv.Itoa(val.Port)) {
			addr[0] = getAddrFromStr(key)
			addr[1] = val
			doc, ok := agent.Documentation[key]
			if ok {
				addr[2] = doc
			}
			return addr, nil
		}
	}
	return addr, fmt.Errorf("any node is alive")
}

type RecoverAgent struct {
	Name     string
	Password string
}

// Recover an agent
func (pl Platform) RecoverAgent(recover RecoverAgent) (Agent, error) {
	var agentWithUnchecked AgentWithUncheckedSimilars
	// Here we follow the indexation criteria:
	// [keys] : [Value] -> [criteria:AgentName] : [Agent]
	err := pl.DataBase.Get(Name+":"+recover.Name, &agentWithUnchecked)
	if err != nil {
		return Agent{}, err
	}
	agent := agentWithUnchecked.Agent
	if IsAuthenticated(recover.Password, &agent) {
		return agent, nil
	}

	return Agent{}, errors.New("error recovering agent")
}

func getAddrFromStr(s string) Addr {
	a := strings.Split(s, ":")
	port, err := strconv.Atoi(a[1])
	if err != nil {
		return Addr{}
	}
	return Addr{
		Ip:   a[0],
		Port: port,
	}
}

// Check if agent is available
// Send over a tcp connection a message 'Alive?\n'
// Wait 5 seconds for response, that should be 'Yes\n'
func NodeIsAlive(endpoint string) bool {
	message, err := MakeRequest(endpoint, IsAlive)
	if err != nil {
		return false
	}
	return message == "Yes"
}

// Get a specific agents matching a criteria, Should be one of next's:
// criteria:
//	ByName
//	ByFunction
// Only one Agent
func (pl Platform) GetAgentsByFunctions(name string) ([]string, error) {
	var agents []string

	// Here we follow the indexation criteria:
	// [keys] : [Value] -> [criteria:AgentName] : [Agent]
	err := pl.DataBase.Get(Function+":"+name, &agents)
	if err != nil {
		return nil, nil
	}

	return agents, nil
}

// Return the name of the agents that are similar to this agent name
func (pl Platform) GetSimilarToAgent(agentName string) []string {
	var agentWithUnchecked AgentWithUncheckedSimilars
	// Here we follow the indexation criteria:
	// [keys] : [Value] -> [criteria:AgentName] : [Agent]
	err := pl.DataBase.Get(Name+":"+agentName, &agentWithUnchecked)
	if err != nil {
		return []string{}
	}
	//agent := agentWithUnchecked.Agent
	UpdateSimilarToAgent(agentWithUnchecked.Agent.Name, &pl)
	return agentWithUnchecked.Agent.Similar
}

type UpdaterAgent struct {
	Name     string
	Password string

	EndpointService []Addr
	IsAliveService  map[string]Addr
	Documentation   map[string]Addr
}

// Return the name of the agents that are similar to this agent name
func (pl Platform) AddEndpoints(agentUpdated UpdaterAgent) error {
	var agentWithUnchecked AgentWithUncheckedSimilars
	// Here we follow the indexation criteria:
	// [keys] : [Value] -> [criteria:AgentName] : [Agent]
	err := pl.DataBase.Get(Name+":"+agentUpdated.Name, &agentWithUnchecked)
	if err != nil {

		return err
	}
	agent := agentWithUnchecked.Agent
	if IsAuthenticated(agentUpdated.Password, &agent) {
		agent.EndpointService = Union(agent.EndpointService, agentUpdated.EndpointService)
		for k, v := range agentUpdated.Documentation {
			agent.Documentation[k] = v
		}
		for k, v := range agentUpdated.IsAliveService {
			agent.IsAliveService[k] = v
		}
	}
	newAg := AgentWithUncheckedSimilars{
		Agent:      agent,
		Uncheckeds: agentWithUnchecked.Uncheckeds,
	}
	err = pl.DataBase.Store(fmt.Sprintf("%s:%s", Name, agent.Name), &newAg)
	if err != nil {
		return err
	}
	go UpdateSimilarToAgent(newAg.Agent.Name, &pl)
	return nil
}

func IsAuthenticated(updatedAgentPassword string, agent2 *Agent) bool {
	return updatedAgentPassword == agent2.Password
}
