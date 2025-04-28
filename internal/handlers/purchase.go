package handlers

import (
	"marketprogo/internal/models"
	"marketprogo/pkg/database"
	"marketprogo/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CreatePurchaseOrderRequest struct {
	SupplierID     uint            `json:"supplier_id" binding:"required"`
	ExpectedDate   time.Time       `json:"expected_date" binding:"required"`
	Currency       string          `json:"currency" binding:"required"`
	ExchangeRate   float64         `json:"exchange_rate" binding:"required"`
	ShippingMethod string          `json:"shipping_method" binding:"required"`
	Notes          string          `json:"notes"`
	Items          []POItemRequest `json:"items" binding:"required,min=1"`
}

type POItemRequest struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	UnitPrice float64 `json:"unit_price" binding:"required,min=0"`
}

type UpdatePurchaseOrderRequest struct {
	Status          string    `json:"status"`
	ExpectedDate    time.Time `json:"expected_date"`
	ContainerNumber string    `json:"container_number"`
	TrackingNumber  string    `json:"tracking_number"`
	Notes           string    `json:"notes"`
	InternalNotes   string    `json:"internal_notes"`
}

func GetPurchaseOrders(c *gin.Context) {
	var pos []models.PurchaseOrder
	query := database.GetDB().Preload("Supplier").Preload("Items.Product")

	// Apply filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if supplierID := c.Query("supplier_id"); supplierID != "" {
		query = query.Where("supplier_id = ?", supplierID)
	}

	if err := query.Find(&pos).Error; err != nil {
		logger.Error.Printf("Failed to get purchase orders: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get purchase orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"purchase_orders": pos})
}

func GetPurchaseOrder(c *gin.Context) {
	id := c.Param("id")
	var po models.PurchaseOrder

	if err := database.GetDB().Preload("Supplier").
		Preload("Items.Product").
		Preload("Documents").
		First(&po, id).Error; err != nil {
		logger.Error.Printf("Failed to get purchase order: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"purchase_order": po})
}

func CreatePurchaseOrder(c *gin.Context) {
	var req CreatePurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx := database.GetDB().Begin()

	// Create purchase order
	po := models.PurchaseOrder{
		SupplierID:     req.SupplierID,
		Status:         models.POStatusDraft,
		OrderDate:      time.Now(),
		ExpectedDate:   req.ExpectedDate,
		Currency:       req.Currency,
		ExchangeRate:   req.ExchangeRate,
		ShippingMethod: req.ShippingMethod,
		Notes:          req.Notes,
	}

	// Create PO items
	var items []models.POItem
	var totalAmount float64

	for _, itemReq := range req.Items {
		var product models.Product
		if err := tx.First(&product, itemReq.ProductID).Error; err != nil {
			tx.Rollback()
			logger.Error.Printf("Failed to find product: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
			return
		}

		item := models.POItem{
			ProductID:   product.ID,
			Quantity:    itemReq.Quantity,
			UnitPrice:   itemReq.UnitPrice,
			TotalAmount: itemReq.UnitPrice * float64(itemReq.Quantity),
			Status:      "pending",
		}

		items = append(items, item)
		totalAmount += item.TotalAmount
	}

	po.TotalAmount = totalAmount
	po.FinalAmount = totalAmount // Add shipping and tax calculations here

	// Create purchase order
	if err := tx.Create(&po).Error; err != nil {
		tx.Rollback()
		logger.Error.Printf("Failed to create purchase order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create purchase order"})
		return
	}

	// Create PO items
	for i := range items {
		items[i].POID = po.ID
		if err := tx.Create(&items[i]).Error; err != nil {
			tx.Rollback()
			logger.Error.Printf("Failed to create PO item: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create purchase order"})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Purchase order created successfully",
		"purchase_order": po,
	})
}

func UpdatePurchaseOrder(c *gin.Context) {
	id := c.Param("id")
	var req UpdatePurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var po models.PurchaseOrder
	if err := database.GetDB().First(&po, id).Error; err != nil {
		logger.Error.Printf("Failed to find purchase order: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase order not found"})
		return
	}

	// Update fields
	if req.Status != "" {
		po.Status = models.POStatus(req.Status)
	}
	if !req.ExpectedDate.IsZero() {
		po.ExpectedDate = req.ExpectedDate
	}
	if req.ContainerNumber != "" {
		po.ContainerNumber = req.ContainerNumber
	}
	if req.TrackingNumber != "" {
		po.TrackingNumber = req.TrackingNumber
	}
	if req.Notes != "" {
		po.Notes = req.Notes
	}
	if req.InternalNotes != "" {
		po.InternalNotes = req.InternalNotes
	}

	// Update status dates
	if po.Status == models.POStatusShipped && po.ShippedDate == nil {
		now := time.Now()
		po.ShippedDate = &now
	}
	if po.Status == models.POStatusReceived && po.ReceivedDate == nil {
		now := time.Now()
		po.ReceivedDate = &now
	}

	if err := database.GetDB().Save(&po).Error; err != nil {
		logger.Error.Printf("Failed to update purchase order: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update purchase order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Purchase order updated successfully",
		"purchase_order": po,
	})
}
