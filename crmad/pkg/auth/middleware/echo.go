package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type Auth[U any, T any] interface {
	Generate(user U) (T, error)
	Verify(accessToken string) (T, error)
}

type UserDataNext[T any] func(echo.Context, T) error

type AuthMiddleware[U any, T any] struct {
	authManager Auth[U, T]
}

func New[U any, T any](auth Auth[U, T]) *AuthMiddleware[U, T] {
	return &AuthMiddleware[U, T]{
		authManager: auth,
	}
}

func (a *AuthMiddleware[U, T]) Do(next UserDataNext[T]) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		token = strings.Replace(token, "Bearer ", "", 1)

		claims, err := a.authManager.Verify(token)
		if err != nil {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "bad creds",
			})
		}

		return next(c, claims)
	}
}
