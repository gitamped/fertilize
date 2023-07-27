package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gitamped/fertilize/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	outfile    string
	pkgs       string
	tmplPath   string
	v          bool
	ignoreList string
)

var rootCmd = &cobra.Command{
	Use:   "fertilize",
	Short: "Fertilize describes Go packages in a generic way.",
	Long: `Fertilize describes Go packages in a generic way.

Use the output json to generate boilerplate code and documentation with
a template engine of your choice.`,
	Run: func(cmd *cobra.Command, args []string) {

		patterns := strings.Split(viper.GetString("pkgs"), ",")
		p := parser.New(patterns...)
		p.ExcludeInterfaces = strings.Split(viper.GetString("ignore"), ",")
		p.Verbose = viper.GetBool("verbose")
		def, err := p.Parse()
		if err != nil {
			panic("err parsing")
		}

		b, err := json.Marshal(def)
		t, _ := os.ReadFile(tmplPath)
		var data map[string]parser.Definition
		json.Unmarshal(b, &data)

		tmpl, _ := template.New("test").Parse(string(t))

		for _, v := range data {
			os.Truncate(filepath.Join(outfile), 0)
			f, err := os.OpenFile(filepath.Join(outfile), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}

			tmpl.Execute(f, v)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&outfile, "out", "", "output file (default: stdout)")
	rootCmd.PersistentFlags().StringVar(&pkgs, "pkgs", "./...", "comma separated list of package patterns")
	rootCmd.PersistentFlags().StringVar(&tmplPath, "tmpl", "handler.tmpl", "template filepath")
	rootCmd.PersistentFlags().BoolVar(&v, "verbose", false, "verbose output (default: false)")
	rootCmd.PersistentFlags().StringVar(&ignoreList, "ignore", "", "comma separated list of interfaces to ignore")

	viper.BindPFlag("out", rootCmd.PersistentFlags().Lookup("out"))
	viper.BindPFlag("pkgs", rootCmd.PersistentFlags().Lookup("pkgs"))
	viper.BindPFlag("tmpl", rootCmd.PersistentFlags().Lookup("tmpl"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("ignore", rootCmd.PersistentFlags().Lookup("ignore"))
}
