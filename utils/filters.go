package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FilterOperator representa los operadores de filtro disponibles
type FilterOperator string

const (
	OpEqual              FilterOperator = "eq"
	OpNotEqual           FilterOperator = "ne"
	OpGreaterThan        FilterOperator = "gt"
	OpGreaterThanOrEqual FilterOperator = "gte"
	OpLessThan           FilterOperator = "lt"
	OpLessThanOrEqual    FilterOperator = "lte"
	OpLike               FilterOperator = "like"
	OpILike              FilterOperator = "ilike"
	OpIn                 FilterOperator = "in"
	OpNotIn              FilterOperator = "nin"
	OpIsNull             FilterOperator = "null"
	OpIsNotNull          FilterOperator = "notnull"
	OpBetween            FilterOperator = "between"
	OpContains           FilterOperator = "contains"
	OpStartsWith         FilterOperator = "startswith"
	OpEndsWith           FilterOperator = "endswith"
)

// FilterCondition representa una condición de filtro
type FilterCondition struct {
	Field    string         `json:"field"`
	Operator FilterOperator `json:"operator"`
	Value    interface{}    `json:"value"`
	Values   []interface{}  `json:"values,omitempty"` // Para operadores IN, BETWEEN
}

// FilterGroup representa un grupo de condiciones con lógica AND/OR
type FilterGroup struct {
	Logic      string            `json:"logic"` // "AND" o "OR"
	Conditions []FilterCondition `json:"conditions"`
	Groups     []FilterGroup     `json:"groups,omitempty"` // Grupos anidados
}

// AdvancedFilter representa un filtro avanzado completo
type AdvancedFilter struct {
	Groups []FilterGroup `json:"groups"`
}

// FilterConfig configuración para filtros por tabla
type FilterConfig struct {
	AllowedFields    map[string]FieldConfig `json:"allowed_fields"`
	DefaultOperators []FilterOperator       `json:"default_operators"`
	MaxConditions    int                    `json:"max_conditions"`
}

// FieldConfig configuración específica para un campo
type FieldConfig struct {
	Type             string           `json:"type"` // string, int, float, bool, date, datetime
	AllowedOperators []FilterOperator `json:"allowed_operators"`
	ValidationRegex  string           `json:"validation_regex,omitempty"`
}

// FilterManager maneja los filtros avanzados
type FilterManager struct {
	configs map[string]FilterConfig
}

// NewFilterManager crea un nuevo manager de filtros
func NewFilterManager() *FilterManager {
	fm := &FilterManager{
		configs: make(map[string]FilterConfig),
	}
	
	// Configurar filtros por defecto para las tablas principales
	fm.setupDefaultConfigs()
	
	return fm
}

