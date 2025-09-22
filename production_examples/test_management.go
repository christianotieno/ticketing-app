package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// Production test data management system
type TestDataManager struct {
	db     *sql.DB
	seeder *TestDataSeeder
	cleaner *TestDataCleaner
}

type TestScenario struct {
	Name        string
	Setup       func(*TestDataManager) error
	Cleanup     func(*TestDataManager) error
	Data        map[string]interface{}
}

// Factory pattern for test data creation
type TestDataFactory struct {
	rand *rand.Rand
}

func NewTestDataFactory(seed int64) *TestDataFactory {
	return &TestDataFactory{
		rand: rand.New(rand.NewSource(seed)),
	}
}

// Generate realistic test data
func (f *TestDataFactory) CreateTestRoute() Route {
	routes := []Route{
		{ID: "R001", Name: "Paris-London", Stops: []Stop{
			{Station: Station{Name: "Paris"}, Distance: 0, StopOrder: 0},
			{Station: Station{Name: "Calais"}, Distance: 300, StopOrder: 1},
			{Station: Station{Name: "London"}, Distance: 450, StopOrder: 2},
		}},
		{ID: "R002", Name: "Paris-Amsterdam", Stops: []Stop{
			{Station: Station{Name: "Paris"}, Distance: 0, StopOrder: 0},
			{Station: Station{Name: "Calais"}, Distance: 300, StopOrder: 1},
			{Station: Station{Name: "Amsterdam"}, Distance: 520, StopOrder: 2},
		}},
	}
	return routes[f.rand.Intn(len(routes))]
}

func (f *TestDataFactory) CreateTestService(route Route) Service {
	serviceIDs := []string{"5160", "5161", "5162", "5163", "5164"}
	dates := []time.Time{
		time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
		time.Date(2024, 1, 16, 9, 15, 0, 0, time.UTC),
	}
	
	return Service{
		ID:       serviceIDs[f.rand.Intn(len(serviceIDs))],
		Route:    route,
		DateTime: dates[f.rand.Intn(len(dates))],
		Carriages: f.createTestCarriages(),
	}
}

