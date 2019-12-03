package core

import "strings"

type Addr struct {
	ip       string
	port     int
}

func Search(collection []string, criteria string) []int {
	responses := make([]string, 0)
	index := 0
	for index, val := range collection {
		if strings.Contains(val, criteria) {
			responses = append(responses, index)
		}
		index++
	}
	return responses
}