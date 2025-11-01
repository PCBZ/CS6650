package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
)

// generateSnowflakeID creates a simple Snowflake-like ID
func generateSnowflakeID() int64 {
	timestamp := time.Now().UnixMilli() << 22
	machineID := int64(1) << 12   // Could be derived from instance ID
	sequence := rand.Int63n(4096) // 12-bit sequence
	return timestamp | machineID | sequence
}

// ShoppingCartAPIDynamo handles shopping cart operations using DynamoDB
type ShoppingCartAPIDynamo struct {
	dynamoClient *dynamodb.Client
}

// NewShoppingCartAPIDynamo creates a new shopping cart API handler
func NewShoppingCartAPIDynamo(client *dynamodb.Client) *ShoppingCartAPIDynamo {
	return &ShoppingCartAPIDynamo{dynamoClient: client}
}

// CreateShoppingCart creates a new shopping cart for a customer
func (api *ShoppingCartAPIDynamo) CreateShoppingCart(c *gin.Context) {
	var req CreateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cartID := generateSnowflakeID()
	cart := map[string]types.AttributeValue{
		"cart_id":     &types.AttributeValueMemberN{Value: strconv.FormatInt(cartID, 10)},
		"customer_id": &types.AttributeValueMemberN{Value: strconv.Itoa(req.CustomerID)},
		"status":      &types.AttributeValueMemberS{Value: "active"},
		"created_at":  &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().Unix(), 10)},
	}

	_, err := api.dynamoClient.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String("shopping_carts"),
		Item:      cart,
	})

	if err != nil {
		log.Printf("Failed to create shopping cart: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cart"})
		return
	}

	response := CreateCartResponse{
		ShoppingCartID: int(cartID), // Convert Snowflake ID to int for API compatibility
		CustomerID:     req.CustomerID,
		Status:         "active",
	}

	c.JSON(http.StatusCreated, response)
}

// GetShoppingCart retrieves a shopping cart with all items
func (api *ShoppingCartAPIDynamo) GetShoppingCart(c *gin.Context) {
	cartID := c.Param("id")
	if cartID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
		return
	}

	// Get cart
	result, err := api.dynamoClient.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String("shopping_carts"),
		Key: map[string]types.AttributeValue{
			"cart_id": &types.AttributeValueMemberN{Value: cartID},
		},
	})

	if err != nil {
		log.Printf("Failed to get shopping cart: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cart"})
		return
	}

	if result.Item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	// Get cart items using Query
	itemsResult, err := api.dynamoClient.Query(context.Background(), &dynamodb.QueryInput{
		TableName:              aws.String("shopping_cart_items"),
		KeyConditionExpression: aws.String("cart_id = :cid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":cid": &types.AttributeValueMemberN{Value: cartID},
		},
	})

	if err != nil {
		log.Printf("Failed to get cart items: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cart items"})
		return
	}

	// Build response
	customerID, _ := strconv.Atoi(result.Item["customer_id"].(*types.AttributeValueMemberN).Value)
	shoppingCartID, _ := strconv.Atoi(result.Item["cart_id"].(*types.AttributeValueMemberN).Value)
	cart := ShoppingCart{
		ShoppingCartID: shoppingCartID,
		CustomerID:     customerID,
		Status:         result.Item["status"].(*types.AttributeValueMemberS).Value,
		Items:          make([]ShoppingCartItem, 0),
	}

	for _, item := range itemsResult.Items {
		productID, _ := strconv.Atoi(item["product_id"].(*types.AttributeValueMemberN).Value)
		quantity, _ := strconv.Atoi(item["quantity"].(*types.AttributeValueMemberN).Value)
		cart.Items = append(cart.Items, ShoppingCartItem{
			ProductID: productID,
			Quantity:  quantity,
		})
	}

	c.JSON(http.StatusOK, cart)
}

// AddItemToCart adds an item to the shopping cart
func (api *ShoppingCartAPIDynamo) AddItemToCart(c *gin.Context) {
	cartID := c.Param("id")
	if cartID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
		return
	}

	var req AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify cart exists
	cartResult, err := api.dynamoClient.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String("shopping_carts"),
		Key: map[string]types.AttributeValue{
			"cart_id": &types.AttributeValueMemberN{Value: cartID},
		},
	})

	if err != nil {
		log.Printf("Failed to verify cart: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify cart"})
		return
	}

	if cartResult.Item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	// Create or update item
	itemID := cartID + "-" + strconv.Itoa(req.ProductID)
	item := map[string]types.AttributeValue{
		"cart_id":    &types.AttributeValueMemberN{Value: cartID},
		"item_id":    &types.AttributeValueMemberS{Value: itemID},
		"product_id": &types.AttributeValueMemberN{Value: strconv.Itoa(req.ProductID)},
		"quantity":   &types.AttributeValueMemberN{Value: strconv.Itoa(req.Quantity)},
		"added_at":   &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().Unix(), 10)},
	}

	_, err = api.dynamoClient.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String("shopping_cart_items"),
		Item:      item,
	})

	if err != nil {
		log.Printf("Failed to add item to cart: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item"})
		return
	}

	response := AddItemResponse{
		CartItemID:     0, // DynamoDB doesn't use auto-increment
		ShoppingCartID: req.ProductID,
		ProductID:      req.ProductID,
		Quantity:       req.Quantity,
	}

	c.JSON(http.StatusOK, response)
}
