package profile

import (
  "errors"

  "github.com/peligro/golang-demo/dto"
  "github.com/peligro/golang-demo/model"
  "gorm.io/gorm"
)

// ModuleItemService maneja la lógica de negocio para ProfileModuleItem
type ModuleItemService struct {
  db *gorm.DB
}

// NewModuleItemService crea una nueva instancia
func NewModuleItemService(db *gorm.DB) *ModuleItemService {
  return &ModuleItemService{db: db}
}

// ListByModuleAndProfile retorna los items asignados a un módulo de un perfil
func (s *ModuleItemService) ListByModuleAndProfile(profileID, moduleID uint) ([]dto.ItemResponse, error) {
  // Primero obtener el ProfileModule
  var profileModule model.ProfileModule
  if err := s.db.Where("profile_id = ? AND module_id = ?", profileID, moduleID).
    First(&profileModule).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, errors.New("el módulo no está asignado a este perfil")
    }
    return nil, err
  }

  // Luego obtener los ProfileModuleItems con preload de Item
  var profileModuleItems []model.ProfileModuleItem
  if err := s.db.Where("profile_module_id = ?", profileModule.ID).
    Preload("Item").
    Find(&profileModuleItems).Error; err != nil {
    return nil, err
  }

  response := make([]dto.ItemResponse, len(profileModuleItems))
  for i, pmi := range profileModuleItems {
    if pmi.Item != nil {
      response[i] = dto.ItemResponse{
        ID:   pmi.ItemID,
        Name: pmi.Item.Name,
        Code: pmi.Item.Code,
      }
    }
  }
  return response, nil
}

// Sync asigna/remueve items de un módulo de perfil (operación sync)
func (s *ModuleItemService) Sync(profileID, moduleID uint, itemIDs []uint) error {
  // Obtener el ProfileModule
  var profileModule model.ProfileModule
  if err := s.db.Where("profile_id = ? AND module_id = ?", profileID, moduleID).
    First(&profileModule).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return errors.New("el módulo no está asignado a este perfil")
    }
    return err
  }

  // Validar que los items existen
  if len(itemIDs) > 0 {
    var count int64
    if err := s.db.Model(&model.Item{}).Where("id IN ?", itemIDs).Count(&count).Error; err != nil {
      return err
    }
    if count != int64(len(itemIDs)) {
      return errors.New("uno o más items no existen")
    }
  }

  // Ejecutar en transacción
  return s.db.Transaction(func(tx *gorm.DB) error {
    // Eliminar asignaciones existentes
    if err := tx.Where("profile_module_id = ?", profileModule.ID).
      Delete(&model.ProfileModuleItem{}).Error; err != nil {
      return err
    }

    // Crear nuevas asignaciones
    if len(itemIDs) > 0 {
      pmis := make([]model.ProfileModuleItem, len(itemIDs))
      for i, itemID := range itemIDs {
        pmis[i] = model.ProfileModuleItem{
          ProfileModuleID: profileModule.ID,
          ItemID:          itemID,
        }
      }
      if err := tx.Create(&pmis).Error; err != nil {
        return err
      }
    }
    return nil
  })
}

// Attach asigna un item individual a un módulo de perfil
func (s *ModuleItemService) Attach(profileID, moduleID, itemID uint) (*dto.ItemResponse, error) {
  // Obtener el ProfileModule
  var profileModule model.ProfileModule
  if err := s.db.Where("profile_id = ? AND module_id = ?", profileID, moduleID).
    First(&profileModule).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, errors.New("el módulo no está asignado a este perfil")
    }
    return nil, err
  }

  // Validar que el item existe
  var item model.Item
  if err := s.db.First(&item, itemID).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, errors.New("el item no existe")
    }
    return nil, err
  }

  // Validar que no esté ya asignado
  var existing model.ProfileModuleItem
  if err := s.db.Where("profile_module_id = ? AND item_id = ?", profileModule.ID, itemID).
    First(&existing).Error; err == nil {
    return nil, errors.New("este item ya está asignado")
  }

  // Crear la asignación
  pmi := model.ProfileModuleItem{
    ProfileModuleID: profileModule.ID,
    ItemID:          itemID,
  }
  if err := s.db.Create(&pmi).Error; err != nil {
    return nil, err
  }

  return &dto.ItemResponse{
    ID:   item.ID,
    Name: item.Name,
    Code: item.Code,
  }, nil
}

// Detach remueve un item de un módulo de perfil
func (s *ModuleItemService) Detach(profileID, moduleID, itemID uint) error {
  // Obtener el ProfileModule
  var profileModule model.ProfileModule
  if err := s.db.Where("profile_id = ? AND module_id = ?", profileID, moduleID).
    First(&profileModule).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return errors.New("el módulo no está asignado a este perfil")
    }
    return err
  }

  // Buscar la asignación
  var pmi model.ProfileModuleItem
  if err := s.db.Where("profile_module_id = ? AND item_id = ?", profileModule.ID, itemID).
    First(&pmi).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return errors.New("el item no está asignado")
    }
    return err
  }

  // Eliminar la asignación
  return s.db.Delete(&pmi).Error
}
