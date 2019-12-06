package util

type RPCBase struct {
	MethodName string
	Args       interface{}
}

type ResponseRPC struct {
	Response string
	Error    error
}
