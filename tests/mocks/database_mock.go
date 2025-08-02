package mocks

import (
	"fmt"
)

// MockDB es un mock de la base de datos para testing
type MockDB struct {
	queries map[string]MockResult
	errors  map[string]error
}

// MockResult representa el resultado de una query mockeada
type MockResult struct {
	Rows     []map[string]interface{}
	LastID   int64
	Affected int64
}

// NewMockDB crea una nueva instancia de MockDB
func NewMockDB() *MockDB {
	return &MockDB{
		queries: make(map[string]MockResult),
		errors:  make(map[string]error),
	}
}

// SetQueryResult configura el resultado para una query específica
func (m *MockDB) SetQueryResult(query string, result MockResult) {
	m.queries[query] = result
}

// SetQueryError configura un error para una query específica
func (m *MockDB) SetQueryError(query string, err error) {
	m.errors[query] = err
}

// MockRows implementa sql.Rows para testing
type MockRows struct {
	columns []string
	rows    []map[string]interface{}
	current int
}

// NewMockRows crea nuevas filas mockeadas
func NewMockRows(columns []string, rows []map[string]interface{}) *MockRows {
	return &MockRows{
		columns: columns,
		rows:    rows,
		current: -1,
	}
}

// Columns devuelve los nombres de las columnas
func (mr *MockRows) Columns() ([]string, error) {
	return mr.columns, nil
}

// Next avanza al siguiente registro
func (mr *MockRows) Next() bool {
	mr.current++
	return mr.current < len(mr.rows)
}

// Scan escanea los valores de la fila actual
func (mr *MockRows) Scan(dest ...interface{}) error {
	if mr.current < 0 || mr.current >= len(mr.rows) {
		return fmt.Errorf("no rows available")
	}

	row := mr.rows[mr.current]
	for i, col := range mr.columns {
		if i < len(dest) {
			if value, exists := row[col]; exists {
				switch d := dest[i].(type) {
				case *int:
					if v, ok := value.(int); ok {
						*d = v
					}
				case *string:
					if v, ok := value.(string); ok {
						*d = v
					}
				case *bool:
					if v, ok := value.(bool); ok {
						*d = v
					}
				}
			}
		}
	}
	return nil
}

// Close cierra las filas
func (mr *MockRows) Close() error {
	return nil
}

// Err devuelve cualquier error que haya ocurrido
func (mr *MockRows) Err() error {
	return nil
}

// MockJugadorService mock del servicio de jugadores
type MockJugadorService struct {
	jugadores map[int]map[string]interface{}
	nextID    int
	errors    map[string]error
}

// NewMockJugadorService crea un nuevo mock del servicio de jugadores
func NewMockJugadorService() *MockJugadorService {
	return &MockJugadorService{
		jugadores: make(map[int]map[string]interface{}),
		nextID:    1,
		errors:    make(map[string]error),
	}
}

// SetError configura un error para un método específico
func (m *MockJugadorService) SetError(method string, err error) {
	m.errors[method] = err
}

// GetJugadores mock para obtener jugadores
func (m *MockJugadorService) GetJugadores() ([]map[string]interface{}, error) {
	if err, exists := m.errors["GetJugadores"]; exists {
		return nil, err
	}

	var jugadores []map[string]interface{}
	for _, jugador := range m.jugadores {
		jugadores = append(jugadores, jugador)
	}
	return jugadores, nil
}

// GetJugador mock para obtener un jugador por ID
func (m *MockJugadorService) GetJugador(id int) (map[string]interface{}, error) {
	if err, exists := m.errors["GetJugador"]; exists {
		return nil, err
	}

	if jugador, exists := m.jugadores[id]; exists {
		return jugador, nil
	}
	return nil, fmt.Errorf("jugador not found")
}

// CreateJugador mock para crear un jugador
func (m *MockJugadorService) CreateJugador(data map[string]interface{}) (map[string]interface{}, error) {
	if err, exists := m.errors["CreateJugador"]; exists {
		return nil, err
	}

	id := m.nextID
	m.nextID++

	jugador := make(map[string]interface{})
	for k, v := range data {
		jugador[k] = v
	}
	jugador["id"] = id

	m.jugadores[id] = jugador
	return jugador, nil
}

