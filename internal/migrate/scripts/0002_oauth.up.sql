create table if not exists oauth_accounts (
	account_id character(36),
	vendor varchar(64) not null,
	vendor_uid varchar(64) not null,
	create_at int not null,
	update_at int not null,
	primary key (account_id, vendor)
);

create index ix_oauth_vendor_uid on oauth_accounts (vendor_uid);
