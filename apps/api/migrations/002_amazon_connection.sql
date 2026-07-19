alter table stores
	add column if not exists region text not null default 'NA',
	add column if not exists last_imported_at timestamptz;

create table if not exists amazon_oauth_states (
	state text primary key,
	organization_id uuid not null references organizations(id) on delete cascade,
	user_id uuid not null references users(id) on delete cascade,
	region text not null,
	marketplace_id text not null,
	created_at timestamptz not null default now(),
	expires_at timestamptz not null,
	used_at timestamptz
);

create index if not exists idx_amazon_oauth_states_org on amazon_oauth_states(organization_id);
