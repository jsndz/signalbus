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
