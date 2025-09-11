package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jsndz/signalbus/pkg/models"
	"github.com/jsndz/signalbus/pkg/repositories"
	"gorm.io/gorm"
)

type DeliveryAttemptService struct {
    repo *repositories.DeliveryAttemptRepository
}

func NewDeliveryAttemptService(db *gorm.DB) *DeliveryAttemptService {
    return &DeliveryAttemptService{repo: repositories.NewDeliveryAttemptRepository(db)}
}

func (s *DeliveryAttemptService) CreateAttempt(notificationID uuid.UUID, channel, provider, status, errMsg string, try int, latency int64) (*models.DeliveryAttempt, error) {
    if channel == "" || provider == "" {
        return nil, errors.New("channel and provider are required")
    }

    attempt := &models.DeliveryAttempt{
        NotificationID: notificationID,
        Channel:        channel,
        Provider:       provider,
        Status:         status,
        Error:          errMsg,
        Try:            try,
        LatencyMs:      latency,
    }
    if err := s.repo.Create(attempt); err != nil {
        return nil, err
    }
    return attempt, nil
}

func (s *DeliveryAttemptService) GetAttempt(id uuid.UUID) (*models.DeliveryAttempt, error) {
    return s.repo.GetByID(id)
}

func (s *DeliveryAttemptService) ListAttempts(notificationID uuid.UUID) ([]models.DeliveryAttempt, error) {
    return s.repo.ListByNotification(notificationID)
}

func (s *DeliveryAttemptService) DeleteAttempt(id uuid.UUID) error {
    return s.repo.Delete(id)
}
