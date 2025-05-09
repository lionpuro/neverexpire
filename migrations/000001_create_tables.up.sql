create table if not exists users (
	id         varchar(255) primary key,
	email      text not null,
	created_at timestamp not null default (now() at time zone 'utc'),
	updated_at timestamp not null default (now() at time zone 'utc')
);
