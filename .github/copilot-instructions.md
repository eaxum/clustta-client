# Clustta Client - AI Coding Instructions

## Project Overview
Clustta is a distributed version control and collaboration system for creative workflows (like GitHub for creative work). This is the **desktop client** built with **Wails v3** (Go backend + Vue 3 frontend).

## Architecture

### Technology Stack
- **Backend**: Go 1.25+ with Wails v3 framework
- **Frontend**: Vue 3 + Vite + Pinia stores
- **Database**: SQLite (per-project `.clst` files)
- **Serialization**: Protocol Buffers for efficient data transmission

### Key Directories
```
services/          # Go services exposed to frontend via Wails bindings
internal/          # Core Go business logic (not directly exposed)
  repository/      # Database models and operations (SQLite)
  auth_service/    # Authentication handling
  sync_service/    # Sync/collaboration logic
frontend/src/
  services/        # Service abstraction layer (Wails bindings or HTTP)
  stores/          # Pinia state management
  instances/       # Platform-specific UI (desktop/, web/, common/)
  lib/             # Utilities including mitt event bus
```

### Service Layer Pattern
Go services in `services/` are registered in `main.go` and auto-generate TypeScript bindings:
```go
// services/asset_services.go - Backend service
func (t *AssetService) GetAssetByID(projectPath, assetId string) (models.Task, error)

// Frontend usage via auto-generated bindings
import { AssetService } from '@/services';
await AssetService.GetAssetByID(projectUri, id);
```

Frontend services are abstracted in `frontend/src/services/index.js` to support both:
- **Desktop (Wails)**: Direct Go bindings via `adapters/wails.js`
- **Web mode**: HTTP REST calls via `adapters/http.js` (set `VITE_PLATFORM=web`)

### Event Communication
- **Go → Frontend**: `app.Event.Emit("event-name", data)` (Wails events)
- **Frontend listens**: `Events.On("event-name", callback)` from `@wailsio/runtime`
- **Frontend internal**: `emitter.emit()` via mitt (`@/lib/mitt`)

Common events: `progress-update`, `fs-change`, `sync-project`, `refresh-browser`

