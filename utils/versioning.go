package utils

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// APIVersion representa una versión de la API
type APIVersion struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Build      string
}

// String devuelve la representación string de la versión
func (v APIVersion) String() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		version += "-" + v.PreRelease
	}
	if v.Build != "" {
		version += "+" + v.Build
	}
	return version
}

// IsCompatible verifica si esta versión es compatible con otra
func (v APIVersion) IsCompatible(other APIVersion) bool {
	// Compatibilidad hacia atrás dentro de la misma versión major
	return v.Major == other.Major && v.Minor >= other.Minor
}

// VersionManager maneja el versionado de la API
type VersionManager struct {
	currentVersion    APIVersion
	supportedVersions []APIVersion
	deprecatedVersions map[string]string // version -> deprecation message
}

// NewVersionManager crea un nuevo manager de versiones
func NewVersionManager() *VersionManager {
	return &VersionManager{
		currentVersion: APIVersion{Major: 1, Minor: 0, Patch: 0},
		supportedVersions: []APIVersion{
			{Major: 1, Minor: 0, Patch: 0},
		},
		deprecatedVersions: make(map[string]string),
	}
}

// ParseVersion parsea una string de versión
func (vm *VersionManager) ParseVersion(versionStr string) (APIVersion, error) {
	// Regex para parsear versiones semánticas
	re := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
	matches := re.FindStringSubmatch(versionStr)
	
	if len(matches) < 4 {
		return APIVersion{}, fmt.Errorf("invalid version format: %s", versionStr)
	}
	
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	
	version := APIVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
	
	if len(matches) > 4 && matches[4] != "" {
		version.PreRelease = matches[4]
	}
	
	if len(matches) > 5 && matches[5] != "" {
		version.Build = matches[5]
	}
	
	return version, nil
}

// ExtractVersionFromRequest extrae la versión de la request
func (vm *VersionManager) ExtractVersionFromRequest(r *http.Request) (APIVersion, error) {
	// Prioridad: Header > Query Parameter > URL Path > Default
	
	// 1. Verificar header Accept-Version
	if acceptVersion := r.Header.Get("Accept-Version"); acceptVersion != "" {
		return vm.ParseVersion(acceptVersion)
	}
	
	// 2. Verificar header API-Version
	if apiVersion := r.Header.Get("API-Version"); apiVersion != "" {
		return vm.ParseVersion(apiVersion)
	}
	
	// 3. Verificar query parameter
	if version := r.URL.Query().Get("version"); version != "" {
		return vm.ParseVersion(version)
	}
	
	// 4. Extraer de la URL path
	if version := vm.extractVersionFromPath(r.URL.Path); version != "" {
		return vm.ParseVersion(version)
	}
	
	// 5. Usar versión por defecto
	return vm.currentVersion, nil
}

// extractVersionFromPath extrae la versión del path de la URL
func (vm *VersionManager) extractVersionFromPath(path string) string {
	// Buscar patrones como /api/v1/ o /v1.2/
	re := regexp.MustCompile(`/v?(\d+(?:\.\d+(?:\.\d+)?)?)/`)
	matches := re.FindStringSubmatch(path)
	
	if len(matches) > 1 {
		return matches[1]
	}
	
	return ""
}

// IsVersionSupported verifica si una versión está soportada
func (vm *VersionManager) IsVersionSupported(version APIVersion) bool {
	for _, supported := range vm.supportedVersions {
		if version.Major == supported.Major && 
		   version.Minor == supported.Minor && 
		   version.Patch == supported.Patch {
			return true
		}
	}
	return false
}

// IsVersionDeprecated verifica si una versión está deprecada
func (vm *VersionManager) IsVersionDeprecated(version APIVersion) (bool, string) {
	versionStr := version.String()
	message, deprecated := vm.deprecatedVersions[versionStr]
	return deprecated, message
}

// AddSupportedVersion agrega una versión soportada
func (vm *VersionManager) AddSupportedVersion(version APIVersion) {
	vm.supportedVersions = append(vm.supportedVersions, version)
}

// DeprecateVersion marca una versión como deprecada
func (vm *VersionManager) DeprecateVersion(version APIVersion, message string) {
	vm.deprecatedVersions[version.String()] = message
}

