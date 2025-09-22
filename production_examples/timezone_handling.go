package main

import (
	"context"
	"fmt"
	"time"
)

// Timezone-aware booking system
type TimezoneAwareBookingService struct {
	repo *BookingRepository
}

type BookingRequest struct {
	ServiceID      string
	Origin         string
	Destination    string
	Passengers     []Passenger
	SeatRequests   []SeatRequest
	DepartureTime  time.Time // Full datetime with timezone
	Timezone       string    // IANA timezone identifier (e.g., "Europe/Paris")
}

type ServiceSchedule struct {
	ServiceID     string
	RouteID       string
	DepartureTime time.Time // UTC time
	ArrivalTime   time.Time // UTC time
	Timezone      string    // Service timezone
}

// Convert local time to UTC for storage
func (s *TimezoneAwareBookingService) ConvertToUTC(localTime time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}
	
	// Parse the local time in the specified timezone
	localTimeInTZ := time.Date(
		localTime.Year(), localTime.Month(), localTime.Day(),
		localTime.Hour(), localTime.Minute(), localTime.Second(),
		localTime.Nanosecond(), loc)
	
	return localTimeInTZ.UTC(), nil
}

// Check seat availability with timezone-aware date matching
func (s *TimezoneAwareBookingService) IsSeatAvailableWithTimezone(ctx context.Context, 
	req BookingRequest) (bool, error) {
	
	// Convert request time to UTC
	utcTime, err := s.ConvertToUTC(req.DepartureTime, req.Timezone)
	if err != nil {
		return false, fmt.Errorf("failed to convert time to UTC: %w", err)
	}
	
	// Get service schedule to validate timezone
	var serviceSchedule ServiceSchedule
	err = s.repo.db.QueryRowContext(ctx, `
		SELECT service_id, route_id, departure_time, arrival_time, timezone
		FROM service_schedules 
		WHERE service_id = $1`, req.ServiceID).Scan(
		&serviceSchedule.ServiceID, &serviceSchedule.RouteID,
		&serviceSchedule.DepartureTime, &serviceSchedule.ArrivalTime,
		&serviceSchedule.Timezone)
	
	if err != nil {
		return false, fmt.Errorf("failed to get service schedule: %w", err)
	}
	
	// Validate that request timezone matches service timezone
	if req.Timezone != serviceSchedule.Timezone {
		return false, fmt.Errorf("timezone mismatch: request %s, service %s", 
			req.Timezone, serviceSchedule.Timezone)
	}
	
	// Check if request time is within service operating window
	// Allow bookings up to 30 minutes before departure
	cutoffTime := serviceSchedule.DepartureTime.Add(-30 * time.Minute)
	if utcTime.Before(cutoffTime) || utcTime.After(serviceSchedule.ArrivalTime) {
		return false, fmt.Errorf("booking time outside service operating window")
	}
	
	// Check seat availability for the specific datetime
	for _, seatReq := range req.SeatRequests {
		var count int
		err = s.repo.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM seat_reservations 
			WHERE service_id = $1 AND carriage_id = $2 AND seat_number = $3 
			AND departure_time = $4`, 
			req.ServiceID, seatReq.CarriageID, seatReq.SeatNumber, utcTime).Scan(&count)
		
		if err != nil {
			return false, fmt.Errorf("failed to check seat availability: %w", err)
		}
		
		if count > 0 {
			return false, nil // Seat is taken
		}
	}
	
	return true, nil
}

// Create booking with timezone information
func (s *TimezoneAwareBookingService) CreateBookingWithTimezone(ctx context.Context, 
	req BookingRequest) (*Booking, error) {
	
	// Validate timezone and convert to UTC
	utcTime, err := s.ConvertToUTC(req.DepartureTime, req.Timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to convert time to UTC: %w", err)
	}
	
	// Check availability
	available, err := s.IsSeatAvailableWithTimezone(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}
	
	if !available {
		return nil, fmt.Errorf("seats not available for requested time")
	}
	
	// Create booking with timezone metadata
	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	bookingID := generateBookingID()
	
	// Insert booking with timezone information
	_, err = tx.ExecContext(ctx, `
		INSERT INTO bookings (booking_id, service_id, departure_time_utc, 
			departure_time_local, timezone, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		bookingID, req.ServiceID, utcTime, req.DepartureTime, req.Timezone, time.Now())
	
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}
	
	// Insert seat reservations
	for i, seatReq := range req.SeatRequests {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO seat_reservations 
			(booking_id, service_id, carriage_id, seat_number, passenger_name,
			 origin, destination, departure_time_utc, timezone)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			bookingID, req.ServiceID, seatReq.CarriageID, seatReq.SeatNumber,
			req.Passengers[i].Name, req.Origin, req.Destination, utcTime, req.Timezone)
		
		if err != nil {
			return nil, fmt.Errorf("failed to create seat reservation: %w", err)
		}
	}
	
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit booking: %w", err)
	}
	
	return &Booking{ID: bookingID}, nil
}

// Timezone-aware conductor queries
func (s *TimezoneAwareBookingService) GetPassengersBoardingAtWithTimezone(ctx context.Context, 
	serviceID, stationName string, localTime time.Time, timezone string) ([]PassengerInfo, error) {
	
	// Convert local time to UTC
	utcTime, err := s.ConvertToUTC(localTime, timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to convert time to UTC: %w", err)
	}
	
	// Query with timezone-aware matching
	query := `
		SELECT passenger_name, seat_number, carriage_id, origin, destination, booking_id
		FROM seat_reservations 
		WHERE service_id = $1 AND origin = $2 
		AND DATE(departure_time_utc AT TIME ZONE timezone) = DATE($3 AT TIME ZONE $4)
		ORDER BY carriage_id, seat_number`
	
	rows, err := s.repo.db.QueryContext(ctx, query, serviceID, stationName, utcTime, timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to query boarding passengers: %w", err)
	}
	defer rows.Close()
	
	var passengers []PassengerInfo
	for rows.Next() {
		var p PassengerInfo
		err := rows.Scan(&p.Name, &p.SeatNumber, &p.CarriageID, 
			&p.Origin, &p.Destination, &p.BookingID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan passenger: %w", err)
		}
		passengers = append(passengers, p)
	}
	
	return passengers, nil
}
