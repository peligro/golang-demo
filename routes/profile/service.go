package profile

import (
  "errors"

  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
  "github.com/peligro/golang-demo/model"
  "gorm.io/gorm"
)

// Service maneja la lógica de negocio para Profiles
type Service struct {
  db *gorm.DB
}

// NewService crea una nueva instancia
func NewService(db *gorm.DB) *Service {
  return &Service{db: db}
}

// ListPaginated retorna perfiles paginados con búsqueda y ordenamiento
func (s *Service) ListPaginated(params common.PaginationParams) (common.PaginatedResponse[dto.ProfileResponse], error) {
  // Whitelist de campos permitidos
  searchableFields := []string{"name"}
  allowedSortFields := []string{"id", "name", "created_at"}

  // Construir query base
  query := s.db.Model(&model.Profile{})

  // Aplicar búsqueda y ordenamiento
  query = params.Apply(query, searchableFields, allowedSortFields)

  // Obtener total para metadatos de paginación
  var total int64
  if err := s.db.Model(&model.Profile{}).Count(&total).Error; err != nil {
    return common.PaginatedResponse[dto.ProfileResponse]{}, err
  }

  // Aplicar paginación y ejecutar
  offset := (params.Page - 1) * params.PerPage
  var profiles []model.Profile
  if err := query.Offset(offset).Limit(params.PerPage).Find(&profiles).Error; err != nil {
    return common.PaginatedResponse[dto.ProfileResponse]{}, err
  }

  // Mapear a DTOs
  response := make([]dto.ProfileResponse, len(profiles))
  for i, p := range profiles {
    response[i] = dto.ProfileResponse{
      ID:          p.ID,
      Name:        p.Name,
      Description: p.Description,
    }
  }

  return common.NewPaginatedResponse(response, int(total), params.Page, params.PerPage), nil
}

// GetByID retorna un perfil específico por su ID
func (s *Service) GetByID(id uint) (*dto.ProfileResponse, error) {
  var profile model.Profile
  if err := s.db.First(&profile, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, common.ErrNotFound
    }
    return nil, err
  }

  return &dto.ProfileResponse{
    ID:          profile.ID,
    Name:        profile.Name,
    Description: profile.Description,
  }, nil
}

// Create crea un nuevo perfil con validación de unicidad
func (s *Service) Create(name, description string) (*dto.ProfileResponse, error) {
  // Validar nombre único (case-insensitive)
  var existing model.Profile
  if err := s.db.Where("LOWER(name) = LOWER(?)", name).First(&existing).Error; err == nil {
    return nil, errors.New("nombre duplicado")
  }

  newProfile := model.Profile{
    Name:        name,
    Description: description,
  }
  if err := s.db.Create(&newProfile).Error; err != nil {
    return nil, err
  }

  return &dto.ProfileResponse{
    ID:          newProfile.ID,
    Name:        newProfile.Name,
    Description: newProfile.Description,
  }, nil
}

// Update actualiza un perfil existente con validación de unicidad
func (s *Service) Update(id uint, name, description string) (*dto.ProfileResponse, error) {
  var profile model.Profile
  if err := s.db.First(&profile, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, common.ErrNotFound
    }
    return nil, err
  }

  // Validar nombre único (excluyendo el propio registro)
  var existing model.Profile
  if err := s.db.Where("LOWER(name) = LOWER(?) AND id != ?", name, id).First(&existing).Error; err == nil {
    return nil, errors.New("nombre duplicado")
  }

  profile.Name = name
  profile.Description = description
  if err := s.db.Save(&profile).Error; err != nil {
    return nil, err
  }

  return &dto.ProfileResponse{
    ID:          profile.ID,
    Name:        profile.Name,
    Description: profile.Description,
  }, nil
}

// Delete elimina un perfil validando dependencias
func (s *Service) Delete(id uint) error {
  var profile model.Profile
  if err := s.db.First(&profile, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return common.ErrNotFound
    }
    return err
  }

  // 🔒 Validar dependencias: ¿está asignado a algún profile_module?
  var count int64
  if err := s.db.Model(&model.ProfileModule{}).Where("profile_id = ?", id).Count(&count).Error; err != nil {
    return err
  }
  if count > 0 {
    return common.ErrHasDependencies
  }

  return s.db.Delete(&profile).Error
}

// ValidateName verifica si un nombre está disponible
func (s *Service) ValidateName(name string, excludeID *uint) error {
  query := s.db.Model(&model.Profile{}).Where("LOWER(name) = LOWER(?)", name)
  if excludeID != nil {
    query = query.Where("id != ?", *excludeID)
  }
  var existing model.Profile
  if err := query.First(&existing).Error; err == nil {
    return errors.New("nombre duplicado")
  }
  return nil
}

// HasDependencies verifica si un perfil tiene dependencias en profile_module
func (s *Service) HasDependencies(id uint) (bool, error) {
  var count int64
  if err := s.db.Model(&model.ProfileModule{}).Where("profile_id = ?", id).Count(&count).Error; err != nil {
    return false, err
  }
  return count > 0, nil
}
