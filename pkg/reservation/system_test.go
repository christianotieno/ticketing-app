package reservation

import (
	"testing"
	"ticketing-app/pkg/domain"
	"time"
)

func setupTestSystem() *System {
	rs := NewSystem()
	
	paris := domain.NewStation("Paris")
	calais := domain.NewStation("Calais")
	amsterdam := domain.NewStation("Amsterdam")
	
	route := domain.NewRoute("R002", "Paris-Amsterdam",
		[]domain.Station{paris, calais, amsterdam},
		[]int{0, 300, 520})
	
	carriages := []domain.Carriage{
		{
			ID: "A",
			Seats: []domain.Seat{
				{Number: "A1", ComfortZone: domain.FirstClass, CarriageID: "A"},
				{Number: "A2", ComfortZone: domain.FirstClass, CarriageID: "A"},
				{Number: "A3", ComfortZone: domain.FirstClass, CarriageID: "A"},
				{Number: "A4", ComfortZone: domain.FirstClass, CarriageID: "A"},
				{Number: "A5", ComfortZone: domain.FirstClass, CarriageID: "A"},
				{Number: "A6", ComfortZone: domain.FirstClass, CarriageID: "A"},
				{Number: "A7", ComfortZone: domain.FirstClass, CarriageID: "A"},
				{Number: "A8", ComfortZone: domain.FirstClass, CarriageID: "A"},
			},
		},
	}
	
	service := domain.NewService("5160", route, 
		time.Date(2021, 4, 1, 8, 0, 0, 0, time.UTC), carriages)
	
	rs.AddRoute(route)
	rs.AddService(service)
	
	return rs
}

