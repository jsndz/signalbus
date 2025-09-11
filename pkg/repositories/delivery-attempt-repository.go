package repositories

import (
	"github.com/google/uuid"
	"github.com/jsndz/signalbus/pkg/models"
	"gorm.io/gorm"
)

type DeliveryAttemptRepository struct {
    db *gorm.DB
}

func NewDeliveryAttemptRepository(db *gorm.DB) *DeliveryAttemptRepository {
    return &DeliveryAttemptRepository{db: db}
}

func (r *DeliveryAttemptRepository) Create(attempt *models.DeliveryAttempt) error {
    return r.db.Create(attempt).Error
}

func (r *DeliveryAttemptRepository) GetByID(id uuid.UUID) (*models.DeliveryAttempt, error) {
    var attempt models.DeliveryAttempt
    if err := r.db.First(&attempt, "id = ?", id).Error; err != nil {
        return nil, err
    }
    return &attempt, nil
}

func (r *DeliveryAttemptRepository) ListByNotification(notificationID uuid.UUID) ([]models.DeliveryAttempt, error) {
    var attempts []models.DeliveryAttempt
    if err := r.db.Where("notification_id = ?", notificationID).
        Find(&attempts).Error; err != nil {
        return nil, err
    }
    return attempts, nil
}

func (r *DeliveryAttemptRepository) Delete(id uuid.UUID) error {
    return r.db.Delete(&models.DeliveryAttempt{}, "id = ?", id).Error
}
