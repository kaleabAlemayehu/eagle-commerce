package main

import (
	"fmt"
	"log"
	"net/http"

	handler "github.com/kaleabAlemayehu/eagle-commerce/api-gateway/internal/handler"
	router "github.com/kaleabAlemayehu/eagle-commerce/api-gateway/internal/router"
)

func main() {
	proxyHandler := handler.NewProxyHandler()
	r := router.NewRouter(proxyHandler)

	port := "8080"
	fmt.Printf("API Gateway starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
