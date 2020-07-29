package devops

import log "github.com/sirupsen/logrus"

func DevOpsLogger() *log.Logger {
	logger := log.New()
	logger.Formatter = &log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "00:00:00 MST",
	}
	logger.SetLevel(log.DebugLevel)
	return logger
}
