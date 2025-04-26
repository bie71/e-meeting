-- Drop indexes
DROP INDEX IF EXISTS idx_reservations_time_range;
DROP INDEX IF EXISTS idx_reservations_status;
DROP INDEX IF EXISTS idx_reservations_user_id;
DROP INDEX IF EXISTS idx_reservations_room_id;

-- Drop tables
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS rooms;