package integrationtest

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	if os.Getenv("TEST_INTEGRATION") != "true" {
		return
	}
	logrus.Infoln("Running integration tests for matchmaking...")

	exitCode := m.Run()
	os.Exit(exitCode)

}
