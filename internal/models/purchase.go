package models

import (
	"time"

	"gorm.io/gorm"
)

type POStatus string

const (
	POStatusDraft     POStatus = "DRAFT"
	POStatusPending   POStatus = "PENDING"
	POStatusApproved  POStatus = "APPROVED"
	POStatusOrdered   POStatus = "ORDERED"
	POStatusShipped   POStatus = "SHIPPED"
	POStatusReceived  POStatus = "RECEIVED"
	POStatusCancelled POStatus = "CANCELLED"
)

type PurchaseOrder struct {
	gorm.Model
	PONumber       string    `gorm:"uniqueIndex;not null" json:"po_number"`
	SupplierID     uint      `json:"supplier_id"`
	Supplier       Supplier  `json:"supplier"`
	Status         POStatus  `gorm:"type:varchar(20);not null" json:"status"`
	OrderDate      time.Time `gorm:"not null" json:"order_date"`
	ExpectedDate   time.Time `json:"expected_date"`
	Currency       string    `gorm:"default:'GBP'" json:"currency"`
	ExchangeRate   float64   `json:"exchange_rate"`
	TotalAmount    float64   `gorm:"not null" json:"total_amount"`
	TaxAmount      float64   `json:"tax_amount"`
	ShippingAmount float64   `json:"shipping_amount"`
	FinalAmount    float64   `gorm:"not null" json:"final_amount"`

	// Shipping
	ShippingMethod  string `json:"shipping_method"`
	ContainerNumber string `json:"container_number"`
	TrackingNumber  string `json:"tracking_number"`

	// Items
	Items []POItem `json:"items" gorm:"foreignKey:POID"`

	// Documents
	Documents []Document `json:"documents" gorm:"many2many:purchase_order_documents;"`

	// Notes
	Notes         string `json:"notes"`
	InternalNotes string `json:"internal_notes"`

	// Dates
	ApprovedDate *time.Time `json:"approved_date"`
	ShippedDate  *time.Time `json:"shipped_date"`
	ReceivedDate *time.Time `json:"received_date"`
}

type POItem struct {
	gorm.Model
	POID             uint          `json:"po_id"`
	PurchaseOrder    PurchaseOrder `gorm:"foreignKey:POID" json:"-"`
	ProductID        uint          `json:"product_id"`
	Product          Product       `json:"product"`
	Quantity         int           `gorm:"not null" json:"quantity"`
	UnitPrice        float64       `gorm:"not null" json:"unit_price"`
	TaxAmount        float64       `json:"tax_amount"`
	TotalAmount      float64       `gorm:"not null" json:"total_amount"`
	ReceivedQuantity int           `gorm:"default:0" json:"received_quantity"`
	Status           string        `gorm:"default:'pending'" json:"status"` // pending, partial, complete
}

type Supplier struct {
	gorm.Model
	Name               string `gorm:"not null" json:"name"`
	Code               string `gorm:"uniqueIndex;not null" json:"code"`
	VATNumber          string `json:"vat_number"`
	RegistrationNumber string `json:"registration_number"`
	Phone              string `json:"phone"`
	Email              string `json:"email"`
	Website            string `json:"website"`
	IsActive           bool   `gorm:"default:true" json:"is_active"`

	// Address
	AddressID uint    `json:"address_id"`
	Address   Address `json:"address"`

	// Contacts
	Contacts []SupplierContact `json:"contacts"`

	// Purchase Orders
	PurchaseOrders []PurchaseOrder `json:"purchase_orders"`
}

type SupplierContact struct {
	gorm.Model
	SupplierID uint     `json:"supplier_id"`
	Supplier   Supplier `json:"-"`
	Name       string   `gorm:"not null" json:"name"`
	Position   string   `json:"position"`
	Email      string   `json:"email"`
	Phone      string   `json:"phone"`
	IsPrimary  bool     `gorm:"default:false" json:"is_primary"`
}

type Document struct {
	gorm.Model
	DocumentableID   uint   `json:"documentable_id"`
	DocumentableType string `gorm:"not null" json:"documentable_type"` // PurchaseOrder, Order, etc.
	FileName         string `gorm:"not null" json:"file_name"`
	FileType         string `gorm:"not null" json:"file_type"`
	FileSize         int64  `json:"file_size"`
	URL              string `gorm:"not null" json:"url"`
	Description      string `json:"description"`
}
