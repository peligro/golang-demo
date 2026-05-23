# 🚀 Tamila SAAS

> Plataforma de gestión empresarial con sistema de permisos granulares, autenticación segura y arquitectura escalable. Base sólida para integración de IA.

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-1.12-000000?logo=go&logoColor=white)](https://gin-gonic.com/)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react&logoColor=black)](https://react.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?logo=redis&logoColor=white)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-24-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)

---

## 📋 Tabla de Contenidos

- [✨ Características](#-características)
- [🏗️ Arquitectura](#️-arquitectura)
- [🛠️ Tecnologías](#️-tecnologías)
- [📁 Estructura del Proyecto](#-estructura-del-proyecto)
- [🚀 Instalación](#-instalación)
- [⚙️ Variables de Entorno](#️-variables-de-entorno)
- [🔐 Sistema de Permisos](#-sistema-de-permisos)
- [📚 Documentación API](#-documentación-api)
- [🧪 Testing](#-testing)
- [🤖 Próximos Pasos: IA](#-próximos-pasos-ia)
- [🤝 Contribuir](#-contribuir)
- [📄 Licencia](#-licencia)

---

## ✨ Características

### 🔐 Seguridad & Autenticación
- ✅ Autenticación con cookies HTTP-only + sesiones en Redis
- ✅ Validación de estado de usuario (activo/inactivo) en cada request
- ✅ Hash de contraseñas con bcrypt (costo 14)
- ✅ Middleware de autorización con respuestas genéricas (fail securely)
- ✅ Protección contra escalada de privilegios con `view_all_admin`

### 🎯 Sistema de Permisos Granulares
- ✅ Módulos configurables con slugs únicos (`/settings/users`, etc.)
- ✅ Items/acciones con códigos identificadores (`crear_usuario`, `editar_perfil`, etc.)
- ✅ Asignación flexible: Perfil → Módulos → Items
- ✅ Llave maestra: `view_all_admin` otorga acceso total

### 🏗️ Arquitectura Escalable
- ✅ Patrón **Handler + Service**: separación clara de responsabilidades
- ✅ Inyección de dependencias vía constructores
- ✅ Sin relaciones inversas en GORM: queries explícitas y controladas
- ✅ DTOs con validaciones y mensajes de error personalizados

### 📦 Módulos Incluidos
| Módulo | Descripción | Endpoints Principales |
|--------|------------|---------------------|
| 👤 **Users** | Gestión de usuarios con metadata | `GET/POST/PUT/DELETE /users` |
| 🔐 **Auth** | Login, logout, refresh, perfil | `POST /auth/login`, `GET /auth/me` |
| 🧩 **Profiles** | Roles/perfiles con permisos asignados | `GET/POST/PUT/DELETE /profiles` |
| 🧱 **Modules** | Definición de módulos del sistema | `GET/POST/PUT/DELETE /modules` |
| ⚡ **Items** | Acciones/permisos granulares | `GET/POST/PUT/DELETE /items` |
| 📊 **States** | Estados genéricos (activo, inactivo, etc.) | `GET/POST/PUT/DELETE /states` |
| 📑 **App Menu** | Menús dinámicos para sidebar | `GET /app-menu-all`, CRUD protegido |
| 🏠 **Home Menu** | Tarjetas para dashboard home | `GET /home-menu-all`, CRUD protegido |

### 🛡️ Calidad & DevOps
- ✅ Swagger automático con anotaciones completas
- ✅ Validaciones con `go-playground/validator`
- ✅ Docker compose para desarrollo (Go, Postgres, Redis, LocalStack)
- ✅ Límites de recursos configurados (CPU, RAM, procesos)
- ✅ `govulncheck` para escaneo de vulnerabilidades

---

## 🏗️ Arquitectura

```text
┌─────────────────────────────────────────┐
│ Frontend (React) │
│ • TypeScript + Bootstrap 5 │
│ • Consumo de API REST │
│ • Renderizado dinámico de menús │
└─────────────────┬───────────────────────┘
│ HTTPS/JSON
▼
┌─────────────────────────────────────────┐
│ Backend (Go + Gin) │
├─────────────────────────────────────────┤
│ 🌐 Router (app/router.go) │
│ ├─ Middleware: CORS, Security, Auth │
│ └─ Registro de rutas por módulo │
│ │
│ 🔐 middleware/ │
│ ├─ auth.go → Validación de sesión│
│ ├─ permission.go → Autorización granular│
│ └─ security.go → Headers OWASP │
│ │
│ 📦 routes/{module}/ │
│ ├─ handler.go → HTTP: bind/validate/respond│
│ ├─ service.go → Lógica de negocio + DB │
│ └─ routes.go → Registro de rutas + middleware│
│ │
│ 🗄️ model/ │
│ ├─ User, Profile, Module, Item, etc. │
│ └─ Sin relaciones inversas (queries explícitas)│
│ │
│ 📦 dto/ │
│ ├─ Request/Response structs │
│ └─ Validaciones + MensajesDeError() │
└─────────────────┬───────────────────────┘
│ GORM
▼
┌─────────────────────────────────────────┐
│ Capa de Datos │
├─────────────────────────────────────────┤
│ 🐘 PostgreSQL 15 + pgvector │
│ • Tablas: user, profile, module, item │
│ • Relaciones: Profile↔Module↔Item │
│ │
│ 🔴 Redis │
│ • Sesiones de usuario (TTL configurable)│
│ • Cache opcional para permisos │
└─────────────────────────────────────────┘
```
 

## 🛠️ Tecnologías

### Backend
| Tecnología | Versión | Propósito |
|-----------|---------|-----------|
| **Go** | 1.26 | Lenguaje principal |
| **Gin** | 1.12 | Framework HTTP |
| **GORM** | 1.31 | ORM para PostgreSQL |
| **go-playground/validator** | 10.30 | Validaciones de entrada |
| **bcrypt** | 0.51 | Hash de contraseñas |
| **go-redis** | 9.19 | Cliente Redis |
| **swaggo/swag** | 1.16 | Generación de Swagger |

### Frontend
| Tecnología | Versión | Propósito |
|-----------|---------|-----------|
| **React** | 18 | UI library |
| **TypeScript** | 5 | Tipado estático |
| **Bootstrap 5** | 5.3 | Componentes UI |
| **Vite** | 5 | Build tool |

### Infraestructura
| Tecnología | Propósito |
|-----------|-----------|
| **Docker + Docker Compose** | Contenedores para desarrollo |
| **PostgreSQL 15 + pgvector** | Base de datos relacional + embeddings |
| **Redis 7** | Cache y sesiones |
| **LocalStack** | Emulación de AWS para desarrollo |

---

## 📁 Estructura del Proyecto

```text
tamila-saas/
├── 📦 golang/ # Backend en Go
│ ├── 📁 app/ # Configuración principal
│ │ └── router.go # Registro de rutas + middleware globales
│ ├── 📁 common/ # Utilidades compartidas
│ │ ├── constants.go # Constantes: slugs, códigos de items
│ │ ├── validation.go # Validadores personalizados
│ │ └── pagination.go # Helper para paginación
│ ├── 📁 database/ # Configuración de DB
│ │ └── database.go # Conexión GORM + pool settings
│ ├── 📁 dto/ # Data Transfer Objects
│ │ ├── user_dto.go # Request/Response + validaciones
│ │ ├── module_dto.go
│ │ └── ...
│ ├── 📁 middleware/ # Middleware reutilizables
│ │ ├── auth.go # Validación de sesión (cookie + Redis)
│ │ ├── permission.go # Autorización granular (RequireModule/Item)
│ │ └── security.go # Headers de seguridad OWASP
│ ├── 📁 model/ # Modelos GORM (sin relaciones inversas)
│ │ ├── user.go
│ │ ├── profile.go
│ │ ├── module.go
│ │ └── ...
│ ├── 📁 pkg/ # Paquetes internos
│ │ └── auth/ # Helpers de auth: hash, sesiones Redis
│ ├── 📁 routes/ # Módulos de negocio
│ │ ├── auth/ # Login, logout, me, refresh
│ │ ├── user/ # CRUD de usuarios
│ │ ├── profile/ # CRUD + gestión de permisos anidados
│ │ ├── module/ # CRUD de módulos (restringido)
│ │ ├── item/ # CRUD de items/acciones
│ │ ├── state/ # CRUD de estados
│ │ ├── app_menu/ # Menús para sidebar dinámico
│ │ └── home_menu/ # Tarjetas para dashboard
│ ├── 📁 docs/ # Documentación Swagger (auto-generada)
│ ├── main.go # Entry point
│ ├── go.mod # Dependencias Go
│ └── Dockerfile # Imagen para desarrollo
│
├── 📦 react/ # Frontend en React (estructura base)
│ ├── src/
│ │ ├── components/ # Componentes reutilizables
│ │ ├── pages/ # Vistas principales
│ │ ├── services/ # Clientes API (axios)
│ │ ├── context/ # AuthContext, PermissionsContext
│ │ └── utils/ # Helpers de frontend
│ ├── package.json
│ └── vite.config.ts
│
├── 🐳 docker/ # Configuración Docker
│ ├── golang/Dockerfile
│ ├── python/Dockerfile
│ └── nginx/default.conf
│
├── docker-compose.yml # Orquestación de servicios
├── .gitignore
└── README.md
```


## 🚀 Instalación

### Requisitos Previos
- Docker + Docker Compose v2+
- Go 1.26+ (para desarrollo local sin Docker)
- Node.js 18+ y npm/yarn (para frontend)

### 1. Clonar el repositorio
```bash
git clone https://github.com/tu-usuario/tamila-saas.git
cd tamila-saas
```

### 2. Copiar ejemplo de backend
```bash
cp golang/.env.example golang/.env
```

# Editar golang/.env con tus configuraciones:
```env
# Copiar y editar este archivo:
# DB_HOST, DB_USER, DB_PASSWORD, REDIS_URL, etc.
ENVIRONMENT=local
...
```

### 3. Levantar servicios con Docker

```bash
# Construir e iniciar contenedores
docker-compose up -d --build

# Ver logs en tiempo real
docker-compose logs -f go-dev
```

### 4. Acceder a los servicios

| Servicio | URL | Credenciales (dev) |
|----------|-----|-------------------|
| 🌐 API Backend | `http://localhost:8082` | - |
| 📚 Swagger UI | `http://localhost:8082/swagger` | - |
| 🐘 PostgreSQL | `localhost:5432` | `laravel` / `secret` |
| 🔴 Redis | `localhost:6379` | - |
| 🗂️ pgAdmin | `http://localhost:5050` | `admin@example.com` / `admin` |



### 5. (Opcional) Desarrollo local sin Docker

```bash
# Backend
cd golang
go mod tidy
go run main.go

# Frontend (en otra terminal)
cd react
npm install
npm run dev
```


## ⚙️ Variables de Entorno
### Backend (golang/.env)

```env
# Entorno
ENVIRONMENT=local          # local, staging, production
PORT=8080

# Base de datos PostgreSQL
DB_HOST=postgres
DB_PORT=5432
DB_NAME=laravel
DB_USER=laravel
DB_PASSWORD=secret
TIMEZONE=America/Santiago

# Redis para sesiones
REDIS_URL=redis://redis:6379/0
SESSION_TTL=86400          # 24 horas en segundos

# Cookies
COOKIE_DOMAIN=localhost
COOKIE_SECURE=false        # true en producción con HTTPS

# Frontend para CORS
FRONTEND_URL=http://localhost:5173
```

### Frontend (react/.env)
```env
VITE_API_BASE_URL=http://localhost:8082
VITE_APP_NAME=Tamila SAAS
```

## 🔐 Sistema de Permisos

### Conceptos Clave

- **Module**: Recurso del sistema (ej: `/settings/users`)
- **Item**: Acción específica dentro de un módulo (ej: `crear_usuario`)
- **Profile**: Rol que agrupa módulos + items asignados
- **view_all_admin**: Item especial que otorga acceso total (llave maestra)

```text
Usuario → UserMetadata → Profile → [ProfileModule] → Module
                                      ↓
                              [ProfileModuleItem] → Item
```

### Flujo de Autorización

```mermaid
graph LR
    A[Request] --> B{AuthMiddleware?}
    B -->|❌ No autenticado | C[401 "No autenticado"]
    B -->|✅ Autenticado | D{RequirePermission?}
    D -->|❌ Sin permisos | E[401 "No autenticado"]
    D -->|✅ Con permisos | F[Handler]
    F --> G[Response]
```

### Ejemplo: Asignar permisos a un perfil

```bash
# 1. Crear perfil "Editor"
curl -X POST http://localhost:8082/profiles \
  -H "Content-Type: application/json" \
  -d '{"name":"Editor","description":"Puede editar contenido"}'

# 2. Asignar módulo "Items" al perfil (ID=2)
curl -X PUT http://localhost:8082/profiles/2/modules \
  -H "Content-Type: application/json" \
  -d '{"modules":[3]}'  # 3 = ID del módulo "Items"

# 3. Asignar items específicos dentro del módulo
curl -X PUT http://localhost:8082/profiles/2/modules/3/items \
  -H "Content-Type: application/json" \
  -d '{"items":[1,2]}'  # 1="ver_items", 2="editar_item"
```

### Respuesta de /auth/me (ejemplo)

```json
{
  "user": {
    "id": 1,
    "name": "Admin",
    "profile": { "id": 1, "name": "Super Admin" },
    "modules": [
      {
        "name": "Users",
        "slug": "/settings/users",
        "items": [
          { "id": 1, "name": "Ver usuarios", "code": "view_all_admin" }
        ]
      }
    ]
  }
}
```

## 📚 Documentación API


### Swagger UI

### Accede a la documentación interactiva en:

```text
http://localhost:8082/swagger/index.html?dark=true
```
### Regenerar documentación

```bash
# Dentro del contenedor Go
docker exec -it go-dev swag init --parseDependency --parseInternal

# O localmente (si tienes swag instalado)
cd golang
swag init --parseDependency --parseInternal
```

### Endpoints Destacados
| Método | Endpoint | Descripción | Auth Requerido |
|--------|----------|-------------|---------------|
| `POST` | `/auth/login` | Iniciar sesión | ❌ No |
| `GET`  | `/auth/me` | Obtener perfil + permisos | ✅ Sí |
| `GET`  | `/app-menu-all` | Menús para sidebar | ❌ No (público) |
| `GET`  | `/home-menu-all` | Tarjetas para dashboard | ❌ No (público) |
| `GET`  | `/users` | Listar usuarios | ✅ Sí + `ModuleUsers` |
| `POST` | `/users` | Crear usuario | ✅ Sí + `crear_usuario` |

> 🔍 **Nota**: Los endpoints protegidos retornan `401 {"message":"No autenticado"}` tanto para usuarios no autenticados como para usuarios sin permisos (fail securely).

### 🔹 Sección: Testing


## 🧪 Testing

### Backend (Go)
```bash
# Ejecutar tests unitarios
cd golang
go test ./...

# Con cobertura
go test -cover ./...

# Escanear vulnerabilidades
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```


## Frontend (React)

```bash
cd react
npm test          # Tests con Jest/Vitest
npm run lint      # ESLint
npm run typecheck # TypeScript
```

## E2E (Futuro)

- Playwright/Cypress para flujos críticos
- Tests de integración API + DB

### 🔹 Sección: Próximos Pasos: IA


## 🤖 Próximos Pasos: IA

Con el core sólido, los siguientes módulos integrarán IA como portafolio:

### 🎯 Módulo 1: Clasificador de Items con IA
```go
// Input: Descripción en lenguaje natural
// Output: Sugerencia de name, code y module

POST /ai/suggest-item
{
  "description": "Permitir que los administradores creen nuevos usuarios"
}
// Response:
{
  "suggestions": {
    "name": "Crear Usuario",
    "code": "crear_usuario", 
    "module_slug": "/settings/users",
    "confidence": 0.94
  }
}
```

### 🎯 Módulo 2: Generador de Descripciones con LLM

```go
// Input: Nombre de perfil + módulos asignados
// Output: Descripción profesional generada

POST /ai/generate-profile-description
{
  "profile_name": "Analista de Ventas",
  "modules": ["/settings/reports", "/settings/customers"]
}
// Response:
{
  "description": "Perfil diseñado para analistas que requieren acceso a reportes de ventas y gestión de clientes, sin permisos administrativos..."
}
```

### 🎯 Módulo 3: Detector de Conflictos de Permisos

```go
// Input: Perfil con items asignados
// Output: Alertas de inconsistencias

POST /ai/audit-permissions
{
  "profile_id": 5
}
// Response:
{
  "warnings": [
    "El item 'eliminar_usuario' requiere 'ver_usuarios' para ser útil",
    "Perfil sin acceso a 'reports' pero con 'export_data'"
  ],
  "score": 78
}
```

> 🚀 Stack de IA propuesto: Ollama (local) + modelos ligeros (Llama 3, Mistral) + LangChain para orquestación.

### 🔹 Sección: Contribuir


## 🤝 Contribuir

1. Fork el repositorio
2. Crea tu rama de feature (`git checkout -b feature/amazing-feature`)
3. Commit tus cambios (`git commit -m 'feat: add amazing feature'`)
4. Push a la rama (`git push origin feature/amazing-feature`)
5. Abre un Pull Request

### Convenciones de Commit
```text
feat:     Nueva funcionalidad
fix:      Corrección de bug
refactor: Mejora de código sin cambiar comportamiento
docs:     Solo documentación
test:     Agregar o corregir tests
chore:    Tareas de mantenimiento (deps, config, etc.)
```

### 🔹 Sección: Licencia + Footer


## 📄 Licencia

Distribuido bajo la licencia MIT. Ver `LICENSE` para más información.

---

> 💡 **Hecho con ❤️ por César Cancino**  
> *"Construyendo el futuro, un commit a la vez"*

---

## ✅ Checklist para antes de lanzar

- [ ] Revisar que `.env.example` esté actualizado
- [ ] Verificar que `govulncheck` no reporte vulnerabilidades críticas
- [ ] Probar flujo completo: login → permisos → CRUD → logout
- [ ] Validar que Swagger se genera sin errores
- [ ] Confirmar que los límites de Docker (ulimits) están aplicados

## 🚀 Para usarlos:

- Copia cada bloque de código
- Pégalo al final de tu README.md en el orden mostrado
- Guarda y commitea:

```bash
git add README.md
git commit -m "docs: completar README con testing, roadmap IA y guía de contribución"
git push origin main
```