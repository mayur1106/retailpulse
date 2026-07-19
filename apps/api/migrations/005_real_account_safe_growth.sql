alter table stores
	add column if not exists region text not null default 'NA';

alter table stores
	drop constraint if exists stores_organization_id_seller_id_key;

create unique index if not exists idx_stores_org_seller_environment
	on stores (organization_id, seller_id, environment);

alter table products
	add column if not exists data_origin text not null default 'amazon',
	add column if not exists category text not null default '',
	add column if not exists cost_price numeric(14,2) not null default 0,
	add column if not exists selling_price numeric(14,2) not null default 0;

alter table orders
	add column if not exists data_origin text not null default 'amazon';

alter table inventory
	add column if not exists data_origin text not null default 'amazon';

alter table campaigns
	add column if not exists data_origin text not null default 'amazon';

alter table campaign_metrics
	add column if not exists data_origin text not null default 'amazon';

alter table financial_transactions
	add column if not exists data_origin text not null default 'amazon';

alter table reports
	add column if not exists store_id uuid references stores(id) on delete cascade,
	add column if not exists data_origin text not null default 'amazon';

alter table daily_metrics
	add column if not exists data_origin text not null default 'amazon';

create unique index if not exists idx_products_org_store_asin_sku_key
	on products (organization_id, store_id, asin, coalesce(sku, ''));

create index if not exists idx_orders_org_store_purchase_date
	on orders (organization_id, store_id, purchase_date desc);

create index if not exists idx_daily_metrics_org_store_date
	on daily_metrics (organization_id, store_id, metric_date desc);

create index if not exists idx_campaigns_org_store
	on campaigns (organization_id, store_id);

create table if not exists advertised_product_metrics (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references stores(id) on delete cascade,
	product_id uuid not null references products(id) on delete cascade,
	campaign_id uuid references campaigns(id) on delete set null,
	metric_date date not null,
	impressions integer not null default 0,
	clicks integer not null default 0,
	spend numeric(14,2) not null default 0,
	sales numeric(14,2) not null default 0,
	orders integer not null default 0,
	data_origin text not null default 'amazon',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, product_id, campaign_id, metric_date)
);

create index if not exists idx_advertised_product_metrics_org_store_date
	on advertised_product_metrics (organization_id, store_id, metric_date desc);
