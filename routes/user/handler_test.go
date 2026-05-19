package user

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
	"github.com/peligro/golang-demo/pkg/auth"
)

// =============================================================================
// TestHandler_Index
// =============================================================================
func TestHandler_Index(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.User{}, &model.UserMetadata{}, &model.Profile{})

	// Seed: perfil + usuarios con metadata
	db.Create(&model.Profile{ID: 1, Name: "Administrador", Description: "Admin"})
	db.Create(&model.Profile{ID: 2, Name: "Editor", Description: "Editor"})
	db.Create(&model.User{ID: 1, Name: "César", Email: "cesar@test.com", Password: "hash123"})
	db.Create(&model.UserMetadata{UserID: 1, Phone: "+56912345678", State: 1, ProfileID: 1})
	db.Create(&model.User{ID: 2, Name: "María", Email: "maria@test.com", Password: "hash456"})
	db.Create(&model.UserMetadata{UserID: 2, Phone: "+56987654321", State: 0, ProfileID: 2})

	handler := NewHandler(db)

	t.Run("Listado sin filtros (paginación por defecto)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users", nil)

		handler.Index(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response common.PaginatedResponse[dto.UserResponse]
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 2)
		assert.Equal(t, 2, response.Pagination.Total)
		assert.Equal(t, 1, response.Pagination.CurrentPage)
		
		// ✅ Verificar que Profile se incluye cuando hay ProfileID
		for _, u := range response.Data {
			if u.ProfileID != 0 {
				assert.NotNil(t, u.Profile)
				assert.Equal(t, u.ProfileID, u.Profile.ID)
			}
		}
	})

	t.Run("Búsqueda por nombre", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users?search=César&field=name", nil)

		handler.Index(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response common.PaginatedResponse[dto.UserResponse]
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, "César", response.Data[0].Name)
		assert.Equal(t, "Administrador", response.Data[0].Profile.Name)
	})

	t.Run("Filtro por state=1 (activos)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users?state=1", nil)

		handler.Index(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response common.PaginatedResponse[dto.UserResponse]
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, 1, response.Data[0].State)
	})

	t.Run("Filtro por profile_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users?profile_id=1", nil)

		handler.Index(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response common.PaginatedResponse[dto.UserResponse]
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, uint(1), response.Data[0].ProfileID)
		assert.Equal(t, "Administrador", response.Data[0].Profile.Name)
	})

	t.Run("Usuario sin perfil asignado (Profile = nil)", func(t *testing.T) {
		db.Create(&model.User{ID: 3, Name: "SinPerfil", Email: "sinperfil@test.com", Password: "hash"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users?search=SinPerfil&field=name", nil)

		handler.Index(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response common.PaginatedResponse[dto.UserResponse]
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, "SinPerfil", response.Data[0].Name)
		assert.Equal(t, uint(0), response.Data[0].ProfileID)
		assert.Nil(t, response.Data[0].Profile)
	})
}

// =============================================================================
// TestHandler_Create
// =============================================================================
func TestHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.User{}, &model.UserMetadata{}, &model.Profile{})
	db.Create(&model.Profile{ID: 1, Name: "Administrador", Description: "Admin"})

	handler := NewHandler(db)

	t.Run("Creación exitosa con password hash + perfil", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/users",
			strings.NewReader(`{
				"name": "Nuevo Usuario",
				"email": "nuevo@test.com",
				"password": "ClaveSegura123!",
				"phone": "+56911111111",
				"state": 1,
				"profile_id": 1
			}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response dto.UserResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Nuevo Usuario", response.Name)
		assert.Equal(t, "nuevo@test.com", response.Email)
		assert.Equal(t, "+56911111111", response.Phone)
		assert.Equal(t, 1, response.State)
		assert.Equal(t, uint(1), response.ProfileID)
		assert.NotNil(t, response.Profile)
		assert.Equal(t, "Administrador", response.Profile.Name)

		// ✅ Verificar que password está hasheado en DB
		var user model.User
		db.First(&user, response.ID)
		assert.NotEqual(t, "ClaveSegura123!", user.Password)
		assert.True(t, auth.CheckPasswordHash("ClaveSegura123!", user.Password))
	})

	t.Run("Creación sin perfil asignado (Profile = nil en respuesta)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/users",
			strings.NewReader(`{
				"name": "Sin Perfil",
				"email": "sinperfil@test.com",
				"password": "Clave123!"
			}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response dto.UserResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, uint(0), response.ProfileID)
		assert.Nil(t, response.Profile)
	})

	t.Run("Email duplicado", func(t *testing.T) {
		db.Create(&model.User{Name: "Existente", Email: "dup@test.com", Password: "hash"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/users",
			strings.NewReader(`{"name":"Otro","email":"dup@test.com","password":"Clave123!"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "email ya está registrado")
	})

	t.Run("Profile_id no existe", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/users",
			strings.NewReader(`{"name":"Test","email":"test@test.com","password":"Clave123!","profile_id":999}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "perfil especificado no existe")
	})

	t.Run("Password muy corto (validación)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/users",
			strings.NewReader(`{"name":"Test","email":"test@test.com","password":"123"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "8 caracteres")
	})
}

// =============================================================================
// TestHandler_Update
// =============================================================================
func TestHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.User{}, &model.UserMetadata{}, &model.Profile{})

	hash, _ := auth.HashPassword("ClaveOriginal123!")
	db.Create(&model.User{ID: 10, Name: "Original", Email: "original@test.com", Password: hash})
	db.Create(&model.UserMetadata{UserID: 10, Phone: "+56900000000", State: 1, ProfileID: 1})
	db.Create(&model.Profile{ID: 1, Name: "Administrador", Description: "Admin"})
	db.Create(&model.Profile{ID: 2, Name: "Editor", Description: "Editor"})

	handler := NewHandler(db)

	t.Run("Actualización exitosa (sin cambiar password)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "10"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/users/10",
			strings.NewReader(`{"name":"Actualizado","email":"actualizado@test.com","phone":"+56999999999","state":0,"profile_id":2}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response dto.UserResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Actualizado", response.Name)
		assert.Equal(t, "actualizado@test.com", response.Email)
		assert.Equal(t, "+56999999999", response.Phone)
		assert.Equal(t, 0, response.State)
		assert.Equal(t, uint(2), response.ProfileID)
		assert.NotNil(t, response.Profile)
		assert.Equal(t, "Editor", response.Profile.Name)
	})

	t.Run("Cambio de password (se hashea)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "10"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/users/10",
			strings.NewReader(`{"password":"NuevaClave456!"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var user model.User
		db.First(&user, 10)
		assert.NotEqual(t, "NuevaClave456!", user.Password)
		assert.True(t, auth.CheckPasswordHash("NuevaClave456!", user.Password))
	})

	t.Run("Email duplicado al actualizar", func(t *testing.T) {
		db.Create(&model.User{Name: "Otro", Email: "dup2@test.com", Password: "hash"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "10"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/users/10",
			strings.NewReader(`{"email":"dup2@test.com"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Update(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "email ya está registrado")
	})

	t.Run("Quitar perfil asignado (ProfileID = 0)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "10"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/users/10",
			strings.NewReader(`{"profile_id":0,"state":1}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response dto.UserResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, uint(0), response.ProfileID)
		assert.Nil(t, response.Profile)
	})
}

// =============================================================================
// TestHandler_Delete
// =============================================================================
func TestHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.User{}, &model.UserMetadata{})

	db.Create(&model.User{ID: 20, Name: "ParaBorrar", Email: "borrar@test.com", Password: "hash"})
	db.Create(&model.UserMetadata{UserID: 20, Phone: "+56912312312", State: 1, ProfileID: 1})

	handler := NewHandler(db)

	t.Run("Eliminación exitosa (usuario + metadata)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "20"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/users/20", nil)

		handler.Delete(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var userCount, metaCount int64
		db.Model(&model.User{}).Where("id = ?", 20).Count(&userCount)
		db.Model(&model.UserMetadata{}).Where("user_id = ?", 20).Count(&metaCount)
		assert.Equal(t, int64(0), userCount)
		assert.Equal(t, int64(0), metaCount)
	})

	t.Run("ID no existe", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/users/999", nil)

		handler.Delete(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// =============================================================================
// TestHandler_Show
// =============================================================================
func TestHandler_Show(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&model.User{}, &model.UserMetadata{}, &model.Profile{})

	hash, _ := auth.HashPassword("Clave123!")
	db.Create(&model.User{ID: 42, Name: "César", Email: "cesar@test.com", Password: hash})
	db.Create(&model.UserMetadata{UserID: 42, Phone: "+56912345678", State: 1, ProfileID: 1})
	db.Create(&model.Profile{ID: 1, Name: "Administrador", Description: "Admin"})

	handler := NewHandler(db)

	t.Run("ID válido con perfil", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "42"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/users/42", nil)

		handler.Show(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response dto.UserResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "César", response.Name)
		assert.Equal(t, 1, response.State)
		assert.Equal(t, uint(1), response.ProfileID)
		assert.NotNil(t, response.Profile)
		assert.Equal(t, "Administrador", response.Profile.Name)
	})

	t.Run("ID válido sin perfil", func(t *testing.T) {
		db.Create(&model.User{ID: 99, Name: "SinPerfil", Email: "sin@test.com", Password: "hash"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "99"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/users/99", nil)

		handler.Show(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response dto.UserResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "SinPerfil", response.Name)
		assert.Equal(t, uint(0), response.ProfileID)
		assert.Nil(t, response.Profile)
	})

	t.Run("ID inválido (string)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "abc"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/users/abc", nil)

		handler.Show(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "ID inválido")
	})

	t.Run("No encontrado", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "999"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/users/999", nil)

		handler.Show(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "no encontrado")
	})
}

// =============================================================================
// TestHandler_Me
// =============================================================================
func TestHandler_Me(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	handler := NewHandler(db)

	t.Run("Requiere autenticación (401)", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users/me", nil)

		handler.Me(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Requiere autenticación")
	})
}