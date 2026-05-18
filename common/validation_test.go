package common

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// mockDTO implementa Validatable para pruebas
type mockDTO struct {
	Name string `json:"name" binding:"required,min=3"`
}

func (m mockDTO) MensajesDeError(ve validator.ValidationErrors) map[string]string {
	errs := make(map[string]string)
	for _, e := range ve {
		if e.Field() == "Name" {
			switch e.Tag() {
			case "required":
				errs["name"] = "El nombre es obligatorio"
			case "min":
				errs["name"] = "El nombre debe tener al menos 3 caracteres"
			}
		}
	}
	return errs
}

func TestParseID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name       string
		paramValue string
		wantID     uint64
		wantOK     bool
		wantStatus int
	}{
		{"Válido", "42", 42, true, http.StatusOK},
		{"Cero", "0", 0, false, http.StatusBadRequest},
		{"Negativo", "-5", 0, false, http.StatusBadRequest},
		{"Letras", "abc", 0, false, http.StatusBadRequest},
		{"Vacío", "", 0, false, http.StatusBadRequest},
		{"Overflow", "999999999999999999999", 0, false, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{{Key: "id", Value: tt.paramValue}}

			gotID, gotOK := ParseID(c, "id")

			assert.Equal(t, tt.wantID, gotID)
			assert.Equal(t, tt.wantOK, gotOK)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestBindAndValidate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       interface{}
		wantOK     bool
		wantStatus int
		checkErrs  map[string]string
	}{
		{"Válido", mockDTO{Name: "Activo"}, true, http.StatusOK, nil},
		{"Vacío", mockDTO{Name: ""}, false, http.StatusBadRequest, map[string]string{"name": "El nombre es obligatorio"}},
		{"Corto", mockDTO{Name: "AB"}, false, http.StatusBadRequest, map[string]string{"name": "El nombre debe tener al menos 3 caracteres"}},
		{"JSON inválido", "{bad-json", false, http.StatusBadRequest, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var reqBody []byte
			if b, ok := tt.body.([]byte); ok {
				reqBody = b
			} else {
				reqBody, _ = json.Marshal(tt.body)
			}

			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			_, gotOK := BindAndValidate[mockDTO](c)

			assert.Equal(t, tt.wantOK, gotOK)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.checkErrs != nil {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				if errs, ok := resp["errors"].(map[string]interface{}); ok {
					for k, v := range tt.checkErrs {
						assert.Equal(t, v, errs[k])
					}
				}
			}
		})
	}
}
