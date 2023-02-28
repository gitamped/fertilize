package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gitamped/fertilize/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	outfile    string
	pkgs       string
	v          bool
	ignoreList string
)

var rootCmd = &cobra.Command{
	Use:   "fertilize",
	Short: "Fertilize describes Go packages in a generic way.",
	Long: `A way to generate package information to use in a template engine.
		   Use to generate boilerplate code and documents.`,
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
		fmt.Printf(string(b))
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
	rootCmd.PersistentFlags().StringVar(&pkgs, "pkgs", "./...", "comma separated list of package patterns (default: ./...).")
	rootCmd.PersistentFlags().BoolVar(&v, "verbose", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&ignoreList, "ignore", "", "comma separated list of interfaces to ignore")

	viper.BindPFlag("out", rootCmd.PersistentFlags().Lookup("out"))
	viper.BindPFlag("pkgs", rootCmd.PersistentFlags().Lookup("pkgs"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("ignore", rootCmd.PersistentFlags().Lookup("ignore"))
}
