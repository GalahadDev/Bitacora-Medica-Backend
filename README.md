# Bitacora Médica - Backend

Este es el backend para la aplicación Bitacora Médica, una API RESTful desarrollada en Go.

## 🚀 Tecnologías

El proyecto utiliza las siguientes tecnologías principales:

- **Lenguaje**: [Go](https://go.dev/) (v1.25.4)
- **Framework Web**: [Gin Gonic](https://github.com/gin-gonic/gin)
- **Base de Datos**: PostgreSQL
- **ORM**: [GORM](https://gorm.io/)
- **Documentación API**: [Swagger](https://github.com/swaggo/swag)
- **Autenticación**: JWT (JSON Web Tokens)

## 📋 Requisitos Previos

Asegúrate de tener instalado:

- [Go](https://go.dev/dl/) 1.25 o superior
- [PostgreSQL](https://www.postgresql.org/)
- Git

## 🛠️ Instalación y Configuración

1. **Clonar el repositorio**

```bash
git clone <URL_DEL_REPOSITORIO>
cd bitacora-medica-backend
```

2. **Instalar dependencias**

```bash
go mod download
```

3. **Configurar Variables de Entorno**

Crea un archivo `.env` en la raíz del proyecto y configura las siguientes variables (basado en `api/config/config.go`):

```env
# Configuración de Base de Datos
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=tu_contraseña
DB_NAME=postgres
DB_PORT=5432

# Configuración del Servidor
PORT=8080

# Autenticación y Seguridad
JWT_SECRET=tu_secreto_super_seguro
SUPABASE_URL=tu_supabase_url
SUPABASE_SERVICE_ROLE_KEY=tu_supabase_key

# Configuración de Correo (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_EMAIL=tu_email@gmail.com
SMTP_PASSWORD=tu_contraseña_aplicacion
```

## ▶️ Ejecución

Para iniciar el servidor en modo desarrollo:

```bash
go run main.go
```

El servidor iniciará por defecto en `http://localhost:8080`.

## 📚 Documentación de la API

Una vez iniciado el servidor, puedes acceder a la documentación interactiva (Swagger UI) en:

(proximamente)

## 📂 Estructura del Proyecto

La estructura principal del código se encuentra en la carpeta `api/`:

- `api/config`: Carga y gestión de configuración.
- `api/database`: Conexión a la base de datos.
- `api/domains`: Definiciones de dominio y modelos de datos.
- `api/handlers`: Controladores de los endpoints HTTP.
- `api/middleware`: Middlewares (Autenticación, Rate limiting, etc).
- `api/services`: Lógica de negocio y servicios.
- `api/utils`: Utilidades generales.
