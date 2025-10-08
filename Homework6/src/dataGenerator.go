package main

import (
	"strconv"
)

// DataGenerator handles the generation of sample product data
type DataGenerator struct {
	categories []string
	brands     []string
}

// NewDataGenerator creates a new DataGenerator instance
func NewDataGenerator() *DataGenerator {
	return &DataGenerator{
		categories: []string{
			"Electronics", "Books", "Home", "Sports", "Clothing",
			"Toys", "Beauty", "Automotive", "Food", "Health",
			"Music", "Movies", "Games", "Tools", "Jewelry",
			"Pet Supplies", "Office", "Baby", "Travel", "Art",
		},
		brands: []string{
			"Alpha", "Beta", "Gamma", "Delta", "Epsilon",
			"Zeta", "Eta", "Theta", "Iota", "Kappa",
			"Lambda", "Mu", "Nu", "Xi", "Omicron",
			"Pi", "Rho", "Sigma", "Tau", "Upsilon",
		},
	}
}

// GenerateProducts creates and returns a list of sample products
func (dg *DataGenerator) GenerateProducts() []*Product {
	products := make([]*Product, 0, 100000)

	for i := 1; i <= 100000; i++ {
		// Use modulo to rotate through categories and brands for variety
		categoryIndex := (i - 1) % len(dg.categories)
		brandIndex := (i - 1) % len(dg.brands)

		product := &Product{
			ProductID:    int32(i),
			SKU:          "SKU-" + strconv.Itoa(i),
			Manufacturer: "Manufacturer " + strconv.Itoa((i%50)+1),
			CategoryID:   int32(categoryIndex + 1),
			Weight:       int32((i % 1000) + 1), // Weight between 1-1000
			SomeOtherID:  int32((i % 999) + 1),
			Category:     dg.categories[categoryIndex],
			Description:  "High-quality product with advanced features",
			Brand:        "Product " + dg.brands[brandIndex] + " " + strconv.Itoa(i),
		}

		products = append(products, product)
	}
	return products
}
