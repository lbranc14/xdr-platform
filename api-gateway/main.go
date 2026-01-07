package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	
	"github.com/luigi/xdr-platform/api/config"
	"github.com/luigi/xdr-platform/api/database"
	"github.com/luigi/xdr-platform/api/handlers"
	"github.com/luigi/xdr-platform/api/routes"
)

func main() {
	// Logger
	logger := log.New(os.Stdout, "[API] ", log.LstdFlags)
	logger.Println("Starting XDR API Gateway...")

	// Charger la configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

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

	logger.Println("Connected to TimescaleDB successfully")

	// Créer l'application Fiber
	app := fiber.New(fiber.Config{
		AppName: "XDR API Gateway v1.0",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Créer les handlers
	eventsHandler := handlers.NewEventsHandler(db)

	// Configurer les routes
	routes.SetupRoutes(app, eventsHandler)

	// Route par défaut
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "XDR API Gateway",
			"version": "1.0",
			"status": "running",
			"endpoints": fiber.Map{
				"health": "/health",
				"events": "/api/v1/events",
				"count":  "/api/v1/events/count",
				"stats":  "/api/v1/events/stats",
			},
		})
	})

	// Channel pour capturer les signaux d'arrêt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Démarrer le serveur dans une goroutine
	go func() {
		port := os.Getenv("API_PORT")
		if port == "" {
			port = "8000"
		}
		
		logger.Printf("API Gateway listening on http://localhost:%s", port)
		logger.Println("Press Ctrl+C to stop")
		
		if err := app.Listen(":" + port); err != nil {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Attendre le signal d'arrêt
	<-sigChan
	logger.Println("Received shutdown signal, stopping API...")

	// Arrêt gracieux
	if err := app.Shutdown(); err != nil {
		logger.Printf("Error during shutdown: %v", err)
	}

	logger.Println("API Gateway stopped")
}
