package output

import (
	"github.com/Sirupsen/logrus"
)

// LoggerModel ...
type LoggerModel struct {
	Logger *logrus.Logger
	Silent bool
}

// NewLogger ...
func NewLogger(silent bool) LoggerModel {
	return LoggerModel{
		Logger: logrus.New(),
		Silent: silent,
	}
}

// Info ...
func (logger *LoggerModel) Info(args ...interface{}) {
	if !logger.Silent {
		logger.Logger.Info(args)
	}
}

// Infof ...
func (logger *LoggerModel) Infof(format string, args ...interface{}) {
	if !logger.Silent {
		logger.Logger.Infof(format, args)
	}
}

// Warn ...
func (logger *LoggerModel) Warn(args ...interface{}) {
	if !logger.Silent {
		logger.Logger.Warn(args)
	}
}

// Debugf ...
func (logger *LoggerModel) Debugf(format string, args ...interface{}) {
	if !logger.Silent {
		logger.Logger.Debugf(format, args)
	}
}

// Errorf ...
func (logger *LoggerModel) Errorf(format string, args ...interface{}) {
	if !logger.Silent {
		logger.Logger.Errorf(format, args)
	}
}
