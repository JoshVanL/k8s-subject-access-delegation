package main

import (
	"time"

	"github.com/sirupsen/logrus"
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

var (
	apiserverURL = "http://127.0.0.1:8001"
	masterURL    string
	kubeconfig   = "/home/josh/.kube/config"
)

func main() {
	log := logrus.NewEntry(logrus.New())
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	exampleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building example clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	exampleInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)

	controller := controller.NewController(kubeClient, exampleClient, kubeInformerFactory, exampleInformerFactory, log)

	go kubeInformerFactory.Start(stopCh)
	go exampleInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		log.Fatalf("Error running controller: %s", err.Error())
	}
}
