package unit

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"copa-litoral-backend/utils"
)

func TestPaginationParams(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected utils.PaginationParams
	}{
		{
			name:  "default parameters",
			query: "",
			expected: utils.PaginationParams{
				Page:   1,
				Limit:  20,
				Offset: 0,
				Sort:   "id",
				Order:  "asc",
			},
		},
		{
			name:  "custom parameters",
			query: "page=2&limit=10&sort=nombre&order=desc&search=test",
			expected: utils.PaginationParams{
				Page:   2,
				Limit:  10,
				Offset: 10,
				Sort:   "nombre",
				Order:  "desc",
				Search: "test",
			},
		},
		{
			name:  "limit exceeds maximum",
			query: "limit=200",
			expected: utils.PaginationParams{
				Page:   1,
				Limit:  100, // Should be capped at max
				Offset: 0,
				Sort:   "id",
				Order:  "asc",
			},
		},
		{
			name:  "invalid page number",
			query: "page=0",
			expected: utils.PaginationParams{
				Page:   1, // Should default to 1
				Limit:  20,
				Offset: 0,
				Sort:   "id",
				Order:  "asc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?"+tt.query, nil)
			params := utils.ParsePaginationParams(req)

			if params.Page != tt.expected.Page {
				t.Errorf("Page: expected %d, got %d", tt.expected.Page, params.Page)
			}
			if params.Limit != tt.expected.Limit {
				t.Errorf("Limit: expected %d, got %d", tt.expected.Limit, params.Limit)
			}
			if params.Offset != tt.expected.Offset {
				t.Errorf("Offset: expected %d, got %d", tt.expected.Offset, params.Offset)
			}
			if params.Sort != tt.expected.Sort {
				t.Errorf("Sort: expected %s, got %s", tt.expected.Sort, params.Sort)
			}
			if params.Order != tt.expected.Order {
				t.Errorf("Order: expected %s, got %s", tt.expected.Order, params.Order)
			}
			if params.Search != tt.expected.Search {
				t.Errorf("Search: expected %s, got %s", tt.expected.Search, params.Search)
			}
		})
	}
}

func TestCalculatePaginationInfo(t *testing.T) {
	tests := []struct {
		name     string
		params   utils.PaginationParams
		total    int
		expected utils.PaginationInfo
	}{
		{
			name: "first page with results",
			params: utils.PaginationParams{
				Page:  1,
				Limit: 10,
			},
			total: 25,
			expected: utils.PaginationInfo{
				Page:       1,
				Limit:      10,
				Total:      25,
				TotalPages: 3,
				HasNext:    true,
				HasPrev:    false,
			},
		},
		{
			name: "middle page",
			params: utils.PaginationParams{
				Page:  2,
				Limit: 10,
			},
			total: 25,
			expected: utils.PaginationInfo{
				Page:       2,
				Limit:      10,
				Total:      25,
				TotalPages: 3,
				HasNext:    true,
				HasPrev:    true,
			},
		},
		{
			name: "last page",
			params: utils.PaginationParams{
				Page:  3,
				Limit: 10,
			},
			total: 25,
			expected: utils.PaginationInfo{
				Page:       3,
				Limit:      10,
				Total:      25,
				TotalPages: 3,
				HasNext:    false,
				HasPrev:    true,
			},
		},
		{
			name: "no results",
			params: utils.PaginationParams{
				Page:  1,
				Limit: 10,
			},
			total: 0,
			expected: utils.PaginationInfo{
				Page:       1,
				Limit:      10,
				Total:      0,
				TotalPages: 0,
				HasNext:    false,
				HasPrev:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := tt.params.CalculatePaginationInfo(tt.total)

			if info.Page != tt.expected.Page {
				t.Errorf("Page: expected %d, got %d", tt.expected.Page, info.Page)
			}
			if info.Limit != tt.expected.Limit {
				t.Errorf("Limit: expected %d, got %d", tt.expected.Limit, info.Limit)
			}
			if info.Total != tt.expected.Total {
				t.Errorf("Total: expected %d, got %d", tt.expected.Total, info.Total)
			}
			if info.TotalPages != tt.expected.TotalPages {
				t.Errorf("TotalPages: expected %d, got %d", tt.expected.TotalPages, info.TotalPages)
			}
			if info.HasNext != tt.expected.HasNext {
				t.Errorf("HasNext: expected %t, got %t", tt.expected.HasNext, info.HasNext)
			}
			if info.HasPrev != tt.expected.HasPrev {
				t.Errorf("HasPrev: expected %t, got %t", tt.expected.HasPrev, info.HasPrev)
			}
		})
	}
}

func TestVersionManager(t *testing.T) {
	vm := utils.NewVersionManager()

	t.Run("parse valid version", func(t *testing.T) {
		version, err := vm.ParseVersion("1.2.3")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if version.Major != 1 || version.Minor != 2 || version.Patch != 3 {
			t.Errorf("Expected version 1.2.3, got %d.%d.%d", version.Major, version.Minor, version.Patch)
		}
	})

	t.Run("parse version with prefix", func(t *testing.T) {
		version, err := vm.ParseVersion("v2.0.0")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if version.Major != 2 || version.Minor != 0 || version.Patch != 0 {
			t.Errorf("Expected version 2.0.0, got %d.%d.%d", version.Major, version.Minor, version.Patch)
		}
	})

	t.Run("parse invalid version", func(t *testing.T) {
		_, err := vm.ParseVersion("invalid")
		if err == nil {
			t.Error("Expected error for invalid version")
		}
	})

	t.Run("version compatibility", func(t *testing.T) {
		v1 := utils.APIVersion{Major: 1, Minor: 2, Patch: 0}
		v2 := utils.APIVersion{Major: 1, Minor: 1, Patch: 0}
		v3 := utils.APIVersion{Major: 2, Minor: 0, Patch: 0}

		if !v1.IsCompatible(v2) {
			t.Error("v1.2.0 should be compatible with v1.1.0")
		}
		if v1.IsCompatible(v3) {
			t.Error("v1.2.0 should not be compatible with v2.0.0")
		}
	})
}

