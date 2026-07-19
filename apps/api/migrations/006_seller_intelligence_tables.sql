create table if not exists product_traffic_metrics (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references stores(id) on delete cascade,
	product_id uuid not null references products(id) on delete cascade,
	metric_date date not null,
	sessions integer not null default 0,
	page_views integer not null default 0,
	buy_box_percentage numeric(7,2) not null default 0,
	units_ordered integer not null default 0,
	ordered_revenue numeric(14,2) not null default 0,
	data_origin text not null default 'amazon',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, product_id, metric_date)
);

create index if not exists idx_product_traffic_metrics_store_date
	on product_traffic_metrics (organization_id, store_id, metric_date desc);

create table if not exists search_term_metrics (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references stores(id) on delete cascade,
	product_id uuid references products(id) on delete cascade,
	campaign_id uuid references campaigns(id) on delete set null,
	search_term text not null,
	keyword_text text not null default '',
	match_type text not null default '',
	metric_date date not null,
	impressions integer not null default 0,
	clicks integer not null default 0,
	spend numeric(14,2) not null default 0,
	sales numeric(14,2) not null default 0,
	orders integer not null default 0,
	data_origin text not null default 'amazon',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, product_id, campaign_id, search_term, metric_date)
);

create index if not exists idx_search_term_metrics_store_date
	on search_term_metrics (organization_id, store_id, metric_date desc);

create table if not exists return_events (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references stores(id) on delete cascade,
	order_id uuid references orders(id) on delete set null,
	product_id uuid references products(id) on delete set null,
	return_date date not null,
	quantity integer not null default 1,
	reason text not null default '',
	status text not null default 'completed',
	refund_amount numeric(14,2) not null default 0,
	currency_code text not null default 'USD',
	data_origin text not null default 'amazon',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, order_id, product_id, return_date, reason)
);

create index if not exists idx_return_events_store_date
	on return_events (organization_id, store_id, return_date desc);

create table if not exists regional_sales_metrics (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references stores(id) on delete cascade,
	marketplace_id uuid references marketplaces(id),
	country_code text not null default '',
	region_code text not null default '',
	city text not null default '',
	metric_date date not null,
	orders integer not null default 0,
	units integer not null default 0,
	revenue numeric(14,2) not null default 0,
	refunds numeric(14,2) not null default 0,
	ad_spend numeric(14,2) not null default 0,
	data_origin text not null default 'amazon',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, country_code, region_code, city, metric_date)
);

create index if not exists idx_regional_sales_metrics_store_date
	on regional_sales_metrics (organization_id, store_id, metric_date desc);

create table if not exists growth_recommendations (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references stores(id) on delete cascade,
	product_id uuid references products(id) on delete cascade,
	campaign_id uuid references campaigns(id) on delete set null,
	region_code text not null default '',
	recommendation_type text not null,
	title text not null,
	reason text not null,
	evidence jsonb not null default '{}'::jsonb,
	impact_score numeric(8,2) not null default 0,
	confidence numeric(5,2) not null default 0,
	status text not null default 'open',
	data_origin text not null default 'amazon',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now()
);

create index if not exists idx_growth_recommendations_store_status
	on growth_recommendations (organization_id, store_id, status, created_at desc);
