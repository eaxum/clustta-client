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
- `GET /v1/projects/{projectId}/assets/{assetId}/dependency-options/{dependencyId}`
- `GET /v1/projects/{projectId}/assets/{assetId}/checkpoint-tags`
- `GET /v1/projects/{projectId}/assets/{assetId}/checkpoints`
- `GET /v1/projects/{projectId}/assets/{assetId}/build-plan`

Workspace returns statuses and directly assigned assets in one response and one
project database transaction. Asset extension filters are optional and
case-insensitive.

Project metadata is cached by active account and studio. Account and studio
switches invalidate the appropriate cache automatically.

## Operations

- `POST /v1/projects/{projectId}/assets/{assetId}/status`
- `POST /v1/projects/{projectId}/assets/{assetId}/open`
- `POST /v1/projects/{projectId}/assets/{assetId}/reveal`
- `POST /v1/projects/{projectId}/assets/{assetId}/checkpoints`
- `POST /v1/projects/{projectId}/assets/{assetId}/dependencies`
- `POST /v1/projects/{projectId}/assets/{assetId}/dependencies/{edgeId}/selector`
- `POST /v1/projects/{projectId}/checkpoint-groups/{groupId}/tags`
- `POST /v1/projects/{projectId}/checkpoints/{checkpointId}/tags`
- `POST /v1/projects/{projectId}/checkpoints/{checkpointId}/tags/{tagId}`
- `DELETE /v1/projects/{projectId}/checkpoint-tags/{tagId}`
- `POST /v1/projects/{projectId}/assets/{assetId}/build`
- `POST /v1/projects/{projectId}/assets/{assetId}/revert`

Open focuses the tracked asset in Clustta Desktop. Reveal opens the system file
manager with the tracked asset selected. Both return `204 No Content`.

Checkpoint, status, build, and revert requests should include a unique
`Idempotency-Key` header. These asynchronous operations return `202 Accepted`
with a job object.

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

## Dependency selectors

Dependency responses are edge records containing `resolution_mode`, selector
references, the currently resolved checkpoint, and `resolution_status`.

Create a floating dependency with:

```json
{
  "dependency_id": "boy-asset-id",
  "dependency_type_id": "reference-type-id",
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
