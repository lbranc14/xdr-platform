package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/luigi/xdr-platform/api/handlers"
)

// SetupRoutes configure toutes les routes de l'API
func SetupRoutes(app *fiber.App, eventsHandler *handlers.EventsHandler) {
	// Route de health check
	app.Get("/health", eventsHandler.HealthCheck)

	// Groupe API v1
	api := app.Group("/api/v1")

	// Routes pour les événements
	events := api.Group("/events")
	events.Get("/", eventsHandler.GetEvents)                    // GET /api/v1/events
	events.Get("/count", eventsHandler.GetEventCount)          // GET /api/v1/events/count
	events.Get("/stats", eventsHandler.GetEventStats)          // GET /api/v1/events/stats
	events.Get("/filter", eventsHandler.GetFilteredEvents)     // GET /api/v1/events/filter
	events.Get("/timeline", eventsHandler.GetTimeRangeStats)   // GET /api/v1/events/timeline
	
	// Routes pour les statistiques détaillées
	stats := api.Group("/stats")
	stats.Get("/detailed", eventsHandler.GetDetailedStats)     // GET /api/v1/stats/detailed
}
