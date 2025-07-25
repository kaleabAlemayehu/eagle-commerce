package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type ProxyHandler struct {
	services map[string]*httputil.ReverseProxy
}

func NewProxyHandler() *ProxyHandler {
	services := make(map[string]*httputil.ReverseProxy)

	// Initialize service proxies
	userServiceURL, _ := url.Parse("http://localhost:8081")
	productServiceURL, _ := url.Parse("http://localhost:8082")
	orderServiceURL, _ := url.Parse("http://localhost:8083")
	paymentServiceURL, _ := url.Parse("http://localhost:8084")

	services["user"] = httputil.NewSingleHostReverseProxy(userServiceURL)
	services["product"] = httputil.NewSingleHostReverseProxy(productServiceURL)
	services["order"] = httputil.NewSingleHostReverseProxy(orderServiceURL)
	services["payment"] = httputil.NewSingleHostReverseProxy(paymentServiceURL)

	return &ProxyHandler{
		services: services,
	}
}

func (h *ProxyHandler) ProxyRequest(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy, exists := h.services[service]
		if !exists {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}

		// Add service identification header
		r.Header.Set("X-Service", service)

		proxy.ServeHTTP(w, r)
	}
}
