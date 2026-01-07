package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/luigi/xdr-platform/api/database"
	"github.com/luigi/xdr-platform/api/models"
)

// Consumer consomme les événements depuis Kafka et les stocke dans TimescaleDB
type Consumer struct {
	reader    *kafka.Reader
	db        *database.TimescaleDB
	batchSize int
	flushInterval time.Duration
	workerCount int
	logger    *log.Logger
	
	eventBuffer   []*models.Event
	bufferMutex   sync.Mutex
	stopChan      chan struct{}
	wg            sync.WaitGroup
	
	// Métriques
	eventsProcessed uint64
	eventsInserted  uint64
	errorCount      uint64
}

// NewConsumer crée un nouveau consumer Kafka
func NewConsumer(brokers []string, topic, groupID string, db *database.TimescaleDB, batchSize int, flushInterval int, workerCount int, logger *log.Logger) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	return &Consumer{
		reader:        reader,
		db:            db,
		batchSize:     batchSize,
		flushInterval: time.Duration(flushInterval) * time.Second,
		workerCount:   workerCount,
		logger:        logger,
		eventBuffer:   make([]*models.Event, 0, batchSize),
		stopChan:      make(chan struct{}),
	}, nil
}

// Start démarre le consumer
func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Printf("Starting consumer with %d workers, batch size: %d, flush interval: %s",
		c.workerCount, c.batchSize, c.flushInterval)

	// Démarrer les workers
	for i := 0; i < c.workerCount; i++ {
		c.wg.Add(1)
		go c.worker(ctx, i)
	}

	// Démarrer le flusher périodique
	c.wg.Add(1)
	go c.periodicFlusher(ctx)

	// Attendre que tous les workers se terminent
	c.wg.Wait()
	return nil
}

// worker lit les messages Kafka et les traite
func (c *Consumer) worker(ctx context.Context, workerID int) {
	defer c.wg.Done()
	c.logger.Printf("Worker %d started", workerID)

	for {
		select {
		case <-ctx.Done():
			c.logger.Printf("Worker %d stopping (context cancelled)", workerID)
			return
		case <-c.stopChan:
			c.logger.Printf("Worker %d stopping (stop signal)", workerID)
			return
		default:
			// Lire un message depuis Kafka
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
				c.logger.Printf("Worker %d: Error fetching message: %v", workerID, err)
				c.errorCount++
				time.Sleep(time.Second)
				continue
			}

			// Décoder l'événement
			var event models.Event
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				c.logger.Printf("Worker %d: Error unmarshaling event: %v", workerID, err)
				c.errorCount++
				// Commit quand même le message pour ne pas le retraiter
				c.reader.CommitMessages(ctx, msg)
				continue
			}

			// Ajouter au buffer
			c.addToBuffer(&event)
			c.eventsProcessed++

			// Commit le message
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Printf("Worker %d: Error committing message: %v", workerID, err)
			}

			// Flush si le buffer est plein
			if c.bufferSize() >= c.batchSize {
				c.flush(ctx)
			}
		}
	}
}

// periodicFlusher flush le buffer périodiquement
func (c *Consumer) periodicFlusher(ctx context.Context) {
	defer c.wg.Done()
	ticker := time.NewTicker(c.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Dernier flush avant de quitter
			c.flush(ctx)
			return
		case <-c.stopChan:
			c.flush(ctx)
			return
		case <-ticker.C:
			c.flush(ctx)
		}
	}
}

// addToBuffer ajoute un événement au buffer de manière thread-safe
func (c *Consumer) addToBuffer(event *models.Event) {
	c.bufferMutex.Lock()
	defer c.bufferMutex.Unlock()
	c.eventBuffer = append(c.eventBuffer, event)
}

// bufferSize retourne la taille actuelle du buffer
func (c *Consumer) bufferSize() int {
	c.bufferMutex.Lock()
	defer c.bufferMutex.Unlock()
	return len(c.eventBuffer)
}

// flush insère tous les événements du buffer dans la base de données
func (c *Consumer) flush(ctx context.Context) {
	c.bufferMutex.Lock()
	if len(c.eventBuffer) == 0 {
		c.bufferMutex.Unlock()
		return
	}

	// Copier le buffer et le vider
	eventsToInsert := make([]*models.Event, len(c.eventBuffer))
	copy(eventsToInsert, c.eventBuffer)
	c.eventBuffer = c.eventBuffer[:0]
	c.bufferMutex.Unlock()

	// Insérer dans la base de données
	start := time.Now()
	if err := c.db.InsertEvents(ctx, eventsToInsert); err != nil {
		c.logger.Printf("Error inserting events: %v", err)
		c.errorCount++
		return
	}

	duration := time.Since(start)
	c.eventsInserted += uint64(len(eventsToInsert))
	c.logger.Printf("Flushed %d events to database in %s (total inserted: %d)",
		len(eventsToInsert), duration, c.eventsInserted)
}

// Stop arrête le consumer de manière gracieuse
func (c *Consumer) Stop() error {
	c.logger.Println("Stopping consumer...")
	close(c.stopChan)
	c.wg.Wait()

	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("error closing Kafka reader: %w", err)
	}

	c.logger.Printf("Consumer stopped. Events processed: %d, inserted: %d, errors: %d",
		c.eventsProcessed, c.eventsInserted, c.errorCount)
	return nil
}

// GetMetrics retourne les métriques du consumer
func (c *Consumer) GetMetrics() map[string]uint64 {
	return map[string]uint64{
		"events_processed": c.eventsProcessed,
		"events_inserted":  c.eventsInserted,
		"errors":           c.errorCount,
		"buffer_size":      uint64(c.bufferSize()),
	}
}
