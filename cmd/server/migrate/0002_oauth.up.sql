begin;

create table if not exists users.oauth(
	user_id uuid not null,
	vendor character varying(64) not null,
	vendor_uid character varying(64) not null,
	create_at timestamp with time zone not null,
	update_at timestamp with time zone not null,
	constraint oauth_pkey primary key (user_id, vendor)
);
comment on table users.oauth is '三方登录 oauth方式';
comment on column users.oauth.vendor is '三方站点名称';
comment on column users.oauth.vendor_uid is '三方站点用户id';

create index if not exists ix_oauth_vendor on users.oauth(vendor, vendor_uid);

commit;
