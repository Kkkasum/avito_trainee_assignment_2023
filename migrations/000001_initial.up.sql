-- users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- segments
CREATE TABLE segments (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(50),
    CONSTRAINT unique_segments_slug UNIQUE (slug)
);
CREATE INDEX idx_segments_slug ON segments(slug);

-- users_segments
CREATE TABLE users_segments (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    segment_id INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT unique_users_segments_user_id_segment_id UNIQUE (user_id, segment_id),
    CONSTRAINT fk_users_segments_user_id FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT fk_users_segments_segment_id FOREIGN KEY (segment_id) REFERENCES segments (id)
);
CREATE INDEX idx_users_segments_user_id ON users_segments(user_id);