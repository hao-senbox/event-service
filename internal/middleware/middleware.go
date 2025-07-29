package middleware

import (
	"event-service/pkg/constants"
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Secured() gin.HandlerFunc {
	return func(context *gin.Context) {
		authorizationHeader := context.GetHeader("Authorization")

		if len(authorizationHeader) == 0 {
			context.AbortWithStatus(http.StatusForbidden)
			return
		}
		
		if !strings.HasPrefix(authorizationHeader, "Bearer ") {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenString := strings.Split(authorizationHeader, " ")[1]

		token, _, _ := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userId, ok := claims[constants.UserID].(string); ok {
				context.Set(constants.UserID, userId)
			}
		}

		context.Set(constants.Token, tokenString)
		context.Next()
	}
}
