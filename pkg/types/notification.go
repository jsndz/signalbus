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

	TextMessage     string                     `json:"text_message,omitempty"`      
    HTMLMessage     string                     `json:"html_message,omitempty"`

	IdempotencyKey  string                 	`json:"idempotency_key"`
	NotificationId  uuid.UUID				`json:"notification_id"`
}

type GetTemplateData struct {
	EventType string    `json:"event_type"`
	Locale    string    `json:"locale"`
}



type PublishRequest struct {
	EventType 	string 					`json:"event_type" binding:"required"`
	Channel 	string 					`json:"channel" binding:"required"`
	UserData    map[string]interface{} 	`json:"data" binding:"required"`
	TextMessage     string 					`json:"text_message"`
	HTMLMessage     string 					`json:"html_message"`
	UserRef			string					`json:"user_ref" binding:"required"`

}


