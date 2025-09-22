package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Production-ready persistence layer
type BookingRepository struct {
	db *sql.DB
}

type BookingRecord struct {
	ID           string
	ServiceID    string
	CarriageID   string
	SeatNumber   string
	PassengerName string
	Origin       string
	Destination  string
	BookingDate  time.Time
	CreatedAt    time.Time
	Version      int // For optimistic locking
}

// Optimistic locking for seat reservations
func (r *BookingRepository) ReserveSeat(ctx context.Context, req SeatReservationRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check seat availability with row-level locking
	var existingBooking string
	err = tx.QueryRowContext(ctx, `
		SELECT booking_id FROM seat_reservations 
		WHERE service_id = $1 AND carriage_id = $2 AND seat_number = $3 
		AND booking_date = $4 
		FOR UPDATE`, req.ServiceID, req.CarriageID, req.SeatNumber, req.BookingDate).Scan(&existingBooking)
	
	if err == nil {
		return fmt.Errorf("seat already booked: %w", ErrSeatUnavailable)
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check seat availability: %w", err)
	}

	// Insert new reservation
	_, err = tx.ExecContext(ctx, `
		INSERT INTO seat_reservations 
		(booking_id, service_id, carriage_id, seat_number, passenger_name, 
		 origin, destination, booking_date, created_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 1)`,
		req.BookingID, req.ServiceID, req.CarriageID, req.SeatNumber,
		req.PassengerName, req.Origin, req.Destination, req.BookingDate, time.Now())
	
	if err != nil {
		return fmt.Errorf("failed to create reservation: %w", err)
	}

	return tx.Commit()
}

// Database schema with proper indexes
const createTablesSQL = `
CREATE TABLE IF NOT EXISTS seat_reservations (
    id SERIAL PRIMARY KEY,
    booking_id VARCHAR(50) NOT NULL,
    service_id VARCHAR(50) NOT NULL,
    carriage_id VARCHAR(10) NOT NULL,
    seat_number VARCHAR(10) NOT NULL,
    passenger_name VARCHAR(255) NOT NULL,
    origin VARCHAR(100) NOT NULL,
    destination VARCHAR(100) NOT NULL,
    booking_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Unique constraint to prevent double booking
    UNIQUE(service_id, carriage_id, seat_number, booking_date),
    
    -- Indexes for common queries
    INDEX idx_service_date (service_id, booking_date),
    INDEX idx_origin_destination (origin, destination),
    INDEX idx_passenger_name (passenger_name)
);

-- Partial index for active bookings only
CREATE INDEX IF NOT EXISTS idx_active_bookings 
ON seat_reservations (service_id, booking_date) 
WHERE booking_date >= CURRENT_DATE;
`
