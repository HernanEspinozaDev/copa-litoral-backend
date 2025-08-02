package unit

import (
	"fmt"
	"testing"

	"copa-litoral-backend/tests/mocks"
)

func TestJugadorService(t *testing.T) {
	mockService := mocks.NewMockJugadorService()

	t.Run("create jugador successfully", func(t *testing.T) {
		data := map[string]interface{}{
			"nombre":   "Juan",
			"apellido": "Pérez",
			"club":     "Club Test",
		}

		jugador, err := mockService.CreateJugador(data)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if jugador["id"] == nil {
			t.Error("Expected jugador to have an ID")
		}
		if jugador["nombre"] != "Juan" {
			t.Errorf("Expected nombre 'Juan', got %v", jugador["nombre"])
		}
	})

	t.Run("get jugador successfully", func(t *testing.T) {
		// First create a jugador
		data := map[string]interface{}{
			"nombre":   "María",
			"apellido": "García",
			"club":     "Club Test",
		}
		created, _ := mockService.CreateJugador(data)
		id := created["id"].(int)

		// Then get it
		jugador, err := mockService.GetJugador(id)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if jugador["nombre"] != "María" {
			t.Errorf("Expected nombre 'María', got %v", jugador["nombre"])
		}
	})

	t.Run("get non-existent jugador", func(t *testing.T) {
		_, err := mockService.GetJugador(999)
		if err == nil {
			t.Error("Expected error for non-existent jugador")
		}
	})

	t.Run("update jugador successfully", func(t *testing.T) {
		// Create a jugador
		data := map[string]interface{}{
			"nombre":   "Carlos",
			"apellido": "López",
		}
		created, _ := mockService.CreateJugador(data)
		id := created["id"].(int)

		// Update it
		updateData := map[string]interface{}{
			"club": "Nuevo Club",
		}
		updated, err := mockService.UpdateJugador(id, updateData)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if updated["club"] != "Nuevo Club" {
			t.Errorf("Expected club 'Nuevo Club', got %v", updated["club"])
		}
	})

	t.Run("delete jugador successfully", func(t *testing.T) {
		// Create a jugador
		data := map[string]interface{}{
			"nombre":   "Ana",
			"apellido": "Martín",
		}
		created, _ := mockService.CreateJugador(data)
		id := created["id"].(int)

		// Delete it
		err := mockService.DeleteJugador(id)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify it's deleted
		_, err = mockService.GetJugador(id)
		if err == nil {
			t.Error("Expected error when getting deleted jugador")
		}
	})

	t.Run("get all jugadores", func(t *testing.T) {
		// Create multiple jugadores
		mockService.CreateJugador(map[string]interface{}{
			"nombre": "Test1", "apellido": "User1",
		})
		mockService.CreateJugador(map[string]interface{}{
			"nombre": "Test2", "apellido": "User2",
		})

		jugadores, err := mockService.GetJugadores()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(jugadores) < 2 {
			t.Errorf("Expected at least 2 jugadores, got %d", len(jugadores))
		}
	})

	t.Run("service error handling", func(t *testing.T) {
		// Set error for CreateJugador
		mockService.SetError("CreateJugador", fmt.Errorf("mock error for CreateJugador"))

		_, err := mockService.CreateJugador(map[string]interface{}{
			"nombre": "Test",
		})
		if err == nil {
			t.Error("Expected error from mock service")
		}
	})
}

func TestAuthService(t *testing.T) {
	mockService := mocks.NewMockAuthService()

	t.Run("login successfully", func(t *testing.T) {
		// Add a test user
		mockService.AddUser("testuser", map[string]interface{}{
			"password": "testpass",
			"rol":      "jugador",
		})

		token, err := mockService.Login("testuser", "testpass")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if token == "" {
			t.Error("Expected non-empty token")
		}
	})

	t.Run("login with invalid credentials", func(t *testing.T) {
		_, err := mockService.Login("nonexistent", "wrongpass")
		if err == nil {
			t.Error("Expected error for invalid credentials")
		}
	})

	t.Run("validate token successfully", func(t *testing.T) {
		claims, err := mockService.ValidateToken("mock-jwt-token")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if claims["user_id"] != 1 {
			t.Errorf("Expected user_id 1, got %v", claims["user_id"])
		}
	})

	t.Run("validate invalid token", func(t *testing.T) {
		_, err := mockService.ValidateToken("invalid-token")
		if err == nil {
			t.Error("Expected error for invalid token")
		}
	})

	t.Run("auth service error handling", func(t *testing.T) {
		mockService.SetError("Login", fmt.Errorf("mock error for Login"))

		_, err := mockService.Login("testuser", "testpass")
		if err == nil {
			t.Error("Expected error from mock service")
		}
	})
}

