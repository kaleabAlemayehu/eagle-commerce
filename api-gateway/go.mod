module github.com/kaleabAlemayehu/eagle-commerce/api-gateway

go 1.24.3

require (
	github.com/go-chi/chi/v5 v5.2.2
	github.com/kaleabAlemayehu/eagle-commerce/shared v0.0.0
)

require github.com/golang-jwt/jwt/v5 v5.2.3 // indirect

replace github.com/kaleabAlemayehu/eagle-commerce/shared => ../shared