func (f *TestDataFactory) CreateTestPassenger() Passenger {
	firstNames := []string{"John", "Jane", "Bob", "Alice", "Charlie", "Diana", "Eve", "Frank"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis"}
	
	return Passenger{
		Name: fmt.Sprintf("%s %s", 
			firstNames[f.rand.Intn(len(firstNames))],
			lastNames[f.rand.Intn(len(lastNames))]),
	}
}

func (f *TestDataFactory) createTestCarriages() []Carriage {
	return []Carriage{
		{
			ID: "A",
			Seats: f.createSeatsForCarriage("A", 12, FirstClass),
		},
		{
			ID: "B",
			Seats: f.createSeatsForCarriage("B", 20, SecondClass),
		},
		{
			ID: "C",
			Seats: f.createSeatsForCarriage("C", 20, SecondClass),
		},
	}
}

func (f *TestDataFactory) createSeatsForCarriage(carriageID string, count int, comfortZone ComfortZone) []Seat {
	seats := make([]Seat, count)
	for i := 0; i < count; i++ {
		seats[i] = Seat{
			Number:      fmt.Sprintf("%s%d", carriageID, i+1),
			ComfortZone: comfortZone,
			CarriageID:  carriageID,
		}
	}
	return seats
}

// Test scenario management
type TestScenarioManager struct {
	scenarios map[string]TestScenario
	manager   *TestDataManager
}

func NewTestScenarioManager(db *sql.DB) *TestScenarioManager {
	return &TestScenarioManager{
		scenarios: make(map[string]TestScenario),
		manager:   NewTestDataManager(db),
	}
}

// Register test scenarios
func (m *TestScenarioManager) RegisterScenario(scenario TestScenario) {
	m.scenarios[scenario.Name] = scenario
}

// Run specific test scenario
func (m *TestScenarioManager) RunScenario(ctx context.Context, scenarioName string) error {
	scenario, exists := m.scenarios[scenarioName]
	if !exists {
		return fmt.Errorf("scenario %s not found", scenarioName)
	}
	
	// Setup
	if err := scenario.Setup(m.manager); err != nil {
		return fmt.Errorf("scenario setup failed: %w", err)
	}
	
	// Cleanup on exit
	defer func() {
		if err := scenario.Cleanup(m.manager); err != nil {
			// Log cleanup error but don't fail the test
			fmt.Printf("Warning: scenario cleanup failed: %v\n", err)
		}
	}()
	
	return nil
}

// Predefined test scenarios
func (m *TestScenarioManager) RegisterStandardScenarios() {
	// Basic booking scenario
	m.RegisterScenario(TestScenario{
		Name: "basic_booking",
		Setup: func(manager *TestDataManager) error {
			factory := NewTestDataFactory(42)
			
			// Create test route and service
			route := factory.CreateTestRoute()
			service := factory.CreateTestService(route)
			
			// Insert into database
			if err := manager.seeder.SeedRoute(route); err != nil {
				return fmt.Errorf("failed to seed route: %w", err)
			}
			if err := manager.seeder.SeedService(service); err != nil {
				return fmt.Errorf("failed to seed service: %w", err)
			}
			
			return nil
		},
		Cleanup: func(manager *TestDataManager) error {
			return manager.cleaner.CleanupTestData("basic_booking")
		},
	})
	
	// Conductor query scenario
	m.RegisterScenario(TestScenario{
		Name: "conductor_queries",
		Setup: func(manager *TestDataManager) error {
			factory := NewTestDataFactory(123)
			
			// Create multiple bookings for testing conductor queries
			route := factory.CreateTestRoute()
			service := factory.CreateTestService(route)
			
			if err := manager.seeder.SeedRoute(route); err != nil {
				return err
			}
			if err := manager.seeder.SeedService(service); err != nil {
				return err
			}
			
			// Create multiple bookings
			for i := 0; i < 5; i++ {
				passenger := factory.CreateTestPassenger()
				booking := Booking{
					ID:         fmt.Sprintf("TEST_%d", i),
					Passengers: []Passenger{passenger},
					Tickets: []Ticket{
						{
							Seat:        Seat{Number: fmt.Sprintf("A%d", i+1), CarriageID: "A"},
							Origin:      Station{Name: "Paris"},
							Destination: Station{Name: "Amsterdam"},
							Service:     service,
							Passenger:   passenger,
						},
					},
					CreatedAt: time.Now(),
				}
				
				if err := manager.seeder.SeedBooking(booking); err != nil {
					return fmt.Errorf("failed to seed booking %d: %w", i, err)
				}
			}
			
			return nil
		},
		Cleanup: func(manager *TestDataManager) error {
			return manager.cleaner.CleanupTestData("conductor_queries")
		},
	})
}

// Test data seeder
type TestDataSeeder struct {
	db *sql.DB
}

func (s *TestDataSeeder) SeedRoute(route Route) error {
	// Implementation to insert route into database
	_, err := s.db.Exec(`
		INSERT INTO routes (id, name) VALUES ($1, $2)
		ON CONFLICT (id) DO NOTHING`, route.ID, route.Name)
	if err != nil {
		return fmt.Errorf("failed to seed route: %w", err)
	}
	
	// Insert stops
	for _, stop := range route.Stops {
		_, err = s.db.Exec(`
			INSERT INTO route_stops (route_id, station_name, distance, stop_order)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (route_id, stop_order) DO NOTHING`,
			route.ID, stop.Station.Name, stop.Distance, stop.StopOrder)
		if err != nil {
			return fmt.Errorf("failed to seed stop: %w", err)
		}
	}
	
	return nil
}

func (s *TestDataSeeder) SeedService(service Service) error {
	_, err := s.db.Exec(`
		INSERT INTO services (id, route_id, departure_time)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING`,
		service.ID, service.Route.ID, service.DateTime)
	if err != nil {
		return fmt.Errorf("failed to seed service: %w", err)
	}
	
	// Insert carriages and seats
	for _, carriage := range service.Carriages {
		_, err = s.db.Exec(`
			INSERT INTO carriages (id, service_id)
			VALUES ($1, $2)
			ON CONFLICT (id, service_id) DO NOTHING`,
			carriage.ID, service.ID)
		if err != nil {
			return fmt.Errorf("failed to seed carriage: %w", err)
		}
		
		for _, seat := range carriage.Seats {
			_, err = s.db.Exec(`
				INSERT INTO seats (number, carriage_id, comfort_zone)
				VALUES ($1, $2, $3)
				ON CONFLICT (number, carriage_id) DO NOTHING`,
				seat.Number, carriage.ID, string(seat.ComfortZone))
			if err != nil {
				return fmt.Errorf("failed to seed seat: %w", err)
			}
		}
	}
	
	return nil
}

func (s *TestDataSeeder) SeedBooking(booking Booking) error {
	_, err := s.db.Exec(`
		INSERT INTO bookings (id, created_at)
		VALUES ($1, $2)
		ON CONFLICT (id) DO NOTHING`,
		booking.ID, booking.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to seed booking: %w", err)
	}
	
	// Insert tickets
	for _, ticket := range booking.Tickets {
		_, err = s.db.Exec(`
			INSERT INTO tickets (booking_id, seat_number, carriage_id, 
				passenger_name, origin, destination, service_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			booking.ID, ticket.Seat.Number, ticket.Seat.CarriageID,
			ticket.Passenger.Name, ticket.Origin.Name, ticket.Destination.Name,
			ticket.Service.ID)
		if err != nil {
			return fmt.Errorf("failed to seed ticket: %w", err)
		}
	}
	
	return nil
}

// Test data cleaner
type TestDataCleaner struct {
	db *sql.DB
}

func (c *TestDataCleaner) CleanupTestData(scenarioName string) error {
	// Clean up test data in reverse dependency order
	queries := []string{
		"DELETE FROM tickets WHERE booking_id LIKE 'TEST_%'",
		"DELETE FROM bookings WHERE id LIKE 'TEST_%'",
		"DELETE FROM seat_reservations WHERE booking_id LIKE 'TEST_%'",
		"DELETE FROM seats WHERE carriage_id IN (SELECT id FROM carriages WHERE service_id LIKE 'TEST_%')",
		"DELETE FROM carriages WHERE service_id LIKE 'TEST_%'",
		"DELETE FROM services WHERE id LIKE 'TEST_%'",
		"DELETE FROM route_stops WHERE route_id LIKE 'TEST_%'",
		"DELETE FROM routes WHERE id LIKE 'TEST_%'",
	}
	
	for _, query := range queries {
		if _, err := c.db.Exec(query); err != nil {
			return fmt.Errorf("failed to cleanup: %w", err)
		}
	}
	
	return nil
}

// Integration test helper
func RunIntegrationTest(t *testing.T, scenarioName string, testFunc func(*testing.T, *TestDataManager)) {
	// Setup test database
	db := setupTestDatabase(t)
	defer db.Close()
	
	manager := NewTestDataManager(db)
	scenarioManager := NewTestScenarioManager(db)
	scenarioManager.RegisterStandardScenarios()
	
	// Run scenario
	if err := scenarioManager.RunScenario(context.Background(), scenarioName); err != nil {
		t.Fatalf("Failed to run scenario %s: %v", scenarioName, err)
	}
	
	// Run test
	testFunc(t, manager)
}

func setupTestDatabase(t *testing.T) *sql.DB {
	// Implementation to setup test database
	// This would typically use a test container or in-memory database
	return nil // Placeholder
}
