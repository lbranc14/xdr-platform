package config

import (
	"fmt"
	"os"
)

// Config contient la configuration du service d'ingestion
type Config struct {
	// Database configuration
	DatabaseURL      string
	DatabaseHost     string
	DatabasePort     string
	DatabaseName     string
	DatabaseUser     string
	DatabasePassword string

	// Kafka configuration
	KafkaBrokers        []string
	KafkaTopicRawEvents string
	KafkaGroupID        string

	// Service configuration
	ServiceName    string
	BatchSize      int
	FlushInterval  int // en secondes
	WorkerCount    int

	// Logging
	LogLevel string
}

// LoadConfig charge la configuration depuis les variables d'environnement
func LoadConfig() (*Config, error) {
	config := &Config{
		// Database
		DatabaseHost:     getEnvOrDefault("DATABASE_HOST", "localhost"),
		DatabasePort:     getEnvOrDefault("DATABASE_PORT", "5432"),
		DatabaseName:     getEnvOrDefault("DATABASE_NAME", "xdr_events"),
		DatabaseUser:     getEnvOrDefault("DATABASE_USER", "xdr_admin"),
		DatabasePassword: getEnvOrDefault("DATABASE_PASSWORD", "xdr_secure_password_2024"),

		// Kafka
		KafkaBrokers:        []string{getEnvOrDefault("KAFKA_BROKERS", "localhost:9092")},
		KafkaTopicRawEvents: getEnvOrDefault("KAFKA_TOPIC_RAW_EVENTS", "raw-events"),
		KafkaGroupID:        getEnvOrDefault("KAFKA_GROUP_ID", "xdr-ingestion-service"),

		// Service
		ServiceName:   "ingestion-service",
		BatchSize:     100,
		FlushInterval: 5,
		WorkerCount:   4,

		// Logging
		LogLevel: getEnvOrDefault("LOG_LEVEL", "info"),
	}

	// Construire l'URL de connexion PostgreSQL
	config.DatabaseURL = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.DatabaseUser,
		config.DatabasePassword,
		config.DatabaseHost,
		config.DatabasePort,
		config.DatabaseName,
	)

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
	if c.DatabaseURL == "" {
		return fmt.Errorf("database_url cannot be empty")
	}
	if len(c.KafkaBrokers) == 0 {
		return fmt.Errorf("kafka_brokers cannot be empty")
	}
	if c.KafkaTopicRawEvents == "" {
		return fmt.Errorf("kafka_topic_raw_events cannot be empty")
	}
	return nil
}

// String retourne une représentation string de la config
func (c *Config) String() string {
	return fmt.Sprintf(
		"Service{Name: %s, DB: %s:%s/%s, Kafka: %v, Topic: %s}",
		c.ServiceName,
		c.DatabaseHost,
		c.DatabasePort,
		c.DatabaseName,
		c.KafkaBrokers,
		c.KafkaTopicRawEvents,
	)
}
