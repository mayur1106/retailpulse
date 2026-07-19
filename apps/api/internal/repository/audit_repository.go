package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository interface {
	Record(ctx context.Context, organizationID uuid.UUID, actorID uuid.UUID, action string, resourceType string, resourceID string, ipAddress string, userAgent string) error
}

type PostgresAuditRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAuditRepository(pool *pgxpool.Pool) *PostgresAuditRepository {
	return &PostgresAuditRepository{pool: pool}
}

func (r *PostgresAuditRepository) Record(ctx context.Context, organizationID uuid.UUID, actorID uuid.UUID, action string, resourceType string, resourceID string, ipAddress string, userAgent string) error {
	_, err := r.pool.Exec(ctx, `
		insert into audit_logs (organization_id, actor_user_id, action, resource_type, resource_id, ip_address, user_agent)
		values ($1, $2, $3, $4, $5, $6, $7)
	`, organizationID, actorID, action, resourceType, resourceID, ipAddress, userAgent)
	return err
}