// setupDefaultConfigs configura los filtros por defecto
func (fm *FilterManager) setupDefaultConfigs() {
	// Configuración para jugadores
	fm.configs["jugadores"] = FilterConfig{
		AllowedFields: map[string]FieldConfig{
			"id":                      {Type: "int", AllowedOperators: []FilterOperator{OpEqual, OpIn, OpNotIn}},
			"nombre":                  {Type: "string", AllowedOperators: []FilterOperator{OpEqual, OpLike, OpILike, OpStartsWith, OpEndsWith}},
			"apellido":                {Type: "string", AllowedOperators: []FilterOperator{OpEqual, OpLike, OpILike, OpStartsWith, OpEndsWith}},
			"categoria_id":            {Type: "int", AllowedOperators: []FilterOperator{OpEqual, OpIn, OpNotIn, OpIsNull}},
			"club":                    {Type: "string", AllowedOperators: []FilterOperator{OpEqual, OpLike, OpILike, OpIsNull}},
			"estado_participacion":    {Type: "string", AllowedOperators: []FilterOperator{OpEqual, OpIn, OpNotIn}},
			"contacto_visible_en_web": {Type: "bool", AllowedOperators: []FilterOperator{OpEqual}},
			"created_at":              {Type: "datetime", AllowedOperators: []FilterOperator{OpEqual, OpGreaterThan, OpLessThan, OpBetween}},
			"updated_at":              {Type: "datetime", AllowedOperators: []FilterOperator{OpEqual, OpGreaterThan, OpLessThan, OpBetween}},
		},
		MaxConditions: 10,
	}

	// Configuración para partidos
	fm.configs["partidos"] = FilterConfig{
		AllowedFields: map[string]FieldConfig{
			"id":           {Type: "int", AllowedOperators: []FilterOperator{OpEqual, OpIn, OpNotIn}},
			"torneo_id":    {Type: "int", AllowedOperators: []FilterOperator{OpEqual, OpIn, OpNotIn}},
			"categoria_id": {Type: "int", AllowedOperators: []FilterOperator{OpEqual, OpIn, OpNotIn}},
			"jugador1_id":  {Type: "int", AllowedOperators: []FilterOperator{OpEqual, OpIn, OpNotIn}},
			"jugador2_id":  {Type: "int", AllowedOperators: []FilterOperator{OpEqual, OpIn, OpNotIn}},
			"fase":         {Type: "string", AllowedOperators: []FilterOperator{OpEqual, OpIn, OpNotIn}},
			"estado":       {Type: "string", AllowedOperators: []FilterOperator{OpEqual, OpIn, OpNotIn}},
			"fecha_agendada": {Type: "date", AllowedOperators: []FilterOperator{OpEqual, OpGreaterThan, OpLessThan, OpBetween, OpIsNull}},
			"ganador_id":     {Type: "int", AllowedOperators: []FilterOperator{OpEqual, OpIn, OpNotIn, OpIsNull}},
			"resultado_aprobado": {Type: "bool", AllowedOperators: []FilterOperator{OpEqual}},
			"created_at":         {Type: "datetime", AllowedOperators: []FilterOperator{OpEqual, OpGreaterThan, OpLessThan, OpBetween}},
		},
		MaxConditions: 15,
	}

	// Configuración para usuarios
	fm.configs["usuarios"] = FilterConfig{
		AllowedFields: map[string]FieldConfig{
			"id":             {Type: "int", AllowedOperators: []FilterOperator{OpEqual, OpIn, OpNotIn}},
			"nombre_usuario": {Type: "string", AllowedOperators: []FilterOperator{OpEqual, OpLike, OpILike}},
			"email":          {Type: "string", AllowedOperators: []FilterOperator{OpEqual, OpLike, OpILike}},
			"rol":            {Type: "string", AllowedOperators: []FilterOperator{OpEqual, OpIn, OpNotIn}},
			"jugador_id":     {Type: "int", AllowedOperators: []FilterOperator{OpEqual, OpIn, OpNotIn, OpIsNull}},
			"created_at":     {Type: "datetime", AllowedOperators: []FilterOperator{OpEqual, OpGreaterThan, OpLessThan, OpBetween}},
		},
		MaxConditions: 8,
	}
}

// ParseAdvancedFilters parsea filtros avanzados desde query parameters
func (fm *FilterManager) ParseAdvancedFilters(query url.Values, tableName string) (AdvancedFilter, error) {
	config, exists := fm.configs[tableName]
	if !exists {
		return AdvancedFilter{}, fmt.Errorf("table %s not configured for filtering", tableName)
	}

	var filter AdvancedFilter
	var mainGroup FilterGroup
	mainGroup.Logic = "AND"

	conditionCount := 0

	// Parsear filtros simples (formato: field[operator]=value)
	for param, values := range query {
		if len(values) == 0 {
			continue
		}

		// Verificar si es un filtro (contiene corchetes)
		if !strings.Contains(param, "[") || !strings.Contains(param, "]") {
			continue
		}

		// Extraer campo y operador
		field, operator, err := fm.parseFilterParam(param)
		if err != nil {
			continue // Ignorar parámetros mal formados
		}

		// Verificar si el campo está permitido
		fieldConfig, allowed := config.AllowedFields[field]
		if !allowed {
			continue // Ignorar campos no permitidos
		}

		// Verificar si el operador está permitido
		if !fm.isOperatorAllowed(operator, fieldConfig.AllowedOperators) {
			continue // Ignorar operadores no permitidos
		}

		// Crear condición
		condition, err := fm.createCondition(field, operator, values[0], fieldConfig)
		if err != nil {
			continue // Ignorar condiciones mal formadas
		}

		mainGroup.Conditions = append(mainGroup.Conditions, condition)
		conditionCount++

		// Verificar límite de condiciones
		if conditionCount >= config.MaxConditions {
			break
		}
	}

	if len(mainGroup.Conditions) > 0 {
		filter.Groups = append(filter.Groups, mainGroup)
	}

	return filter, nil
}

