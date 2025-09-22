package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Optimized conductor queries with proper indexing and caching
type ConductorQueryService struct {
	repo  *BookingRepository
	cache *QueryCache
}

type PassengerInfo struct {
	Name         string
	SeatNumber   string
	CarriageID   string
	Origin       string
	Destination  string
	BookingID    string
}

// Optimized boarding query with index usage
func (c *ConductorQueryService) GetPassengersBoardingAt(ctx context.Context, 
	serviceID, stationName string, date time.Time) ([]PassengerInfo, error) {
	
	// Use index on (service_id, booking_date, origin)
	query := `
		SELECT passenger_name, seat_number, carriage_id, origin, destination, booking_id
		FROM seat_reservations 
		WHERE service_id = $1 AND booking_date = $2 AND origin = $3
		ORDER BY carriage_id, seat_number`
	
	rows, err := c.repo.db.QueryContext(ctx, query, serviceID, date, stationName)
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

// Optimized alighting query
func (c *ConductorQueryService) GetPassengersAlightingAt(ctx context.Context, 
	serviceID, stationName string, date time.Time) ([]PassengerInfo, error) {
	
	query := `
		SELECT passenger_name, seat_number, carriage_id, origin, destination, booking_id
		FROM seat_reservations 
		WHERE service_id = $1 AND booking_date = $2 AND destination = $3
		ORDER BY carriage_id, seat_number`
	
	rows, err := c.repo.db.QueryContext(ctx, query, serviceID, date, stationName)
	if err != nil {
		return nil, fmt.Errorf("failed to query alighting passengers: %w", err)
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

// Optimized between-stations query with route validation
func (c *ConductorQueryService) GetPassengersBetweenStations(ctx context.Context, 
	serviceID, station1, station2 string, date time.Time) ([]PassengerInfo, error) {
	
	// First get route information to validate station order
	var routeStops []string
	routeQuery := `
		SELECT stop_name FROM route_stops rs
		JOIN services s ON s.route_id = rs.route_id
		WHERE s.service_id = $1
		ORDER BY rs.stop_order`
	
	rows, err := c.repo.db.QueryContext(ctx, routeQuery, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get route stops: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var stopName string
		if err := rows.Scan(&stopName); err != nil {
			return nil, fmt.Errorf("failed to scan stop: %w", err)
		}
		routeStops = append(routeStops, stopName)
	}
	
	// Find station indices
	station1Index := -1
	station2Index := -1
	for i, stop := range routeStops {
		if stop == station1 {
			station1Index = i
		}
		if stop == station2 {
			station2Index = i
		}
	}
	
	if station1Index == -1 || station2Index == -1 {
		return nil, fmt.Errorf("station not found on route")
	}
	
	// Ensure correct order
	if station1Index > station2Index {
		station1Index, station2Index = station2Index, station1Index
	}
	
	// Query passengers whose journey spans the requested segment
	query := `
		SELECT sr.passenger_name, sr.seat_number, sr.carriage_id, 
		       sr.origin, sr.destination, sr.booking_id
		FROM seat_reservations sr
		JOIN route_stops origin_stops ON origin_stops.stop_name = sr.origin
		JOIN route_stops dest_stops ON dest_stops.stop_name = sr.destination
		JOIN services s ON s.service_id = sr.service_id
		WHERE sr.service_id = $1 AND sr.booking_date = $2
		AND origin_stops.route_id = s.route_id AND dest_stops.route_id = s.route_id
		AND origin_stops.stop_order <= $3 AND dest_stops.stop_order >= $4
		ORDER BY sr.carriage_id, sr.seat_number`
	
	rows, err = c.repo.db.QueryContext(ctx, query, serviceID, date, station1Index, station2Index)
	if err != nil {
		return nil, fmt.Errorf("failed to query passengers between stations: %w", err)
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

// Optimized seat lookup
func (c *ConductorQueryService) GetPassengerOnSeat(ctx context.Context, 
	serviceID, carriageID, seatNumber string, date time.Time) (*PassengerInfo, error) {
	
	query := `
		SELECT passenger_name, seat_number, carriage_id, origin, destination, booking_id
		FROM seat_reservations 
		WHERE service_id = $1 AND carriage_id = $2 AND seat_number = $3 AND booking_date = $4`
	
	var p PassengerInfo
	err := c.repo.db.QueryRowContext(ctx, query, serviceID, carriageID, seatNumber, date).
		Scan(&p.Name, &p.SeatNumber, &p.CarriageID, &p.Origin, &p.Destination, &p.BookingID)
	
	if err == sql.ErrNoRows {
		return nil, nil // No passenger found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query passenger on seat: %w", err)
	}
	
	return &p, nil
}
