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

// SystemEvent représente un événement système
type SystemEvent struct {
	EventName   string                 `json:"event_name"`
	EventID     int                    `json:"event_id"`
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
}

// NetworkEvent représente un événement réseau
type NetworkEvent struct {
	Protocol      string `json:"protocol"`
	SourceIP      string `json:"source_ip"`
	SourcePort    int    `json:"source_port"`
	DestIP        string `json:"dest_ip"`
	DestPort      int    `json:"dest_port"`
	BytesSent     int64  `json:"bytes_sent"`
	BytesReceived int64  `json:"bytes_received"`
	State         string `json:"state"`
}

// ProcessEvent représente un événement de processus
type ProcessEvent struct {
	PID             int      `json:"pid"`
	Name            string   `json:"name"`
	CommandLine     string   `json:"command_line"`
	ExecutablePath  string   `json:"executable_path"`
	ParentPID       int      `json:"parent_pid"`
	Username        string   `json:"username"`
	CPUPercent      float64  `json:"cpu_percent"`
	MemoryPercent   float64  `json:"memory_percent"`
	MemoryBytes     uint64   `json:"memory_bytes"`
	CreateTime      int64    `json:"create_time"`
	NumThreads      int32    `json:"num_threads"`
	Status          string   `json:"status"`
	OpenFiles       []string `json:"open_files,omitempty"`
	Connections     int      `json:"connections"`
}

// AgentInfo représente les informations de l'agent
type AgentInfo struct {
	AgentID       string    `json:"agent_id"`
	Hostname      string    `json:"hostname"`
	IPAddress     string    `json:"ip_address"`
	OSType        string    `json:"os_type"`
	OSVersion     string    `json:"os_version"`
	AgentVersion  string    `json:"agent_version"`
	Status        string    `json:"status"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}
