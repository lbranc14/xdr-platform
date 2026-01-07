package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/luigi/xdr-platform/api/models"
)

// TimescaleDB gère la connexion à la base de données
type TimescaleDB struct {
	db *sql.DB
}

// NewTimescaleDB crée une nouvelle connexion à TimescaleDB
func NewTimescaleDB(databaseURL string) (*TimescaleDB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configurer le pool de connexions
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Tester la connexion
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &TimescaleDB{db: db}, nil
}

// InsertEvents insère un batch d'événements dans la base de données
func (ts *TimescaleDB) InsertEvents(ctx context.Context, events []*models.Event) error {
	if len(events) == 0 {
		return nil
	}

	// Préparer la requête d'insertion
	query := `
		INSERT INTO raw_events (
			timestamp, agent_id, hostname, event_type, severity,
			raw_data, source_ip, destination_ip, process_name,
			process_pid, username, tags, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	// Préparer la transaction
	tx, err := ts.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Préparer le statement
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Insérer tous les événements
	for _, event := range events {
		// Convertir raw_data en JSONB
		rawDataJSON, err := json.Marshal(event.RawData)
		if err != nil {
			return fmt.Errorf("failed to marshal raw_data: %w", err)
		}

		// Convertir metadata en JSONB (peut être nil)
		var metadataJSON interface{}
		if event.Metadata != nil && len(event.Metadata) > 0 {
			data, err := json.Marshal(event.Metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal metadata: %w", err)
			}
			metadataJSON = data
		} else {
			metadataJSON = nil
		}

		// Convertir les valeurs nullables
		var sourceIP, destIP, processName, username interface{}
		var processPID interface{}

		if event.SourceIP != "" {
			sourceIP = event.SourceIP
		}
		if event.DestinationIP != "" {
			destIP = event.DestinationIP
		}
		if event.ProcessName != "" {
			processName = event.ProcessName
		}
		if event.ProcessPID != 0 {
			processPID = event.ProcessPID
		}
		if event.Username != "" {
			username = event.Username
		}

		// Exécuter l'insertion
		_, err = stmt.ExecContext(
			ctx,
			event.Timestamp,
			event.AgentID,
			event.Hostname,
			event.EventType,
			event.Severity,
			rawDataJSON,
			sourceIP,
			destIP,
			processName,
			processPID,
			username,
			pq.Array(event.Tags),
			metadataJSON,
		)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	// Commit la transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetEventCount retourne le nombre total d'événements
func (ts *TimescaleDB) GetEventCount(ctx context.Context) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM raw_events"
	err := ts.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get event count: %w", err)
	}
	return count, nil
}

// GetRecentEvents retourne les N derniers événements
func (ts *TimescaleDB) GetRecentEvents(ctx context.Context, limit int) ([]*models.Event, error) {
	query := `
		SELECT 
			timestamp, agent_id, hostname, event_type, severity,
			raw_data, source_ip, destination_ip, process_name,
			process_pid, username, tags, metadata
		FROM raw_events
		ORDER BY timestamp DESC
		LIMIT $1
	`

	rows, err := ts.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		var rawDataJSON, metadataJSON []byte
		var sourceIP, destIP, processName, username sql.NullString
		var processPID sql.NullInt64

		err := rows.Scan(
			&event.Timestamp,
			&event.AgentID,
			&event.Hostname,
			&event.EventType,
			&event.Severity,
			&rawDataJSON,
			&sourceIP,
			&destIP,
			&processName,
			&processPID,
			&username,
			pq.Array(&event.Tags),
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		// Décoder raw_data
		if err := json.Unmarshal(rawDataJSON, &event.RawData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal raw_data: %w", err)
		}

		// Décoder metadata si présent
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		// Assigner les valeurs nullables
		if sourceIP.Valid {
			event.SourceIP = sourceIP.String
		}
		if destIP.Valid {
			event.DestinationIP = destIP.String
		}
		if processName.Valid {
			event.ProcessName = processName.String
		}
		if processPID.Valid {
			event.ProcessPID = int(processPID.Int64)
		}
		if username.Valid {
			event.Username = username.String
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, nil
}

// Close ferme la connexion à la base de données
func (ts *TimescaleDB) Close() error {
	if ts.db != nil {
		return ts.db.Close()
	}
	return nil
}

// HealthCheck vérifie que la base de données est accessible
func (ts *TimescaleDB) HealthCheck(ctx context.Context) error {
	return ts.db.PingContext(ctx)
}

// GetFilteredEvents retourne les événements filtrés par critères
func (ts *TimescaleDB) GetFilteredEvents(ctx context.Context, filters map[string]interface{}, limit int, offset int) ([]*models.Event, error) {
	query := `
		SELECT 
			timestamp, agent_id, hostname, event_type, severity,
			raw_data, source_ip, destination_ip, process_name,
			process_pid, username, tags, metadata
		FROM raw_events
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	// Filtre par type d'événement
	if eventType, ok := filters["event_type"].(string); ok && eventType != "" {
		query += fmt.Sprintf(" AND event_type = $%d", argPos)
		args = append(args, eventType)
		argPos++
	}

	// Filtre par sévérité
	if severity, ok := filters["severity"].(string); ok && severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", argPos)
		args = append(args, severity)
		argPos++
	}

	// Filtre par période
	if startTime, ok := filters["start_time"].(time.Time); ok {
		query += fmt.Sprintf(" AND timestamp >= $%d", argPos)
		args = append(args, startTime)
		argPos++
	}

	if endTime, ok := filters["end_time"].(time.Time); ok {
		query += fmt.Sprintf(" AND timestamp <= $%d", argPos)
		args = append(args, endTime)
		argPos++
	}

	// Filtre par hostname
	if hostname, ok := filters["hostname"].(string); ok && hostname != "" {
		query += fmt.Sprintf(" AND hostname = $%d", argPos)
		args = append(args, hostname)
		argPos++
	}

	query += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := ts.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query filtered events: %w", err)
	}
	defer rows.Close()

	return ts.scanEvents(rows)
}

