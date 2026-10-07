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
| `CORS_ORIGINS` | Solo si el front llama directo a la API | Dominios del frontend separados por coma. Con el proxy de Vercel no hace falta. |
| `SERVER_ENV` | No | `development` (default), `test` o `production`. En producción la cookie de sesión es `Secure` y `SameSite=None`. |
| `PORT` | No | Puerto. Cloud Run, Render y Railway lo definen solos. |
| `UPLOADS_DIR` | No | Carpeta de los archivos subidos. Default `uploads`; en Docker `/uploads`. |
| `UPLOADS_DISABLED` | No | `true` apaga la subida de archivos y los aportes son solo links. Para hostings sin disco, como Cloud Run. |
| `HONEYCOMB_API_KEY` | No | Si está, se mandan trazas a Honeycomb. |

## Deploy

Gratis y rápido: **Neon** (base de datos) + **Google Cloud Run** (API) +
**Vercel** (frontend). Cloud Run apaga la API cuando nadie la usa y la prende
en menos de un segundo, porque la imagen es un binario de Go. La guía paso a
paso, con capturas de qué elegir en cada pantalla, está en el doc del proyecto.

1. **Neon** (https://neon.tech): creá un proyecto en la región São Paulo y
   copiá la connection string (`postgresql://...?sslmode=require`).
2. **Cloud Run**: con el [CLI de gcloud](https://cloud.google.com/sdk/docs/install)
   instalado y un proyecto con facturación activada (el free tier no cobra,
   pero Google pide tarjeta):
   ```sh
   gcloud config set project TU_PROYECTO
   gcloud services enable run.googleapis.com cloudbuild.googleapis.com artifactregistry.googleapis.com
   ./scripts/deploy-cloudrun.sh
   ```
   La primera vez pide la `DATABASE_URL`, genera el `JWT_SECRET` y deja los
   aportes en modo solo links (`UPLOADS_DISABLED=true`, Cloud Run no tiene
   disco persistente). Al final muestra la URL de la API.
3. **Vercel**: importar `chedul-frontend`, poner la URL de Cloud Run en su
   `vercel.json` y `VITE_API_URL=/api/v1` (ver su README).

**Deploy automático:** `.github/workflows/deploy_cloud_run.yml` despliega en
cada merge a main una vez que existan la variable `GCP_PROJECT` y el secret
`GCP_SA_KEY` en el repo. Sin eso no hace nada.

Las migraciones se aplican solas al iniciar, así que no hay que correr nada
contra la base.

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
| GET 🔒 | `/aportes/config` | Si se pueden subir archivos (`subida_archivos`) y el tamaño máximo |
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
