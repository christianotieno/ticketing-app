package domain

import (
	"testing"
	"time"
)

func TestRoute_IsValidOriginDestination(t *testing.T) {
	route := NewRoute("R001", "Test Route",
		[]Station{
			NewStation("A"),
			NewStation("B"),
			NewStation("C"),
		},
		[]int{0, 100, 200})
	
	tests := []struct {
		origin      string
		destination string
		expected    bool
	}{
		{"A", "B", true},
		{"A", "C", true},
		{"B", "C", true},
		{"B", "A", false},
		{"C", "A", false},
		{"C", "B", false},
		{"A", "D", false},
		{"D", "A", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.origin+"_to_"+tt.destination, func(t *testing.T) {
			result := route.IsValidOriginDestination(tt.origin, tt.destination)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestService_GetSeatByID(t *testing.T) {
	route := NewRoute("R001", "Test Route", []Station{NewStation("A")}, []int{0})
	carriages := []Carriage{
		{
			ID: "A",
			Seats: []Seat{
				{Number: "A1", ComfortZone: FirstClass, CarriageID: "A"},
				{Number: "A2", ComfortZone: FirstClass, CarriageID: "A"},
			},
		},
	}
	service := NewService("S001", route, time.Now(), carriages)
	
	// Test finding existing seat
	seat, found := service.GetSeatByID("A", "A1")
	if !found {
		t.Errorf("Expected to find seat A1")
	}
	if found && seat.Number != "A1" {
		t.Errorf("Expected seat number A1, got %s", seat.Number)
	}
	
	// Test finding non-existing seat
	_, found = service.GetSeatByID("A", "A99")
	if found {
		t.Errorf("Expected not to find seat A99")
	}
	
	// Test finding seat in non-existing carriage
	_, found = service.GetSeatByID("Z", "A1")
	if found {
		t.Errorf("Expected not to find seat A1 in carriage Z")
	}
}

func TestBooking_String(t *testing.T) {
	booking := NewBooking("B001", 
		[]Passenger{{Name: "John"}, {Name: "Jane"}},
		[]Ticket{
			{Passenger: Passenger{Name: "John"}},
			{Passenger: Passenger{Name: "Jane"}},
		})
	
	result := booking.String()
	expected := "Booking B001: 2 passengers, 2 tickets"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestNewRoute(t *testing.T) {
	stations := []Station{NewStation("A"), NewStation("B")}
	distances := []int{0, 100}
	
	route := NewRoute("R001", "Test Route", stations, distances)
	
	if route.ID != "R001" {
		t.Errorf("Expected route ID R001, got %s", route.ID)
	}
	if route.Name != "Test Route" {
		t.Errorf("Expected route name 'Test Route', got %s", route.Name)
	}
	if len(route.Stops) != 2 {
		t.Errorf("Expected 2 stops, got %d", len(route.Stops))
	}
	if route.Stops[0].Distance != 0 {
		t.Errorf("Expected first stop distance 0, got %d", route.Stops[0].Distance)
	}
	if route.Stops[1].Distance != 100 {
		t.Errorf("Expected second stop distance 100, got %d", route.Stops[1].Distance)
	}
}

func TestNewRoute_PanicOnMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for mismatched stations and distances")
		}
	}()
	
	stations := []Station{NewStation("A"), NewStation("B")}
	distances := []int{0} // Only one distance for two stations
	
	NewRoute("R001", "Test Route", stations, distances)
}
