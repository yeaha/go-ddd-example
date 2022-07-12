begin;

create schema if not exists users;

create table if not exists users.users (
	id uuid not null,
	email character varying(255) not null,
	password character varying(36) not null,
	setting jsonb,
	create_at timestamp with time zone not null,
	update_at timestamp with time zone not null,
	constraint users_pkey primary key (id)
);
comment on table users.users is '系统账号';
create unique index if not exists users_email_ukey on users.users(email);

commit;
