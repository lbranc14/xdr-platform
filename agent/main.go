package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/luigi/xdr-platform/agent/collectors"
	"github.com/luigi/xdr-platform/agent/config"
	"github.com/luigi/xdr-platform/agent/models"
	"github.com/luigi/xdr-platform/agent/shipper"
	"github.com/luigi/xdr-platform/agent/utils"
)

func main() {
	// Créer le logger
	logger := utils.NewLogger()
	logger.Info("Starting XDR Agent...")

	// Charger la configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration: %v", err)
	}

	// Valider la configuration
	if err := cfg.Validate(); err != nil {
		logger.Fatal("Invalid configuration: %v", err)
	}

	logger.Info("Configuration loaded: %s", cfg.String())

	// Créer le shipper Kafka
	kafkaShipper, err := shipper.NewKafkaShipper(cfg.KafkaBrokers, cfg.KafkaTopicRawEvents, logger)
	if err != nil {
		logger.Fatal("Failed to create Kafka shipper: %v", err)
	}
	defer kafkaShipper.Close()

	// Créer les collecteurs
	var activeCollectors []Collector

	if cfg.EnableSystemCollector {
		systemCollector := collectors.NewSystemCollector(logger, cfg.AgentID, cfg.Hostname)
		activeCollectors = append(activeCollectors, systemCollector)
		logger.Info("System collector enabled")
	}

	if cfg.EnableNetworkCollector {
		networkCollector := collectors.NewNetworkCollector(logger, cfg.AgentID, cfg.Hostname)
		activeCollectors = append(activeCollectors, networkCollector)
		logger.Info("Network collector enabled")
	}

	if cfg.EnableProcessCollector {
		processCollector := collectors.NewProcessCollector(logger, cfg.AgentID, cfg.Hostname)
		activeCollectors = append(activeCollectors, processCollector)
		logger.Info("Process collector enabled")
	}

	if len(activeCollectors) == 0 {
		logger.Fatal("No collectors enabled, please enable at least one collector")
	}

	logger.Info("Started %d collectors", len(activeCollectors))

	// Context pour gérer l'arrêt gracieux
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel pour capturer les signaux d'arrêt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Ticker pour la collecte périodique
	collectionTicker := time.NewTicker(cfg.CollectionInterval)
	defer collectionTicker.Stop()

	// Ticker pour le heartbeat
	heartbeatTicker := time.NewTicker(cfg.HeartbeatInterval)
	defer heartbeatTicker.Stop()

	logger.Info("Agent is running, collecting every %s", cfg.CollectionInterval)
	logger.Info("Press Ctrl+C to stop")

	// Boucle principale
	for {
		select {
		case <-ctx.Done():
			logger.Info("Context cancelled, shutting down...")
			return

		case <-sigChan:
			logger.Info("Received shutdown signal, stopping agent...")
			cancel()
			return

		case <-collectionTicker.C:
			// Collecter et envoyer les événements
			go collectAndShip(activeCollectors, kafkaShipper, logger)

		case <-heartbeatTicker.C:
			// Envoyer un heartbeat
			go sendHeartbeat(cfg, kafkaShipper, logger)
		}
	}
}

// Collector interface pour tous les collecteurs
type Collector interface {
	Collect() ([]*models.Event, error)
}

// collectAndShip collecte les événements de tous les collecteurs et les envoie
func collectAndShip(collectors []Collector, shipper *shipper.KafkaShipper, logger *utils.Logger) {
	logger.Debug("Starting collection cycle...")

	var allEvents []*models.Event

	// Collecter depuis tous les collecteurs
	for _, collector := range collectors {
		events, err := collector.Collect()
		if err != nil {
			logger.Error("Collection failed: %v", err)
			continue
		}
		allEvents = append(allEvents, events...)
	}

	if len(allEvents) == 0 {
		logger.Debug("No events collected in this cycle")
		return
	}

	// Envoyer tous les événements à Kafka
	if err := shipper.Ship(allEvents); err != nil {
		logger.Error("Failed to ship events: %v", err)
		return
	}

	logger.Info("Collection cycle completed: %d events shipped", len(allEvents))
}

// sendHeartbeat envoie un heartbeat pour indiquer que l'agent est actif
func sendHeartbeat(cfg *config.Config, shipper *shipper.KafkaShipper, logger *utils.Logger) {
	logger.Debug("Sending heartbeat...")

	// Créer un événement heartbeat
	heartbeat := &models.Event{
		Timestamp: time.Now(),
		AgentID:   cfg.AgentID,
		Hostname:  cfg.Hostname,
		EventType: models.EventTypeSystem,
		Severity:  models.SeverityLow,
		RawData: map[string]interface{}{
			"heartbeat": true,
			"version":   cfg.AgentVersion,
		},
		Tags: []string{"heartbeat"},
	}

	if err := shipper.Ship([]*models.Event{heartbeat}); err != nil {
		logger.Error("Failed to send heartbeat: %v", err)
	}
}
