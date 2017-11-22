package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	clientset "github.com/joshvanl/k8s-subject-access-delegation/pkg/client/clientset/versioned"
	informers "github.com/joshvanl/k8s-subject-access-delegation/pkg/client/informers/externalversions"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/controller"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/signals"
)

const FlagApiServerURL = "api-url"
const FlagKubeConfig = "kube-config"

var RootCmd = &cobra.Command{
	Use:   "subject-access-delegation",
	Short: "<< Subject Access Delegation via Role Bindings onto resources in Kubernetes using event and time based triggers",
	Run: func(cmd *cobra.Command, args []string) {
		log := LogLevel(cmd)

		var masterURL string

		kubeconfig, err := cmd.PersistentFlags().GetString(FlagKubeConfig)
		if err != nil {
			log.Fatal(err)
		}

		kubeconfig, err = homedir.Expand(kubeconfig)
		if err != nil {
			log.Fatalf("unable to expand config directory ('%s'): %v", kubeconfig, err)
		}

		cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
		if err != nil {
			log.Fatalf("error building kubeconfig: %v", err)
		}

		kubeClient, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			log.Fatalf("error building kubernetes clientset: %v", err)
		}

		exampleClient, err := clientset.NewForConfig(cfg)
		if err != nil {
			log.Fatalf("error building example clientset: %v", err)
		}

		kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
		exampleInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)
		controller := controller.NewController(kubeClient, exampleClient, kubeInformerFactory, exampleInformerFactory, log)

		stopCh := signals.SetupSignalHandler()

		go kubeInformerFactory.Start(stopCh)
		go exampleInformerFactory.Start(stopCh)

		if err = controller.Run(2, stopCh); err != nil {
			log.Fatalf("error running controller: %s", err.Error())
		}

	},
}

func init() {
	RootCmd.PersistentFlags().Int("log-level", 1, "Set the log level of output. 0-Fatal 1-Info 2-Debug")
	RootCmd.Flag("log-level").Shorthand = "l"
	RootCmd.PersistentFlags().StringP(FlagApiServerURL, "u", "http://127.0.0.1:8001", "Set URL of Kubernetes API")
	RootCmd.PersistentFlags().StringP(FlagKubeConfig, "c", "~/.kube/config", "Path to kube config")
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
