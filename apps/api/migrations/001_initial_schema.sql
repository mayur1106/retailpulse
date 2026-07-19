create extension if not exists pgcrypto;

create table organizations (
	id uuid primary key default gen_random_uuid(),
	name text not null,
	slug text not null unique,
	status text not null default 'active',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create table roles (
	id uuid primary key default gen_random_uuid(),
	name text not null unique,
	description text not null
);

create table permissions (
	id uuid primary key default gen_random_uuid(),
	name text not null unique,
	description text not null
);

create table role_permissions (
	role_id uuid not null references roles(id) on delete cascade,
	permission_id uuid not null references permissions(id) on delete cascade,
	primary key (role_id, permission_id)
);

create table users (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	role_id uuid not null references roles(id),
	email text not null unique,
	name text not null,
	password_hash text not null,
	status text not null default 'active',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create index idx_users_organization_id on users(organization_id);

create table auth_sessions (
	id uuid primary key,
	user_id uuid not null references users(id) on delete cascade,
	refresh_hash text not null unique,
	user_agent text not null default '',
	ip_address text not null default '',
	expires_at timestamptz not null,
	revoked_at timestamptz,
	created_at timestamptz not null default now(),
	last_used_at timestamptz not null default now()
);

create index idx_auth_sessions_user_id on auth_sessions(user_id);

create table marketplaces (
	id uuid primary key default gen_random_uuid(),
	amazon_marketplace_id text not null unique,
	country_code text not null,
	name text not null,
	region text not null,
	currency_code text not null
);

create table stores (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	name text not null,
	seller_id text not null,
	encrypted_refresh_token bytea,
	status text not null default 'pending_authorization',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (organization_id, seller_id)
);

create table store_marketplaces (
	store_id uuid not null references stores(id) on delete cascade,
	marketplace_id uuid not null references marketplaces(id),
	primary key (store_id, marketplace_id)
);

create table products (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid references stores(id) on delete cascade,
	asin text not null,
	sku text,
	title text not null,
	status text not null default 'active',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (organization_id, store_id, asin, sku)
);

create table orders (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid references stores(id) on delete cascade,
	amazon_order_id text not null,
	marketplace_id uuid references marketplaces(id),
	order_status text not null,
	purchase_date timestamptz not null,
	order_total numeric(14,2) not null default 0,
	currency_code text not null,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (organization_id, store_id, amazon_order_id)
);

create table order_items (
	id uuid primary key default gen_random_uuid(),
	order_id uuid not null references orders(id) on delete cascade,
	product_id uuid references products(id),
	amazon_order_item_id text not null,
	asin text not null,
	sku text,
	title text not null,
	quantity_ordered integer not null default 0,
	item_price numeric(14,2) not null default 0,
	currency_code text not null,
	unique (order_id, amazon_order_item_id)
);

create table inventory (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	product_id uuid not null references products(id) on delete cascade,
	fulfillable_quantity integer not null default 0,
	inbound_quantity integer not null default 0,
	reserved_quantity integer not null default 0,
	updated_at timestamptz not null default now(),
	unique (organization_id, product_id)
);

create table campaigns (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid references stores(id) on delete cascade,
	amazon_campaign_id text not null,
	name text not null,
	campaign_type text not null,
	status text not null,
	budget numeric(14,2) not null default 0,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (organization_id, store_id, amazon_campaign_id)
);

create table campaign_metrics (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	campaign_id uuid not null references campaigns(id) on delete cascade,
	metric_date date not null,
	impressions integer not null default 0,
	clicks integer not null default 0,
	spend numeric(14,2) not null default 0,
	sales numeric(14,2) not null default 0,
	orders integer not null default 0,
	unique (campaign_id, metric_date)
);

create table keywords (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	campaign_id uuid not null references campaigns(id) on delete cascade,
	amazon_keyword_id text not null,
	match_type text not null,
	keyword_text text not null,
	status text not null,
	bid numeric(14,2),
	unique (campaign_id, amazon_keyword_id)
);

create table keyword_metrics (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	keyword_id uuid not null references keywords(id) on delete cascade,
	metric_date date not null,
	impressions integer not null default 0,
	clicks integer not null default 0,
	spend numeric(14,2) not null default 0,
	sales numeric(14,2) not null default 0,
	orders integer not null default 0,
	unique (keyword_id, metric_date)
);

create table financial_transactions (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid references stores(id) on delete cascade,
	transaction_type text not null,
	amount numeric(14,2) not null,
	currency_code text not null,
	posted_at timestamptz not null,
	raw_payload jsonb not null default '{}'::jsonb
);

create table settlements (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid references stores(id) on delete cascade,
	amazon_settlement_id text not null,
	start_date timestamptz not null,
	end_date timestamptz not null,
	deposit_date timestamptz,
	total_amount numeric(14,2) not null default 0,
	currency_code text not null,
	unique (organization_id, store_id, amazon_settlement_id)
);

create table reviews (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	product_id uuid references products(id) on delete cascade,
	rating integer not null,
	title text not null default '',
	body text not null default '',
	reviewed_at timestamptz not null,
	created_at timestamptz not null default now()
);

create table daily_metrics (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid references stores(id) on delete cascade,
	metric_date date not null,
	revenue numeric(14,2) not null default 0,
	profit numeric(14,2) not null default 0,
	units_sold integer not null default 0,
	ad_spend numeric(14,2) not null default 0,
	refunds numeric(14,2) not null default 0,
	unique (organization_id, store_id, metric_date)
);

create table ai_insights (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	insight_type text not null,
	title text not null,
	body text not null,
	priority text not null default 'medium',
	status text not null default 'open',
	created_at timestamptz not null default now()
);

create table notifications (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	user_id uuid references users(id) on delete cascade,
	notification_type text not null,
	title text not null,
	body text not null,
	read_at timestamptz,
	created_at timestamptz not null default now()
);

create table reports (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	created_by uuid references users(id),
	report_type text not null,
	format text not null,
	status text not null default 'queued',
	storage_key text,
	created_at timestamptz not null default now()
);

create table audit_logs (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid references organizations(id) on delete set null,
	actor_user_id uuid references users(id) on delete set null,
	action text not null,
	resource_type text not null,
	resource_id text not null,
	ip_address text not null default '',
	user_agent text not null default '',
	created_at timestamptz not null default now()
);

create table api_tokens (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	name text not null,
	token_hash text not null unique,
	scopes text[] not null default '{}',
	expires_at timestamptz,
	created_at timestamptz not null default now()
);

create table sync_history (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid references stores(id) on delete cascade,
	sync_type text not null,
	status text not null,
	started_at timestamptz not null default now(),
	finished_at timestamptz,
	records_processed integer not null default 0,
	error_message text
);

create table webhook_events (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid references organizations(id) on delete cascade,
	store_id uuid references stores(id) on delete cascade,
	provider text not null,
	event_type text not null,
	payload jsonb not null,
	processed_at timestamptz,
	created_at timestamptz not null default now()
);

insert into roles (name, description) values
	('admin', 'RetailPulse platform administrator'),
	('organization_owner', 'Organization owner with full tenant access'),
	('seller', 'Amazon seller account user'),
	('manager', 'Manager with operational access'),
	('viewer', 'Read-only organization user')
on conflict (name) do nothing;

insert into permissions (name, description) values
	('organization.manage', 'Manage organization settings'),
	('users.manage', 'Invite and manage users'),
	('stores.connect', 'Connect Amazon stores'),
	('analytics.read', 'Read analytics dashboards'),
	('reports.manage', 'Create and schedule reports')
on conflict (name) do nothing;

insert into role_permissions (role_id, permission_id)
select r.id, p.id
from roles r
cross join permissions p
where r.name in ('admin', 'organization_owner')
on conflict do nothing;

insert into role_permissions (role_id, permission_id)
select r.id, p.id
from roles r
join permissions p on p.name in ('stores.connect', 'analytics.read', 'reports.manage')
where r.name = 'manager'
on conflict do nothing;

insert into role_permissions (role_id, permission_id)
select r.id, p.id
from roles r
join permissions p on p.name in ('stores.connect', 'analytics.read', 'reports.manage')
where r.name = 'seller'
on conflict do nothing;

insert into role_permissions (role_id, permission_id)
select r.id, p.id
from roles r
join permissions p on p.name = 'analytics.read'
where r.name = 'viewer'
on conflict do nothing;

insert into marketplaces (amazon_marketplace_id, country_code, name, region, currency_code) values
	('ATVPDKIKX0DER', 'US', 'Amazon.com', 'NA', 'USD'),
	('A2EUQ1WTGCTBG2', 'CA', 'Amazon.ca', 'NA', 'CAD'),
	('A1AM78C64UM0Y8', 'MX', 'Amazon.com.mx', 'NA', 'MXN'),
	('A1PA6795UKMFR9', 'DE', 'Amazon.de', 'EU', 'EUR'),
	('A1F83G8C2ARO7P', 'GB', 'Amazon.co.uk', 'EU', 'GBP'),
	('A13V1IB3VIYZZH', 'FR', 'Amazon.fr', 'EU', 'EUR'),
	('APJ6JRA9NG5V4', 'IT', 'Amazon.it', 'EU', 'EUR'),
	('A1RKKUPIHCS9HS', 'ES', 'Amazon.es', 'EU', 'EUR'),
	('A21TJRUUN4KGV', 'IN', 'Amazon.in', 'FE', 'INR'),
	('A1VC38T7YXB528', 'JP', 'Amazon.co.jp', 'FE', 'JPY'),
	('A39IBJ37TRP1C6', 'AU', 'Amazon.com.au', 'FE', 'AUD')
on conflict (amazon_marketplace_id) do nothing;
