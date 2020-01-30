package common

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionEnvFromRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "ws://anycable.dev:9292/cable?token=secretos", nil)
	req.Header.Set("Cookies", "cookie=nookie;")
	req.Header.Set("Origin", "anycable.dev")
	req.Header.Set("X-Request-Id", "20200130")
	req.Header.Set("X-Api-Token", "42")
	req.Header.Set("Accept-Language", "ru")

	env := SessionEnvFromRequest(req)

	assert.Equal(t, "ws://anycable.dev:9292/cable?token=secretos", env.URL)
	assert.Equal(t, "/cable", env.Path)
	assert.Equal(t, "token=secretos", env.Query)
	assert.Equal(t, "anycable.dev", env.Host)
	assert.Equal(t, "9292", env.Port)
	assert.Equal(t, "ws", env.Scheme)
	assert.Equal(t, "anycable.dev", env.Origin)
	assert.Equal(t, "cookie=nookie;", env.Cookies)
	assert.Equal(t, "192.0.2.1", env.RemoteAddr)
}
