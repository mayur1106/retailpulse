package domain

import (
	"time"

	"github.com/google/uuid"
)

type RoleName string

const (
	RoleAdmin  RoleName = "admin"
	RoleOwner  RoleName = "organization_owner"
	RoleSeller RoleName = "seller"
	RoleMgr    RoleName = "manager"
	RoleView   RoleName = "viewer"
)

type Organization struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Role struct {
	ID   uuid.UUID
	Name RoleName
}

type User struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	RoleID         uuid.UUID
	Email          string
	Name           string
	PasswordHash   string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AuthSession struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	RefreshHash string
	UserAgent   string
	IPAddress   string
	ExpiresAt   time.Time
	RevokedAt   *time.Time
	CreatedAt   time.Time
	LastUsedAt  time.Time
}

type Principal struct {
	UserID         uuid.UUID `json:"userId"`
	OrganizationID uuid.UUID `json:"organizationId"`
	Role           RoleName  `json:"role"`
	Email          string    `json:"email"`
}

type TokenPair struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type AuthResponse struct {
	User   Principal `json:"user"`
	Tokens TokenPair `json:"tokens"`
}
