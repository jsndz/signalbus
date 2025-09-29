package types

import "github.com/google/uuid"

type NotifyRequest struct {
	EventType      string                 `json:"event_type" binding:"required"`
	UserRef			string					`json:"user_ref" binding:"required"`
	UserData       map[string]interface{} `json:"data" binding:"required"`
	TemplateData   map[string]interface{} `json:"template_data,omitempty"`
}

type KafkaStreamData struct {
	GetTemplateData *GetTemplateData       	`json:"get_template_data"`
	InTemplateData  map[string]interface{} 	`json:"in_template_data,omitempty"`
	RecieverData    map[string]interface{} 	`json:"reciever_data,omitempty"`
	IdempotencyKey  string                 	`json:"idempotency_key"`
	NotificationId  uuid.UUID				`json:"notification_id"`
}

type GetTemplateData struct {
	EventType string    `json:"event_type"`
	Locale    string    `json:"locale"`
	TenantID  uuid.UUID `json:"tenant_id"`
}

type PublishMessage struct {
	IdempotencyKey string `json:"idempotency_key"`
	Content        string `json:"content"`
}

type PublishRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
	Message        string `json:"message"`
}


