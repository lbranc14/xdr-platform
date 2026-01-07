package shipper

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	"github.com/luigi/xdr-platform/agent/models"
	"github.com/luigi/xdr-platform/agent/utils"
)

// KafkaShipper envoie les événements vers Kafka
type KafkaShipper struct {
	writer *kafka.Writer
	logger *utils.Logger
	topic  string
}

// NewKafkaShipper crée un nouveau shipper Kafka
func NewKafkaShipper(brokers []string, topic string, logger *utils.Logger) (*KafkaShipper, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers list is empty")
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	logger.Info("Kafka shipper initialized with brokers: %v, topic: %s", brokers, topic)

	return &KafkaShipper{
		writer: writer,
		logger: logger,
		topic:  topic,
	}, nil
}

// Ship envoie un lot d'événements vers Kafka
func (ks *KafkaShipper) Ship(events []*models.Event) error {
	if len(events) == 0 {
		return nil
	}

	messages := make([]kafka.Message, 0, len(events))

	for _, event := range events {
		// Sérialiser l'événement en JSON
		data, err := json.Marshal(event)
		if err != nil {
			ks.logger.Error("Failed to marshal event: %v", err)
			continue
		}

		// Créer le message Kafka
		message := kafka.Message{
			Key:   []byte(event.AgentID),
			Value: data,
		}

		messages = append(messages, message)
	}

	// Envoyer tous les messages en batch
	err := ks.writer.WriteMessages(context.Background(), messages...)
	if err != nil {
		return fmt.Errorf("failed to write messages to kafka: %w", err)
	}

	ks.logger.Info("Successfully shipped %d events to Kafka topic '%s'", len(messages), ks.topic)
	return nil
}

// Close ferme la connexion Kafka
func (ks *KafkaShipper) Close() error {
	if ks.writer != nil {
		ks.logger.Info("Closing Kafka writer...")
		return ks.writer.Close()
	}
	return nil
}
