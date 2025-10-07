package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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

// SearchProducts handles GET /products/search?q={query}
func (api *ProductAPI) SearchProducts(c *gin.Context) {
	startTime := time.Now()

	// Get query parameter
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, NewInvalidInputError("Query parameter 'q' is required"))
		return
	}

	// Convert query to lowercase for case-insensitive search
	queryLower := strings.ToLower(query)

	// Get all product IDs
	allIDs := api.store.GetAllProductIDs()

	var matches []Product
	totalFound := 0
	productsChecked := 0
	maxToCheck := 100
	maxResults := 20

	// Check exactly 100 products then stop
	for _, id := range allIDs {
		if productsChecked >= maxToCheck {
			break
		}

		productsChecked++

		product, exists := api.store.GetProduct(id)
		if !exists {
			continue
		}

		// Search in name (using Brand field as name) and category (case-insensitive)
		productName := strings.ToLower(product.Brand)
		productCategory := strings.ToLower(product.Category)

		if strings.Contains(productName, queryLower) || strings.Contains(productCategory, queryLower) {
			totalFound++
			// Only add to results if we haven't reached the max results limit
			if len(matches) < maxResults {
				matches = append(matches, *product)
			}
		}
	}

	// Calculate search time
	searchTime := time.Since(startTime).Round(time.Millisecond)

	response := SearchResponse{
		Products:   matches,
		TotalFound: totalFound,
		SearchTime: searchTime.String(),
	}

	c.JSON(http.StatusOK, response)
}