// UpdateJugador mock para actualizar un jugador
func (m *MockJugadorService) UpdateJugador(id int, data map[string]interface{}) (map[string]interface{}, error) {
	if err, exists := m.errors["UpdateJugador"]; exists {
		return nil, err
	}

	if _, exists := m.jugadores[id]; !exists {
		return nil, fmt.Errorf("jugador not found")
	}

	for k, v := range data {
		m.jugadores[id][k] = v
	}

	return m.jugadores[id], nil
}

// DeleteJugador mock para eliminar un jugador
func (m *MockJugadorService) DeleteJugador(id int) error {
	if err, exists := m.errors["DeleteJugador"]; exists {
		return err
	}

	if _, exists := m.jugadores[id]; !exists {
		return fmt.Errorf("jugador not found")
	}

	delete(m.jugadores, id)
	return nil
}

// MockAuthService mock del servicio de autenticación
type MockAuthService struct {
	users  map[string]map[string]interface{}
	errors map[string]error
}

// NewMockAuthService crea un nuevo mock del servicio de autenticación
func NewMockAuthService() *MockAuthService {
	return &MockAuthService{
		users:  make(map[string]map[string]interface{}),
		errors: make(map[string]error),
	}
}

// SetError configura un error para un método específico
func (m *MockAuthService) SetError(method string, err error) {
	m.errors[method] = err
}

// AddUser agrega un usuario al mock
func (m *MockAuthService) AddUser(username string, user map[string]interface{}) {
	m.users[username] = user
}

// Login mock para autenticación
func (m *MockAuthService) Login(username, password string) (string, error) {
	if err, exists := m.errors["Login"]; exists {
		return "", err
	}

	if user, exists := m.users[username]; exists {
		if storedPassword, ok := user["password"].(string); ok && storedPassword == password {
			return "mock-jwt-token", nil
		}
	}

	return "", fmt.Errorf("invalid credentials")
}

// ValidateToken mock para validación de token
func (m *MockAuthService) ValidateToken(token string) (map[string]interface{}, error) {
	if err, exists := m.errors["ValidateToken"]; exists {
		return nil, err
	}

	if token == "mock-jwt-token" {
		return map[string]interface{}{
			"user_id": 1,
			"username": "testuser",
			"rol": "jugador",
		}, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// MockHTTPClient mock para cliente HTTP
type MockHTTPClient struct {
	responses map[string]MockHTTPResponse
	errors    map[string]error
}

// MockHTTPResponse respuesta HTTP mockeada
type MockHTTPResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

// NewMockHTTPClient crea un nuevo mock de cliente HTTP
func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		responses: make(map[string]MockHTTPResponse),
		errors:    make(map[string]error),
	}
}

// SetResponse configura una respuesta para una URL específica
func (m *MockHTTPClient) SetResponse(url string, response MockHTTPResponse) {
	m.responses[url] = response
}

// SetError configura un error para una URL específica
func (m *MockHTTPClient) SetError(url string, err error) {
	m.errors[url] = err
}

// Get simula una petición GET
func (m *MockHTTPClient) Get(url string) (MockHTTPResponse, error) {
	if err, exists := m.errors[url]; exists {
		return MockHTTPResponse{}, err
	}

	if response, exists := m.responses[url]; exists {
		return response, nil
	}

	return MockHTTPResponse{StatusCode: 404, Body: "Not Found"}, nil
}

// MockEmailService mock del servicio de email
type MockEmailService struct {
	sentEmails []MockEmail
	errors     map[string]error
}

// MockEmail representa un email enviado
type MockEmail struct {
	To      string
	Subject string
	Body    string
}

// NewMockEmailService crea un nuevo mock del servicio de email
func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		sentEmails: make([]MockEmail, 0),
		errors:     make(map[string]error),
	}
}

// SetError configura un error para un método específico
func (m *MockEmailService) SetError(method string, err error) {
	m.errors[method] = err
}

// SendEmail mock para envío de email
func (m *MockEmailService) SendEmail(to, subject, body string) error {
	if err, exists := m.errors["SendEmail"]; exists {
		return err
	}

	m.sentEmails = append(m.sentEmails, MockEmail{
		To:      to,
		Subject: subject,
		Body:    body,
	})

	return nil
}

// GetSentEmails devuelve los emails enviados
func (m *MockEmailService) GetSentEmails() []MockEmail {
	return m.sentEmails
}

// ClearSentEmails limpia la lista de emails enviados
func (m *MockEmailService) ClearSentEmails() {
	m.sentEmails = make([]MockEmail, 0)
}
