# Harmony Frontend

The web interface for Harmony, built with SvelteKit and Tailwind CSS.

## Tech Stack

- **Framework**: SvelteKit 2 with Svelte 5
- **Styling**: Tailwind CSS 4
- **Language**: TypeScript
- **Icons**: Lucide Svelte

## Development

### Prerequisites

- Node.js 20+
- npm or pnpm

### Setup

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

The development server runs at http://localhost:5173.

### Environment Variables

Create a `.env` file for local development:

```env
PUBLIC_API_URL=http://localhost:8080/api/v1
```

## Project Structure

```
src/
├── routes/                 # SvelteKit pages
│   ├── +layout.svelte      # Root layout
│   ├── +page.svelte        # Home page
│   ├── album/[id]/         # Album detail
│   ├── artist/[id]/        # Artist detail
│   ├── library/            # Library page
│   ├── playlist/[id]/      # Playlist detail
│   ├── search/             # Search page
│   └── settings/           # Settings page
├── lib/
│   ├── api/                # API client modules
│   ├── audio/              # Audio playback
│   ├── components/
│   │   ├── layout/         # Layout components
│   │   └── ui/             # Base UI components
│   ├── stores/             # Svelte stores
│   └── utils/              # Utility functions
├── app.css                 # Global styles
└── app.html                # HTML template
```

## Stores

### Player Store (`lib/stores/player.ts`)

```typescript
import { currentTrack, isPlaying, volume, queue } from '$lib/stores/player';
```

### Theme Store (`lib/stores/theme.ts`)

```typescript
import { theme, accentColor, setTheme, setAccentColor } from '$lib/stores/theme';

// Theme: 'dark' | 'light' | 'system'
setTheme('dark');

// Accent: 'green' | 'blue' | 'purple' | 'pink' | 'orange' | 'red'
setAccentColor('green');
```

## Building for Production

```bash
npm run build
```

The build output is in the `build/` directory.

## Docker

```bash
docker build -t harmony-frontend .
docker run -p 3000:3000 harmony-frontend
```
