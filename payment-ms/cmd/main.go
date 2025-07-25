package main

import (
	"log"
	"net/http"

	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/application/service"
	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/infrastructure/external"
	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/infrastructure/repository"
	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/interfaces/http/handler"
	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/interfaces/http/router"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/config"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/database"

	"github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
)

func main() {
	cfg := config.Load()

	db, err := database.NewMongoDB(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer db.Close()

	natsClient, err := messaging.NewNATSClient(cfg.NATS.URL)
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer natsClient.Close()

	paymentRepo := repository.NewMongoPaymentRepository(db.Database)
	paymentGetway := external.NewMockPaymentGateway()
	paymentService := service.NewPaymentService(paymentRepo, natsClient, paymentGetway)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	r := router.NewRouter(paymentHandler)
	port := "8084"
	log.Printf("Product service starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))

}
