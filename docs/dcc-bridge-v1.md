# DCC Bridge API v1

The DCC Bridge listens on `127.0.0.1:1173`. Except for `GET /health`, requests
must send the bearer token stored in Clustta's `.bridge-token` application-data
file.

All project and asset routes use stable IDs. DCC clients must not pass or retain
the project database URI. DCC clients should send their selected studio in the
`X-Clustta-Studio` header so project resolution is independent of the studio
currently displayed in Clustta Desktop.

## Discovery

- `GET /v1/capabilities`
- `GET /v1/bootstrap`
- `GET /v1/context?filePath=<absolute-path>`
- `GET /v1/projects`

Bootstrap returns the API version, accounts, studios, local projects, active
account, and active studio in one request. Add `?refresh=true` to refresh the
cached studio and project catalog.

## Project data

- `GET /v1/projects/{projectId}/workspace?ext=.blend`
- `GET /v1/projects/{projectId}/assets?ext=.blend`
- `GET /v1/projects/{projectId}/statuses`
- `GET /v1/projects/{projectId}/assets/{assetId}/dependencies`
- `GET /v1/projects/{projectId}/dependency-types`
- `GET /v1/projects/{projectId}/assets/{assetId}/dependency-options/{dependencyId}`
- `GET /v1/projects/{projectId}/assets/{assetId}/checkpoint-tags`
- `GET /v1/projects/{projectId}/assets/{assetId}/checkpoints`
- `GET /v1/projects/{projectId}/assets/{assetId}/build-plan`

Workspace returns statuses and directly assigned assets in one response and one
project database transaction. Asset extension filters are optional and
case-insensitive.

Project metadata is cached by active account and studio. Account and studio
switches invalidate the appropriate cache automatically.

Context resolution uses a persistent, account-specific local project index. The
index maps canonical working directories to project IDs, Studio names, and
local `.clst` database paths. Resolution selects the most specific containing
working directory without fetching Studio project catalogs or relying on the
Studio currently displayed in Clustta Desktop.

The index is loaded into memory when the Bridge starts and persisted under the
Clustta application-data `project-indexes` directory. Successful project create,
download, upload, open, working-directory update, and local removal operations
update it. Existing installations are backfilled from local Personal and shared
Studio `.clst` files without contacting remote Studios. Account switches replace
the in-memory index.

After selecting a project, the Bridge validates the indexed database metadata
and performs a direct project-relative asset lookup. Special linked or pending
paths retain a compatibility scan. `asset` is null when the file is inside the
project but is not tracked, allowing a DCC to offer an explicit tracking action.
A path outside every indexed local project returns `404 Not Found`.

Checksum-derived file status is cached in memory by normalized path, file
identity, size, nanosecond modification time, and checkpoint set. Repeated
context requests for an unchanged file do not rehash it. A changed file or
checkpoint set invalidates the cached status before the response is returned.

## Operations

- `POST /v1/projects/{projectId}/assets`
- `POST /v1/projects/{projectId}/integrations/ayon/workfiles/resolve`
- `POST /v1/projects/{projectId}/integrations/ayon/workfiles/prepare`
- `POST /v1/projects/{projectId}/integrations/ayon/workfiles/import`
- `POST /v1/projects/{projectId}/assets/{assetId}/status`
- `POST /v1/projects/{projectId}/assets/{assetId}/open`
- `POST /v1/projects/{projectId}/assets/{assetId}/reveal`
- `POST /v1/projects/{projectId}/assets/{assetId}/checkpoints`
- `POST /v1/projects/{projectId}/assets/{assetId}/dependencies`
- `POST /v1/projects/{projectId}/assets/{assetId}/dependencies/import`
- `POST /v1/projects/{projectId}/assets/{assetId}/dependencies/{edgeId}/selector`
- `POST /v1/projects/{projectId}/checkpoint-groups/{groupId}/tags`
- `POST /v1/projects/{projectId}/checkpoints/{checkpointId}/tags`
- `POST /v1/projects/{projectId}/checkpoints/{checkpointId}/tags/{tagId}`
- `DELETE /v1/projects/{projectId}/checkpoint-tags/{tagId}`
- `POST /v1/projects/{projectId}/assets/{assetId}/build`
- `POST /v1/projects/{projectId}/assets/{assetId}/sync`
- `POST /v1/projects/{projectId}/assets/{assetId}/get-latest`
- `POST /v1/projects/{projectId}/assets/{assetId}/revert`

