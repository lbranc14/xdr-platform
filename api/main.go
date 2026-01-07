package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/luigi/xdr-platform/api/config"
	"github.com/luigi/xdr-platform/api/database"
	"github.com/luigi/xdr-platform/api/ingestion"
)

func main() {
	// Logger
	logger := log.New(os.Stdout, "[INGESTION] ", log.LstdFlags|log.Lshortfile)
	logger.Println("Starting XDR Ingestion Service...")

	// Charger la configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Valider la configuration
	if err := cfg.Validate(); err != nil {
		logger.Fatalf("Invalid configuration: %v", err)
	}

	logger.Printf("Configuration loaded: %s", cfg.String())

	// Connexion à TimescaleDB
	logger.Println("Connecting to TimescaleDB...")
	db, err := database.NewTimescaleDB(cfg.DatabaseURL)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Vérifier la connexion
	ctx := context.Background()
	if err := db.HealthCheck(ctx); err != nil {
		logger.Fatalf("Database health check failed: %v", err)
	}
	logger.Println("Connected to TimescaleDB successfully")

	// Afficher le nombre d'événements existants
	count, err := db.GetEventCount(ctx)
	if err != nil {
		logger.Printf("Warning: failed to get event count: %v", err)
	} else {
		logger.Printf("Current event count in database: %d", count)
	}

	// Créer le consumer Kafka
	logger.Println("Creating Kafka consumer...")
	consumer, err := ingestion.NewConsumer(
		cfg.KafkaBrokers,
		cfg.KafkaTopicRawEvents,
		cfg.KafkaGroupID,
		db,
		cfg.BatchSize,
		cfg.FlushInterval,
		cfg.WorkerCount,
		logger,
	)
	if err != nil {
		logger.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	logger.Printf("Kafka consumer created: brokers=%v, topic=%s, group=%s",
		cfg.KafkaBrokers, cfg.KafkaTopicRawEvents, cfg.KafkaGroupID)

	// Context avec annulation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel pour capturer les signaux d'arrêt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Démarrer le consumer dans une goroutine
	go func() {
		if err := consumer.Start(ctx); err != nil {
			logger.Printf("Consumer error: %v", err)
		}
	}()

	// Afficher les métriques périodiquement
	metricsTicker := time.NewTicker(30 * time.Second)
	defer metricsTicker.Stop()

	logger.Println("Ingestion service is running...")
	logger.Println("Press Ctrl+C to stop")

	// Boucle principale
	for {
		select {
		case <-sigChan:
			logger.Println("Received shutdown signal, stopping service...")
			cancel()

			// Arrêter le consumer proprement
			if err := consumer.Stop(); err != nil {
				logger.Printf("Error stopping consumer: %v", err)
			}

			// Afficher les métriques finales
			metrics := consumer.GetMetrics()
			logger.Printf("Final metrics: %+v", metrics)

			logger.Println("Service stopped")
			return

		case <-metricsTicker.C:
			// Afficher les métriques
			metrics := consumer.GetMetrics()
			logger.Printf("Metrics: events_processed=%d, events_inserted=%d, errors=%d, buffer_size=%d",
				metrics["events_processed"],
				metrics["events_inserted"],
				metrics["errors"],
				metrics["buffer_size"],
			)

			// Afficher le nombre total d'événements
			count, err := db.GetEventCount(ctx)
			if err == nil {
				logger.Printf("Total events in database: %d", count)
			}
		}
	}
}
