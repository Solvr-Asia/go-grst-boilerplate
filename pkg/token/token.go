package token

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"aidanwoods.dev/go-paseto"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID      string   `json:"userId"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	CompanyCode string   `json:"companyCode"`
}

type TokenService struct {
	secretKey  paseto.V4SymmetricKey
	expiration time.Duration
}

// NewTokenService creates a new token service with PASETO v4
// secretKey should be a hex-encoded 32-byte string
func NewTokenService(secretKeyString string, expirationHours int) *TokenService {
	var key paseto.V4SymmetricKey
	
	// Try to decode from hex if it looks like hex
	if len(secretKeyString) == 64 { // 32 bytes in hex = 64 chars
		keyBytes, err := hex.DecodeString(secretKeyString)
		if err == nil && len(keyBytes) == 32 {
			// Successfully decoded hex key
			key, _ = paseto.V4SymmetricKeyFromBytes(keyBytes)
			return &TokenService{
				secretKey:  key,
				expiration: time.Duration(expirationHours) * time.Hour,
			}
		}
	}
	
	// Otherwise, derive key from the string
	keyBytes := []byte(secretKeyString)
	
	// Ensure exactly 32 bytes for PASETO v4
	var keyArray [32]byte
	if len(keyBytes) < 32 {
		// Pad with random bytes if too short (but copy what we have)
		copy(keyArray[:], keyBytes)
		// Fill rest with deterministic data based on the key
		for i := len(keyBytes); i < 32; i++ {
			keyArray[i] = byte(i ^ len(keyBytes))
		}
	} else {
		// Use first 32 bytes if too long
		copy(keyArray[:], keyBytes[:32])
	}
	
	key, _ = paseto.V4SymmetricKeyFromBytes(keyArray[:])

	return &TokenService{
		secretKey:  key,
		expiration: time.Duration(expirationHours) * time.Hour,
	}
}

// GenerateSecretKey generates a new random 32-byte secret key and returns it as hex
func GenerateSecretKey() string {
	key := paseto.NewV4SymmetricKey()
	return hex.EncodeToString(key.ExportBytes())
}

// GenerateToken creates a new PASETO token with user claims
func (ts *TokenService) GenerateToken(userID, email string, roles []string, companyCode string) (string, error) {
	token := paseto.NewToken()
	
	now := time.Now()
	
	// Set registered claims
	token.SetIssuedAt(now)
	token.SetNotBefore(now)
	token.SetExpiration(now.Add(ts.expiration))
	
	// Set custom claims
	token.SetString("userId", userID)
	token.SetString("email", email)
	token.Set("roles", roles)
	token.SetString("companyCode", companyCode)
	
	// Encrypt the token (v4.local)
	encrypted := token.V4Encrypt(ts.secretKey, nil)
	return encrypted, nil
}

// ValidateToken validates and decrypts a PASETO token
func (ts *TokenService) ValidateToken(tokenString string) (*Claims, error) {
	parser := paseto.NewParser()
	
	// Add validation rules
	parser.AddRule(paseto.NotExpired())
	parser.AddRule(paseto.ValidAt(time.Now()))
	
	// Parse and decrypt the token
	token, err := parser.ParseV4Local(ts.secretKey, tokenString, nil)
	if err != nil {
		// Check if error message contains expiration-related text
		errMsg := err.Error()
		if contains(errMsg, "expired") || contains(errMsg, "expiration") || contains(errMsg, "exp") {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	
	// Extract claims
	claims := &Claims{}
	
	// Get userId
	if err := token.Get("userId", &claims.UserID); err != nil {
		return nil, ErrInvalidToken
	}
	
	// Get email
	if err := token.Get("email", &claims.Email); err != nil {
		return nil, ErrInvalidToken
	}
	
	// Get companyCode
	if err := token.Get("companyCode", &claims.CompanyCode); err != nil {
		return nil, ErrInvalidToken
	}
	
	// Get roles array
	if err := token.Get("roles", &claims.Roles); err != nil {
		return nil, ErrInvalidToken
	}
	
	return claims, nil
}

// GetSecretKeyHex exports the secret key as hex string (for backup/migration)
func (ts *TokenService) GetSecretKeyHex() string {
	return hex.EncodeToString(ts.secretKey.ExportBytes())
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func init() {
	// Ensure crypto/rand is available
	var b [1]byte
	_, _ = rand.Read(b[:])
}
