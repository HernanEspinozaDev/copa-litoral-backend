package tests

import (
	"testing"
)

// TestBasic es una prueba básica para verificar que el sistema de testing funciona
func TestBasic(t *testing.T) {
	t.Log("Sistema de testing funcionando correctamente")
	
	// Prueba simple de suma
	result := 2 + 2
	expected := 4
	
	if result != expected {
		t.Errorf("Esperado %d, obtenido %d", expected, result)
	}
}

// TestStringOperations prueba operaciones básicas con strings
func TestStringOperations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"single word", "test", "test"},
		{"multiple words", "hello world", "hello world"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input != tt.expected {
				t.Errorf("Test %s failed: expected %s, got %s", tt.name, tt.expected, tt.input)
			}
		})
	}
}

// BenchmarkStringConcat benchmark para concatenación de strings
func BenchmarkStringConcat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = "hello" + " " + "world"
	}
}

// TestTableDriven ejemplo de prueba table-driven
func TestTableDriven(t *testing.T) {
	testCases := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive numbers", 2, 3, 5},
		{"zero", 0, 5, 5},
		{"negative", -1, 1, 0},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.a + tc.b
			if result != tc.expected {
				t.Errorf("Expected %d + %d = %d, got %d", tc.a, tc.b, tc.expected, result)
			}
		})
	}
}
