package pleasantries

import (
	"github.com/gitamped/fertilize/examples/testdata/services"
	"github.com/gitamped/seed/server"
)

// GreeterService is a polite API.
// You will love it.
// strapline: "A lovely greeter service"
type GreeterService interface {
	// Greet creates a Greeting for one or more people.
	// featured: true
	Greet(GreetRequest, server.GenericRequest) GreetResponse
	// GetGreetings gets a range of saved Greetings.
	// featured: false
	GetGreetings(GetGreetingsRequest, server.GenericRequest) GetGreetingsResponse
}

// GreetRequest is the request object for GreeterService.Greet.
type GreetRequest struct {
	// Names are the names of the people to greet.
	// example: ["Mat", "David"]
	Names []string
}

// GreetResponse is the response object containing a
// person's greeting.
type GreetResponse struct {
	// Greeting is the greeted person's Greeting.
	Greeting *Greeting
}

// GetGreetingsRequest is the request object for GreeterService.GetGreetings.
// featured: true
type GetGreetingsRequest struct {
	// Page describes which page of data to get.
	Page services.Page `tagtest:"value,option1,option2"`
}

// GetGreetingsResponse is the respponse object for GreeterService.GetGreetings.
// featured: false
type GetGreetingsResponse struct {
	Greetings      []Greeting `json:"greetings"`
	GreetingsCount int        `json:"count,omitempty"`
}

// Greeting contains the pleasentry.
type Greeting struct {
	// Text is the message.
	// example: "Hello there"
	Text string
}

type GreeterServicer struct{}

func New() GreeterService {
	return &GreeterServicer{}
}

func (GreeterServicer) Greet(r GreetRequest, gr server.GenericRequest) GreetResponse {
	panic("not implemented")
}

func (GreeterServicer) GetGreetings(ggr GetGreetingsRequest, gr server.GenericRequest) GetGreetingsResponse {
	panic("not implemented")
}

