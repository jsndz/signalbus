package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)



type Policy struct {
    ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Topic     string    `gorm:"size:100;not null"`
    Channels  pq.StringArray  `gorm:"type:text[];not null"`
    Locale    string `gorm:"size:100;not null"`
    CreatedAt time.Time `gorm:"autoCreateTime"`

}

type IdempotencyKey struct {
    Key       string    `gorm:"primaryKey;size:64;index:idx_tenant_key"`
    RequestHash string  `gorm:"not null"`
    Response   string   `gorm:"type:jsonb"` 
    StatusCode int      `gorm:"not null"`
    CreatedAt  time.Time `gorm:"autoCreateTime"`


}


