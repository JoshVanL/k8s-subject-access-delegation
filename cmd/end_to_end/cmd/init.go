package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/end_to_end/test"
)

const FlagLogLevel = "log-level"
const FlagTestFile = "test-file"

var RootCmd = &cobra.Command{
	Use:   "end-to-end",
	Short: "- Binary used to complete end to end testing of Subject Access Delegation.",
	Run: func(cmd *cobra.Command, args []string) {
		log := LogLevel(cmd)

		filename, err := cmd.PersistentFlags().GetString(FlagTestFile)
		if err != nil {
			log.Fatalf("failed to get test file from flag: %v", err)
		}

		testingSuite, err := test.NewSuite(filename, log)
		if err != nil {
			log.Fatalf("failed to create testing suite: %v", err)
		}

		if err := testingSuite.RunTests(); err != nil {
			log.Fatalf("error running end to end tests: %v", err)
		}

		log.Infof("== All tests passed. ==")
	},
}

func init() {
	RootCmd.PersistentFlags().IntP(FlagLogLevel, "l", 1, "Set the log level of output. 0-Fatal 1-Info 2-Debug")
	RootCmd.PersistentFlags().StringP(FlagTestFile, "t", "tests.yaml", "File path of the yaml test file.")
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func LogLevel(cmd *cobra.Command) *logrus.Entry {
	logger := logrus.New()

	i, err := cmd.PersistentFlags().GetInt("log-level")
	if err != nil {
		logrus.Fatalf("failed to get log level of flag: %s", err)
	}
	if i < 0 || i > 2 {
		logrus.Fatalf("not a valid log level")
	}
	switch i {
	case 0:
		logger.Level = logrus.FatalLevel
	case 1:
		logger.Level = logrus.InfoLevel
	case 2:
		logger.Level = logrus.DebugLevel
	}

	return logrus.NewEntry(logger)
}
