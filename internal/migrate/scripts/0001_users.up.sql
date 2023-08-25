begin;

create schema if not exists account;

create table if not exists account.accounts (
	id uuid not null,
	email character varying(255) not null,
	password character varying(36) not null,
	setting jsonb,
	create_at timestamp with time zone not null,
	update_at timestamp with time zone not null,
	constraint accounts_pkey primary key (id)
);
comment on table account.accounts is '系统账号';
create unique index if not exists accounts_email_ukey on account.accounts(email);

commit;