// VersionMiddleware middleware para manejo de versiones
func (vm *VersionManager) VersionMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extraer versión de la request
			requestedVersion, err := vm.ExtractVersionFromRequest(r)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid API version: %s", err.Error()), http.StatusBadRequest)
				return
			}
			
			// Verificar si la versión está soportada
			if !vm.IsVersionSupported(requestedVersion) {
				errorResponse := map[string]interface{}{
					"success": false,
					"message": "API version not supported",
					"error": map[string]interface{}{
						"code": "UNSUPPORTED_VERSION",
						"details": map[string]interface{}{
							"requested_version": requestedVersion.String(),
							"supported_versions": vm.GetSupportedVersionsStrings(),
							"current_version": vm.currentVersion.String(),
						},
					},
				}
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				WriteJSONResponse(w, errorResponse)
				return
			}
			
			// Verificar si la versión está deprecada
			if deprecated, message := vm.IsVersionDeprecated(requestedVersion); deprecated {
				w.Header().Set("Deprecation", "true")
				w.Header().Set("Sunset", message)
				w.Header().Set("Link", fmt.Sprintf("</api/v%d>; rel=\"successor-version\"", vm.currentVersion.Major))
			}
			
			// Agregar headers de versión a la respuesta
			w.Header().Set("API-Version", requestedVersion.String())
			w.Header().Set("API-Current-Version", vm.currentVersion.String())
			
			// Agregar versión al contexto de la request
			r = r.WithContext(SetVersionInContext(r.Context(), requestedVersion))
			
			next.ServeHTTP(w, r)
		})
	}
}

// GetSupportedVersionsStrings devuelve las versiones soportadas como strings
func (vm *VersionManager) GetSupportedVersionsStrings() []string {
	versions := make([]string, len(vm.supportedVersions))
	for i, version := range vm.supportedVersions {
		versions[i] = version.String()
	}
	return versions
}

// VersionedRouter crea un router con soporte para versiones
type VersionedRouter struct {
	routers map[string]*mux.Router
	vm      *VersionManager
}

// NewVersionedRouter crea un nuevo router versionado
func NewVersionedRouter(vm *VersionManager) *VersionedRouter {
	return &VersionedRouter{
		routers: make(map[string]*mux.Router),
		vm:      vm,
	}
}

// GetRouter obtiene el router para una versión específica
func (vr *VersionedRouter) GetRouter(version APIVersion) *mux.Router {
	versionStr := version.String()
	if router, exists := vr.routers[versionStr]; exists {
		return router
	}
	
	// Crear nuevo router para esta versión
	router := mux.NewRouter()
	vr.routers[versionStr] = router
	return router
}

// RegisterVersionedHandler registra un handler para una versión específica
func (vr *VersionedRouter) RegisterVersionedHandler(version APIVersion, path string, handler http.HandlerFunc, methods ...string) {
	router := vr.GetRouter(version)
	route := router.HandleFunc(path, handler)
	if len(methods) > 0 {
		route.Methods(methods...)
	}
}

// ServeHTTP implementa http.Handler
func (vr *VersionedRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extraer versión de la request
	version, err := vr.vm.ExtractVersionFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid API version", http.StatusBadRequest)
		return
	}
	
	// Obtener router para la versión
	router := vr.GetRouter(version)
	router.ServeHTTP(w, r)
}

// ContentNegotiationMiddleware middleware para negociación de contenido
func ContentNegotiationMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verificar Accept header
			acceptHeader := r.Header.Get("Accept")
			
			// Si no especifica Accept, usar JSON por defecto
			if acceptHeader == "" {
				r.Header.Set("Accept", "application/json")
			}
			
			// Verificar si acepta JSON
			if !strings.Contains(acceptHeader, "application/json") && 
			   !strings.Contains(acceptHeader, "*/*") {
				errorResponse := map[string]interface{}{
					"success": false,
					"message": "Content type not acceptable",
					"error": map[string]interface{}{
						"code": "NOT_ACCEPTABLE",
						"details": map[string]interface{}{
							"accepted_types": []string{"application/json"},
							"requested_type": acceptHeader,
						},
					},
				}
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotAcceptable)
				WriteJSONResponse(w, errorResponse)
				return
			}
			
			// Establecer Content-Type por defecto
			w.Header().Set("Content-Type", "application/json")
			
			next.ServeHTTP(w, r)
		})
	}
}

// Clave para el contexto de versión
type versionContextKey string

const VersionKey versionContextKey = "api_version"

// GetVersionFromContext obtiene la versión del contexto
func GetVersionFromContext(ctx context.Context) (APIVersion, bool) {
	version, ok := ctx.Value(VersionKey).(APIVersion)
	return version, ok
}

// SetVersionInContext establece la versión en el contexto
func SetVersionInContext(ctx context.Context, version APIVersion) context.Context {
	return context.WithValue(ctx, VersionKey, version)
}
