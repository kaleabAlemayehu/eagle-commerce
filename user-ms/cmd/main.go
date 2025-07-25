package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/application/service"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/infrastructure/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/infrastructure/repository"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/interfaces/http/handler"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/interfaces/http/router"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/config"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/database"
	sharedMessaging "github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
)

// @title User Service API
// @version 1.0
// @description This is a user service API for eCommerce application
// @host localhost:8081
// @BasePath /api/v1
func main() {
	cfg := config.Load()

	// Connect to MongoDB
	db, err := database.NewMongoDB(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer db.Close()

	// Connect to NATS
	natsClient, err := sharedMessaging.NewNATSClient(cfg.NATS.URL)
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer natsClient.Close()
	nats := messaging.NewUserEventPublisher(natsClient)

	// listening for the incoming events from other services
	messaging.NewUserEventHandler(natsClient).StartListening()

	// Initialize dependencies
	userRepo := repository.NewMongoUserRepository(db.Database)
	userService := service.NewUserService(userRepo, nats)
	userHandler := handler.NewUserHandler(userService)

	// Setup router
	r := router.NewRouter(userHandler)

	// Start server
	port := "8081"
	fmt.Printf("User service starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