### External Server Integration
Clustta connects to two external servers for collaboration:
- **[clustta-server](https://github.com/eaxum/clustta-server)** - Global authentication server
- **[clustta-studio](https://github.com/eaxum/clustta-studio)** - Studio/team management server

For local development with these servers, set build flags in `build/platform/Taskfile.yml`:
```bash
-ldflags="-X clustta/internal/constants.host=http://127.0.0.1:5000"
```

## Development Commands

```bash
# Run development client (primary command)
make client           # or: wails3 dev

# Build for production
make build            # Creates MSIX (Windows) or Mac App Store build

# Install dev dependencies
make install-dev      # Installs wails3 CLI

# Run tests
go test ./...

# Protocol buffer regeneration (after editing schema.proto)
protoc --go_out=. .\internal\repository\schema.proto
pbjs -t static-module -w es6 --keep-case -o .\frontend\src\lib\repositorypb.js .\internal\repository\schema.proto
```

## Core Domain Concepts

| Term | Meaning |
|------|---------|
| **Entity/Collection** | Folder-like container for organizing assets |
| **Task/Asset** | Individual file being version-controlled |
| **Checkpoint** | A saved version/snapshot of assets |
| **Project** | A `.clst` SQLite database containing all metadata |
| **Studio** | Team/organization server for collaboration |

## Code Style Conventions

### Comments & Docstrings
- Keep docstrings/comments to **one line (max two)**. State what the function does, not how or why in detail.
- Do not over-explain: avoid restating the implementation, edge cases, or background rationale in comments.
- No inline comments except at major blocks.
- Applies to all languages (Go, JS/Vue, YAML, shell, etc.).

### Go Services
- One line comment preceding functions (two only when truly necessary)
- See `services/collection_service.go` for reference pattern:
```go
// GetCollectionCount returns the total number of collections in the project.
func (t *CollectionService) GetCollectionCount(projectPath string) (int, error) {
```

### Vue Components (`<script setup>`)
Organize sections in this order with comment headers (see `frontend/src/assets/boilerplate.js`):
```javascript
// imports
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import emitter from '@/lib/mitt';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import GridView from '@/instances/desktop/components/GridView.vue';
import PageState from '@/instances/common/components/PageState.vue';

// services
import { AssetService, CollectionService, SyncService } from '@/services';

// stores (alphabetically)
const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const iconStore = useIconStore();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

// refs (alphabetically)
const browserRoot = ref(null);
const isLoading = ref(false);
const searchQuery = ref('');

// computed properties (alphabetically, but dependencies first)
const canModifyEntity = computed(() => { ... });
const filteredAssets = computed(() => { ... });
const isTasksModified = computed(() => { ... });

// Note: If a computed depends on another computed, the dependency must come first.
// Example: filteredConflicts depends on taskConflicts, so taskConflicts must be defined before filteredConflicts.

// methods/functions (alphabetically)
const clearSearch = async () => { ... };
const createEntity = () => { ... };
const deleteMultipleItems = async () => { ... };
const handleClickOutside = (event) => { ... };
const refresh = async () => { ... };

// watchers
watch(() => projectStore.activeProject, async () => { ... });

// lifecycle hooks
onMounted(async () => { ... });
onUnmounted(() => { ... });
```

### Component Template Style
- **Inline elements**: Keep component and element tags to 2-3 lines max, don't break every prop/attribute onto new lines
- **Spacing**: One blank line between elements in `<template>` section
- **No comments**: Avoid comments in templates except for temporarily disabled elements
- **Icons**: Always use `getAppIcon('icon-name')` helper, never direct paths
```vue
<!-- Good -->
<div class="container" :class="{ 'active': isActive }" @click="handleClick">

<ActionButton :icon="getAppIcon('edit')" v-tooltip="'Rename'" :buttonFunction="startRename" />

<!-- Avoid -->
<div 
  class="container" 
  :class="{ 'active': isActive }" 
  @click="handleClick"
>
<ActionButton 
  :icon="getAppIcon('edit')" 
  v-tooltip="'Rename'" 
  :buttonFunction="startRename" 
/>
```

### Scrollbar Styling
Use consistent scrollbar styling across all scrollable containers:
```css
.scroll-container::-webkit-scrollbar {
  width: 4px;
}

.scroll-container::-webkit-scrollbar-thumb {
  border-radius: var(--small-radius);
  background-color: var(--light-steel);
}

.scroll-container::-webkit-scrollbar-track {
  border-radius: var(--small-radius);
}
```

### Refactoring Checklist
When cleaning up components:
- [ ] Remove unused imports, functions, computed properties, and CSS classes
- [ ] Remove unused variables within methods
- [ ] Alphabetize all imports, stores, refs, computed props, and methods
- [ ] For computed props: dependencies must come before dependents (then alphabetize within same level)
- [ ] Add section comment headers per boilerplate
- [ ] Ensure functions have 1-2 line preceding comments (no inline comments)
- [ ] Consolidate component and element tags to inline format
- [ ] No lines between template elements except major blocks
- [ ] Remove template comments (except temporarily disabled elements)
- [ ] Remove commented-out CSS and unused CSS classes
- [ ] Apply standard scrollbar styling to all scrollable containers

## Key Patterns

### Database Access
Projects use SQLite databases (`.clst` files). Always:
```go
dbConn, err := utils.OpenDb(projectPath)
defer dbConn.Close()
tx, err := dbConn.Beginx()  // Use transactions
```

### Pinia Store Convention
Stores use Options API style with clear state/getters/actions separation:
```javascript
export const useProjectStore = defineStore("projects", {
  state: () => ({ activeProject: null }),
  getters: { ... },
  actions: { async loadProject() { ... } }
});
```

### Component Organization
- `instances/desktop/` - Desktop-specific components
- `instances/web/` - Web-only components  
- `instances/common/` - Shared components
- Modals go in `modals/`, pages in `pages/`, reusable UI in `components/`

### Sync Conflict Handling
Conflicts during sync are stored in `useSyncConflictStore` and displayed via `SyncConflictModal.vue`. Resolution options: Rename (local) or Merge (with server).

## Testing
Tests are in `*_test/` directories alongside their packages:
- `internal/repository_test/`
- `internal/auth_service_test/`

Run with: `go test ./...`

## Platform-Specific Code
OS-specific implementations use build tags:
- `fs_service_windows.go` / `fs_service_darwin.go` / `fs_service_linux.go`
- `monitor.go` (Windows fullscreen) / `monitor_stub.go` (other platforms)

## Important Files
- `main.go` - App entry, service registration, menu setup
- `build/config.yml` - Wails build configuration
- `internal/repository/schema.sql` - Database schema
- `internal/repository/schema.proto` - Protobuf definitions
- `frontend/src/App.vue` - Root Vue component with global event listeners

## Security Audit Reference
A comprehensive security audit checklist for all three Clustta repos (server, client, studio) is maintained at `docs/SECURITY-AUDIT.md` in the client repo. It covers:

1. **Authentication & Session Management** — sessions, API tokens, OAuth, rate limiting
2. **Authorization & Access Control** — IDOR, role enforcement, frontend-only guards
3. **Input Validation & Injection** — SQL injection, path traversal, command injection, XSS
4. **File System & Storage Security** — `.clst` access, R2 presigned URLs, symlinks, temp files
5. **Network & Transport Security** — TLS, CORS, presigned URL transport
6. **Secrets & Configuration Management** — hardcoded secrets, leaked credentials, error messages
7. **Denial of Service & Resource Exhaustion** — unbounded uploads, goroutine leaks, SQLite locking
8. **Dependency & Supply Chain** — Go module CVEs, npm CVEs
9. **Desktop Client Specific (Wails v3)** — IPC boundary, local server, deep links, auto-update
10. **Data Integrity & Cryptography** — chunk hashing, checkpoint integrity
11. **Logging & Monitoring** — sensitive data in logs, audit trails

When performing security work, consult `docs/SECURITY-AUDIT.md` for the full checklist with attack scenarios and expected output format.
