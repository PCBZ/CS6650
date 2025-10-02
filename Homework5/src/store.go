package main

import (
	"sync"
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
