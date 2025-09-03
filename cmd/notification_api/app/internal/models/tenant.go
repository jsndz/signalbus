package models

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
    ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Name          string    `gorm:"size:100;not null"`
    QuotaDaily    int       `gorm:"default:1000"`
    QuotaMonthly  int       `gorm:"default:30000"`
    CreatedAt     time.Time `gorm:"autoCreateTime"`
    UpdatedAt     time.Time `gorm:"autoUpdateTime"`

    APIKeys  []APIKey  `gorm:"constraint:OnDelete:CASCADE"`
    Policies []Policy  `gorm:"constraint:OnDelete:CASCADE"`
}

type APIKey struct {
    ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    TenantID  uuid.UUID `gorm:"type:uuid;not null;index"`
    Hash      string    `gorm:"not null"`
    Scopes    []string  `gorm:"type:text[]"`
    CreatedAt time.Time `gorm:"autoCreateTime"`

    Tenant    Tenant     `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
}

type Policy struct {
    ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    TenantID  uuid.UUID `gorm:"type:uuid;not null;index"`
    Topic     string    `gorm:"size:100;not null"`
    Channels  []string  `gorm:"type:text[];not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`

    Tenant    Tenant	`gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
}