package core

import (
	"github.com/sirupsen/logrus"
	"strconv"
)

type Agent struct {
	Name     string
	Function string
	Password string

	EndpointService []Addr
	IsAliveService  map[string]Addr
	Documentation   map[string]Addr
	Similar         []string
	TestCases       []TestCase
}

func NewAgent(name, functionality string, endpoints []Addr, alive, doc map[string]Addr, testCases []TestCase, password string) *Agent {
	agent := &Agent{
		Name:            name,
		Function:        functionality,
		Password:        password,
		EndpointService: endpoints,
		IsAliveService:  alive,
		Documentation:   doc,
		Similar:         []string{},
		TestCases:       testCases,
	}
	return agent
}

func UpdateSimilarToAgent(agent *Agent, platform *Platform) {
	var agents []string
	var similar int
	unlockKey := func(key string) {
		go func() {
			for {
				err := platform.DataBase.Unlock(key)
				if err != nil {
					continue
				}
				break
			}
		}()
	}
	if agent.Similar == nil {
		similar = 0
	} else {
		similar = len(agent.Similar)
	}
	// Here we follow the indexation criteria:
	// [keys] : [Value] -> [criteria:AgentName] : [Agent]
	err := platform.DataBase.Get(Function+":"+agent.Function, &agents)
	if err != nil {
		return
	}
	for _, val := range agents {
		if val == agent.Name {
			continue
		}
		var tempAgent Agent
		// Get Agents with the same function
		err := platform.DataBase.Get(Name+":"+val, &tempAgent)
		if err != nil {
			continue
		}
		if AreCompatibles(&tempAgent, agent) {
			err := platform.DataBase.Lock(Name + ":" + val)
			if err != nil {
				continue
			}
			defer unlockKey(Name + ":" + val)
			tempAgent.Similar = append(tempAgent.Similar, val)
			err = platform.DataBase.Store(Name+":"+val, &tempAgent)
			if err != nil {
				continue
			}
		}
		if AreCompatibles(agent, &tempAgent) {
			agent.Similar = append(agent.Similar, tempAgent.Name)
		}
	}
	if similar != len(agent.Similar) {
		err := platform.DataBase.Lock(Name + ":" + agent.Name)
		if err != nil {
			logrus.Warn("Couldn't store agent")
			return
		}
		defer unlockKey(Name + ":" + agent.Name)
		err = platform.DataBase.Store(Name+":"+agent.Name, &agent)
		if err != nil {
			logrus.Warn("Couldn't store agent")
		}
	}
}

func AreCompatibles(tempAgent, agent *Agent) bool {
	for key, val := range tempAgent.IsAliveService {
		// get if the endpoint is alive
		if NodeIsAlive(val.Ip + ":" + strconv.Itoa(val.Port)) {
			accepted := 0
			// Check that all test cases follow the criteria
			for _, testCase := range agent.TestCases {
				// Check if are compatibles the test cases
				if checkTestCase(testCase, key) {
					accepted++
					continue
				}
				break
			}
			return accepted == len(agent.TestCases)
		}
	}
	return false
}

func checkTestCase(testCase TestCase, host string) bool {
	response, err := MakeRequest(host, testCase.Input)
	if err != nil {
		return false
	}
	return response == testCase.Output
}
