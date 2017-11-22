package cmd

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
)

var Version struct {
	Version   string
	BuildDate string
	Commit    string
}

var AppName string = "subject-access-delegation"

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: fmt.Sprintf("Version number of %s", AppName),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s", AppName)

		v := reflect.ValueOf(Version)

		for i := 0; i < v.NumField(); i++ {
			fmt.Printf(
				" %s: %s",
				strings.ToLower(v.Type().Field(i).Name),
				v.Field(i).Interface(),
			)
		}

		fmt.Println("")
	},
}
