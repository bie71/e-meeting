-- Create rooms table
CREATE TABLE rooms_new (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    capacity INT NOT NULL,
    price_per_hour DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'available',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create reservations table
CREATE TABLE reservations_new (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_new_id UUID NOT NULL REFERENCES rooms_new(id),
    user_id UUID NOT NULL REFERENCES users(id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    visitor_count INT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_time_range CHECK (end_time > start_time)
);

-- Create indexes
CREATE INDEX idx_reservations_room_id ON reservations_new(room_new_id);
CREATE INDEX idx_reservations_user_id ON reservations_new(user_id);
CREATE INDEX idx_reservations_status ON reservations_new(status);
CREATE INDEX idx_reservations_time_range ON reservations_new(start_time, end_time);