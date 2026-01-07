package config

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

// Config contient toute la configuration de l'agent
type Config struct {
	// Agent configuration
	AgentID              string
	AgentVersion         string
	Hostname             string
	CollectionInterval   time.Duration
	HeartbeatInterval    time.Duration

	// Kafka configuration
	KafkaBrokers         []string
	KafkaTopicRawEvents  string
	KafkaGroupID         string

	// Collectors configuration
	EnableSystemCollector  bool
	EnableNetworkCollector bool
	EnableProcessCollector bool

	// Logging
	LogLevel string
	LogFile  string

	// Performance
	MaxEventsPerBatch int
	BufferSize        int
}

// LoadConfig charge la configuration depuis les variables d'environnement
func LoadConfig() (*Config, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}

	// Générer un AgentID unique s'il n'existe pas
	agentID := os.Getenv("AGENT_ID")
	if agentID == "" || agentID == "auto-generated" {
		agentID = fmt.Sprintf("agent-%s", uuid.New().String()[:8])
	}

	// Collection interval
	collectionInterval, err := time.ParseDuration(getEnvOrDefault("AGENT_COLLECTION_INTERVAL", "30s"))
	if err != nil {
		collectionInterval = 30 * time.Second
	}

	// Heartbeat interval
	heartbeatInterval, err := time.ParseDuration(getEnvOrDefault("AGENT_HEARTBEAT_INTERVAL", "60s"))
	if err != nil {
		heartbeatInterval = 60 * time.Second
	}

	config := &Config{
		AgentID:              agentID,
		AgentVersion:         getEnvOrDefault("AGENT_VERSION", "1.0.0"),
		Hostname:             hostname,
		CollectionInterval:   collectionInterval,
		HeartbeatInterval:    heartbeatInterval,

		// Kafka configuration
		KafkaBrokers:         []string{getEnvOrDefault("KAFKA_BROKERS", "localhost:9092")},
		KafkaTopicRawEvents:  getEnvOrDefault("KAFKA_TOPIC_RAW_EVENTS", "raw-events"),
		KafkaGroupID:         getEnvOrDefault("KAFKA_GROUP_ID", "xdr-agent-group"),

		// Collectors
		EnableSystemCollector:  getEnvOrDefault("ENABLE_SYSTEM_COLLECTOR", "true") == "true",
		EnableNetworkCollector: getEnvOrDefault("ENABLE_NETWORK_COLLECTOR", "true") == "true",
		EnableProcessCollector: getEnvOrDefault("ENABLE_PROCESS_COLLECTOR", "true") == "true",

		// Logging
		LogLevel: getEnvOrDefault("LOG_LEVEL", "info"),
		LogFile:  getEnvOrDefault("LOG_FILE", ""),

		// Performance
		MaxEventsPerBatch: 100,
		BufferSize:        1000,
	}

	return config, nil
}

// getEnvOrDefault retourne la valeur d'une variable d'environnement ou une valeur par défaut
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Validate valide la configuration
func (c *Config) Validate() error {
	if c.AgentID == "" {
		return fmt.Errorf("agent_id cannot be empty")
	}
	if c.Hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}
	if len(c.KafkaBrokers) == 0 {
		return fmt.Errorf("kafka_brokers cannot be empty")
	}
	if c.CollectionInterval <= 0 {
		return fmt.Errorf("collection_interval must be positive")
	}
	return nil
}

// String retourne une représentation string de la config (pour logging)
func (c *Config) String() string {
	return fmt.Sprintf(
		"Agent{ID: %s, Hostname: %s, Version: %s, Interval: %s}",
		c.AgentID,
		c.Hostname,
		c.AgentVersion,
		c.CollectionInterval,
	)
}
