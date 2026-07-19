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

type AmazonRepository interface {
	CreateOAuthState(ctx context.Context, state string, principal domain.Principal, region string, marketplaceID string, expiresAt time.Time) error
	ConsumeOAuthState(ctx context.Context, state string) (domain.Principal, string, string, error)
	UpsertStore(ctx context.Context, organizationID uuid.UUID, name string, sellerID string, region string, environment string, encryptedRefreshToken []byte) (domain.AmazonStore, error)
	ListStores(ctx context.Context, organizationID uuid.UUID) ([]domain.AmazonStore, error)
	GetStore(ctx context.Context, organizationID uuid.UUID, storeID uuid.UUID) (domain.AmazonStore, []byte, error)
	FindMarketplaceID(ctx context.Context, amazonMarketplaceID string) (uuid.UUID, error)
	UpsertOrders(ctx context.Context, organizationID uuid.UUID, storeID uuid.UUID, orders []ImportedOrder) (int, error)
	UpsertInventory(ctx context.Context, organizationID uuid.UUID, storeID uuid.UUID, items []ImportedInventory) (int, error)
	UpsertFinancialEvents(ctx context.Context, organizationID uuid.UUID, storeID uuid.UUID, events []ImportedFinancialEvent) (int, error)
	UpsertReportRecords(ctx context.Context, organizationID uuid.UUID, storeID uuid.UUID, records []ImportedReport) (int, error)
	UpsertCampaigns(ctx context.Context, organizationID uuid.UUID, storeID uuid.UUID, campaigns []ImportedCampaign) (int, error)
	RecordSync(ctx context.Context, organizationID uuid.UUID, storeID uuid.UUID, syncType string, status string, processed int, errMsg string, startedAt time.Time, finishedAt time.Time) error
}

type ImportedOrder struct {
	AmazonOrderID       string
	AmazonMarketplaceID string
	OrderStatus         string
	PurchaseDate        time.Time
	OrderTotal          float64
	CurrencyCode        string
	Items               []ImportedOrderItem
}

type ImportedOrderItem struct {
	AmazonOrderItemID string
	ASIN              string
	SKU               string
	Title             string
	Quantity          int
	ItemPrice         float64
	CurrencyCode      string
}

type ImportedInventory struct {
	ASIN        string
	SKU         string
	Title       string
	Fulfillable int
	Inbound     int
	Reserved    int
}
type ImportedFinancialEvent struct {
	EventType  string
	Amount     float64
	Currency   string
	PostedAt   time.Time
	RawPayload []byte
}
type ImportedReport struct {
	ReportID   string
	ReportType string
	Status     string
	CreatedAt  time.Time
}

type ImportedCampaign struct {
	CampaignID   string
	Channel      string
	Name         string
	CampaignType string
	Status       string
	Budget       float64
}

type PostgresAmazonRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAmazonRepository(pool *pgxpool.Pool) *PostgresAmazonRepository {
	return &PostgresAmazonRepository{pool: pool}
}

