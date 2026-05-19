package module

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/peligro/golang-demo/common"
	"github.com/peligro/golang-demo/dto"
	"github.com/peligro/golang-demo/model"
)

// =============================================================================
// TestHandler_Index
// =============================================================================
func TestHandler_Index(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.Module{})

	// Seed con slugs tipo path
	db.Create(&model.Module{Name: "Usuarios", Slug: "/setting/users"})
	db.Create(&model.Module{Name: "Reportes", Slug: "/setting/reports"})
	db.Create(&model.Module{Name: "Configuración", Slug: "/setting/config"})

	handler := NewHandler(db)

	t.Run("Listado sin filtros (paginación por defecto)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/modules", nil)

		handler.Index(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.PaginatedResponse[dto.ModuleResponse]
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		// Validar estructura compatible con React
		assert.Len(t, response.Data, 3)
		assert.Equal(t, 3, response.Pagination.Total)
		assert.Equal(t, 1, response.Pagination.CurrentPage)
		assert.Equal(t, 20, response.Pagination.PerPage)
		assert.Equal(t, 1, response.Pagination.LastPage) // ceil(3/20) = 1
	})

	t.Run("Búsqueda por nombre con field=name", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// ← Tu componente usa ?field=nombreModulo, aquí field=name
		c.Request = httptest.NewRequest(http.MethodGet, "/modules?search=Report&field=name", nil)

		handler.Index(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.PaginatedResponse[dto.ModuleResponse]
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, "Reportes", response.Data[0].Name)
		assert.Equal(t, "/setting/reports", response.Data[0].Slug)
	})

	t.Run("Paginación: page=1, per_page=2", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/modules?page=1&per_page=2", nil)

		handler.Index(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.PaginatedResponse[dto.ModuleResponse]
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 2)
		assert.Equal(t, 3, response.Pagination.Total)
		assert.Equal(t, 2, response.Pagination.LastPage) // ceil(3/2) = 2
		assert.Equal(t, 1, response.Pagination.CurrentPage)
	})

	t.Run("Ordenamiento por nombre ASC", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/modules?sort_by=name&sort_dir=asc", nil)

		handler.Index(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.PaginatedResponse[dto.ModuleResponse]
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		// "Configuración" debería ser primero alfabéticamente
		assert.Equal(t, "Configuración", response.Data[0].Name)
	})

	t.Run("Búsqueda con campo no permitido (debería ignorar search)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/modules?search=Usuarios&field=slug_invalido", nil)

		handler.Index(c)

		assert.Equal(t, http.StatusOK, w.Code)
		// Debería retornar todos los registros porque el campo no está en whitelist
		var response common.PaginatedResponse[dto.ModuleResponse]
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 3, response.Pagination.Total)
	})
}

// =============================================================================
// TestHandler_Show
// =============================================================================
func TestHandler_Show(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.Module{})
	db.Create(&model.Module{ID: 42, Name: "Usuarios", Slug: "/setting/users"})

	handler := NewHandler(db)

	t.Run("ID válido", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "42"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/modules/42", nil)

		handler.Show(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response dto.ModuleResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "/setting/users", response.Slug)
	})

	t.Run("ID inválido (string)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "abc"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/modules/abc", nil)

		handler.Show(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "ID inválido")
	})

	t.Run("No encontrado", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/modules/999", nil)

		handler.Show(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// =============================================================================
// TestHandler_Create
// =============================================================================
func TestHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.Module{})

	handler := NewHandler(db)

	t.Run("Creación exitosa con slug path", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/modules",
			strings.NewReader(`{"name":"Seguridad","slug":"/setting/security"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response dto.ModuleResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "/setting/security", response.Slug)

		var count int64
		db.Model(&model.Module{}).Where("slug = ?", "/setting/security").Count(&count)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Slug inválido (no comienza con /)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/modules",
			strings.NewReader(`{"name":"Test","slug":"invalid-path"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "debe comenzar con /")
	})

	t.Run("Nombre duplicado", func(t *testing.T) {
		db.Create(&model.Module{Name: "Existente", Slug: "/setting/existente"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/modules",
			strings.NewReader(`{"name":"Existente","slug":"/setting/otro"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "duplicado")
	})

	t.Run("Slug duplicado", func(t *testing.T) {
		db.Create(&model.Module{Name: "Otro", Slug: "/setting/duplicado"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/modules",
			strings.NewReader(`{"name":"Nuevo","slug":"/setting/duplicado"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "path ya está en uso")
	})
}

// =============================================================================
// TestHandler_Update
// =============================================================================
func TestHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.Module{})
	db.Create(&model.Module{ID: 10, Name: "Original", Slug: "/setting/original"})

	handler := NewHandler(db)

	t.Run("Actualización exitosa de nombre y slug", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "10"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/modules/10",
			strings.NewReader(`{"name":"Modificado","slug":"/setting/modified"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response dto.ModuleResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "/setting/modified", response.Slug)

		var module model.Module
		db.First(&module, 10)
		assert.Equal(t, "Modificado", module.Name)
		assert.Equal(t, "/setting/modified", module.Slug)
	})

	t.Run("Slug duplicado en otro registro", func(t *testing.T) {
		db.Create(&model.Module{Name: "Otro", Slug: "/setting/otro"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "10"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/modules/10",
			strings.NewReader(`{"name":"Actualizado","slug":"/setting/otro"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Update(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "path ya está en uso")
	})
}

// =============================================================================
// TestHandler_Delete
// =============================================================================
func TestHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.Module{}, &model.ProfileModule{})
	db.Create(&model.Module{ID: 20, Name: "ParaBorrar", Slug: "/setting/todelete"})

	handler := NewHandler(db)

	t.Run("Eliminación exitosa", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "20"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/modules/20", nil)

		handler.Delete(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var count int64
		db.Model(&model.Module{}).Where("id = ?", 20).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Con dependencias (profile_module)", func(t *testing.T) {
		db.Create(&model.Module{ID: 30, Name: "ConDependencia", Slug: "/setting/withdep"})
		db.Create(&model.ProfileModule{ModuleID: 30})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "30"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/modules/30", nil)

		handler.Delete(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "tiene registros asociados")
	})
}