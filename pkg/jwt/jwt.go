package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	jwt.RegisteredClaims
}

type TokenService struct {
	secretKey  string
	expiration time.Duration
}

func NewTokenService(secretKey string, expirationHours int) *TokenService {
	return &TokenService{
		secretKey:  secretKey,
		expiration: time.Duration(expirationHours) * time.Hour,
	}
}

func (ts *TokenService) GenerateToken(userID, email string, roles []string, companyCode string) (string, error) {
	claims := &Claims{
		UserID:      userID,
		Email:       email,
		Roles:       roles,
		CompanyCode: companyCode,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ts.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(ts.secretKey))
}

func (ts *TokenService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(ts.secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
