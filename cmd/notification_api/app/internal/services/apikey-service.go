package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jsndz/signalbus/pkg/models"
	"github.com/jsndz/signalbus/pkg/repositories"
	"gorm.io/gorm"
)

type APIKeyService struct {
	repo *repositories.APIKeyRepository
}

func NewAPIKeyService(db *gorm.DB) *APIKeyService {
	return &APIKeyService{repo: repositories.NewAPIKeyRepository(db)}
}

func (s *APIKeyService) CreateAPIKey(tenantID uuid.UUID, hash string, scopes []string) (*models.APIKey, error) {
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant ID is required")
	}
	if hash == "" {
		return nil, errors.New("hash is required")
	}

	apiKey := &models.APIKey{
		TenantID: tenantID,
		Hash:     hash,
		Scopes:   scopes,
	}
	if err := s.repo.Create(apiKey); err != nil {
		return nil, err
	}
	return apiKey, nil
}

func (s *APIKeyService) GetAPIKeyByID(id uuid.UUID) (*models.APIKey, error) {
	return s.repo.GetByID(id)
}

func (s *APIKeyService) ListAPIKeys(tenantID uuid.UUID) ([]models.APIKey, error) {
	return s.repo.ListByTenant(tenantID)
}

func (s *APIKeyService) DeleteAPIKey(id uuid.UUID) error {
	return s.repo.Delete(id)
}
