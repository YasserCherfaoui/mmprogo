package models

import (
	"time"

	"gorm.io/gorm"
)

type ContractStatus string

const (
	ContractStatusDraft     ContractStatus = "DRAFT"
	ContractStatusActive    ContractStatus = "ACTIVE"
	ContractStatusPaused    ContractStatus = "PAUSED"
	ContractStatusExpired   ContractStatus = "EXPIRED"
	ContractStatusCancelled ContractStatus = "CANCELLED"
)

type Contract struct {
	gorm.Model
	ContractNumber string         `gorm:"uniqueIndex;not null" json:"contract_number"`
	CompanyID      uint           `json:"company_id"`
	Company        *Company       `json:"company"`
	Status         ContractStatus `gorm:"type:varchar(20);not null" json:"status"`
	StartDate      time.Time      `gorm:"not null" json:"start_date"`
	EndDate        time.Time      `json:"end_date"`
	AutoRenew      bool           `gorm:"default:false" json:"auto_renew"`
	RenewalPeriod  int            `json:"renewal_period"` // in months
	PaymentTerms   int            `json:"payment_terms"`  // in days
	Notes          string         `json:"notes"`

	// Items
	Items []ContractItem `json:"items"`

	// Schedule
	Schedule *ContractSchedule `json:"schedule"`

	// Documents
	Documents []Document `json:"documents" gorm:"many2many:contract_documents;"`
}

type ContractItem struct {
	gorm.Model
	ContractID  uint      `json:"contract_id"`
	Contract    *Contract `json:"-"`
	ProductID   uint      `json:"product_id"`
	Product     *Product  `json:"product"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	UnitPrice   float64   `gorm:"not null" json:"unit_price"`
	TaxAmount   float64   `json:"tax_amount"`
	TotalAmount float64   `gorm:"not null" json:"total_amount"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
}

type ContractSchedule struct {
	gorm.Model
	ContractID     uint       `json:"contract_id"`
	Contract       *Contract  `json:"-"`
	Frequency      string     `gorm:"not null" json:"frequency"` // daily, weekly, monthly
	DayOfWeek      *int       `json:"day_of_week"`               // 0-6 (Sunday-Saturday)
	DayOfMonth     *int       `json:"day_of_month"`              // 1-31
	TimeOfDay      string     `json:"time_of_day"`               // HH:MM format
	LastGenerated  *time.Time `json:"last_generated"`
	NextGeneration *time.Time `json:"next_generation"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
}

type ContractOrder struct {
	gorm.Model
	ContractID    uint              `json:"contract_id"`
	Contract      *Contract         `json:"contract"`
	OrderID       uint              `json:"order_id"`
	Order         *Order            `json:"order"`
	ScheduleID    uint              `json:"schedule_id"`
	Schedule      *ContractSchedule `json:"schedule"`
	GeneratedDate time.Time         `gorm:"not null" json:"generated_date"`
	Status        string            `gorm:"default:'pending'" json:"status"` // pending, processed, failed
	Error         string            `json:"error"`
}
