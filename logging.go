package graphql

import (
	"os"

	stackdriver "github.com/icco/logrus-stackdriver-formatter"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// InitLogging initializes a logger to send things to stackdriver.
func InitLogging() *logrus.Logger {
	log.Formatter = stackdriver.NewFormatter()
	log.Level = logrus.DebugLevel
	log.SetOutput(os.Stdout)

	log.Info("Logger successfully initialised!")

	return log
}