The AYON prepare endpoint accepts AYON project, folder hierarchy, task, task
type, application identifiers, and an optional AYON-resolved workfile path. It
returns an existing mapped workfile or a validated target path for a new DCC
file. The requested path must resolve inside the Clustta project working
directory. It creates missing generic collections when permitted and does not
require a linked integration or Clustta workfile template.

```json
{
  "ayon_project_id": "ayon-project-id",
  "folder_id": "ayon-folder-id",
  "folder_path": "/assets/characters/hero/",
  "workfile_path": "C:/project/assets/characters/hero/Hero_Modeling.blend",
  "folders": [
    {
      "id": "ayon-folder-id",
      "parent_id": "",
      "name": "hero",
      "path": "/assets/characters/hero/",
      "type": "asset"
    }
  ],
  "task_id": "ayon-task-id",
  "task_name": "modeling",
  "task_type_id": "modeling-type-id",
  "task_type": "Modeling",
  "application": "blender",
  "extension": ".blend"
}
```

When `requires_import` is true, the DCC creates and saves its native file at
`target_path`. The client then submits the same context and path to the import
endpoint:

```json
{
  "context": {},
  "file_path": "C:/project/assets/characters/hero/Modeling.blend"
}
```

Import is asynchronous. It validates the canonical path and atomically creates
the asset, initial checkpoint, assignment, and AYON DCC mapping. The active role
requires `create_asset`, `create_checkpoint`, and `assign_asset` permissions.
Creating missing hierarchy also requires `create_collection`.

The older resolve endpoint remains available for clients using configured AYON
project, folder, and template mappings.

Open focuses the tracked asset in Clustta Desktop. Reveal opens the system file
manager with the tracked asset selected. Both return `204 No Content`.

### Track an existing file

Create an asset and its initial checkpoint from an existing file inside the
project working directory:

```json
{
  "filePath": "C:/project/assets/character.blend",
  "message": "v0001",
  "previewPath": "",
  "assignToCurrentUser": false,
  "sync": true
}
```

The bridge resolves and validates the canonical path, creates missing collection
folders when permitted, and creates an unassigned resource by default. The
active role requires `create_asset` and `create_checkpoint`. Set
`assignToCurrentUser` to `true` to create an assigned task for the active user;
this also requires `assign_asset`. `create_collection` is required when the
relative file path contains a collection that is not tracked yet.

Tracking is asynchronous and returns `202 Accepted` with a job. The completed
job result contains `asset`, `checkpoint`, and `created`. Repeating the request
for an existing tracked asset returns the asset and checkpoint with `created`
set to `false`. When `assignToCurrentUser` is `true`, an existing unassigned
resource is converted to a task and assigned to the active user. An asset
assigned to another user is rejected. Send an `Idempotency-Key` header so a
retried request returns the original job.

The request may also include `sourceAssetId` and `sourceCheckpointId`, using the
same source-selection rules as checkpoint creation. A blank message is stored as
`v0001`. Set `sync` to override the desktop sync-after-checkpoint preference.

Checkpoint, status, build, revert, sync, Get Latest, and dependency import
requests should include a unique `Idempotency-Key` header. These asynchronous
operations return `202 Accepted` with a job object.

Checkpoint requests accept:

```json
{
  "filePath": "C:/project/asset.blend",
  "message": "Checkpoint comment",
  "previewPath": "",
  "useAsThumbnail": false
}
```

`message` is optional. When it is omitted or blank, the bridge stores the next
asset version as the comment, starting at `v0001`. Explicit comments are
trimmed and stored as provided.

The bridge verifies that `filePath` is the tracked asset path. Integration
publishing and comment forwarding happen in the bridge.

Asset sync pushes the current asset, its checkpoints, chunks, previews, and
related metadata to the configured Studio. Get Latest first refreshes project
metadata, resolves a new dependency build plan, and restores that exact plan.
Both operations are asynchronous. Get Latest rejects locally modified files
unless `allow_modified` is true after explicit DCC confirmation.

## Dependency selectors

Dependency responses are edge records containing `resolution_mode`, selector
references, the currently resolved checkpoint, `resolution_status`, and a
`dependency` object with the referenced asset's current path and file state.

Dependency creation and import always use the project's `linked` dependency
type; callers do not supply a dependency type. The import endpoint accepts
direct DCC file references. The Bridge canonicalizes every path, rejects
references outside the project, resolves tracked assets, skips existing edges
idempotently, and returns added, existing, and unresolved references.
DCC-provided `kind` values are descriptive and are not trusted for path or asset
resolution.

