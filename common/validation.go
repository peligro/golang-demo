package common

import (
  "errors"
  "net/http"
  "regexp"
  "strconv"

  "github.com/gin-gonic/gin"
  "github.com/go-playground/validator/v10"
)

// Validatable es la interfaz que deben implementar los DTOs
type Validatable interface {
  MensajesDeError(ve validator.ValidationErrors) map[string]string
}

// BindAndValidate valida el body JSON y devuelve el DTO o responde con error
func BindAndValidate[T Validatable](c *gin.Context) (T, bool) {
  var body T
  var zero T

  if err := c.ShouldBindJSON(&body); err != nil {
    var verr validator.ValidationErrors
    if errors.As(err, &verr) {
      mensajes := body.MensajesDeError(verr)
      c.JSON(http.StatusBadRequest, gin.H{
        "status":  "error",
        "message": "Validación fallida",
        "errors":  mensajes,
      })
      return zero, false
    }
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "Cuerpo de solicitud inválido",
    })
    return zero, false
  }
  return body, true
}

// ParseID valida y convierte un param "id" a uint64 (previene SQL injection)
func ParseID(c *gin.Context, paramName string) (uint64, bool) {
  idStr := c.Param(paramName)
  id, err := strconv.ParseUint(idStr, 10, 64)
  if err != nil || id == 0 {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "ID inválido",
      "field":   paramName,
    })
    return 0, false
  }
  return id, true
}

// SetupValidator registra validaciones personalizadas
func SetupValidator(validate *validator.Validate) {
  // Validación "slug": solo letras minúsculas, números y guiones bajos
  validate.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
    val := fl.Field().String()
    return regexp.MustCompile(`^[a-z0-9_]+$`).MatchString(val)
  })
} // ← ✅ FALTABA ESTA LLAVE DE CIERRE