package integration

import (
	"testing"

	"copa-litoral-backend/tests"
)

func TestDatabaseConnection(t *testing.T) {
	tests.RunTestWithCleanup(t, func(t *testing.T, cfg *tests.TestConfig) {
		// Test basic database connectivity
		err := cfg.DB.Ping()
		if err != nil {
			t.Fatalf("Failed to ping database: %v", err)
		}
	})
}

func TestJugadorCRUD(t *testing.T) {
	tests.RunTestWithCleanup(t, func(t *testing.T, cfg *tests.TestConfig) {
		// Create categoria first
		categoriaID, err := tests.CreateTestCategoria(cfg.DB, "Primera")
		tests.AssertNoError(t, err, "Create categoria")

		// Test Create
		jugadorID, err := tests.CreateTestJugador(cfg.DB, "Juan", "Pérez", &categoriaID)
		tests.AssertNoError(t, err, "Create jugador")
		tests.AssertNotNil(t, jugadorID, "Jugador ID should not be nil")

		// Test Read
		var nombre, apellido, club string
		var retrievedCategoriaID *int
		query := "SELECT nombre, apellido, club, categoria_id FROM jugadores WHERE id = $1"
		err = cfg.DB.QueryRow(query, jugadorID).Scan(&nombre, &apellido, &club, &retrievedCategoriaID)
		tests.AssertNoError(t, err, "Read jugador")
		tests.AssertEqual(t, "Juan", nombre, "Jugador nombre")
		tests.AssertEqual(t, "Pérez", apellido, "Jugador apellido")
		tests.AssertEqual(t, "Club Test", club, "Jugador club")
		tests.AssertEqual(t, categoriaID, *retrievedCategoriaID, "Jugador categoria_id")

		// Test Update
		updateQuery := "UPDATE jugadores SET club = $1 WHERE id = $2"
		_, err = cfg.DB.Exec(updateQuery, "Nuevo Club", jugadorID)
		tests.AssertNoError(t, err, "Update jugador")

		// Verify update
		var updatedClub string
		err = cfg.DB.QueryRow("SELECT club FROM jugadores WHERE id = $1", jugadorID).Scan(&updatedClub)
		tests.AssertNoError(t, err, "Read updated jugador")
		tests.AssertEqual(t, "Nuevo Club", updatedClub, "Updated club")

		// Test Delete
		deleteQuery := "DELETE FROM jugadores WHERE id = $1"
		result, err := cfg.DB.Exec(deleteQuery, jugadorID)
		tests.AssertNoError(t, err, "Delete jugador")

		rowsAffected, err := result.RowsAffected()
		tests.AssertNoError(t, err, "Get rows affected")
		tests.AssertEqual(t, int64(1), rowsAffected, "Rows affected by delete")

		// Verify deletion
		var count int
		err = cfg.DB.QueryRow("SELECT COUNT(*) FROM jugadores WHERE id = $1", jugadorID).Scan(&count)
		tests.AssertNoError(t, err, "Count after delete")
		tests.AssertEqual(t, 0, count, "Jugador should be deleted")
	})
}

