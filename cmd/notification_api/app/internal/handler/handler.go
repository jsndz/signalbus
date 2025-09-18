package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jsndz/signalbus/cmd/notification_api/app/internal/services"
	"github.com/jsndz/signalbus/pkg/gosms"
	"github.com/jsndz/signalbus/pkg/kafka"
	"github.com/jsndz/signalbus/pkg/models"
	"github.com/jsndz/signalbus/pkg/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	service *services.NotificationService
}

func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{service: services.NewNotificationService(db)}
}

func (h *NotificationHandler) Notify(p *kafka.Producer, tdb *gorm.DB,ndb *gorm.DB, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := services.NewTenantService(tdb)
		apiKey := c.GetHeader("X-API-Key")

		log.Info("Incoming HTTP request",
			zap.String("endpoint", "/notify"),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()),
		)

		var req types.NotifyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "couldn't unmarshal JSON: " + err.Error(),
			})
			return
		}

		tenant, err := s.GetTenantByAPIKey(apiKey)
		if err != nil || tenant == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
			})
			return
		}

		var record models.IdempotencyKey
		if err := tdb.Where("key = ? AND tenant_id = ?", req.IdempotencyKey, tenant.ID).First(&record).Error; err == nil {
			c.Data(record.StatusCode, "application/json", []byte(record.Response))
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
		} else  {
			log.Error("failed to query idempotency key", zap.Error(err))
		}


		reqBytes, _ := json.Marshal(struct {
			EventType    string                 `json:"event_type"`
			Data         map[string]interface{} `json:"data"`
			TemplateData map[string]interface{} `json:"template_data,omitempty"`
		}{
			EventType:    req.EventType,
			Data:         req.UserData,
			TemplateData: req.TemplateData,
		})

		sum := sha256.Sum256(reqBytes)
		reqHash := hex.EncodeToString(sum[:])

		payloads, err := splitPayload(req.UserData, req.TemplateData)
		if err != nil {
			log.Error("failed to split payload", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		matched := false
		ctx := context.Background()

		for _, policy := range tenant.Policies {
			if policy.Topic != req.EventType {
				continue
			}
			matched = true

			locale := "en-US"
			if v, ok := req.TemplateData["locale"]; ok {
				if sLocale, ok := v.(string); ok && sLocale != "" {
					locale = sLocale
				}
			}

			for _, channel := range policy.Channels {
				topic := "notification." + channel
				for _, pl := range payloads {
					notification_id,_:= h.service.CreateNotification(policy.TenantID,req.EventType,req.UserRef)	

					msg := types.KafkaStreamData{
						IdempotencyKey: req.IdempotencyKey,
						RecieverData:   pl.RecieverData,
						InTemplateData: pl.InTemplateData,
						GetTemplateData: &types.GetTemplateData{
							EventType: req.EventType,
							Locale:    locale,
							TenantID:  tenant.ID,
						},
						NotificationId:notification_id,
					}
					if channel == "sms"{
						num ,ok :=pl.RecieverData["to"].(string)
						if !ok || num == "" {
							c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid 'to' field for SMS"})
        					return
						}
						to,err := gosms.NormalizeSMS(num)
						if err != nil{
							c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        					return
						}
						pl.RecieverData["to"] = to
					}
					msgBytes, err := json.Marshal(msg)
					if err != nil {
						log.Error("failed to marshal kafka message", zap.Error(err))
						continue
					}
					key := []byte(tenant.ID.String())

					if err := p.Publish(ctx, topic, key, msgBytes); err != nil {
						log.Error("failed to publish message", zap.String("topic", topic), zap.Error(err))
					} else {
						log.Debug("published message", zap.String("topic", topic), zap.String("tenant", tenant.ID.String()))
						
					}
				}
			}
		}

		if !matched {
			log.Warn("no policy matched event", zap.String("event_type", req.EventType), zap.String("tenant", tenant.ID.String()))
			c.JSON(http.StatusBadRequest, gin.H{"error": "no policy matches event type"})
			return
		}

		resp := gin.H{"message": "Notification accepted"}
		respBytes, _ := json.Marshal(resp)

		if err := tdb.Create(&models.IdempotencyKey{
			Key:         req.IdempotencyKey,
			TenantID:    tenant.ID,
			RequestHash: reqHash,
			Response:    string(respBytes),
			StatusCode:  http.StatusAccepted,
		}).Error; err != nil {
			log.Error("failed to create idempotency record", zap.Error(err))
		}
		c.JSON(http.StatusAccepted, resp)
	}
}

