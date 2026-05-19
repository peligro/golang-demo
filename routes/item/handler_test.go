package item

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
	db.AutoMigrate(&model.Item{})

	// Seed
	db.Create(&model.Item{Name: "crear_usuario"})
	db.Create(&model.Item{Name: "leer_usuario"})

	handler := NewHandler(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/items", nil)

	handler.Index(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ItemsResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, uint(2), response[0].ID) // id desc
	assert.Equal(t, "leer_usuario", response[0].Name)
}

// =============================================================================
// TestHandler_Show
// =============================================================================
func TestHandler_Show(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.Item{})
	db.Create(&model.Item{ID: 42, Name: "crear_usuario"})

	handler := NewHandler(db)

	t.Run("ID válido", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "42"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/items/42", nil)

		handler.Show(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "crear_usuario")
	})

	t.Run("ID inválido (string)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "abc"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/items/abc", nil)

		handler.Show(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "ID inválido")
	})

	t.Run("No encontrado", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/items/999", nil)

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
	db.AutoMigrate(&model.Item{})

	handler := NewHandler(db)

	t.Run("Creación exitosa", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/items",
			strings.NewReader(`{"name":"exportar_reporte"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "exportar_reporte")

		var count int64
		db.Model(&model.Item{}).Where("name = ?", "exportar_reporte").Count(&count)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Nombre vacío (validación)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/items",
			strings.NewReader(`{"name":""}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "obligatorio")
	})

	t.Run("Nombre muy corto (validación)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/items",
			strings.NewReader(`{"name":"ab"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "3 caracteres")
	})

	t.Run("Nombre duplicado", func(t *testing.T) {
		db.Create(&model.Item{Name: "existente_item"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/items",
			strings.NewReader(`{"name":"Existente_Item"}`)) // ← case-insensitive
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
	db.AutoMigrate(&model.Item{})
	db.Create(&model.Item{ID: 10, Name: "original_item"})

	handler := NewHandler(db)

	t.Run("Actualización exitosa", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "10"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/items/10",
			strings.NewReader(`{"name":"actualizado_item"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "actualizado_item")

		var item model.Item
		db.First(&item, 10)
		assert.Equal(t, "actualizado_item", item.Name)
	})

	t.Run("ID no existe", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/items/999",
			strings.NewReader(`{"name":"nuevo_item"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Update(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "no encontrado")
	})

	t.Run("Nombre duplicado (otro registro)", func(t *testing.T) {
		db.Create(&model.Item{Name: "otro_item"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "10"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/items/10",
			strings.NewReader(`{"name":"Otro_Item"}`)) // ← case-insensitive
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
	// Migrar también ProfileModuleItem para validar dependencias
	db.AutoMigrate(&model.Item{}, &model.ProfileModuleItem{})
	db.Create(&model.Item{ID: 20, Name: "para_borrar"})

	handler := NewHandler(db)

	t.Run("Eliminación exitosa", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "20"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/items/20", nil)

		handler.Delete(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "eliminado")

		var count int64
		db.Model(&model.Item{}).Where("id = ?", 20).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ID no existe", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/items/999", nil)

		handler.Delete(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Con dependencias (simulado)", func(t *testing.T) {
		// Seed: item + profile_module_item que lo referencia
		db.Create(&model.Item{ID: 30, Name: "con_dependencia"})
		db.Create(&model.ProfileModuleItem{ItemID: 30}) // ← Dependencia simulada

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "30"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/items/30", nil)

		handler.Delete(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "tiene registros asociados")
	})
}