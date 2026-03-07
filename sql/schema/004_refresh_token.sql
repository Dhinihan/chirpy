-- +goose Up
CREATE TABLE refresh_token (
	token text primary key,
	created_at timestamp not null default CURRENT_TIMESTAMP, 
	updated_at timestamp not null default CURRENT_TIMESTAMP, 
	user_id uuid not null references users(id) ON DELETE cascade, 
	expires_at timestamp not null,
	revoked_at timestamp null
);

-- +goose Down
DROP TABLE refresh_token;