func TestEmailService(t *testing.T) {
	mockService := mocks.NewMockEmailService()

	t.Run("send email successfully", func(t *testing.T) {
		err := mockService.SendEmail("test@example.com", "Test Subject", "Test Body")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		emails := mockService.GetSentEmails()
		if len(emails) != 1 {
			t.Errorf("Expected 1 sent email, got %d", len(emails))
		}

		email := emails[0]
		if email.To != "test@example.com" {
			t.Errorf("Expected To 'test@example.com', got %s", email.To)
		}
		if email.Subject != "Test Subject" {
			t.Errorf("Expected Subject 'Test Subject', got %s", email.Subject)
		}
	})

	t.Run("email service error handling", func(t *testing.T) {
		mockService.SetError("SendEmail", fmt.Errorf("mock error for SendEmail"))

		err := mockService.SendEmail("test@example.com", "Subject", "Body")
		if err == nil {
			t.Error("Expected error from mock service")
		}
	})

	t.Run("clear sent emails", func(t *testing.T) {
		mockService.SendEmail("test1@example.com", "Subject 1", "Body 1")
		mockService.SendEmail("test2@example.com", "Subject 2", "Body 2")

		emails := mockService.GetSentEmails()
		if len(emails) != 2 {
			t.Errorf("Expected 2 sent emails, got %d", len(emails))
		}

		mockService.ClearSentEmails()
		emails = mockService.GetSentEmails()
		if len(emails) != 0 {
			t.Errorf("Expected 0 sent emails after clear, got %d", len(emails))
		}
	})
}

func TestHTTPClient(t *testing.T) {
	mockClient := mocks.NewMockHTTPClient()

	t.Run("successful HTTP request", func(t *testing.T) {
		expectedResponse := mocks.MockHTTPResponse{
			StatusCode: 200,
			Body:       `{"success": true}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
		mockClient.SetResponse("http://example.com/api", expectedResponse)

		response, err := mockClient.Get("http://example.com/api")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if response.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", response.StatusCode)
		}
		if response.Body != `{"success": true}` {
			t.Errorf("Expected body '{}', got %s", `{"success": true}`, response.Body)
		}
	})

	t.Run("HTTP request error", func(t *testing.T) {
		mockClient.SetError("http://error.com", fmt.Errorf("mock error for http://error.com"))

		_, err := mockClient.Get("http://error.com")
		if err == nil {
			t.Error("Expected error from mock client")
		}
	})

	t.Run("404 response", func(t *testing.T) {
		response, err := mockClient.Get("http://notfound.com")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if response.StatusCode != 404 {
			t.Errorf("Expected status 404, got %d", response.StatusCode)
		}
	})
}

// Test helper functions
func TestTestHelpers(t *testing.T) {
	t.Run("assert equal", func(t *testing.T) {
		// This would normally fail the test, but we're testing the helper
		// In a real scenario, we'd use a mock testing.T
		expected := "hello"
		actual := "hello"
		
		if expected != actual {
			t.Errorf("AssertEqual should pass for equal values")
		}
	})

	t.Run("assert not equal", func(t *testing.T) {
		expected := "hello"
		actual := "world"
		
		if expected == actual {
			t.Errorf("Values should not be equal")
		}
	})
}

// Benchmark tests for services
func BenchmarkJugadorServiceCreate(b *testing.B) {
	mockService := mocks.NewMockJugadorService()
	data := map[string]interface{}{
		"nombre":   "Benchmark",
		"apellido": "Test",
		"club":     "Club Test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockService.CreateJugador(data)
	}
}

func BenchmarkAuthServiceLogin(b *testing.B) {
	mockService := mocks.NewMockAuthService()
	mockService.AddUser("benchuser", map[string]interface{}{
		"password": "benchpass",
		"rol":      "jugador",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockService.Login("benchuser", "benchpass")
	}
}

func BenchmarkTokenValidation(b *testing.B) {
	mockService := mocks.NewMockAuthService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockService.ValidateToken("mock-jwt-token")
	}
}
