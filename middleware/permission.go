package middleware

import (
  "net/http"

  "github.com/gin-gonic/gin"
  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/model"
  "gorm.io/gorm"
)

// RequirePermission retorna un middleware que valida permisos
// Siempre retorna 401 genérico para no revelar información
func RequirePermission(db *gorm.DB, moduleSlug string, itemCode string, specialCodes ...string) gin.HandlerFunc {
  return func(c *gin.Context) {
    // 1. Obtener userID del contexto
    userID, exists := GetUserIDFromContext(c)
    if !exists {
      // 🔒 Genérico: no revelar si falta cookie o es inválida
      c.JSON(http.StatusUnauthorized, gin.H{
        "status":  "error",
        "message": "No autenticado",
      })
      c.Abort()
      return
    }

    // 2. Cargar metadata con perfil
    var meta model.UserMetadata
    if err := db.Where("user_id = ?", userID).
      Preload("Profile").
      First(&meta).Error; err != nil {
      // 🔒 Genérico: no revelar si el usuario existe o no
      c.JSON(http.StatusUnauthorized, gin.H{
        "status":  "error",
        "message": "No autenticado",
      })
      c.Abort()
      return
    }

    // 3. Recopilar permisos del usuario
    userItemCodes := make(map[string]bool)
    userModuleSlugs := make(map[string]bool)

    if meta.Profile != nil {
      var profileModules []model.ProfileModule
      if err := db.Where("profile_id = ?", meta.Profile.ID).
        Preload("Module").
        Find(&profileModules).Error; err != nil {
        profileModules = []model.ProfileModule{}
      }

      for _, pm := range profileModules {
        if pm.Module != nil {
          userModuleSlugs[pm.Module.Slug] = true

          var profileModuleItems []model.ProfileModuleItem
          if err := db.Where("profile_module_id = ?", pm.ID).
            Preload("Item").
            Find(&profileModuleItems).Error; err != nil {
            continue
          }

          for _, pmi := range profileModuleItems {
            if pmi.Item != nil {
              userItemCodes[pmi.Item.Code] = true
            }
          }
        }
      }
    }

    // 4. Verificar permisos especiales
    for _, code := range specialCodes {
      if userItemCodes[code] {
        c.Next()
        return
      }
    }

    // 5. Verificar módulo específico
    if moduleSlug != "" && userModuleSlugs[moduleSlug] {
      if itemCode == "" {
        c.Next()
        return
      }
      if itemCode != "" && userItemCodes[itemCode] {
        c.Next()
        return
      }
    }

    // 6. 🔒 SIN PERMISOS - Respuesta genérica 401 (sin revelar nada)
    c.JSON(http.StatusUnauthorized, gin.H{
      "status":  "error",
      "message": "No autenticado",
    })
    c.Abort()
  }
}

// RequireModule: solo verifica que el usuario tenga el módulo
func RequireModule(db *gorm.DB, moduleSlug string) gin.HandlerFunc {
  return RequirePermission(db, moduleSlug, "", common.ViewAllAdminCode)
}

// RequireItem: verifica módulo + item específico
func RequireItem(db *gorm.DB, moduleSlug, itemCode string) gin.HandlerFunc {
  return RequirePermission(db, moduleSlug, itemCode, common.ViewAllAdminCode)
}

// RequireAnySpecialCode: verifica si tiene CUALQUIERA de los códigos especiales
func RequireAnySpecialCode(db *gorm.DB, codes ...string) gin.HandlerFunc {
  return RequirePermission(db, "", "", codes...)
}