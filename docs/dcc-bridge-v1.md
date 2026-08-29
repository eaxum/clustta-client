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
- `GET /v1/projects/{projectId}/assets/{assetId}/checkpoint-group-tags`
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
- `POST /v1/projects/{projectId}/checkpoint-groups/{groupId}/tags/{tagId}`
- `DELETE /v1/projects/{projectId}/checkpoint-group-tags/{tagId}`
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
`checkpoint_group_tag_id`. The unused selector field must be omitted or empty.
The selector update endpoint accepts the same three selector fields without the
dependency IDs.

Dependency options return the active checkpoints and compatible checkpoint
group tags for the dependency asset. Tag creation and movement require a
finalized multi-asset checkpoint group and `manage_dependencies` permission for
the affected assets. Referenced tags cannot be deleted.

Build plans resolve the complete graph to dependency-first exact checkpoint
entries. Each plan reports conflicts, missing chunks, locally modified files,
and a fingerprint. A plan with conflicts cannot be executed.

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
