package handler

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"

	"office-uninterruptible_power_supply/internal/app/ds"
)

const jwtPrefix = "Bearer "

func (h *Handler) WithAuthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtStr := c.GetHeader("Authorization")
		if !strings.HasPrefix(jwtStr, jwtPrefix) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		jwtStr = jwtStr[len(jwtPrefix):]

		err := h.redis.CheckJWTInBlacklist(c.Request.Context(), jwtStr)
		if err == nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		if !errors.Is(err, redis.Nil) {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		secret := os.Getenv("JWT_SECRET")
		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		claims, ok := token.Claims.(*ds.JWTClaims)
		if !ok {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}

func currentUserFromContext(c *gin.Context) uint {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	id, _ := userID.(uint)
	return id
}