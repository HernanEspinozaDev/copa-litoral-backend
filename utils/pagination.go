package utils

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// PaginationParams representa los parámetros de paginación
type PaginationParams struct {
	Page     int `json:"page"`
	Limit    int `json:"limit"`
	Offset   int `json:"offset"`
	Sort     string `json:"sort,omitempty"`
	Order    string `json:"order,omitempty"`
	Search   string `json:"search,omitempty"`
	Filters  map[string]interface{} `json:"filters,omitempty"`
}

// PaginationInfo contiene información sobre la paginación
type PaginationInfo struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
	NextPage   *int `json:"next_page,omitempty"`
	PrevPage   *int `json:"prev_page,omitempty"`
}

// PaginatedResponse representa una respuesta paginada
type PaginatedResponse struct {
	Success    bool           `json:"success"`
	Message    string         `json:"message"`
	Data       interface{}    `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
	Timestamp  string         `json:"timestamp"`
}

// PaginationConfig configuración por defecto para paginación
type PaginationConfig struct {
	DefaultLimit int
	MaxLimit     int
	DefaultSort  string
	DefaultOrder string
}

// DefaultPaginationConfig configuración por defecto
var DefaultPaginationConfig = PaginationConfig{
	DefaultLimit: 20,
	MaxLimit:     100,
	DefaultSort:  "id",
	DefaultOrder: "asc",
}

// ParsePaginationParams extrae y valida los parámetros de paginación de la request
func ParsePaginationParams(r *http.Request, config ...PaginationConfig) PaginationParams {
	cfg := DefaultPaginationConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	params := PaginationParams{
		Page:    1,
		Limit:   cfg.DefaultLimit,
		Sort:    cfg.DefaultSort,
		Order:   cfg.DefaultOrder,
		Filters: make(map[string]interface{}),
	}

	query := r.URL.Query()

	// Parsear página
	if pageStr := query.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	// Parsear límite
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			if limit > cfg.MaxLimit {
				params.Limit = cfg.MaxLimit
			} else {
				params.Limit = limit
			}
		}
	}

	// Parsear ordenamiento
	if sort := query.Get("sort"); sort != "" {
		params.Sort = sanitizeSortField(sort)
	}

	// Parsear orden
	if order := query.Get("order"); order != "" {
		if strings.ToLower(order) == "desc" {
			params.Order = "desc"
		} else {
			params.Order = "asc"
		}
	}

	// Parsear búsqueda
	params.Search = query.Get("search")

	// Parsear filtros específicos
	params.Filters = parseFilters(query)

	// Calcular offset
	params.Offset = (params.Page - 1) * params.Limit

	return params
}

// parseFilters extrae filtros específicos de los query parameters
func parseFilters(query url.Values) map[string]interface{} {
	filters := make(map[string]interface{})

	// Filtros comunes que se pueden aplicar
	filterFields := []string{
		"categoria_id", "estado", "activo", "torneo_id", "jugador_id",
		"fecha_desde", "fecha_hasta", "club", "rol",
	}

	for _, field := range filterFields {
		if value := query.Get(field); value != "" {
			// Intentar convertir a entero si es posible
			if intValue, err := strconv.Atoi(value); err == nil {
				filters[field] = intValue
			} else if boolValue, err := strconv.ParseBool(value); err == nil {
				filters[field] = boolValue
			} else {
				filters[field] = value
			}
		}
	}

	return filters
}

// sanitizeSortField sanitiza el campo de ordenamiento para prevenir inyección SQL
func sanitizeSortField(sort string) string {
	// Lista blanca de campos permitidos para ordenamiento
	allowedFields := map[string]bool{
		"id":                      true,
		"nombre":                  true,
		"apellido":                true,
		"created_at":              true,
		"updated_at":              true,
		"fecha_agendada":          true,
		"estado":                  true,
		"categoria_id":            true,
		"nombre_usuario":          true,
		"email":                   true,
		"anio":                    true,
		"fecha_inicio":            true,
		"fecha_fin":               true,
	}

	// Limpiar el campo
	sort = strings.TrimSpace(strings.ToLower(sort))
	
	// Verificar si está en la lista blanca
	if allowedFields[sort] {
		return sort
	}

	// Si no está permitido, usar campo por defecto
	return "id"
}

// BuildSQLQuery construye la query SQL con paginación y filtros
func (p PaginationParams) BuildSQLQuery(baseQuery string, tableName string) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Agregar condiciones de búsqueda
	if p.Search != "" {
		searchConditions := []string{}
		searchFields := getSearchFields(tableName)
		
		for _, field := range searchFields {
			searchConditions = append(searchConditions, fmt.Sprintf("%s ILIKE $%d", field, argIndex))
			args = append(args, "%"+p.Search+"%")
			argIndex++
		}
		
		if len(searchConditions) > 0 {
			conditions = append(conditions, "("+strings.Join(searchConditions, " OR ")+")")
		}
	}

	// Agregar filtros específicos
	for field, value := range p.Filters {
		if value != nil && value != "" {
			conditions = append(conditions, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	// Construir query completa
	query := baseQuery
	
	// Agregar WHERE si hay condiciones
	if len(conditions) > 0 {
		if strings.Contains(strings.ToUpper(query), "WHERE") {
			query += " AND " + strings.Join(conditions, " AND ")
		} else {
			query += " WHERE " + strings.Join(conditions, " AND ")
		}
	}

	// Agregar ORDER BY
	query += fmt.Sprintf(" ORDER BY %s %s", p.Sort, strings.ToUpper(p.Order))

	// Agregar LIMIT y OFFSET
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, p.Limit, p.Offset)

	return query, args
}

// getSearchFields devuelve los campos de búsqueda para una tabla
func getSearchFields(tableName string) []string {
	searchFields := map[string][]string{
		"jugadores": {"nombre", "apellido", "club"},
		"usuarios":  {"nombre_usuario", "email"},
		"torneos":   {"nombre", "frase_destacada"},
		"categorias": {"nombre"},
		"partidos":  {"fase"},
		"noticias":  {"titulo", "contenido"},
	}

	if fields, exists := searchFields[tableName]; exists {
		return fields
	}

	// Campos por defecto
	return []string{"nombre"}
}

// BuildCountQuery construye la query para contar el total de registros
func (p PaginationParams) BuildCountQuery(baseQuery string, tableName string) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Agregar condiciones de búsqueda (mismo lógica que BuildSQLQuery)
	if p.Search != "" {
		searchConditions := []string{}
		searchFields := getSearchFields(tableName)
		
		for _, field := range searchFields {
			searchConditions = append(searchConditions, fmt.Sprintf("%s ILIKE $%d", field, argIndex))
			args = append(args, "%"+p.Search+"%")
			argIndex++
		}
		
		if len(searchConditions) > 0 {
			conditions = append(conditions, "("+strings.Join(searchConditions, " OR ")+")")
		}
	}

	// Agregar filtros específicos
	for field, value := range p.Filters {
		if value != nil && value != "" {
			conditions = append(conditions, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	// Construir query de conteo
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	
	// Agregar WHERE si hay condiciones
	if len(conditions) > 0 {
		countQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	return countQuery, args
}

// CalculatePaginationInfo calcula la información de paginación
func (p PaginationParams) CalculatePaginationInfo(total int) PaginationInfo {
	totalPages := int(math.Ceil(float64(total) / float64(p.Limit)))
	
	info := PaginationInfo{
		Page:       p.Page,
		Limit:      p.Limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    p.Page < totalPages,
		HasPrev:    p.Page > 1,
	}

	// Calcular página siguiente
	if info.HasNext {
		nextPage := p.Page + 1
		info.NextPage = &nextPage
	}

	// Calcular página anterior
	if info.HasPrev {
		prevPage := p.Page - 1
		info.PrevPage = &prevPage
	}

	return info
}

// PaginatedQuery ejecuta una query paginada
func PaginatedQuery(db *sql.DB, params PaginationParams, baseQuery string, tableName string) ([]map[string]interface{}, PaginationInfo, error) {
	// Obtener total de registros
	countQuery, countArgs := params.BuildCountQuery(baseQuery, tableName)
	var total int
	err := db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, PaginationInfo{}, fmt.Errorf("error counting records: %w", err)
	}

	// Calcular información de paginación
	paginationInfo := params.CalculatePaginationInfo(total)

	// Si no hay registros, devolver resultado vacío
	if total == 0 {
		return []map[string]interface{}{}, paginationInfo, nil
	}

	// Ejecutar query paginada
	query, args := params.BuildSQLQuery(baseQuery, tableName)
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, PaginationInfo{}, fmt.Errorf("error executing paginated query: %w", err)
	}
	defer rows.Close()

	// Obtener nombres de columnas
	columns, err := rows.Columns()
	if err != nil {
		return nil, PaginationInfo{}, fmt.Errorf("error getting columns: %w", err)
	}

	// Procesar resultados
	var results []map[string]interface{}
	for rows.Next() {
		// Crear slice de interfaces para escanear
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Escanear fila
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, PaginationInfo{}, fmt.Errorf("error scanning row: %w", err)
		}

		// Crear mapa de resultado
		result := make(map[string]interface{})
		for i, column := range columns {
			val := values[i]
			if val != nil {
				// Convertir []byte a string si es necesario
				if b, ok := val.([]byte); ok {
					result[column] = string(b)
				} else {
					result[column] = val
				}
			} else {
				result[column] = nil
			}
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, PaginationInfo{}, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, paginationInfo, nil
}

// CreatePaginatedResponse crea una respuesta paginada estándar
func CreatePaginatedResponse(data interface{}, pagination PaginationInfo, message string) PaginatedResponse {
	if message == "" {
		message = "Data retrieved successfully"
	}

	return PaginatedResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: pagination,
		Timestamp:  GetCurrentTimestamp(),
	}
}
