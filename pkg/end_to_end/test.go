package end_to_end

import (
	//"bufio"
	//"fmt"

	"github.com/sirupsen/logrus"
)

func RunTests(log *logrus.Entry) error {
	firstTest, err := NewCommand("echo", []string{"hello!"})
	if err != nil {
		return fmt.Errorf("failed to create echo command: %v", err)
	}

	if err := firstTest.Run(); err != nil {
		return err
	}

	//if err := firstTest.Start(); err != nil {
	//	return fmt.Errorf("error starting echo command: %v", err)
	//}

	//stdout, err := firstTest.StdoutScan()
	//if err != nil {
	//	return fmt.Errorf("failed to get command stdout: %v", err)
	//}

	//log.Infof("stdout: '%s'", stdout)

	//scanner := bufio.NewScanner(firstTest.Stdout)
	//scanner.Split(bufio.ScanWords)
	//for scanner.Scan() {
	//	m := scanner.Text()
	//	fmt.Printf(">>%s\n", m)
	//}

	//if err := firstTest.Wait(); err != nil {
	//	return fmt.Errorf("error waiting echo command: %v", err)
	//}

	//log.Infof("All command run successfully!")

	return nil
}
