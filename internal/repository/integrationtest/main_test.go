package integrationtest

import (
	"os"
	"testing"

	"github.com/redis/rueidis"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	if os.Getenv("TEST_INTEGRATION") != "true" {
		return
	}
	logrus.Infoln("Running integration tests for commonbehaviour...")
	exitCode := m.Run()
	os.Exit(exitCode)
}

func flushAll(t *testing.T, redisClient rueidis.Client) {
	assert.NoError(t, redisClient.Do(t.Context(), redisClient.B().Flushall().Build()).Error())
}
