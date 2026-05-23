package app_menu

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

// ListAll retorna todos los menús ordenados (sin paginación)
func (s *Service) ListAll() ([]dto.AppMenuResponse, error) {
  var menus []model.AppMenu
  // "order" es palabra reservada en SQL, lo escapamos
  if err := s.db.Order("\"order\" ASC, id ASC").Find(&menus).Error; err != nil {
    return nil, err
  }

  response := make([]dto.AppMenuResponse, len(menus))
  for i, m := range menus {
    response[i] = s.toResponse(m)
  }
  return response, nil
}

// ListPaginated retorna menús paginados con búsqueda y ordenamiento
func (s *Service) ListPaginated(params common.PaginationParams) (common.PaginatedResponse[dto.AppMenuResponse], error) {
  searchableFields := []string{"label", "title"}
  allowedSortFields := []string{"id", "label", "title", "order", "created_at"}

  query := s.db.Model(&model.AppMenu{})
  query = params.Apply(query, searchableFields, allowedSortFields)

  var total int64
  if err := s.db.Model(&model.AppMenu{}).Count(&total).Error; err != nil {
    return common.PaginatedResponse[dto.AppMenuResponse]{}, err
  }

  offset := (params.Page - 1) * params.PerPage
  var menus []model.AppMenu
  if err := query.Offset(offset).Limit(params.PerPage).Find(&menus).Error; err != nil {
    return common.PaginatedResponse[dto.AppMenuResponse]{}, err
  }

  response := make([]dto.AppMenuResponse, len(menus))
  for i, m := range menus {
    response[i] = s.toResponse(m)
  }

  return common.NewPaginatedResponse(response, int(total), params.Page, params.PerPage), nil
}

// GetByID retorna un menú específico
func (s *Service) GetByID(id uint) (*dto.AppMenuResponse, error) {
  var menu model.AppMenu
  if err := s.db.First(&menu, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, common.ErrNotFound
    }
    return nil, err
  }
  res := s.toResponse(menu)
  return &res, nil
}

// Create crea un nuevo menú validando FKs
func (s *Service) Create(req dto.AppMenuCreateRequest) (*dto.AppMenuResponse, error) {
  // Validar ParentID si se proporciona
  if req.ParentID != nil && *req.ParentID != 0 {
    var count int64
    s.db.Model(&model.AppMenu{}).Where("id = ?", *req.ParentID).Count(&count)
    if count == 0 {
      return nil, errors.New("el menú padre no existe")
    }
  }

  // Validar ModuleID si se proporciona
  if req.ModuleID != nil && *req.ModuleID != 0 {
    var count int64
    s.db.Model(&model.Module{}).Where("id = ?", *req.ModuleID).Count(&count)
    if count == 0 {
      return nil, errors.New("el módulo no existe")
    }
  }

  newMenu := model.AppMenu{
    Label:    req.Label,
    Title:    req.Title,
    Icon:     req.Icon,
    Order:    req.Order,
    ParentID: req.ParentID,
    ModuleID: req.ModuleID, // ← Ahora ambos son *uint, compatible
  }

  if err := s.db.Create(&newMenu).Error; err != nil {
    return nil, err
  }

  res := s.toResponse(newMenu)
  return &res, nil
}

// Update actualiza un menú existente
func (s *Service) Update(id uint, req dto.AppMenuUpdateRequest) (*dto.AppMenuResponse, error) {
  var menu model.AppMenu
  if err := s.db.First(&menu, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, common.ErrNotFound
    }
    return nil, err
  }

  // Validar ParentID si cambia
  if req.ParentID != nil && *req.ParentID != 0 {
    if *req.ParentID == id {
      return nil, errors.New("un menú no puede ser su propio padre")
    }
    var count int64
    s.db.Model(&model.AppMenu{}).Where("id = ?", *req.ParentID).Count(&count)
    if count == 0 {
      return nil, errors.New("el menú padre no existe")
    }
  }

  // Validar ModuleID si cambia
  if req.ModuleID != nil && *req.ModuleID != 0 {
    var count int64
    s.db.Model(&model.Module{}).Where("id = ?", *req.ModuleID).Count(&count)
    if count == 0 {
      return nil, errors.New("el módulo no existe")
    }
  }

  // Aplicar cambios solo si no están vacíos/nulos
  if req.Label != "" {
    menu.Label = req.Label
  }
  if req.Title != "" {
    menu.Title = req.Title
  }
  if req.Icon != "" {
    menu.Icon = req.Icon
  }
  // Order: permitir 0 explícito
  menu.Order = req.Order
  if req.ParentID != nil {
    menu.ParentID = req.ParentID
  }
  if req.ModuleID != nil {
    menu.ModuleID = req.ModuleID // ← Ambos *uint, compatible
  }

  if err := s.db.Save(&menu).Error; err != nil {
    return nil, err
  }

  res := s.toResponse(menu)
  return &res, nil
}

// Delete elimina un menú (valida que no tenga hijos)
func (s *Service) Delete(id uint) error {
  var menu model.AppMenu
  if err := s.db.First(&menu, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return common.ErrNotFound
    }
    return err
  }

  // Validar que no tenga menús hijos
  var childrenCount int64
  s.db.Model(&model.AppMenu{}).Where("parent_id = ?", id).Count(&childrenCount)
  if childrenCount > 0 {
    return errors.New("no se puede eliminar: tiene menús hijos asociados")
  }

  return s.db.Delete(&menu).Error
}

// toResponse helper para mapear model -> dto
func (s *Service) toResponse(m model.AppMenu) dto.AppMenuResponse {
  return dto.AppMenuResponse{
    ID:        m.ID,
    Label:     m.Label,
    Title:     m.Title,
    Icon:      m.Icon,
    Order:     m.Order,
    ParentID:  m.ParentID,  // ← Ambos *uint
    ModuleID:  m.ModuleID,  // ← Ambos *uint
    CreatedAt: m.CreatedAt,
    UpdatedAt: m.UpdatedAt,
  }
}