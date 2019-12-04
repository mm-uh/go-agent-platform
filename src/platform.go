package core

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

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
	err := pl.DataBase.GetLock(Name, names)
	if err != nil {
		return false
	}
	taken := CheckWord(names, agent.Name)
	if taken {
		go func() {
			for {
				err := pl.DataBase.StoreLock(Name, names)
				if err != nil {
					continue
				}
				break
			}
		}()
		return false
	}
	names = AddWord(names, agent.Name)
	err = pl.DataBase.StoreLock(Name, names)
	if err != nil {
		return false
	}
	eraseName := func() {
		for {
			names = &Trie{}
			err := pl.DataBase.GetLock(Name, names)
			if err != nil {
				continue
			}
			RemoveWord(names, agent.Name)
			err = pl.DataBase.StoreLock(Name, names)
			if err != nil {
				continue
			}
			break
		}

	}

	functions := &Trie{}

	err = pl.DataBase.GetLock(Function, functions)
	if err != nil {
		go eraseName()
		return false
	}
	exist := CheckWord(functions, agent.Function)
	if !exist {
		functions = AddWord(functions, agent.Function)
	}
	err = pl.DataBase.StoreLock(Function, functions)
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
	err = pl.DataBase.GetLock(fmt.Sprintf("%s:%s", Function, agent.Function), &agentsByFunction)
	if err != nil {
		go eraseName()
		return false
	}
	agentsByFunction = append(agentsByFunction, agent.Name)
	err = pl.DataBase.StoreLock(fmt.Sprintf("%s:%s", Function, agent.Function), agentsByFunction)
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
// Only one Agent
// Response reference:
// Response Should contain 3 Addr
// Response[0] Agent Addr
// Response[1] Agent Is Alive endpoint Addr
// Response[2] Agent Documentation Addr
func (pl Platform) LocateAgent(name string) ([3]Addr, error) {
	var agent Agent
	// Here we follow the indexation criteria:
	// [keys] : [Value] -> [criteria:AgentName] : [Agent]
	err := pl.DataBase.Get(Name+":"+name, &agent)
	if err != nil {
		return [3]Addr{}, err
	}
	addr := [3]Addr{}
	for key, val := range agent.IsAliveService {
		if isAlive(val.Ip + ":" + strconv.Itoa(val.Port)) {
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
func isAlive(endpoint string) bool {

	conn, err := net.Dial("tcp", endpoint)
	if err != nil {
		return false
	}
	err = conn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return false
	}
	text := "Alive?"
	// send to socket
	_, err = fmt.Fprintf(conn, text+"\n")
	if err != nil {
		return false
	}

	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return false
	}
	return message == "Yes\n"
}

// Get a specific agents matching a criteria, Should be one of next's:
// criteria:
//	ByName
//	ByFunction
// Only one Agent
func (pl Platform) GetAgentsByFunctions(name string) ([][3]Addr, error) {
	var agents []string
	// Here we follow the indexation criteria:
	// [keys] : [Value] -> [criteria:AgentName] : [Agent]
	err := pl.DataBase.Get(Function+":"+name, &agents)
	if err != nil {
		return nil, nil
	}
	response := make([][3]Addr, 0)
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
	var agent Agent
	// Here we follow the indexation criteria:
	// [keys] : [Value] -> [criteria:AgentName] : [Agent]
	err := pl.DataBase.Get(Name+":"+agentName, &agent)
	if err != nil {
		return nil
	}
	UpdateSimilarToAgent(&agent, &pl)
	return agent.Similar
}
