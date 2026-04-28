CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL DEFAULT '',
    role       VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'user')),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users (id, email, password_hash, role) 
VALUES 
    ('00000000-0000-0000-0000-000000000001', 'admin@example.com', '', 'admin'),
    ('00000000-0000-0000-0000-000000000002', 'user@example.com', '', 'user')
ON CONFLICT (email) DO NOTHING;

CREATE TABLE IF NOT EXISTS rooms (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    capacity    INTEGER,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS schedules (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id      UUID NOT NULL UNIQUE REFERENCES rooms(id) ON DELETE CASCADE,
    days_of_week INTEGER[] NOT NULL,
    start_time   TIME NOT NULL,
    end_time     TIME NOT NULL,
    created_at   TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS slots (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id    UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ NOT NULL,
    UNIQUE (room_id, start_time)
);

CREATE TABLE IF NOT EXISTS bookings (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slot_id    UUID NOT NULL REFERENCES slots(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status     VARCHAR(50) NOT NULL CHECK (status IN ('active', 'cancelled')),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_active_slot ON bookings (slot_id) WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings (user_id);
