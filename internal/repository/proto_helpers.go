package repository

import (
	"clustta/internal/chunk_service"
	"clustta/internal/repository/models"
	"clustta/internal/repository/repositorypb"
)

// --- Conversion helpers ---

func ToPbUsers(users []models.User) []*repositorypb.User {
	pb := make([]*repositorypb.User, len(users))
	for i, u := range users {
		pb[i] = &repositorypb.User{
			Id:        u.Id,
			Mtime:     int64(u.MTime),
			AddedAt:   u.AddedAt,
			Username:  u.Username,
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Photo:     u.Photo,
			RoleId:    u.RoleId,
			Synced:    u.Synced,
			Role:      u.Role.Name, // or other field if needed
		}
	}
	return pb
}

func ToPbCollectionTypes(collectionTypes []models.CollectionType) []*repositorypb.CollectionType {
	pb := make([]*repositorypb.CollectionType, len(collectionTypes))
	for i, et := range collectionTypes {
		pb[i] = &repositorypb.CollectionType{
			Id:     et.Id,
			Mtime:  int64(et.MTime),
			Name:   et.Name,
			Icon:   et.Icon,
			Synced: et.Synced,
		}
	}
	return pb
}

func ToPbAssetTypes(types []models.AssetType) []*repositorypb.AssetType {
	pb := make([]*repositorypb.AssetType, len(types))
	for i, t := range types {
		pb[i] = &repositorypb.AssetType{
			Id:     t.Id,
			Mtime:  int64(t.MTime),
			Name:   t.Name,
			Icon:   t.Icon,
			Synced: t.Synced,
		}
	}
	return pb
}

func ToPbAssets(assets []models.Asset) []*repositorypb.Asset {
	pb := make([]*repositorypb.Asset, len(assets))
	for i, t := range assets {
		pb[i] = &repositorypb.Asset{
			Id:          t.Id,
			Mtime:       int64(t.MTime),
			CreatedAt:   t.CreatedAt,
			Name:        t.Name,
			Description: t.Description,
			Extension:   t.Extension,
			IsResource:  t.IsResource,
			StatusId:    t.StatusId,
			AssetTypeId:  t.AssetTypeId,
			CollectionId:    t.CollectionId,
			AssigneeId:  t.AssigneeId,
			AssignerId:  t.AssignerId,
			IsLink:      t.IsLink,
			Pointer:     t.Pointer,
			PreviewId:   t.PreviewId,
			Trashed:     t.Trashed,
			Synced:      t.Synced,
		}
	}
	return pb
}

func ToPbCollections(collections []models.Collection) []*repositorypb.Collection {
	pb := make([]*repositorypb.Collection, len(collections))
	for i, e := range collections {
		pb[i] = &repositorypb.Collection{
			Id:           e.Id,
			Mtime:        int64(e.MTime),
			CreatedAt:    e.CreatedAt,
			Name:         e.Name,
			Description:  e.Description,
			CollectionPath:   e.CollectionPath,
			Trashed:      e.Trashed,
			CollectionTypeId: e.CollectionTypeId,
			ParentId:     e.ParentId,
			PreviewId:    e.PreviewId,
			Synced:       e.Synced,
			IsLibrary:    e.IsLibrary,
		}
	}
	return pb
}

func ToPbCollectionAssignees(collectionAssignees []models.CollectionAssignee) []*repositorypb.CollectionAssignee {
	pb := make([]*repositorypb.CollectionAssignee, len(collectionAssignees))
	for i, ea := range collectionAssignees {
		pb[i] = &repositorypb.CollectionAssignee{
			Id:         ea.Id,
			Mtime:      int64(ea.MTime),
			CollectionId:   ea.CollectionId,
			AssigneeId: ea.AssigneeId,
			AssignerId: ea.AssignerId,
			Synced:     ea.Synced,
		}
	}
	return pb
}

func ToPbAssetDependencies(assetDependencies []models.AssetDependency) []*repositorypb.AssetDependency {
	pb := make([]*repositorypb.AssetDependency, len(assetDependencies))
	for i, td := range assetDependencies {
		pb[i] = &repositorypb.AssetDependency{
			Id:               td.Id,
			Mtime:            int64(td.MTime),
			AssetId:           td.AssetId,
			DependencyId:     td.DependencyId,
			DependencyTypeId: td.DependencyTypeId,
			Synced:           td.Synced,
		}
	}
	return pb
}

func ToPbCollectionDependencies(collectionDependencies []models.CollectionDependency) []*repositorypb.CollectionDependency {
	pb := make([]*repositorypb.CollectionDependency, len(collectionDependencies))
	for i, ed := range collectionDependencies {
		pb[i] = &repositorypb.CollectionDependency{
			Id:               ed.Id,
			Mtime:            int64(ed.MTime),
			AssetId:           ed.AssetId,
			DependencyId:     ed.DependencyId,
			DependencyTypeId: ed.DependencyTypeId,
			Synced:           ed.Synced,
		}
	}
	return pb
}

func ToPbWorkflows(workflows []models.Workflow) []*repositorypb.Workflow {
	pb := make([]*repositorypb.Workflow, len(workflows))
	for i, w := range workflows {
		pb[i] = &repositorypb.Workflow{
			Id:     w.Id,
			Mtime:  int64(w.MTime),
			Name:   w.Name,
			Synced: w.Synced,
		}
	}
	return pb
}

