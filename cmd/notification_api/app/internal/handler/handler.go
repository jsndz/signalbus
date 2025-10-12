package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jsndz/signalbus/cmd/notification_api/app/internal/services"
	"github.com/jsndz/signalbus/pkg/gosms"
	"github.com/jsndz/signalbus/pkg/kafka"
	"github.com/jsndz/signalbus/pkg/models"
	"github.com/jsndz/signalbus/pkg/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	service *services.NotificationService
}

func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{service: services.NewNotificationService(db)}
}

func (h *NotificationHandler) Notify(p *kafka.Producer, tdb *gorm.DB, ndb *gorm.DB, log *zap.Logger, tracer trace.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := services.NewTenantService(tdb)
		tracer_context, span := tracer.Start(c.Request.Context(), "handle-sending")
		defer span.End()
		idem_key := c.GetHeader("X-Idempotency-Key")
		tenant_id, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "couldn't get tenant ID ",
			})
			return
		}
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

		tenantUUID, ok := tenant_id.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid tenant ID format",
			})
			return
		}
		tenant, err := s.GetTenant(tenantUUID)
		if err != nil || tenant == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid API key",
			})
			return
		}

		var record models.IdempotencyKey
		err = tdb.Where("key = ? AND tenant_id = ?", idem_key, tenant.ID).First(&record).Error
		if err == nil {
			c.Data(record.StatusCode, "application/json", []byte(record.Response))
			return
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
		if len(tenant.Policies) == 0 {
			log.Warn("tenant has no notification policies configured", zap.String("tenant", tenant.ID.String()))
		}

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
					_, dbSpan := tracer.Start(tracer_context, "create-notification")
					if channel == "sms" {
						num, ok := pl.RecieverData["to"].(string)
						if !ok || num == "" {
							c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid 'to' field for SMS"})
							return
						}
						to, err := gosms.NormalizeSMS(num)
						if err != nil {
							c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
							return
						}
						pl.RecieverData["to"] = to
					}
					notification_id, err := h.service.CreateNotification(tenant.ID, req.EventType, channel, req.UserRef)

					if err != nil {
						dbSpan.RecordError(err)
						dbSpan.SetStatus(codes.Error, err.Error())
						log.Error("failed to create notification", zap.Error(err))
						c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create notification"})
						return
					}
					dbSpan.SetStatus(codes.Ok, "notification created")

					 

					msg := types.KafkaStreamData{
						IdempotencyKey: idem_key,
						RecieverData:   pl.RecieverData,
						InTemplateData: pl.InTemplateData,
						GetTemplateData: &types.GetTemplateData{
							EventType: req.EventType,
							Locale:    locale,
							TenantID:  tenant.ID,
						},
						NotificationId: notification_id,
					}

					msgBytes, err := json.Marshal(msg)
					if err != nil {
						log.Error("failed to marshal kafka message", zap.Error(err))
						continue
					}
					ctx := c.Request.Context()
					key := []byte(tenant.ID.String())
					headers := make(map[string]string)
					otel.GetTextMapPropagator().Inject(tracer_context, propagation.MapCarrier(headers))
					_, kafkaSpan := tracer.Start(tracer_context, "publish-kafka")
					if err := p.Publish(ctx, topic, key, msgBytes); err != nil {
						kafkaSpan.RecordError(err)
						kafkaSpan.SetStatus(codes.Error, err.Error())
						log.Error("failed to publish message", zap.String("topic", topic), zap.Error(err))
					} else {
						log.Debug("published message", zap.String("topic", topic), zap.String("tenant", tenant.ID.String()))
						kafkaSpan.SetStatus(codes.Ok, "published")

					}
					dbSpan.End()
					 kafkaSpan.End()
				}
			}
		}

		if !matched {
			log.Warn("no policy matched event", zap.String("event_type", req.EventType), zap.String("tenant", tenant.ID.String()))
			c.JSON(http.StatusBadRequest, gin.H{"error": "no policy matches event type"})
			span.SetStatus(codes.Error, "no policy matched event")
			return
		}

		resp := gin.H{"message": "Notification accepted"}
		respBytes, _ := json.Marshal(resp)

		if err := tdb.Create(&models.IdempotencyKey{
			Key:         idem_key,
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

func PublishNotification(
	ctx context.Context,
	producer *kafka.Producer,
	logger *zap.Logger,
	tenantID uuid.UUID,
	channel string,
	idempotencyKey string,
	msg types.KafkaStreamData,
) error {
	topic := fmt.Sprintf("notification.%s", channel)

	msg.IdempotencyKey = idempotencyKey

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		logger.Error("failed to marshal Kafka message", zap.Error(err))
		return fmt.Errorf("marshal kafka message: %w", err)
	}

	key := []byte(tenantID.String())

	if err := producer.Publish(ctx, topic, key, msgBytes); err != nil {
		logger.Error("failed to publish Kafka message",
			zap.String("topic", topic),
			zap.Error(err),
		)
		return fmt.Errorf("publish kafka message: %w", err)
	}

	logger.Info("Kafka message published successfully",
		zap.String("topic", topic),
		zap.String("tenant_id", tenantID.String()),
		zap.String("notification_id", msg.NotificationId.String()),
	)
	return nil
}
func (h *NotificationHandler) Publish(
	p *kafka.Producer,
	tdb *gorm.DB,
	ndb *gorm.DB,
	log *zap.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		idemKey := c.GetHeader("X-Idempotency-Key")
		tenantIDVal, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing tenant ID"})
			return
		}
		tenantID, ok := tenantIDVal.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID format"})
			return
		}

		var record models.IdempotencyKey
		if err := tdb.Where("key = ? AND tenant_id = ?", idemKey, tenantID).First(&record).Error; err == nil {
			c.Data(record.StatusCode, "application/json", []byte(record.Response))
			return
		}

		var req types.PublishRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
			return
		}

		payloads, err := splitPayload(req.UserData, nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		if req.Channel == "sms" {
			for _, pl := range payloads {
				if to, ok := pl.RecieverData["to"].(string); ok {
					num, err := gosms.NormalizeSMS(to)
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
						return
					}
					pl.RecieverData["to"] = num
				}
			}
		}

		for _, pl := range payloads {
			notificationID, err := h.service.CreateNotification(
				tenantID, req.EventType, req.Channel, req.UserRef,
			)
			if err != nil {
				log.Error("failed to create notification", zap.Error(err))
				continue
			}

			msg := types.KafkaStreamData{
				RecieverData:   pl.RecieverData,
				TextMessage:    req.TextMessage,
				HTMLMessage:    req.HTMLMessage,
				NotificationId: notificationID,
			}

			if err := PublishNotification(c.Request.Context(), p, log, tenantID, req.Channel, idemKey, msg); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish notification"})
				return
			}
		}

		resp := gin.H{"message": "Notification published"}
		respBytes, _ := json.Marshal(resp)

		if err :=tdb.Create(&models.IdempotencyKey{
			Key:         idemKey,
			TenantID:    tenantID,
			Response:    string(respBytes),
			StatusCode:  http.StatusAccepted,
		});
		err != nil {
			log.Error("failed to create idempotency record", zap.Error(err.Error))
		}

		c.JSON(http.StatusAccepted, resp)
	}
}

