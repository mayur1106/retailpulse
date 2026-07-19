create table if not exists commerce_homepage_sections (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	section_key text not null,
	section_type text not null default 'product_grid',
	title text not null default '',
	subtitle text not null default '',
	layout text not null default 'grid',
	image_url text not null default '',
	cta_label text not null default '',
	cta_href text not null default '',
	category_slug text not null default '',
	product_source text not null default 'all',
	max_items integer not null default 12,
	content jsonb not null default '{}'::jsonb,
	sort_order integer not null default 0,
	status text not null default 'active',
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, section_key)
);

create table if not exists commerce_payment_methods (
	id uuid primary key default gen_random_uuid(),
	organization_id uuid not null references organizations(id) on delete cascade,
	store_id uuid not null references commerce_stores(id) on delete cascade,
	code text not null,
	name text not null,
	provider text not null default 'manual',
	instructions text not null default '',
	sort_order integer not null default 0,
	status text not null default 'inactive',
	settings jsonb not null default '{}'::jsonb,
	created_at timestamptz not null default now(),
	updated_at timestamptz not null default now(),
	unique (store_id, code)
);

create index if not exists idx_commerce_homepage_sections_store_status
	on commerce_homepage_sections (store_id, status, sort_order);

create index if not exists idx_commerce_payment_methods_store_status
	on commerce_payment_methods (store_id, status, sort_order);
