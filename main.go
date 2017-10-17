package main

import (
	"github.com/sirupsen/logrus"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/controller"
)

var (
	apiserverURL = "http://127.0.0.1:8001"
)

func main() {
	log := logrus.NewEntry(logrus.New())

	con, err := controller.New(apiserverURL, log)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Finished populating shared informer cache.")

	con.Work()
}
