CREATE TABLE IF NOT EXISTS addresses (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    street_house VARCHAR(255) NOT NULL,
    apartment VARCHAR(50),
    entrance VARCHAR(50),
    floor VARCHAR(50),
    door_code VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);