func ToPbWorkflowAssets(workflowAssets []models.WorkflowAsset) []*repositorypb.WorkflowAsset {
	pb := make([]*repositorypb.WorkflowAsset, len(workflowAssets))
	for i, wt := range workflowAssets {
		pb[i] = &repositorypb.WorkflowAsset{
			Id:               wt.Id,
			Mtime:            int64(wt.MTime),
			Name:             wt.Name,
			TemplateId:       wt.TemplateId,
			IsResource:       wt.IsResource,
			WorkflowId:       wt.WorkflowId,
			AssetTypeId:       wt.AssetTypeId,
			WorkflowCollectionId: wt.WorkflowCollectionId,
			IsLink:           wt.IsLink,
			Pointer:          wt.Pointer,
			Synced:           wt.Synced,
		}
	}
	return pb
}

func ToPbWorkflowCollections(workflowCollections []models.WorkflowCollection) []*repositorypb.WorkflowCollection {
	pb := make([]*repositorypb.WorkflowCollection, len(workflowCollections))
	for i, we := range workflowCollections {
		pb[i] = &repositorypb.WorkflowCollection{
			Id:           we.Id,
			Mtime:        int64(we.MTime),
			Name:         we.Name,
			WorkflowId:   we.WorkflowId,
			CollectionTypeId: we.CollectionTypeId,
			ParentId:     we.ParentId,
			Synced:       we.Synced,
		}
	}
	return pb
}

func ToPbWorkflowLinks(workflowLinks []models.WorkflowLink) []*repositorypb.WorkflowLink {
	pb := make([]*repositorypb.WorkflowLink, len(workflowLinks))
	for i, wl := range workflowLinks {
		pb[i] = &repositorypb.WorkflowLink{
			Id:                 wl.Id,
			Mtime:              int64(wl.MTime),
			Name:               wl.Name,
			CollectionTypeId:       wl.CollectionTypeId,
			WorkflowId:         wl.WorkflowId,
			LinkedWorkflowId:   wl.LinkedWorkflowId,
			LinkedWorkflowName: wl.LinkedWorkflowName,
			Synced:             wl.Synced,
		}
	}
	return pb
}

func ToPbDependencyTypes(dependencyTypes []models.DependencyType) []*repositorypb.DependencyType {
	pb := make([]*repositorypb.DependencyType, len(dependencyTypes))
	for i, dt := range dependencyTypes {
		pb[i] = &repositorypb.DependencyType{
			Id:     dt.Id,
			Mtime:  int64(dt.MTime),
			Name:   dt.Name,
			Synced: dt.Synced,
		}
	}
	return pb
}

func ToPbStatuses(statuses []models.Status) []*repositorypb.Status {
	pb := make([]*repositorypb.Status, len(statuses))
	for i, s := range statuses {
		pb[i] = &repositorypb.Status{
			Id:        s.Id,
			Mtime:     int64(s.MTime),
			Name:      s.Name,
			ShortName: s.ShortName,
			Color:     s.Color,
			Synced:    s.Synced,
		}
	}
	return pb
}

func ToPbTags(tags []models.Tag) []*repositorypb.Tag {
	pb := make([]*repositorypb.Tag, len(tags))
	for i, t := range tags {
		pb[i] = &repositorypb.Tag{
			Id:     t.Id,
			Mtime:  int64(t.MTime),
			Name:   t.Name,
			Synced: t.Synced,
		}
	}
	return pb
}

func ToPbAssetTags(assetTags []models.AssetTag) []*repositorypb.AssetTag {
	pb := make([]*repositorypb.AssetTag, len(assetTags))
	for i, tt := range assetTags {
		pb[i] = &repositorypb.AssetTag{
			Id:     tt.Id,
			Mtime:  int64(tt.MTime),
			AssetId: tt.AssetId,
			TagId:  tt.TagId,
			Synced: tt.Synced,
		}
	}
	return pb
}

func ToPbCheckpoints(checkpoints []models.Checkpoint) []*repositorypb.Checkpoint {
	pb := make([]*repositorypb.Checkpoint, len(checkpoints))
	for i, c := range checkpoints {
		pb[i] = &repositorypb.Checkpoint{
			Id:             c.Id,
			Mtime:          int64(c.MTime),
			CreatedAt:      c.CreatedAt,
			AssetId:         c.AssetId,
			XxhashChecksum: c.XXHashChecksum,
			TimeModified:   int64(c.TimeModified),
			FileSize:       int64(c.FileSize),
			Comment:        c.Comment,
			Chunks:         c.Chunks,
			AuthorUid:      c.AuthorUID,
			PreviewId:      c.PreviewId,
			Trashed:        c.Trashed,
			Synced:         c.Synced,
			GroupId:        c.GroupId,
		}
	}
	return pb
}

