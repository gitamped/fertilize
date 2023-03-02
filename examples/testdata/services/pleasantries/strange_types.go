package pleasantries

import "github.com/gitamped/seed/server"

type StrangeTypesService interface {
	DoSomethingStrange(DoSomethingStrangeRequest, server.GenericRequest) DoSomethingStrangeResponse
}

type DoSomethingStrangeRequest struct {
	Anything interface{}
}

type DoSomethingStrangeResponse struct {
	Value interface{}
	Size  int
}

type StrangeTypesServicer struct{}

func (StrangeTypesServicer) DoSomethingStrange(d DoSomethingStrangeRequest, r server.GenericRequest) DoSomethingStrangeResponse {
	panic("not implemented")
}
