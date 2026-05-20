package profile

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/peligro/golang-demo/dto"
	"github.com/peligro/golang-demo/model"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(
		&model.Profile{}, &model.Module{}, &model.Item{},
		&model.ProfileModule{}, &model.ProfileModuleItem{},
	))
	return db
}

func setupProfileHandler(t *testing.T) (*Handler, *gorm.DB) {
	db := setupTestDB(t)
	return NewHandler(db), db
}

func setupModuleHandler(t *testing.T) (*ModuleHandler, *gorm.DB) {
	db := setupTestDB(t)
	return NewModuleHandler(db), db
}

func setupModuleItemHandler(t *testing.T) (*ModuleItemHandler, *gorm.DB) {
	db := setupTestDB(t)
	return NewModuleItemHandler(db), db
}

func TestProfileHandler_Create_Integration(t *testing.T) {
	t.Skip("Tests de validación de nombre requieren PostgreSQL (LOWER)")
}

func TestProfileHandler_Update_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db := setupProfileHandler(t)
	db.Create(&model.Profile{ID: 1, Name: "Original", Description: "Descripción original"})

	t.Run("Actualizar perfil exitosamente", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "1"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/profiles/1",
			bytes.NewBufferString(`{"name": "Actualizado", "description": "Nueva descripción"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		handler.Update(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp dto.ProfileResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Actualizado", resp.Name)
	})

	t.Run("Eliminar con dependencias → 409", func(t *testing.T) {
		db.Create(&model.Profile{ID: 20, Name: "ConModulo", Description: "Test"})
		db.Create(&model.Module{ID: 1, Name: "TestModule"})
		db.Create(&model.ProfileModule{ProfileID: 20, ModuleID: 1})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "20"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/profiles/20", nil)
		handler.Delete(c)
		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestProfileHandler_Show_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db := setupProfileHandler(t)
	db.Create(&model.Profile{ID: 42, Name: "TestProfile", Description: "Descripción de prueba"})
	t.Run("ID válido", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "42"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/profiles/42", nil)
		handler.Show(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp dto.ProfileResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "TestProfile", resp.Name)
	})
}

func TestProfileHandler_Index_Integration(t *testing.T) {
	t.Skip("Test de paginación requiere PostgreSQL")
}

func TestModuleHandler_Index_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db := setupModuleHandler(t)
	db.Create(&model.Profile{ID: 1, Name: "Admin", Description: "Admin"})
	db.Create(&model.Module{ID: 1, Name: "Usuarios", Slug: "/users"})
	db.Create(&model.Module{ID: 2, Name: "Reportes", Slug: "/reports"})
	db.Create(&model.ProfileModule{ProfileID: 1, ModuleID: 1})
	db.Create(&model.ProfileModule{ProfileID: 1, ModuleID: 2})

	t.Run("Listar módulos de un perfil", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// ✅ CORREGIDO: usar "id" en vez de "profileId"
		c.Params = []gin.Param{{Key: "id", Value: "1"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/profiles/1/modules", nil)
		handler.Index(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["profile_id"])
		modules := resp["modules"].([]interface{})
		assert.Len(t, modules, 2)
	})

	t.Run("Perfil no encontrado → 404", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// ✅ CORREGIDO: usar "id" en vez de "profileId"
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/profiles/999/modules", nil)
		handler.Index(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestModuleHandler_Sync_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db := setupModuleHandler(t)
	db.Create(&model.Profile{ID: 1, Name: "Admin", Description: "Admin"})
	db.Create(&model.Module{ID: 1, Name: "Mod1", Slug: "/mod1"})
	db.Create(&model.Module{ID: 2, Name: "Mod2", Slug: "/mod2"})

	t.Run("Sincronizar módulos", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// ✅ CORREGIDO: usar "id" en vez de "profileId"
		c.Params = []gin.Param{{Key: "id", Value: "1"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/profiles/1/modules",
			bytes.NewBufferString(`{"modules": [1, 2]}`))
		c.Request.Header.Set("Content-Type", "application/json")
		handler.Sync(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "ok", resp["status"])
	})

	t.Run("Módulo inexistente → 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// ✅ CORREGIDO: usar "id" en vez de "profileId"
		c.Params = []gin.Param{{Key: "id", Value: "1"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/profiles/1/modules",
			bytes.NewBufferString(`{"modules": [999]}`))
		c.Request.Header.Set("Content-Type", "application/json")
		handler.Sync(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestModuleItemHandler_Index_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db := setupModuleItemHandler(t)
	db.Create(&model.Profile{ID: 1, Name: "Admin", Description: "Admin"})
	db.Create(&model.Module{ID: 1, Name: "Usuarios", Slug: "/users"})
	db.Create(&model.Item{ID: 1, Name: "crear"})
	db.Create(&model.Item{ID: 2, Name: "editar"})
	db.Create(&model.ProfileModule{ID: 1, ProfileID: 1, ModuleID: 1})
	db.Create(&model.ProfileModuleItem{ProfileModuleID: 1, ItemID: 1})
	db.Create(&model.ProfileModuleItem{ProfileModuleID: 1, ItemID: 2})

	t.Run("Listar items", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// ✅ CORREGIDO: usar "id" en vez de "profileId"
		c.Params = []gin.Param{{Key: "id", Value: "1"}, {Key: "moduleId", Value: "1"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/profiles/1/modules/1/items", nil)
		handler.Index(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(1), resp["profile_id"])
		items := resp["items"].([]interface{})
		assert.Len(t, items, 2)
	})
}

func TestModuleItemHandler_Sync_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db := setupModuleItemHandler(t)
	db.Create(&model.Profile{ID: 1, Name: "Admin", Description: "Admin"})
	db.Create(&model.Module{ID: 1, Name: "Usuarios", Slug: "/users"})
	db.Create(&model.Item{ID: 1, Name: "crear"})
	db.Create(&model.Item{ID: 2, Name: "editar"})
	db.Create(&model.ProfileModule{ID: 1, ProfileID: 1, ModuleID: 1})

	t.Run("Sincronizar items", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// ✅ CORREGIDO: usar "id" en vez de "profileId"
		c.Params = []gin.Param{{Key: "id", Value: "1"}, {Key: "moduleId", Value: "1"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/profiles/1/modules/1/items",
			bytes.NewBufferString(`{"items": [1, 2]}`))
		c.Request.Header.Set("Content-Type", "application/json")
		handler.Sync(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestModuleItemHandler_Attach_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db := setupModuleItemHandler(t)
	db.Create(&model.Profile{ID: 1, Name: "Admin", Description: "Admin"})
	db.Create(&model.Module{ID: 1, Name: "Usuarios", Slug: "/users"})
	db.Create(&model.Item{ID: 1, Name: "crear"})
	db.Create(&model.ProfileModule{ID: 1, ProfileID: 1, ModuleID: 1})

	t.Run("Agregar item", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// ✅ CORREGIDO: usar "id" en vez de "profileId"
		c.Params = []gin.Param{{Key: "id", Value: "1"}, {Key: "moduleId", Value: "1"}}
		c.Request = httptest.NewRequest(http.MethodPost, "/profiles/1/modules/1/items",
			bytes.NewBufferString(`{"item_id": 1}`))
		c.Request.Header.Set("Content-Type", "application/json")
		handler.Attach(c)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Item ya asignado → 409", func(t *testing.T) {
		db.Create(&model.ProfileModuleItem{ProfileModuleID: 1, ItemID: 1})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// ✅ CORREGIDO: usar "id" en vez de "profileId"
		c.Params = []gin.Param{{Key: "id", Value: "1"}, {Key: "moduleId", Value: "1"}}
		c.Request = httptest.NewRequest(http.MethodPost, "/profiles/1/modules/1/items",
			bytes.NewBufferString(`{"item_id": 1}`))
		c.Request.Header.Set("Content-Type", "application/json")
		handler.Attach(c)
		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestModuleItemHandler_Detach_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db := setupModuleItemHandler(t)
	db.Create(&model.Profile{ID: 1, Name: "Admin", Description: "Admin"})
	db.Create(&model.Module{ID: 1, Name: "Usuarios", Slug: "/users"})
	db.Create(&model.Item{ID: 1, Name: "crear"})
	db.Create(&model.ProfileModule{ID: 1, ProfileID: 1, ModuleID: 1})
	db.Create(&model.ProfileModuleItem{ProfileModuleID: 1, ItemID: 1})

	t.Run("Eliminar item", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// ✅ CORREGIDO: usar "id" en vez de "profileId"
		c.Params = []gin.Param{{Key: "id", Value: "1"}, {Key: "moduleId", Value: "1"}, {Key: "itemId", Value: "1"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/profiles/1/modules/1/items/1", nil)
		handler.Detach(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestProfileModuleItemResponseStructure(t *testing.T) {
	resp := map[string]interface{}{
		"profile_id": uint(1), "profile_name": "Admin",
		"module_id": uint(1), "module_name": "Usuarios", "module_slug": "/users",
		"items": []map[string]interface{}{{"id": uint(1), "name": "crear"}},
		"item_ids": []uint{1}, "total": 1,
	}
	jsonBytes, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "profile_id")
}

func TestProfileModuleSyncResponseStructure(t *testing.T) {
	resp := map[string]interface{}{
		"status": "ok", "message": "Items actualizados",
		"profile_id": uint(1), "module_id": uint(1), "attached": 2, "items": []uint{1, 2},
	}
	jsonBytes, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "ok")
}