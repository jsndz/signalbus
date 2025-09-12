package models

import (
	"time"

	"github.com/google/uuid"
)

type Template struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Name        string    `gorm:"size:100;not null;index"` 
	Channel     string    `gorm:"type:varchar(20);not null;index"`
	ContentType string    `gorm:"type:varchar(20);not null;index"`
	Locale      string    `gorm:"size:10;default:'en-US';index"`
	Content     string    `gorm:"type:text;not null"` 
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}