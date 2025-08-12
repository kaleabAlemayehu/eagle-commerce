package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kaleabAlemayehu/eagle-commerce/product-ms/internal/application/dto"
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

// @Summary Create a new product
// @Description Create a new product with details
// @Tags products
// @Accept json
// @Produce json
// @Param product body dto.CreateProductRequest true "Product data"
// @Success 201 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /products [post]
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
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

	if err := h.productService.CreateProduct(r.Context(), product); err != nil {
		if validationErrors := utils.GetValidationErrors(err); len(validationErrors) > 0 {
			utils.SendValidationErrorResponse(w, validationErrors)
			return
		}
		utils.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	productRes := h.toProductResponse(product)
	utils.SendSuccessResponse(w, http.StatusCreated, productRes)
}

// @Summary Get product by ID
// @Description Get product details by ID
// @Tags products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /products/{id} [get]
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	product, err := h.productService.GetProduct(r.Context(), id)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusNotFound, "Product not found")
		return
	}

	productRes := h.toProductResponse(product)
	utils.SendSuccessResponse(w, http.StatusOK, productRes)
}

// @Summary List products
// @Description Get paginated list of products
// @Tags products
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param category query string false "Category filter"
// @Success 200 {object} dto.Response
// @Router /products [get]
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	category := r.URL.Query().Get("category")

	products, err := h.productService.ListProducts(r.Context(), limit, offset, category)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	productsList := h.toProductResponseList(products)
	productsRes := dto.ProductListResponse{
		Products: productsList,
		Total:    len(productsList),
	}

	utils.SendSuccessResponse(w, http.StatusOK, productsRes)
}

// @Summary Search products
// @Description Search products by name or description
// @Tags products
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} dto.Response
// @Router /products/search [get]
func (h *ProductHandler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		utils.SendErrorResponse(w, http.StatusBadRequest, "Search query is required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	products, err := h.productService.SearchProducts(r.Context(), query, limit, offset)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	productList := h.toProductResponseList(products)
	productsRes := dto.ProductSearchResponse{
		Products: productList,
		Query:    query,
		Total:    len(productList),
	}

	utils.SendSuccessResponse(w, http.StatusOK, productsRes)
}

// @Summary Update product
// @Description Update product details
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body dto.CreateProductRequest true "Product data"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /products/{id} [put]
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
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

	updatedProduct, err := h.productService.UpdateProduct(r.Context(), id, product)
	if err != nil {
		if validationErrors := utils.GetValidationErrors(err); len(validationErrors) > 0 {
			utils.SendValidationErrorResponse(w, validationErrors)
			return
		}
		utils.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	productRes := h.toProductResponse(updatedProduct)
	utils.SendSuccessResponse(w, http.StatusOK, productRes)
}

// @Summary Delete product by ID
// @Description Delete product details by ID
// @Tags products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.productService.DeleteProduct(r.Context(), id); err != nil {
		utils.SendErrorResponse(w, http.StatusNotFound, "Product not found")
		return
	}

	utils.SendSuccessResponse(w, http.StatusOK, "Product Deleted Successfully")
}

// @Summary Check stoke of a new product
// @Description Check the stock of product
// @Tags products
// @Accept json
// @Produce json
// @Param product body dto.StockCheckRequest true "CheckStock data"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /products/check-stock [post]
func (h *ProductHandler) CheckStock(w http.ResponseWriter, r *http.Request) {
	var req dto.StockCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	ok, n, err := h.productService.CheckStock(r.Context(), req.ProductID, req.Quantity)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusNotFound, "Product not found")
		return
	}
	res := dto.StockCheckResponse{
		ProductID: req.ProductID,
		Available: ok,
		Stock:     n,
		Requested: req.Quantity,
	}
	utils.SendSuccessResponse(w, http.StatusOK, res)
}

// @Summary Check stoke of a new product
// @Description Check the stock of product
// @Tags products
// @Accept json
// @Produce json
// @Param product body dto.StockUpdateRequest true "CheckStock data"
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /products/reserve-stock [post]
func (h *ProductHandler) ReserveStock(w http.ResponseWriter, r *http.Request) {
	var req dto.StockUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := utils.ValidateStruct(req); err != nil {
		utils.SendValidationErrorResponse(w, err)
		return
	}
	if err := h.productService.ReserveStock(r.Context(), req.ProductID, req.Quantity); err != nil {
		utils.SendErrorResponse(w, http.StatusNotFound, "Product not found")
		return
	}
	utils.SendSuccessResponse(w, http.StatusOK, "Stock reserved successfully")
}

func (h *ProductHandler) toProductResponse(p *domain.Product) dto.ProductResponse {
	return dto.ProductResponse{
		ID:          p.ID.String(), // Convert ObjectID to string
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		Category:    p.Category,
		Images:      p.Images,
		Active:      p.Active,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func (h *ProductHandler) toProductResponseList(products []*domain.Product) []dto.ProductResponse {
	res := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		res[i] = h.toProductResponse(p)
	}
	return res
}
