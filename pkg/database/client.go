package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


func InitDB(dsn string) (*gorm.DB,error){
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Coudn't run postgres")
	}
	return db,nil

}