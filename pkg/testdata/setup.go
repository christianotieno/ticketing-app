package testdata

import (
	"ticketing-app/pkg/domain"
	"ticketing-app/pkg/reservation"
	"time"
)

func SetupTestData() *reservation.System {
	rs := reservation.NewSystem()
	
	paris := domain.NewStation("Paris")
	london := domain.NewStation("London")
	calais := domain.NewStation("Calais")
	dover := domain.NewStation("Dover")
	amsterdam := domain.NewStation("Amsterdam")
	antwerp := domain.NewStation("Antwerp")
	utrecht := domain.NewStation("Utrecht")
	berlin := domain.NewStation("Berlin")
	hannover := domain.NewStation("Hannover")
	
	routeParisLondon := domain.NewRoute("R001", "Paris-London", 
		[]domain.Station{paris, calais, dover, london},
		[]int{0, 300, 380, 450})
	
	routeParisAmsterdam := domain.NewRoute("R002", "Paris-Amsterdam",
		[]domain.Station{paris, calais, antwerp, amsterdam},
		[]int{0, 300, 420, 520})
	
	routeAmsterdamBerlin := domain.NewRoute("R003", "Amsterdam-Berlin",
		[]domain.Station{amsterdam, utrecht, hannover, berlin},
		[]int{0, 45, 280, 450})
	
	rs.AddRoute(routeParisLondon)
	rs.AddRoute(routeParisAmsterdam)
	rs.AddRoute(routeAmsterdamBerlin)
	
	carriages := createCarriages()
	
	service5160 := domain.NewService("5160", routeParisAmsterdam, 
		time.Date(2021, 4, 1, 8, 0, 0, 0, time.UTC), carriages)
	
	service5161 := domain.NewService("5161", routeParisAmsterdam,
		time.Date(2021, 12, 20, 8, 0, 0, 0, time.UTC), carriages)
	
	service5162 := domain.NewService("5162", routeParisAmsterdam,
		time.Date(2021, 4, 2, 10, 0, 0, 0, time.UTC), carriages)
	
	rs.AddService(service5160)
	rs.AddService(service5161)
	rs.AddService(service5162)
	
	return rs
}

func createCarriages() []domain.Carriage {
	carriages := []domain.Carriage{
		// Carriage A - First Class
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
				{Number: "A9", ComfortZone: domain.FirstClass, CarriageID: "A"},
				{Number: "A10", ComfortZone: domain.FirstClass, CarriageID: "A"},
				{Number: "A11", ComfortZone: domain.FirstClass, CarriageID: "A"},
				{Number: "A12", ComfortZone: domain.FirstClass, CarriageID: "A"},
			},
		},
		// Carriage H - Second Class
		{
			ID: "H",
			Seats: []domain.Seat{
				{Number: "H1", ComfortZone: domain.SecondClass, CarriageID: "H"},
				{Number: "H2", ComfortZone: domain.SecondClass, CarriageID: "H"},
				{Number: "H3", ComfortZone: domain.SecondClass, CarriageID: "H"},
				{Number: "H4", ComfortZone: domain.SecondClass, CarriageID: "H"},
				{Number: "H5", ComfortZone: domain.SecondClass, CarriageID: "H"},
				{Number: "H6", ComfortZone: domain.SecondClass, CarriageID: "H"},
				{Number: "H7", ComfortZone: domain.SecondClass, CarriageID: "H"},
				{Number: "H8", ComfortZone: domain.SecondClass, CarriageID: "H"},
				{Number: "H9", ComfortZone: domain.SecondClass, CarriageID: "H"},
				{Number: "H10", ComfortZone: domain.SecondClass, CarriageID: "H"},
			},
		},
		// Carriage N - First Class
		{
			ID: "N",
			Seats: []domain.Seat{
				{Number: "N1", ComfortZone: domain.FirstClass, CarriageID: "N"},
				{Number: "N2", ComfortZone: domain.FirstClass, CarriageID: "N"},
				{Number: "N3", ComfortZone: domain.FirstClass, CarriageID: "N"},
				{Number: "N4", ComfortZone: domain.FirstClass, CarriageID: "N"},
				{Number: "N5", ComfortZone: domain.FirstClass, CarriageID: "N"},
				{Number: "N6", ComfortZone: domain.FirstClass, CarriageID: "N"},
				{Number: "N7", ComfortZone: domain.FirstClass, CarriageID: "N"},
				{Number: "N8", ComfortZone: domain.FirstClass, CarriageID: "N"},
				{Number: "N9", ComfortZone: domain.FirstClass, CarriageID: "N"},
				{Number: "N10", ComfortZone: domain.FirstClass, CarriageID: "N"},
			},
		},
		// Carriage T - Second Class
		{
			ID: "T",
			Seats: []domain.Seat{
				{Number: "T1", ComfortZone: domain.SecondClass, CarriageID: "T"},
				{Number: "T2", ComfortZone: domain.SecondClass, CarriageID: "T"},
				{Number: "T3", ComfortZone: domain.SecondClass, CarriageID: "T"},
				{Number: "T4", ComfortZone: domain.SecondClass, CarriageID: "T"},
				{Number: "T5", ComfortZone: domain.SecondClass, CarriageID: "T"},
				{Number: "T6", ComfortZone: domain.SecondClass, CarriageID: "T"},
				{Number: "T7", ComfortZone: domain.SecondClass, CarriageID: "T"},
				{Number: "T8", ComfortZone: domain.SecondClass, CarriageID: "T"},
				{Number: "T9", ComfortZone: domain.SecondClass, CarriageID: "T"},
				{Number: "T10", ComfortZone: domain.SecondClass, CarriageID: "T"},
			},
		},
	}
	
	return carriages
}
