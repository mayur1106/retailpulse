alter table stores
	add column if not exists environment text not null default 'production';

alter table stores
	drop constraint if exists stores_environment_check;

alter table stores
	add constraint stores_environment_check check (environment in ('production', 'sandbox'));
