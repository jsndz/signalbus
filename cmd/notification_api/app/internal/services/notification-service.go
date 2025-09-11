package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jsndz/signalbus/pkg/models"
	"github.com/jsndz/signalbus/pkg/repositories"
	"gorm.io/gorm"
)

type NotificationService struct {
    repo *repositories.NotificationRepository
}

func NewNotificationService(db *gorm.DB) *NotificationService {
    return &NotificationService{repo: repositories.NewNotificationRepository(db)}
}

func (s *NotificationService) CreateNotification(tenantID uuid.UUID, topic, userRef string) (*models.Notification, error) {
    if topic == "" {
        return nil, errors.New("notification topic cannot be empty")
    }
    notification := &models.Notification{
        TenantID: tenantID,
        Topic:    topic,
        UserRef:  userRef,
        Status:   "pending",
    }
    if err := s.repo.Create(notification); err != nil {
        return nil, err
    }
    return notification, nil
}

func (s *NotificationService) GetNotification(id uuid.UUID) (*models.Notification, error) {
    return s.repo.GetByID(id)
}

func (s *NotificationService) ListNotifications() ([]models.Notification, error) {
    return s.repo.List()
}

func (s *NotificationService) DeleteNotification(id uuid.UUID) error {
    return s.repo.Delete(id)
}

func (s *NotificationService) UpdateStatus(id uuid.UUID, status string) error {
    return s.repo.UpdateStatus(id, status)
}