func TestUsuarioCRUD(t *testing.T) {
	tests.RunTestWithCleanup(t, func(t *testing.T, cfg *tests.TestConfig) {
		// Test Create
		userID, err := tests.CreateTestUser(cfg.DB, "testuser", "test@example.com", "jugador")
		tests.AssertNoError(t, err, "Create user")
		tests.AssertNotNil(t, userID, "User ID should not be nil")

		// Test Read
		var username, email, rol string
		query := "SELECT nombre_usuario, email, rol FROM usuarios WHERE id = $1"
		err = cfg.DB.QueryRow(query, userID).Scan(&username, &email, &rol)
		tests.AssertNoError(t, err, "Read user")
		tests.AssertEqual(t, "testuser", username, "Username")
		tests.AssertEqual(t, "test@example.com", email, "Email")
		tests.AssertEqual(t, "jugador", rol, "Rol")

		// Test Update
		updateQuery := "UPDATE usuarios SET email = $1 WHERE id = $2"
		_, err = cfg.DB.Exec(updateQuery, "newemail@example.com", userID)
		tests.AssertNoError(t, err, "Update user")

		// Verify update
		var updatedEmail string
		err = cfg.DB.QueryRow("SELECT email FROM usuarios WHERE id = $1", userID).Scan(&updatedEmail)
		tests.AssertNoError(t, err, "Read updated user")
		tests.AssertEqual(t, "newemail@example.com", updatedEmail, "Updated email")

		// Test Delete
		deleteQuery := "DELETE FROM usuarios WHERE id = $1"
		result, err := cfg.DB.Exec(deleteQuery, userID)
		tests.AssertNoError(t, err, "Delete user")

		rowsAffected, err := result.RowsAffected()
		tests.AssertNoError(t, err, "Get rows affected")
		tests.AssertEqual(t, int64(1), rowsAffected, "Rows affected by delete")
	})
}

func TestTorneoCRUD(t *testing.T) {
	tests.RunTestWithCleanup(t, func(t *testing.T, cfg *tests.TestConfig) {
		// Test Create
		torneoID, err := tests.CreateTestTorneo(cfg.DB, "Copa Test 2024", 2024)
		tests.AssertNoError(t, err, "Create torneo")
		tests.AssertNotNil(t, torneoID, "Torneo ID should not be nil")

		// Test Read
		var nombre string
		var anio int
		query := "SELECT nombre, anio FROM torneos WHERE id = $1"
		err = cfg.DB.QueryRow(query, torneoID).Scan(&nombre, &anio)
		tests.AssertNoError(t, err, "Read torneo")
		tests.AssertEqual(t, "Copa Test 2024", nombre, "Torneo nombre")
		tests.AssertEqual(t, 2024, anio, "Torneo anio")

		// Test Update
		updateQuery := "UPDATE torneos SET nombre = $1 WHERE id = $2"
		_, err = cfg.DB.Exec(updateQuery, "Copa Test Actualizada 2024", torneoID)
		tests.AssertNoError(t, err, "Update torneo")

		// Verify update
		var updatedNombre string
		err = cfg.DB.QueryRow("SELECT nombre FROM torneos WHERE id = $1", torneoID).Scan(&updatedNombre)
		tests.AssertNoError(t, err, "Read updated torneo")
		tests.AssertEqual(t, "Copa Test Actualizada 2024", updatedNombre, "Updated nombre")

		// Test Delete
		deleteQuery := "DELETE FROM torneos WHERE id = $1"
		result, err := cfg.DB.Exec(deleteQuery, torneoID)
		tests.AssertNoError(t, err, "Delete torneo")

		rowsAffected, err := result.RowsAffected()
		tests.AssertNoError(t, err, "Get rows affected")
		tests.AssertEqual(t, int64(1), rowsAffected, "Rows affected by delete")
	})
}

func TestRelationshipConstraints(t *testing.T) {
	tests.RunTestWithCleanup(t, func(t *testing.T, cfg *tests.TestConfig) {
		// Create categoria
		categoriaID, err := tests.CreateTestCategoria(cfg.DB, "Primera")
		tests.AssertNoError(t, err, "Create categoria")

		// Create jugador with categoria
		jugadorID, err := tests.CreateTestJugador(cfg.DB, "Juan", "Pérez", &categoriaID)
		tests.AssertNoError(t, err, "Create jugador")

		// Try to delete categoria that has jugadores (should fail due to foreign key)
		_, err = cfg.DB.Exec("DELETE FROM categorias WHERE id = $1", categoriaID)
		tests.AssertError(t, err, "Should not be able to delete categoria with jugadores")

		// Delete jugador first
		_, err = cfg.DB.Exec("DELETE FROM jugadores WHERE id = $1", jugadorID)
		tests.AssertNoError(t, err, "Delete jugador")

		// Now should be able to delete categoria
		_, err = cfg.DB.Exec("DELETE FROM categorias WHERE id = $1", categoriaID)
		tests.AssertNoError(t, err, "Delete categoria after removing jugadores")
	})
}