func Publish(p *kafka.Producer, db *gorm.DB, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := services.NewTenantService(db)
		apiKey := c.GetHeader("X-API-Key")
		topicParam := c.Param("topic")

		log.Info("Incoming HTTP request",
			zap.String("endpoint", "/publish/"+topicParam),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()),
		)

		var req types.PublishRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "couldn't unmarshal JSON: " + err.Error(),
			})
			return
		}

		tenantID, err := s.CheckForValidTenant(apiKey)
		if err != nil || tenantID == uuid.Nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
			})
			return
		}

		var record models.IdempotencyKey
		if err := db.Where("key = ? AND tenant_id = ?", req.IdempotencyKey, tenantID).First(&record).Error; err == nil {
			c.Data(record.StatusCode, "application/json", []byte(record.Response))
			return
		}

		msg := types.PublishMessage{
			IdempotencyKey: req.IdempotencyKey,
			Content:        req.Message,
		}
		msgBytes, err := json.Marshal(msg)
		if err != nil {
			log.Error("failed to marshal message", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		ctx := context.Background()
		if err := p.Publish(ctx, "notification." + topicParam, []byte(tenantID.String()), msgBytes); err != nil {
			log.Error("failed to publish message", zap.String("topic", topicParam), zap.Error(err))
		}

		resp := gin.H{"message": "Notification accepted"}
		respBytes, _ := json.Marshal(resp)

		sum := sha256.Sum256(msgBytes)
		reqHash := hex.EncodeToString(sum[:])

		if err := db.Create(&models.IdempotencyKey{
			Key:         req.IdempotencyKey,
			TenantID:    tenantID,
			RequestHash: reqHash,
			Response:    string(respBytes),
			StatusCode:  http.StatusAccepted,
		}).Error; err != nil {
			log.Error("failed to create idempotency record", zap.Error(err))
		}

		c.JSON(http.StatusAccepted, resp)
	}
}

type normalizedPayload struct {
	RecieverData   map[string]interface{}
	InTemplateData map[string]interface{}
}

func splitPayload(data map[string]interface{}, templateData map[string]interface{}) ([]normalizedPayload, error) {
	candidateKeys := []string{"recipients", "receivers", "users", "to", "targets"}

	if data == nil {
		return []normalizedPayload{{RecieverData: map[string]interface{}{}, InTemplateData: templateData}}, nil
	}

	for _, key := range candidateKeys {
		if raw, ok := data[key]; ok {
			switch arr := raw.(type) {
			case []interface{}:
				out := make([]normalizedPayload, 0, len(arr))
				for _, item := range arr {
					rdata := copyMap(data)
					delete(rdata, key)
					rec := make(map[string]interface{})
					switch it := item.(type) {
					case map[string]interface{}:
						rec = it
					default:
						rec["value"] = it
					}
					out = append(out, normalizedPayload{
						RecieverData:   rec,
						InTemplateData: templateData,
					})
				}
				return out, nil

			case map[string]interface{}:
				rec := arr
				return []normalizedPayload{{RecieverData: rec, InTemplateData: templateData}}, nil

			default:
				rec := map[string]interface{}{"value": arr}
				return []normalizedPayload{{RecieverData: rec, InTemplateData: templateData}}, nil
			}
		}
	}
	return []normalizedPayload{{RecieverData: data, InTemplateData: templateData}}, nil
}

func copyMap(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
