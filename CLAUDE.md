# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LocalSpace is a Wails v2 desktop application for local-first personal file management and AI-assisted organization. It is built with a Go backend (Wails runtime) and a Vue 3 + TypeScript frontend. Users import local files (images, documents, videos, archives, installers, etc.) into a managed library, then search, open, and organize them by tags, descriptions, keywords, and collections. Optional AI configuration can generate tags and descriptions during import.

## Common Development Commands

The project uses Wails v2. Frontend tooling is npm-based; backend tooling is Go modules.

### Frontend (inside `frontend/`)

- `npm run dev` — start the Vite dev server on port 5173 (`strictPort: true`). Useful for standalone frontend work when Wails is not running.
- `npm run build` — build the production frontend bundle into `frontend/dist`.
- `npm run preview` — preview the built frontend bundle.
- `npm install` — install frontend dependencies.

There are no configured frontend test scripts in `frontend/package.json`.

### Backend (Go)

- `go test ./...` — run all Go tests.
- `go test ./app/services` — run tests for a specific package (e.g. `services`).
- `go test -run TestName ./app/...` — run a single test by name.
- `go build .` — compile the Go binary without Wails (does not embed the frontend).
- `go mod tidy` — clean up Go module dependencies.

### Full Application (Wails)

- `wails dev` — start the Wails development environment: builds the frontend, runs the Go backend, opens the desktop window, and enables live reload.
- `wails build` — build a production desktop binary for the current platform.
- `wails build -platform windows` — build a Windows binary.
- `wails build -platform windows/amd64` — build a Windows amd64 binary.

Wails embeds `frontend/dist` via the `//go:embed all:frontend/dist` directive in `main.go`, so the frontend must be built before `wails build` or `go build .` will succeed with the embedded assets.

## High-level Architecture

### Wails Layout

- `main.go` — entry point. Creates `app.App`, configures Wails options (window, asset server, bindings, macOS chrome, lifecycle hooks), and calls `wails.Run`.
- `app/app.go` — the Wails-bound application struct. All public methods on `App` are exposed to the frontend. `Startup` initializes the database and services; `Shutdown` and `BeforeClose` handle cleanup and prevent closing while background jobs are running.
- `frontend/` — Vue 3 SPA. Built with Vite, outputs to `frontend/dist`, which is embedded by the Go binary.

### Backend Layers

Backend code is organized into layers under `app/`:

- `app/models/` — domain models and shared constants (`File`, `Config`, `Job`, `AIConfig`, `ThemeConfig`, `StorageLayoutConfig`, `OpenWithConfig`, etc.).
- `app/database/` — SQLite setup, migrations, connection pool configuration, and transaction helpers. `database.NewSQLiteDB` opens the database and runs migrations.
- `app/repositories/` — data access layer. `SQLiteDBWrapper` exposes the raw `*sql.DB` to repositories. Repositories include `FileRepository`, `ConfigRepository`, and `JobRepository`.
- `app/services/` — business logic layer. Services include `FileService`, `ConfigService`, `StorageService`, `AIService`, `AgentService`, `ThumbnailService`, and `JobService`. `App.initializeApp` wires them together.
- `app/utils/` — utility helpers for checksums, metadata extraction, thumbnails, image/video/document parsing.
- `app/agents/` — metadata-generation agent runtime. `EinoMetadataAgent` builds prompts and executes them through `DefaultAgentRuntime`.
- `app/tools/` — runtime tools available to agents. `WebSearchTool` performs bounded HTTP search during metadata analysis when enabled.

### Backend Service Flow

1. `App.Startup` creates the `data/` directory next to the executable and opens `data/localspace.db`.
2. `database.NewSQLiteDB` runs migrations and seeds initial file types.
3. Repositories are created from the database connection.
4. Services are initialized from repositories: `StorageService`, `AIService`, `AgentService`, `ThumbnailService`, `FileService`, `ConfigService`, and `JobService`.
5. `JobService` registers handlers (e.g. `BatchImportHandler`) and normalizes any unfinished jobs from a previous run.
6. All public methods on `App` are bound to Wails and callable from the frontend as `window.go.app.App.MethodName`.

