package core

import (
	"errors"
	"strconv"
)

type AgentWithUncheckedSimilars struct {
	Agent      Agent
	Uncheckeds []UncheckedSimilar
}

type UncheckedSimilar struct {
	Name      string
	MyPeer    bool
	OtherPeer bool
}

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

func UpdateSimilarToAgent(agentName string, platform *Platform) {
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

	var agent AgentWithUncheckedSimilars
	platform.DataBase.Get(Name+":"+agentName, &agent)

	newUnchecked := make([]UncheckedSimilar, 0)
	for _, un := range agent.Uncheckeds {
		different := false
		var otherAgent AgentWithUncheckedSimilars
		key := Name + ":" + un.Name
		err := platform.DataBase.Lock(key)
		if err != nil {
			continue
		}
		defer unlockKey(key)
		err = platform.DataBase.Get(Name+":"+un.Name, &otherAgent)
		if err != nil {
			continue
		}

		if !un.MyPeer {
			ok, err := AreCompatibles(&agent.Agent, &otherAgent.Agent)
			if err == nil {
				un.MyPeer = true
				if !ok {
					different = true
				}
			}
		}

		if !un.OtherPeer && !different {
			ok, err := AreCompatibles(&otherAgent.Agent, &agent.Agent)
			if err == nil {
				un.OtherPeer = true
				if !ok {
					different = true
				}
			}
		}

		nU := UncheckedSimilar{
			Name:      un.Name,
			MyPeer:    un.MyPeer,
			OtherPeer: un.OtherPeer,
		}

		if un.OtherPeer && un.MyPeer {
			if !different {
				agent.Agent.Similar = append(agent.Agent.Similar, un.Name)
				otherAgent.Agent.Similar = append(otherAgent.Agent.Similar, agent.Agent.Name)
				for i, ag := range otherAgent.Uncheckeds {
					if ag.Name == agent.Agent.Name {
						otherAgent.Uncheckeds = removeUncheckSimilar(otherAgent.Uncheckeds, i)
					}
				}
			}
		} else {

			if !different {
				newUnchecked = append(newUnchecked, nU)
				otherAgent.Uncheckeds = append(otherAgent.Uncheckeds, UncheckedSimilar{
					Name:      agent.Agent.Name,
					MyPeer:    un.OtherPeer,
					OtherPeer: un.MyPeer,
				})
			} else {
				for i, ag := range otherAgent.Uncheckeds {
					if ag.Name == agent.Agent.Name {
						otherAgent.Uncheckeds = removeUncheckSimilar(otherAgent.Uncheckeds, i)
						break
					}
				}

			}

		}

		platform.DataBase.Store(key, &otherAgent)

	}

	agent.Uncheckeds = newUnchecked

	platform.DataBase.Store(Name+":"+agentName, &agent)

}

func AreCompatibles(tempAgent, agent *Agent) (bool, error) {
	for key, val := range tempAgent.IsAliveService {
		// get if the endpoint is alive
		if NodeIsAlive(val.Ip + ":" + strconv.Itoa(val.Port)) {
			// Check that all test cases follow the criteria
			for _, testCase := range agent.TestCases {
				// Check if are compatibles the test cases
				if !checkTestCase(testCase, key) {
					return false, nil
				}
				break
			}
			return true, nil
		}
	}
	return false, errors.New("There is not available endpoint")
}

func checkTestCase(testCase TestCase, host string) bool {
	response, err := MakeRequest(host, testCase.Input)
	if err != nil {
		return false
	}
	return response == testCase.Output
}

func NewAgentWithUncheckedSimilars(name, functionality string, endpoints []Addr, alive, doc map[string]Addr, testCases []TestCase, password string) *AgentWithUncheckedSimilars {
	return &AgentWithUncheckedSimilars{
		Agent:      *NewAgent(name, functionality, endpoints, alive, doc, testCases, password),
		Uncheckeds: make([]UncheckedSimilar, 0),
	}
}

func removeUncheckSimilar(list []UncheckedSimilar, index int) []UncheckedSimilar {
	return append(list[:index], list[index+1:]...)
}