func (r *PostgresAmazonRepository) CreateOAuthState(ctx context.Context, state string, principal domain.Principal, region string, marketplaceID string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		insert into amazon_oauth_states (state, organization_id, user_id, region, marketplace_id, expires_at)
		values ($1, $2, $3, $4, $5, $6)
	`, state, principal.OrganizationID, principal.UserID, region, marketplaceID, expiresAt)
	return err
}

func (r *PostgresAmazonRepository) ConsumeOAuthState(ctx context.Context, state string) (domain.Principal, string, string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Principal{}, "", "", err
	}
	defer tx.Rollback(ctx)

	var principal domain.Principal
	var region string
	var marketplaceID string
	err = tx.QueryRow(ctx, `
		select s.organization_id, s.user_id, u.email, ro.name, s.region, s.marketplace_id
		from amazon_oauth_states s
		join users u on u.id = s.user_id
		join roles ro on ro.id = u.role_id
		where s.state = $1 and s.used_at is null and s.expires_at > now()
		for update
	`, state).Scan(&principal.OrganizationID, &principal.UserID, &principal.Email, &principal.Role, &region, &marketplaceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Principal{}, "", "", domain.ErrInvalidToken
	}
	if err != nil {
		return domain.Principal{}, "", "", err
	}
	if _, err := tx.Exec(ctx, `update amazon_oauth_states set used_at = now() where state = $1`, state); err != nil {
		return domain.Principal{}, "", "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Principal{}, "", "", err
	}
	return principal, region, marketplaceID, nil
}

func (r *PostgresAmazonRepository) UpsertStore(ctx context.Context, organizationID uuid.UUID, name string, sellerID string, region string, environment string, encryptedRefreshToken []byte) (domain.AmazonStore, error) {
	var store domain.AmazonStore
	err := r.pool.QueryRow(ctx, `
		insert into stores (organization_id, name, seller_id, region, environment, encrypted_refresh_token, status)
		values ($1, $2, $3, $4, $5, $6, 'connected')
		on conflict (organization_id, seller_id, environment) do update
		set name = excluded.name,
			region = excluded.region,
			environment = excluded.environment,
			encrypted_refresh_token = excluded.encrypted_refresh_token,
			status = 'connected',
			updated_at = now()
		returning id, organization_id, name, seller_id, region, environment, status, last_imported_at, created_at, updated_at
	`, organizationID, name, sellerID, region, environment, encryptedRefreshToken).Scan(&store.ID, &store.OrganizationID, &store.Name, &store.SellerID, &store.Region, &store.Environment, &store.Status, &store.LastImportedAt, &store.CreatedAt, &store.UpdatedAt)
	return store, err
}

func (r *PostgresAmazonRepository) ListStores(ctx context.Context, organizationID uuid.UUID) ([]domain.AmazonStore, error) {
	rows, err := r.pool.Query(ctx, `
		select id, organization_id, name, seller_id, region, environment, status, last_imported_at, created_at, updated_at
		from stores
		where organization_id = $1
		order by created_at desc
	`, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stores []domain.AmazonStore
	for rows.Next() {
		var store domain.AmazonStore
		if err := rows.Scan(&store.ID, &store.OrganizationID, &store.Name, &store.SellerID, &store.Region, &store.Environment, &store.Status, &store.LastImportedAt, &store.CreatedAt, &store.UpdatedAt); err != nil {
			return nil, err
		}
		stores = append(stores, store)
	}
	return stores, rows.Err()
}

func (r *PostgresAmazonRepository) GetStore(ctx context.Context, organizationID uuid.UUID, storeID uuid.UUID) (domain.AmazonStore, []byte, error) {
	var store domain.AmazonStore
	var encrypted []byte
	err := r.pool.QueryRow(ctx, `
		select id, organization_id, name, seller_id, region, environment, status, last_imported_at, created_at, updated_at, encrypted_refresh_token
		from stores
		where organization_id = $1 and id = $2
	`, organizationID, storeID).Scan(&store.ID, &store.OrganizationID, &store.Name, &store.SellerID, &store.Region, &store.Environment, &store.Status, &store.LastImportedAt, &store.CreatedAt, &store.UpdatedAt, &encrypted)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AmazonStore{}, nil, domain.ErrNotFound
	}
	return store, encrypted, err
}

func (r *PostgresAmazonRepository) FindMarketplaceID(ctx context.Context, amazonMarketplaceID string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `select id from marketplaces where amazon_marketplace_id = $1`, amazonMarketplaceID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, domain.ErrNotFound
	}
	return id, err
}

func (r *PostgresAmazonRepository) UpsertOrders(ctx context.Context, organizationID uuid.UUID, storeID uuid.UUID, orders []ImportedOrder) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	imported := 0
	for _, order := range orders {
		marketplaceID, err := r.FindMarketplaceID(ctx, order.AmazonMarketplaceID)
		if err != nil {
			marketplaceID = uuid.Nil
		}
		var nullableMarketplace any
		if marketplaceID != uuid.Nil {
			nullableMarketplace = marketplaceID
		}
		var orderID uuid.UUID
		err = tx.QueryRow(ctx, `
			insert into orders (organization_id, store_id, amazon_order_id, marketplace_id, order_status, purchase_date, order_total, currency_code, data_origin)
			values ($1, $2, $3, $4, $5, $6, $7, $8, 'amazon')
			on conflict (organization_id, store_id, amazon_order_id) do update
			set marketplace_id = excluded.marketplace_id,
				order_status = excluded.order_status,
				purchase_date = excluded.purchase_date,
				order_total = excluded.order_total,
				currency_code = excluded.currency_code,
				data_origin = 'amazon',
				updated_at = now()
			returning id
		`, organizationID, storeID, order.AmazonOrderID, nullableMarketplace, order.OrderStatus, order.PurchaseDate, order.OrderTotal, order.CurrencyCode).Scan(&orderID)
		if err != nil {
			return 0, err
		}
		for _, item := range order.Items {
			var productID uuid.UUID
			err = tx.QueryRow(ctx, `insert into products(organization_id,store_id,asin,sku,title,status,data_origin) values($1,$2,$3,nullif($4,''),$5,'active','amazon') on conflict(organization_id,store_id,asin,coalesce(sku,'')) do update set title=excluded.title,data_origin='amazon',updated_at=now() returning id`, organizationID, storeID, item.ASIN, item.SKU, item.Title).Scan(&productID)
			if err != nil {
				return 0, err
			}
			_, err = tx.Exec(ctx, `insert into order_items(order_id,product_id,amazon_order_item_id,asin,sku,title,quantity_ordered,item_price,currency_code) values($1,$2,$3,$4,nullif($5,''),$6,$7,$8,$9) on conflict(order_id,amazon_order_item_id) do update set product_id=excluded.product_id,quantity_ordered=excluded.quantity_ordered,item_price=excluded.item_price,title=excluded.title`, orderID, productID, item.AmazonOrderItemID, item.ASIN, item.SKU, item.Title, item.Quantity, item.ItemPrice, item.CurrencyCode)
			if err != nil {
				return 0, err
			}
		}
		imported++
	}
	if _, err := tx.Exec(ctx, `update stores set last_imported_at = now(), updated_at = now() where id = $1`, storeID); err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return imported, nil
}

func (r *PostgresAmazonRepository) RecordSync(ctx context.Context, organizationID uuid.UUID, storeID uuid.UUID, syncType string, status string, processed int, errMsg string, startedAt time.Time, finishedAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		insert into sync_history (organization_id, store_id, sync_type, status, started_at, finished_at, records_processed, error_message)
		values ($1, $2, $3, $4, $5, $6, $7, nullif($8, ''))
	`, organizationID, storeID, syncType, status, startedAt, finishedAt, processed, errMsg)
	return err
}
