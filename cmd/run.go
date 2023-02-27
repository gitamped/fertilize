package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gitamped/fertilize/parser"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fertilize",
	Short: "Fertilize describes Go packages in a generic way.",
	Long: `A way to generate package information to use in a template engine.
		   Use to generate boilerplate code and documents.`,
	Run: func(cmd *cobra.Command, args []string) {
		p := parser.New([]string{"github.com/gitamped/fertilize/testdata/services/pleasantries"}...)
		def, err := p.Parse()
		if err != nil {
			panic("err parsing")
		}
		b, err := json.Marshal(def)
		fmt.Printf(string(b))
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
