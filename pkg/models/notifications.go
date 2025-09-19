package models

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
    ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    TenantID  uuid.UUID `gorm:"type:uuid;not null;index"`
    Topic     string    `gorm:"size:100;not null;index"`
    UserRef   string    `gorm:"size:100;index"`
    Status    string    `gorm:"size:50;not null;index"` // pending, delivered, failed
    CreatedAt time.Time `gorm:"autoCreateTime"`

    Tenant Tenant `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE"`
}


type DeliveryAttempt struct {
    ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    NotificationID uuid.UUID `gorm:"type:uuid;not null;index"`
    Channel        string    `gorm:"size:50;not null"`
    Provider       string    `gorm:"size:50;not null"`
    Status         string    `gorm:"size:50;not null"`
    Error          string    `gorm:"type:text"`
    Try            int       `gorm:"not null"`
    LatencyMs      int64     `gorm:"not null"`
    Message        []byte       
    CreatedAt      time.Time `gorm:"autoCreateTime"`

    Notification Notification `gorm:"foreignKey:NotificationID;constraint:OnDelete:CASCADE"`
}