func ToPbRoles(roles []models.Role) []*repositorypb.Role {
	pb := make([]*repositorypb.Role, len(roles))
	for i, r := range roles {
		pb[i] = &repositorypb.Role{
			Id:     r.Id,
			Mtime:  int64(r.MTime),
			Name:   r.Name,
			Synced: r.Synced,

			ViewCollection:   r.ViewCollection,
			CreateCollection: r.CreateCollection,
			UpdateCollection: r.UpdateCollection,
			DeleteCollection: r.DeleteCollection,

			ViewAsset:   r.ViewAsset,
			CreateAsset: r.CreateAsset,
			UpdateAsset: r.UpdateAsset,
			DeleteAsset: r.DeleteAsset,

			ViewTemplate:   r.ViewTemplate,
			CreateTemplate: r.CreateTemplate,
			UpdateTemplate: r.UpdateTemplate,
			DeleteTemplate: r.DeleteTemplate,

			ViewCheckpoint:   r.ViewCheckpoint,
			CreateCheckpoint: r.CreateCheckpoint,
			DeleteCheckpoint: r.DeleteCheckpoint,

			PullChunk: r.PullChunk,

			AssignAsset:   r.AssignAsset,
			UnassignAsset: r.UnassignAsset,

			AddUser:    r.AddUser,
			RemoveUser: r.RemoveUser,
			ChangeRole: r.ChangeRole,

			ChangeStatus:  r.ChangeStatus,
			SetDoneAsset:   r.SetDoneAsset,
			SetRetakeAsset: r.SetRetakeAsset,

			ViewDoneAsset:       r.ViewDoneAsset,
			ManageDependencies: r.ManageDependencies,
		}
	}
	return pb
}

func ToPbUserRoles(userRoles []models.UserRole) []*repositorypb.UserRole {
	pb := make([]*repositorypb.UserRole, len(userRoles))
	for i, ur := range userRoles {
		pb[i] = &repositorypb.UserRole{
			Mtime:  int64(ur.MTime),
			UserId: ur.UserUID,
			RoleId: ur.RoleId,
			Synced: ur.Synced,
		}
	}
	return pb
}

func ToPbTemplates(templates []models.Template) []*repositorypb.Template {
	pb := make([]*repositorypb.Template, len(templates))
	for i, t := range templates {
		pb[i] = &repositorypb.Template{
			Id:             t.Id,
			Mtime:          int64(t.MTime),
			Name:           t.Name,
			Extension:      t.Extension,
			Chunks:         t.Chunks,
			XxhashChecksum: t.XxhashChecksum,
			FileSize:       int64(t.FileSize),
			Trashed:        t.Trashed,
			Synced:         t.Synced,
		}
	}
	return pb
}

func ToPbPreviews(previews []models.Preview) []*repositorypb.Preview {
	pb := make([]*repositorypb.Preview, len(previews))
	for i, p := range previews {
		pb[i] = &repositorypb.Preview{
			Hash:      p.Hash,
			Preview:   p.Preview,
			Extension: p.Extension,
		}
	}
	return pb
}

func ToPbChunkInfos(chunkInfos []chunk_service.ChunkInfo) []*repositorypb.ChunkInfo {
	pb := make([]*repositorypb.ChunkInfo, len(chunkInfos))
	for i, p := range chunkInfos {
		pb[i] = &repositorypb.ChunkInfo{
			Hash: p.Hash,
			Size: int64(p.Size),
		}
	}
	return pb
}

func ToPbTombs(tombs []Tomb) []*repositorypb.Tomb {
	pb := make([]*repositorypb.Tomb, len(tombs))
	for i, t := range tombs {
		pb[i] = &repositorypb.Tomb{
			Id:        t.Id,
			Mtime:     int64(t.Mtime),
			TableName: t.TableName,
			Synced:    t.Synced,
		}
	}
	return pb
}

// type FullAsset struct {
// 	Id              string `db:"id" json:"id"`
// 	MTime           int    `db:"mtime" json:"mtime"`
// 	CreatedAt       string `db:"created_at" json:"created_at"`
// 	Name            string `db:"name" json:"name"`
// 	Description     string `db:"description" json:"description"`
// 	Extension       string `db:"extension" json:"extension"`
// 	IsResource      bool   `db:"is_resource" json:"is_resource"`
// 	StatusId        string `db:"status_id" json:"status_id"`
// 	StatusShortName string `db:"status_short_name" json:"status_short_name"`
// 	AssetTypeId      string `db:"asset_type_id" json:"asset_type_id"`
// 	AssetTypeName    string `db:"asset_type_name" json:"asset_type_name"`
// 	AssetTypeIcon    string `db:"asset_type_icon" json:"asset_type_icon"`
// 	CollectionId        string `db:"collection_id" json:"collection_id"`
// 	CollectionName      string `db:"collection_name" json:"collection_name"`
// 	CollectionPath      string `db:"collection_path" json:"collection_path"`
// 	AssetPath        string `db:"asset_path" json:"asset_path"`
// 	AssigneeId      string `db:"assignee_id" json:"assignee_id"`
// 	AssigneeEmail   string `db:"assignee_email" json:"assignee_email"`
// 	AssigneeName    string `db:"assignee_name" json:"assignee_name"`
// 	AssignerId      string `db:"assigner_id" json:"assigner_id"`
// 	AssignerEmail   string `db:"assigner_email" json:"assigner_email"`
// 	AssignerName    string `db:"assigner_name" json:"assigner_name"`
// 	// RelationshipType string   `db:"relationship_type" json:"relationship_type"`
// 	IsDependency    bool     `db:"is_dependency" json:"is_dependency"`
// 	DependencyLevel int      `db:"dependency_level" json:"-"`
// 	FilePath        string   `db:"file_path" json:"file_path"`
// 	Tags            []string `db:"-" json:"tags"`
// 	TagsRaw         string   `db:"tags" json:"-"`
// 	// Tags             []string `db:"tags" json:"tags"`
// 	CollectionDependencies    []string `db:"-" json:"collection_dependencies"`
// 	CollectionDependenciesRaw string   `db:"collection_dependencies" json:"-"`
// 	Dependencies          []string `db:"-" json:"dependencies"`
// 	DependenciesRaw       string   `db:"dependencies" json:"-"`
// 	// Dependencies     []string `db:"dependencies" json:"dependencies"`
// 	FileStatus       string       `db:"file_status" json:"file_status"`
// 	Status           Status       `db:"status" json:"status"`
// 	IsLink           bool         `db:"is_link" json:"is_link"`
// 	Pointer          string       `db:"pointer" json:"pointer"`
// 	PreviewId        string       `db:"preview_id" json:"preview_id"`
// 	Preview          []byte       `db:"preview" json:"preview"`
// 	PreviewExtension string       `db:"preview_extension" json:"preview_extension"`
// 	Checkpoints      []Checkpoint `db:"-" json:"checkpoints"`
// 	Trashed          bool         `db:"trashed" json:"trashed"`
// 	Synced           bool         `db:"synced" json:"synced"`
// 	Type             string       `db:"type" json:"type"`
// }

