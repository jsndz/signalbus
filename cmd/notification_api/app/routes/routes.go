package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/cmd/notification_api/app/internal/handler"
	"github.com/jsndz/signalbus/pkg/kafka"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Notifications(router *gin.RouterGroup,p *kafka.Producer,tdb *gorm.DB,ndb *gorm.DB,log *zap.Logger ){
	notificationHandler :=handler.NewNotificationHandler(ndb)

	router.POST("/",notificationHandler.Notify(p,tdb,ndb,log))
}

func Tenants(r *gin.RouterGroup,db *gorm.DB,log *zap.Logger ){
	tenantHandler :=handler.NewTenantHandler(db)
	r.POST("/", tenantHandler.CreateTenant)
	r.GET("/", tenantHandler.ListTenants)
	r.GET("/:id", tenantHandler.GetTenant)
	r.DELETE("/:id", tenantHandler.DeleteTenant)
	r.POST("/policies", tenantHandler.CreatePolicy)
}

func Templates(r *gin.RouterGroup,db *gorm.DB,log *zap.Logger ){
	templateHandler :=handler.NewTemplateHandler(db)
	r.POST("/", templateHandler.CreateTemplate)
	r.GET("/:id", templateHandler.GetTemplateByID)
	r.GET("/", templateHandler.ListTemplates)
	r.PUT("/:id", templateHandler.UpdateTemplate)
	r.DELETE("/:id", templateHandler.DeleteTemplate)
}

func APIKeys(r *gin.RouterGroup, db *gorm.DB, log *zap.Logger) {
	apiKeyHandler := handler.NewAPIKeyHandler(db)

	r.POST("/", apiKeyHandler.CreateAPIKey)
	r.GET("/", apiKeyHandler.ListAPIKeys)
	r.DELETE("/:id", apiKeyHandler.DeleteAPIKey)
}


func Policies(r *gin.RouterGroup, db *gorm.DB, log *zap.Logger) {
	policyHandler := handler.NewPolicyHandler(db)

	r.POST("/", policyHandler.CreatePolicy)
	r.GET("/", policyHandler.ListPolicies)
	r.DELETE("/:id", policyHandler.DeletePolicy)
}
