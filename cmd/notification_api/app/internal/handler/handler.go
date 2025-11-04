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
	notificationService *services.NotificationService
	policyService *services.PolicyService
}

func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{notificationService: services.NewNotificationService(db),policyService: services.NewPolicyService(db)}
}

func (h *NotificationHandler) Notify(p *kafka.Producer,db *gorm.DB, log *zap.Logger, tracer trace.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		tracer_context, span := tracer.Start(c.Request.Context(), "handle-sending")
		defer span.End()
		idem_key := c.GetHeader("X-Idempotency-Key")
		
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

		

		var record models.IdempotencyKey
		err := db.Where("key = ? ", idem_key).First(&record).Error
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
		policy,err := h.policyService.GetPolicyByTopic(req.EventType)
		if err != nil {
			log.Error("failed to get policy payload", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}
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
				notification_id, err := h.notificationService.CreateNotification(req.EventType, channel, req.UserRef)

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
					},
					NotificationId: notification_id,
				}

				msgBytes, err := json.Marshal(msg)
				if err != nil {
					log.Error("failed to marshal kafka message", zap.Error(err))
					continue
				}
				ctx := c.Request.Context()
				headers := make(map[string]string)
				otel.GetTextMapPropagator().Inject(tracer_context, propagation.MapCarrier(headers))
				_, kafkaSpan := tracer.Start(tracer_context, "publish-kafka")
				if err := p.Publish(ctx, topic, []byte(idem_key), msgBytes); err != nil {
					kafkaSpan.RecordError(err)
					kafkaSpan.SetStatus(codes.Error, err.Error())
					log.Error("failed to publish message", zap.String("topic", topic), zap.Error(err))
				} else {
					log.Debug("published message", zap.String("topic", topic), )
					kafkaSpan.SetStatus(codes.Ok, "published")

				}
				dbSpan.End()
					kafkaSpan.End()
			}
		}

		resp := gin.H{"message": "Notification accepted"}
		respBytes, _ := json.Marshal(resp)

		if err := db.Create(&models.IdempotencyKey{
			Key:         idem_key,
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

	if err := producer.Publish(ctx, topic, []byte(idempotencyKey), msgBytes); err != nil {
		logger.Error("failed to publish Kafka message",
			zap.String("topic", topic),
			zap.Error(err),
		)
		return fmt.Errorf("publish kafka message: %w", err)
	}

	logger.Info("Kafka message published successfully",
		zap.String("topic", topic),
		zap.String("notification_id", msg.NotificationId.String()),
	)
	return nil
}
func (h *NotificationHandler) Publish(
	p *kafka.Producer,
	db *gorm.DB,
	log *zap.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		idemKey := c.GetHeader("X-Idempotency-Key")
	
		var record models.IdempotencyKey
		if err := db.Where("key = ?", idemKey).First(&record).Error; err == nil {
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
			notificationID, err := h.notificationService.CreateNotification(
				 req.EventType, req.Channel, req.UserRef,
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

			if err := PublishNotification(c.Request.Context(), p, log, req.Channel, idemKey, msg); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish notification"})
				return
			}
		}

		resp := gin.H{"message": "Notification published"}
		respBytes, _ := json.Marshal(resp)

		if err :=db.Create(&models.IdempotencyKey{
			Key:         idemKey,
			Response:    string(respBytes),
			StatusCode:  http.StatusAccepted,
		});
		err != nil {
			log.Error("failed to create idempotency record", zap.Error(err.Error))
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

		notification, err := h.notificationService.GetNotification(id)
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

		attempt , err := h.notificationService.DLQNotification(id)
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