// GetEventsByTimeRange retourne les événements groupés par intervalle de temps
func (ts *TimescaleDB) GetEventsByTimeRange(ctx context.Context, interval string, hours int) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		SELECT 
			time_bucket('%s', timestamp) AS bucket,
			event_type,
			severity,
			COUNT(*) as count
		FROM raw_events
		WHERE timestamp > NOW() - INTERVAL '%d hours'
		GROUP BY bucket, event_type, severity
		ORDER BY bucket DESC
	`, interval, hours)

	rows, err := ts.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query time range stats: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var bucket time.Time
		var eventType, severity string
		var count int

		if err := rows.Scan(&bucket, &eventType, &severity, &count); err != nil {
			return nil, fmt.Errorf("failed to scan time range row: %w", err)
		}

		results = append(results, map[string]interface{}{
			"timestamp":  bucket,
			"event_type": eventType,
			"severity":   severity,
			"count":      count,
		})
	}

	return results, nil
}

// GetStatsBySeverity retourne les stats par sévérité
func (ts *TimescaleDB) GetStatsBySeverity(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT severity, COUNT(*) as count
		FROM raw_events
		GROUP BY severity
	`

	rows, err := ts.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query severity stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var severity string
		var count int

		if err := rows.Scan(&severity, &count); err != nil {
			return nil, fmt.Errorf("failed to scan severity row: %w", err)
		}

		stats[severity] = count
	}

	return stats, nil
}

// GetStatsByType retourne les stats par type
func (ts *TimescaleDB) GetStatsByType(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT event_type, COUNT(*) as count
		FROM raw_events
		GROUP BY event_type
	`

	rows, err := ts.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query type stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var eventType string
		var count int

		if err := rows.Scan(&eventType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan type row: %w", err)
		}

		stats[eventType] = count
	}

	return stats, nil
}

// scanEvents est une fonction helper pour scanner les événements
func (ts *TimescaleDB) scanEvents(rows *sql.Rows) ([]*models.Event, error) {
	var events []*models.Event
	for rows.Next() {
		event := &models.Event{}
		var rawDataJSON, metadataJSON []byte
		var sourceIP, destIP, processName, username sql.NullString
		var processPID sql.NullInt64

		err := rows.Scan(
			&event.Timestamp,
			&event.AgentID,
			&event.Hostname,
			&event.EventType,
			&event.Severity,
			&rawDataJSON,
			&sourceIP,
			&destIP,
			&processName,
			&processPID,
			&username,
			pq.Array(&event.Tags),
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		// Décoder raw_data
		if err := json.Unmarshal(rawDataJSON, &event.RawData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal raw_data: %w", err)
		}

		// Décoder metadata si présent
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		// Assigner les valeurs nullables
		if sourceIP.Valid {
			event.SourceIP = sourceIP.String
		}
		if destIP.Valid {
			event.DestinationIP = destIP.String
		}
		if processName.Valid {
			event.ProcessName = processName.String
		}
		if processPID.Valid {
			event.ProcessPID = int(processPID.Int64)
		}
		if username.Valid {
			event.Username = username.String
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, nil
}