func ToPbFullAssets(assets []models.Asset) []*repositorypb.FullAsset {
	pb := make([]*repositorypb.FullAsset, len(assets))
	for i, t := range assets {
		pb[i] = &repositorypb.FullAsset{
			Id:                    t.Id,
			Mtime:                 int64(t.MTime),
			CreatedAt:             t.CreatedAt,
			Name:                  t.Name,
			Description:           t.Description,
			Extension:             t.Extension,
			IsResource:            t.IsResource,
			StatusId:              t.StatusId,
			StatusShortName:       t.StatusShortName,
			AssetTypeId:            t.AssetTypeId,
			AssetTypeName:          t.AssetTypeName,
			AssetTypeIcon:          t.AssetTypeIcon,
			CollectionId:              t.CollectionId,
			CollectionName:            t.CollectionName,
			CollectionPath:            t.CollectionPath,
			AssetPath:              t.AssetPath,
			AssigneeId:            t.AssigneeId,
			AssigneeEmail:         t.AssigneeEmail,
			AssigneeName:          t.AssigneeName,
			AssignerId:            t.AssignerId,
			AssignerEmail:         t.AssignerEmail,
			AssignerName:          t.AssignerName,
			IsDependency:          t.IsDependency,
			DependencyLevel:       int32(t.DependencyLevel),
			FilePath:              t.FilePath,
			Tags:                  t.Tags,
			TagsRaw:               t.TagsRaw,
			CollectionDependencies:    t.CollectionDependencies,
			CollectionDependenciesRaw: t.CollectionDependenciesRaw,
			Dependencies:          t.Dependencies,
			DependenciesRaw:       t.DependenciesRaw,
			FileStatus:            t.FileStatus,
			Status:                ToPbStatus(t.Status),
			IsLink:                t.IsLink,
			Pointer:               t.Pointer,
			PreviewId:             t.PreviewId,
			Preview:               t.Preview,
			PreviewExtension:      t.PreviewExtension,
			Checkpoints:           ToPbCheckpoints(t.Checkpoints),
			Trashed:               t.Trashed,
			Synced:                t.Synced,
			Type:                  "asset",
		}
	}
	return pb
}

// Helper for single Status
func ToPbStatus(s models.Status) *repositorypb.Status {
	return &repositorypb.Status{
		Id:        s.Id,
		Mtime:     int64(s.MTime),
		Name:      s.Name,
		ShortName: s.ShortName,
		Color:     s.Color,
		Synced:    s.Synced,
	}
}

// --- From Conversion helpers ---

func FromPbUser(pb *repositorypb.User) models.User {
	return models.User{
		Id:        pb.Id,
		MTime:     int(pb.Mtime),
		AddedAt:   pb.AddedAt,
		Username:  pb.Username,
		Email:     pb.Email,
		FirstName: pb.FirstName,
		LastName:  pb.LastName,
		Photo:     pb.Photo,
		RoleId:    pb.RoleId,
		Synced:    pb.Synced,
		Role:      models.Role{Name: pb.Role},
	}
}

func FromPbUsers(pbs []*repositorypb.User) []models.User {
	users := make([]models.User, len(pbs))
	for i, pb := range pbs {
		users[i] = FromPbUser(pb)
	}
	return users
}

func FromPbCollectionType(pb *repositorypb.CollectionType) models.CollectionType {
	return models.CollectionType{
		Id:     pb.Id,
		MTime:  int(pb.Mtime),
		Name:   pb.Name,
		Icon:   pb.Icon,
		Synced: pb.Synced,
	}
}

func FromPbCollectionTypes(pbs []*repositorypb.CollectionType) []models.CollectionType {
	collectionTypes := make([]models.CollectionType, len(pbs))
	for i, pb := range pbs {
		collectionTypes[i] = FromPbCollectionType(pb)
	}
	return collectionTypes
}

func FromPbAssetType(pb *repositorypb.AssetType) models.AssetType {
	return models.AssetType{
		Id:     pb.Id,
		MTime:  int(pb.Mtime),
		Name:   pb.Name,
		Icon:   pb.Icon,
		Synced: pb.Synced,
	}
}

func FromPbAssetTypes(pbs []*repositorypb.AssetType) []models.AssetType {
	assetTypes := make([]models.AssetType, len(pbs))
	for i, pb := range pbs {
		assetTypes[i] = FromPbAssetType(pb)
	}
	return assetTypes
}

