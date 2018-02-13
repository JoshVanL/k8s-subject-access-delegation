package end_to_end

import (
	//"bufio"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/end_to_end/command"
)

func RunTests(log *logrus.Entry) error {
	var err error

	firstTest, err := command.New("echo", []string{"hello there!"})
	if err != nil {
		return fmt.Errorf("failed to create echo command: %v", err)
	}

	go func() {
		err = firstTest.Run()
	}()

	firstTest.Wait()

	if err != nil {
		fmt.Errorf("error when running command: %v", err)
	}

	fmt.Printf("stdout: '%s'\n", firstTest.Stdout())
	fmt.Printf("stderr: '%s'\n", firstTest.Stderr())

	return nil
}
