DROP INDEX IF EXISTS idx_bookings_user_id;
DROP INDEX IF EXISTS idx_bookings_active_slot;

DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS slots;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS users;
