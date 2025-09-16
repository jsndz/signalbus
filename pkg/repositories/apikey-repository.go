package repositories

import (
	"github.com/google/uuid"
	"github.com/jsndz/signalbus/pkg/models"
	"gorm.io/gorm"
)


type APIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(apiKey *models.APIKey) error {
	return r.db.Create(apiKey).Error
}

func (r *APIKeyRepository) GetByID(id uuid.UUID) (*models.APIKey, error) {
	var apiKey models.APIKey
	if err := r.db.First(&apiKey, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *APIKeyRepository) ListByTenant(tenantID uuid.UUID) ([]models.APIKey, error) {
	var apiKeys []models.APIKey
	if err := r.db.Where("tenant_id = ?", tenantID).Find(&apiKeys).Error; err != nil {
		return nil, err
	}
	return apiKeys, nil
}

func (r *APIKeyRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.APIKey{}, "id = ?", id).Error
}

