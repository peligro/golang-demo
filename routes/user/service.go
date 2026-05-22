package user

import (
  "errors"
  "fmt"
  "strings"

  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
  "github.com/peligro/golang-demo/model"
  authpkg "github.com/peligro/golang-demo/pkg/auth"
  "gorm.io/gorm"
)

// Service maneja la lógica de negocio para Users
type Service struct {
  db *gorm.DB
}

// NewService crea una nueva instancia
func NewService(db *gorm.DB) *Service {
  return &Service{db: db}
}

// ListPaginated retorna usuarios paginados con filtros y búsqueda
func (s *Service) ListPaginated(params userListParams) (common.PaginatedResponse[dto.UserResponse], error) {
  // Query base: usar comillas dobles para escapar "user" (keyword reservada en PostgreSQL)
  query := s.db.Table("\"user\"").
    Joins("LEFT JOIN user_metadata ON user_metadata.user_id = \"user\".id")

  // Aplicar búsqueda textual (name/email)
  if params.Search != "" && params.Field != "" {
    query = query.Where(fmt.Sprintf("\"user\".%s ILIKE ?", params.Field), "%"+params.Search+"%")
  }

  // Aplicar filtro por estado
  if params.State != nil {
    query = query.Where("user_metadata.state = ?", *params.State)
  }

  // Aplicar filtro por perfil
  if params.ProfileID != nil {
    query = query.Where("user_metadata.profile_id = ?", *params.ProfileID)
  }

  // Aplicar ordenamiento
  if isValidSortField(params.SortBy) {
    dir := strings.ToUpper(params.SortDir)
    if dir != "ASC" && dir != "DESC" {
      dir = "DESC"
    }
    query = query.Order(fmt.Sprintf("\"user\".%s %s", params.SortBy, dir))
  } else {
    query = query.Order("\"user\".id DESC")
  }

  // Obtener total para paginación
  var total int64
  if err := s.db.Table("\"user\"").Count(&total).Error; err != nil {
    return common.PaginatedResponse[dto.UserResponse]{}, err
  }

  // Aplicar paginación y ejecutar
  offset := (params.Page - 1) * params.PerPage
  var users []model.User
  if err := query.Offset(offset).Limit(params.PerPage).Find(&users).Error; err != nil {
    return common.PaginatedResponse[dto.UserResponse]{}, err
  }

  // Mapear a DTOs
  response := make([]dto.UserResponse, len(users))
  for i, u := range users {
    resp, err := s.mapUserToResponse(u)
    if err != nil {
      return common.PaginatedResponse[dto.UserResponse]{}, err
    }
    response[i] = *resp // ✅ Dereferenciar el puntero al asignar
  }

  return common.NewPaginatedResponse(response, int(total), params.Page, params.PerPage), nil
}

// GetByID retorna un usuario específico con metadata y perfil
func (s *Service) GetByID(id uint) (*dto.UserResponse, error) {
  var user model.User
  if err := s.db.First(&user, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, common.ErrNotFound
    }
    return nil, err
  }

  return s.mapUserToResponse(user)
}

// Create crea un nuevo usuario con validaciones y hash de password
func (s *Service) Create(name, email, password, phone string, state *int, profileID uint) (*dto.UserResponse, error) {
  // Validar unicidad de email
  if err := s.validateEmailUnique(0, email); err != nil {
    return nil, errors.New("email duplicado")
  }

  // Validar que el profile_id exista si se proporciona
  if profileID != 0 {
    if err := s.validateProfileExists(profileID); err != nil {
      return nil, errors.New("perfil no existe")
    }
  }

  // Hashear password con bcrypt
  hashedPassword, err := authpkg.HashPassword(password)
  if err != nil {
    return nil, errors.New("error al procesar contraseña")
  }

  // Calcular state por defecto
  stateValue := 1
  if state != nil {
    stateValue = *state
  }

  // Ejecutar en transacción
  var newUser model.User
  var newMeta model.UserMetadata

  err = s.db.Transaction(func(tx *gorm.DB) error {
    newUser = model.User{
      Name:     name,
      Email:    email,
      Password: hashedPassword,
    }
    if err := tx.Create(&newUser).Error; err != nil {
      return fmt.Errorf("error al crear usuario: %w", err)
    }

    // Crear metadata si hay datos relevantes
    if phone != "" || state != nil || profileID != 0 {
      newMeta = model.UserMetadata{
        UserID:    newUser.ID,
        Phone:     phone,
        State:     stateValue,
        ProfileID: profileID,
      }
      if err := tx.Create(&newMeta).Error; err != nil {
        return fmt.Errorf("error al guardar metadata: %w", err)
      }
    }
    return nil
  })

  if err != nil {
    return nil, err
  }

  // Retornar respuesta (sin password)
  return s.mapUserToResponse(newUser)
}