Create a floating dependency with:

```json
{
  "dependency_id": "boy-asset-id",
  "resolution_mode": "floating"
}
```

For `pinned`, provide `checkpoint_id`. For `tagged`, provide
`asset_checkpoint_tag_id`. The unused selector field must be omitted or empty.
The selector update endpoint accepts the same three selector fields without the
dependency IDs.

Dependency options return the active checkpoints and checkpoint tag assignments
for the dependency asset. Tag names are project-wide, while a tag identifies at
most one checkpoint in each asset history. Applying the same tag to another
checkpoint moves that asset's assignment. Batch assignment applies the tag to
the latest checkpoint for each asset sharing the operation `group_id`. Tag
changes require `manage_dependencies` permission for the affected assets, and
referenced assignments cannot be deleted.

Build plans resolve the complete graph to dependency-first exact checkpoint
entries. Each plan reports conflicts, missing chunks, locally modified files,
and a fingerprint. A plan with conflicts cannot be executed.

The desktop graph uses the separate Wails method
`AssetService.ResolveDependencyGraphPlan(projectPath, assetId)`. Its response
contains `entries` (`asset_id`, `checkpoint_id`, optional `dependency_edge_id`)
and `conflicts` using the build conflict format. It resolves the complete graph
without reading local asset files, checking chunks, or loading checkpoint
previews. It requires checkpoint read permission and access to the root asset,
but does not require download permission. Build readiness fields and execution
fingerprints are not included; the bridge build-plan and execution APIs retain
their readiness checks and existing contracts.

The direct desktop view shows only the root asset's outgoing dependencies and
selectors, without a transitive conflict scan. Full graph resolves all reachable
assets and collection contents without a depth limit, retaining direct links as
well as nested relationships. `GetRecursiveDependencies` accepts depth `1` for
direct dependencies and `0` for the full graph; returned items include `parentId`
and `parentIds` so shared collection relationships can all be drawn.
Requests run in parallel where possible. The graph is drawn once its requested
selectors and, in full mode, conflicts are resolved. Failed requests preserve
the previous graph; responses from a previous selection or closed graph are ignored.

Build requests accept:

```json
{
  "plan_fingerprint": "<fingerprint returned by build-plan>",
  "allow_modified": false
}
```

The bridge resolves the graph again before writing files and rejects stale
fingerprints. Set `allow_modified` only after the DCC has explicitly confirmed
that locally modified dependency files may be overwritten. Execution downloads
missing chunks and restores the exact checkpoint IDs in dependency-first order.

## Jobs

- `GET /v1/jobs/{jobId}`
- `POST /v1/jobs/{jobId}/cancel`

Job states are `queued`, `running`, `cancelling`, `cancelled`, `succeeded`, and
`failed`. Checkpoint and status jobs are not cancellable after submission.

## Desktop transfer coordination

Desktop checkpoint downloads and asset/collection fetches now have independent Activity IDs and cancellation. Existing Bridge job contracts are unchanged. Bridge build/revert and checkpoint/sync service calls use project admission and return a busy error when they conflict with active desktop transfers. See [Activity transfers](activity-transfers.md) for scope and verification.

### Checkpoint sources

Checkpoint creation accepts optional `sourceAssetId` and `sourceCheckpointId`.
With only `sourceAssetId`, the bridge resolves the latest local checkpoint once
before starting the creation job. An explicit checkpoint must belong to the
specified source asset. Responses include nullable `source_checkpoint_id`.
The reference is fixed and never follows later checkpoints automatically.

Project schema 2.2 adds one nullable source column. Source selections are
validated by the service; database triggers reject cycles and protect referenced
checkpoints and assets from deletion/trash. The column is a logical reference
rather than a foreign key because permission-filtered sync can omit source
metadata. An unavailable source keeps its ID and is displayed as unavailable.

Desktop checkpoint editing changes comment, source, and optional checkpoint tags
atomically. Comment/source edits require checkpoint creation permission on the
asset; tag changes also require `manage_dependencies`. A null tag list leaves
tags unchanged; an empty list removes them. No audit history is recorded.

Checkpoint metadata pushes require remote project schema 2.2. Upgrade the studio
server alongside the client; older servers can ignore edits to existing
checkpoints. Pulling from an older server is blocked when local source references
would otherwise be lost.
