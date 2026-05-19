package common

import (
	"gorm.io/gorm"
)

// PaginationParams parámetros para paginación y búsqueda (compatible con frontend)
type PaginationParams struct {
	Page    int    `form:"page" binding:"min=1"`
	PerPage int    `form:"per_page" binding:"min=1,max=100"`
	Search  string `form:"search"`          // Búsqueda por texto
	Field   string `form:"field"`           // Campo a buscar (ej: name)
	SortBy  string `form:"sort_by"`         // Campo para ordenar
	SortDir string `form:"sort_dir"`        // Dirección: asc/desc
}

// DefaultPagination retorna valores por defecto
func DefaultPagination() PaginationParams {
	return PaginationParams{
		Page:    1,
		PerPage: 20,
		SortBy:  "id",
		SortDir: "desc",
	}
}

// Apply aplica filtros, búsqueda y ordenamiento a una query de GORM
func (p PaginationParams) Apply(db *gorm.DB, searchableFields []string, allowedSortFields []string) *gorm.DB {
	// Búsqueda (solo si hay término y campo permitido)
	if p.Search != "" && p.Field != "" {
		// Validar que el campo esté en la whitelist
		fieldAllowed := false
		for _, f := range searchableFields {
			if f == p.Field {
				fieldAllowed = true
				break
			}
		}
		if fieldAllowed {
			db = db.Where(p.Field+" LIKE ?", "%"+p.Search+"%")
		}
	}

	// Ordenamiento (solo campos permitidos)
	if p.SortBy != "" {
		sortAllowed := false
		for _, f := range allowedSortFields {
			if f == p.SortBy {
				sortAllowed = true
				break
			}
		}
		if sortAllowed {
			dir := "ASC"
			if p.SortDir == "desc" {
				dir = "DESC"
			}
			db = db.Order(p.SortBy + " " + dir)
		}
	}

	return db
}

// PaginationInfo estructura compatible con tu componente React
type PaginationInfo struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}

// PaginatedResponse formato estándar para respuestas paginadas (compatible con React)
type PaginatedResponse[T any] struct {
	Data       []T            `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
}

// NewPaginatedResponse crea una respuesta paginada con metadatos
func NewPaginatedResponse[T any](data []T, total, page, perPage int) PaginatedResponse[T] {
	lastPage := (total + perPage - 1) / perPage // ceil division
	return PaginatedResponse[T]{
		Data: data,
		Pagination: PaginationInfo{
			CurrentPage: page,
			PerPage:     perPage,
			Total:       total,
			LastPage:    lastPage,
		},
	}
}