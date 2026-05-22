package state

import (
  "errors"

  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
  "github.com/peligro/golang-demo/model"
  "gorm.io/gorm"
)

// Service maneja la lógica de negocio para States
type Service struct {
  db *gorm.DB
}

// NewService crea una nueva instancia
func NewService(db *gorm.DB) *Service {
  return &Service{db: db}
}

// ListAll retorna todos los estados ordenados por ID descendente
func (s *Service) ListAll() ([]dto.StateResponse, error) {
  var states []model.State
  if err := s.db.Order("id desc").Find(&states).Error; err != nil {
    return nil, err
  }

  response := make([]dto.StateResponse, len(states))
  for i, st := range states {
    response[i] = dto.StateResponse{
      ID:   st.ID,
      Name: st.Name,
    }
  }
  return response, nil
}

// GetByID retorna un estado específico por su ID
func (s *Service) GetByID(id uint) (*dto.StateResponse, error) {
  var state model.State
  if err := s.db.First(&state, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, common.ErrNotFound
    }
    return nil, err
  }

  return &dto.StateResponse{
    ID:   state.ID,
    Name: state.Name,
  }, nil
}

// Create crea un nuevo estado con validación de unicidad
func (s *Service) Create(name string) (*dto.StateResponse, error) {
  // Validar nombre único (case-insensitive)
  var existing model.State
  if err := s.db.Where("LOWER(name) = LOWER(?)", name).First(&existing).Error; err == nil {
    return nil, errors.New("nombre duplicado")
  }

  newState := model.State{Name: name}
  if err := s.db.Create(&newState).Error; err != nil {
    return nil, err
  }

  return &dto.StateResponse{
    ID:   newState.ID,
    Name: newState.Name,
  }, nil
}

// Update actualiza un estado existente con validación de unicidad
func (s *Service) Update(id uint, name string) (*dto.StateResponse, error) {
  var state model.State
  if err := s.db.First(&state, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, common.ErrNotFound
    }
    return nil, err
  }

  // Validar nombre único (excluyendo el propio registro)
  var existing model.State
  if err := s.db.Where("LOWER(name) = LOWER(?) AND id != ?", name, id).First(&existing).Error; err == nil {
    return nil, errors.New("nombre duplicado")
  }

  state.Name = name
  if err := s.db.Save(&state).Error; err != nil {
    return nil, err
  }

  return &dto.StateResponse{
    ID:   state.ID,
    Name: state.Name,
  }, nil
}

// Delete elimina un estado validando dependencias
func (s *Service) Delete(id uint) error {
  var state model.State
  if err := s.db.First(&state, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return common.ErrNotFound
    }
    return err
  }

  // 🔒 Validar dependencias: ¿está asignado a algún user_metadata?
  var count int64
  if err := s.db.Model(&model.UserMetadata{}).Where("state = ?", id).Count(&count).Error; err != nil {
    return err
  }
  if count > 0 {
    return common.ErrHasDependencies
  }

  return s.db.Delete(&state).Error
}

// ValidateName verifica si un nombre está disponible (para validación en frontend)
func (s *Service) ValidateName(name string, excludeID *uint) error {
  query := s.db.Model(&model.State{}).Where("LOWER(name) = LOWER(?)", name)
  if excludeID != nil {
    query = query.Where("id != ?", *excludeID)
  }
  var existing model.State
  if err := query.First(&existing).Error; err == nil {
    return errors.New("nombre duplicado")
  }
  return nil
}

// HasDependencies verifica si un estado tiene dependencias en user_metadata
func (s *Service) HasDependencies(id uint) (bool, error) {
  var count int64
  if err := s.db.Model(&model.UserMetadata{}).Where("state = ?", id).Count(&count).Error; err != nil {
    return false, err
  }
  return count > 0, nil
}
