package main

import (
	"log"
	"net/http"

	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/application/service"
	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/infrastructure/external"
	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/infrastructure/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/infrastructure/repository"
	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/interfaces/http/handler"
	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/interfaces/http/router"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/config"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/database"

	sharedMessaging "github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
	sharedMiddleware "github.com/kaleabAlemayehu/eagle-commerce/shared/middleware"
)

func main() {
	cfg := config.Load()

	db, err := database.NewMongoDB(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer db.Close()

	natsClient, err := sharedMessaging.NewNATSClient(cfg.NATS.URL)
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer natsClient.Close()
	paymentPublisher := messaging.NewPaymentEventPublisher(natsClient)

	paymentRepo := repository.NewMongoPaymentRepository(db.Database)
	paymentGetway := external.NewMockPaymentGateway()
	paymentService := service.NewPaymentService(paymentRepo, paymentPublisher, paymentGetway)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	if err := messaging.NewPaymentEventHandler(paymentService, natsClient).StartListening(); err != nil {
		log.Fatal("Failed to listen events from NATS:", err)
	}

	auth := sharedMiddleware.NewAuth(cfg.JWTSecret)
	r := router.NewRouter(paymentHandler, auth)
	port := "8084"
	log.Printf("Product service starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))

}
