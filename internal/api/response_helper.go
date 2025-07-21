package api

import (
	"event-service/internal/models"
	"github.com/gin-gonic/gin"
)

func SendSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, models.APIResponse{
		StatusCode: statusCode,
		Message: message,
		Data: data,
	})
}

func SendError( c* gin.Context, statusCode int, err error, errorCode string) {
	c.JSON(statusCode, models.APIResponse {
		StatusCode: statusCode,
		Error: err.Error(),
		ErrorCode: errorCode,
	})
}