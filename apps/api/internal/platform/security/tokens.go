package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"retailpulse/apps/api/internal/domain"
)

type TokenManager struct {
	accessSecret  []byte
	refreshSecret []byte
}

type AccessClaims struct {
	OrganizationID string          `json:"organizationId"`
	Role           domain.RoleName `json:"role"`
	Email          string          `json:"email"`
	jwt.RegisteredClaims
}

func NewTokenManager(accessSecret string, refreshSecret string) TokenManager {
	return TokenManager{accessSecret: []byte(accessSecret), refreshSecret: []byte(refreshSecret)}
}

func (m TokenManager) CreateAccessToken(principal domain.Principal, ttl time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(ttl)
	claims := AccessClaims{
		OrganizationID: principal.OrganizationID.String(),
		Role:           principal.Role,
		Email:          principal.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   principal.UserID.String(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    "retailpulse-ai",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.accessSecret)
	return signed, expiresAt, err
}

func (m TokenManager) CreateRefreshToken() (string, string, error) {
	random := make([]byte, 48)
	if _, err := rand.Read(random); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(random)
	return token, m.HashRefreshToken(token), nil
}

func (m TokenManager) HashRefreshToken(token string) string {
	sum := sha256.Sum256(append(m.refreshSecret, []byte(token)...))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func (m TokenManager) ParseAccessToken(tokenValue string) (domain.Principal, error) {
	parsed, err := jwt.ParseWithClaims(tokenValue, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.accessSecret, nil
	})
	if err != nil || !parsed.Valid {
		return domain.Principal{}, domain.ErrInvalidToken
	}
	claims, ok := parsed.Claims.(*AccessClaims)
	if !ok {
		return domain.Principal{}, domain.ErrInvalidToken
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return domain.Principal{}, domain.ErrInvalidToken
	}
	orgID, err := uuid.Parse(claims.OrganizationID)
	if err != nil {
		return domain.Principal{}, domain.ErrInvalidToken
	}
	return domain.Principal{
		UserID:         userID,
		OrganizationID: orgID,
		Role:           claims.Role,
		Email:          claims.Email,
	}, nil
}
