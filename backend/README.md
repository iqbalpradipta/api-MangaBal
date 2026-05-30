# Manga API Backend

Go Echo backend for Manga API with MVC structure, background Python ingest jobs, and BalStorage-backed manga page storage.

## Setup

1. Copy environment file:

```powershell
Copy-Item .env.example .env
```

2. Fill these values:

```env
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASS=
DB_NAME=manga_api
ADMIN_TOKEN=change-this-admin-token
INGEST_INTERNAL_TOKEN=change-this-secret
BALSTORAGE_BASE_URL=http://localhost:8000/api/v1
BALSTORAGE_EMAIL=admin@example.com
BALSTORAGE_PASSWORD=password123
```

3. Run the API:

```powershell
go run .
```

## Important Endpoints

Swagger UI:

```text
GET /swagger
https://manga.iqbalpradipta.my.id/swagger
```

Public read endpoints:

```text
GET /api/v1/health
GET /api/v1/manga?page=1&limit=20
GET /api/v1/manga/search?q=solo
GET /api/v1/manga/:slug
GET /api/v1/manga/:slug/chapters
GET /api/v1/manga/:slug/chapters/:chapter
GET /api/v1/genres
```

Admin ingest endpoints require:

```http
X-Admin-Token: <ADMIN_TOKEN>
```

```text
POST /api/v1/admin/ingest/all
POST /api/v1/admin/ingest/series
POST /api/v1/admin/ingest/chapter
GET /api/v1/admin/ingest/jobs
GET /api/v1/admin/ingest/jobs/:id
POST /api/v1/admin/ingest/jobs/:id/cancel
```

Internal endpoints are called by Python and require:

```http
X-Internal-Token: <INGEST_INTERNAL_TOKEN>
```

## Safer First Test

Start with a single chapter before running all ingest:

```powershell
Invoke-RestMethod `
  -Method Post `
  -Headers @{ "X-Admin-Token" = "<ADMIN_TOKEN>" } `
  -ContentType "application/json" `
  -Body '{"slug":"mumumu","chapter":1}' `
  http://localhost:8001/api/v1/admin/ingest/chapter
```

After the job finishes, read the data:

```text
GET /api/v1/manga/mumumu/chapters/1
```
