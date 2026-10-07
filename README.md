# Chedul-core

API de Chedul en Go (Echo + Bun + Postgres). Maneja cuentas de alumnos, estado
académico con correlativas, el módulo de aportes (apuntes compartidos) y el
calendario (fechas de parciales y horario de cursada).

## Desarrollo local

Con Docker:

```sh
docker compose up --build
```

La API queda en http://localhost:8080 y aplica las migraciones sola al iniciar
(incluye el plan de Ingeniería en Sistemas con sus 74 correlativas).

Sin Docker, con un Postgres propio en `127.0.0.1:5432` (usuario y clave
`chedul`, base `chedul-db`):

```sh
CONFIG_DIR=configuration/ JWT_SECRET=secreto-de-desarrollo go run ./cmd/api
```

## Tests

Los tests de la API necesitan un Postgres. Hay dos formas:

```sh
# Con Docker: cada suite levanta un contenedor con testcontainers
just test

# Con un Postgres ya levantado: se crea una base nueva por suite
TEST_DATABASE_URL="postgres://chedul:chedul@127.0.0.1:5432/postgres?sslmode=disable" \
  CONFIG_DIR=../configuration/ go test ./...
```

## Variables de entorno

| Variable | Obligatoria | Para qué |
| --- | --- | --- |
| `JWT_SECRET` | Sí | Firma de los tokens. En producción, mínimo 32 caracteres. |
| `DATABASE_URL` | En producción | Conexión completa a Postgres. Si no está, se usa `configuration/*.yml` + `DB_HOST` / `DB_PASSWORD`. |
| `CORS_ORIGINS` | En producción | Dominios del frontend separados por coma, por ejemplo `https://chedul.vercel.app`. |
| `SERVER_ENV` | No | `development` (default), `test` o `production`. En producción la cookie de sesión es `Secure` y `SameSite=None`. |
| `PORT` | No | Puerto. Fly, Render y Railway lo definen solos. |
| `UPLOADS_DIR` | No | Carpeta de los archivos subidos. Default `uploads`; en Docker `/uploads`. |
| `HONEYCOMB_API_KEY` | No | Si está, se mandan trazas a Honeycomb. |

## Deploy

La forma más simple, y gratis para empezar:

1. **Base de datos en [Neon](https://neon.tech)**: crear un proyecto y copiar
   la connection string (`postgres://...?sslmode=require`).
2. **API en [Fly.io](https://fly.io)** con el `fly.toml` de este repo:
   ```sh
   fly launch --no-deploy --copy-config
   fly volumes create chedul_uploads --size 1
   fly secrets set JWT_SECRET="$(openssl rand -hex 32)" \
     DATABASE_URL="postgres://..." \
     CORS_ORIGINS="https://tu-front.vercel.app"
   fly deploy
   ```
   El volumen guarda los archivos de los aportes; sin él se pierden en cada deploy.
3. **Frontend en Vercel o Netlify** (ver el README de chedul-frontend) con
   `VITE_API_URL=https://chedul-core.fly.dev/api/v1`.

Railway o Render también sirven: usan el mismo `Dockerfile`, hay que definir
las mismas variables y montar un volumen en `/uploads`.

## API

Base: `/api/v1`. Las rutas marcadas con 🔒 piden `Authorization: Bearer <accessToken>`.

| Método | Ruta | Qué hace |
| --- | --- | --- |
| POST | `/signup` | Crea una cuenta (`nombre`, `email`, `carrera_id`, `password`) |
| POST | `/login` | Devuelve `accessToken` (15 min) y deja la cookie `refreshToken` (30 días) |
| POST | `/refresh-token` | Nuevo `accessToken` usando la cookie |
| POST | `/logout` | Borra la cookie |
| GET | `/carreras`, `/carreras/:id` | Carreras |
| GET | `/materias?carrera_id=` , `/materias/:id` | Materias con sus correlativas |
| GET | `/condicion` | Condiciones posibles (Cursando, Regularizada, Aprobada) |
| GET/PUT/DELETE 🔒 | `/alumnos/me` | Perfil propio |
| GET 🔒 | `/alumnos/me/progreso` | Materias aprobadas, regulares, cursando y cuáles se pueden cursar |
| GET 🔒 | `/condicion_alumno` | Mi condición en cada materia |
| PUT/DELETE 🔒 | `/condicion_alumno/:materia_id` | Guardar (`condicion_id`, `nota`) o volver a pendiente |
| GET 🔒 | `/aportes` | Listado. Filtros: `materia_id`, `tag_id`, `q`, `mios=1`, `favoritos=1`, `orden=recientes\|populares`, `pagina`, `limite` |
| POST 🔒 | `/aportes` | Subir (multipart: `titulo`, `descripcion`, `materia_id`, `tag_id`, `link` y/o `archivo`) |
| GET 🔒 | `/aportes/tags` | Tipos de aporte |
| GET/PUT/DELETE 🔒 | `/aportes/:id` | Ver, editar o borrar (solo el autor edita y borra) |
| GET 🔒 | `/aportes/:id/archivo` | Descargar el archivo |
| POST/DELETE 🔒 | `/aportes/:id/favorito` | Marcar o desmarcar favorito |
| GET 🔒 | `/eventos?desde=AAAA-MM-DD&hasta=AAAA-MM-DD` | Mis eventos en un rango (hasta 400 días) |
| POST 🔒 | `/eventos` | Nuevo evento (`titulo`, `tipo`: parcial/final/entrega/recordatorio/otro, `fecha`, `hora` HH:MM opcional, `materia_id` opcional, `descripcion`) |
| PUT/DELETE 🔒 | `/eventos/:id` | Editar o borrar un evento propio |
| GET 🔒 | `/clases` | Mi horario semanal |
| POST 🔒 | `/clases` | Nueva clase (`titulo`, `dia` 1=lunes…7=domingo, `hora_inicio`, `hora_fin`, `aula`, `materia_id` opcional) |
| PUT/DELETE 🔒 | `/clases/:id` | Editar o borrar una clase propia |

Archivos permitidos: PDF, imágenes (PNG, JPG, WEBP), Word, Excel, PowerPoint,
TXT, Markdown y ZIP, hasta 25 MB (configurable en `uploads.max_size_mb`).
