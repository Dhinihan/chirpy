-- +goose up
CREATE TABLE chirps (
	id uuid primary key,
	created_at timestamp not null default CURRENT_TIMESTAMP,
	updated_at timestamp not null default CURRENT_TIMESTAMP,
	body text not null,
	user_id uuid not null references users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chirps;