// parseFilterParam parsea un parámetro de filtro (ej: "nombre[like]" -> "nombre", "like")
func (fm *FilterManager) parseFilterParam(param string) (string, FilterOperator, error) {
	re := regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\[([a-zA-Z]+)\]$`)
	matches := re.FindStringSubmatch(param)
	
	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid filter parameter format: %s", param)
	}

	field := matches[1]
	operator := FilterOperator(matches[2])

	return field, operator, nil
}

// isOperatorAllowed verifica si un operador está permitido
func (fm *FilterManager) isOperatorAllowed(operator FilterOperator, allowedOperators []FilterOperator) bool {
	for _, allowed := range allowedOperators {
		if operator == allowed {
			return true
		}
	}
	return false
}

// createCondition crea una condición de filtro
func (fm *FilterManager) createCondition(field string, operator FilterOperator, value string, fieldConfig FieldConfig) (FilterCondition, error) {
	condition := FilterCondition{
		Field:    field,
		Operator: operator,
	}

	// Validar y convertir valor según el tipo de campo
	convertedValue, err := fm.convertValue(value, fieldConfig.Type)
	if err != nil {
		return condition, fmt.Errorf("invalid value for field %s: %w", field, err)
	}

	// Manejar operadores especiales
	switch operator {
	case OpIn, OpNotIn:
		// Para operadores IN, dividir por comas
		values := strings.Split(value, ",")
		var convertedValues []interface{}
		for _, v := range values {
			converted, err := fm.convertValue(strings.TrimSpace(v), fieldConfig.Type)
			if err != nil {
				continue // Ignorar valores mal formados
			}
			convertedValues = append(convertedValues, converted)
		}
		condition.Values = convertedValues
	case OpBetween:
		// Para BETWEEN, esperar dos valores separados por coma
		values := strings.Split(value, ",")
		if len(values) != 2 {
			return condition, fmt.Errorf("BETWEEN operator requires exactly 2 values")
		}
		var convertedValues []interface{}
		for _, v := range values {
			converted, err := fm.convertValue(strings.TrimSpace(v), fieldConfig.Type)
			if err != nil {
				return condition, fmt.Errorf("invalid BETWEEN value: %w", err)
			}
			convertedValues = append(convertedValues, converted)
		}
		condition.Values = convertedValues
	case OpIsNull, OpIsNotNull:
		// Estos operadores no necesitan valor
		condition.Value = nil
	default:
		condition.Value = convertedValue
	}

	return condition, nil
}

// convertValue convierte un valor string al tipo apropiado
func (fm *FilterManager) convertValue(value string, fieldType string) (interface{}, error) {
	switch fieldType {
	case "int":
		return strconv.Atoi(value)
	case "float":
		return strconv.ParseFloat(value, 64)
	case "bool":
		return strconv.ParseBool(value)
	case "date":
		return time.Parse("2006-01-02", value)
	case "datetime":
		// Intentar varios formatos de fecha/hora
		formats := []string{
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05",
			"2006-01-02",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, value); err == nil {
				return t, nil
			}
		}
		return nil, fmt.Errorf("invalid datetime format: %s", value)
	default:
		return value, nil
	}
}

// BuildSQLConditions construye las condiciones SQL para el filtro
func (fm *FilterManager) BuildSQLConditions(filter AdvancedFilter) (string, []interface{}, error) {
	if len(filter.Groups) == 0 {
		return "", nil, nil
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	for i, group := range filter.Groups {
		groupCondition, groupArgs, newArgIndex, err := fm.buildGroupConditions(group, argIndex)
		if err != nil {
			return "", nil, fmt.Errorf("error building group %d conditions: %w", i, err)
		}

		if groupCondition != "" {
			conditions = append(conditions, "("+groupCondition+")")
			args = append(args, groupArgs...)
			argIndex = newArgIndex
		}
	}

	if len(conditions) == 0 {
		return "", nil, nil
	}

	// Unir grupos con AND (por defecto)
	finalCondition := strings.Join(conditions, " AND ")
	return finalCondition, args, nil
}

// buildGroupConditions construye las condiciones SQL para un grupo
func (fm *FilterManager) buildGroupConditions(group FilterGroup, startArgIndex int) (string, []interface{}, int, error) {
	var conditions []string
	var args []interface{}
	argIndex := startArgIndex

	// Procesar condiciones del grupo
	for _, condition := range group.Conditions {
		conditionSQL, conditionArgs, newArgIndex, err := fm.buildConditionSQL(condition, argIndex)
		if err != nil {
			return "", nil, argIndex, fmt.Errorf("error building condition for field %s: %w", condition.Field, err)
		}

		if conditionSQL != "" {
			conditions = append(conditions, conditionSQL)
			args = append(args, conditionArgs...)
			argIndex = newArgIndex
		}
	}

	// Procesar grupos anidados
	for _, nestedGroup := range group.Groups {
		nestedCondition, nestedArgs, newArgIndex, err := fm.buildGroupConditions(nestedGroup, argIndex)
		if err != nil {
			return "", nil, argIndex, fmt.Errorf("error building nested group: %w", err)
		}

		if nestedCondition != "" {
			conditions = append(conditions, "("+nestedCondition+")")
			args = append(args, nestedArgs...)
			argIndex = newArgIndex
		}
	}

	if len(conditions) == 0 {
		return "", nil, argIndex, nil
	}

	// Unir condiciones con la lógica del grupo
	logic := "AND"
	if group.Logic == "OR" {
		logic = "OR"
	}

	groupCondition := strings.Join(conditions, " "+logic+" ")
	return groupCondition, args, argIndex, nil
}

// buildConditionSQL construye la SQL para una condición específica
func (fm *FilterManager) buildConditionSQL(condition FilterCondition, startArgIndex int) (string, []interface{}, int, error) {
	field := condition.Field
	argIndex := startArgIndex

	switch condition.Operator {
	case OpEqual:
		return fmt.Sprintf("%s = $%d", field, argIndex), []interface{}{condition.Value}, argIndex + 1, nil
	case OpNotEqual:
		return fmt.Sprintf("%s != $%d", field, argIndex), []interface{}{condition.Value}, argIndex + 1, nil
	case OpGreaterThan:
		return fmt.Sprintf("%s > $%d", field, argIndex), []interface{}{condition.Value}, argIndex + 1, nil
	case OpGreaterThanOrEqual:
		return fmt.Sprintf("%s >= $%d", field, argIndex), []interface{}{condition.Value}, argIndex + 1, nil
	case OpLessThan:
		return fmt.Sprintf("%s < $%d", field, argIndex), []interface{}{condition.Value}, argIndex + 1, nil
	case OpLessThanOrEqual:
		return fmt.Sprintf("%s <= $%d", field, argIndex), []interface{}{condition.Value}, argIndex + 1, nil
	case OpLike:
		return fmt.Sprintf("%s LIKE $%d", field, argIndex), []interface{}{condition.Value}, argIndex + 1, nil
	case OpILike:
		return fmt.Sprintf("%s ILIKE $%d", field, argIndex), []interface{}{condition.Value}, argIndex + 1, nil
	case OpStartsWith:
		return fmt.Sprintf("%s ILIKE $%d", field, argIndex), []interface{}{fmt.Sprintf("%s%%", condition.Value)}, argIndex + 1, nil
	case OpEndsWith:
		return fmt.Sprintf("%s ILIKE $%d", field, argIndex), []interface{}{fmt.Sprintf("%%%s", condition.Value)}, argIndex + 1, nil
	case OpContains:
		return fmt.Sprintf("%s ILIKE $%d", field, argIndex), []interface{}{fmt.Sprintf("%%%s%%", condition.Value)}, argIndex + 1, nil
	case OpIsNull:
		return fmt.Sprintf("%s IS NULL", field), []interface{}{}, argIndex, nil
	case OpIsNotNull:
		return fmt.Sprintf("%s IS NOT NULL", field), []interface{}{}, argIndex, nil
	case OpIn:
		if len(condition.Values) == 0 {
			return "", nil, argIndex, fmt.Errorf("IN operator requires at least one value")
		}
		placeholders := make([]string, len(condition.Values))
		args := make([]interface{}, len(condition.Values))
		for i, value := range condition.Values {
			placeholders[i] = fmt.Sprintf("$%d", argIndex+i)
			args[i] = value
		}
		sql := fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", "))
		return sql, args, argIndex + len(condition.Values), nil
	case OpNotIn:
		if len(condition.Values) == 0 {
			return "", nil, argIndex, fmt.Errorf("NOT IN operator requires at least one value")
		}
		placeholders := make([]string, len(condition.Values))
		args := make([]interface{}, len(condition.Values))
		for i, value := range condition.Values {
			placeholders[i] = fmt.Sprintf("$%d", argIndex+i)
			args[i] = value
		}
		sql := fmt.Sprintf("%s NOT IN (%s)", field, strings.Join(placeholders, ", "))
		return sql, args, argIndex + len(condition.Values), nil
	case OpBetween:
		if len(condition.Values) != 2 {
			return "", nil, argIndex, fmt.Errorf("BETWEEN operator requires exactly 2 values")
		}
		sql := fmt.Sprintf("%s BETWEEN $%d AND $%d", field, argIndex, argIndex+1)
		return sql, condition.Values, argIndex + 2, nil
	default:
		return "", nil, argIndex, fmt.Errorf("unsupported operator: %s", condition.Operator)
	}
}

// GetFilterConfig obtiene la configuración de filtros para una tabla
func (fm *FilterManager) GetFilterConfig(tableName string) (FilterConfig, bool) {
	config, exists := fm.configs[tableName]
	return config, exists
}
