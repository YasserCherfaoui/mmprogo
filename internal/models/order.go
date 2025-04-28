package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "PENDING"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusShipped    OrderStatus = "SHIPPED"
	OrderStatusDelivered  OrderStatus = "DELIVERED"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
	OrderStatusReturned   OrderStatus = "RETURNED"
)

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "PENDING"
	PaymentStatusPaid     PaymentStatus = "PAID"
	PaymentStatusFailed   PaymentStatus = "FAILED"
	PaymentStatusRefunded PaymentStatus = "REFUNDED"
)

// TODO: RFQ - Request for Quote
type Order struct {
	gorm.Model
	OrderNumber    string        `gorm:"uniqueIndex;not null" json:"order_number"`
	UserID         uint          `json:"user_id"`
	User           User          `json:"user"`
	CompanyID      *uint         `json:"company_id,omitempty"`
	Company        *Company      `json:"company,omitempty"`
	Status         OrderStatus   `gorm:"type:varchar(20);not null" json:"status"`
	PaymentStatus  PaymentStatus `gorm:"type:varchar(20);not null" json:"payment_status"`
	TotalAmount    float64       `gorm:"not null" json:"total_amount"`
	TaxAmount      float64       `json:"tax_amount"`
	ShippingAmount float64       `json:"shipping_amount"`
	DiscountAmount float64       `json:"discount_amount"`
	FinalAmount    float64       `gorm:"not null" json:"final_amount"`

	// Shipping
	ShippingAddressID uint    `json:"shipping_address_id"`
	ShippingAddress   Address `json:"shipping_address"`
	ShippingMethod    string  `json:"shipping_method"`
	TrackingNumber    string  `json:"tracking_number"`

	// Payment
	PaymentMethod    string     `json:"payment_method"`
	PaymentReference string     `json:"payment_reference"`
	PaymentDate      *time.Time `json:"payment_date"`

	// Order Items
	Items []OrderItem `json:"items"`

	// Notes
	CustomerNotes string `json:"customer_notes"`
	AdminNotes    string `json:"admin_notes"`

	// Dates
	OrderDate     time.Time  `gorm:"not null" json:"order_date"`
	ShippedDate   *time.Time `json:"shipped_date"`
	DeliveredDate *time.Time `json:"delivered_date"`
}

type OrderItem struct {
	gorm.Model
	OrderID        uint    `json:"order_id"`
	Order          Order   `json:"-"`
	ProductID      uint    `json:"product_id"`
	Product        Product `json:"product"`
	Quantity       int     `gorm:"not null" json:"quantity"`
	UnitPrice      float64 `gorm:"not null" json:"unit_price"`
	TaxAmount      float64 `json:"tax_amount"`
	DiscountAmount float64 `json:"discount_amount"`
	TotalAmount    float64 `gorm:"not null" json:"total_amount"`

	// Inventory tracking
	InventoryItemID *uint          `json:"inventory_item_id,omitempty"`
	InventoryItem   *InventoryItem `json:"inventory_item,omitempty"`

	// Status
	Status string `gorm:"default:'active'" json:"status"` // active, cancelled, returned
}

type Invoice struct {
	gorm.Model
	OrderID          uint       `json:"order_id"`
	Order            Order      `json:"order"`
	InvoiceNumber    string     `gorm:"uniqueIndex;not null" json:"invoice_number"`
	IssueDate        time.Time  `gorm:"not null" json:"issue_date"`
	DueDate          time.Time  `json:"due_date"`
	Amount           float64    `gorm:"not null" json:"amount"`
	TaxAmount        float64    `json:"tax_amount"`
	Status           string     `gorm:"default:'pending'" json:"status"` // pending, paid, overdue, cancelled
	PaymentDate      *time.Time `json:"payment_date"`
	PaymentMethod    string     `json:"payment_method"`
	PaymentReference string     `json:"payment_reference"`
	Notes            string     `json:"notes"`
}
