package agent

// knowledgeBase contains curated documentation about Clustta's features and concepts.
// Used to provide the LLM with context about how Clustta works so it can answer user questions.
var knowledgeBase = map[string]string{
	"overview": `Clustta is a distributed version control and collaboration system for creative workflows — like Git/GitHub but designed for creative work (3D, VFX, animation, game dev). It manages binary files, large assets, and creative pipelines.

Key concepts:
- Projects: A .clst SQLite database file that stores all metadata for a creative project
- Collections (also called Entities): Folder-like containers that organize assets (e.g., "Characters", "Environments", "Props")
- Assets (also called Tasks): Individual files being version-controlled (e.g., a Blender file, a texture, a rig)
- Checkpoints: Saved versions/snapshots of assets — like Git commits but for binary files
- Studios: Team/organization servers for collaboration
- Statuses: Track the state of assets (e.g., Todo, In Progress, Review, Done)
- Task Types: Categorize assets by their role (e.g., Model, Rig, Animation, Texture)
- Templates: Predefined file templates that can be used when creating new assets
- Tags: Labels for organizing and filtering assets`,

	"collections": `Collections are folder-like containers for organizing assets in a Clustta project.

Creating a collection:
- Give it a name, optionally a description, entity type, and parent collection
- Collections can be nested (parent-child hierarchy)
- Collections can be marked as "library" collections for shared resources

Collection operations:
- Create, rename, delete collections
- View all collections or get children of a specific collection
- Each collection has a file path on disk where its assets are stored
- Collections can have assignees (team members responsible for that section)
- Collections track state: untracked files, modified assets, outdated assets`,

	"assets": `Assets are individual files being version-controlled in Clustta.

Creating an asset:
- Provide a name, task type (e.g., Model, Rig, Animation), and target collection
- Optionally: description, tags, template, preview image
- Assets can be "resources" (shared/reusable) or regular task assets
- Assets can be "link" assets that point to external files

Asset operations:
- Create, rename, delete, duplicate assets
- Change status (e.g., from "Todo" to "In Progress")
- Assign/unassign team members
- Move assets between collections
- Change asset type
- Add preview images
- Create checkpoints (save versions)
- View dependencies between assets`,

	"checkpoints": `Checkpoints are saved versions of assets — similar to Git commits but designed for large binary files.

How checkpoints work:
- When you checkpoint an asset, Clustta saves the current state of the file
- Files are chunked and deduplicated for efficient storage
- Each checkpoint has a comment, author, and timestamp
- You can revert to any previous checkpoint
- Checkpoints can be synced to a studio server for collaboration

Creating a checkpoint:
- Select modified or untracked assets
- Provide a comment describing the changes
- Checkpoints are grouped by operation (batch checkpoint)`,

	"sync": `Syncing pushes and pulls project data between your local machine and a studio server.

How sync works:
- Push: Upload your local checkpoints to the studio server
- Pull: Download new checkpoints from the studio server
- Sync: Push and pull in one operation
- Conflicts: When the same asset is modified locally and remotely, Clustta detects conflicts
- Conflict resolution: Choose to rename your version or merge with the server version

Prerequisites:
- Project must be connected to a studio server
- You need appropriate permissions (collaborator role)

Related operations:
- Clone: Download an entire project from a studio server
- Discard changes: Revert local modifications`,

	"studios": `Studios are team/organization servers for collaboration in Clustta.

How studios work:
- A studio is a self-hosted server (clustta-studio) that manages team collaboration
- Projects are published to studios for sharing with team members
- Team members can clone, sync, push, and pull project data

Studio operations:
- Register/connect to a studio server
- Add/remove collaborators with specific roles
- Manage roles and permissions
- Pin/unpin projects for quick access
- View studio status and usage`,

	"collaborators": `To add collaborators to a Clustta project:

Prerequisites:
1. Your project must be connected to a studio server
2. The studio server must be running and accessible
3. You need admin or appropriate permissions

Steps:
1. Connect your project to a studio (if not already connected)
2. Navigate to the project settings or studio panel
3. Add collaborators by their email address or username
4. Assign roles to control permissions (view, create, update, delete)

Roles control what collaborators can do:
- View entities/tasks
- Create entities/tasks
- Update entities/tasks
- Delete entities/tasks

Collaborators can then clone the project and sync changes through the studio server.`,

	"dependencies": `Dependencies in Clustta link assets that rely on each other.

How dependencies work:
- An asset can depend on other assets (e.g., a character rig depends on the character model)
- Dependencies can be at the asset level or collection level
- Dependency types categorize the relationship (defined per project)
- When an asset's dependency is updated, Clustta can flag the dependent asset as "outdated" or "rebuildable"

Use cases:
- A texture asset depends on a UV layout
- An animation asset depends on a character rig
- A lighting scene depends on environment models

Dependency presets:
- You can save dependency configurations as presets for reuse
- Presets store common dependency patterns for your pipeline`,

	"integrations": `Clustta integrates with external production tracking tools.

Currently supported:
- Kitsu: Open-source production tracker for animation/VFX

How integration works:
1. Authenticate with the external service (provide credentials)
2. Link your Clustta project to an external project
3. Sync data: map Clustta collections to external entities, assets to external tasks
4. Push status updates, previews, and progress to the external tool

Integration operations:
- Authenticate with credentials
- View available external projects
- Link/unlink projects
- Preview sync changes before executing
- Execute sync to push/pull data`,

	"workflows": `Workflows define the pipeline stages and structure for a project.

How workflows work:
- A workflow is a template that defines what collections and assets a project section should have
- Workflows contain: entities (collection templates), tasks (asset templates), and links (connections to other workflows)
- When you "add" a workflow, it creates the collections and assets defined in the template

Use cases:
- Define a standard character pipeline: Concept → Model → Rig → Animation
- Define an environment pipeline: Concept → Block-out → Detail → Lighting
- Reuse pipeline structures across projects or sections`,

	"templates": `Templates are predefined files used when creating new assets.

How templates work:
- Create a template by providing a name and file path
- When creating a new asset, select a template to pre-populate the file
- Templates are stored per-project
- Common templates: empty Blender files, Maya scenes, Photoshop documents

Template operations:
- Create, rename, delete templates
- Change the file associated with a template
- List all available templates`,

	"tags": `Tags are labels for organizing and filtering assets.

How tags work:
- Tags are defined per project
- Multiple tags can be applied to a single asset
- Tags can be added when creating an asset or updated later
- Use tags to filter and find assets quickly

Common tag patterns:
- Priority: high, medium, low
- Department: modeling, rigging, animation, compositing
- Phase: concept, production, final
- Custom labels specific to your project`,

	"statuses": `Statuses track the current state of assets in a project.

Default statuses typically include:
- Todo: Not yet started
- In Progress: Currently being worked on
- Review: Ready for review
- Done: Completed and approved

Status operations:
- View all available statuses
- Change the status of one or more assets
- Statuses have a name, short name, and color
- Statuses are project-specific and can be customized`,

	"file_operations": `Clustta manages files on the local file system alongside the project database.

How file management works:
- Each project has a working directory where files are stored on disk
- Collections map to directories
- Assets map to files within those directories
- Checkpoints store binary diffs of file versions
- Untracked files: files on disk not yet registered as assets
- Modified files: tracked assets whose files have changed since last checkpoint

File operations available:
- View file status (modified, untracked, outdated)
- Open asset files in their default application
- Monitor file system changes in real-time
- Ignore specific files or patterns (like .gitignore)`,

	"scripts": `You can generate scripts to perform batch operations on Clustta project files.

Common script scenarios:
- Blender background render: Run Blender in background mode to render assets
  Example: blender -b file.blend -o //output -a
- Batch file conversion: Convert file formats using command-line tools
- Automated exports: Export assets from DCCs (Digital Content Creation tools)
- File analysis: Scan assets for issues (missing textures, broken references)

Scripts are generated for review and manual execution — they are never auto-executed by the agent.`,
}
