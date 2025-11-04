package repositories

import (
	"github.com/google/uuid"
	"github.com/jsndz/signalbus/pkg/models"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}


func (r *NotificationRepository) Create(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *NotificationRepository) GetByID(id uuid.UUID) (*models.Notification, error) {
	var notification models.Notification
	if err := r.db.First(&notification, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *NotificationRepository) List() ([]models.Notification, error) {
	var notifications []models.Notification
	if err := r.db.Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *NotificationRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Notification{}, "id = ?", id).Error
}

func (r *NotificationRepository) UpdateStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *NotificationRepository) CreateAttempt(attempt *models.DeliveryAttempt) error {
	return r.db.Create(attempt).Error
}

func (r *NotificationRepository) GetAttemptByID(id uuid.UUID) (*models.DeliveryAttempt, error) {
	var attempt models.DeliveryAttempt
	if err := r.db.First(&attempt, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (r *NotificationRepository) ListAttemptsByNotification(notificationID uuid.UUID) ([]models.DeliveryAttempt, error) {
	var attempts []models.DeliveryAttempt
	if err := r.db.Where("notification_id = ?", notificationID).
		Find(&attempts).Error; err != nil {
		return nil, err
	}
	return attempts, nil
}

func (r *NotificationRepository) DeleteAttempt(id uuid.UUID) error {
	return r.db.Delete(&models.DeliveryAttempt{}, "id = ?", id).Error
}

func (r *NotificationRepository) GetDLQByNotificationID(id uuid.UUID) (*models.DeliveryAttempt, error) {
    var out struct {
        
        Message  []byte
        Channel  string
    }

    if err := r.db.Table("delivery_attempts AS da").
        Joins("JOIN notifications n ON n.id = da.notification_id").
        Select(" da.message, da.channel, da.id, da.notification_id, da.provider, da.status, da.error, da.try, da.latency_ms, da.created_at").
        Where("da.notification_id = ? AND da.status = ?", id, "dlq").
        Scan(&out).Error; err != nil {
        return nil, err
    }

    attempt := &models.DeliveryAttempt{
        NotificationID: id,
        Channel:        out.Channel,
        Status:         "dlq",
        Message:        out.Message,
    }

    return attempt, nil
}
