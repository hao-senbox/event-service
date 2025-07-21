package api

import (

	"github.com/gin-gonic/gin"
)

func RegisterEmergencyRouters(r *gin.Engine, emergencyService service.EmergencyService) {
	handlers := NewEmergencyService(emergencyService)

	emergencyGroup := r.Group("/api/v1/emergency", Secured())
	{
		emergencyGroup.POST("", handlers.CreateEmergency)
	}	
}