-- +goose up
CREATE TABLE users (
	id uuid primary key,
	created_at timestamp not null default CURRENT_TIMESTAMP,
	updated_at timestamp not null default CURRENT_TIMESTAMP,
	email text not null unique
);

-- +goose Down
DROP TABLE users;
