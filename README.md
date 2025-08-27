# Eagle-Commerce Microservices Platform

A scalable e-commerce platform built with Go microservices architecture, featuring clean architecture principles, event-driven communication, and containerized deployment.

## 🏗️ Architecture Overview

This platform implements a microservices architecture with the following components:

- **API Gateway**: Centralized entry point with authentication, rate limiting, and request routing
- **User Service**: User management, authentication, and profile operations
- **Product Service**: Product catalog, inventory, and search functionality
- **Order Service**: Order processing, order history, and order status management
- **Payment Service**: Payment processing with external gateway integration
- **Shared Libraries**: Common utilities, models, and middleware

## 🛠️ Technology Stack

- **Language**: Go 1.24+
- **Database**: MongoDB
- **Message Broker**: NATS
- **API Documentation**: Swagger/OpenAPI
- **Containerization**: Docker & Docker Compose
- **Architecture Pattern**: Hexagonal Architecture with DDD principles

## 📁 Project Structure

```
eagle-ecommerce/
├── shared/              # Shared libraries and utilities
├── user-ms/        # User management service
├── product-ms/     # Product catalog service
├── order-ms/       # Order processing service
└── payment-ms/     # Payment processing service
├── api-gateway/        # API Gateway service
├── docker-compose.yml  # Development environment setup
├── Makefile           # Build and deployment commands
└── go.work           # Go workspace configuration
```

Each microservice follows Clean Architecture principles:
- **Domain**: Business entities and rules
- **Application**: Use cases and DTOs
- **Infrastructure**: External concerns (DB, messaging, external APIs)
- **Interfaces**: HTTP handlers and routing

## 🚀 Quick Start

### Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose
- Make (optional, for using Makefile commands)
- Nix (optional, for using Nix-shell )

### Development Setup

1. **Clone the repository**
   ```bash
   git clone git@github.com:kaleabAlemayehu/eagle-commerce.git
   cd eagle-ecommerce
   ```

2. **Start infrastructure services**
   ```bash
   make docker-up
   ```

3. **Install dependencies**
   ```bash
   go work sync
   ```

4. **Run services locally**
   ```bash
   # Start all services
   make run
   ```

### Docker Deployment

Run the entire platform with Docker Compose:

```bash
docker-compose up --build
```

## 📚 API Documentation

Each service provides Swagger documentation:

- **User Service**: http://localhost:8081/swagger/
- **Product Service**: http://localhost:8082/swagger/
- **Order Service**: http://localhost:8083/swagger/
- **Payment Service**: http://localhost:8084/swagger/

## 🔧 Configuration

Services are configured through environment variables. Key configurations:

### Database
```env
MONGO_URI=mongodb://localhost:27017
DATABASE_NAME=ecommerce
```

### Message Broker
```env
NATS_URL=nats://localhost:4222
```

### Service Discovery
```env
USER_SERVICE_URL=http://localhost:8081
PRODUCT_SERVICE_URL=http://localhost:8082
ORDER_SERVICE_URL=http://localhost:8083
PAYMENT_SERVICE_URL=http://localhost:8084
```

## 🔄 Event-Driven Architecture

The platform uses NATS for asynchronous communication between services:

### Event Types
- `user.created` - Published when a user registers
- `user.updated` - Published when user profile is updated
- `user.deleted` - Published when user profile is deleted
- `product.created` - Published when a product is added
- `product.stock.updated`- Published when a product stock updated
- `stock.check.response` - Published as response to a product stock check
- `product.updated` - Published when product details change
- `order.created` - Published when an order is placed
- `order.cancelled` - Published when an order is cancelled
- `order.status.changed` - Published when order status updates
- `refund.requested`- Published when payment refuned is requested
- `stock.check` - Published when a product stock get checked when ordering
- `stock.reserve` - Published when a product stock get reserved
- `payment.processed` - Published when payment is completed
- `payment.failed` - Published when payment fails
- `payment.refunded` - Published when payment is refunded


## 📦 Available Make Commands

```bash
make build              # Build all services
make run                # Run all services locally
make docker-build       # Build Docker images
make docker-up          # Start with Docker Compose
make docker-down        # Stop Docker Compose
make clean              # Clean build artifacts
make lint               # Run linter
make swagger            # Generate Swagger docs
```

## 🐳 Docker Services

The docker-compose.yml includes:

- **MongoDB**: Database (port 27017)
- **NATS**: Message broker (port 4222)
- **API Gateway**: Main entry point (port 8080)
- **User Service**: User management (port 8081)
- **Product Service**: Product catalog (port 8082)
- **Order Service**: Order processing (port 8083)
- **Payment Service**: Payment processing (port 8084)

## 🔐 Authentication & Authorization

The platform uses JWT-based authentication:

1. Users authenticate through the User Service
2. JWT tokens are validated by the API Gateway
3. Service-to-service communication uses internal authentication

## 📊 Monitoring & Observability

- **Health Checks**: Each service exposes `/health` endpoint

## 🔧 Development Workflow

1. **Feature Development**
   - Create feature branch
   - Implement changes following clean architecture
   - Add/update tests
   - Update documentation

2. **Code Quality**
   - Run linter: `make lint`

3. **API Changes**
   - Update swagger annotations
   - Regenerate docs: `make swagger`
   - Update API documentation

## 🚀 Deployment

### Production Considerations

1. **Environment Variables**: Set production configurations
2. **Database**: Use MongoDB replica sets for production
3. **Message Broker**: Configure NATS clustering

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style

- Follow Go conventions and best practices
- Use `gofmt` for code formatting
- Add appropriate comments and documentation
- Write tests for new functionality

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Happy Coding! 🎉**
