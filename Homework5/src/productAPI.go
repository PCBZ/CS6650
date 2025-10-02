package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ProductAPI contains all product-related business logic
type ProductAPI struct {
	store *ProductStore
}

// NewProductAPI returns a new instance of ProductAPI
func NewProductAPI() *ProductAPI {
	return &ProductAPI{
		store: productStore,
	}
}

// GetProduct handles GET /products/{productId}
func (api *ProductAPI) GetProduct(c *gin.Context) {
	// Parse productId from path parameter
	productIDStr := c.Param("productId")
	productID64, err := strconv.ParseInt(productIDStr, 10, 32)
	if err != nil || productID64 < 1 {
		c.JSON(http.StatusBadRequest, NewInvalidInputError("Product ID must be a positive integer"))
		return
	}

	productID := int32(productID64)

	// Get product from store
	product, exists := api.store.GetProduct(productID)
	if !exists {
		c.JSON(http.StatusNotFound, NewProductNotFoundError(productID))
		return
	}

	c.JSON(http.StatusOK, product)
}

// AddProductDetails handles POST /products/{productId}/details
func (api *ProductAPI) AddProductDetails(c *gin.Context) {
	// Parse productId from path parameter
	productIDStr := c.Param("productId")
	productID64, err := strconv.ParseInt(productIDStr, 10, 32)
	if err != nil || productID64 < 1 {
		c.JSON(http.StatusBadRequest, NewInvalidInputError("Product ID must be a positive integer"))
		return
	}

	productID := int32(productID64)

	// Parse request body
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, NewInvalidInputError(err.Error()))
		return
	}

	// Check product ID matches path parameter
	if product.ProductID != productID {
		c.JSON(http.StatusBadRequest, NewInvalidInputError("Product ID in request body must match path parameter"))
		return
	}

	// Validate the product
	if err := product.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, NewInvalidInputError(err.Error()))
		return
	}

	// Store the product
	api.store.SetProduct(productID, &product)

	// Return 204 No Content
	c.Status(http.StatusNoContent)
}
