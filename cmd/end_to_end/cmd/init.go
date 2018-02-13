package cmd

import (
	"fmt"
	"os"
	//"time"

	//"github.com/hashicorp/go-multierror"
	//"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/end_to_end"
)

const FlagApiServerURL = "api-url"
const FlagKubeConfig = "kube-config"
const FlagWorkers = "worker-threads"
const FlagLogLevel = "log-level"
const FlagNTPHosts = "ntp-hosts"

var RootCmd = &cobra.Command{
	Use:   "end-to-end",
	Short: "- Binary used to complete end to end testing of Subject Access Delegation.",
	Run: func(cmd *cobra.Command, args []string) {
		log := LogLevel(cmd)

		if err := end_to_end.RunTests(log); err != nil {
			log.Fatalf("error running end to end tests: %v", err)
		}
	},
}

func init() {
	RootCmd.PersistentFlags().IntP(FlagLogLevel, "l", 1, "Set the log level of output. 0-Fatal 1-Info 2-Debug")
	//RootCmd.PersistentFlags().StringP(FlagApiServerURL, "u", "http://127.0.0.1:8001", "Set URL of Kubernetes API")
	//RootCmd.PersistentFlags().StringP(FlagKubeConfig, "c", "~/.kube/config", "Path to kube config")
	//RootCmd.PersistentFlags().IntP(FlagWorkers, "w", 2, "Number of worker threads for controller")
	//RootCmd.PersistentFlags().StringSliceP(FlagNTPHosts, "n", []string{""}, "Optional list of host URLs of ntp servers to ensure correct time")
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
