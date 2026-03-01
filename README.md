# LeadFlowCRM - Sistema de Gestión de Leads

LeadFlowCRM es un sistema profesional de gestión de clientes potenciales (leads) construido con Go, MongoDB y Bootstrap 5.

## Características

- **Gestión de Leads**: CRUD completo con estados configurables
- **Asignación a Vendedores**: Asigna leads a tu equipo de ventas
- **Historial de Cambios**: Seguimiento completo de modificaciones
- **Notas Internas**: Agregar notas privadas por cada lead
- **Dashboard Analítico**: Métricas en tiempo real con gráficos
- **Importación CSV**: Importa múltiples leads desde archivos CSV
- **Paginación**: Manejo eficiente de grandes cantidades de datos
- **Soft Delete**: Eliminación lógica con recuperación

## Tecnologías

- **Backend**: Go 1.21 con arquitectura limpia
- **Base de Datos**: MongoDB 7.0
- **Frontend**: HTML5, Bootstrap 5, Chart.js
- **Autenticación**: JWT con roles (admin/user)
- **Docker**: Multi-stage build con docker-compose

## Estructura del Proyecto

```
/
├── cmd/server/           # Punto de entrada
├── internal/
│   ├── config/         # Configuración
│   ├── handlers/       # Controladores HTTP
│   ├── services/      # Lógica de negocio
│   ├── repositories/  # Acceso a datos
│   ├── models/         # Modelos de datos
│   ├── middlewares/    # Middlewares HTTP
│   └── utils/          # Utilidades
├── web/
│   ├── templates/      # Plantillas HTML
│   └── static/         # CSS, JS, imágenes
├── .env                # Variables de entorno
├── Dockerfile          # Imagen Docker
├── docker-compose.yml  # Orquestación
└── go.mod              # Dependencias Go
```

## Requisitos

- Go 1.21+
- MongoDB 7.0 (local o remoto)

## Instalación y Ejecución Local

```bash
# Clonar el repositorio
git clone <repo-url>
cd Programaci-n-Estructurada-C-

# Asegúrate de que MongoDB esté ejecutándose en localhost:27017
# O actualiza MONGODB_URI en el archivo .env

# Instalar dependencias
go mod tidy

# Ejecutar el servidor
go run cmd/server/main.go

# La aplicación estará disponible en http://localhost:8080
```

### Con Docker (Alternativo)

```bash
# Iniciar servicios
docker-compose up --build

# La aplicación estará disponible en http://localhost:8080
```

## API Endpoints

### Autenticación
- `POST /api/auth/register` - Registrar usuario
- `POST /api/auth/login` - Iniciar sesión

### Usuarios (requiere auth)
- `GET /api/users/profile` - Perfil del usuario actual
- `GET /api/users` - Listar todos los usuarios

### Leads (requiere auth)
- `GET /api/leads` - Listar leads con filtros y paginación
- `POST /api/leads` - Crear lead
- `GET /api/leads/:id` - Obtener lead por ID
- `PUT /api/leads/:id` - Actualizar lead
- `DELETE /api/leads/:id` - Eliminar lead (soft delete)
- `GET /api/leads/:id/history` - Historial de cambios
- `POST /api/leads/:id/notes` - Agregar nota
- `GET /api/leads/:id/notes` - Listar notas
- `DELETE /api/leads/:id/notes/:noteId` - Eliminar nota
- `GET /api/leads/stats` - Estadísticas del dashboard
- `POST /api/leads/import` - Importar CSV

## Colecciones MongoDB

- `users` - Usuarios del sistema
- `leads` - Leads/clientes potenciales
- `lead_history` - Historial de cambios
- `lead_notes` - Notas internas

## Seguridad

- JWT Authentication
- Roles: admin, user
- Rate limiting
- CORS configurado
- Hash de contraseñas con bcrypt
- Índices en email y teléfono

## Dashboard

El dashboard incluye:
- Métricas principales (total, asignados, sin asignar, valor)
- Gráfico de estado por estado (Chart.js)
- Gráfico de Leads por fuente
- Tabla de últimos leads con paginación

---

## 👨‍💻 Desarrollado por Isaac Esteban Haro Torres

**Ingeniero en Sistemas · Full Stack · Automatización · Data**

- 📧 Email: zackharo1@gmail.com
- 📱 WhatsApp: 098805517
- 💻 GitHub: https://github.com/ieharo1
- 🌐 Portafolio: https://ieharo1.github.io/portafolio-isaac.haro/

---

© 2026 Isaac Esteban Haro Torres - Todos los derechos reservados.
