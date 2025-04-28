package database

import (
	"fmt"
	"marketprogo/config"
	"marketprogo/internal/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto migrate all models
	err = DB.AutoMigrate(
		&models.User{},
		&models.Company{},
		&models.Address{},
		&models.Product{},
		&models.ProductImage{},
		&models.Category{},
		&models.InventoryItem{},
		&models.Warehouse{},
		&models.ProductSpecification{},
		&models.Order{},
		&models.OrderItem{},
		&models.Invoice{},
		&models.PurchaseOrder{},
		&models.POItem{},
		&models.Supplier{},
		&models.SupplierContact{},
		&models.Document{},
		&models.Contract{},
		&models.ContractItem{},
		&models.ContractSchedule{},
		&models.ContractOrder{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto migrate database: %v", err)
	}

	return nil
}

func GetDB() *gorm.DB {
	return DB
}
