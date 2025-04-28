package models

import (
	"time"

	"gorm.io/gorm"
)

type DocumentType string

const (
	DocumentTypeInvoice     DocumentType = "INVOICE"
	DocumentTypePackingSlip DocumentType = "PACKING_SLIP"
	DocumentTypePO          DocumentType = "PURCHASE_ORDER"
	DocumentTypeContract    DocumentType = "CONTRACT"
)

type InDocument struct {
	gorm.Model
	DocumentableID   uint         `json:"documentable_id"`
	DocumentableType string       `gorm:"not null" json:"documentable_type"`
	Type             DocumentType `gorm:"not null" json:"type"`
	FileName         string       `gorm:"not null" json:"file_name"`
	FileType         string       `gorm:"not null" json:"file_type"`
	FileSize         int64        `json:"file_size"`
	S3Bucket         string       `gorm:"not null" json:"s3_bucket"`
	S3Key            string       `gorm:"not null" json:"s3_key"`
	URL              string       `gorm:"not null" json:"url"`
	Description      string       `json:"description"`
	ExpiresAt        *time.Time   `json:"expires_at"`
}
