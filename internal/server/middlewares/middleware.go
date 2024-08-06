package token

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenHeader := c.Request().Header.Get("Authorization")
			if tokenHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Token is required")
			}

			tokenValue := tokenHeader[7:]
			token, _, err := new(jwt.Parser).ParseUnverified(tokenValue, jwt.MapClaims{})
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}

			expiration := claims["exp"].(float64)
			if expiration < float64(time.Now().Unix()) {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}

			userId, ok := claims["sub"].(string)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}

			c.Set("userId", userId)

			return next(c)
		}
	}
}
