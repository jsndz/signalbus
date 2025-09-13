package repositories

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jsndz/signalbus/pkg/models"
	"gorm.io/gorm"
)

type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) Create(template *models.Template) error {
	return r.db.Create(template).Error
}

func (r *TemplateRepository) GetByID(id uuid.UUID) (*models.Template, error) {
	var template models.Template
	if err := r.db.First(&template, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepository) GetByNameAndTenant(tenantID uuid.UUID, name, locale string) (*models.Template, error) {
	var template models.Template
	if err := r.db.Where("tenant_id = ? AND name = ? AND locale = ?", tenantID, name, locale).
		First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepository) List(tenantID uuid.UUID) ([]models.Template, error) {
	var templates []models.Template
	if err := r.db.Where("tenant_id = ?", tenantID).Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *TemplateRepository) Update(template *models.Template) error {
	if template.ID == uuid.Nil {
		return errors.New("invalid template ID")
	}
	return r.db.Save(template).Error
}

func (r *TemplateRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Template{}, "id = ?", id).Error
}

func (r *TemplateRepository) GetByLookup(tenantID uuid.UUID, channel, name, locale, contentType string) (*models.Template, error) {
	var template models.Template
	err := r.db.Where(
		"tenant_id = ? AND channel = ? AND name = ? AND locale = ? AND content_type = ?",
		tenantID, channel, name, locale, contentType,
	).First(&template).Error

	if err != nil {
		return nil, err
	}
	return &template, nil
}
