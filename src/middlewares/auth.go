package middlewares

import (
	"lexi/books/utils"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

func Authenticate(keys utils.Keys) gin.HandlerFunc {
	return func(context *gin.Context) {
		token := context.Request.Header.Get("Authorization")

		if token == "" {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorized"})
			return
		}

		if err := keys.VerifyToken(token); err != nil {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorized"})
			return
		}

		userId, roles, err := keys.ExtractClaims(token)
		if err != nil {
			context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Not authorized"})
			return
		}

		context.Set("userId", userId)
		context.Set("roles", roles)
		context.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(context *gin.Context) {
		roles, _ := context.Get("roles")

		userRoles, ok := roles.([]string)
		if !ok || !slices.Contains(userRoles, role) {
			context.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
			return
		}

		context.Next()
	}
}
