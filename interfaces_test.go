package main

import (
	"testing"
)

func TestHTTPResponse(t *testing.T) {
	response := HTTPResponse{
		StatusCode: 200,
		Body:       "test body",
	}
	
	if response.GetStatusCode() != 200 {
		t.Errorf("Expected status code 200, got %d", response.GetStatusCode())
	}
	
	if response.GetBody() != "test body" {
		t.Errorf("Expected body 'test body', got '%v'", response.GetBody())
	}
}

func TestMockHTTPClient(t *testing.T) {
	client := &MockHTTPClient{}
	
	response := client.Post("/test", "test data")
	if response.GetStatusCode() != 200 {
		t.Errorf("Expected status code 200, got %d", response.GetStatusCode())
	}
	
	response = client.Get("/test")
	if response.GetStatusCode() != 200 {
		t.Errorf("Expected status code 200, got %d", response.GetStatusCode())
	}
}
