package handlers

import (
	"marketprogo/internal/models"
	"marketprogo/pkg/database"
	"marketprogo/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateOrderRequest struct {
	ShippingAddressID uint               `json:"shipping_address_id" binding:"required"`
	ShippingMethod    string             `json:"shipping_method" binding:"required"`
	PaymentMethod     string             `json:"payment_method" binding:"required"`
	CustomerNotes     string             `json:"customer_notes"`
	Items             []OrderItemRequest `json:"items" binding:"required,min=1"`
}

type OrderItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type UpdateOrderRequest struct {
	Status        string `json:"status"`
	PaymentStatus string `json:"payment_status"`
	AdminNotes    string `json:"admin_notes"`
}

func GetOrders(c *gin.Context) {
	var orders []models.Order
	query := database.GetDB().Preload("User").Preload("Items.Product")

	// Apply filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if paymentStatus := c.Query("payment_status"); paymentStatus != "" {
		query = query.Where("payment_status = ?", paymentStatus)
	}
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&orders).Error; err != nil {
		logger.Error.Printf("Failed to get orders: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func GetOrder(c *gin.Context) {
	id := c.Param("id")
	var order models.Order

	if err := database.GetDB().Preload("User").
		Preload("Items.Product").
		Preload("ShippingAddress").
		First(&order, id).Error; err != nil {
		logger.Error.Printf("Failed to get order: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

func CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get user ID from JWT token
	userID := uint(1) // Temporary hardcoded user ID

	// Start transaction
	tx := database.GetDB().Begin()

	// Create order
	order := models.Order{
		UserID:            userID,
		Status:            models.OrderStatusPending,
		PaymentStatus:     models.PaymentStatusPending,
		ShippingAddressID: req.ShippingAddressID,
		ShippingMethod:    req.ShippingMethod,
		PaymentMethod:     req.PaymentMethod,
		CustomerNotes:     req.CustomerNotes,
		OrderDate:         time.Now(),
	}

	// Calculate order totals
	var totalAmount float64
	var items []models.OrderItem

	for _, itemReq := range req.Items {
		var product models.Product
		if err := tx.First(&product, itemReq.ProductID).Error; err != nil {
			tx.Rollback()
			logger.Error.Printf("Failed to find product: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
			return
		}

		// Check inventory
		var inventory models.InventoryItem
		if err := tx.Where("product_id = ? AND status = 'active'", product.ID).
			First(&inventory).Error; err != nil {
			tx.Rollback()
			logger.Error.Printf("Failed to find inventory: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product out of stock"})
			return
		}

		if inventory.Quantity-inventory.Reserved < itemReq.Quantity {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock"})
			return
		}

		// Create order item
		orderItem := models.OrderItem{
			ProductID:       product.ID,
			Quantity:        itemReq.Quantity,
			UnitPrice:       product.BasePrice,
			TotalAmount:     product.BasePrice * float64(itemReq.Quantity),
			InventoryItemID: &inventory.ID,
		}

		items = append(items, orderItem)
		totalAmount += orderItem.TotalAmount

		// Reserve inventory
		inventory.Reserved += itemReq.Quantity
		if err := tx.Save(&inventory).Error; err != nil {
			tx.Rollback()
			logger.Error.Printf("Failed to update inventory: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
			return
		}
	}

	order.TotalAmount = totalAmount
	order.FinalAmount = totalAmount // Add shipping and tax calculations here

	// Create order
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		logger.Error.Printf("Failed to create order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Create order items
	for i := range items {
		items[i].OrderID = order.ID
		if err := tx.Create(&items[i]).Error; err != nil {
			tx.Rollback()
			logger.Error.Printf("Failed to create order item: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"order":   order,
	})
}

func UpdateOrder(c *gin.Context) {
	id := c.Param("id")
	var req UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.Order
	if err := database.GetDB().First(&order, id).Error; err != nil {
		logger.Error.Printf("Failed to find order: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Update fields
	if req.Status != "" {
		order.Status = models.OrderStatus(req.Status)
	}
	if req.PaymentStatus != "" {
		order.PaymentStatus = models.PaymentStatus(req.PaymentStatus)
	}
	if req.AdminNotes != "" {
		order.AdminNotes = req.AdminNotes
	}

	// Update order status dates
	if order.Status == models.OrderStatusShipped && order.ShippedDate == nil {
		now := time.Now()
		order.ShippedDate = &now
	}
	if order.Status == models.OrderStatusDelivered && order.DeliveredDate == nil {
		now := time.Now()
		order.DeliveredDate = &now
	}

	if err := database.GetDB().Save(&order).Error; err != nil {
		logger.Error.Printf("Failed to update order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order updated successfully",
		"order":   order,
	})
}
