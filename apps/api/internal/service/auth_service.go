package service

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"retailpulse/apps/api/internal/domain"
	"retailpulse/apps/api/internal/platform/security"
	"retailpulse/apps/api/internal/repository"
)

type AuthService struct {
	authRepo        repository.AuthRepository
	auditRepo       repository.AuditRepository
	passwords       security.PasswordHasher
	tokens          security.TokenManager
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type RegisterInput struct {
	OrganizationName string
	Name             string
	Email            string
	Password         string
	Role             domain.RoleName
	IPAddress        string
	UserAgent        string
}

type LoginInput struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}

func NewAuthService(authRepo repository.AuthRepository, auditRepo repository.AuditRepository, passwords security.PasswordHasher, tokens security.TokenManager, accessTokenTTL time.Duration, refreshTokenTTL time.Duration) *AuthService {
	return &AuthService{
		authRepo:        authRepo,
		auditRepo:       auditRepo,
		passwords:       passwords,
		tokens:          tokens,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (domain.AuthResponse, error) {
	passwordHash, err := s.passwords.Hash(input.Password)
	if err != nil {
		return domain.AuthResponse{}, err
	}
	role := input.Role
	if role == "" {
		role = domain.RoleOwner
	}
	principal, err := s.authRepo.CreateOrganizationWithUser(ctx, input.OrganizationName, slugify(input.OrganizationName), input.Name, strings.ToLower(input.Email), passwordHash, role)
	if err != nil {
		return domain.AuthResponse{}, err
	}
	tokens, err := s.issueSession(ctx, principal, input.IPAddress, input.UserAgent)
	if err != nil {
		return domain.AuthResponse{}, err
	}
	_ = s.auditRepo.Record(ctx, principal.OrganizationID, principal.UserID, "auth.register", "user", principal.UserID.String(), input.IPAddress, input.UserAgent)
	return domain.AuthResponse{User: principal, Tokens: tokens}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (domain.AuthResponse, error) {
	user, role, err := s.authRepo.FindUserByEmail(ctx, input.Email)
	if err != nil {
		return domain.AuthResponse{}, err
	}
	if user.Status != "active" || !s.passwords.Compare(user.PasswordHash, input.Password) {
		return domain.AuthResponse{}, domain.ErrInvalidCredential
	}
	principal := domain.Principal{
		UserID:         user.ID,
		OrganizationID: user.OrganizationID,
		Role:           role,
		Email:          user.Email,
	}
	tokens, err := s.issueSession(ctx, principal, input.IPAddress, input.UserAgent)
	if err != nil {
		return domain.AuthResponse{}, err
	}
	_ = s.auditRepo.Record(ctx, principal.OrganizationID, principal.UserID, "auth.login", "user", principal.UserID.String(), input.IPAddress, input.UserAgent)
	return domain.AuthResponse{User: principal, Tokens: tokens}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string, ipAddress string, userAgent string) (domain.TokenPair, error) {
	oldHash := s.tokens.HashRefreshToken(refreshToken)
	session, err := s.authRepo.FindSessionByRefreshHash(ctx, oldHash)
	if err != nil {
		return domain.TokenPair{}, err
	}
	if session.RevokedAt != nil || time.Now().UTC().After(session.ExpiresAt) {
		return domain.TokenPair{}, domain.ErrInvalidToken
	}
	principal, err := s.authRepo.FindPrincipalByUserID(ctx, session.UserID)
	if err != nil {
		return domain.TokenPair{}, err
	}
	accessToken, accessExpiresAt, err := s.tokens.CreateAccessToken(principal, s.accessTokenTTL)
	if err != nil {
		return domain.TokenPair{}, err
	}
	newRefreshToken, newRefreshHash, err := s.tokens.CreateRefreshToken()
	if err != nil {
		return domain.TokenPair{}, err
	}
	if err := s.authRepo.RotateSession(ctx, session.ID, oldHash, newRefreshHash, time.Now().UTC().Add(s.refreshTokenTTL)); err != nil {
		return domain.TokenPair{}, err
	}
	_ = s.auditRepo.Record(ctx, principal.OrganizationID, principal.UserID, "auth.refresh", "auth_session", session.ID.String(), ipAddress, userAgent)
	return domain.TokenPair{AccessToken: accessToken, RefreshToken: newRefreshToken, ExpiresAt: accessExpiresAt}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return domain.ErrInvalidToken
	}
	return s.authRepo.RevokeSession(ctx, s.tokens.HashRefreshToken(refreshToken))
}

func (s *AuthService) issueSession(ctx context.Context, principal domain.Principal, ipAddress string, userAgent string) (domain.TokenPair, error) {
	accessToken, accessExpiresAt, err := s.tokens.CreateAccessToken(principal, s.accessTokenTTL)
	if err != nil {
		return domain.TokenPair{}, err
	}
	refreshToken, refreshHash, err := s.tokens.CreateRefreshToken()
	if err != nil {
		return domain.TokenPair{}, err
	}
	session := domain.AuthSession{
		ID:          uuid.New(),
		UserID:      principal.UserID,
		RefreshHash: refreshHash,
		UserAgent:   userAgent,
		IPAddress:   ipAddress,
		ExpiresAt:   time.Now().UTC().Add(s.refreshTokenTTL),
	}
	if err := s.authRepo.CreateSession(ctx, session); err != nil {
		return domain.TokenPair{}, err
	}
	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiresAt,
	}, nil
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	value = re.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "organization"
	}
	return value
}
