# Open Experience Phase 1 Implementation Plan

> **For agentic workers:** Steps use checkbox (`- [ ]`) syntax for tracking. Execute tasks in order because the UI and service layers depend on the new config contracts.

**Goal:** Implement Phase 1 of the open-experience redesign so LocalSpace can open files with user-preferred apps instead of relying only on the system default program.

**Architecture:** Keep the existing Wails service boundary intact, store open-with rules in the existing `configs` table as a JSON object, resolve open behavior in the backend, and expose a focused settings UI for file-type defaults plus extension overrides.

**Tech Stack:** Go, Wails, SQLite, Vue 3, Pinia-style frontend structure, Go testing package.

## Global Constraints

- Do not break the current `OpenFileLocation` behavior.
- Preserve a system-default fallback path when no preferred app is configured.
- Do not require migration of existing files or directories in Phase 1.
- Do not implement complex launch arguments in Phase 1; only support `appPath + filePath`.
- Keep Windows path handling safe for spaces and Chinese characters.
- Prefer adding dedicated config methods over spreading JSON parsing across unrelated services.
- Ensure invalid preferred-app configuration does not block file access; always offer system-default fallback.

---

## File Map

### Create
- `frontend/src/components/OpenWithConfig.vue` — settings UI for preferred app configuration.

### Modify
- `app/models/config.go` — add open-with config model types.
- `app/repositories/config_repository.go` — load/save open-with config via `configs`.
- `app/repositories/config_repository_test.go` — config JSON round-trip tests.
- `app/services/config_service.go` — expose open-with config methods.
- `app/app.go` — add Wails-facing methods for open-with config and preferred open behavior.
- `frontend/src/api/index.ts` — add API wrappers.
- `frontend/src/types/index.ts` — add TS types for open-with config.
- `frontend/src/views/SettingsView.vue` — add the new settings section.
- `frontend/src/components/FileCard.vue` — add “preferred open / system default open” menu actions.
- `docs/2026-07-19-localspace-打开体验与存储结构重构设计.md` — optionally annotate execution status after implementation.

### Transitional / Do Not Expand
- `app/services/storage_service.go` — do not change storage layout in Phase 1.
- `app/services/file_service.go` — only touch if needed for open behavior reuse; do not mix in storage-rule work.

---

## Task 1: Add backend config models for preferred open rules

**Files:**
- Modify: `app/models/config.go`
- Modify: `app/models/config_test.go` if needed

**Interfaces:**
- Produces:
  - `type OpenWithConfig struct { ByFileType map[string]string; ByExtension map[string]string }`
  - optional helper types if needed for future validation

- [ ] **Step 1: Add the failing model test**

Add a small test that asserts the struct supports both file-type and extension mappings.

```go
func TestOpenWithConfigSupportsTypeAndExtensionMappings(t *testing.T) {
	config := OpenWithConfig{
		ByFileType: map[string]string{
			"video": "C:\\Program Files\\DAUM\\PotPlayer\\PotPlayerMini64.exe",
		},
		ByExtension: map[string]string{
			".mkv": "C:\\Program Files\\DAUM\\PotPlayer\\PotPlayerMini64.exe",
		},
	}

	if config.ByFileType["video"] == "" {
		t.Fatal("expected video mapping to be present")
	}
	if config.ByExtension[".mkv"] == "" {
		t.Fatal("expected .mkv mapping to be present")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./app/models -run TestOpenWithConfigSupportsTypeAndExtensionMappings -count=1`

Expected: FAIL because `OpenWithConfig` does not exist yet.

- [ ] **Step 3: Add the minimal model**

In `app/models/config.go`, add:

```go
type OpenWithConfig struct {
	ByFileType  map[string]string `json:"byFileType"`
	ByExtension map[string]string `json:"byExtension"`
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./app/models -run TestOpenWithConfigSupportsTypeAndExtensionMappings -count=1`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add app/models/config.go app/models/config_test.go
git commit -m "feat: add open-with config model"
```

---

## Task 2: Persist open-with config in the repository layer

**Files:**
- Modify: `app/repositories/config_repository.go`
- Modify: `app/repositories/config_repository_test.go`

**Interfaces:**
- Produces:
  - `func (r *ConfigRepository) GetOpenWithConfig() (*models.OpenWithConfig, error)`
  - `func (r *ConfigRepository) SetOpenWithConfig(config *models.OpenWithConfig) error`

- [ ] **Step 1: Add the failing repository tests**

Add tests for default behavior and JSON round-trip.

```go
func TestGetOpenWithConfigReturnsDefaultWhenMissing(t *testing.T) {
	repo := newTestConfigRepository(t)

	config, err := repo.GetOpenWithConfig()
	if err != nil {
		t.Fatalf("GetOpenWithConfig returned error: %v", err)
	}
	if config.ByFileType == nil || config.ByExtension == nil {
		t.Fatal("expected default config maps to be initialized")
	}
}

