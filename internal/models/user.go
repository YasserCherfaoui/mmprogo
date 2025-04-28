package models

import (
	"time"

	"gorm.io/gorm"
)

type UserType string

const (
	UserTypeB2C   UserType = "B2C"
	UserTypeB2B   UserType = "B2B"
	UserTypeAdmin UserType = "ADMIN"
)

type User struct {
	gorm.Model
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Phone        string    `json:"phone"`
	UserType     UserType  `gorm:"type:varchar(10);not null" json:"user_type"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	LastLogin    time.Time `json:"last_login"`

	// B2B specific fields
	CompanyID *uint  `json:"company_id"`
	Role      string `json:"role"`

	// Addresses
	Addresses []Address `json:"addresses" gorm:"foreignKey:UserID"`
}

type Company struct {
	gorm.Model
	Name               string  `gorm:"not null" json:"name"`
	VATNumber          string  `json:"vat_number"`
	RegistrationNumber string  `json:"registration_number"`
	Phone              string  `json:"phone"`
	Email              string  `json:"email"`
	Website            string  `json:"website"`
	IsVerified         bool    `gorm:"default:false" json:"is_verified"`
	CreditLimit        float64 `json:"credit_limit"`
	PaymentTerms       int     `json:"payment_terms"` // in days

	// Address
	AddressID uint `json:"address_id"`

	// Users
	Users []User `json:"users" gorm:"foreignKey:CompanyID"`
}

type Address struct {
	gorm.Model
	StreetAddress1 string `gorm:"not null" json:"street_address1"`
	StreetAddress2 string `json:"street_address2"`
	City           string `gorm:"not null" json:"city"`
	State          string `json:"state"`
	PostalCode     string `gorm:"not null" json:"postal_code"`
	Country        string `gorm:"not null" json:"country"`
	IsDefault      bool   `gorm:"default:false" json:"is_default"`

	// Relations
	UserID uint `json:"user_id"`
}
