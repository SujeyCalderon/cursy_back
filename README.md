# Cursy Backend 📚

Backend API para **Cursy**, una aplicación de intercambio de cursos estilo "trueque". Los usuarios publican sus propios cursos para desbloquear el acceso al contenido de otros.

## 🛠️ Tecnologías

- **Go 1.21+** - Lenguaje de programación
- **Gin** - Framework web HTTP
- **MongoDB** - Base de datos NoSQL
- **JWT** - Autenticación con tokens

## 📁 Estructura del Proyecto

```
cursy_back/
├── config/         # Configuración de base de datos
├── controllers/    # Controladores HTTP
├── middleware/     # Middleware (JWT auth)
├── models/         # Modelos de datos
├── routes/         # Definición de rutas
├── utils/          # Utilidades (password hash, JWT)
└── main.go         # Punto de entrada
```

## 🚀 Instalación

### Prerrequisitos

- Go 1.21 o superior
- MongoDB (local o Atlas)

### Configuración

1. **Clonar el repositorio**
   ```bash
   git clone https://github.com/tu-usuario/cursy_back.git
   cd cursy_back
   ```

2. **Configurar variables de entorno**
   ```bash
   cp .env.example .env
   ```
   Edita `.env` con tus credenciales:
   ```env
   MONGO_URI=mongodb+srv://usuario:password@cluster.mongodb.net/
   DB_NAME=cursy_db
   JWT_SECRET=tu-clave-secreta-aqui
   PORT=8080
   ```

3. **Instalar dependencias**
   ```bash
   go mod download
   ```

4. **Ejecutar el servidor**
   ```bash
   go run main.go
   ```
   
   Para producción:
   ```bash
   GIN_MODE=release go run main.go
   ```

## 📡 API Endpoints

### Autenticación (Públicos)
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Registrar usuario |
| POST | `/api/v1/auth/login` | Iniciar sesión |
| POST | `/api/v1/auth/recover-password` | Recuperar contraseña |

### Autenticación (Protegidos)
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | `/api/v1/auth/logout` | Cerrar sesión |
| DELETE | `/api/v1/auth/account` | Eliminar cuenta |

### Cursos (Requieren JWT)
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/v1/courses` | Feed de cursos publicados |
| POST | `/api/v1/courses` | Crear nuevo curso |
| GET | `/api/v1/courses/:id` | Detalle del curso* |
| PUT | `/api/v1/courses/:id` | Actualizar curso propio |
| DELETE | `/api/v1/courses/:id` | Eliminar curso propio |
| PUT | `/api/v1/courses/:id/publish` | Publicar curso |
| POST | `/api/v1/courses/:id/save` | Guardar curso* |
| DELETE | `/api/v1/courses/:id/save` | Quitar de guardados |

*Requiere tener un curso publicado (lógica de trueque)

### Perfil (Requieren JWT)
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/v1/profile` | Mi perfil |
| PUT | `/api/v1/profile` | Actualizar perfil |
| GET | `/api/v1/profile/courses` | Mis cursos |
| GET | `/api/v1/profile/saved` | Cursos guardados |

### Health Check
| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/health` | Estado del servidor |

## 🔐 Autenticación

Incluye el token JWT en el header:
```
Authorization: Bearer <tu-token>
```

## 📝 Ejemplo de Uso

```bash
# Registrar usuario
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Password123","name":"Usuario","ine_url":"https://..."}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Password123"}'

# Crear curso (con token)
curl -X POST http://localhost:8080/api/v1/courses \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Mi Curso","description":"Descripción","blocks":[{"type":"header","content":"Intro"}]}'
```

## 🚀 Deploy

### Railway / Render / Fly.io

1. Configura las variables de entorno en el dashboard
2. El servidor escuchará en el puerto definido por `PORT`

### Docker (opcional)

```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .
EXPOSE 8080
CMD ["./main"]
```

## 📄 Licencia

MIT License

---

Desarrollado con ❤️ para el intercambio de conocimiento
