begin;

create table if not exists account.oauth(
	account_id uuid not null,
	vendor character varying(64) not null,
	vendor_uid character varying(64) not null,
	create_at timestamp with time zone not null,
	update_at timestamp with time zone not null,
	constraint oauth_pkey primary key (account_id, vendor)
);
comment on table account.oauth is '三方登录 oauth方式';
comment on column account.oauth.vendor is '三方站点名称';
comment on column account.oauth.vendor_uid is '三方站点用户id';

create index if not exists ix_oauth_vendor on account.oauth(vendor, vendor_uid);

commit;
