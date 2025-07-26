package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kaleabAlemayehu/eagle-commerce/order-ms/internal/application/dto"
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

// @Summary Create a new order
// @Description Create a new order with items and address
// @Tags orders
// @Accept json
// @Produce json
// @Param order body dto.CreateOrderRequest true "Order data"
// @Success 201 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOrderRequest
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
	orderRes := h.toOrderResponse(order)

	h.sendSuccessResponse(w, http.StatusCreated, orderRes)
}

// @Summary Get order list
// @Description Get orders as list
// @Tags orders
// @Produce json
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /orders [get]
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	order, err := h.orderService.ListOrders(limit, offset)
	if err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "Order not found")
		return
	}
	orders := h.toOrderListResponse(order)
	h.sendSuccessResponse(w, http.StatusOK, orders)
}

// @Summary Get order by ID
// @Description Get order details by ID
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	order, err := h.orderService.GetOrder(id)
	if err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "Order not found")
		return
	}
	orderRes := h.toOrderResponse(order)

	h.sendSuccessResponse(w, http.StatusOK, orderRes)
}

// @Summary Get user orders
// @Description Get orders for a specific user
// @Tags orders
// @Produce json
// @Param user_id path string true "User ID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} dto.Response
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
	ordersRes := h.toOrderListResponse(orders)

	h.sendSuccessResponse(w, http.StatusOK, ordersRes)
}

// @Summary Update order status
// @Description Update the status of an order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param status body dto.UpdateOrderStatusRequest true "Status update"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /orders/{id}/status [put]
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.UpdateOrderStatusRequest
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
	updatedOrderRes := h.toOrderResponse(updatedOrder)
	h.sendSuccessResponse(w, http.StatusOK, updatedOrderRes)
}

// @Summary Cancel order
// @Description Cancel an order
// @Tags orders
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /orders/{id}/cancel [put]
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.orderService.CancelOrder(id); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	updatedOrder, _ := h.orderService.GetOrder(id)
	updatedOrderRes := h.toOrderResponse(updatedOrder)
	h.sendSuccessResponse(w, http.StatusOK, updatedOrderRes)
}

func (h *OrderHandler) sendSuccessResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dto.Response{
		Success: true,
		Data:    data,
	})
}

func (h *OrderHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dto.Response{
		Success: false,
		Error:   message,
	})
}

func (h *OrderHandler) sendValidationErrorResponse(w http.ResponseWriter, errors interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(dto.Response{
		Success: false,
		Error:   "Validation failed",
		Errors:  errors,
	})
}

func (h *OrderHandler) toOrderListResponse(orders []*domain.Order) []*dto.OrderResponse {
	res := make([]*dto.OrderResponse, len(orders))

	for i, o := range orders {
		res[i] = h.toOrderResponse(o)
	}
	return res

}

func (h *OrderHandler) toOrderResponse(order *domain.Order) *dto.OrderResponse {
	return &dto.OrderResponse{
		ID:        order.ID.String(),
		UserID:    order.UserID,
		Items:     h.toOrderItemListResponse(order.Items),
		Total:     order.Total,
		Status:    string(order.Status),
		Address:   h.toAddressResponse(order.Address),
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}
}

func (h *OrderHandler) toOrderItemListResponse(orderItems []domain.OrderItem) []dto.OrderItemResponse {
	res := make([]dto.OrderItemResponse, len(orderItems))
	for i, o := range orderItems {
		res[i] = h.toOrderItem(o)
	}
	return res
}

func (h *OrderHandler) toAddressResponse(adr domain.Address) dto.AddressResponse {
	return dto.AddressResponse{
		Street:  adr.Street,
		State:   adr.State,
		City:    adr.City,
		ZipCode: adr.ZipCode,
		Country: adr.Country,
	}
}

func (h *OrderHandler) toOrderItem(orderItem domain.OrderItem) dto.OrderItemResponse {
	return dto.OrderItemResponse{
		ProductID: orderItem.ProductID,
		Name:      orderItem.Name,
		Price:     orderItem.Price,
		Quantity:  orderItem.Quantity,
		Subtotal:  orderItem.Price * float64(orderItem.Quantity),
	}
}
