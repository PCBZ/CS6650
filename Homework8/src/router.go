package main

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter configures and returns the main Gin router
func SetupRouter(orderAPI *OrderAPI, productAPI *ProductAPI, cartAPI CartAPI) *gin.Engine {
	// Set gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Health check endpoint for ALB
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "order-processing-api",
		})
	})

	// Setup API v1 routes
	setupV1Routes(router, orderAPI, productAPI, cartAPI)

	return router
}

// setupV1Routes configures all v1 API routes
func setupV1Routes(router *gin.Engine, orderAPI *OrderAPI, productAPI *ProductAPI, cartAPI CartAPI) {
	v1 := router.Group("/v1")
	{
		// Products domain routes
		setupProductsRoutes(v1, productAPI)
		// Orders domain routes
		setupOrdersRoutes(v1, orderAPI)
		// Shopping carts domain routes
		setupShoppingCartsRoutes(v1, cartAPI)
	}
}

// setupProductsRoutes configures product-related routes
func setupProductsRoutes(rg *gin.RouterGroup, productAPI *ProductAPI) {
	products := rg.Group("/products")
	{
		products.GET("/search", productAPI.SearchProducts)
		products.GET("/:productId", productAPI.GetProduct)
		products.POST("/:productId/details", productAPI.AddProductDetails)
	}
}

// setupOrdersRoutes configures order-related routes
func setupOrdersRoutes(rg *gin.RouterGroup, orderAPI *OrderAPI) {
	orders := rg.Group("/orders")
	{
		// Synchronous order processing endpoint (Phase 1)
		orders.POST("/sync", orderAPI.ProcessOrderSync)
		// Asynchronous order processing endpoint (Phase 3)
		orders.POST("/async", orderAPI.ProcessOrderAsync)
		// Statistics for monitoring during load tests
		orders.GET("/stats", orderAPI.GetOrderStats)
	}
}

// setupShoppingCartsRoutes configures shopping cart-related routes
func setupShoppingCartsRoutes(rg *gin.RouterGroup, cartAPI CartAPI) {
	carts := rg.Group("/shopping-carts")
	{
		// Create new shopping cart
		carts.POST("", cartAPI.CreateShoppingCart)
		// Get shopping cart with all items
		carts.GET("/:id", cartAPI.GetShoppingCart)
		// Add or update item in cart
		carts.POST("/:id/items", cartAPI.AddItemToCart)
	}
}
