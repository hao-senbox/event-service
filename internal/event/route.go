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
		eventGroup.GET("/:id", handler.GetEventByID)
		eventGroup.PUT("/:id", handler.UpdateEvent)
		eventGroup.DELETE("/:id", handler.DeleteEvent)
		eventGroup.PUT("toggle-send/:id", handler.ToggleSendEventNotifications)
		eventGroup.PUT("toggle-show/:id", handler.ToggleShowEventNotifications)
		eventGroup.POST("/trigger", handler.SendEventNotifications)
	}
}