func TestTransactions(t *testing.T) {
	tests.RunTestWithCleanup(t, func(t *testing.T, cfg *tests.TestConfig) {
		// Test successful transaction
		tx, err := cfg.DB.Begin()
		tests.AssertNoError(t, err, "Begin transaction")

		// Create categoria within transaction
		var categoriaID int
		err = tx.QueryRow("INSERT INTO categorias (nombre, created_at, updated_at) VALUES ($1, NOW(), NOW()) RETURNING id", "Transacción Test").Scan(&categoriaID)
		tests.AssertNoError(t, err, "Create categoria in transaction")

		// Commit transaction
		err = tx.Commit()
		tests.AssertNoError(t, err, "Commit transaction")

		// Verify categoria exists
		var count int
		err = cfg.DB.QueryRow("SELECT COUNT(*) FROM categorias WHERE id = $1", categoriaID).Scan(&count)
		tests.AssertNoError(t, err, "Count categorias after commit")
		tests.AssertEqual(t, 1, count, "Categoria should exist after commit")

		// Test rollback transaction
		tx2, err := cfg.DB.Begin()
		tests.AssertNoError(t, err, "Begin second transaction")

		// Create another categoria
		var categoriaID2 int
		err = tx2.QueryRow("INSERT INTO categorias (nombre, created_at, updated_at) VALUES ($1, NOW(), NOW()) RETURNING id", "Rollback Test").Scan(&categoriaID2)
		tests.AssertNoError(t, err, "Create categoria in second transaction")

		// Rollback transaction
		err = tx2.Rollback()
		tests.AssertNoError(t, err, "Rollback transaction")

		// Verify categoria doesn't exist
		err = cfg.DB.QueryRow("SELECT COUNT(*) FROM categorias WHERE id = $1", categoriaID2).Scan(&count)
		tests.AssertNoError(t, err, "Count categorias after rollback")
		tests.AssertEqual(t, 0, count, "Categoria should not exist after rollback")
	})
}

func TestComplexQueries(t *testing.T) {
	tests.RunTestWithCleanup(t, func(t *testing.T, cfg *tests.TestConfig) {
		// Setup test data
		categoriaID, err := tests.CreateTestCategoria(cfg.DB, "Primera")
		tests.AssertNoError(t, err, "Create categoria")

		jugador1ID, err := tests.CreateTestJugador(cfg.DB, "Juan", "Pérez", &categoriaID)
		tests.AssertNoError(t, err, "Create jugador 1")

		jugador2ID, err := tests.CreateTestJugador(cfg.DB, "María", "García", &categoriaID)
		tests.AssertNoError(t, err, "Create jugador 2")

		torneoID, err := tests.CreateTestTorneo(cfg.DB, "Copa Test", 2024)
		tests.AssertNoError(t, err, "Create torneo")

		// Test JOIN query
		query := `
			SELECT j.nombre, j.apellido, c.nombre as categoria_nombre
			FROM jugadores j
			JOIN categorias c ON j.categoria_id = c.id
			WHERE c.id = $1
			ORDER BY j.nombre
		`
		rows, err := cfg.DB.Query(query, categoriaID)
		tests.AssertNoError(t, err, "Execute JOIN query")
		defer rows.Close()

		var jugadores []map[string]string
		for rows.Next() {
			var nombre, apellido, categoriaNombre string
			err := rows.Scan(&nombre, &apellido, &categoriaNombre)
			tests.AssertNoError(t, err, "Scan JOIN result")
			
			jugadores = append(jugadores, map[string]string{
				"nombre":           nombre,
				"apellido":         apellido,
				"categoria_nombre": categoriaNombre,
			})
		}

		tests.AssertEqual(t, 2, len(jugadores), "Should have 2 jugadores")
		tests.AssertEqual(t, "Juan", jugadores[0]["nombre"], "First jugador nombre")
		tests.AssertEqual(t, "María", jugadores[1]["nombre"], "Second jugador nombre")

		// Test aggregate query
		var totalJugadores int
		err = cfg.DB.QueryRow("SELECT COUNT(*) FROM jugadores WHERE categoria_id = $1", categoriaID).Scan(&totalJugadores)
		tests.AssertNoError(t, err, "Count jugadores by categoria")
		tests.AssertEqual(t, 2, totalJugadores, "Total jugadores in categoria")

		// Clean up
		cfg.DB.Exec("DELETE FROM jugadores WHERE id IN ($1, $2)", jugador1ID, jugador2ID)
		cfg.DB.Exec("DELETE FROM torneos WHERE id = $1", torneoID)
		cfg.DB.Exec("DELETE FROM categorias WHERE id = $1", categoriaID)
	})
}

