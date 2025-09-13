package models

import (
	"time"

	"github.com/google/uuid"
)

type Template struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Name        string    `gorm:"size:100;not null"`
	Channel     string    `gorm:"type:varchar(20);not null"`
	ContentType string    `gorm:"type:varchar(20);not null"`
	Locale      string    `gorm:"size:10;default:'en-US'"`
	Content     string    `gorm:"type:text;not null"` 
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`

	_ struct{} `gorm:"uniqueIndex:idx_template_unique,priority:1"`
}
