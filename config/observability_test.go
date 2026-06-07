package config

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterObservabilityRoutes(t *testing.T) {
	app := fiber.New()

	registerObservabilityRoutes(app, "test_service")

	metricsResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/metrics", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, metricsResp.StatusCode)

	docsResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, docsResp.StatusCode)
}