func TestDatabaseMigrations(t *testing.T) {
	tests.RunTestWithCleanup(t, func(t *testing.T, cfg *tests.TestConfig) {
		// Test that all expected tables exist
		expectedTables := []string{
			"usuarios",
			"categorias",
			"jugadores",
			"torneos",
			"partidos",
			"schema_migrations",
		}

		for _, table := range expectedTables {
			var exists bool
			query := `
				SELECT EXISTS (
					SELECT FROM information_schema.tables 
					WHERE table_schema = 'public' 
					AND table_name = $1
				)
			`
			err := cfg.DB.QueryRow(query, table).Scan(&exists)
			tests.AssertNoError(t, err, "Check table existence")
			tests.AssertEqual(t, true, exists, "Table "+table+" should exist")
		}

		// Test that schema_migrations table has entries
		var migrationCount int
		err := cfg.DB.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&migrationCount)
		tests.AssertNoError(t, err, "Count migrations")
		
		if migrationCount == 0 {
			t.Log("Warning: No migrations found in schema_migrations table")
		}
	})
}

func TestConnectionPooling(t *testing.T) {
	tests.RunTestWithCleanup(t, func(t *testing.T, cfg *tests.TestConfig) {
		// Test multiple concurrent connections
		const numConnections = 10
		done := make(chan bool, numConnections)
		errors := make(chan error, numConnections)

		for i := 0; i < numConnections; i++ {
			go func(id int) {
				defer func() { done <- true }()
				
				// Execute a simple query
				var result int
				err := cfg.DB.QueryRow("SELECT $1", id).Scan(&result)
				if err != nil {
					errors <- err
					return
				}
				
				if result != id {
					errors <- testing.ErrExample
					return
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numConnections; i++ {
			<-done
		}

		// Check for errors
		select {
		case err := <-errors:
			t.Errorf("Connection pooling test failed: %v", err)
		default:
			// No errors, test passed
		}
	})
}

// Benchmark tests for database operations
func BenchmarkDatabaseInsert(b *testing.B) {
	cfg := tests.SetupTestDB()
	defer tests.TeardownTestDB()

	// Clean up before benchmark
	tests.CleanupTestData(cfg.DB)

	// Create a categoria for the benchmark
	categoriaID, _ := tests.CreateTestCategoria(cfg.DB, "Benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tests.CreateTestJugador(cfg.DB, "Benchmark", "User", &categoriaID)
	}
}

func BenchmarkDatabaseSelect(b *testing.B) {
	cfg := tests.SetupTestDB()
	defer tests.TeardownTestDB()

	// Setup test data
	tests.CleanupTestData(cfg.DB)
	categoriaID, _ := tests.CreateTestCategoria(cfg.DB, "Benchmark")
	jugadorID, _ := tests.CreateTestJugador(cfg.DB, "Benchmark", "User", &categoriaID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var nombre string
		cfg.DB.QueryRow("SELECT nombre FROM jugadores WHERE id = $1", jugadorID).Scan(&nombre)
	}
}
