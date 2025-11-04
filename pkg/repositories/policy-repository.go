package repositories

import (
	"github.com/google/uuid"
	"github.com/jsndz/signalbus/pkg/models"
	"gorm.io/gorm"
)
type PolicyRepository struct {
	db *gorm.DB
}

func NewPolicyRepository(db *gorm.DB) *PolicyRepository {
	return &PolicyRepository{db: db}
}

func (r *PolicyRepository) Create(policy *models.Policy) error {
	return r.db.Create(policy).Error
}

func (r *PolicyRepository) GetByTopic(topic string) (*models.Policy, error) {
	var policy models.Policy
	if err := r.db.First(&policy, "topic = ?", topic).Error; err != nil {
		return nil, err
	}
	return &policy, nil
}

func (r *PolicyRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Policy{}, "id = ?", id).Error
}
