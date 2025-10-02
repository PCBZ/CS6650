package main

// Global product store
var productStore = NewProductStore()

func main() {
	router := SetupRouter()

	// Start server on port 8080
	if err := router.Run(":8080"); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