func FromPbAsset(pb *repositorypb.Asset) models.Asset {
	return models.Asset{
		Id:          pb.Id,
		MTime:       int(pb.Mtime),
		CreatedAt:   pb.CreatedAt,
		Name:        pb.Name,
		Description: pb.Description,
		Extension:   pb.Extension,
		IsResource:  pb.IsResource,
		StatusId:    pb.StatusId,
		AssetTypeId:  pb.AssetTypeId,
		CollectionId:    pb.CollectionId,
		AssigneeId:  pb.AssigneeId,
		AssignerId:  pb.AssignerId,
		IsLink:      pb.IsLink,
		Pointer:     pb.Pointer,
		PreviewId:   pb.PreviewId,
		Trashed:     pb.Trashed,
		Synced:      pb.Synced,
	}
}

func FromPbAssets(pbs []*repositorypb.Asset) []models.Asset {
	assets := make([]models.Asset, len(pbs))
	for i, pb := range pbs {
		assets[i] = FromPbAsset(pb)
	}
	return assets
}

func FromPbCollection(pb *repositorypb.Collection) models.Collection {
	return models.Collection{
		Id:           pb.Id,
		MTime:        int(pb.Mtime),
		CreatedAt:    pb.CreatedAt,
		Name:         pb.Name,
		Description:  pb.Description,
		CollectionPath:   pb.CollectionPath,
		Trashed:      pb.Trashed,
		CollectionTypeId: pb.CollectionTypeId,
		ParentId:     pb.ParentId,
		PreviewId:    pb.PreviewId,
		Synced:       pb.Synced,
		IsLibrary:    pb.IsLibrary,
	}
}

func FromPbCollections(pbs []*repositorypb.Collection) []models.Collection {
	collections := make([]models.Collection, len(pbs))
	for i, pb := range pbs {
		collections[i] = FromPbCollection(pb)
	}
	return collections
}

func FromPbCollectionAssignee(pb *repositorypb.CollectionAssignee) models.CollectionAssignee {
	return models.CollectionAssignee{
		Id:         pb.Id,
		MTime:      int(pb.Mtime),
		CollectionId:   pb.CollectionId,
		AssigneeId: pb.AssigneeId,
		AssignerId: pb.AssignerId,
		Synced:     pb.Synced,
	}
}

func FromPbCollectionAssignees(pbs []*repositorypb.CollectionAssignee) []models.CollectionAssignee {
	collectionAssignees := make([]models.CollectionAssignee, len(pbs))
	for i, pb := range pbs {
		collectionAssignees[i] = FromPbCollectionAssignee(pb)
	}
	return collectionAssignees
}

func FromPbAssetDependency(pb *repositorypb.AssetDependency) models.AssetDependency {
	return models.AssetDependency{
		Id:               pb.Id,
		MTime:            int(pb.Mtime),
		AssetId:           pb.AssetId,
		DependencyId:     pb.DependencyId,
		DependencyTypeId: pb.DependencyTypeId,
		Synced:           pb.Synced,
	}
}

func FromPbAssetDependencies(pbs []*repositorypb.AssetDependency) []models.AssetDependency {
	assetDependencies := make([]models.AssetDependency, len(pbs))
	for i, pb := range pbs {
		assetDependencies[i] = FromPbAssetDependency(pb)
	}
	return assetDependencies
}

func FromPbCollectionDependency(pb *repositorypb.CollectionDependency) models.CollectionDependency {
	return models.CollectionDependency{
		Id:               pb.Id,
		MTime:            int(pb.Mtime),
		AssetId:           pb.AssetId,
		DependencyId:     pb.DependencyId,
		DependencyTypeId: pb.DependencyTypeId,
		Synced:           pb.Synced,
	}
}

func FromPbCollectionDependencies(pbs []*repositorypb.CollectionDependency) []models.CollectionDependency {
	collectionDependencies := make([]models.CollectionDependency, len(pbs))
	for i, pb := range pbs {
		collectionDependencies[i] = FromPbCollectionDependency(pb)
	}
	return collectionDependencies
}

func FromPbWorkflow(pb *repositorypb.Workflow) models.Workflow {
	return models.Workflow{
		Id:     pb.Id,
		MTime:  int(pb.Mtime),
		Name:   pb.Name,
		Synced: pb.Synced,
	}
}

func FromPbWorkflows(pbs []*repositorypb.Workflow) []models.Workflow {
	workflows := make([]models.Workflow, len(pbs))
	for i, pb := range pbs {
		workflows[i] = FromPbWorkflow(pb)
	}
	return workflows
}

func FromPbWorkflowAsset(pb *repositorypb.WorkflowAsset) models.WorkflowAsset {
	return models.WorkflowAsset{
		Id:               pb.Id,
		MTime:            int(pb.Mtime),
		Name:             pb.Name,
		TemplateId:       pb.TemplateId,
		IsResource:       pb.IsResource,
		WorkflowId:       pb.WorkflowId,
		AssetTypeId:       pb.AssetTypeId,
		WorkflowCollectionId: pb.WorkflowCollectionId,
		IsLink:           pb.IsLink,
		Pointer:          pb.Pointer,
		Synced:           pb.Synced,
	}
}

