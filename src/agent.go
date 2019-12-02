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
