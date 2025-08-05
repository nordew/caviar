package rest

import (
	"caviar/internal/dto"
	"caviar/internal/types"
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initProductRoutes(r *gin.RouterGroup) {
	products := r.Group("/products")
	
	products.POST("/search", h.listProducts)

	productsProtected := products.Group("/", h.AuthMiddleware())
	productsProtected.POST("/", h.createProduct)
	productsProtected.PUT("/:id", h.updateProduct)
	productsProtected.DELETE("/:id", h.deleteProduct)
}

// ListProducts godoc
// @Summary Search products
// @Description Search and filter products with pagination
// @Tags products
// @Accept json
// @Produce json
// @Param filter body types.ProductFilter false "Product filter parameters"
// @Success 200 {array} models.Product "List of products"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/products/search [post]
func (h *Handler) listProducts(c *gin.Context) {
	var filter types.ProductFilter
	
	if err := c.ShouldBindJSON(&filter); err != nil {
		if err == io.EOF {
			filter = *types.DefaultProductFilter()
		} else {
			h.handleError(c, err)
			return
		}
	}

	products, err := h.productService.List(context.Background(), false, &filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, products)
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with variants and details
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product body dto.ProductCreateDTO true "Product creation data"
// @Success 200 {object} dto.ProductCreateDTO "Created product"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/products [post]
func (h *Handler) createProduct(c *gin.Context) {
	var product dto.ProductCreateDTO
	if err := c.ShouldBindJSON(&product); err != nil {
		h.handleError(c, err)
		return
	}

	err := h.productService.Create(context.Background(), &product)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, product)
}

// UpdateProduct godoc
// @Summary Update an existing product
// @Description Update product details, variants, and other information
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param product body dto.ProductUpdateDTO true "Product update data"
// @Success 200 {object} dto.ProductUpdateDTO "Updated product"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/products/{id} [put]
func (h *Handler) updateProduct(c *gin.Context) {
	var product dto.ProductUpdateDTO
	if err := c.ShouldBindJSON(&product); err != nil {
		h.handleError(c, err)
		return
	}

	err := h.productService.Update(context.Background(), &product)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product and all its variants
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} map[string]string "Success message"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Product not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/products/{id} [delete]
func (h *Handler) deleteProduct(c *gin.Context) {
	id := c.Param("id")

	err := h.productService.Delete(context.Background(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}