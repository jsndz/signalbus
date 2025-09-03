package database

import (
	"log"

	"gorm.io/gorm"
)

func MigrateDB(db *gorm.DB,models ...interface{}) error{
	err:= db.AutoMigrate(
		models...
	)
	if err!=nil {
		return err
	}
	log.Printf("Database migrated successfully with %d models", len(models))
	return nil
}