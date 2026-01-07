package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/luigi/xdr-platform/api/database"
)

// EventsHandler gère les requêtes liées aux événements
type EventsHandler struct {
	db *database.TimescaleDB
}

// NewEventsHandler crée un nouveau handler pour les événements
func NewEventsHandler(db *database.TimescaleDB) *EventsHandler {
	return &EventsHandler{db: db}
}

// GetEvents retourne une liste paginée d'événements
// GET /api/v1/events?limit=100&offset=0&event_type=system&severity=high
func (h *EventsHandler) GetEvents(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Récupérer les paramètres de pagination
	limit := c.QueryInt("limit", 50)
	if limit > 1000 {
		limit = 1000 // Maximum 1000 événements par requête
	}

	// Récupérer les événements
	events, err := h.db.GetRecentEvents(ctx, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve events",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(events),
		"events":  events,
	})
}

// GetEventCount retourne le nombre total d'événements
// GET /api/v1/events/count
func (h *EventsHandler) GetEventCount(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := h.db.GetEventCount(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get event count",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   count,
	})
}

// GetEventStats retourne des statistiques sur les événements
// GET /api/v1/events/stats
func (h *EventsHandler) GetEventStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Pour l'instant, on retourne juste le count total
	// TODO: Ajouter des stats par type, sévérité, etc.
	count, err := h.db.GetEventCount(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get statistics",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"stats": fiber.Map{
			"total_events": count,
			"last_updated": time.Now(),
		},
	})
}

// HealthCheck vérifie que l'API et la base de données fonctionnent
// GET /api/health
func (h *EventsHandler) HealthCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Vérifier la connexion à la base de données
	err := h.db.HealthCheck(ctx)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "unhealthy",
			"error":  "Database connection failed",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "healthy",
		"timestamp": time.Now(),
		"service": "xdr-api",
	})
}

// GetFilteredEvents retourne des événements filtrés
// GET /api/v1/events/filter?event_type=system&severity=high&limit=50
func (h *EventsHandler) GetFilteredEvents(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Construire les filtres
	filters := make(map[string]interface{})
	
	if eventType := c.Query("event_type"); eventType != "" {
		filters["event_type"] = eventType
	}
	
	if severity := c.Query("severity"); severity != "" {
		filters["severity"] = severity
	}
	
	if hostname := c.Query("hostname"); hostname != "" {
		filters["hostname"] = hostname
	}

	// Filtres temporels
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filters["start_time"] = t
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filters["end_time"] = t
		}
	}

	// Pagination
	limit := c.QueryInt("limit", 50)
	if limit > 1000 {
		limit = 1000
	}
	offset := c.QueryInt("offset", 0)

	// Récupérer les événements filtrés
	events, err := h.db.GetFilteredEvents(ctx, filters, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve filtered events",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(events),
		"filters": filters,
		"events":  events,
	})
}

// GetTimeRangeStats retourne des stats par intervalle de temps
// GET /api/v1/events/timeline?interval=1h&hours=24
func (h *EventsHandler) GetTimeRangeStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	interval := c.Query("interval", "1 hour")
	hours := c.QueryInt("hours", 24)

	stats, err := h.db.GetEventsByTimeRange(ctx, interval, hours)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get timeline stats",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"interval": interval,
		"hours": hours,
		"data": stats,
	})
}

// GetDetailedStats retourne des statistiques détaillées
// GET /api/v1/events/stats/detailed
func (h *EventsHandler) GetDetailedStats(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stats par sévérité
	severityStats, err := h.db.GetStatsBySeverity(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get severity stats",
			"details": err.Error(),
		})
	}

	// Stats par type
	typeStats, err := h.db.GetStatsByType(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get type stats",
			"details": err.Error(),
		})
	}

	// Total count
	totalCount, err := h.db.GetEventCount(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get total count",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"stats": fiber.Map{
			"total_events": totalCount,
			"by_severity": severityStats,
			"by_type": typeStats,
			"last_updated": time.Now(),
		},
	})
}
