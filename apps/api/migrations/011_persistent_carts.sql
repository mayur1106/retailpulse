create table if not exists commerce_carts (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	customer_id uuid references commerce_customers(id) on delete set null,
	cart_token text not null,
	visitor_id text not null default '',
	email text not null default '',
	name text not null default '',
	phone text not null default '',
	status text not null default 'active',
	currency_code text not null default 'INR',
	subtotal numeric(14,2) not null default 0,
	item_count integer not null default 0,
	metadata jsonb not null default '{}'::jsonb,
	first_seen_at timestamptz not null default now(),
	last_activity_at timestamptz not null default now(),
	checkout_started_at timestamptz,
	abandoned_at timestamptz,
	converted_at timestamptz,
	converted_order_id uuid references commerce_orders(id) on delete set null,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, cart_token)
);

create table if not exists commerce_cart_items (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	cart_id uuid not null references commerce_carts(id) on delete cascade,
	product_id uuid references commerce_products(id) on delete set null,
	variant_id uuid references commerce_product_variants(id) on delete set null,
	product_title text not null,
	variant_title text not null default '',
	sku text not null default '',
	image_url text not null default '',
	unit_price numeric(14,2) not null default 0,
	quantity integer not null default 1,
	added_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	removed_at timestamptz
);

create unique index if not exists idx_commerce_cart_items_active_variant
	on commerce_cart_items (cart_id, coalesce(product_id, '00000000-0000-0000-0000-000000000000'::uuid), coalesce(variant_id, '00000000-0000-0000-0000-000000000000'::uuid))
	where removed_at is null;

create index if not exists idx_commerce_carts_store_status_activity
	on commerce_carts (organization_id, store_id, status, last_activity_at desc);

create index if not exists idx_commerce_cart_items_product_active
	on commerce_cart_items (organization_id, store_id, product_id)
	where removed_at is null;
