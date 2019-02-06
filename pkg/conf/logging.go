package conf

import (
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

// ConfigureLogging will take the logging configuration and also adds
// a few default parameters
func ConfigureLogging(config *LoggingConfig) (*logrus.Entry, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(config.ReportCaller)
	logrus.SetOutput(os.Stdout)

	level, err := logrus.ParseLevel(strings.ToUpper(config.Level))
	if err != nil {
		return nil, err
	}
	logrus.SetLevel(level)

	// always use the fulltimestamp
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		DisableTimestamp: false,
		ForceColors:      true,
	})

	return logrus.StandardLogger().WithField("hostname", hostname), nil
}
