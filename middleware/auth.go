package middleware

import (
	"crypto/subtle"
	"os"

	"github.com/labstack/echo/v4"
)

func Authenticate() func(username, password string, c echo.Context) (bool, error) {
	return func(username, password string, c echo.Context) (bool, error) {
		if subtle.ConstantTimeCompare([]byte(username), []byte(os.Getenv("ADMIN_USERNAME"))) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(os.Getenv("ADMIN_PASSWORD"))) == 1 {
			return true, nil
		}
		return false, nil
	}
}
