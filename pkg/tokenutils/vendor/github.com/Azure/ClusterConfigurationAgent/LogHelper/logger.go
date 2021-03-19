package LogHelper

import (
	"fmt"
	"io"
	"os"
)

const (
	SecureStringLogFilePath = "/var/log/token"
)

type LogWriter interface {
	Info(format string, params ...interface{})
	Error(format string, params ...interface{})
	LogSecureString(format string, params ...interface{})
}

// Using Logger as the logging class for CorrelationId and potentially extensible in the future.
// For logs without CorrelationId we are still using golang log class. TODO: Change to klog
type Logger struct {
	InfoWriter  io.Writer
	ErrorWriter io.Writer
}

func (a *Logger) InitializeLogger(armId string, location string, correlationId string, agentType string, agentName string, agentVersion string) {
	GenevaInfoWriter := &GenevaInfoWriter{}
	GenevaInfoWriter.Initialize(armId, location, correlationId, agentType, agentName, agentVersion)
	GenevaErrorWriter := &GenevaErrorWriter{}
	GenevaErrorWriter.Initialize(armId, location, correlationId, agentType, agentName, agentVersion)

	a.InfoWriter = io.MultiWriter(GenevaInfoWriter)
	a.ErrorWriter = io.MultiWriter(GenevaErrorWriter)
}

func (a Logger) Info(format string, params ...interface{}) {
	if len(params) < 1 {
		_, _ = a.InfoWriter.Write([]byte(format))

	} else {
		_, _ = a.InfoWriter.Write([]byte(fmt.Sprintf(format, params...)))
	}
}

func (a Logger) Error(format string, params ...interface{}) {
	if len(params) < 1 {
		_, _ = a.ErrorWriter.Write([]byte(format))

	} else {
		_, _ = a.ErrorWriter.Write([]byte(fmt.Sprintf(format, params...)))
	}
}

func (a Logger) LogSecureString(format string, params ...interface{}) {
	format = fmt.Sprintf("%s\n", format)
	message := format
	if len(params) > 1 {
		message = fmt.Sprintf(format, params...)
	}
	f, err := os.OpenFile(SecureStringLogFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(message)
	_ = f.Close()
}
