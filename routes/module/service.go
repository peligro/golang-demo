package module

import (
  "errors"

  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
  "github.com/peligro/golang-demo/model"
  "gorm.io/gorm"
)

// Service maneja la lógica de negocio para Modules
type Service struct {
  db *gorm.DB
}

// NewService crea una nueva instancia
func NewService(db *gorm.DB) *Service {
  return &Service{db: db}
}

// ListPaginated retorna módulos paginados con búsqueda y ordenamiento
func (s *Service) ListPaginated(params common.PaginationParams) (common.PaginatedResponse[dto.ModuleResponse], error) {
  // Whitelist de campos permitidos
  searchableFields := []string{"name"}
  allowedSortFields := []string{"id", "name", "slug", "created_at"}

  // Construir query base
  query := s.db.Model(&model.Module{})

  // Aplicar búsqueda y ordenamiento
  query = params.Apply(query, searchableFields, allowedSortFields)

  // Obtener total para metadatos de paginación
  var total int64
  if err := s.db.Model(&model.Module{}).Count(&total).Error; err != nil {
    return common.PaginatedResponse[dto.ModuleResponse]{}, err
  }

  // Aplicar paginación y ejecutar
  offset := (params.Page - 1) * params.PerPage
  var modules []model.Module
  if err := query.Offset(offset).Limit(params.PerPage).Find(&modules).Error; err != nil {
    return common.PaginatedResponse[dto.ModuleResponse]{}, err
  }

  // Mapear a DTOs
  response := make([]dto.ModuleResponse, len(modules))
  for i, m := range modules {
    response[i] = dto.ModuleResponse{
      ID:   m.ID,
      Name: m.Name,
      Slug: m.Slug,
    }
  }

  return common.NewPaginatedResponse(response, int(total), params.Page, params.PerPage), nil
}

// GetByID retorna un módulo específico por su ID
func (s *Service) GetByID(id uint) (*dto.ModuleResponse, error) {
  var module model.Module
  if err := s.db.First(&module, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, common.ErrNotFound
    }
    return nil, err
  }

  return &dto.ModuleResponse{
    ID:   module.ID,
    Name: module.Name,
    Slug: module.Slug,
  }, nil
}

// Create crea un nuevo módulo con validaciones de unicidad
func (s *Service) Create(name, slug string) (*dto.ModuleResponse, error) {
  // Validar nombre único (case-insensitive)
  var existingName model.Module
  if err := s.db.Where("LOWER(name) = LOWER(?)", name).First(&existingName).Error; err == nil {
    return nil, errors.New("nombre duplicado")
  }

  // Validar slug único (case-sensitive)
  var existingSlug model.Module
  if err := s.db.Where("slug = ?", slug).First(&existingSlug).Error; err == nil {
    return nil, errors.New("path duplicado")
  }

  newModule := model.Module{
    Name: name,
    Slug: slug,
  }
  if err := s.db.Create(&newModule).Error; err != nil {
    return nil, err
  }

  return &dto.ModuleResponse{
    ID:   newModule.ID,
    Name: newModule.Name,
    Slug: newModule.Slug,
  }, nil
}

// Update actualiza un módulo existente con validaciones
func (s *Service) Update(id uint, name, slug string) (*dto.ModuleResponse, error) {
  var module model.Module
  if err := s.db.First(&module, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, common.ErrNotFound
    }
    return nil, err
  }

  // Validar nombre único (excluyendo el propio registro)
  var existingName model.Module
  if err := s.db.Where("LOWER(name) = LOWER(?) AND id != ?", name, id).First(&existingName).Error; err == nil {
    return nil, errors.New("nombre duplicado")
  }

  // Validar slug único (excluyendo el propio registro)
  var existingSlug model.Module
  if err := s.db.Where("slug = ? AND id != ?", slug, id).First(&existingSlug).Error; err == nil {
    return nil, errors.New("path duplicado")
  }

  module.Name = name
  module.Slug = slug
  if err := s.db.Save(&module).Error; err != nil {
    return nil, err
  }

  return &dto.ModuleResponse{
    ID:   module.ID,
    Name: module.Name,
    Slug: module.Slug,
  }, nil
}

// Delete elimina un módulo validando dependencias
func (s *Service) Delete(id uint) error {
  var module model.Module
  if err := s.db.First(&module, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return common.ErrNotFound
    }
    return err
  }

  // 🔒 Validar dependencias: ¿está asignado a algún profile_module?
  var count int64
  if err := s.db.Model(&model.ProfileModule{}).Where("module_id = ?", id).Count(&count).Error; err != nil {
    return err
  }
  if count > 0 {
    return common.ErrHasDependencies
  }

  return s.db.Delete(&module).Error
}

// ValidateName verifica si un nombre está disponible
func (s *Service) ValidateName(name string, excludeID *uint) error {
  query := s.db.Model(&model.Module{}).Where("LOWER(name) = LOWER(?)", name)
  if excludeID != nil {
    query = query.Where("id != ?", *excludeID)
  }
  var existing model.Module
  if err := query.First(&existing).Error; err == nil {
    return errors.New("nombre duplicado")
  }
  return nil
}

// ValidateSlug verifica si un slug/path está disponible
func (s *Service) ValidateSlug(slug string, excludeID *uint) error {
  query := s.db.Model(&model.Module{}).Where("slug = ?", slug)
  if excludeID != nil {
    query = query.Where("id != ?", *excludeID)
  }
  var existing model.Module
  if err := query.First(&existing).Error; err == nil {
    return errors.New("path duplicado")
  }
  return nil
}

// HasDependencies verifica si un módulo tiene dependencias en profile_module
func (s *Service) HasDependencies(id uint) (bool, error) {
  var count int64
  if err := s.db.Model(&model.ProfileModule{}).Where("module_id = ?", id).Count(&count).Error; err != nil {
    return false, err
  }
  return count > 0, nil
}