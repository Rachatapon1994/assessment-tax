package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	t.Parallel()
	os.Setenv("ADMIN_USERNAME", "admin")
	os.Setenv("ADMIN_PASSWORD", "secret")

	tests := []struct {
		auth           string
		wantStatusCode int
	}{
		{"admin:secret", http.StatusOK},
		{"admin:wrong-secret", http.StatusUnauthorized},
	}

	for _, tc := range tests {
		e := echo.New()
		mw := Authenticate()
		e.Use(middleware.BasicAuth(mw))
		e.GET("/admin", func(c echo.Context) error { return c.String(http.StatusOK, "[]") })
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		auth := "basic " + base64.StdEncoding.EncodeToString([]byte(tc.auth))
		req.Header.Set(echo.HeaderAuthorization, auth)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, tc.wantStatusCode, rec.Code)
	}
}