func FromPbWorkflowAssets(pbs []*repositorypb.WorkflowAsset) []models.WorkflowAsset {
	workflowAssets := make([]models.WorkflowAsset, len(pbs))
	for i, pb := range pbs {
		workflowAssets[i] = FromPbWorkflowAsset(pb)
	}
	return workflowAssets
}

func FromPbWorkflowCollection(pb *repositorypb.WorkflowCollection) models.WorkflowCollection {
	return models.WorkflowCollection{
		Id:           pb.Id,
		MTime:        int(pb.Mtime),
		Name:         pb.Name,
		WorkflowId:   pb.WorkflowId,
		CollectionTypeId: pb.CollectionTypeId,
		ParentId:     pb.ParentId,
		Synced:       pb.Synced,
	}
}

func FromPbWorkflowCollections(pbs []*repositorypb.WorkflowCollection) []models.WorkflowCollection {
	workflowCollections := make([]models.WorkflowCollection, len(pbs))
	for i, pb := range pbs {
		workflowCollections[i] = FromPbWorkflowCollection(pb)
	}
	return workflowCollections
}

func FromPbWorkflowLink(pb *repositorypb.WorkflowLink) models.WorkflowLink {
	return models.WorkflowLink{
		Id:                 pb.Id,
		MTime:              int(pb.Mtime),
		Name:               pb.Name,
		CollectionTypeId:       pb.CollectionTypeId,
		WorkflowId:         pb.WorkflowId,
		LinkedWorkflowId:   pb.LinkedWorkflowId,
		LinkedWorkflowName: pb.LinkedWorkflowName,
		Synced:             pb.Synced,
	}
}

func FromPbWorkflowLinks(pbs []*repositorypb.WorkflowLink) []models.WorkflowLink {
	workflowLinks := make([]models.WorkflowLink, len(pbs))
	for i, pb := range pbs {
		workflowLinks[i] = FromPbWorkflowLink(pb)
	}
	return workflowLinks
}

func FromPbDependencyType(pb *repositorypb.DependencyType) models.DependencyType {
	return models.DependencyType{
		Id:     pb.Id,
		MTime:  int(pb.Mtime),
		Name:   pb.Name,
		Synced: pb.Synced,
	}
}

func FromPbDependencyTypes(pbs []*repositorypb.DependencyType) []models.DependencyType {
	dependencyTypes := make([]models.DependencyType, len(pbs))
	for i, pb := range pbs {
		dependencyTypes[i] = FromPbDependencyType(pb)
	}
	return dependencyTypes
}

func FromPbStatus(pb *repositorypb.Status) models.Status {
	return models.Status{
		Id:        pb.Id,
		MTime:     int(pb.Mtime),
		Name:      pb.Name,
		ShortName: pb.ShortName,
		Color:     pb.Color,
		Synced:    pb.Synced,
	}
}

func FromPbStatuses(pbs []*repositorypb.Status) []models.Status {
	statuses := make([]models.Status, len(pbs))
	for i, pb := range pbs {
		statuses[i] = FromPbStatus(pb)
	}
	return statuses
}

func FromPbTag(pb *repositorypb.Tag) models.Tag {
	return models.Tag{
		Id:     pb.Id,
		MTime:  int(pb.Mtime),
		Name:   pb.Name,
		Synced: pb.Synced,
	}
}

func FromPbTags(pbs []*repositorypb.Tag) []models.Tag {
	tags := make([]models.Tag, len(pbs))
	for i, pb := range pbs {
		tags[i] = FromPbTag(pb)
	}
	return tags
}

func FromPbAssetTag(pb *repositorypb.AssetTag) models.AssetTag {
	return models.AssetTag{
		Id:     pb.Id,
		MTime:  int(pb.Mtime),
		AssetId: pb.AssetId,
		TagId:  pb.TagId,
		Synced: pb.Synced,
	}
}

func FromPbAssetTags(pbs []*repositorypb.AssetTag) []models.AssetTag {
	assetTags := make([]models.AssetTag, len(pbs))
	for i, pb := range pbs {
		assetTags[i] = FromPbAssetTag(pb)
	}
	return assetTags
}

func FromPbCheckpoint(pb *repositorypb.Checkpoint) models.Checkpoint {
	return models.Checkpoint{
		Id:             pb.Id,
		MTime:          int(pb.Mtime),
		CreatedAt:      pb.CreatedAt,
		AssetId:         pb.AssetId,
		XXHashChecksum: pb.XxhashChecksum,
		TimeModified:   int(pb.TimeModified),
		FileSize:       int(pb.FileSize),
		Comment:        pb.Comment,
		Chunks:         pb.Chunks,
		AuthorUID:      pb.AuthorUid,
		PreviewId:      pb.PreviewId,
		Trashed:        pb.Trashed,
		Synced:         pb.Synced,
		GroupId:        pb.GroupId,
	}
}

func FromPbCheckpoints(pbs []*repositorypb.Checkpoint) []models.Checkpoint {
	checkpoints := make([]models.Checkpoint, len(pbs))
	for i, pb := range pbs {
		checkpoints[i] = FromPbCheckpoint(pb)
	}
	return checkpoints
}

