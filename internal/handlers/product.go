package handlers

import (
	"marketprogo/internal/models"
	"marketprogo/pkg/database"
	"marketprogo/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreateProductRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	SKU         string   `json:"sku" binding:"required"`
	Barcode     string   `json:"barcode"`
	BasePrice   float64  `json:"base_price" binding:"required"`
	B2BPrice    float64  `json:"b2b_price"`
	CostPrice   float64  `json:"cost_price"`
	Weight      float64  `json:"weight"`
	WeightUnit  string   `json:"weight_unit"`
	CategoryIDs []uint   `json:"category_ids"`
	Images      []string `json:"images"`
}

type UpdateProductRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Barcode     string   `json:"barcode"`
	BasePrice   float64  `json:"base_price"`
	B2BPrice    float64  `json:"b2b_price"`
	CostPrice   float64  `json:"cost_price"`
	Weight      float64  `json:"weight"`
	WeightUnit  string   `json:"weight_unit"`
	IsActive    *bool    `json:"is_active"`
	CategoryIDs []uint   `json:"category_ids"`
	Images      []string `json:"images"`
}

func GetProducts(c *gin.Context) {
	var products []models.Product
	query := database.GetDB().Preload("Categories").Preload("Images")

	// Apply filters
	if categoryID := c.Query("category_id"); categoryID != "" {
		query = query.Joins("JOIN product_categories ON product_categories.product_id = products.id").
			Where("product_categories.category_id = ?", categoryID)
	}

	if isActive := c.Query("is_active"); isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("is_active = ?", active)
	}

	if err := query.Find(&products).Error; err != nil {
		logger.Error.Printf("Failed to get products: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

func GetProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	if err := database.GetDB().Preload("Categories").
		Preload("Images").
		Preload("Specifications").
		Preload("InventoryItems").
		First(&product, id).Error; err != nil {
		logger.Error.Printf("Failed to get product: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product})
}

func CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		SKU:         req.SKU,
		Barcode:     req.Barcode,
		BasePrice:   req.BasePrice,
		B2BPrice:    req.B2BPrice,
		CostPrice:   req.CostPrice,
		Weight:      req.Weight,
		WeightUnit:  req.WeightUnit,
		IsActive:    true,
	}
	// Create images
	for _, image := range req.Images {
		image := models.ProductImage{
			URL: image,
		}
		product.Images = append(product.Images, image)
	}

	// Start transaction
	tx := database.GetDB().Begin()

	// Create product
	if err := tx.Create(&product).Error; err != nil {
		tx.Rollback()
		logger.Error.Printf("Failed to create product: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	// Associate categories
	if len(req.CategoryIDs) > 0 {
		var categories []models.Category
		if err := tx.Find(&categories, req.CategoryIDs).Error; err != nil {
			tx.Rollback()
			logger.Error.Printf("Failed to find categories: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
			return
		}
		if err := tx.Model(&product).Association("Categories").Replace(categories); err != nil {
			tx.Rollback()
			logger.Error.Printf("Failed to associate categories: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
		"product": product,
	})
}

func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var product models.Product
	if err := database.GetDB().First(&product, id).Error; err != nil {
		logger.Error.Printf("Failed to find product: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Update fields
	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Barcode != "" {
		product.Barcode = req.Barcode
	}
	if req.BasePrice != 0 {
		product.BasePrice = req.BasePrice
	}
	if req.B2BPrice != 0 {
		product.B2BPrice = req.B2BPrice
	}
	if req.CostPrice != 0 {
		product.CostPrice = req.CostPrice
	}
	if req.Weight != 0 {
		product.Weight = req.Weight
	}
	if req.WeightUnit != "" {
		product.WeightUnit = req.WeightUnit
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	// Start transaction
	tx := database.GetDB().Begin()

	// Update product
	if err := tx.Save(&product).Error; err != nil {
		tx.Rollback()
		logger.Error.Printf("Failed to update product: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	// Update categories if provided
	if len(req.CategoryIDs) > 0 {
		var categories []models.Category
		if err := tx.Find(&categories, req.CategoryIDs).Error; err != nil {
			tx.Rollback()
			logger.Error.Printf("Failed to find categories: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
			return
		}
		if err := tx.Model(&product).Association("Categories").Replace(categories); err != nil {
			tx.Rollback()
			logger.Error.Printf("Failed to update categories: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully",
		"product": product,
	})
}

func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	if err := database.GetDB().First(&product, id).Error; err != nil {
		logger.Error.Printf("Failed to find product: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Soft delete
	if err := database.GetDB().Delete(&product).Error; err != nil {
		logger.Error.Printf("Failed to delete product: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
