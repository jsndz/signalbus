package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jsndz/signalbus/cmd/notification_api/app/internal/services"
	"gorm.io/gorm"
)

type APIKeyHandler struct {
	service *services.APIKeyService
}

func NewAPIKeyHandler(db *gorm.DB) *APIKeyHandler {
	return &APIKeyHandler{service: services.NewAPIKeyService(db)}
}

func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	var req struct {
		TenantID string   `json:"tenant_id" binding:"required"`
		Hash     string   `json:"hash" binding:"required"`
		Scopes   []string `json:"scopes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID"})
		return
	}

	apiKey, err := h.service.CreateAPIKey(tenantID, req.Hash, req.Scopes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, apiKey)
}

func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	tenantIDParam := c.Query("tenant_id")
	tenantID, err := uuid.Parse(tenantIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID"})
		return
	}

	apiKeys, err := h.service.ListAPIKeys(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, apiKeys)
}

func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid APIKey ID"})
		return
	}

	if err := h.service.DeleteAPIKey(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
