package common

// Códigos de items especiales para permisos globales
const (
  ViewAllAdminCode = "view_all_admin"

  ModuleUsers     = "/settings/users"
  ModuleProfiles  = "/settings/profiles"  // ← Esta es la que usamos
  ModuleItems     = "/settings/items"
  ModuleModules   = "/settings/modules"
  ModuleStates    = "/settings/states"
  ModuleAppMenu   = "/settings/app-menu"
  ModuleHome      = "/settings/home-menu"
)

// Items/permisos granulares para el módulo Profiles
const (
  // CRUD básico de perfiles
  ProfileCreate  = "crear_perfil"
  ProfileEdit    = "editar_perfil"
  ProfileDelete  = "eliminar_perfil"
  
  // Gestión de módulos asignados a perfiles
  ProfileAssignModules   = "asignar_modulos"
  ProfileUnassignModules = "desasignar_modulos"
  
  // Gestión de permisos granulares (items) dentro de módulos de perfil
  ProfileAssignPermissions   = "asignar_permisos"
  ProfileUnassignPermissions = "desasignar_permisos"
)