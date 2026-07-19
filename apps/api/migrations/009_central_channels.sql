create table if not exists commerce_sales_channels (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	code text not null,
	name text not null,
	channel_type text not null,
	status text not null default 'inactive',
	settings jsonb not null default '{}'::jsonb,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, code)
);

create table if not exists commerce_channel_listings (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	channel_id uuid not null references commerce_sales_channels(id) on delete cascade,
	product_id uuid not null references commerce_products(id) on delete cascade,
	variant_id uuid references commerce_product_variants(id) on delete cascade,
	external_product_id text not null default '',
	external_variant_id text not null default '',
	external_sku text not null default '',
	external_url text not null default '',
	listing_status text not null default 'not_listed',
	sync_status text not null default 'pending',
	channel_price numeric(14,2),
	channel_quantity integer,
	last_synced_at timestamptz,
	error_message text not null default '',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (channel_id, product_id, variant_id)
);

create index if not exists idx_commerce_channel_listings_product
	on commerce_channel_listings (organization_id, store_id, product_id);

create index if not exists idx_commerce_channel_listings_sync
	on commerce_channel_listings (organization_id, store_id, sync_status, updated_at desc);

alter table commerce_orders
	add column if not exists channel_code text not null default 'website',
	add column if not exists external_order_id text not null default '';

create index if not exists idx_commerce_orders_channel
	on commerce_orders (organization_id, store_id, channel_code, placed_at desc);
