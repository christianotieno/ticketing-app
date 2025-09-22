package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Optimized seat availability with caching
type SeatAvailabilityCache struct {
	cache map[string]bool // key: "serviceID:carriageID:seatNumber:date"
	mutex sync.RWMutex
	ttl   time.Duration
}

type OptimizedReservationSystem struct {
	repo  *BookingRepository
	cache *SeatAvailabilityCache
}

// O(1) seat availability check with cache
func (s *OptimizedReservationSystem) IsSeatAvailable(ctx context.Context, 
	serviceID, carriageID, seatNumber string, date time.Time) (bool, error) {
	
	key := fmt.Sprintf("%s:%s:%s:%s", serviceID, carriageID, seatNumber, date.Format("2006-01-02"))
	
	// Check cache first
	s.cache.mutex.RLock()
	if available, exists := s.cache.cache[key]; exists {
		s.cache.mutex.RUnlock()
		return available, nil
	}
	s.cache.mutex.RUnlock()
	
	// Cache miss - check database with optimized query
	var count int
	err := s.repo.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM seat_reservations 
		WHERE service_id = $1 AND carriage_id = $2 AND seat_number = $3 
		AND booking_date = $4`, serviceID, carriageID, seatNumber, date).Scan(&count)
	
	if err != nil {
		return false, fmt.Errorf("failed to check seat availability: %w", err)
	}
	
	available := count == 0
	
	// Update cache
	s.cache.mutex.Lock()
	s.cache.cache[key] = available
	s.cache.mutex.Unlock()
	
	return available, nil
}

// Batch seat availability check for multiple seats
func (s *OptimizedReservationSystem) CheckMultipleSeats(ctx context.Context, 
	seatRequests []SeatRequest, serviceID string, date time.Time) (map[string]bool, error) {
	
	// Build efficient batch query
	query := `
		SELECT carriage_id, seat_number, COUNT(*) as booked_count
		FROM seat_reservations 
		WHERE service_id = $1 AND booking_date = $2 
		AND (carriage_id, seat_number) IN (`
	
	args := []interface{}{serviceID, date}
	placeholders := make([]string, len(seatRequests))
	
	for i, req := range seatRequests {
		placeholders[i] = fmt.Sprintf("($%d, $%d)", len(args)+1, len(args)+2)
		args = append(args, req.CarriageID, req.SeatNumber)
	}
	
	query += fmt.Sprintf("%s) GROUP BY carriage_id, seat_number", 
		fmt.Sprintf("%s", placeholders[0]))
	for i := 1; i < len(placeholders); i++ {
		query += fmt.Sprintf(", %s", placeholders[i])
	}
	
	rows, err := s.repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to check multiple seats: %w", err)
	}
	defer rows.Close()
	
	// Build result map
	bookedSeats := make(map[string]bool)
	for rows.Next() {
		var carriageID, seatNumber string
		var count int
		if err := rows.Scan(&carriageID, &seatNumber, &count); err != nil {
			return nil, fmt.Errorf("failed to scan seat result: %w", err)
		}
		key := fmt.Sprintf("%s:%s", carriageID, seatNumber)
		bookedSeats[key] = count > 0
	}
	
	// Build availability map
	availability := make(map[string]bool)
	for _, req := range seatRequests {
		key := fmt.Sprintf("%s:%s", req.CarriageID, req.SeatNumber)
		availability[key] = !bookedSeats[key] // Available if not booked
	}
	
	return availability, nil
}
