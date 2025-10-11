package main

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter configures and returns the main Gin router
func SetupRouter() *gin.Engine {
	// Set gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Health check endpoint for ALB
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "product-api",
		})
	})

	// Setup API v1 routes
	setupV1Routes(router)

	return router
}

// setupV1Routes configures all v1 API routes
func setupV1Routes(router *gin.Engine) {
	v1 := router.Group("/v1")
	{
		// Products domain routes
		setupProductsRoutes(v1)
	}
}

// setupProductsRoutes configures product-related routes
func setupProductsRoutes(rg *gin.RouterGroup) {
	products := rg.Group("/products")
	productAPI := NewProductAPI()
	{
		products.GET("/search", productAPI.SearchProducts)
		products.GET("/:productId", productAPI.GetProduct)
		products.POST("/:productId/details", productAPI.AddProductDetails)
	}
}
