CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS reservations (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    room_id INTEGER NOT NULL REFERENCES rooms(id),
    snack_id INTEGER REFERENCES snacks(id),
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    company VARCHAR(255),
    total_participant INTEGER NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'booked',
    notes TEXT,
    duration INTEGER NOT NULL,
    total_price_room DECIMAL(10,2) NOT NULL,
    total_price_snack DECIMAL(10,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
); 