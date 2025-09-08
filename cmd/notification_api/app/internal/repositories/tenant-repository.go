package repositories

import (
	"github.com/google/uuid"
	"github.com/jsndz/signalbus/cmd/notification_api/app/internal/models"
	"gorm.io/gorm"
)

type TenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

func (r *TenantRepository) CreateTenant(tenant *models.Tenant) error {
	return r.db.Create(tenant).Error
}

func (r *TenantRepository) GetTenantByID(id uuid.UUID) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := r.db.Preload("APIKeys").Preload("Policies").First(&tenant, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepository) ListTenants() ([]models.Tenant, error) {
	var tenants []models.Tenant
	if err := r.db.Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func (r *TenantRepository) DeleteTenant(id uuid.UUID) error {
	return r.db.Delete(&models.Tenant{}, "id = ?", id).Error
}

// API Keys
func (r *TenantRepository) CreateAPIKey(apiKey *models.APIKey) error {
	return r.db.Create(apiKey).Error
}

func (r *TenantRepository) GetAPIKeyByHash(hash string) (*models.APIKey, error) {
	var apiKey models.APIKey
	if err := r.db.First(&apiKey, "hash = ?", hash).Error; err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// Policies
func (r *TenantRepository) CreatePolicy(policy *models.Policy) error {
	return r.db.Create(policy).Error
}

func (r *TenantRepository) GetPoliciesByTenant(tenantID uuid.UUID) ([]models.Policy, error) {
	var policies []models.Policy
	if err := r.db.Where("tenant_id = ?", tenantID).Find(&policies).Error; err != nil {
		return nil, err
	}
	return policies, nil
}


func (r *TenantRepository) GetTenantByAPIKey(apiKey string) (*models.Tenant, error){
	var key models.APIKey
    if err := r.db.Where("hash = ?", apiKey).First(&key).Error; err != nil {
        return nil, err
    }

    var tenant models.Tenant
    if err := r.db. Preload("Policies").First(&tenant, "id = ?", key.TenantID).Error; err != nil {
        return nil, err
    }

    return &tenant, nil
}
