# Jenkins Deploy Setup

This deploy setup follows the pattern from `D:\personal\discordStorage\deploy`.

## Required Jenkins Credentials

Create these Jenkins credentials:

```text
target-host
```

Type: Secret text.

Value example:

```text
user@your-server-ip
```

```text
target-host-ssh
```

Type: SSH Username with private key.

Use the private key that can SSH into the target server.

## Server Requirements

Install on the target server:

```bash
docker
docker compose
nginx
postgresql client/tools if needed
```

PostgreSQL can run on the host, another server, or managed DB. If PostgreSQL runs on the Docker host, use:

```env
DB_HOST=host.docker.internal
```

## First Deployment

Jenkins will copy the repository to:

```text
~/ScrapingManga
```

On first deploy, the script creates:

```text
~/ScrapingManga/deploy/.env.production
```

Fill real values there, then rerun Jenkins.

Important values:

```env
DB_PASS=
DB_NAME=db_manga
ADMIN_TOKEN=
INGEST_INTERNAL_TOKEN=
BALSTORAGE_EMAIL=
BALSTORAGE_PASSWORD=
```

## Jenkins Job

Create a Pipeline job and point it to:

```text
deploy/Jenkinsfile
```

Pipeline stages:

```text
Checkout
Test Go backend + compile Python
Deploy via rsync and docker compose
Verify health + Swagger
```

## URLs

After deploy:

```text
http://<server>:8001/api/v1/health
http://<server>:8001/swagger
```

If using nginx, edit:

```text
deploy/nginx/manga-api.conf
```

The default domain is:

```text
manga.iqbalpradipta.my.id
```

Jenkins enables:

```text
/etc/nginx/sites-enabled/manga-api
```

Make sure the SSL certificate exists before the nginx reload step:

```bash
sudo certbot --nginx -d manga.iqbalpradipta.my.id
```