func FromPbRole(pb *repositorypb.Role) models.Role {
	return models.Role{
		Id:     pb.Id,
		MTime:  int(pb.Mtime),
		Name:   pb.Name,
		Synced: pb.Synced,

		ViewCollection:   pb.ViewCollection,
		CreateCollection: pb.CreateCollection,
		UpdateCollection: pb.UpdateCollection,
		DeleteCollection: pb.DeleteCollection,

		ViewAsset:   pb.ViewAsset,
		CreateAsset: pb.CreateAsset,
		UpdateAsset: pb.UpdateAsset,
		DeleteAsset: pb.DeleteAsset,

		ViewTemplate:   pb.ViewTemplate,
		CreateTemplate: pb.CreateTemplate,
		UpdateTemplate: pb.UpdateTemplate,
		DeleteTemplate: pb.DeleteTemplate,

		ViewCheckpoint:   pb.ViewCheckpoint,
		CreateCheckpoint: pb.CreateCheckpoint,
		DeleteCheckpoint: pb.DeleteCheckpoint,

		PullChunk: pb.PullChunk,

		AssignAsset:   pb.AssignAsset,
		UnassignAsset: pb.UnassignAsset,

		AddUser:    pb.AddUser,
		RemoveUser: pb.RemoveUser,
		ChangeRole: pb.ChangeRole,

		ChangeStatus:  pb.ChangeStatus,
		SetDoneAsset:   pb.SetDoneAsset,
		SetRetakeAsset: pb.SetRetakeAsset,

		ViewDoneAsset:       pb.ViewDoneAsset,
		ManageDependencies: pb.ManageDependencies,
	}
}

func FromPbRoles(pbs []*repositorypb.Role) []models.Role {
	roles := make([]models.Role, len(pbs))
	for i, pb := range pbs {
		roles[i] = FromPbRole(pb)
	}
	return roles
}

func FromPbUserRole(pb *repositorypb.UserRole) models.UserRole {
	return models.UserRole{
		MTime:   int(pb.Mtime),
		UserUID: pb.UserId,
		RoleId:  pb.RoleId,
		Synced:  pb.Synced,
	}
}

func FromPbUserRoles(pbs []*repositorypb.UserRole) []models.UserRole {
	userRoles := make([]models.UserRole, len(pbs))
	for i, pb := range pbs {
		userRoles[i] = FromPbUserRole(pb)
	}
	return userRoles
}

func FromPbTemplate(pb *repositorypb.Template) models.Template {
	return models.Template{
		Id:             pb.Id,
		MTime:          int(pb.Mtime),
		Name:           pb.Name,
		Extension:      pb.Extension,
		Chunks:         pb.Chunks,
		XxhashChecksum: pb.XxhashChecksum,
		FileSize:       int(pb.FileSize),
		Trashed:        pb.Trashed,
		Synced:         pb.Synced,
	}
}

func FromPbTemplates(pbs []*repositorypb.Template) []models.Template {
	templates := make([]models.Template, len(pbs))
	for i, pb := range pbs {
		templates[i] = FromPbTemplate(pb)
	}
	return templates
}

func FromPbPreview(pb *repositorypb.Preview) models.Preview {
	return models.Preview{
		Hash:      pb.Hash,
		Preview:   pb.Preview,
		Extension: pb.Extension,
	}
}

func FromPbPreviews(pbs []*repositorypb.Preview) []models.Preview {
	previews := make([]models.Preview, len(pbs))
	for i, pb := range pbs {
		previews[i] = FromPbPreview(pb)
	}
	return previews
}

func FromPbChunkInfo(pb *repositorypb.ChunkInfo) chunk_service.ChunkInfo {
	return chunk_service.ChunkInfo{
		Hash: pb.Hash,
		Size: int(pb.Size),
	}
}

func FromPbChunkInfos(pbs []*repositorypb.ChunkInfo) []chunk_service.ChunkInfo {
	chunkInfos := make([]chunk_service.ChunkInfo, len(pbs))
	for i, pb := range pbs {
		chunkInfos[i] = FromPbChunkInfo(pb)
	}
	return chunkInfos
}

func FromPbTomb(pb *repositorypb.Tomb) Tomb {
	return Tomb{
		Id:        pb.Id,
		Mtime:     int(pb.Mtime),
		TableName: pb.TableName,
		Synced:    pb.Synced,
	}
}

func FromPbTombs(pbs []*repositorypb.Tomb) []Tomb {
	tombs := make([]Tomb, len(pbs))
	for i, pb := range pbs {
		tombs[i] = FromPbTomb(pb)
	}
	return tombs
}

func ToPbIntegrationProjects(integrations []models.IntegrationProject) []*repositorypb.IntegrationProject {
	pb := make([]*repositorypb.IntegrationProject, len(integrations))
	for i, ip := range integrations {
		pb[i] = &repositorypb.IntegrationProject{
			Id:                  ip.Id,
			Mtime:               int64(ip.MTime),
			IntegrationId:       ip.IntegrationId,
			ExternalProjectId:   ip.ExternalProjectId,
			ExternalProjectName: ip.ExternalProjectName,
			ApiUrl:              ip.ApiUrl,
			SyncOptions:         ip.SyncOptions,
			LinkedByUserId:      ip.LinkedByUserId,
			LinkedAt:            ip.LinkedAt,
			Enabled:             ip.Enabled,
			Synced:              ip.Synced,
		}
	}
	return pb
}

