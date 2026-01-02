# Harmony

A self-hosted music streaming application with a modern, Spotify-inspired interface.

![Harmony Screenshot](docs/screenshot.png)

## Features

- **Modern Web Interface** - Clean, responsive design built with SvelteKit and Tailwind CSS
- **Audio Streaming** - Stream your music library with support for seeking and quality selection
- **Library Management** - Automatic scanning and metadata extraction from your music files
- **Playlists** - Create, edit, and manage custom playlists
- **Search** - Fast search across tracks, albums, and artists
- **Transcoding** - On-the-fly audio transcoding with ffmpeg
- **Dark/Light Theme** - Customizable themes with accent color options
- **Keyboard Shortcuts** - Full keyboard navigation support
- **Container Ready** - Easy deployment with Docker or Podman

## Supported Formats

MP3, FLAC, WAV, OGG, M4A, AAC, OPUS, WMA

## Quick Start

### Prerequisites

- **Docker** (with Docker Compose) or **Podman** (with podman-compose)
- A directory containing your music library

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/harmony.git
   cd harmony
   ```

2. **Configure environment**
   ```bash
   cp .env.example .env
   ```

   Edit `.env` and set `MEDIA_PATH` to your music library:
   ```env
   MEDIA_PATH=/path/to/your/music
   ```

3. **Start the application**

   Using the Makefile (auto-detects runtime):
   ```bash
   make up
   ```

   Or using the shell script:
   ```bash
   ./harmony.sh up
   ```

   Or directly with your preferred runtime:
   ```bash
   # Docker
   docker compose up -d

   # Podman
   podman-compose up -d
   ```

4. **Access Harmony**

   Open http://localhost:3000 in your browser.

5. **Scan your library**

   Go to Library and click the scan button to index your music.

## Configuration

All configuration is done via environment variables. Copy `.env.example` to `.env` and customize:

| Variable | Default | Description |
|----------|---------|-------------|
| `MEDIA_PATH` | (required) | Path to your music library |
| `API_PORT` | `8080` | Backend API port |
| `FRONTEND_PORT` | `3000` | Frontend web port |
| `DB_PATH` | `/data/harmony.db` | SQLite database location |
| `REDIS_URL` | `redis://redis:6379` | Redis connection string |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `SCAN_ON_STARTUP` | `false` | Auto-scan library on startup |
| `TZ` | `UTC` | Timezone for timestamps |

See `.env.example` for all available options.

## Container Deployment

Harmony supports both **Docker** and **Podman** container runtimes. The Makefile and shell script automatically detect which runtime is available.

### Management Commands

Using the Makefile:

```bash
make up              # Start all services
make down            # Stop all services
make restart         # Restart all services
make build           # Build container images
make logs            # Follow container logs
make logs-backend    # Backend logs only
make logs-frontend   # Frontend logs only
make status          # Show service status
make scan            # Trigger library scan
make shell-backend   # Shell into backend container
make shell-frontend  # Shell into frontend container
make info            # Show runtime information
make clean           # Remove containers and images
make prune           # Remove unused resources
```

Using the shell script:

```bash
./harmony.sh up       # Start all services
./harmony.sh down     # Stop all services
./harmony.sh restart  # Restart all services
./harmony.sh build    # Build container images
./harmony.sh logs     # Follow logs
./harmony.sh status   # Show service status
./harmony.sh scan     # Trigger library scan
./harmony.sh shell backend   # Shell into backend
./harmony.sh info     # Show runtime information
```

### Direct Commands

**Docker:**
```bash
docker compose up -d
docker compose down
docker compose logs -f
```

**Podman:**
```bash
podman-compose up -d
podman-compose down
podman-compose logs -f

# Or with podman compose plugin
podman compose up -d
```

### Production with SSL (Traefik)

For production deployments with HTTPS:

1. Configure your domain in `.env`:
   ```env
   DOMAIN=music.example.com
   ACME_EMAIL=admin@example.com
   ```

2. Start with Traefik:
   ```bash
   make up-prod
   # or
   ./harmony.sh up-prod
   # or directly:
   docker compose -f docker-compose.yml -f docker-compose.traefik.yml up -d
   ```

### Development Mode

For local development with debug logging:

```bash
make up
```

The `docker-compose.override.yml` automatically enables development settings when present.

### Building Images

```bash
# Build with version tag
make build

# Build with all metadata (direct command)
docker compose build \
  --build-arg VERSION=1.0.0 \
  --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD)
```

## Development Setup

### Backend (Go)

```bash
cd backend
go mod download
go run ./cmd/server
```

