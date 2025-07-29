package helper

import (
	"github.com/gin-gonic/gin"
)

const (
	ErrInvalidOperation = "ERR_INVALID_OPERATION"
	ErrInvalidRequest   = "ERR_INVALID_REQUEST"
)

type APIResponse struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
	ErrorCode  string      `json:"error_code,omitempty"`
}

func SendSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		StatusCode: statusCode,
		Message: message,
		Data: data,
	})
}

func SendError( c* gin.Context, statusCode int, err error, errorCode string) {
	c.JSON(statusCode, APIResponse {
		StatusCode: statusCode,
		Error: err.Error(),
		ErrorCode: errorCode,
	})
}