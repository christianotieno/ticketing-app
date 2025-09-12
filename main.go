package main

import (
	"fmt"
	"ticketing-app/pkg/domain"
	"ticketing-app/pkg/reservation"
	"ticketing-app/pkg/testdata"
	"time"
)

func main() {
	fmt.Println("=== Ticketing System Demo ===")
	
	rs := testdata.SetupTestData()
	
	runTestScenarios(rs)
	
	runConductorQueries(rs)
}

func runTestScenarios(rs *reservation.System) {
	fmt.Println("\n=== Test Scenarios ===")
	
	fmt.Println("\nScenario 1: Booking 2 first-class seats (A11 & A12) from Paris to Amsterdam")
	booking1, err := rs.MakeReservation(domain.ReservationRequest{
		ServiceID: "5160",
		Origin:    "Paris",
		Destination: "Amsterdam",
		Passengers: []domain.Passenger{
			{Name: "John Doe"},
			{Name: "Jane Smith"},
		},
		SeatRequests: []domain.SeatRequest{
			{CarriageID: "A", SeatNumber: "A11"},
			{CarriageID: "A", SeatNumber: "A12"},
		},
		Date: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
	})
	
	if err != nil {
		fmt.Printf("Booking failed: %v\n", err)
	} else {
		fmt.Printf("Booking successful: %s\n", booking1.String())
		fmt.Printf("   Passengers: %s, %s\n", booking1.Passengers[0].Name, booking1.Passengers[1].Name)
		fmt.Printf("   Seats: %s, %s\n", booking1.Tickets[0].Seat.Number, booking1.Tickets[1].Seat.Number)
	}
	
	fmt.Println("\nScenario 2: Attempting duplicate booking (should fail)")
	_, err = rs.MakeReservation(domain.ReservationRequest{
		ServiceID: "5160",
		Origin:    "Paris",
		Destination: "Amsterdam",
		Passengers: []domain.Passenger{
			{Name: "Bob Wilson"},
			{Name: "Alice Brown"},
		},
		SeatRequests: []domain.SeatRequest{
			{CarriageID: "A", SeatNumber: "A11"},
			{CarriageID: "A", SeatNumber: "A12"},
		},
		Date: time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC),
	})
	
	if err != nil {
		fmt.Printf("Expected failure: %v\n", err)
	} else {
		fmt.Println("Booking should have failed but succeeded")
	}
	
	fmt.Println("\nScenario 3: Mixed class booking from London to Amsterdam (via Paris)")
	booking3, err := rs.MakeReservation(domain.ReservationRequest{
		ServiceID: "5162",
		Origin:    "Paris",
		Destination: "Amsterdam",
		Passengers: []domain.Passenger{
			{Name: "Charlie Davis"},
			{Name: "Diana Prince"},
		},
		SeatRequests: []domain.SeatRequest{
			{CarriageID: "H", SeatNumber: "H1"},
			{CarriageID: "N", SeatNumber: "N5"},
		},
		Date: time.Date(2025, 4, 2, 0, 0, 0, 0, time.UTC),
	})
	
	if err != nil {
		fmt.Printf("Booking failed: %v\n", err)
	} else {
		fmt.Printf("Booking successful: %s\n", booking3.String())
		fmt.Printf("   Passengers: %s, %s\n", booking3.Passengers[0].Name, booking3.Passengers[1].Name)
		fmt.Printf("   Seats: %s (%s), %s (%s)\n", 
			booking3.Tickets[0].Seat.Number, booking3.Tickets[0].Seat.ComfortZone,
			booking3.Tickets[1].Seat.Number, booking3.Tickets[1].Seat.ComfortZone)
	}
	
	fmt.Println("\nScenario 4: Attempting duplicate mixed class booking (should fail)")
	_, err = rs.MakeReservation(domain.ReservationRequest{
		ServiceID: "5162",
		Origin:    "Paris",
		Destination: "Amsterdam",
		Passengers: []domain.Passenger{
			{Name: "Eve Johnson"},
			{Name: "Frank Miller"},
		},
		SeatRequests: []domain.SeatRequest{
			{CarriageID: "H", SeatNumber: "H1"},
			{CarriageID: "N", SeatNumber: "N5"},
		},
		Date: time.Date(2021, 4, 2, 0, 0, 0, 0, time.UTC),
	})
	
	if err != nil {
		fmt.Printf("Expected failure: %v\n", err)
	} else {
		fmt.Println("Booking should have failed but succeeded")
	}
	
	fmt.Println("\nSetting up additional bookings for conductor queries...")
	
	_, err = rs.MakeReservation(domain.ReservationRequest{
		ServiceID: "5161",
		Origin:    "Paris",
		Destination: "Amsterdam",
		Passengers: []domain.Passenger{{Name: "Conductor Test Passenger"}},
		SeatRequests: []domain.SeatRequest{{CarriageID: "A", SeatNumber: "A11"}},
		Date: time.Date(2021, 12, 20, 0, 0, 0, 0, time.UTC),
	})
	
	if err != nil {
		fmt.Printf("Warning: Could not create test booking: %v\n", err)
	} else {
		fmt.Println("Test booking created for conductor queries")
	}
}

func runConductorQueries(rs *reservation.System) {
	fmt.Println("\n=== Conductor Queries ===")
	
	fmt.Println("\n1. Passengers boarding at London:")
	passengers := rs.GetPassengersBoardingAt("5160", "London", time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC))
	fmt.Printf("   Passengers boarding at London on service 5160: %d\n", len(passengers))
	
	passengers = rs.GetPassengersBoardingAt("5160", "Paris", time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC))
	fmt.Printf("   Passengers boarding at Paris on service 5160: %d\n", len(passengers))
	for _, p := range passengers {
		fmt.Printf("     - %s\n", p.Name)
	}
	
	fmt.Println("\n2. Passengers leaving at Paris:")
	passengers = rs.GetPassengersAlightingAt("5160", "Paris", time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC))
	fmt.Printf("   Passengers leaving at Paris on service 5160: %d\n", len(passengers))
	
	passengers = rs.GetPassengersAlightingAt("5160", "Amsterdam", time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC))
	fmt.Printf("   Passengers leaving at Amsterdam on service 5160: %d\n", len(passengers))
	for _, p := range passengers {
		fmt.Printf("     - %s\n", p.Name)
	}
	
	fmt.Println("\n3. Passengers traveling between Calais and Paris:")
	passengers = rs.GetPassengersBetweenStations("5160", "Calais", "Amsterdam", time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC))
	fmt.Printf("   Passengers traveling between Calais and Amsterdam on service 5160: %d\n", len(passengers))
	for _, p := range passengers {
		fmt.Printf("     - %s\n", p.Name)
	}
	
	fmt.Println("\n4. Passenger on seat A11 in service 5161 on December 20th:")
	passenger, found := rs.GetPassengerOnSeat("5161", "A", "A11", time.Date(2021, 12, 20, 0, 0, 0, 0, time.UTC))
	if found {
		fmt.Printf("   Passenger on seat A11: %s\n", passenger.Name)
	} else {
		fmt.Printf("   No passenger found on seat A11\n")
	}
}