### Frontend (SvelteKit)

```bash
cd frontend
npm install
npm run dev
```

The frontend development server runs at http://localhost:5173.

## API Reference

### Tracks

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/tracks` | List tracks (paginated) |
| GET | `/api/v1/tracks/:id` | Get track details |
| GET | `/api/v1/tracks/:id/stream` | Stream audio file |

### Albums

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/albums` | List albums |
| GET | `/api/v1/albums/:id` | Get album with tracks |

### Artists

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/artists` | List artists |
| GET | `/api/v1/artists/:id` | Get artist with albums |

### Playlists

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/playlists` | List playlists |
| POST | `/api/v1/playlists` | Create playlist |
| GET | `/api/v1/playlists/:id` | Get playlist with tracks |
| PUT | `/api/v1/playlists/:id` | Update playlist |
| DELETE | `/api/v1/playlists/:id` | Delete playlist |
| POST | `/api/v1/playlists/:id/tracks` | Add track to playlist |
| DELETE | `/api/v1/playlists/:id/tracks/:trackId` | Remove track |

### Search & Discovery

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/search?q=` | Global search |
| GET | `/api/v1/recent` | Recently added |
| GET | `/api/v1/random` | Random tracks/albums |

### Library Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/library/scan` | Start library scan |
| GET | `/api/v1/library/scan/status` | Get scan progress |
| POST | `/api/v1/library/scan/cancel` | Cancel running scan |
| GET | `/api/v1/library/stats` | Library statistics |

### Artwork

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/artwork/:type/:id` | Get artwork image |

Query parameters: `size` (thumbnail, small, medium, large)

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Space` | Play / Pause |
| `N` | Next track |
| `P` | Previous track |
| `M` | Mute / Unmute |
| `S` | Toggle shuffle |
| `R` | Toggle repeat |
| `Q` | Toggle queue panel |
| `Shift + ←/→` | Seek backward/forward |
| `Shift + ↑/↓` | Volume up/down |
| `/` | Focus search |

## Architecture

```
harmony/
├── backend/                 # Go backend API
│   ├── cmd/server/          # Application entry point
│   └── internal/
│       ├── config/          # Configuration management
│       ├── database/        # SQLite + Redis
│       ├── handlers/        # HTTP handlers
│       ├── models/          # Data models
│       ├── scanner/         # Media scanning
│       ├── services/        # Business logic
│       └── transcoder/      # Audio transcoding
├── frontend/                # SvelteKit frontend
│   └── src/
│       ├── lib/
│       │   ├── api/         # API client
│       │   ├── audio/       # Audio controller
│       │   ├── components/  # UI components
│       │   └── stores/      # Svelte stores
│       └── routes/          # Page routes
└── docker-compose.yml       # Docker orchestration
```

## Tech Stack

- **Backend**: Go 1.22, Gin, GORM, SQLite, Redis
- **Frontend**: SvelteKit 2, Svelte 5, Tailwind CSS 4, TypeScript
- **Audio**: ffmpeg for transcoding
- **Deployment**: Docker or Podman with Compose

## Troubleshooting

### Library not scanning

- Ensure `MEDIA_PATH` is correctly set and accessible
- Check that the media directory is mounted read-only in the container
- Review logs: `make logs-backend` or `docker compose logs backend`

### Audio not playing

- Verify the track file exists and is readable
- Check browser console for errors
- Ensure ffmpeg is installed in the backend container

### Redis connection issues

- Redis is optional; the app works without it (caching disabled)
- Check Redis container health: `make status` or `docker compose ps`

### Podman-specific issues

**SELinux denials (Fedora/RHEL):**
```bash
# Add :Z suffix to volume mounts if needed
podman-compose down
# Edit docker-compose.yml volumes to add :Z suffix
# e.g., ${MEDIA_PATH:-./media}:/media:ro,Z
podman-compose up -d
```

**Rootless Podman permissions:**
```bash
# Ensure your user owns the data directories
podman unshare chown -R $(id -u):$(id -g) ./data
```

**podman-compose not found:**
```bash
# Install via pip
pip install podman-compose

# Or use podman compose plugin (Podman 4.1+)
podman compose up -d
```

### Container resource limits

Resource limits are configured in `docker-compose.yml`. Adjust if needed:

```yaml
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 1G
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Commit changes: `git commit -am 'Add new feature'`
4. Push to branch: `git push origin feature/my-feature`
5. Submit a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Inspired by Spotify's user interface
- Icons by [Lucide](https://lucide.dev/)
- Font: [Inter](https://rsms.me/inter/)
