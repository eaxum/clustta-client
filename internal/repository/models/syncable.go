package models

// Syncable is implemented by every project-data model that participates in
// sync. The methods expose the two fields the generic merge helpers need:
// the row's id and its modification timestamp. Implementations are pure
// field accessors - no logic. Keeping them next to their structs (rather
// than centralised in the sync package) makes drift impossible when a
// struct gains, loses, or renames its Id/MTime field.
type Syncable interface {
	SyncId() string
	SyncMTime() int
}

func (a Asset) SyncId() string                        { return a.Id }
func (a Asset) SyncMTime() int                        { return a.MTime }
func (a AssetDependency) SyncId() string              { return a.Id }
func (a AssetCheckpointTag) SyncId() string           { return a.Id }
func (a AssetCheckpointTag) SyncMTime() int           { return int(a.MTime) }
func (a AssetDependency) SyncMTime() int              { return a.MTime }
func (a AssetTag) SyncId() string                     { return a.Id }
func (a AssetTag) SyncMTime() int                     { return a.MTime }
func (a AssetType) SyncId() string                    { return a.Id }
func (a AssetType) SyncMTime() int                    { return a.MTime }
func (c Checkpoint) SyncId() string                   { return c.Id }
func (c Checkpoint) SyncMTime() int                   { return c.MTime }
func (c Collection) SyncId() string                   { return c.Id }
func (c Collection) SyncMTime() int                   { return c.MTime }
func (c CollectionAssignee) SyncId() string           { return c.Id }
func (c CollectionAssignee) SyncMTime() int           { return c.MTime }
func (c CollectionDependency) SyncId() string         { return c.Id }
func (c CollectionDependency) SyncMTime() int         { return c.MTime }
func (c CollectionType) SyncId() string               { return c.Id }
func (c CollectionType) SyncMTime() int               { return c.MTime }
func (d DependencyType) SyncId() string               { return d.Id }
func (d DependencyType) SyncMTime() int               { return d.MTime }
func (i IntegrationAssetMapping) SyncId() string      { return i.Id }
func (i IntegrationAssetMapping) SyncMTime() int      { return i.MTime }
func (i IntegrationCollectionMapping) SyncId() string { return i.Id }
func (i IntegrationCollectionMapping) SyncMTime() int { return i.MTime }
func (i IntegrationProject) SyncId() string           { return i.Id }
func (i IntegrationProject) SyncMTime() int           { return i.MTime }
func (r Role) SyncId() string                         { return r.Id }
func (r Role) SyncMTime() int                         { return r.MTime }
func (s Status) SyncId() string                       { return s.Id }
func (s Status) SyncMTime() int                       { return s.MTime }
func (t Tag) SyncId() string                          { return t.Id }
func (t Tag) SyncMTime() int                          { return t.MTime }
func (t Template) SyncId() string                     { return t.Id }
func (t Template) SyncMTime() int                     { return t.MTime }
func (u User) SyncId() string                         { return u.Id }
func (u User) SyncMTime() int                         { return u.MTime }
func (w Workflow) SyncId() string                     { return w.Id }
func (w Workflow) SyncMTime() int                     { return w.MTime }
func (w WorkflowAsset) SyncId() string                { return w.Id }
func (w WorkflowAsset) SyncMTime() int                { return w.MTime }
func (w WorkflowCollection) SyncId() string           { return w.Id }
func (w WorkflowCollection) SyncMTime() int           { return w.MTime }
func (w WorkflowLink) SyncId() string                 { return w.Id }
func (w WorkflowLink) SyncMTime() int                 { return w.MTime }
