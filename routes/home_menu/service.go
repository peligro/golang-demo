package home_menu

import (
  "errors"

  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
  "github.com/peligro/golang-demo/model"
  "gorm.io/gorm"
)

type Service struct {
  db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
  return &Service{db: db}
}

// ListAll retorna todos los menús del home ordenados (sin paginación)
func (s *Service) ListAll() ([]dto.HomeMenuResponse, error) {
  var menus []model.HomeMenu
  // "order" es palabra reservada en SQL, lo escapamos
  if err := s.db.Order("\"order\" ASC, id ASC").Find(&menus).Error; err != nil {
    return nil, err
  }

  response := make([]dto.HomeMenuResponse, len(menus))
  for i, m := range menus {
    response[i] = s.toResponse(m)
  }
  return response, nil
}

// ListPaginated retorna menús del home paginados con búsqueda y ordenamiento
func (s *Service) ListPaginated(params common.PaginationParams) (common.PaginatedResponse[dto.HomeMenuResponse], error) {
  searchableFields := []string{"title", "description", "slug"}
  allowedSortFields := []string{"id", "title", "order", "created_at"}

  query := s.db.Model(&model.HomeMenu{})
  query = params.Apply(query, searchableFields, allowedSortFields)

  var total int64
  if err := s.db.Model(&model.HomeMenu{}).Count(&total).Error; err != nil {
    return common.PaginatedResponse[dto.HomeMenuResponse]{}, err
  }

  offset := (params.Page - 1) * params.PerPage
  var menus []model.HomeMenu
  if err := query.Offset(offset).Limit(params.PerPage).Find(&menus).Error; err != nil {
    return common.PaginatedResponse[dto.HomeMenuResponse]{}, err
  }

  response := make([]dto.HomeMenuResponse, len(menus))
  for i, m := range menus {
    response[i] = s.toResponse(m)
  }

  return common.NewPaginatedResponse(response, int(total), params.Page, params.PerPage), nil
}

// GetByID retorna un menú del home específico
func (s *Service) GetByID(id uint) (*dto.HomeMenuResponse, error) {
  var menu model.HomeMenu
  if err := s.db.First(&menu, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, common.ErrNotFound
    }
    return nil, err
  }
  res := s.toResponse(menu)
  return &res, nil
}

// Create crea un nuevo menú del home validando FKs
func (s *Service) Create(req dto.HomeMenuCreateRequest) (*dto.HomeMenuResponse, error) {
  // Validar ModuleID si se proporciona
  if req.ModuleID != nil && *req.ModuleID != 0 {
    var count int64
    s.db.Model(&model.Module{}).Where("id = ?", *req.ModuleID).Count(&count)
    if count == 0 {
      return nil, errors.New("el módulo no existe")
    }
  }

  newMenu := model.HomeMenu{
    Title:       req.Title,
    Icon:        req.Icon,
    Color:       req.Color,
    Description: req.Description,
    Slug:        req.Slug,
    Order:       req.Order,
    ModuleID:    req.ModuleID,
  }

  if err := s.db.Create(&newMenu).Error; err != nil {
    return nil, err
  }

  res := s.toResponse(newMenu)
  return &res, nil
}

// Update actualiza un menú del home existente
func (s *Service) Update(id uint, req dto.HomeMenuUpdateRequest) (*dto.HomeMenuResponse, error) {
  var menu model.HomeMenu
  if err := s.db.First(&menu, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, common.ErrNotFound
    }
    return nil, err
  }

  // Validar ModuleID si cambia
  if req.ModuleID != nil && *req.ModuleID != 0 {
    var count int64
    s.db.Model(&model.Module{}).Where("id = ?", *req.ModuleID).Count(&count)
    if count == 0 {
      return nil, errors.New("el módulo no existe")
    }
  }

  // Aplicar cambios solo si no están vacíos
  if req.Title != "" {
    menu.Title = req.Title
  }
  if req.Icon != "" {
    menu.Icon = req.Icon
  }
  if req.Color != "" {
    menu.Color = req.Color
  }
  if req.Description != "" {
    menu.Description = req.Description
  }
  if req.Slug != "" {
    menu.Slug = req.Slug
  }
  // Order: permitir 0 explícito
  menu.Order = req.Order
  if req.ModuleID != nil {
    menu.ModuleID = req.ModuleID
  }

  if err := s.db.Save(&menu).Error; err != nil {
    return nil, err
  }

  res := s.toResponse(menu)
  return &res, nil
}

// Delete elimina un menú del home
func (s *Service) Delete(id uint) error {
  var menu model.HomeMenu
  if err := s.db.First(&menu, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return common.ErrNotFound
    }
    return err
  }

  return s.db.Delete(&menu).Error
}

// toResponse helper para mapear model -> dto
func (s *Service) toResponse(m model.HomeMenu) dto.HomeMenuResponse {
  return dto.HomeMenuResponse{
    ID:          m.ID,
    Title:       m.Title,
    Icon:        m.Icon,
    Color:       m.Color,
    Description: m.Description,
    Slug:        m.Slug,
    Order:       m.Order,
    ModuleID:    m.ModuleID,
    CreatedAt:   m.CreatedAt,
    UpdatedAt:   m.UpdatedAt,
  }
}
