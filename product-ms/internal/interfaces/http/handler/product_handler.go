package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/utils"
)

type ProductHandler struct {
	productService domain.ProductService
}

func NewProductHandler(productService domain.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

type CreateProductRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	Price       float64  `json:"price" validate:"required,gt=0"`
	Stock       int      `json:"stock" validate:"gte=0"`
	Category    string   `json:"category" validate:"required"`
	Images      []string `json:"images"`
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// @Summary Create a new product
// @Description Create a new product with details
// @Tags products
// @Accept json
// @Produce json
// @Param product body CreateProductRequest true "Product data"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Router /products [post]
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	product := &domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		Images:      req.Images,
	}

	if err := h.productService.CreateProduct(product); err != nil {
		if validationErrors := utils.GetValidationErrors(err); len(validationErrors) > 0 {
			h.sendValidationErrorResponse(w, validationErrors)
			return
		}
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.sendSuccessResponse(w, http.StatusCreated, product)
}

// @Summary Get product by ID
// @Description Get product details by ID
// @Tags products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} Response
// @Failure 404 {object} Response
// @Router /products/{id} [get]
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	product, err := h.productService.GetProduct(id)
	if err != nil {
		h.sendErrorResponse(w, http.StatusNotFound, "Product not found")
		return
	}

	h.sendSuccessResponse(w, http.StatusOK, product)
}

// @Summary List products
// @Description Get paginated list of products
// @Tags products
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param category query string false "Category filter"
// @Success 200 {object} Response
// @Router /products [get]
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	category := r.URL.Query().Get("category")

	products, err := h.productService.ListProducts(limit, offset, category)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccessResponse(w, http.StatusOK, products)
}

// @Summary Search products
// @Description Search products by name or description
// @Tags products
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} Response
// @Router /products/search [get]
func (h *ProductHandler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Search query is required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	products, err := h.productService.SearchProducts(query, limit, offset)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccessResponse(w, http.StatusOK, products)
}

// @Summary Update product
// @Description Update product details
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body CreateProductRequest true "Product data"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Router /products/{id} [put]
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	product := &domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		Images:      req.Images,
	}

	if err := h.productService.UpdateProduct(id, product); err != nil {
		if validationErrors := utils.GetValidationErrors(err); len(validationErrors) > 0 {
			h.sendValidationErrorResponse(w, validationErrors)
			return
		}
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	updatedProduct, _ := h.productService.GetProduct(id)
	h.sendSuccessResponse(w, http.StatusOK, updatedProduct)
}

func (h *ProductHandler) sendSuccessResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    data,
	})
}

func (h *ProductHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error:   message,
	})
}

func (h *ProductHandler) sendValidationErrorResponse(w http.ResponseWriter, errors interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error:   "Validation failed",
		Errors:  errors,
	})
}