func ToPbIntegrationCollectionMappings(mappings []models.IntegrationCollectionMapping) []*repositorypb.IntegrationCollectionMapping {
	pb := make([]*repositorypb.IntegrationCollectionMapping, len(mappings))
	for i, cm := range mappings {
		pb[i] = &repositorypb.IntegrationCollectionMapping{
			Id:               cm.Id,
			Mtime:            int64(cm.MTime),
			IntegrationId:    cm.IntegrationId,
			ExternalId:       cm.ExternalId,
			ExternalType:     cm.ExternalType,
			ExternalName:     cm.ExternalName,
			ExternalParentId: cm.ExternalParentId,
			ExternalPath:     cm.ExternalPath,
			ExternalMetadata: cm.ExternalMetadata,
			CollectionId:     cm.CollectionId,
			SyncedAt:         cm.SyncedAt,
			Synced:           cm.Synced,
		}
	}
	return pb
}

func ToPbIntegrationAssetMappings(mappings []models.IntegrationAssetMapping) []*repositorypb.IntegrationAssetMapping {
	pb := make([]*repositorypb.IntegrationAssetMapping, len(mappings))
	for i, am := range mappings {
		pb[i] = &repositorypb.IntegrationAssetMapping{
			Id:                     am.Id,
			Mtime:                  int64(am.MTime),
			IntegrationId:          am.IntegrationId,
			ExternalId:             am.ExternalId,
			ExternalName:           am.ExternalName,
			ExternalParentId:       am.ExternalParentId,
			ExternalType:           am.ExternalType,
			ExternalStatus:         am.ExternalStatus,
			ExternalAssignees:      am.ExternalAssignees,
			ExternalMetadata:       am.ExternalMetadata,
			AssetId:                am.AssetId,
			LastPushedCheckpointId: am.LastPushedCheckpointId,
			SyncedAt:               am.SyncedAt,
			Synced:                 am.Synced,
		}
	}
	return pb
}

// FromPbIntegrationProject converts a protobuf IntegrationProject to a model.
func FromPbIntegrationProject(pb *repositorypb.IntegrationProject) models.IntegrationProject {
	return models.IntegrationProject{
		Id:                  pb.Id,
		MTime:               int(pb.Mtime),
		IntegrationId:       pb.IntegrationId,
		ExternalProjectId:   pb.ExternalProjectId,
		ExternalProjectName: pb.ExternalProjectName,
		ApiUrl:              pb.ApiUrl,
		SyncOptions:         pb.SyncOptions,
		LinkedByUserId:      pb.LinkedByUserId,
		LinkedAt:            pb.LinkedAt,
		Enabled:             pb.Enabled,
		Synced:              pb.Synced,
	}
}

// FromPbIntegrationProjects converts a slice of protobuf IntegrationProjects to models.
func FromPbIntegrationProjects(pbs []*repositorypb.IntegrationProject) []models.IntegrationProject {
	items := make([]models.IntegrationProject, len(pbs))
	for i, pb := range pbs {
		items[i] = FromPbIntegrationProject(pb)
	}
	return items
}

// FromPbIntegrationCollectionMapping converts a protobuf IntegrationCollectionMapping to a model.
func FromPbIntegrationCollectionMapping(pb *repositorypb.IntegrationCollectionMapping) models.IntegrationCollectionMapping {
	return models.IntegrationCollectionMapping{
		Id:               pb.Id,
		MTime:            int(pb.Mtime),
		IntegrationId:    pb.IntegrationId,
		ExternalId:       pb.ExternalId,
		ExternalType:     pb.ExternalType,
		ExternalName:     pb.ExternalName,
		ExternalParentId: pb.ExternalParentId,
		ExternalPath:     pb.ExternalPath,
		ExternalMetadata: pb.ExternalMetadata,
		CollectionId:     pb.CollectionId,
		SyncedAt:         pb.SyncedAt,
		Synced:           pb.Synced,
	}
}

// FromPbIntegrationCollectionMappings converts a slice of protobuf IntegrationCollectionMappings to models.
func FromPbIntegrationCollectionMappings(pbs []*repositorypb.IntegrationCollectionMapping) []models.IntegrationCollectionMapping {
	items := make([]models.IntegrationCollectionMapping, len(pbs))
	for i, pb := range pbs {
		items[i] = FromPbIntegrationCollectionMapping(pb)
	}
	return items
}

// FromPbIntegrationAssetMapping converts a protobuf IntegrationAssetMapping to a model.
func FromPbIntegrationAssetMapping(pb *repositorypb.IntegrationAssetMapping) models.IntegrationAssetMapping {
	return models.IntegrationAssetMapping{
		Id:                     pb.Id,
		MTime:                  int(pb.Mtime),
		IntegrationId:          pb.IntegrationId,
		ExternalId:             pb.ExternalId,
		ExternalName:           pb.ExternalName,
		ExternalParentId:       pb.ExternalParentId,
		ExternalType:           pb.ExternalType,
		ExternalStatus:         pb.ExternalStatus,
		ExternalAssignees:      pb.ExternalAssignees,
		ExternalMetadata:       pb.ExternalMetadata,
		AssetId:                pb.AssetId,
		LastPushedCheckpointId: pb.LastPushedCheckpointId,
		SyncedAt:               pb.SyncedAt,
		Synced:                 pb.Synced,
	}
}

// FromPbIntegrationAssetMappings converts a slice of protobuf IntegrationAssetMappings to models.
func FromPbIntegrationAssetMappings(pbs []*repositorypb.IntegrationAssetMapping) []models.IntegrationAssetMapping {
	items := make([]models.IntegrationAssetMapping, len(pbs))
	for i, pb := range pbs {
		items[i] = FromPbIntegrationAssetMapping(pb)
	}
	return items
}
