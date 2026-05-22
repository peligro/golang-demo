package profile

import (
  "errors"

  "github.com/peligro/golang-demo/dto"
  "github.com/peligro/golang-demo/model"
  "gorm.io/gorm"
)

// ModuleService maneja la lógica de negocio para ProfileModule
type ModuleService struct {
  db *gorm.DB
}

// NewModuleService crea una nueva instancia
func NewModuleService(db *gorm.DB) *ModuleService {
  return &ModuleService{db: db}
}

// ListByProfile retorna los módulos asignados a un perfil (sin relaciones inversas)
func (s *ModuleService) ListByProfile(profileID uint) ([]dto.ModuleResponse, error) {
  // Query separada: ProfileModule → Module
  var profileModules []model.ProfileModule
  if err := s.db.Where("profile_id = ?", profileID).
    Preload("Module").
    Find(&profileModules).Error; err != nil {
    return nil, err
  }

  response := make([]dto.ModuleResponse, len(profileModules))
  for i, pm := range profileModules {
    if pm.Module != nil {
      response[i] = dto.ModuleResponse{
        ID:   pm.ModuleID,
        Name: pm.Module.Name,
        Slug: pm.Module.Slug,
      }
    }
  }
  return response, nil
}

// Sync asigna/remueve módulos de un perfil (operación sync)
func (s *ModuleService) Sync(profileID uint, moduleIDs []uint) error {
  // Validar que los módulos existen
  if len(moduleIDs) > 0 {
    var count int64
    if err := s.db.Model(&model.Module{}).Where("id IN ?", moduleIDs).Count(&count).Error; err != nil {
      return err
    }
    if count != int64(len(moduleIDs)) {
      return errors.New("uno o más módulos no existen")
    }
  }

  // Ejecutar en transacción
  return s.db.Transaction(func(tx *gorm.DB) error {
    // Eliminar asignaciones existentes
    if err := tx.Where("profile_id = ?", profileID).Delete(&model.ProfileModule{}).Error; err != nil {
      return err
    }

    // Crear nuevas asignaciones
    if len(moduleIDs) > 0 {
      pms := make([]model.ProfileModule, len(moduleIDs))
      for i, moduleID := range moduleIDs {
        pms[i] = model.ProfileModule{
          ProfileID: profileID,
          ModuleID:  moduleID,
        }
      }
      if err := tx.Create(&pms).Error; err != nil {
        return err
      }
    }
    return nil
  })
}
