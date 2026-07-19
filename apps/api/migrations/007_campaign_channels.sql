alter table campaigns
	add column if not exists channel text not null default 'amazon_ads';

create index if not exists idx_campaigns_store_channel
	on campaigns (organization_id, store_id, channel);

update campaigns
set channel = 'amazon_ads'
where channel is null or channel = '';
