create table if not exists api_keys (
	id          varchar(64) primary key,
	hash        text not null,
	user_id     varchar(255) not null,
	created_at  timestamp not null default (now() at time zone 'utc'),
	constraint fk_api_keys_user_id
		foreign key (user_id)
		references users (id)
		on delete cascade,
	constraint uq_api_keys_hash
		unique (hash)
);
