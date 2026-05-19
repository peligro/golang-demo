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

	// Seed
	db.Create(&model.Module{Name: "Usuarios", Description: "Gestión de usuarios"})
	db.Create(&model.Module{Name: "Reportes", Description: "Generación de reportes"})

	handler := NewHandler(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/modules", nil)

	handler.Index(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ModulesResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, uint(2), response[0].ID) // id desc
	assert.Equal(t, "Reportes", response[0].Name)
}

// =============================================================================
// TestHandler_Show
// =============================================================================
func TestHandler_Show(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.Module{})
	db.Create(&model.Module{ID: 42, Name: "Usuarios", Description: "Gestión de usuarios"})

	handler := NewHandler(db)

	t.Run("ID válido", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "42"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/modules/42", nil)

		handler.Show(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Usuarios")
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
	db.AutoMigrate(&model.Module{})

	handler := NewHandler(db)

	t.Run("Creación exitosa", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/modules",
			strings.NewReader(`{"name":"Seguridad","description":"Gestión de permisos y accesos"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Seguridad")

		var count int64
		db.Model(&model.Module{}).Where("name = ?", "Seguridad").Count(&count)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Nombre vacío (validación)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/modules",
			strings.NewReader(`{"name":"","description":"Descripción válida de 20 caracteres"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "obligatorio")
	})

	t.Run("Descripción muy corta (validación)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/modules",
			strings.NewReader(`{"name":"Test","description":"corto"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "10 caracteres")
	})

	t.Run("Nombre duplicado", func(t *testing.T) {
		// Seed previo con descripción válida
		db.Create(&model.Module{Name: "Existente", Description: "Descripción válida de 25 caracteres"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/modules",
			strings.NewReader(`{"name":"Existente","description":"Otra descripción válida de 30 caracteres"}`))
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
	db.AutoMigrate(&model.Module{})
	db.Create(&model.Module{ID: 10, Name: "Original", Description: "Descripción original válida de 30 caracteres"})

	handler := NewHandler(db)

	t.Run("Actualización exitosa", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "10"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/modules/10",
			strings.NewReader(`{"name":"Modificado","description":"Nueva descripción válida de 35 caracteres"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Modificado")

		var module model.Module
		db.First(&module, 10)
		assert.Equal(t, "Modificado", module.Name)
		assert.Equal(t, "Nueva descripción válida de 35 caracteres", module.Description)
	})

	t.Run("ID no existe", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/modules/999",
			strings.NewReader(`{"name":"Nuevo","description":"Descripción válida de 25 caracteres"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Update(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "no encontrado")
	})

	t.Run("Nombre duplicado (otro registro)", func(t *testing.T) {
		db.Create(&model.Module{Name: "Otro", Description: "Descripción de otro módulo válida de 35 caracteres"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "10"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/modules/10",
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
	// Migrar también ProfileModule para validar dependencias
	db.AutoMigrate(&model.Module{}, &model.ProfileModule{})
	db.Create(&model.Module{ID: 20, Name: "ParaBorrar", Description: "Desc"})

	handler := NewHandler(db)

	t.Run("Eliminación exitosa", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "20"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/modules/20", nil)

		handler.Delete(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "eliminado")

		// Verificar que ya no existe
		var count int64
		db.Model(&model.Module{}).Where("id = ?", 20).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ID no existe", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/modules/999", nil)

		handler.Delete(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Con dependencias (simulado)", func(t *testing.T) {
		// Seed: módulo + profile_module que lo referencia
		db.Create(&model.Module{ID: 30, Name: "ConDependencia", Description: "Desc"})
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