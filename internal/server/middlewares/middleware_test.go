package token_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	token "github.com/jfelipearaujo-org/ms-product-catalog/internal/server/middlewares"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func generateToken(t *testing.T, userId string, expire time.Duration) string {
	claims := jwt.MapClaims{
		"sub": userId,
		"exp": time.Now().Add(expire).Unix(),
	}

	if userId == "no-claims" {
		claims = jwt.MapClaims{
			"exp": time.Now().Add(expire).Unix(),
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte("my-secret"))
	assert.NoError(t, err)

	return fmt.Sprintf("Bearer %s", tokenString)
}

func TestMiddleware(t *testing.T) {
	t.Run("Should authorize when token is valid", func(t *testing.T) {
		// Arrange
		userId := uuid.NewString()

		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Authorization", generateToken(t, userId, time.Minute*1))
		res := httptest.NewRecorder()

		e := echo.New()
		e.Use(token.Middleware())
		e.GET("/", func(c echo.Context) error {
			userId := c.Get("userId").(string)
			return c.String(http.StatusOK, userId)
		})

		// Act
		e.ServeHTTP(res, req)

		// Assert
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, userId, res.Body.String())
	})

	t.Run("Should not authorize when token is invalid", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Authorization", "invalid-token")
		res := httptest.NewRecorder()

		e := echo.New()
		e.Use(token.Middleware())
		e.GET("/", func(c echo.Context) error {
			userId := c.Get("userId").(string)
			return c.String(http.StatusOK, userId)
		})

		// Act
		e.ServeHTTP(res, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("Should not authorize when token is missing", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(echo.GET, "/", nil)
		res := httptest.NewRecorder()

		e := echo.New()
		e.Use(token.Middleware())
		e.GET("/", func(c echo.Context) error {
			userId := c.Get("userId").(string)
			return c.String(http.StatusOK, userId)
		})

		// Act
		e.ServeHTTP(res, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("Should not authorize when token is expired", func(t *testing.T) {
		// Arrange
		userId := uuid.NewString()

		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Authorization", generateToken(t, userId, time.Second*-1))
		res := httptest.NewRecorder()

		e := echo.New()
		e.Use(token.Middleware())
		e.GET("/", func(c echo.Context) error {
			userId := c.Get("userId").(string)
			return c.String(http.StatusOK, userId)
		})

		// Act
		e.ServeHTTP(res, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	t.Run("Should not authorize when token doest not have user id", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Authorization", generateToken(t, "no-claims", time.Minute*1))
		res := httptest.NewRecorder()

		e := echo.New()
		e.Use(token.Middleware())
		e.GET("/", func(c echo.Context) error {
			userId := c.Get("userId").(string)
			return c.String(http.StatusOK, userId)
		})

		// Act
		e.ServeHTTP(res, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})
}
