package LogHelper

import (
	"fmt"
	"os"
	"strings"
)

// Both GenevaInfoWriter and GenevaErrorWriter are used by logger class to format the logs with correlationId.
// This is extensible in the future to be a multiwriter used to write to stdout/stderr and a file likely when FluentD changes are made.
type GenevaInfoWriter struct {
	*Serializer
}

func (s *GenevaInfoWriter) Initialize(armId string, location string, correlationId string, agentType string, agentName string, agentVersion string) {
	s.Serializer = &Serializer{
		Armid:         armId,
		AgentType:     agentType,
		Location:      location,
		LogType:       TraceType,
		LogLevel:      InformationLevel,
		CorrelationId: correlationId,
		AgentName:     agentName,
		AgentVersion:  agentVersion,
	}
}

func (s *GenevaInfoWriter) Write(p []byte) (n int, err error) {
	escapedMsg := strings.TrimSpace(string(p))
	genevaFormattedValue := fmt.Sprintf("%s\n", s.Format(escapedMsg))
	n, err = fmt.Fprint(os.Stdout, genevaFormattedValue)
	msg := "Error in infoWriter when printing geneva formatted value: %v"
	if err != nil {
		fmt.Printf(s.Format(msg))
		return n, fmt.Errorf(msg, err)
	}
	return len(p), nil
}

type GenevaErrorWriter struct {
	*Serializer
}

func (s *GenevaErrorWriter) Initialize(armId string, location string, correlationId string, agentType string, agentName string, agentVersion string) {
	s.Serializer = &Serializer{
		Armid:         armId,
		AgentType:     agentType,
		Location:      location,
		LogType:       TraceType,
		LogLevel:      ErrorLevel,
		CorrelationId: correlationId,
		AgentName:     agentName,
		AgentVersion:  agentVersion,
	}
}

func (s *GenevaErrorWriter) Write(p []byte) (n int, err error) {
	escapedMsg := strings.TrimSpace(string(p))
	genevaFormattedValue := fmt.Sprintf("%s\n", s.Format(escapedMsg))
	n, err = fmt.Fprint(os.Stderr, genevaFormattedValue)
	msg := fmt.Sprintf("Error in errorWriter when printing geneva formatted value: %v", err)
	if err != nil {
		fmt.Printf(s.Format(msg))
		return n, fmt.Errorf(msg)
	}
	return len(p), nil
}