// Update actualiza un usuario existente con validaciones
func (s *Service) Update(id uint, name, email, password, phone string, state *int, profileID uint) (*dto.UserResponse, error) {
  // Buscar usuario existente
  var user model.User
  if err := s.db.First(&user, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return nil, common.ErrNotFound
    }
    return nil, err
  }

  // Validar unicidad de email (si cambia)
  if email != "" && !strings.EqualFold(email, user.Email) {
    if err := s.validateEmailUnique(user.ID, email); err != nil {
      return nil, errors.New("email duplicado")
    }
  }

  // Validar profile_id
  if profileID != 0 {
    if err := s.validateProfileExists(profileID); err != nil {
      return nil, errors.New("perfil no existe")
    }
  }

  // Ejecutar en transacción
  err := s.db.Transaction(func(tx *gorm.DB) error {
    // Actualizar campos del usuario
    updates := make(map[string]interface{})
    if name != "" {
      updates["name"] = name
    }
    if email != "" {
      updates["email"] = email
    }
    if password != "" {
      hashed, err := authpkg.HashPassword(password)
      if err != nil {
        return fmt.Errorf("error al hashear contraseña: %w", err)
      }
      updates["password"] = hashed
    }

    if len(updates) > 0 {
      if err := tx.Model(&user).Updates(updates).Error; err != nil {
        return fmt.Errorf("error al actualizar usuario: %w", err)
      }
    }

    // Actualizar o crear metadata (upsert)
    var meta model.UserMetadata
    if err := tx.Where("user_id = ?", user.ID).FirstOrCreate(&meta, model.UserMetadata{UserID: user.ID}).Error; err != nil {
      return fmt.Errorf("error al consultar metadata: %w", err)
    }

    metaUpdates := make(map[string]interface{})
    if phone != "" {
      metaUpdates["phone"] = phone
    }
    if state != nil {
      metaUpdates["state"] = *state
    }
    if profileID != 0 {
      metaUpdates["profile_id"] = profileID
    }

    if len(metaUpdates) > 0 {
      if err := tx.Model(&meta).Updates(metaUpdates).Error; err != nil {
        return fmt.Errorf("error al actualizar metadata: %w", err)
      }
    }
    return nil
  })

  if err != nil {
    return nil, err
  }

  // Recargar y retornar respuesta
  return s.mapUserToResponse(user)
}

// Delete elimina un usuario y su metadata
func (s *Service) Delete(id uint) error {
  var user model.User
  if err := s.db.First(&user, id).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return common.ErrNotFound
    }
    return err
  }

  err := s.db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Where("user_id = ?", id).Delete(&model.UserMetadata{}).Error; err != nil {
      return fmt.Errorf("error al eliminar metadata: %w", err)
    }
    if err := tx.Delete(&user).Error; err != nil {
      return fmt.Errorf("error al eliminar usuario: %w", err)
    }
    return nil
  })

  return err
}

// validateEmailUnique verifica que un email no esté registrado (excluyendo un ID opcional)
func (s *Service) validateEmailUnique(excludeID uint, email string) error {
  query := s.db.Model(&model.User{}).Where("email ILIKE ?", email)
  if excludeID != 0 {
    query = query.Where("id != ?", excludeID)
  }
  var existing model.User
  if err := query.First(&existing).Error; err == nil {
    return errors.New("email duplicado")
  }
  return nil
}

// validateProfileExists verifica que un profile_id exista
func (s *Service) validateProfileExists(profileID uint) error {
  var profile model.Profile
  if err := s.db.First(&profile, profileID).Error; err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
      return errors.New("perfil no existe")
    }
    return err
  }
  return nil
}

// mapUserToResponse convierte model.User a dto.UserResponse (sin password, con metadata)
func (s *Service) mapUserToResponse(user model.User) (*dto.UserResponse, error) {
  var meta model.UserMetadata
  if err := s.db.Where("user_id = ?", user.ID).Preload("Profile").First(&meta).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, fmt.Errorf("error al cargar metadata: %w", err)
  }

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
    // Modules se carga en otro endpoint (/auth/me)
  }, nil
}

// isValidSortField verifica si un campo está permitido para ordenamiento
func isValidSortField(field string) bool {
  allowed := map[string]bool{
    "id": true, "name": true, "email": true, "created_at": true,
  }
  return allowed[field]
}

// userListParams parámetros para listado de usuarios (privado del servicio)
type userListParams struct {
  Page      int
  PerPage   int
  Search    string
  Field     string
  State     *int
  ProfileID *uint
  SortBy    string
  SortDir   string
}