func TestSetOpenWithConfigRoundTrips(t *testing.T) {
	repo := newTestConfigRepository(t)

	expected := &models.OpenWithConfig{
		ByFileType: map[string]string{"video": "C:\\Tools\\PotPlayer.exe"},
		ByExtension: map[string]string{".mkv": "C:\\Tools\\PotPlayer.exe"},
	}

	if err := repo.SetOpenWithConfig(expected); err != nil {
		t.Fatalf("SetOpenWithConfig returned error: %v", err)
	}

	actual, err := repo.GetOpenWithConfig()
	if err != nil {
		t.Fatalf("GetOpenWithConfig returned error: %v", err)
	}
	if actual.ByFileType["video"] != expected.ByFileType["video"] {
		t.Fatalf("expected video mapping %q, got %q", expected.ByFileType["video"], actual.ByFileType["video"])
	}
	if actual.ByExtension[".mkv"] != expected.ByExtension[".mkv"] {
		t.Fatalf("expected .mkv mapping %q, got %q", expected.ByExtension[".mkv"], actual.ByExtension[".mkv"])
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./app/repositories -run "Test(GetOpenWithConfigReturnsDefaultWhenMissing|SetOpenWithConfigRoundTrips)" -count=1`

Expected: FAIL because repository methods do not exist.

- [ ] **Step 3: Implement repository JSON persistence**

Use a new config key:

- `open_with_config`

Implementation guidelines:

- Missing config should return an initialized empty config:

```go
&models.OpenWithConfig{
	ByFileType:  map[string]string{},
	ByExtension: map[string]string{},
}
```

- Save as JSON string via the existing `configs` table and `Set`.
- Read via `Get`, then `json.Unmarshal`.

- [ ] **Step 4: Run the repository tests to verify they pass**

Run: `go test ./app/repositories -run "Test(GetOpenWithConfigReturnsDefaultWhenMissing|SetOpenWithConfigRoundTrips)" -count=1`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add app/repositories/config_repository.go app/repositories/config_repository_test.go
git commit -m "feat: persist open-with config"
```

---

## Task 3: Expose open-with config through ConfigService and Wails app methods

**Files:**
- Modify: `app/services/config_service.go`
- Modify: `app/app.go`

**Interfaces:**
- Produces:
  - `func (s *ConfigService) GetOpenWithConfig() (*models.OpenWithConfig, error)`
  - `func (s *ConfigService) UpdateOpenWithConfig(config models.OpenWithConfig) error`
  - `func (a *App) GetOpenWithConfig() (*models.OpenWithConfig, error)`
  - `func (a *App) UpdateOpenWithConfig(config models.OpenWithConfig) error`
  - `func (a *App) SelectExecutable() (string, error)`

- [ ] **Step 1: Add the failing app-level tests if app tests already cover exported methods**

If current app tests are lightweight, add at least service-level compile coverage. If not practical, use direct implementation plus broader integration tests in later tasks.

- [ ] **Step 2: Implement ConfigService passthrough methods**

Add thin wrappers around repository methods.

- [ ] **Step 3: Add Wails-facing app methods**

In `app/app.go`:

- expose config get/update methods
- add `SelectExecutable()` using an open-file dialog filtered to `.exe` on Windows

Suggested filter:

```go
runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
	Title: "选择打开软件",
	Filters: []runtime.FileFilter{
		{DisplayName: "可执行文件", Pattern: "*.exe"},
		{DisplayName: "所有文件", Pattern: "*.*"},
	},
})
```

- [ ] **Step 4: Verify compile/test coverage**

Run: `go test ./app/... -count=1`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add app/services/config_service.go app/app.go
git commit -m "feat: expose open-with config methods"
```

---

## Task 4: Implement preferred-open resolution and safe fallback behavior

**Files:**
- Modify: `app/app.go`
- Optional small helper extracted in `app/app.go` or nearby if it improves readability

**Interfaces:**
- Produces:
  - `func (a *App) OpenFileWithPreferredApp(id uint) error`
  - `func (a *App) OpenFileWithSystemDefault(id uint) error`
  - helper resolution logic for extension-first, type-second, system-default fallback

- [ ] **Step 1: Add failing tests for resolution order in a focused helper**

Because `exec.Command` is hard to test directly, add a pure helper such as:

```go
func resolvePreferredApp(fileName, fileType string, config *models.OpenWithConfig) string
```

Test:

```go
func TestResolvePreferredAppPrefersExtensionOverType(t *testing.T) {
	config := &models.OpenWithConfig{
		ByFileType: map[string]string{"video": "C:\\Tools\\VLC.exe"},
		ByExtension: map[string]string{".mkv": "C:\\Tools\\PotPlayer.exe"},
	}

	appPath := resolvePreferredApp("episode01.mkv", "video", config)
	if appPath != "C:\\Tools\\PotPlayer.exe" {
		t.Fatalf("expected extension mapping to win, got %q", appPath)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./app -run TestResolvePreferredAppPrefersExtensionOverType -count=1`

Expected: FAIL because helper does not exist.

- [ ] **Step 3: Implement helper and new open methods**

Behavior:

1. Load file record by ID
2. Load open-with config
3. Resolve preferred app by:
   - normalized extension such as `.mkv`
   - then `file.FileType`
4. If no preferred app is found, fall back to system-default open
5. If preferred app path is configured but missing on disk:
   - return a descriptive error, or
   - optionally fall back to system default with a warning path in later UI handling

For Phase 1, prefer this behavior:

- `OpenFileWithPreferredApp`:
  - if configured app exists, use it
  - if configured app is invalid, return explicit error
  - if not configured, use system default

- `OpenFileWithSystemDefault`:
  - always use existing `openFileWithDefaultApp`

Windows execution:

```go
cmd := exec.Command(appPath, filePath)
```

- [ ] **Step 4: Run targeted tests**

Run: `go test ./app -run TestResolvePreferredAppPrefersExtensionOverType -count=1`

Expected: PASS

- [ ] **Step 5: Run broader Go tests**

Run: `go test ./app/... -count=1`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add app/app.go app/app_test.go
git commit -m "feat: add preferred app open behavior"
```

---

## Task 5: Add frontend API and type definitions

**Files:**
- Modify: `frontend/src/api/index.ts`
- Modify: `frontend/src/types/index.ts`

**Interfaces:**
- Produces TS types:
  - `OpenWithConfig`
- Produces API wrappers:
  - `api.openWith.getConfig()`
  - `api.openWith.updateConfig(config)`
  - `api.system.selectExecutable()`
  - `api.file.openPreferred(id)`
  - `api.file.openSystemDefault(id)`

- [ ] **Step 1: Add type definitions**

Suggested type:

```ts
export interface OpenWithConfig {
  byFileType: Record<string, string>
  byExtension: Record<string, string>
}
```

- [ ] **Step 2: Extend Wails declarations and API wrapper**

Add declarations for:

- `GetOpenWithConfig`
- `UpdateOpenWithConfig`
- `SelectExecutable`
- `OpenFileWithPreferredApp`
- `OpenFileWithSystemDefault`

Use direct reject/throw style for write actions and explicit open actions where failures should surface to the UI.

- [ ] **Step 3: Run frontend type/build validation**

Run: `npm.cmd run build`

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add frontend/src/api/index.ts frontend/src/types/index.ts
git commit -m "feat: add frontend open-with api types"
```

---

## Task 6: Build the OpenWith settings component

**Files:**
- Create: `frontend/src/components/OpenWithConfig.vue`

**Interfaces:**
- Consumes:
  - `api.openWith.getConfig`
  - `api.openWith.updateConfig`
  - `api.system.selectExecutable`
- Produces:
  - a settings form for file-type defaults
  - a settings form for extension overrides
  - save/reset flow

- [ ] **Step 1: Component behavior**

The component should support:

1. Load config on mount
2. Edit file-type defaults for at least:
   - `video`
   - `music`
   - `image`
   - `document`
   - `archive`
   - `installer`
   - `other`
3. Add/remove extension-level overrides
4. Browse for executable path
5. Save config
6. Clear a mapping to fall back to system default

- [ ] **Step 2: UX rules**

- Show a short description under the section:
  - “优先按扩展名匹配，其次按文件类型匹配；未配置时使用系统默认打开方式。”
- Extension input must normalize to lowercase and ensure it starts with `.`
- Empty values should be stripped before save

- [ ] **Step 3: Suggested internal data shape**

Use:

```ts
const config = ref<OpenWithConfig>({
  byFileType: {},
  byExtension: {},
})
```

For extension overrides, consider using a local editable array:

```ts
type ExtensionRule = { extension: string; appPath: string }
```

Convert between array form and object form on load/save.

- [ ] **Step 4: Run frontend build validation**

Run: `npm.cmd run build`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/OpenWithConfig.vue
git commit -m "feat: add open-with settings component"
```

---

## Task 7: Integrate the settings component into SettingsView

**Files:**
- Modify: `frontend/src/views/SettingsView.vue`

**Interfaces:**
- Consumes: `OpenWithConfig.vue`
- Produces: a new settings card/section named `打开方式`

- [ ] **Step 1: Add the new section**

Suggested section placement:

- after `AI 配置`
- before `主题设置`

Suggested card title:

- `打开方式`

Suggested subtitle/description:

- `为不同类型文件指定 LocalSpace 内部优先使用的打开软件。`

- [ ] **Step 2: Preserve current layout quality**

The section should visually match existing settings cards and not collapse the current grid on desktop/mobile.

- [ ] **Step 3: Run frontend build validation**

Run: `npm.cmd run build`

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/SettingsView.vue
git commit -m "feat: add open-with section to settings"
```

---

## Task 8: Update file-card actions to support preferred open and system open

**Files:**
- Modify: `frontend/src/components/FileCard.vue`

**Interfaces:**
- Consumes:
  - `api.file.openPreferred`
  - `api.file.openSystemDefault`
- Produces:
  - new menu items
  - default open behavior aligned with preferred-open design

- [ ] **Step 1: Adjust file open behavior**

Change primary open behavior:

- card click / open action should call preferred open

Keep fallback behavior in backend, not duplicated in the component.

- [ ] **Step 2: Update menu items**

Add:

- `用默认软件打开`
- `用系统默认打开`

Recommended ordering:

1. 文件详情
2. 用默认软件打开
3. 用系统默认打开
4. 打开文件夹
5. 重命名
6. 编辑标签和描述
7. 删除文件

- [ ] **Step 3: Error handling**

If preferred open fails because configured app path is invalid:

- show explicit alert:
  - `指定打开软件不可用，请到设置中检查路径，或使用系统默认打开。`

- [ ] **Step 4: Run frontend build validation**

Run: `npm.cmd run build`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/FileCard.vue
git commit -m "feat: support preferred and system open actions"
```

---

## Task 9: End-to-end verification and documentation touch-up

**Files:**
- Optional Modify: `docs/2026-07-19-localspace-打开体验与存储结构重构设计.md`
- Optional Modify: this plan file to mark execution notes

- [ ] **Step 1: Backend regression**

Run:

```bash
go test ./app/... -count=1
```

Expected: PASS

- [ ] **Step 2: Frontend regression**

Run:

```bash
npm.cmd run build
```

Expected: PASS

- [ ] **Step 3: Manual verification checklist**

Verify these flows manually:

1. 未配置任何规则时，点击文件仍能按系统默认打开
2. 配置 `video -> VLC.exe` 后，视频文件从 LocalSpace 内直接走 VLC
3. 配置 `.mkv -> PotPlayer.exe` 后，`.mkv` 优先走 PotPlayer，而其他视频仍走类型规则
4. 配置无效路径时，用户会收到明确错误提示
5. “用系统默认打开” 始终可用
6. “打开文件夹” 行为未回归

- [ ] **Step 4: Add execution note**

Once complete, add a short note near the top of this plan:

```md
> Execution note: Implemented on 2026-07-19. Backend tests and frontend build passed.
```

- [ ] **Step 5: Commit**

```bash
git add docs/2026-07-19-localspace-打开体验与存储结构重构设计.md docs/superpowers/plans/2026-07-19-open-experience-phase1-plan.md
git commit -m "docs: finalize preferred open phase 1 plan"
```

---

## Self-Review

### Spec coverage

- Preferred app configuration model is covered by Tasks 1-3.
- Preferred-open resolution and system-default fallback are covered by Task 4.
- Frontend API and type wiring are covered by Task 5.
- Settings UI is covered by Tasks 6-7.
- File-card open flow changes are covered by Task 8.
- Regression checks and manual verification are covered by Task 9.

### Scope control

- No storage layout migration is included.
- No collection-name changes are included.
- No launch-argument templating is included.
- No directory watcher work is included.

### Execution order

- Backend config contract first
- Backend open behavior second
- Frontend types and APIs third
- Settings UI fourth
- File action integration fifth

