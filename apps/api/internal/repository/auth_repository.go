package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"retailpulse/apps/api/internal/domain"
)

type AuthRepository interface {
	CreateOrganizationWithUser(ctx context.Context, orgName string, slug string, userName string, email string, passwordHash string, role domain.RoleName) (domain.Principal, error)
	FindUserByEmail(ctx context.Context, email string) (domain.User, domain.RoleName, error)
	FindPrincipalByUserID(ctx context.Context, userID uuid.UUID) (domain.Principal, error)
	CreateSession(ctx context.Context, session domain.AuthSession) error
	FindSessionByRefreshHash(ctx context.Context, refreshHash string) (domain.AuthSession, error)
	RotateSession(ctx context.Context, sessionID uuid.UUID, oldHash string, newHash string, expiresAt time.Time) error
	RevokeSession(ctx context.Context, refreshHash string) error
}

type PostgresAuthRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAuthRepository(pool *pgxpool.Pool) *PostgresAuthRepository {
	return &PostgresAuthRepository{pool: pool}
}

func (r *PostgresAuthRepository) CreateOrganizationWithUser(ctx context.Context, orgName string, slug string, userName string, email string, passwordHash string, role domain.RoleName) (domain.Principal, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Principal{}, err
	}
	defer tx.Rollback(ctx)

	var orgID uuid.UUID
	err = tx.QueryRow(ctx, `
		insert into organizations (name, slug)
		values ($1, $2)
		returning id
	`, orgName, slug).Scan(&orgID)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Principal{}, domain.ErrConflict
		}
		return domain.Principal{}, err
	}

	var roleID uuid.UUID
	err = tx.QueryRow(ctx, `select id from roles where name = $1`, role).Scan(&roleID)
	if err != nil {
		return domain.Principal{}, err
	}

	var userID uuid.UUID
	err = tx.QueryRow(ctx, `
		insert into users (organization_id, role_id, email, name, password_hash)
		values ($1, $2, $3, $4, $5)
		returning id
	`, orgID, roleID, email, userName, passwordHash).Scan(&userID)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Principal{}, domain.ErrConflict
		}
		return domain.Principal{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Principal{}, err
	}

	return domain.Principal{
		UserID:         userID,
		OrganizationID: orgID,
		Role:           role,
		Email:          email,
	}, nil
}

func (r *PostgresAuthRepository) FindUserByEmail(ctx context.Context, email string) (domain.User, domain.RoleName, error) {
	var user domain.User
	var role domain.RoleName
	err := r.pool.QueryRow(ctx, `
		select u.id, u.organization_id, u.role_id, u.email, u.name, u.password_hash, u.status, u.created_at, u.updated_at, r.name
		from users u
		join roles r on r.id = u.role_id
		where lower(u.email) = lower($1)
	`, email).Scan(&user.ID, &user.OrganizationID, &user.RoleID, &user.Email, &user.Name, &user.PasswordHash, &user.Status, &user.CreatedAt, &user.UpdatedAt, &role)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, "", domain.ErrInvalidCredential
	}
	return user, role, err
}

func (r *PostgresAuthRepository) FindPrincipalByUserID(ctx context.Context, userID uuid.UUID) (domain.Principal, error) {
	var principal domain.Principal
	err := r.pool.QueryRow(ctx, `
		select u.id, u.organization_id, r.name, u.email
		from users u
		join roles r on r.id = u.role_id
		where u.id = $1 and u.status = 'active'
	`, userID).Scan(&principal.UserID, &principal.OrganizationID, &principal.Role, &principal.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Principal{}, domain.ErrInvalidToken
	}
	return principal, err
}

func (r *PostgresAuthRepository) CreateSession(ctx context.Context, session domain.AuthSession) error {
	_, err := r.pool.Exec(ctx, `
		insert into auth_sessions (id, user_id, refresh_hash, user_agent, ip_address, expires_at, last_used_at)
		values ($1, $2, $3, $4, $5, $6, now())
	`, session.ID, session.UserID, session.RefreshHash, session.UserAgent, session.IPAddress, session.ExpiresAt)
	return err
}

func (r *PostgresAuthRepository) FindSessionByRefreshHash(ctx context.Context, refreshHash string) (domain.AuthSession, error) {
	var session domain.AuthSession
	err := r.pool.QueryRow(ctx, `
		select id, user_id, refresh_hash, user_agent, ip_address, expires_at, revoked_at, created_at, last_used_at
		from auth_sessions
		where refresh_hash = $1
	`, refreshHash).Scan(&session.ID, &session.UserID, &session.RefreshHash, &session.UserAgent, &session.IPAddress, &session.ExpiresAt, &session.RevokedAt, &session.CreatedAt, &session.LastUsedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AuthSession{}, domain.ErrInvalidToken
	}
	return session, err
}

func (r *PostgresAuthRepository) RotateSession(ctx context.Context, sessionID uuid.UUID, oldHash string, newHash string, expiresAt time.Time) error {
	tag, err := r.pool.Exec(ctx, `
		update auth_sessions
		set refresh_hash = $1, expires_at = $2, last_used_at = now()
		where id = $3 and refresh_hash = $4 and revoked_at is null and expires_at > now()
	`, newHash, expiresAt, sessionID, oldHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInvalidToken
	}
	return nil
}

func (r *PostgresAuthRepository) RevokeSession(ctx context.Context, refreshHash string) error {
	_, err := r.pool.Exec(ctx, `
		update auth_sessions
		set revoked_at = now()
		where refresh_hash = $1 and revoked_at is null
	`, refreshHash)
	return err
}

func isUniqueViolation(err error) bool {
	return err != nil && (contains(err.Error(), "SQLSTATE 23505") || contains(err.Error(), "unique constraint"))
}

func contains(value string, fragment string) bool {
	for i := 0; i+len(fragment) <= len(value); i++ {
		if value[i:i+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
