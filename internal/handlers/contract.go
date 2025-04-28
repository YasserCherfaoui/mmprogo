package handlers

import (
	"marketprogo/internal/models"
	"marketprogo/pkg/database"
	"marketprogo/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateContractRequest struct {
	CompanyID     uint                     `json:"company_id" binding:"required"`
	StartDate     time.Time                `json:"start_date" binding:"required"`
	EndDate       time.Time                `json:"end_date" binding:"required"`
	AutoRenew     bool                     `json:"auto_renew"`
	RenewalPeriod int                      `json:"renewal_period"` // in months
	PaymentTerms  int                      `json:"payment_terms"`  // in days
	Notes         string                   `json:"notes"`
	Items         []ContractItemRequest    `json:"items" binding:"required,min=1"`
	Schedule      *ContractScheduleRequest `json:"schedule"`
}

type ContractItemRequest struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	UnitPrice float64 `json:"unit_price" binding:"required,min=0"`
}

type ContractScheduleRequest struct {
	Frequency  string `json:"frequency" binding:"required,oneof=daily weekly monthly"`
	DayOfWeek  *int   `json:"day_of_week"`                    // 0-6 (Sunday-Saturday)
	DayOfMonth *int   `json:"day_of_month"`                   // 1-31
	TimeOfDay  string `json:"time_of_day" binding:"required"` // HH:MM format
}

type UpdateContractRequest struct {
	Status       string `json:"status"`
	AutoRenew    bool   `json:"auto_renew"`
	PaymentTerms int    `json:"payment_terms"`
	Notes        string `json:"notes"`
}

func GetContracts(c *gin.Context) {
	var contracts []models.Contract
	query := database.GetDB().Preload("Company").Preload("Items.Product").Preload("Schedule")

	// Apply filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if companyID := c.Query("company_id"); companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}

	if err := query.Find(&contracts).Error; err != nil {
		logger.Error.Printf("Failed to get contracts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get contracts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"contracts": contracts})
}

func GetContract(c *gin.Context) {
	id := c.Param("id")
	var contract models.Contract

	if err := database.GetDB().Preload("Company").
		Preload("Items.Product").
		Preload("Schedule").
		Preload("Documents").
		First(&contract, id).Error; err != nil {
		logger.Error.Printf("Failed to get contract: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Contract not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"contract": contract})
}

func CreateContract(c *gin.Context) {
	var req CreateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx := database.GetDB().Begin()

	// Create contract
	contract := models.Contract{
		CompanyID:     req.CompanyID,
		Status:        models.ContractStatusDraft,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		AutoRenew:     req.AutoRenew,
		RenewalPeriod: req.RenewalPeriod,
		PaymentTerms:  req.PaymentTerms,
		Notes:         req.Notes,
	}

	// Create contract items
	var items []models.ContractItem
	var totalAmount float64

	for _, itemReq := range req.Items {
		var product models.Product
		if err := tx.First(&product, itemReq.ProductID).Error; err != nil {
			tx.Rollback()
			logger.Error.Printf("Failed to find product: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
			return
		}

		item := models.ContractItem{
			ProductID:   product.ID,
			Quantity:    itemReq.Quantity,
			UnitPrice:   itemReq.UnitPrice,
			TotalAmount: itemReq.UnitPrice * float64(itemReq.Quantity),
			IsActive:    true,
		}

		items = append(items, item)
		totalAmount += item.TotalAmount
	}

	// Create contract schedule if provided
	if req.Schedule != nil {
		schedule := models.ContractSchedule{
			Frequency:  req.Schedule.Frequency,
			DayOfWeek:  req.Schedule.DayOfWeek,
			DayOfMonth: req.Schedule.DayOfMonth,
			TimeOfDay:  req.Schedule.TimeOfDay,
			IsActive:   true,
		}
		contract.Schedule = &schedule
	}

	// Create contract
	if err := tx.Create(&contract).Error; err != nil {
		tx.Rollback()
		logger.Error.Printf("Failed to create contract: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contract"})
		return
	}

	// Create contract items
	for i := range items {
		items[i].ContractID = contract.ID
		if err := tx.Create(&items[i]).Error; err != nil {
			tx.Rollback()
			logger.Error.Printf("Failed to create contract item: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contract"})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Contract created successfully",
		"contract": contract,
	})
}

func UpdateContract(c *gin.Context) {
	id := c.Param("id")
	var req UpdateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var contract models.Contract
	if err := database.GetDB().First(&contract, id).Error; err != nil {
		logger.Error.Printf("Failed to find contract: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Contract not found"})
		return
	}

	// Update fields
	if req.Status != "" {
		contract.Status = models.ContractStatus(req.Status)
	}
	if req.AutoRenew {
		contract.AutoRenew = req.AutoRenew
	}
	if req.PaymentTerms > 0 {
		contract.PaymentTerms = req.PaymentTerms
	}
	if req.Notes != "" {
		contract.Notes = req.Notes
	}

	if err := database.GetDB().Save(&contract).Error; err != nil {
		logger.Error.Printf("Failed to update contract: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contract"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Contract updated successfully",
		"contract": contract,
	})
}
