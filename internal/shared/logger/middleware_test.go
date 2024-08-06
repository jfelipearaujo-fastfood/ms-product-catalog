package logger

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	t.Run("Should return middleware", func(t *testing.T) {
		// Arrange

		// Act
		middleware := Middleware()

		// Assert
		assert.NotNil(t, middleware)
	})

	t.Run("Should return middleware with request logger", func(t *testing.T) {
		// Arrange
		middleware := Middleware()

		trigger := func(_ echo.Context) error {
			return echo.NewHTTPError(http.StatusBadRequest, errors.New("this is a test"))
		}

		req := httptest.NewRequest(echo.GET, "/", nil)
		resp := httptest.NewRecorder()

		e := echo.New()
		e.Use(middleware)

		ctx := e.NewContext(req, resp)

		// Act
		err := trigger(ctx)

		// Assert
		assert.Error(t, err)
	})
}
