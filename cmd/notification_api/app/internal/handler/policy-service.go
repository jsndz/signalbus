package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jsndz/signalbus/cmd/notification_api/app/internal/services"
	"gorm.io/gorm"
)
type PolicyHandler struct {
	service *services.PolicyService
}

func NewPolicyHandler(db *gorm.DB) *PolicyHandler {
	return &PolicyHandler{service: services.NewPolicyService(db)}
}

func (h *PolicyHandler) CreatePolicy(c *gin.Context) {
	var req struct {
		TenantID string   `json:"tenant_id" binding:"required"`
		Topic    string   `json:"topic" binding:"required"`
		Channels []string `json:"channels" binding:"required"`
		Locale   string   `json:"locale" binding:"required"`
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

	policy, err := h.service.CreatePolicy(tenantID, req.Topic, req.Channels, req.Locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, policy)
}

func (h *PolicyHandler) ListPolicies(c *gin.Context) {
	tenantIDParam := c.Query("tenant_id")
	tenantID, err := uuid.Parse(tenantIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID"})
		return
	}

	policies, err := h.service.ListPolicies(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, policies)
}

func (h *PolicyHandler) DeletePolicy(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid Policy ID"})
		return
	}

	if err := h.service.DeletePolicy(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
