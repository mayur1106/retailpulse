create table if not exists commerce_stores (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	name text not null,
	slug text not null,
	domain text not null default '',
	logo_url text not null default '',
	currency_code text not null default 'INR',
	country_code text not null default 'IN',
	timezone text not null default 'Asia/Kolkata',
	status text not null default 'draft',
	settings jsonb not null default '{}'::jsonb,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (organization_id, slug)
);

create table if not exists commerce_categories (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	name text not null,
	slug text not null,
	description text not null default '',
	image_url text not null default '',
	sort_order integer not null default 0,
	status text not null default 'active',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, slug)
);

create table if not exists commerce_products (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	category_id uuid references commerce_categories(id) on delete set null,
	title text not null,
	slug text not null,
	description text not null default '',
	sku text not null default '',
	brand text not null default '',
	status text not null default 'draft',
	price numeric(14,2) not null default 0,
	compare_at_price numeric(14,2) not null default 0,
	cost_price numeric(14,2) not null default 0,
	currency_code text not null default 'INR',
	images jsonb not null default '[]'::jsonb,
	options jsonb not null default '{}'::jsonb,
	tags text[] not null default '{}',
	seo_title text not null default '',
	seo_description text not null default '',
	is_featured boolean not null default false,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, slug)
);

create table if not exists commerce_product_variants (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	product_id uuid not null references commerce_products(id) on delete cascade,
	title text not null,
	sku text not null,
	color text not null default '',
	size text not null default '',
	price numeric(14,2) not null default 0,
	compare_at_price numeric(14,2) not null default 0,
	cost_price numeric(14,2) not null default 0,
	stock_quantity integer not null default 0,
	reserved_quantity integer not null default 0,
	low_stock_threshold integer not null default 5,
	status text not null default 'active',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, sku)
);

create table if not exists commerce_customers (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	email text not null,
	name text not null,
	phone text not null default '',
	country_code text not null default 'IN',
	region_code text not null default '',
	city text not null default '',
	total_spent numeric(14,2) not null default 0,
	order_count integer not null default 0,
	tags text[] not null default '{}',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, email)
);

create table if not exists commerce_orders (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	customer_id uuid references commerce_customers(id) on delete set null,
	order_number text not null,
	status text not null default 'pending',
	payment_status text not null default 'pending',
	fulfillment_status text not null default 'unfulfilled',
	currency_code text not null default 'INR',
	subtotal numeric(14,2) not null default 0,
	discount_total numeric(14,2) not null default 0,
	shipping_total numeric(14,2) not null default 0,
	tax_total numeric(14,2) not null default 0,
	total numeric(14,2) not null default 0,
	coupon_code text not null default '',
	shipping_address jsonb not null default '{}'::jsonb,
	tracking_number text not null default '',
	placed_at timestamptz not null default now(),
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, order_number)
);

create table if not exists commerce_order_items (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	order_id uuid not null references commerce_orders(id) on delete cascade,
	product_id uuid references commerce_products(id) on delete set null,
	variant_id uuid references commerce_product_variants(id) on delete set null,
	title text not null,
	sku text not null default '',
	quantity integer not null default 1,
	unit_price numeric(14,2) not null default 0,
	total_price numeric(14,2) not null default 0
);

create table if not exists commerce_coupons (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	code text not null,
	name text not null,
	discount_type text not null default 'percentage',
	discount_value numeric(14,2) not null default 0,
	minimum_order_value numeric(14,2) not null default 0,
	usage_limit integer not null default 0,
	used_count integer not null default 0,
	starts_at timestamptz,
	expires_at timestamptz,
	status text not null default 'active',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, code)
);

create table if not exists commerce_shipping_zones (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	name text not null,
	country_code text not null default 'IN',
	region_codes text[] not null default '{}',
	rate_type text not null default 'flat',
	rate numeric(14,2) not null default 0,
	free_shipping_threshold numeric(14,2) not null default 0,
	estimated_days_min integer not null default 3,
	estimated_days_max integer not null default 7,
	cod_enabled boolean not null default true,
	status text not null default 'active',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create table if not exists commerce_returns (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	order_id uuid references commerce_orders(id) on delete set null,
	product_id uuid references commerce_products(id) on delete set null,
	reason text not null default '',
	status text not null default 'requested',
	refund_amount numeric(14,2) not null default 0,
	requested_at timestamptz not null default now(),
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create table if not exists commerce_cms_pages (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	slug text not null,
	title text not null,
	body text not null default '',
	status text not null default 'published',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, slug)
);

create index if not exists idx_commerce_products_store_status on commerce_products (store_id, status, updated_at desc);
create index if not exists idx_commerce_orders_store_placed_at on commerce_orders (store_id, placed_at desc);
create index if not exists idx_commerce_customers_store_city on commerce_customers (store_id, region_code, city);
create index if not exists idx_commerce_variants_product on commerce_product_variants (product_id);
