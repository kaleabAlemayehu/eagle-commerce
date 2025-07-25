package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/utils"
)

type OrderHandler struct {
	orderService domain.OrderService
}

func NewOrderHandler(orderService domain.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

type CreateOrderRequest struct {
	UserID  string                   `json:"user_id" validate:"required"`
	Items   []CreateOrderItemRequest `json:"items" validate:"required,dive"`
	Address AddressRequest           `json:"address" validate:"required"`
}

type CreateOrderItemRequest struct {
	ProductID string  `json:"product_id" validate:"required"`
	Name      string  `json:"name" validate:"required"`
	Price     float64 `json:"price" validate:"gt=0"`
	Quantity  int     `json:"quantity" validate:"gt=0"`
}

type AddressRequest struct {
	Street  string `json:"street" validate:"required"`
	City    string `json:"city" validate:"required"`
	State   string `json:"state" validate:"required"`
	ZipCode string `json:"zip_code" validate:"required"`
	Country string `json:"country" validate:"required"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// @Summary Create a new order
// @Description Create a new order with items and address
// @Tags orders
// @Accept json
// @Produce json
// @Param order body CreateOrderRequest true "Order data"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Convert request to domain model
	items := make([]domain.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = domain.OrderItem{
			ProductID: item.ProductID,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  item.Quantity,
		}
	}

	order := &domain.Order{
		UserID: req.UserID,
		Items:  items,
		Address: domain.Address{
			Street:  req.Address.Street,
			City:    req.Address.City,
			State:   req.Address.State,
			ZipCode: req.Address.ZipCode,
			Country: req.Address.Country,
		},
	}

	if err := h.orderService.CreateOrder(order); err != nil {
		if validationErrors := utils.GetValidationErrors(err); len(validationErrors) > 0 {
			h.sendValidationErrorResponse(w, validationErrors)
			return
		}
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.sendSuccessResponse(w, http.StatusCreated, order)
}

// @Summary Get order by ID
// @Description Get order details by ID
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} Response
// @Failure 404 {object} Response
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	order, err := h.orderService.GetOrder(id)
	if err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "Order not found")
		return
	}

	h.sendSuccessResponse(w, http.StatusOK, order)
}

// @Summary Get user orders
// @Description Get orders for a specific user
// @Tags orders
// @Produce json
// @Param user_id path string true "User ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} Response
// @Router /orders/user/{user_id} [get]
func (h *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	orders, err := h.orderService.GetOrdersByUser(userID, limit, offset)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccessResponse(w, http.StatusOK, orders)
}

// @Summary Update order status
// @Description Update the status of an order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param status body UpdateOrderStatusRequest true "Status update"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /orders/{id}/status [put]
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	status := domain.OrderStatus(req.Status)
	if err := h.orderService.UpdateOrderStatus(id, status); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	updatedOrder, _ := h.orderService.GetOrder(id)
	h.sendSuccessResponse(w, http.StatusOK, updatedOrder)
}

// @Summary Cancel order
// @Description Cancel an order
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /orders/{id}/cancel [put]
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.orderService.CancelOrder(id); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	updatedOrder, _ := h.orderService.GetOrder(id)
	h.sendSuccessResponse(w, http.StatusOK, updatedOrder)
}

func (h *OrderHandler) sendSuccessResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    data,
	})
}

func (h *OrderHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error:   message,
	})
}

func (h *OrderHandler) sendValidationErrorResponse(w http.ResponseWriter, errors interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error:   "Validation failed",
		Errors:  errors,
	})
}
