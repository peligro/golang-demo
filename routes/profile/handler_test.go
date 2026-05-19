package profile

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
	db.AutoMigrate(&model.Profile{})

	// Seed
	db.Create(&model.Profile{Name: "Administrador", Description: "Acceso completo al sistema"})
	db.Create(&model.Profile{Name: "Editor", Description: "Puede editar contenido"})
	db.Create(&model.Profile{Name: "Cliente", Description: "Acceso limitado"})

	handler := NewHandler(db)

	t.Run("Listado sin filtros (paginación por defecto)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/profiles", nil)

		handler.Index(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.PaginatedResponse[dto.ProfileResponse]
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
		c.Request = httptest.NewRequest(http.MethodGet, "/profiles?search=Admin&field=name", nil)

		handler.Index(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.PaginatedResponse[dto.ProfileResponse]
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, "Administrador", response.Data[0].Name)
	})

	t.Run("Paginación: page=1, per_page=2", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/profiles?page=1&per_page=2", nil)

		handler.Index(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.PaginatedResponse[dto.ProfileResponse]
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
		c.Request = httptest.NewRequest(http.MethodGet, "/profiles?sort_by=name&sort_dir=asc", nil)

		handler.Index(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.PaginatedResponse[dto.ProfileResponse]
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		// "Administrador" debería ser primero alfabéticamente
		assert.Equal(t, "Administrador", response.Data[0].Name)
	})

	t.Run("Búsqueda con campo no permitido (debería ignorar search)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/profiles?search=Admin&field=description_invalid", nil)

		handler.Index(c)

		assert.Equal(t, http.StatusOK, w.Code)
		// Debería retornar todos los registros porque el campo no está en whitelist
		var response common.PaginatedResponse[dto.ProfileResponse]
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
	db.AutoMigrate(&model.Profile{})
	db.Create(&model.Profile{ID: 42, Name: "Administrador", Description: "Acceso completo"})

	handler := NewHandler(db)

	t.Run("ID válido", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "42"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/profiles/42", nil)

		handler.Show(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Administrador")
	})

	t.Run("ID inválido (string)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "abc"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/profiles/abc", nil)

		handler.Show(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "ID inválido")
	})

	t.Run("No encontrado", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/profiles/999", nil)

		handler.Show(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "no encontrado")
	})
}

// =============================================================================
// TestHandler_Create
// =============================================================================
func TestHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.Profile{})

	handler := NewHandler(db)

	t.Run("Creación exitosa", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/profiles",
			strings.NewReader(`{"name":"Soporte","description":"Acceso para equipo de soporte técnico"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Soporte")

		var count int64
		db.Model(&model.Profile{}).Where("name = ?", "Soporte").Count(&count)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Nombre vacío (validación)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/profiles",
			strings.NewReader(`{"name":"","description":"Descripción válida de 25 caracteres"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "obligatorio")
	})

	t.Run("Descripción muy corta (validación)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/profiles",
			strings.NewReader(`{"name":"Test","description":"corto"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "10 caracteres")
	})

	t.Run("Nombre duplicado", func(t *testing.T) {
		db.Create(&model.Profile{Name: "Existente", Description: "Descripción válida de 30 caracteres"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/profiles",
			strings.NewReader(`{"name":"Existente","description":"Otra descripción válida de 35 caracteres"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "duplicado")
	})
}

// =============================================================================
// TestHandler_Update
// =============================================================================
func TestHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.Profile{})
	db.Create(&model.Profile{ID: 10, Name: "Original", Description: "Descripción original válida de 30 caracteres"})

	handler := NewHandler(db)

	t.Run("Actualización exitosa", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "10"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/profiles/10",
			strings.NewReader(`{"name":"Modificado","description":"Nueva descripción válida de 40 caracteres"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Modificado")

		var profile model.Profile
		db.First(&profile, 10)
		assert.Equal(t, "Modificado", profile.Name)
		assert.Equal(t, "Nueva descripción válida de 40 caracteres", profile.Description)
	})

	t.Run("ID no existe", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/profiles/999",
			strings.NewReader(`{"name":"Nuevo","description":"Descripción válida de 25 caracteres"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Update(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "no encontrado")
	})

	t.Run("Nombre duplicado (otro registro)", func(t *testing.T) {
		db.Create(&model.Profile{Name: "Otro", Description: "Descripción de otro perfil válida de 35 caracteres"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "10"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/profiles/10",
			strings.NewReader(`{"name":"Otro","description":"Descripción actualizada válida de 40 caracteres"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Update(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "duplicado")
	})
}

// =============================================================================
// TestHandler_Delete
// =============================================================================
func TestHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.Profile{}, &model.ProfileModule{})
	db.Create(&model.Profile{ID: 20, Name: "ParaBorrar", Description: "Desc"})

	handler := NewHandler(db)

	t.Run("Eliminación exitosa", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "20"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/profiles/20", nil)

		handler.Delete(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "eliminado")

		var count int64
		db.Model(&model.Profile{}).Where("id = ?", 20).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ID no existe", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/profiles/999", nil)

		handler.Delete(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Con dependencias (simulado)", func(t *testing.T) {
		db.Create(&model.Profile{ID: 30, Name: "ConDependencia", Description: "Desc"})
		db.Create(&model.ProfileModule{ProfileID: 30})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "30"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/profiles/30", nil)

		handler.Delete(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "tiene registros asociados")
	})
}