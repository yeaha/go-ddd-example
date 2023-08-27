create table if not exists accounts (
	id character(36) primary key,
	email varchar(255) not null,
	password varchar(36) not null,
	setting json,
	create_at int not null,
	update_at int not null
);
create unique index if not exists accounts_email_ukey on accounts(email);
