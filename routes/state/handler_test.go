package state

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
	db.AutoMigrate(&model.State{})

	// Seed
	db.Create(&model.State{Name: "Activo"})
	db.Create(&model.State{Name: "Inactivo"})

	handler := NewHandler(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/states", nil)

	handler.Index(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.StatesResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, uint(2), response[0].ID) // id desc
	assert.Equal(t, "Inactivo", response[0].Name)
}

// =============================================================================
// TestHandler_Show
// =============================================================================
func TestHandler_Show(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.State{})
	db.Create(&model.State{ID: 42, Name: "Activo"})

	handler := NewHandler(db)

	t.Run("ID válido", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "42"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/states/42", nil)

		handler.Show(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Activo")
	})

	t.Run("ID inválido (string)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "abc"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/states/abc", nil)

		handler.Show(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "ID inválido")
	})

	t.Run("No encontrado", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/states/999", nil)

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
	db.AutoMigrate(&model.State{})

	handler := NewHandler(db)

	t.Run("Creación exitosa", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/states",
			strings.NewReader(`{"name":"Pendiente"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Pendiente")

		// Verificar que se guardó en DB
		var count int64
		db.Model(&model.State{}).Where("name = ?", "Pendiente").Count(&count)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Nombre vacío (validación)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/states",
			strings.NewReader(`{"name":""}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "obligatorio")
	})

	t.Run("Nombre duplicado", func(t *testing.T) {
		db.Create(&model.State{Name: "Existente"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/states",
			strings.NewReader(`{"name":"Existente"}`))
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
	db.AutoMigrate(&model.State{})
	db.Create(&model.State{ID: 10, Name: "Original"})

	handler := NewHandler(db)

	t.Run("Actualización exitosa", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "10"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/states/10",
			strings.NewReader(`{"name":"Modificado"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Modificado")

		// Verificar cambio en DB
		var state model.State
		db.First(&state, 10)
		assert.Equal(t, "Modificado", state.Name)
	})

	t.Run("ID no existe", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/states/999",
			strings.NewReader(`{"name":"Nuevo"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Update(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Nombre duplicado (otro registro)", func(t *testing.T) {
		db.Create(&model.State{Name: "Otro"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "10"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/states/10",
			strings.NewReader(`{"name":"Otro"}`))
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
	// Migrar también UserMetadata para que funcione la validación de dependencias
	db.AutoMigrate(&model.State{}, &model.UserMetadata{})
	db.Create(&model.State{ID: 20, Name: "ParaBorrar"})

	handler := NewHandler(db)

	t.Run("Eliminación exitosa", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "20"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/states/20", nil)

		handler.Delete(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "eliminado")

		// Verificar que ya no existe
		var count int64
		db.Model(&model.State{}).Where("id = ?", 20).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ID no existe", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/states/999", nil)

		handler.Delete(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Con dependencias (simulado)", func(t *testing.T) {
		// Seed: estado + metadata que lo referencia
		db.Create(&model.State{ID: 30, Name: "ConDependencia"})
		db.Create(&model.UserMetadata{State: 30}) // ← Dependencia simulada

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "30"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/states/30", nil)

		handler.Delete(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "tiene registros asociados")
	})
}