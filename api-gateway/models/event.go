package models

import "time"

// EventType représente le type d'événement collecté
type EventType string

const (
	EventTypeSystem  EventType = "system"
	EventTypeNetwork EventType = "network"
	EventTypeProcess EventType = "process"
	EventTypeFile    EventType = "file"
)

// Severity représente la sévérité d'un événement
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Event représente un événement de sécurité collecté
type Event struct {
	Timestamp      time.Time              `json:"timestamp"`
	AgentID        string                 `json:"agent_id"`
	Hostname       string                 `json:"hostname"`
	EventType      EventType              `json:"event_type"`
	Severity       Severity               `json:"severity"`
	RawData        map[string]interface{} `json:"raw_data"`
	SourceIP       string                 `json:"source_ip,omitempty"`
	DestinationIP  string                 `json:"destination_ip,omitempty"`
	ProcessName    string                 `json:"process_name,omitempty"`
	ProcessPID     int                    `json:"process_pid,omitempty"`
	Username       string                 `json:"username,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}
