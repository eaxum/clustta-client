# DCC Bridge API v1

The DCC Bridge listens on `127.0.0.1:1173`. Except for `GET /health`, requests
must send the bearer token stored in Clustta's `.bridge-token` application-data
file.

All project and asset routes use stable IDs. DCC clients must not pass or retain
the project database URI.

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
- `GET /v1/projects/{projectId}/assets/{assetId}/checkpoints`

Workspace returns statuses and directly assigned assets in one response and one
project database transaction. Asset extension filters are optional and
case-insensitive.

Project metadata is cached by active account and studio. Account and studio
switches invalidate the appropriate cache automatically.

## Operations

- `POST /v1/projects/{projectId}/assets/{assetId}/status`
- `POST /v1/projects/{projectId}/assets/{assetId}/checkpoints`
- `POST /v1/projects/{projectId}/assets/{assetId}/build`
- `POST /v1/projects/{projectId}/assets/{assetId}/revert`

Mutating requests should include a unique `Idempotency-Key` header. Operations
return `202 Accepted` with a job object.

Checkpoint requests accept:

```json
{
  "filePath": "C:/project/asset.blend",
  "message": "Checkpoint comment",
  "previewPath": "",
  "useAsThumbnail": false
}
```

The bridge verifies that `filePath` is the tracked asset path. Integration
publishing and comment forwarding happen in the bridge.

## Jobs

- `GET /v1/jobs/{jobId}`
- `POST /v1/jobs/{jobId}/cancel`

Job states are `queued`, `running`, `cancelling`, `cancelled`, `succeeded`, and
`failed`. Checkpoint and status jobs are not cancellable after submission.
