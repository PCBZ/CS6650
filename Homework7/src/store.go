package main

import (
	"sync"
	"time"
)

// ProductStore provides thread-safe in-memory storage for products using sync.Map
type ProductStore struct {
	products sync.Map // map[int32]*Product
}

// NewProductStore creates a new ProductStore instance
func NewProductStore() *ProductStore {
	return &ProductStore{}
}

// GetProduct retrieves a product by its ID
func (ps *ProductStore) GetProduct(id int32) (*Product, bool) {
	value, exists := ps.products.Load(id)
	if !exists {
		return nil, false
	}

	product, ok := value.(*Product)
	if !ok {
		return nil, false
	}

	return product, true
}

// SetProduct stores a product with the given ID
func (ps *ProductStore) SetProduct(id int32, product *Product) {
	ps.products.Store(id, product)
}

// GetAllProductIDs returns a slice of all product IDs for iteration
func (ps *ProductStore) GetAllProductIDs() []int32 {
	var ids []int32
	ps.products.Range(func(key, value interface{}) bool {
		if id, ok := key.(int32); ok {
			ids = append(ids, id)
		}
		return true
	})
	return ids
}

// GetLimitedProductIDs returns a slice of up to maxCount product IDs for efficient iteration
func (ps *ProductStore) GetLimitedProductIDs(maxCount int) []int32 {
	var ids []int32
	count := 0
	ps.products.Range(func(key, value interface{}) bool {
		if count >= maxCount {
			return false // Stop iteration
		}
		if id, ok := key.(int32); ok {
			ids = append(ids, id)
			count++
		}
		return true
	})
	return ids
}

// GetProductsCount returns the total number of products in the store
func (ps *ProductStore) GetProductsCount() int {
	count := 0
	ps.products.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// === ORDER STORE ===

// OrderStore provides thread-safe in-memory storage for orders
type OrderStore struct {
	orders sync.Map // map[string]*Order
	stats  OrderStats
	mutex  sync.RWMutex
}

// NewOrderStore creates a new OrderStore instance
func NewOrderStore() *OrderStore {
	return &OrderStore{
		stats: OrderStats{},
	}
}

// GetOrder retrieves an order by its ID
func (os *OrderStore) GetOrder(id string) (*Order, bool) {
	value, exists := os.orders.Load(id)
	if !exists {
		return nil, false
	}

	order, ok := value.(*Order)
	if !ok {
		return nil, false
	}

	return order, true
}

// SetOrder stores an order with the given ID and updates stats
func (os *OrderStore) SetOrder(id string, order *Order) {
	os.orders.Store(id, order)
	os.updateStats()
}

// GetStats returns current order processing statistics
func (os *OrderStore) GetStats() OrderStats {
	os.mutex.RLock()
	defer os.mutex.RUnlock()

	os.updateStats()
	return os.stats
}

// updateStats recalculates order statistics
func (os *OrderStore) updateStats() {
	var pending, processing, completed int
	var totalTime time.Duration
	var count int

	os.orders.Range(func(key, value interface{}) bool {
		order, ok := value.(*Order)
		if !ok {
			return true
		}

		count++
		switch order.Status {
		case "pending":
			pending++
		case "processing":
			processing++
		case "completed":
			completed++
			// Calculate processing time for completed orders
			if !order.CreatedAt.IsZero() {
				totalTime += time.Since(order.CreatedAt)
			}
		}
		return true
	})

	os.mutex.Lock()
	defer os.mutex.Unlock()

	os.stats.TotalOrders = count
	os.stats.PendingOrders = pending
	os.stats.ProcessingOrders = processing
	os.stats.CompletedOrders = completed

	if completed > 0 {
		os.stats.AverageTime = float64(totalTime.Milliseconds()) / float64(completed)
	}
}
