package migrations

import (
	"github.com/jinzhu/gorm"
	"jobs-svc/internal/models"
	"log"
)

// handle schema changes
func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(&models.Job{}, &models.Application{})
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
	log.Println("Migration applied successfully")
}
