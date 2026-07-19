insert into roles (name, description) values
	('seller', 'Amazon seller account user')
on conflict (name) do nothing;

insert into role_permissions (role_id, permission_id)
select r.id, p.id
from roles r
join permissions p on p.name in ('stores.connect', 'analytics.read', 'reports.manage')
where r.name = 'seller'
on conflict do nothing;

with org as (
	insert into organizations (name, slug)
	values ('RetailPulse Demo', 'retailpulse-demo')
	on conflict (slug) do update set name = excluded.name
	returning id
),
owner_role as (
	select id from roles where name = 'organization_owner'
)
insert into users (organization_id, role_id, email, name, password_hash, status)
select org.id, owner_role.id, 'owner@retailpulse.local', 'RetailPulse Owner', crypt('RetailPulse@12345', gen_salt('bf')), 'active'
from org, owner_role
on conflict (email) do update set
	organization_id = excluded.organization_id,
	role_id = excluded.role_id,
	name = excluded.name,
	password_hash = excluded.password_hash,
	status = excluded.status;

with org as (
	select id from organizations where slug = 'retailpulse-demo'
),
seller_role as (
	select id from roles where name = 'seller'
)
insert into users (organization_id, role_id, email, name, password_hash, status)
select org.id, seller_role.id, 'seller@retailpulse.local', 'RetailPulse Seller', crypt('RetailPulse@12345', gen_salt('bf')), 'active'
from org, seller_role
on conflict (email) do update set
	organization_id = excluded.organization_id,
	role_id = excluded.role_id,
	name = excluded.name,
	password_hash = excluded.password_hash,
	status = excluded.status;
