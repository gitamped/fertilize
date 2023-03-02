package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gitamped/fertilize/parser"
)

func main() {

	patterns := []string{"github.com/gitamped/fertilize/examples/testdata/services/pleasantries"}
	p := parser.New(patterns...)
	p.ExcludeInterfaces = []string{"Welcomer", "Ignorer"}
	p.Verbose = false
	def, err := p.Parse()
	if err != nil {
		panic("err parsing")
	}
	b, err := json.Marshal(def)
	t, _ := ioutil.ReadFile("templates/handlers.tmpl")
	var data map[string]parser.Definition
	json.Unmarshal(b, &data)

	tmpl, _ := template.New("test").Parse(string(t))

	for k, v := range data {
		p := strings.Replace(k, "github.com/gitamped/fertilize/examples/", "", -1)
		os.Truncate(filepath.Join(p, "handlers.go"), 0)
		f, err := os.OpenFile(filepath.Join(p, "handlers.go"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		tmpl.Execute(f, v)
		if err != nil {
			log.Fatal(err)
		}
	}
}
