package domain

import (
	"fmt"
	"time"
)

type Station struct {
	Name string
}

type Stop struct {
	Station   Station
	Distance  int 
	StopOrder int 
}

type Route struct {
	ID    string
	Name  string
	Stops []Stop
}

type ComfortZone string

const (
	FirstClass  ComfortZone = "first-class"
	SecondClass ComfortZone = "second-class"
)

type Seat struct {
	Number       string
	ComfortZone  ComfortZone
	CarriageID   string
}

type Carriage struct {
	ID    string
	Seats []Seat
}

type Service struct {
	ID        string
	Route     Route
	DateTime  time.Time
	Carriages []Carriage
}

type Passenger struct {
	Name string
}

type Ticket struct {
	Seat         Seat
	Origin       Station
	Destination  Station
	Service      Service
	Passenger    Passenger
}

type Booking struct {
	ID        string
	Passengers []Passenger
	Tickets   []Ticket
	CreatedAt time.Time
}

type ReservationRequest struct {
	ServiceID    string
	Origin       string
	Destination  string
	Passengers   []Passenger
	SeatRequests []SeatRequest
	Date         time.Time
}

type SeatRequest struct {
	CarriageID string
	SeatNumber string
}

func NewStation(name string) Station {
	return Station{Name: name}
}

func NewRoute(id, name string, stations []Station, distances []int) Route {
	if len(stations) != len(distances) {
		panic("number of stations must equal number of distances")
	}
	
	stops := make([]Stop, len(stations))
	for i, station := range stations {
		stops[i] = Stop{
			Station:   station,
			Distance:  distances[i],
			StopOrder: i,
		}
	}
	
	return Route{
		ID:    id,
		Name:  name,
		Stops: stops,
	}
}

func NewService(id string, route Route, dateTime time.Time, carriages []Carriage) Service {
	return Service{
		ID:        id,
		Route:     route,
		DateTime:  dateTime,
		Carriages: carriages,
	}
}

func NewBooking(id string, passengers []Passenger, tickets []Ticket) Booking {
	return Booking{
		ID:         id,
		Passengers: passengers,
		Tickets:    tickets,
		CreatedAt:  time.Now(),
	}
}

func (r Route) GetStationByName(name string) (Station, bool) {
	for _, stop := range r.Stops {
		if stop.Station.Name == name {
			return stop.Station, true
		}
	}
	return Station{}, false
}

func (r Route) GetStopIndex(stationName string) (int, bool) {
	for i, stop := range r.Stops {
		if stop.Station.Name == stationName {
			return i, true
		}
	}
	return -1, false
}

func (r Route) IsValidOriginDestination(origin, destination string) bool {
	originIndex, originFound := r.GetStopIndex(origin)
	destIndex, destFound := r.GetStopIndex(destination)
	
	if !originFound || !destFound {
		return false
	}
	
	return originIndex < destIndex
}

func (s Service) GetSeatByID(carriageID, seatNumber string) (Seat, bool) {
	for _, carriage := range s.Carriages {
		if carriage.ID == carriageID {
			for _, seat := range carriage.Seats {
				if seat.Number == seatNumber {
					return seat, true
				}
			}
		}
	}
	return Seat{}, false
}

func (b Booking) String() string {
	return fmt.Sprintf("Booking %s: %d passengers, %d tickets", b.ID, len(b.Passengers), len(b.Tickets))
}
