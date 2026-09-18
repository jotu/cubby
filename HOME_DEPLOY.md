# HOME_DEPLOY (Pi 5, local network only)

This is the recommended home setup for Cubby: **simple, maintainable, extendable**.

## Stack
- Raspberry Pi 5 (16GB recommended)
- Raspberry Pi OS 64-bit
- Docker Compose
- SQLite data on host path (`/opt/cubby/data`)

## Migration strategy (simple option)
- Use Cubby’s **built-in migrations** on startup (`backend/internal/db/db.go`).
- No extra Flyway/Liquibase service needed for this home setup.
- Upgrade routine: **backup DB → deploy new version → app starts and applies pending migrations**.

---

## 1) Host setup

```bash
sudo apt update && sudo apt -y upgrade
sudo apt -y install git curl sqlite3 ufw
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
sudo reboot
```

## 2) Deploy Cubby

```bash
sudo mkdir -p /opt/cubby && sudo chown -R $USER:$USER /opt/cubby
cd /opt/cubby
git clone https://github.com/joacim/cubby.git app
mkdir -p /opt/cubby/data/uploads /opt/cubby/backups
cd /opt/cubby/app
```

Create `compose.home.yml`:

```yaml
services:
  cubby:
    build: .
    ports:
      - "192.168.1.50:8080:8080"
    volumes:
      - /opt/cubby/data:/data
    environment:
      - CUBBY_PORT=8080
      - CUBBY_DB_PATH=/data/cubby.db
      - CUBBY_UPLOAD_DIR=/data/uploads
    restart: unless-stopped
    read_only: true
    security_opt:
      - no-new-privileges:true
    cap_drop:
      - ALL
```

Start:

```bash
docker compose -f compose.home.yml up -d --build
docker compose -f compose.home.yml ps
curl -fsS http://192.168.1.50:8080/healthz
```

The health endpoint is a liveness check. A successful response confirms the
container is serving HTTP; it does not replace the backup and restore checks
below.

## 3) LAN-only protection

```bash
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow from 192.168.1.0/24 to any port 8080 proto tcp
sudo ufw enable
```

Also: no router port-forwarding, no UPnP.

## 4) Backup / restore

Backup:

```bash
ts=$(date +%F-%H%M%S)
sqlite3 /opt/cubby/data/cubby.db \
  "PRAGMA wal_checkpoint(FULL);" \
  ".backup '/opt/cubby/backups/cubby-${ts}.db'"
```

Restore:

```bash
cd /opt/cubby/app
docker compose -f compose.home.yml stop
cp /opt/cubby/backups/cubby-YYYY-MM-DD-HHMMSS.db /opt/cubby/data/cubby.db
docker compose -f compose.home.yml up -d
```

## 5) Upgrade routine

```bash
cd /opt/cubby/app
# 1) backup first
ts=$(date +%F-%H%M%S)
sqlite3 /opt/cubby/data/cubby.db "PRAGMA wal_checkpoint(FULL);" \
  ".backup '/opt/cubby/backups/cubby-${ts}.db'"

# 2) update and redeploy
git pull
docker compose -f compose.home.yml up -d --build
docker compose -f compose.home.yml ps
curl -fsS http://192.168.1.50:8080/healthz
```

## 6) If you still want Flyway/Liquibase

- For SQLite home use, this adds complexity with little benefit.
- If you must choose one, prefer **Flyway** over Liquibase for SQLite.
- Keep it as a separate one-shot admin task, not always-on service.
