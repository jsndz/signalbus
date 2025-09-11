package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jsndz/signalbus/pkg/models"
	"github.com/jsndz/signalbus/pkg/repositories"
	"gorm.io/gorm"
)

type TenantService struct {
	repo *repositories.TenantRepository
}

func NewTenantService(db *gorm.DB) *TenantService {
	return &TenantService{repo: repositories.NewTenantRepository(db)}
}

func (s *TenantService) CreateTenant(name string) (*models.Tenant, error) {
	tenant := &models.Tenant{Name: name}
	if err := s.repo.CreateTenant(tenant); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *TenantService) GetTenant(id uuid.UUID) (*models.Tenant, error) {
	return s.repo.GetTenantByID(id)
}

func (s *TenantService) ListTenants() ([]models.Tenant, error) {
	return s.repo.ListTenants()
}

func (s *TenantService) DeleteTenant(id uuid.UUID) error {
	return s.repo.DeleteTenant(id)
}

func (s *TenantService) CreateAPIKey(tenantID uuid.UUID, hash string, scopes []string) (*models.APIKey, error) {
	apiKey := &models.APIKey{TenantID: tenantID, Hash: hash, Scopes: scopes}
	if err := s.repo.CreateAPIKey(apiKey); err != nil {
		return nil, err
	}
	return apiKey, nil
}

func (s *TenantService) CreatePolicy(tenantID uuid.UUID, topic string, channels []string) (*models.Policy, error) {
	if len(channels) == 0 {
		return nil, errors.New("policy must have at least one channel")
	}
	policy := &models.Policy{TenantID: tenantID, Topic: topic, Channels: channels}
	if err := s.repo.CreatePolicy(policy); err != nil {
		return nil, err
	}
	return policy, nil
}


func (s *TenantService) GetTenantByAPIKey(apiKey string) (*models.Tenant, error) {
	tenant ,err := s.repo.GetTenantByAPIKey(apiKey)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}


func (s *TenantService) CheckForValidTenant(apiKey string) (uuid.UUID, error) {
	exist ,err := s.repo.CheckIfTenantExist(apiKey)
	return exist,err
}
