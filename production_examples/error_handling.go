package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Standard Go error handling with error wrapping and sentinel errors
var (
	// Sentinel errors for common conditions
	ErrSeatUnavailable     = errors.New("seat is not available")
	ErrServiceNotFound     = errors.New("service not found")
	ErrInvalidRoute        = errors.New("invalid route")
	ErrSeatNotFound        = errors.New("seat not found")
	ErrPassengerSeatMismatch = errors.New("passenger count does not match seat count")
	ErrBookingNotFound     = errors.New("booking not found")
	ErrInvalidTimezone     = errors.New("invalid timezone")
	ErrBookingTimeExpired  = errors.New("booking time has expired")
)

// Error types for structured error handling
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %s: %s (value: %v)", e.Field, e.Message, e.Value)
}

type BusinessRuleError struct {
	Rule    string
	Message string
	Context map[string]interface{}
}

func (e BusinessRuleError) Error() string {
	return fmt.Sprintf("business rule violation [%s]: %s", e.Rule, e.Message)
}

// Production error handling with proper wrapping
type ProductionBookingService struct {
	repo *BookingRepository
}

// MakeReservation with comprehensive error handling
func (s *ProductionBookingService) MakeReservation(ctx context.Context, 
	req BookingRequest) (*Booking, error) {
	
	// Input validation with structured errors
	if err := s.validateBookingRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	// Check service existence
	service, err := s.getService(ctx, req.ServiceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("service %s: %w", req.ServiceID, ErrServiceNotFound)
		}
		return nil, fmt.Errorf("failed to get service: %w", err)
	}
	
	// Validate route
	if !s.isValidRoute(service, req.Origin, req.Destination) {
		return nil, fmt.Errorf("route from %s to %s for service %s: %w", 
			req.Origin, req.Destination, req.ServiceID, ErrInvalidRoute)
	}
	
	// Check seat availability with proper error context
	available, err := s.checkSeatAvailability(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to check seat availability: %w", err)
	}
	
	if !available {
		return nil, fmt.Errorf("requested seats: %w", ErrSeatUnavailable)
	}
	
	// Create booking with transaction
	booking, err := s.createBookingTransaction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}
	
	return booking, nil
}

// Input validation with detailed error reporting
func (s *ProductionBookingService) validateBookingRequest(req BookingRequest) error {
	var validationErrors []ValidationError
	
	if req.ServiceID == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "ServiceID",
			Value:   req.ServiceID,
			Message: "service ID is required",
		})
	}
	
	if req.Origin == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "Origin",
			Value:   req.Origin,
			Message: "origin station is required",
		})
	}
	
	if req.Destination == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "Destination",
			Value:   req.Destination,
			Message: "destination station is required",
		})
	}
	
	if len(req.Passengers) == 0 {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "Passengers",
			Value:   len(req.Passengers),
			Message: "at least one passenger is required",
		})
	}
	
	if len(req.Passengers) != len(req.SeatRequests) {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "Passengers/SeatRequests",
			Value:   fmt.Sprintf("passengers:%d, seats:%d", len(req.Passengers), len(req.SeatRequests)),
			Message: "passenger count must match seat request count",
		})
	}
	
	// Validate timezone
	if req.Timezone != "" {
		if _, err := time.LoadLocation(req.Timezone); err != nil {
			validationErrors = append(validationErrors, ValidationError{
				Field:   "Timezone",
				Value:   req.Timezone,
				Message: fmt.Sprintf("invalid timezone: %v", err),
			})
		}
	}
	
	if len(validationErrors) > 0 {
		// Return the first validation error with context about total errors
		return fmt.Errorf("validation failed (%d errors): %w", 
			len(validationErrors), validationErrors[0])
	}
	
	return nil
}

// Check seat availability with detailed error context
func (s *ProductionBookingService) checkSeatAvailability(ctx context.Context, 
	req BookingRequest) (bool, error) {
	
	for i, seatReq := range req.SeatRequests {
		// Check if seat exists
		exists, err := s.seatExists(ctx, req.ServiceID, seatReq.CarriageID, seatReq.SeatNumber)
		if err != nil {
			return false, fmt.Errorf("failed to check seat existence: %w", err)
		}
		
		if !exists {
			return false, fmt.Errorf("seat %s in carriage %s: %w", 
				seatReq.SeatNumber, seatReq.CarriageID, ErrSeatNotFound)
		}
		
		// Check if seat is available
		available, err := s.isSeatAvailable(ctx, req.ServiceID, seatReq.CarriageID, 
			seatReq.SeatNumber, req.DepartureTime)
		if err != nil {
			return false, fmt.Errorf("failed to check availability for seat %s: %w", 
				seatReq.SeatNumber, err)
		}
		
		if !available {
			return false, fmt.Errorf("seat %s in carriage %s for passenger %s: %w", 
				seatReq.SeatNumber, seatReq.CarriageID, req.Passengers[i].Name, ErrSeatUnavailable)
		}
	}
	
	return true, nil
}

// Business rule validation with context
func (s *ProductionBookingService) validateBusinessRules(ctx context.Context, 
	req BookingRequest) error {
	
	// Check advance booking limits
	if err := s.checkAdvanceBookingLimit(req.DepartureTime); err != nil {
		return fmt.Errorf("advance booking rule violation: %w", err)
	}
	
	// Check booking time windows
	if err := s.checkBookingTimeWindow(req.DepartureTime); err != nil {
		return fmt.Errorf("booking time window violation: %w", err)
	}
	
	// Check passenger limits
	if err := s.checkPassengerLimits(req.Passengers); err != nil {
		return fmt.Errorf("passenger limit violation: %w", err)
	}
	
	return nil
}

// Error handling with retry logic for transient failures
func (s *ProductionBookingService) createBookingWithRetry(ctx context.Context, 
	req BookingRequest, maxRetries int) (*Booking, error) {
	
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		booking, err := s.createBookingTransaction(ctx, req)
		if err == nil {
			return booking, nil
		}
		
		// Check if error is retryable
		if !s.isRetryableError(err) {
			return nil, fmt.Errorf("non-retryable error: %w", err)
		}
		
		lastErr = err
		
		// Exponential backoff
		if attempt < maxRetries {
			backoff := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
			case <-time.After(backoff):
				// Continue to next attempt
			}
		}
	}
	
	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries+1, lastErr)
}

// Helper function to determine if error is retryable
func (s *ProductionBookingService) isRetryableError(err error) bool {
	// Database connection errors are retryable
	if errors.Is(err, sql.ErrConnDone) || errors.Is(err, sql.ErrTxDone) {
		return true
	}
	
	// Timeout errors are retryable
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	
	// Business rule violations are not retryable
	if errors.Is(err, ErrSeatUnavailable) || errors.Is(err, ErrInvalidRoute) {
		return false
	}
	
	// Default to non-retryable for safety
	return false
}