func TestSystem_MakeReservation(t *testing.T) {
	rs := setupTestSystem()
	
	tests := []struct {
		name    string
		request domain.ReservationRequest
		wantErr bool
		errCode string
	}{
		{
			name: "Valid first booking",
			request: domain.ReservationRequest{
				ServiceID: "5160",
				Origin:    "Paris",
				Destination: "Amsterdam",
				Passengers: []domain.Passenger{{Name: "John Doe"}},
				SeatRequests: []domain.SeatRequest{{CarriageID: "A", SeatNumber: "A1"}},
				Date: time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "Duplicate booking should fail",
			request: domain.ReservationRequest{
				ServiceID: "5160",
				Origin:    "Paris",
				Destination: "Amsterdam",
				Passengers: []domain.Passenger{{Name: "Jane Smith"}},
				SeatRequests: []domain.SeatRequest{{CarriageID: "A", SeatNumber: "A1"}},
				Date: time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: true,
			errCode: "SEAT_ALREADY_BOOKED",
		},
		{
			name: "Invalid service ID",
			request: domain.ReservationRequest{
				ServiceID: "9999",
				Origin:    "Paris",
				Destination: "Amsterdam",
				Passengers: []domain.Passenger{{Name: "Bob Wilson"}},
				SeatRequests: []domain.SeatRequest{{CarriageID: "A", SeatNumber: "A2"}},
				Date: time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: true,
			errCode: "SERVICE_NOT_FOUND",
		},
		{
			name: "Invalid route",
			request: domain.ReservationRequest{
				ServiceID: "5160",
				Origin:    "Amsterdam",
				Destination: "Paris", // Reverse direction
				Passengers: []domain.Passenger{{Name: "Alice Brown"}},
				SeatRequests: []domain.SeatRequest{{CarriageID: "A", SeatNumber: "A3"}},
				Date: time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: true,
			errCode: "INVALID_ROUTE",
		},
		{
			name: "Seat not found",
			request: domain.ReservationRequest{
				ServiceID: "5160",
				Origin:    "Paris",
				Destination: "Amsterdam",
				Passengers: []domain.Passenger{{Name: "Charlie Davis"}},
				SeatRequests: []domain.SeatRequest{{CarriageID: "Z", SeatNumber: "Z1"}},
				Date: time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: true,
			errCode: "SEAT_NOT_FOUND",
		},
		{
			name: "Passenger seat count mismatch",
			request: domain.ReservationRequest{
				ServiceID: "5160",
				Origin:    "Paris",
				Destination: "Amsterdam",
				Passengers: []domain.Passenger{{Name: "Diana Prince"}, {Name: "Eve Johnson"}},
				SeatRequests: []domain.SeatRequest{{CarriageID: "A", SeatNumber: "A4"}},
				Date: time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: true,
			errCode: "PASSENGER_SEAT_MISMATCH",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			booking, err := rs.MakeReservation(tt.request)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if reservationErr, ok := err.(ReservationError); ok {
					if reservationErr.Code != tt.errCode {
						t.Errorf("Expected error code %s, got %s", tt.errCode, reservationErr.Code)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
				if booking == nil {
					t.Errorf("Expected booking but got nil")
					return
				}
				if len(booking.Tickets) != len(tt.request.Passengers) {
					t.Errorf("Expected %d tickets, got %d", len(tt.request.Passengers), len(booking.Tickets))
				}
			}
		})
	}
}

func TestSystem_GetPassengersBoardingAt(t *testing.T) {
	rs := setupTestSystem()
	
	_, err := rs.MakeReservation(domain.ReservationRequest{
		ServiceID: "5160",
		Origin:    "Paris",
		Destination: "Amsterdam",
		Passengers: []domain.Passenger{{Name: "Test Passenger"}},
		SeatRequests: []domain.SeatRequest{{CarriageID: "A", SeatNumber: "A5"}},
		Date: time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Failed to create test booking: %v", err)
	}
	
	passengers := rs.GetPassengersBoardingAt("5160", "Paris", time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC))
	if len(passengers) != 1 {
		t.Errorf("Expected 1 passenger boarding at Paris, got %d", len(passengers))
	}
	if len(passengers) > 0 && passengers[0].Name != "Test Passenger" {
		t.Errorf("Expected passenger 'Test Passenger', got '%s'", passengers[0].Name)
	}
	
	passengers = rs.GetPassengersBoardingAt("5160", "Amsterdam", time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC))
	if len(passengers) != 0 {
		t.Errorf("Expected 0 passengers boarding at Amsterdam, got %d", len(passengers))
	}
}

func TestSystem_GetPassengersAlightingAt(t *testing.T) {
	rs := setupTestSystem()
	
	_, err := rs.MakeReservation(domain.ReservationRequest{
		ServiceID: "5160",
		Origin:    "Paris",
		Destination: "Amsterdam",
		Passengers: []domain.Passenger{{Name: "Test Passenger"}},
		SeatRequests: []domain.SeatRequest{{CarriageID: "A", SeatNumber: "A6"}},
		Date: time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Failed to create test booking: %v", err)
	}
	
	passengers := rs.GetPassengersAlightingAt("5160", "Amsterdam", time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC))
	if len(passengers) != 1 {
		t.Errorf("Expected 1 passenger alighting at Amsterdam, got %d", len(passengers))
	}
	if len(passengers) > 0 && passengers[0].Name != "Test Passenger" {
		t.Errorf("Expected passenger 'Test Passenger', got '%s'", passengers[0].Name)
	}
	
	passengers = rs.GetPassengersAlightingAt("5160", "Paris", time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC))
	if len(passengers) != 0 {
		t.Errorf("Expected 0 passengers alighting at Paris, got %d", len(passengers))
	}
}

func TestSystem_GetPassengersBetweenStations(t *testing.T) {
	rs := setupTestSystem()
	
	_, err := rs.MakeReservation(domain.ReservationRequest{
		ServiceID: "5160",
		Origin:    "Paris",
		Destination: "Amsterdam",
		Passengers: []domain.Passenger{{Name: "Test Passenger"}},
		SeatRequests: []domain.SeatRequest{{CarriageID: "A", SeatNumber: "A7"}},
		Date: time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Failed to create test booking: %v", err)
	}
	
	passengers := rs.GetPassengersBetweenStations("5160", "Calais", "Amsterdam", time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC))
	if len(passengers) != 1 {
		t.Errorf("Expected 1 passenger between Calais and Amsterdam, got %d", len(passengers))
	}
	
	passengers = rs.GetPassengersBetweenStations("5160", "Paris", "Calais", time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC))
	if len(passengers) != 1 {
		t.Errorf("Expected 1 passenger between Paris and Calais, got %d", len(passengers))
	}
}

func TestSystem_GetPassengerOnSeat(t *testing.T) {
	rs := setupTestSystem()
	
	_, err := rs.MakeReservation(domain.ReservationRequest{
		ServiceID: "5160",
		Origin:    "Paris",
		Destination: "Amsterdam",
		Passengers: []domain.Passenger{{Name: "Test Passenger"}},
		SeatRequests: []domain.SeatRequest{{CarriageID: "A", SeatNumber: "A8"}},
		Date: time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Failed to create test booking: %v", err)
	}
	
	passenger, found := rs.GetPassengerOnSeat("5160", "A", "A8", time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC))
	if !found {
		t.Errorf("Expected to find passenger on seat A8")
	}
	if found && passenger.Name != "Test Passenger" {
		t.Errorf("Expected passenger 'Test Passenger', got '%s'", passenger.Name)
	}
	
	_, found = rs.GetPassengerOnSeat("5160", "A", "A9", time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC))
	if found {
		t.Errorf("Expected not to find passenger on empty seat A9")
	}
}
