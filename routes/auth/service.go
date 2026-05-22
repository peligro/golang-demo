package auth

import (
  "errors"

  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
  "github.com/peligro/golang-demo/model"
  authpkg "github.com/peligro/golang-demo/pkg/auth"
  "gorm.io/gorm"
)

// Service maneja la lógica de negocio de autenticación
type Service struct {
  db *gorm.DB
}

// NewService crea una nueva instancia
func NewService(db *gorm.DB) *Service {
  return &Service{db: db}
}

// Login autentica al usuario y crea sesión
// Retorna: user DTO, token para cookie, error
func (s *Service) Login(email, password string) (*dto.UserResponse, string, error) {
  // Buscar usuario
  var user model.User
  if err := s.db.Where("email ILIKE ?", email).First(&user).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, "", errors.New("credenciales inválidas")
    }
    return nil, "", err
  }

  // Buscar metadata para validar state
  var meta model.UserMetadata
  if err := s.db.Where("user_id = ?", user.ID).First(&meta).Error; err != nil {
    return nil, "", err
  }

  // Validar estado
  if meta.State != 1 {
    return nil, "", errors.New("usuario inactivo")
  }

  // Verificar contraseña
  if !user.CheckPassword(password) {
    return nil, "", errors.New("credenciales inválidas")
  }

  // Generar token y crear sesión
  token, err := authpkg.GenerateToken()
  if err != nil {
    return nil, "", err
  }
  if err := authpkg.CreateSession(token, user.ID, user.Email); err != nil {
    return nil, "", err
  }

  // Preparar respuesta (sin password)
  var profileSummary *dto.ProfileSummary
  if meta.Profile != nil {
    profileSummary = &dto.ProfileSummary{
      ID:   meta.Profile.ID,
      Name: meta.Profile.Name,
    }
  }

  return &dto.UserResponse{
    ID:        user.ID,
    Name:      user.Name,
    Email:     user.Email,
    Date:      common.FormatDate(user.CreatedAt),
    Time:      common.FormatTime(user.CreatedAt),
    Phone:     meta.Phone,
    State:     meta.State,
    ProfileID: meta.ProfileID,
    Profile:   profileSummary,
  }, token, nil
}

// Logout invalida la sesión actual
func (s *Service) Logout(token string) error {
  return authpkg.DeleteSession(token)
}

// GetMe retorna los datos del usuario autenticado con módulos y permisos
func (s *Service) GetMe(userID uint) (*dto.UserResponse, error) {
  // Cargar usuario
  var user model.User
  if err := s.db.First(&user, userID).Error; err != nil {
    return nil, err
  }

  // Cargar metadata con perfil (1:1)
  var meta model.UserMetadata
  if err := s.db.Where("user_id = ?", userID).
    Preload("Profile").
    First(&meta).Error; err != nil {
    return nil, err
  }

  // Cargar módulos asignados al perfil (many-to-many vía ProfileModule)
  var profileModules []model.ProfileModule
  if meta.Profile != nil {
    // Query separada: ProfileModule → Module
    if err := s.db.Where("profile_id = ?", meta.Profile.ID).
      Preload("Module").
      Find(&profileModules).Error; err != nil {
      return nil, err
    }
  }

  // Convertir a DTO con items (queries separadas para ProfileModuleItem → Item)
  modules := s.convertProfileModulesToResponse(profileModules)

  var profileSummary *dto.ProfileSummary
  if meta.Profile != nil {
    profileSummary = &dto.ProfileSummary{
      ID:   meta.Profile.ID,
      Name: meta.Profile.Name,
    }
  }

  return &dto.UserResponse{
    ID:        user.ID,
    Name:      user.Name,
    Email:     user.Email,
    Date:      common.FormatDate(user.CreatedAt),
    Time:      common.FormatTime(user.CreatedAt),
    Phone:     meta.Phone,
    State:     meta.State,
    ProfileID: meta.ProfileID,
    Profile:   profileSummary,
    Modules:   modules, // ← Módulos con sus items/permisos
  }, nil
}

// convertProfileModulesToResponse convierte []ProfileModule a []UserModuleResponse
// Carga items con queries separadas (sin relaciones inversas en modelos)
func (s *Service) convertProfileModulesToResponse(profileModules []model.ProfileModule) []dto.UserModuleResponse {
  result := make([]dto.UserModuleResponse, 0, len(profileModules))
  
  for _, pm := range profileModules {
    // Saltar si el módulo está nil
    if pm.Module == nil {
      continue
    }
    
    // 🔥 Cargar items para este módulo con query separada
    // ProfileModuleItem → Item (sin preload inverso en Module)
    var profileModuleItems []model.ProfileModuleItem
    _ = s.db.Where("profile_module_id = ?", pm.ID).
           Preload("Item").
           Find(&profileModuleItems)
    
    // Recopilar items únicos usando un map para evitar duplicados
    itemMap := make(map[uint]dto.ItemResponse)
    for _, pmi := range profileModuleItems {
      if pmi.Item != nil {
        itemMap[pmi.Item.ID] = dto.ItemResponse{
          ID:   pmi.Item.ID,
          Name: pmi.Item.Name,
          Code: pmi.Item.Code, // ← Incluir code para permisos en frontend
        }
      }
    }
    
    // Convertir map a slice
    items := make([]dto.ItemResponse, 0, len(itemMap))
    for _, item := range itemMap {
      items = append(items, item)
    }
    
    // Agregar módulo solo si tiene nombre válido
    if pm.Module.Name != "" {
      result = append(result, dto.UserModuleResponse{
        Name:  pm.Module.Name,
        Slug:  pm.Module.Slug,
        Items: items,
      })
    }
  }
  
  return result
}