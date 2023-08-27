create table if not exists oauth_accounts (
	account_id character(36) primary key,
	vendor varchar(64) not null,
	vendor_uid varchar(64) not null,
	create_at int not null,
	update_at int not null
);

create unique index if not exists ux_oauth_vendor on oauth_accounts(account_id, vendor);
