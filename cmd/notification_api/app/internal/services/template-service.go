package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jsndz/signalbus/pkg/models"
	"github.com/jsndz/signalbus/pkg/repositories"
	"gorm.io/gorm"
)

type TemplateService struct {
	repo *repositories.TemplateRepository
}
func NewTemplateService(db *gorm.DB) *TemplateService {
	return &TemplateService{repo: repositories.NewTemplateRepository(db)}
}

func (s *TemplateService) CreateTemplate(template *models.Template) error {
	if template.TenantID == uuid.Nil {
		return errors.New("tenant ID is required")
	}
	if template.Name == "" {
		return errors.New("template name is required")
	}
	if template.Channel == "" {
		return errors.New("channel is required")
	}
	if template.ContentType == "" {
		return errors.New("content type is required")
	}
	if template.Content == "" {
		return errors.New("content is required")
	}

	if template.Locale == "" {
		template.Locale = "en-US"
	}

	return s.repo.Create(template)
}

func (s *TemplateService) GetTemplateByID(id uuid.UUID) (*models.Template, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid template ID")
	}
	return s.repo.GetByID(id)
}

func (s *TemplateService) GetTemplateByName(tenantID uuid.UUID, name, locale string) (*models.Template, error) {
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant ID is required")
	}
	if name == "" {
		return nil, errors.New("template name is required")
	}
	if locale == "" {
		locale = "en-US"
	}
	return s.repo.GetByNameAndTenant(tenantID, name, locale)
}

func (s *TemplateService) ListTemplates(tenantID uuid.UUID) ([]models.Template, error) {
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant ID is required")
	}
	return s.repo.List(tenantID)
}

func (s *TemplateService) UpdateTemplate(template *models.Template) error {
	if template.ID == uuid.Nil {
		return errors.New("invalid template ID")
	}
	return s.repo.Update(template)
}

func (s *TemplateService) DeleteTemplate(id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid template ID")
	}
	return s.repo.Delete(id)
}

func (s *TemplateService) LookupTemplate(tenantID uuid.UUID, channel, name, locale, contentType string) (*models.Template, error) {
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant ID is required")
	}
	if channel == "" || name == "" || contentType == "" {
		return nil, errors.New("channel, name and content type are required")
	}
	if locale == "" {
		locale = "en-US"
	}
	return s.repo.GetByLookup(tenantID, channel, name, locale, contentType)
}
