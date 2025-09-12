# Ticketing System

A simple train ticketing system that manages seat reservations and bookings.

## What it does

- Book train tickets with specific seats
- Prevent double-booking of seats
- Track passengers and their journeys
- Query passenger information

## Quick Start

### Run locally

```bash
make start
```

### Run with Docker

```bash
make docker
```

### Run tests

```bash
make test
```

### Build the app

```bash
make build
```

## Available Commands

```bash
make help    # Show all commands
make run     # Run the app
make build   # Build the app
make test    # Run tests
make docker  # Build and run with Docker
make clean   # Clean up files
```

## How it works

The system has trains that run on routes between stations. Each train has carriages with seats that can be booked by passengers.

### Example Routes

- Paris → Calais → Amsterdam
- Paris → Calais → London  
- Amsterdam → Utrecht → Hannover → Berlin

### Booking a ticket

```go
booking, err := rs.MakeReservation(ReservationRequest{
    ServiceID: "5160",
    Origin: "Paris",
    Destination: "Amsterdam", 
    Passengers: []Passenger{{Name: "John Doe"}},
    SeatRequests: []SeatRequest{{CarriageID: "A", SeatNumber: "A11"}},
    Date: time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC),
})
```

## Docker

The app is containerized and ready to run anywhere:

```bash
# Build and run in one command
make docker

# Or step by step
make docker-build
make docker-run
```

## Files

### Main Package

- `main.go` - Application entry point
- `interfaces.go` - REST interfaces (Response, HTTPClient)
- `interfaces_test.go` - Tests for interfaces

### Domain Package (`pkg/domain/`)

- `models.go` - Core data structures (Station, Route, Service, Booking, etc.)
- `models_test.go` - Tests for domain models

### Reservation Package (`pkg/reservation/`)

- `system.go` - Booking logic and reservation system
- `system_test.go` - Tests for reservation system

### Test Data Package (`pkg/testdata/`)

- `setup.go` - Sample routes, trains, and test data setup

### Infrastructure

- `Dockerfile` - Docker configuration
- `Makefile` - Build commands
- `README.md` - This documentation
