package token

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenService_GenerateAndValidateToken(t *testing.T) {
	ts := NewTokenService("test-secret-key-123456789012", 24)

	userID := "user123"
	email := "test@example.com"
	roles := []string{"admin", "user"}
	companyCode := "COMP001"

	// Generate token
	token, err := ts.GenerateToken(userID, email, roles, companyCode)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := ts.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, roles, claims.Roles)
	assert.Equal(t, companyCode, claims.CompanyCode)
}

func TestTokenService_InvalidToken(t *testing.T) {
	ts := NewTokenService("test-secret-key-123456789012", 24)

	// Test with invalid token
	_, err := ts.ValidateToken("invalid-token")
	assert.ErrorIs(t, err, ErrInvalidToken)

	// Test with empty token
	_, err = ts.ValidateToken("")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestTokenService_ExpiredToken(t *testing.T) {
	// Create service with very short expiration (1 second)
	ts := &TokenService{
		secretKey:  NewTokenService("test-secret", 1).secretKey,
		expiration: 1 * time.Millisecond,
	}

	token, err := ts.GenerateToken("user123", "test@example.com", []string{"user"}, "COMP001")
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Validate expired token
	_, err = ts.ValidateToken(token)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestTokenService_DifferentSecretKeys(t *testing.T) {
	ts1 := NewTokenService("secret-key-1", 24)
	ts2 := NewTokenService("secret-key-2", 24)

	// Generate token with first service
	token, err := ts1.GenerateToken("user123", "test@example.com", []string{"user"}, "COMP001")
	require.NoError(t, err)

	// Try to validate with different secret key
	_, err = ts2.ValidateToken(token)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestTokenService_ShortSecretKey(t *testing.T) {
	// Test with short secret key (should be padded)
	ts := NewTokenService("short", 24)
	
	token, err := ts.GenerateToken("user123", "test@example.com", []string{"user"}, "COMP001")
	require.NoError(t, err)
	
	claims, err := ts.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user123", claims.UserID)
}

func TestTokenService_LongSecretKey(t *testing.T) {
	// Test with very long secret key (should be truncated)
	longKey := "this-is-a-very-long-secret-key-that-exceeds-32-bytes-and-should-be-truncated"
	ts := NewTokenService(longKey, 24)
	
	token, err := ts.GenerateToken("user123", "test@example.com", []string{"user"}, "COMP001")
	require.NoError(t, err)
	
	claims, err := ts.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user123", claims.UserID)
}

func TestTokenService_EmptyRoles(t *testing.T) {
	ts := NewTokenService("test-secret-key-123456789012", 24)
	
	// Generate token with empty roles
	token, err := ts.GenerateToken("user123", "test@example.com", []string{}, "COMP001")
	require.NoError(t, err)
	
	claims, err := ts.ValidateToken(token)
	require.NoError(t, err)
	assert.Empty(t, claims.Roles)
}

func TestTokenService_MultipleRoles(t *testing.T) {
	ts := NewTokenService("test-secret-key-123456789012", 24)
	
	roles := []string{"admin", "user", "moderator", "viewer"}
	token, err := ts.GenerateToken("user123", "test@example.com", roles, "COMP001")
	require.NoError(t, err)
	
	claims, err := ts.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, roles, claims.Roles)
}