### Job System

Batch import is implemented as a background job system:

- `JobService` manages job lifecycle: submit, resume, cancel, timeout, and shutdown recovery.
- `JobPolicy` controls per-job-type concurrency, exclusivity, background execution, and recoverability.
- `BatchImportHandler` implements the `JobHandler` interface and executes a multi-file import plan with temp-file staging and resumable progress.
- Jobs are persisted in the database. Running jobs block application shutdown via `App.BeforeClose`, which prompts the user to continue or exit.
- `JobService` emits Wails runtime events through the emitter set by `App`, so the frontend can observe progress.

### AI / Metadata Generation

Metadata generation has two paths, with the agent path preferred when enabled:

1. Legacy path: `AIService` calls the OpenAI-compatible model directly via Eino to generate tags and descriptions.
2. Agent path: `AgentService` uses `EinoMetadataAgent` with `DefaultAgentRuntime`. When the user enables the agent and web search, the runtime may call the `web_search` tool to gather snippets, then ask the model again to produce tags and description. Failures fall back to an empty result so file import is not blocked.

Agent configuration and trace fields (`ToolsAvailable`, `ToolsUsed`, `SearchQueries`, `FallbackReason`) are documented in `docs/agent-integration.md`.

### Frontend Architecture

- `frontend/src/main.ts` — bootstraps the Vue app with Pinia and Vue Router.
- `frontend/src/router/index.ts` — hash-based routing (`createWebHashHistory`) with routes for Files (`/files`), Collections (`/collections`), Import (`/import`), Tasks (`/tasks`), Settings (`/settings`), and Documentation (`/documentation`).
- `frontend/src/store/` — Pinia stores. `files.ts`, `jobs.ts`, and `theme.ts` manage state for their domains.
- `frontend/src/api/index.ts` — wraps all Wails backend calls in a typed `api` object. Provides safe fallbacks when `window.go.app.App` is not available (e.g. browser-only development).
- `frontend/src/views/` — top-level route views.
- `frontend/src/components/` — reusable UI components.
- `frontend/vite.config.ts` — Vite config with the `@ -> src` alias, dev port 5173, and build output `../frontend/dist`.

### Frontend / Backend Communication

The frontend calls the backend through Wails runtime bindings. `api/index.ts` is the single source of truth for these calls. The backend can push events to the frontend via `runtime.EventsEmit` and the `JobService` emitter.

### Storage and Configuration

- Imported files are stored under configured storage directories. Master directories (`StorageDir` with `ParentID`/`IsDefault`) can contain file-type subdirectories. The storage layout is controlled by `StorageLayoutConfig`.
- Thumbnails are cached in `data/thumbnails` next to the executable.
- Configuration (storage paths, AI settings, theme, open-with mappings, storage layout) is stored in SQLite via `ConfigRepository`.
- The database and thumbnails are created next to the executable in the `data/` directory, so the app expects to be run from a directory where it can write.

### Important Notes from the README

- The app is local-first: files, database, and thumbnails live on the local machine.
- First-time setup should start with Settings → Storage Directory configuration.
- AI-assisted metadata generation is optional; configure it in Settings → AI Config.
- The main workflows are: import files, add metadata, search/open files, and manage background jobs in Tasks.

### Dependencies to Be Aware Of

- Go: Wails v2 (`github.com/wailsapp/wails/v2`), Eino (`github.com/cloudwego/eino`), OpenAI-compatible Eino component (`github.com/cloudwego/eino-ext/components/model/openai`), SQLite driver (`github.com/glebarez/sqlite`), GORM, Testify.
- Frontend: Vue 3, Vue Router, Pinia, Vite, TypeScript.

### Build Artifacts

- `frontend/dist` — built frontend bundle, embedded by the Go binary. Do not edit directly; run `npm run build` in `frontend/` to regenerate.
- `build/bin/` — default Wails output directory for the built executable.
- `build/windows/` — Windows manifest and icon assets.
