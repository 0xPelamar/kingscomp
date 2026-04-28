package integrationtest

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("TEST_INTEGRATION") != "true" {
		return
	}
	fmt.Println("Running integration tests...")
	exitCode := m.Run()
	os.Exit(exitCode)

}
