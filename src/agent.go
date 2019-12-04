package core

import (
	"github.com/sirupsen/logrus"
	"strconv"
)

type Agent struct {
	Name     string
	Function string

	EndpointService []Addr
	IsAliveService  map[string]Addr
	Documentation   map[string]Addr
	Similar         []string
	TestCases       []TestCase
}

func NewAgent(name, functionality string, endpoints []Addr, alive, doc map[string]Addr, testCases []TestCase) *Agent {
	agent := &Agent{
		Name:            name,
		Function:        functionality,
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
		var tempAgent Agent
		// Get Agents with the same function
		err := platform.DataBase.Get(Name+":"+val, &tempAgent)
		if err != nil {
			continue
		}
		if AreCompatibles(&tempAgent, agent) {
			tempAgent.Similar = append(tempAgent.Similar, val)
			err := platform.DataBase.StoreLock(Name+":"+val, &tempAgent)
			if err != nil {
				continue
			}
		}
		if AreCompatibles(agent, &tempAgent) {
			agent.Similar = append(agent.Similar, tempAgent.Name)
		}
	}
	if similar != len(agent.Similar) {
		err := platform.DataBase.StoreLock(Name+":"+agent.Name, &agent)
		if err != nil {
			logrus.Warn("Couldn't store agent")
		}
	}
}

func AreCompatibles(tempAgent , agent *Agent) bool {
	for key, val := range tempAgent.IsAliveService {
		// get if the endpoint is alive
		if isAlive(val.Ip + ":" + strconv.Itoa(val.Port)) {
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
