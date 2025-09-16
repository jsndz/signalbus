package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jsndz/signalbus/pkg/models"
	"github.com/jsndz/signalbus/pkg/repositories"
	"gorm.io/gorm"
)
type PolicyService struct {
	repo *repositories.PolicyRepository
}

func NewPolicyService(db *gorm.DB) *PolicyService {
	return &PolicyService{repo: repositories.NewPolicyRepository(db)}
}

func (s *PolicyService) CreatePolicy(tenantID uuid.UUID, topic string, channels []string, locale string) (*models.Policy, error) {
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant ID is required")
	}
	if topic == "" {
		return nil, errors.New("topic is required")
	}
	if len(channels) == 0 {
		return nil, errors.New("at least one channel is required")
	}
	if locale == "" {
		return nil, errors.New("locale is required")
	}

	policy := &models.Policy{
		TenantID: tenantID,
		Topic:    topic,
		Channels: channels,
		Locale:   locale,
	}
	if err := s.repo.Create(policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *PolicyService) GetPolicyByID(id uuid.UUID) (*models.Policy, error) {
	return s.repo.GetByID(id)
}

func (s *PolicyService) ListPolicies(tenantID uuid.UUID) ([]models.Policy, error) {
	return s.repo.ListByTenant(tenantID)
}

func (s *PolicyService) DeletePolicy(id uuid.UUID) error {
	return s.repo.Delete(id)
}