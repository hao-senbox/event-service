package event

import (
	"event-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, handler *EventHandler) {
	eventGroup := r.Group("api/v1/events", middleware.Secured())
	{
		eventGroup.POST("", handler.CreateEvent)
		eventGroup.GET("", handler.GetAllEvents)
		eventGroup.POST("/trigger", handler.SendEventNotifications)
	}
}