func TestFilterManager(t *testing.T) {
	fm := utils.NewFilterManager()

	t.Run("parse simple filters", func(t *testing.T) {
		values := url.Values{}
		values.Set("nombre[like]", "Juan")
		values.Set("categoria_id[eq]", "1")

		filter, err := fm.ParseAdvancedFilters(values, "jugadores")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if len(filter.Groups) != 1 {
			t.Errorf("Expected 1 group, got %d", len(filter.Groups))
		}

		group := filter.Groups[0]
		if len(group.Conditions) != 2 {
			t.Errorf("Expected 2 conditions, got %d", len(group.Conditions))
		}
	})

	t.Run("reject invalid field", func(t *testing.T) {
		values := url.Values{}
		values.Set("invalid_field[eq]", "value")

		filter, err := fm.ParseAdvancedFilters(values, "jugadores")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Should ignore invalid fields
		if len(filter.Groups) > 0 && len(filter.Groups[0].Conditions) > 0 {
			t.Error("Expected no conditions for invalid field")
		}
	})

	t.Run("build SQL conditions", func(t *testing.T) {
		filter := utils.AdvancedFilter{
			Groups: []utils.FilterGroup{
				{
					Logic: "AND",
					Conditions: []utils.FilterCondition{
						{
							Field:    "nombre",
							Operator: utils.OpLike,
							Value:    "Juan%",
						},
						{
							Field:    "categoria_id",
							Operator: utils.OpEqual,
							Value:    1,
						},
					},
				},
			},
		}

		sql, args, err := fm.BuildSQLConditions(filter)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if sql == "" {
			t.Error("Expected non-empty SQL")
		}
		if len(args) != 2 {
			t.Errorf("Expected 2 arguments, got %d", len(args))
		}
	})
}

func TestResponseManager(t *testing.T) {
	rm := utils.NewResponseManager(false, true)

	t.Run("write success response", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		data := map[string]interface{}{
			"id":   1,
			"name": "Test",
		}

		rm.WriteSuccessResponse(recorder, req, http.StatusOK, "Success", data, nil)

		if recorder.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", recorder.Code)
		}

		contentType := recorder.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}
	})

	t.Run("write error response", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		rm.WriteErrorResponse(recorder, req, http.StatusBadRequest, utils.ErrInvalidInput, "Invalid data", nil)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("write validation errors", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", nil)

		errors := []utils.ErrorDetail{
			{
				Code:    utils.ErrMissingField,
				Message: "Name is required",
				Field:   "name",
			},
			{
				Code:    utils.ErrInvalidFormat,
				Message: "Email format is invalid",
				Field:   "email",
			},
		}

		rm.WriteValidationErrors(recorder, req, errors)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", recorder.Code)
		}
	})
}

func TestJWTUtils(t *testing.T) {
	secret := "test-secret-key"

	t.Run("generate and parse JWT", func(t *testing.T) {
		// Test JWT generation with correct parameters
		userID := 1
		rol := "jugador"

		// Note: This test assumes JWT utility functions exist
		// If they don't exist yet, this test will fail and we'll need to implement them
		token, err := utils.GenerateJWT(userID, rol, secret)
		if err != nil {
			t.Errorf("Expected no error generating JWT, got %v", err)
		}

		if token == "" {
			t.Error("Expected non-empty token")
		}

		parsedClaims, err := utils.ParseJWT(token, secret)
		if err != nil {
			t.Errorf("Expected no error parsing JWT, got %v", err)
		}

		if parsedClaims.UserID != 1 {
			t.Errorf("Expected user_id 1, got %v", parsedClaims.UserID)
		}
	})

	t.Run("parse invalid JWT", func(t *testing.T) {
		_, err := utils.ParseJWT("invalid-token", secret)
		if err == nil {
			t.Error("Expected error for invalid token")
		}
	})
}

func TestInputSanitization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean input",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "input with HTML",
			input:    "<script>alert('xss')</script>Hello",
			expected: "Hello", // Should strip HTML
		},
		{
			name:     "input with SQL injection attempt",
			input:    "'; DROP TABLE users; --",
			expected: "'; DROP TABLE users; --", // Should be escaped when used in queries
		},
		{
			name:     "input with special characters",
			input:    "Hello & <World>",
			expected: "Hello &amp; &lt;World&gt;", // Should escape HTML entities
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Implement SanitizeInput function in utils
			// result := utils.SanitizeInput(tt.input)
			// if result != tt.expected {
			//	t.Errorf("Expected %s, got %s", tt.expected, result)
			// }
			t.Skip("SanitizeInput function not implemented yet")
		})
	}
}

func BenchmarkPaginationParsing(b *testing.B) {
	req := httptest.NewRequest("GET", "/?page=5&limit=50&sort=nombre&order=desc&search=test", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.ParsePaginationParams(req)
	}
}

func BenchmarkFilterParsing(b *testing.B) {
	fm := utils.NewFilterManager()
	values := url.Values{}
	values.Set("nombre[like]", "Juan")
	values.Set("categoria_id[eq]", "1")
	values.Set("created_at[gte]", "2024-01-01")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fm.ParseAdvancedFilters(values, "jugadores")
	}
}
