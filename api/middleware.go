package api

import (
	"bank/token"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(authorizationHeader, "Bearer ") {
			err := &ApiError{Status: http.StatusUnauthorized, Err: "authorization header must start with Bearer"}
			ctx.AbortWithStatusJSON(err.Status, err)
			return
		}

		tokenString := strings.TrimPrefix(authorizationHeader, "Bearer ")
		payload, err := tokenMaker.VerifyToken(tokenString)
		if err != nil {
			err := &ApiError{Status: http.StatusUnauthorized, Err: "invalid token"}
			ctx.AbortWithStatusJSON(err.Status, err)
			return
		}

		ctx.Set("auth_payload", payload)
		ctx.Next()
	}
}