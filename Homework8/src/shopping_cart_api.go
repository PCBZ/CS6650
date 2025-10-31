package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ShoppingCartAPI handles shopping cart operations using GORM
type ShoppingCartAPI struct {
	db *gorm.DB
}

// NewShoppingCartAPI creates a new shopping cart API handler
func NewShoppingCartAPI(db *gorm.DB) *ShoppingCartAPI {
	return &ShoppingCartAPI{db: db}
}

// CreateCartRequest represents the request to create a new shopping cart
type CreateCartRequest struct {
	CustomerID int `json:"customer_id" binding:"required,gt=0"`
}

// CreateCartResponse represents the response after creating a cart
type CreateCartResponse struct {
	ShoppingCartID int    `json:"shopping_cart_id"`
	CustomerID     int    `json:"customer_id"`
	Status         string `json:"status"`
}

// AddItemRequest represents the request to add an item to a cart
type AddItemRequest struct {
	ProductID int `json:"product_id" binding:"required,gt=0"`
	Quantity  int `json:"quantity" binding:"required,gt=0"`
}

// AddItemResponse represents the response after adding an item
type AddItemResponse struct {
	CartItemID     int `json:"cart_item_id"`
	ShoppingCartID int `json:"shopping_cart_id"`
	ProductID      int `json:"product_id"`
	Quantity       int `json:"quantity"`
}

// CreateShoppingCart creates a new shopping cart for a customer
// POST /shopping-carts
func (api *ShoppingCartAPI) CreateShoppingCart(c *gin.Context) {
	if api.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Database not available",
		})
		return
	}

	var req CreateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create new shopping cart using GORM transaction
	cart := ShoppingCart{
		CustomerID: req.CustomerID,
		Status:     "active",
	}

	if err := api.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&cart).Error
	}); err != nil {
		log.Printf("Failed to create shopping cart: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create cart",
		})
		return
	}

	response := CreateCartResponse{
		ShoppingCartID: cart.ShoppingCartID,
		CustomerID:     cart.CustomerID,
		Status:         cart.Status,
	}

	c.JSON(http.StatusCreated, response)
}

// GetShoppingCart retrieves a shopping cart with all items
// GET /shopping-carts/:id
func (api *ShoppingCartAPI) GetShoppingCart(c *gin.Context) {
	if api.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Database not available",
		})
		return
	}

	cartIDStr := c.Param("id")
	cartID, err := strconv.Atoi(cartIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid cart ID",
		})
		return
	}

	var cart ShoppingCart

	// Use GORM's Preload to fetch cart with items (JOIN)
	result := api.db.Preload("Items").First(&cart, cartID)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Cart not found",
			})
			return
		}
		log.Printf("Failed to query shopping cart: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve cart",
		})
		return
	}

	c.JSON(http.StatusOK, cart)
}

// AddItemToCart adds or updates an item in the shopping cart
// POST /shopping-carts/:id/items
func (api *ShoppingCartAPI) AddItemToCart(c *gin.Context) {
	if api.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Database not available",
		})
		return
	}

	cartIDStr := c.Param("id")
	cartID, err := strconv.Atoi(cartIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid cart ID",
		})
		return
	}

	var req AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var cartItemID int
	var finalQuantity int

	// Use GORM transaction for multi-table operations
	err = api.db.Transaction(func(tx *gorm.DB) error {
		// Verify cart exists
		var cart ShoppingCart
		if err := tx.First(&cart, cartID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("CART_NOT_FOUND")
			}
			return err
		}

		// Verify product exists
		var product Product
		if err := tx.First(&product, req.ProductID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("PRODUCT_NOT_FOUND")
			}
			return err
		}

		// Check if item already exists in cart
		var existingItem ShoppingCartItem
		result := tx.Where("shopping_cart_id = ? AND product_id = ?", cartID, req.ProductID).
			First(&existingItem)

		if result.Error == nil {
			// Item exists, update quantity
			existingItem.Quantity += req.Quantity
			if err := tx.Save(&existingItem).Error; err != nil {
				return err
			}
			cartItemID = existingItem.CartItemID
			finalQuantity = existingItem.Quantity
		} else if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Item doesn't exist, create new
			newItem := ShoppingCartItem{
				ShoppingCartID: cartID,
				ProductID:      req.ProductID,
				Quantity:       req.Quantity,
			}
			if err := tx.Create(&newItem).Error; err != nil {
				return err
			}
			cartItemID = newItem.CartItemID
			finalQuantity = newItem.Quantity
		} else {
			return result.Error
		}

		// Update cart's updated_at timestamp
		if err := tx.Model(&cart).Update("updated_at", gorm.Expr("NOW()")).Error; err != nil {
			log.Printf("Failed to update cart timestamp: %v", err)
			// Continue anyway, as the main operation succeeded
		}

		return nil
	})

	if err != nil {
		if err.Error() == "CART_NOT_FOUND" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Cart not found",
			})
			return
		}
		if err.Error() == "PRODUCT_NOT_FOUND" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Product not found",
			})
			return
		}
		log.Printf("Failed to add item to cart: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add item",
		})
		return
	}

	response := AddItemResponse{
		CartItemID:     cartItemID,
		ShoppingCartID: cartID,
		ProductID:      req.ProductID,
		Quantity:       finalQuantity,
	}

	c.JSON(http.StatusOK, response)
}
