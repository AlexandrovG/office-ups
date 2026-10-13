package handler

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"office-uninterruptible_power_supply/internal/app/ds"
)

// RegisterAPI godoc
// @Summary      Регистрация
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  ds.RegisterRequest  true  "login и password"
// @Success      201  {object}  ds.UserSerializer
// @Failure      400
// @Failure      409
// @Router       /auth/register [post]
func (h *Handler) RegisterAPI(c *gin.Context) {
	var req ds.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	user, err := h.repo.CreateUser(req.Login, string(hash))
	if err != nil {
		c.Status(http.StatusConflict)
		return
	}

	c.JSON(http.StatusCreated, ds.UserSerializer{
		ID:          user.ID,
		Login:       user.UserLogin,
		IsModerator: user.IsModerator,
	})
}

// LoginAPI godoc
// @Summary      Аутентификация
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  ds.RegisterRequest  true  "login и password"
// @Success      200  {object}  object
// @Failure      400
// @Failure      403
// @Router       /auth/login [post]
func (h *Handler) LoginAPI(c *gin.Context) {
	var req ds.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUserByLogin(req.Login)
	if err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.Status(http.StatusForbidden)
		return
	}

	secret := os.Getenv("JWT_SECRET")
	expiresIn := 24 * time.Hour
	if d, err := time.ParseDuration(os.Getenv("JWT_EXPIRES_IN")); err == nil {
		expiresIn = d
	}

	claims := ds.JWTClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "office-uninterruptible-power-supply",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	strToken, err := token.SignedString([]byte(secret))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"expires_in":   int(expiresIn.Seconds()),
		"access_token": strToken,
		"token_type":   "Bearer",
	})
}

// LogoutAPI godoc
// @Summary      Выход
// @Tags         auth
// @Security     BearerAuth
// @Success      200
// @Failure      400
// @Router       /auth/logout [post]
func (h *Handler) LogoutAPI(c *gin.Context) {
	jwtStr := c.GetHeader("Authorization")
	if len(jwtStr) < 7 || jwtStr[:7] != "Bearer " {
		c.Status(http.StatusBadRequest)
		return
	}
	jwtStr = jwtStr[7:]

	expiresIn := 24 * time.Hour
	if d, err := time.ParseDuration(os.Getenv("JWT_EXPIRES_IN")); err == nil {
		expiresIn = d
	}

	if err := h.redis.WriteJWTToBlacklist(c.Request.Context(), jwtStr, expiresIn); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}