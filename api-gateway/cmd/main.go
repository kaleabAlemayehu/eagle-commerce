package main

import (
	"fmt"
	"log"
	"net/http"

	handler "github.com/kaleabAlemayehu/eagle-commerce/api-gateway/internal/handler"
	router "github.com/kaleabAlemayehu/eagle-commerce/api-gateway/internal/router"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/config"
	sharedMiddleware "github.com/kaleabAlemayehu/eagle-commerce/shared/middleware"
)

func main() {
	cfg := config.Load()
	proxyHandler := handler.NewProxyHandler()
	secret := config.Load().JWTSecret
	auth := sharedMiddleware.NewAuth(secret)
	r := router.NewRouter(proxyHandler, auth, cfg.AllowedOrigins)
	port := "8080"
	fmt.Printf("API Gateway starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
