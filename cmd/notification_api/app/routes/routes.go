package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jsndz/signalbus/cmd/notification_api/app/internal/handler"
	"github.com/jsndz/signalbus/pkg/kafka"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Notifications(router *gin.RouterGroup,p *kafka.Producer,db *gorm.DB,log *zap.Logger ){
	router.GET("/notify",handler.Notify(p,db,log))
}


func Tenants(r *gin.RouterGroup,db *gorm.DB,log *zap.Logger ){
	tenantHandler :=handler.NewTenantHandler(db)
	r.POST("/tenants", tenantHandler.CreateTenant)
	r.GET("/tenants", tenantHandler.ListTenants)
	r.GET("/tenants/:id", tenantHandler.GetTenant)
	r.DELETE("/tenants/:id", tenantHandler.DeleteTenant)
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