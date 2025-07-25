package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/application/dto"
	"github.com/kaleabAlemayehu/eagle-commerce/payment-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/utils"
)

type PaymentHandler struct {
	paymentService domain.PaymentService
}

func NewPaymentHandler(paymentService domain.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// TODO: for webhook
func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {

}

// @Summary Process a payment
// @Description Process payment for an order
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body dto.ProcessPaymentRequest true "Payment data"
// @Success 201 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /payments [post]
func (h *PaymentHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	var req dto.ProcessPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	payment := &domain.Payment{
		OrderID:  req.OrderID,
		UserID:   req.UserID,
		Amount:   req.Amount,
		Currency: req.Currency,
		Method:   domain.PaymentMethod(req.Method),
	}

	if err := h.paymentService.ProcessPayment(payment); err != nil {
		if validationErrors := utils.GetValidationErrors(err); len(validationErrors) > 0 {
			h.sendValidationErrorResponse(w, validationErrors)
			return
		}
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	response := dto.PaymentResponse{
		ID:            payment.ID.Hex(),
		OrderID:       payment.OrderID,
		UserID:        payment.UserID,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Status:        string(payment.Status),
		Method:        string(payment.Method),
		TransactionID: payment.TransactionID,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}

	h.sendSuccessResponse(w, http.StatusCreated, response)
}

// @Summary Get payment by ID
// @Description Get payment details by ID
// @Tags payments
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /payments/{id} [get]
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	payment, err := h.paymentService.GetPayment(id)
	if err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "Payment not found")
		return
	}

	response := dto.PaymentResponse{
		ID:            payment.ID.Hex(),
		OrderID:       payment.OrderID,
		UserID:        payment.UserID,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Status:        string(payment.Status),
		Method:        string(payment.Method),
		TransactionID: payment.TransactionID,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}

	h.sendSuccessResponse(w, http.StatusOK, response)
}

// @Summary Get payment by order ID
// @Description Get payment details by order ID
// @Tags payments
// @Produce json
// @Param order_id path string true "Order ID"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /payments/order/{order_id} [get]
func (h *PaymentHandler) GetPaymentByOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "order_id")

	payment, err := h.paymentService.GetPaymentByOrder(orderID)
	if err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "Payment not found")
		return
	}

	response := dto.PaymentResponse{
		ID:            payment.ID.Hex(),
		OrderID:       payment.OrderID,
		UserID:        payment.UserID,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Status:        string(payment.Status),
		Method:        string(payment.Method),
		TransactionID: payment.TransactionID,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}

	h.sendSuccessResponse(w, http.StatusOK, response)
}

// @Summary Refund payment
// @Description Refund a completed payment
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param refund body dto.RefundPaymentRequest false "Refund details"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /payments/{id}/refund [post]
func (h *PaymentHandler) RefundPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.RefundPaymentRequest
	json.NewDecoder(r.Body).Decode(&req) // Optional body

	if err := h.paymentService.RefundPayment(id); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	updatedPayment, _ := h.paymentService.GetPayment(id)
	response := dto.PaymentResponse{
		ID:            updatedPayment.ID.Hex(),
		OrderID:       updatedPayment.OrderID,
		UserID:        updatedPayment.UserID,
		Amount:        updatedPayment.Amount,
		Currency:      updatedPayment.Currency,
		Status:        string(updatedPayment.Status),
		Method:        string(updatedPayment.Method),
		TransactionID: updatedPayment.TransactionID,
		CreatedAt:     updatedPayment.CreatedAt,
		UpdatedAt:     updatedPayment.UpdatedAt,
	}

	h.sendSuccessResponse(w, http.StatusOK, response)
}

// @Summary List payments
// @Description Get paginated list of payments
// @Tags payments
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} dto.Response
// @Router /payments [get]
func (h *PaymentHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	payments, err := h.paymentService.ListPayments(limit, offset)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	var responses []dto.PaymentResponse
	for _, payment := range payments {
		responses = append(responses, dto.PaymentResponse{
			ID:            payment.ID.Hex(),
			OrderID:       payment.OrderID,
			UserID:        payment.UserID,
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			Status:        string(payment.Status),
			Method:        string(payment.Method),
			TransactionID: payment.TransactionID,
			CreatedAt:     payment.CreatedAt,
			UpdatedAt:     payment.UpdatedAt,
		})
	}

	h.sendSuccessResponse(w, http.StatusOK, responses)
}

func (h *PaymentHandler) sendSuccessResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dto.Response{
		Success: true,
		Data:    data,
	})
}

func (h *PaymentHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(dto.Response{
		Success: false,
		Error:   message,
	})
}

func (h *PaymentHandler) sendValidationErrorResponse(w http.ResponseWriter, errors interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(dto.Response{
		Success: false,
		Error:   "Validation failed",
		Errors:  errors,
	})
}
