package main

import (
	"encoding/json"
	"fmt"

	"github.com/gitamped/fertilize/parser"
)

func main() {
	p := parser.New([]string{"github.com/gitamped/fertilize/testdata/services/pleasantries"}...)
	def, err := p.Parse()
	if err != nil {
		panic("err parsing")
	}
	b, err := json.Marshal(def)
	fmt.Printf(string(b))
}