func (h *NotificationHandler) Publish1(p *kafka.Producer, tdb *gorm.DB, ndb *gorm.DB, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		idem_key := c.GetHeader("X-Idempotency-Key")
		tenant_id, exists := c.Get("tenant_id")

		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't get tenant ID"})
			return
		}

		tenantID, ok := tenant_id.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID format"})
			return
		}

		var record models.IdempotencyKey
		err := tdb.Where("key = ? AND tenant_id = ?", idem_key, tenantID).First(&record).Error
		if err == nil {
			c.Data(record.StatusCode, "application/json", []byte(record.Response))
			return
		}

		var req types.PublishRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't unmarshal JSON: " + err.Error()})
			return
		}

		reqBytes, _ := json.Marshal(struct {
			EventType   string                 `json:"event_type"`
			Channel     string                 `json:"channel"`
			Data        map[string]interface{} `json:"data"`
			TextMessage string                 `json:"text_message,omitempty"`
			HTMLMessage string                 `json:"html_message,omitempty"`
		}{
			EventType:   req.EventType,
			Channel:     req.Channel,
			Data:        req.UserData,
			TextMessage: req.TextMessage,
			HTMLMessage: req.HTMLMessage,
		})

		sum := sha256.Sum256(reqBytes)
		reqHash := hex.EncodeToString(sum[:])

		payloads, err := splitPayload(req.UserData, nil)
		if err != nil {
			log.Error("failed to split payload", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		topic := "notification." + req.Channel

		if req.Channel == "sms" {
			for _, pl := range payloads {
				num, ok := pl.RecieverData["to"].(string)
				if !ok || num == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid 'to' field for SMS"})
					return
				}
				to, err := gosms.NormalizeSMS(num)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				pl.RecieverData["to"] = to
			}
		}

		for _, pl := range payloads {
			notification_id, err := h.service.CreateNotification(tenantID, req.EventType,req.Channel, req.UserRef)
			if err != nil {
				log.Error("failed to create notification", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create notification"})
				return
			}

			msg := types.KafkaStreamData{
				IdempotencyKey: idem_key,
				RecieverData:   pl.RecieverData,
				TextMessage:    req.TextMessage,
				HTMLMessage:    req.HTMLMessage,
				NotificationId: notification_id,
			}

			msgBytes, err := json.Marshal(msg)
			if err != nil {
				log.Error("failed to marshal kafka message", zap.Error(err))
				continue
			}

			ctx := c.Request.Context()
			key := []byte(tenantID.String())

			if err := p.Publish(ctx, topic, key, msgBytes); err != nil {
				log.Error("failed to publish message", zap.String("topic", topic), zap.Error(err))
			} else {
				log.Debug("published message", zap.String("topic", topic), zap.String("tenant", tenantID.String()))
			}
		}

		resp := gin.H{"message": "Notification published"}
		respBytes, _ := json.Marshal(resp)

		if err := tdb.Create(&models.IdempotencyKey{
			Key:         idem_key,
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

func (h *NotificationHandler) GetNotification(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
			return
		}

		notification, err := h.service.GetNotification(id)
		if err != nil {
			log.Error("failed to fetch notification",
				zap.String("id", idStr),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch notification"})
			return
		}
		if notification == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
			return
		}

		c.JSON(http.StatusOK, notification)
	}
}

func (h *NotificationHandler) RedriveNotification(log *zap.Logger, p *kafka.Producer) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
			return
		}

		attempt, tenantId , err := h.service.DLQNotification(id)
		if err != nil {
			log.Error("failed to fetch DLQ notification",
				zap.String("notification_id", idStr),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch DLQ notification"})
			return
		}

		if attempt.Status != "dlq" {
			c.JSON(http.StatusNotFound, gin.H{"error": "no DLQ delivery attempt found"})
			return
		}

		if len(attempt.Message) == 0 {
			log.Warn("skipping redrive due to empty message payload",
				zap.String("attempt_id", attempt.ID.String()))
			c.JSON(http.StatusBadRequest, gin.H{"error": "empty message payload in DLQ"})
			return
		}

		var msg types.KafkaStreamData
		if err := json.Unmarshal(attempt.Message, &msg); err != nil {
			log.Error("failed to unmarshal DLQ message",
				zap.String("attempt_id", attempt.ID.String()),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid DLQ message format"})
			return
		}

		msg.NotificationId = attempt.NotificationID

		err = PublishNotification(
			c.Request.Context(),
			p,
			log,
			tenantId,
			attempt.Channel,
			"", 
			msg,
		)
		if err != nil {
			log.Error("failed to redrive notification",
				zap.String("attempt_id", attempt.ID.String()),
				zap.String("channel", attempt.Channel),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to republish DLQ message"})
			return
		}

		log.Info("successfully redrived DLQ message",
			zap.String("attempt_id", attempt.ID.String()),
			zap.String("notification_id", attempt.NotificationID.String()),
			zap.String("channel", attempt.Channel))

		c.JSON(http.StatusOK, gin.H{
			"message":          "redrive successful",
			"notification_id":  attempt.NotificationID,
			"channel":          attempt.Channel,
			"previous_attempt": attempt.ID,
		})
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
