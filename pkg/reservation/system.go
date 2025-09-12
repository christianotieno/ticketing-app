package reservation

import (
	"fmt"
	"ticketing-app/pkg/domain"
	"time"
)

type ReservationError struct {
	Message string
	Code    string
}

func (e ReservationError) Error() string {
	return e.Message
}

type System struct {
	bookings      map[string]domain.Booking
	services      map[string]domain.Service
	routes        map[string]domain.Route
	nextBookingID int
}

func NewSystem() *System {
	return &System{
		bookings:      make(map[string]domain.Booking),
		services:      make(map[string]domain.Service),
		routes:        make(map[string]domain.Route),
		nextBookingID: 1,
	}
}

func (rs *System) AddRoute(route domain.Route) {
	rs.routes[route.ID] = route
}

func (rs *System) AddService(service domain.Service) {
	rs.services[service.ID] = service
}

func (rs *System) MakeReservation(req domain.ReservationRequest) (*domain.Booking, error) {
	service, exists := rs.services[req.ServiceID]
	if !exists {
		return nil, ReservationError{
			Message: fmt.Sprintf("Service %s not found", req.ServiceID),
			Code:    "SERVICE_NOT_FOUND",
		}
	}

	if !service.Route.IsValidOriginDestination(req.Origin, req.Destination) {
		return nil, ReservationError{
			Message: fmt.Sprintf("Invalid route from %s to %s for service %s", req.Origin, req.Destination, req.ServiceID),
			Code:    "INVALID_ROUTE",
		}
	}

	if len(req.Passengers) != len(req.SeatRequests) {
		return nil, ReservationError{
			Message: "Number of passengers must match number of seat requests",
			Code:    "PASSENGER_SEAT_MISMATCH",
		}
	}

	originStation, _ := service.Route.GetStationByName(req.Origin)
	destStation, _ := service.Route.GetStationByName(req.Destination)
	
	tickets := make([]domain.Ticket, len(req.Passengers))
	
	for i, seatReq := range req.SeatRequests {
		seat, exists := service.GetSeatByID(seatReq.CarriageID, seatReq.SeatNumber)
		if !exists {
			return nil, ReservationError{
				Message: fmt.Sprintf("Seat %s in carriage %s not found in service %s", seatReq.SeatNumber, seatReq.CarriageID, req.ServiceID),
				Code:    "SEAT_NOT_FOUND",
			}
		}

		if rs.isSeatBooked(req.ServiceID, seatReq.CarriageID, seatReq.SeatNumber, req.Date) {
			return nil, ReservationError{
				Message: fmt.Sprintf("Seat %s in carriage %s is already booked for service %s", seatReq.SeatNumber, seatReq.CarriageID, req.ServiceID),
				Code:    "SEAT_ALREADY_BOOKED",
			}
		}

		tickets[i] = domain.Ticket{
			Seat:        seat,
			Origin:      originStation,
			Destination: destStation,
			Service:     service,
			Passenger:   req.Passengers[i],
		}
	}

	bookingID := fmt.Sprintf("B%04d", rs.nextBookingID)
	rs.nextBookingID++
	
	booking := domain.NewBooking(bookingID, req.Passengers, tickets)
	rs.bookings[bookingID] = booking

	return &booking, nil
}

func (rs *System) isSeatBooked(serviceID, carriageID, seatNumber string, date time.Time) bool {
	for _, booking := range rs.bookings {
		for _, ticket := range booking.Tickets {
			if ticket.Service.ID == serviceID &&
				ticket.Seat.CarriageID == carriageID &&
				ticket.Seat.Number == seatNumber &&
				rs.isSameDate(ticket.Service.DateTime, date) {
				return true
			}
		}
	}
	return false
}

func (rs *System) isSameDate(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func (rs *System) GetBooking(bookingID string) (*domain.Booking, bool) {
	booking, exists := rs.bookings[bookingID]
	return &booking, exists
}

func (rs *System) GetAllBookings() []domain.Booking {
	bookings := make([]domain.Booking, 0, len(rs.bookings))
	for _, booking := range rs.bookings {
		bookings = append(bookings, booking)
	}
	return bookings
}

func (rs *System) GetPassengersBoardingAt(serviceID, stationName string, date time.Time) []domain.Passenger {
	var passengers []domain.Passenger
	
	for _, booking := range rs.bookings {
		for _, ticket := range booking.Tickets {
			if ticket.Service.ID == serviceID &&
				ticket.Origin.Name == stationName &&
				rs.isSameDate(ticket.Service.DateTime, date) {
				passengers = append(passengers, ticket.Passenger)
			}
		}
	}
	
	return passengers
}

func (rs *System) GetPassengersAlightingAt(serviceID, stationName string, date time.Time) []domain.Passenger {
	var passengers []domain.Passenger
	
	for _, booking := range rs.bookings {
		for _, ticket := range booking.Tickets {
			if ticket.Service.ID == serviceID &&
				ticket.Destination.Name == stationName &&
				rs.isSameDate(ticket.Service.DateTime, date) {
				passengers = append(passengers, ticket.Passenger)
			}
		}
	}
	
	return passengers
}

func (rs *System) GetPassengersBetweenStations(serviceID, station1, station2 string, date time.Time) []domain.Passenger {
	var passengers []domain.Passenger
	
	service, exists := rs.services[serviceID]
	if !exists {
		return passengers
	}
	
	stop1Index, found1 := service.Route.GetStopIndex(station1)
	stop2Index, found2 := service.Route.GetStopIndex(station2)
	
	if !found1 || !found2 {
		return passengers
	}
	
	if stop1Index >= stop2Index {
		stop1Index, stop2Index = stop2Index, stop1Index
	}
	
	for _, booking := range rs.bookings {
		for _, ticket := range booking.Tickets {
			if ticket.Service.ID == serviceID && rs.isSameDate(ticket.Service.DateTime, date) {
				originIndex, _ := service.Route.GetStopIndex(ticket.Origin.Name)
				destIndex, _ := service.Route.GetStopIndex(ticket.Destination.Name)
				
				if originIndex <= stop1Index && destIndex >= stop2Index {
					passengers = append(passengers, ticket.Passenger)
				}
			}
		}
	}
	
	return passengers
}

func (rs *System) GetPassengerOnSeat(serviceID, carriageID, seatNumber string, date time.Time) (*domain.Passenger, bool) {
	for _, booking := range rs.bookings {
		for _, ticket := range booking.Tickets {
			if ticket.Service.ID == serviceID &&
				ticket.Seat.CarriageID == carriageID &&
				ticket.Seat.Number == seatNumber &&
				rs.isSameDate(ticket.Service.DateTime, date) {
				return &ticket.Passenger, true
			}
		}
	}
	return nil, false
}
