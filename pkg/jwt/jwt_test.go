package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	ts := NewTokenService("test-secret-key", 24)

	userID := "user-123"
	email := "test@example.com"
	roles := []string{"admin", "user"}
	companyCode := "COMPANY-001"

	token, err := ts.GenerateToken(userID, email, roles, companyCode)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateToken_Success(t *testing.T) {
	ts := NewTokenService("test-secret-key", 24)

	userID := "user-123"
	email := "test@example.com"
	roles := []string{"admin", "user"}
	companyCode := "COMPANY-001"

	token, err := ts.GenerateToken(userID, email, roles, companyCode)
	assert.NoError(t, err)

	claims, err := ts.ValidateToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, roles, claims.Roles)
	assert.Equal(t, companyCode, claims.CompanyCode)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	ts := NewTokenService("test-secret-key", 24)

	claims, err := ts.ValidateToken("invalid-token")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	ts1 := NewTokenService("secret-key-1", 24)
	ts2 := NewTokenService("secret-key-2", 24)

	token, err := ts1.GenerateToken("user-123", "test@example.com", []string{"user"}, "")
	assert.NoError(t, err)

	claims, err := ts2.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	// Create token service with very short expiration
	ts := &TokenService{
		secretKey:  "test-secret-key",
		expiration: -1 * time.Hour, // Already expired
	}

	token, err := ts.GenerateToken("user-123", "test@example.com", []string{"user"}, "")
	assert.NoError(t, err)

	claims, err := ts.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrExpiredToken, err)
}

func TestTokenService_DifferentExpirations(t *testing.T) {
	tests := []struct {
		name       string
		expiration int
	}{
		{"1 hour", 1},
		{"24 hours", 24},
		{"168 hours (1 week)", 168},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTokenService("test-secret", tt.expiration)
			token, err := ts.GenerateToken("user-123", "test@example.com", []string{"user"}, "")

			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			claims, err := ts.ValidateToken(token)
			assert.NoError(t, err)
			assert.NotNil(t, claims)
		})
	}
}
