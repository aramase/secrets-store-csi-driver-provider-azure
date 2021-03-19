package LogHelper

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// These are expected in the service to be in this EXACT format
const (
	TraceType     = "Trace"
	ExceptionType = "Exception"
)

// These are expected in the service to be in this EXACT format
// Update AGENT_TYPE in the fluent bit container spec in the charts for all agents if any changes are made here.
const (
	ConfigAgentType    = "ConfigAgent"
	ConnectAgentType   = "ConnectAgent"
	ApplianceAgentType = "ApplianceAgent"
)

// These are expected in the service to be in this EXACT format
// Update AGENT_NAME in the fluent bit container spec in the charts for all agents if any changes are made here.
const (
	ConfigAgentName           = "ConfigAgent"
	ControllerManager         = "ControllerManager"
	ConnectAgentName          = "ConnectAgent"
	MetricsAgentName          = "MetricsAgent"
	ClusterMetadataAgentName  = "ClusterMetadataAgent"
	ApplianceConnectAgentName = "ApplianceConnectAgent"
)

const (
	VerboseLevel     = "Verbose"
	WarningLevel     = "Warning"
	InformationLevel = "Information"
	ErrorLevel       = "Error"
)

const (
	Role            = "ClusterConfigAgent"
	ProdEnvironment = "prod"
)

type LogMessage struct {
	Message        string `json:"Message"`
	LogType        string `json:"LogType"`
	LogLevel       string `json:"LogLevel"`
	Environment    string `json:"Environment"`
	Role           string `json:"Role"` //
	Location       string `json:"Location"`
	ArmId          string `json:"ArmId"`
	CorrelationId  string `json:"CorrelationId"`
	AgentName      string `json:"AgentName"`
	AgentVersion   string `json:"AgentVersion"`
	AgentTimestamp string `json:"AgentTimestamp"`
}

type Serializer struct {
	Armid         string
	Location      string
	CorrelationId string
	LogType       string
	LogLevel      string
	AgentType     string
	AgentName     string
	AgentVersion  string
}

func (s *Serializer) Format(message string) string {
	logmessage := LogMessage{}
	// Later try to gather the Log level based on regex parsing of the logs to understand the error
	logmessage.LogLevel = s.LogLevel
	logmessage.LogType = s.AgentType + s.LogType
	logmessage.CorrelationId = s.CorrelationId

	logmessage.Location = s.Location
	logmessage.Environment = ProdEnvironment
	logmessage.Message = message
	logmessage.Role = Role

	logmessage.ArmId = s.Armid
	logmessage.AgentName = s.AgentName
	logmessage.AgentVersion = s.AgentVersion
	logmessage.AgentTimestamp = fmt.Sprintf("%s", time.Now().Format("2006/01/02 15:04:05"))

	result, err := json.Marshal(logmessage)
	if err != nil {
		log.Printf("Unable to serialize the log %v", err)
		return ""
	}

	return string(result)
}
