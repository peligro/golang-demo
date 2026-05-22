package item

import (
  "errors"

  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
  "github.com/peligro/golang-demo/model"
  "gorm.io/gorm"
)

// Service maneja la lógica de negocio para Items
type Service struct {
  db *gorm.DB
}

// NewService crea una nueva instancia
func NewService(db *gorm.DB) *Service {
  return &Service{db: db}
}

// ListAll retorna todos los items ordenados por ID descendente
func (s *Service) ListAll() ([]dto.ItemResponse, error) {
  var items []model.Item
  if err := s.db.Order("id desc").Find(&items).Error; err != nil {
    return nil, err
  }

  response := make([]dto.ItemResponse, len(items))
  for i, it := range items {
    response[i] = dto.ItemResponse{
      ID:   it.ID,
      Name: it.Name,
      Code: it.Code,
    }
  }
  return response, nil
}

// GetByID retorna un item específico por su ID
func (s *Service) GetByID(id uint) (*dto.ItemResponse, error) {
  var item model.Item
  if err := s.db.First(&item, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, common.ErrNotFound
    }
    return nil, err
  }

  return &dto.ItemResponse{
    ID:   item.ID,
    Name: item.Name,
    Code: item.Code,
  }, nil
}

// Create crea un nuevo item con validaciones de unicidad
func (s *Service) Create(name, code string) (*dto.ItemResponse, error) {
  // Validar nombre único (case-insensitive)
  var existingName model.Item
  if err := s.db.Where("LOWER(name) = LOWER(?)", name).First(&existingName).Error; err == nil {
    return nil, errors.New("nombre duplicado")
  }

  // Validar code único (case-sensitive)
  var existingCode model.Item
  if err := s.db.Where("code = ?", code).First(&existingCode).Error; err == nil {
    return nil, errors.New("código duplicado")
  }

  newItem := model.Item{
    Name: name,
    Code: code,
  }
  if err := s.db.Create(&newItem).Error; err != nil {
    return nil, err
  }

  return &dto.ItemResponse{
    ID:   newItem.ID,
    Name: newItem.Name,
    Code: newItem.Code,
  }, nil
}

// Update actualiza un item existente con validaciones
func (s *Service) Update(id uint, name, code string) (*dto.ItemResponse, error) {
  var item model.Item
  if err := s.db.First(&item, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, common.ErrNotFound
    }
    return nil, err
  }

  // Validar nombre único (excluyendo el propio registro)
  var existingName model.Item
  if err := s.db.Where("LOWER(name) = LOWER(?) AND id != ?", name, id).First(&existingName).Error; err == nil {
    return nil, errors.New("nombre duplicado")
  }

  // Validar code único (excluyendo el propio registro)
  var existingCode model.Item
  if err := s.db.Where("code = ? AND id != ?", code, id).First(&existingCode).Error; err == nil {
    return nil, errors.New("código duplicado")
  }

  item.Name = name
  item.Code = code
  if err := s.db.Save(&item).Error; err != nil {
    return nil, err
  }

  return &dto.ItemResponse{
    ID:   item.ID,
    Name: item.Name,
    Code: item.Code,
  }, nil
}

// Delete elimina un item validando dependencias
func (s *Service) Delete(id uint) error {
  var item model.Item
  if err := s.db.First(&item, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return common.ErrNotFound
    }
    return err
  }

  // 🔒 Validar dependencias: ¿está asignado a algún profile_module_item?
  var count int64
  if err := s.db.Model(&model.ProfileModuleItem{}).Where("item_id = ?", id).Count(&count).Error; err != nil {
    return err
  }
  if count > 0 {
    return common.ErrHasDependencies
  }

  return s.db.Delete(&item).Error
}

// ValidateName verifica si un nombre está disponible (para validación en frontend)
func (s *Service) ValidateName(name string, excludeID *uint) error {
  query := s.db.Model(&model.Item{}).Where("LOWER(name) = LOWER(?)", name)
  if excludeID != nil {
    query = query.Where("id != ?", *excludeID)
  }
  var existing model.Item
  if err := query.First(&existing).Error; err == nil {
    return errors.New("nombre duplicado")
  }
  return nil
}

// ValidateCode verifica si un código está disponible (para validación en frontend)
func (s *Service) ValidateCode(code string, excludeID *uint) error {
  query := s.db.Model(&model.Item{}).Where("code = ?", code)
  if excludeID != nil {
    query = query.Where("id != ?", *excludeID)
  }
  var existing model.Item
  if err := query.First(&existing).Error; err == nil {
    return errors.New("código duplicado")
  }
  return nil
}

// HasDependencies verifica si un item tiene dependencias en profile_module_item
func (s *Service) HasDependencies(id uint) (bool, error) {
  var count int64
  if err := s.db.Model(&model.ProfileModuleItem{}).Where("item_id = ?", id).Count(&count).Error; err != nil {
    return false, err
  }
  return count > 0, nil
}
