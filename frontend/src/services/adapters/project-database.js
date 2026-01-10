/**
 * Project Database Manager
 * 
 * Manages SQLite databases in the browser using sql.js.
 * Each project gets its own database, persisted to IndexedDB.
 */

import initSqlJs from 'sql.js';
import { repository } from '@/lib/repositorypb.js';

// sql.js SQL instance (singleton)
let SQL = null;

// Map of project name -> database instance
const databases = new Map();

// IndexedDB database name
const IDB_NAME = 'clustta_project_dbs';
const IDB_STORE = 'databases';

/**
 * Initialize sql.js (load WASM)
 */
async function initSQL() {
  if (SQL) return SQL;
  
  SQL = await initSqlJs({
    // Load sql.js WASM from CDN
    locateFile: file => `https://sql.js.org/dist/${file}`
  });
  
  return SQL;
}

/**
 * Open IndexedDB for persistence
 */
function openIndexedDB() {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(IDB_NAME, 1);
    
    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result);
    
    request.onupgradeneeded = (event) => {
      const db = event.target.result;
      if (!db.objectStoreNames.contains(IDB_STORE)) {
        db.createObjectStore(IDB_STORE);
      }
    };
  });
}

/**
 * Save database to IndexedDB
 */
async function saveToIndexedDB(projectName, data) {
  const idb = await openIndexedDB();
  return new Promise((resolve, reject) => {
    const tx = idb.transaction(IDB_STORE, 'readwrite');
    const store = tx.objectStore(IDB_STORE);
    const request = store.put(data, projectName);
    
    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve();
    
    tx.oncomplete = () => idb.close();
  });
}

/**
 * Load database from IndexedDB
 */
async function loadFromIndexedDB(projectName) {
  const idb = await openIndexedDB();
  return new Promise((resolve, reject) => {
    const tx = idb.transaction(IDB_STORE, 'readonly');
    const store = tx.objectStore(IDB_STORE);
    const request = store.get(projectName);
    
    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result);
    
    tx.oncomplete = () => idb.close();
  });
}

/**
 * Create database schema
 */
function createSchema(db) {
  // Entity (Collection) table
  db.run(`
    CREATE TABLE IF NOT EXISTS entity (
      id TEXT PRIMARY KEY,
      created_at TEXT,
      mtime INTEGER,
      name TEXT,
      description TEXT,
      entity_type_id TEXT,
      parent_id TEXT,
      entity_path TEXT,
      preview_id TEXT DEFAULT '',
      is_library INTEGER DEFAULT 0,
      trashed INTEGER DEFAULT 0,
      synced INTEGER DEFAULT 1
    )
  `);

  // Task table
  db.run(`
    CREATE TABLE IF NOT EXISTS task (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      created_at TEXT,
      name TEXT,
      description TEXT,
      extension TEXT,
      is_resource INTEGER DEFAULT 0,
      status_id TEXT,
      task_type_id TEXT,
      entity_id TEXT,
      assignee_id TEXT,
      assigner_id TEXT,
      is_link INTEGER DEFAULT 0,
      pointer TEXT,
      preview_id TEXT,
      trashed INTEGER DEFAULT 0,
      synced INTEGER DEFAULT 1
    )
  `);

  // Checkpoint table
  db.run(`
    CREATE TABLE IF NOT EXISTS task_checkpoint (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      created_at TEXT,
      task_id TEXT,
      xxhash_checksum TEXT,
      time_modified INTEGER,
      file_size INTEGER,
      comment TEXT,
      chunks TEXT,
      author_id TEXT,
      preview_id TEXT,
      group_id TEXT,
      trashed INTEGER DEFAULT 0,
      synced INTEGER DEFAULT 1
    )
  `);

  // User table
  db.run(`
    CREATE TABLE IF NOT EXISTS user (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      added_at TEXT,
      username TEXT,
      email TEXT,
      first_name TEXT,
      last_name TEXT,
      photo BLOB,
      role_id TEXT,
      synced INTEGER DEFAULT 1
    )
  `);

  // Role table
  db.run(`
    CREATE TABLE IF NOT EXISTS role (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      name TEXT,
      view_entity INTEGER DEFAULT 0,
      create_entity INTEGER DEFAULT 0,
      update_entity INTEGER DEFAULT 0,
      delete_entity INTEGER DEFAULT 0,
      view_task INTEGER DEFAULT 0,
      create_task INTEGER DEFAULT 0,
      update_task INTEGER DEFAULT 0,
      delete_task INTEGER DEFAULT 0,
      view_template INTEGER DEFAULT 0,
      create_template INTEGER DEFAULT 0,
      update_template INTEGER DEFAULT 0,
      delete_template INTEGER DEFAULT 0,
      view_checkpoint INTEGER DEFAULT 0,
      create_checkpoint INTEGER DEFAULT 0,
      delete_checkpoint INTEGER DEFAULT 0,
      pull_chunk INTEGER DEFAULT 0,
      assign_task INTEGER DEFAULT 0,
      unassign_task INTEGER DEFAULT 0,
      add_user INTEGER DEFAULT 0,
      remove_user INTEGER DEFAULT 0,
      change_role INTEGER DEFAULT 0,
      change_status INTEGER DEFAULT 0,
      set_done_task INTEGER DEFAULT 0,
      set_retake_task INTEGER DEFAULT 0,
      view_done_task INTEGER DEFAULT 0,
      manage_dependencies INTEGER DEFAULT 0,
      synced INTEGER DEFAULT 1
    )
  `);

  // Status table
  db.run(`
    CREATE TABLE IF NOT EXISTS status (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      name TEXT,
      short_name TEXT,
      color TEXT,
      synced INTEGER DEFAULT 1
    )
  `);

  // Entity type table
  db.run(`
    CREATE TABLE IF NOT EXISTS entity_type (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      name TEXT,
      icon TEXT,
      synced INTEGER DEFAULT 1
    )
  `);

  // Task type table
  db.run(`
    CREATE TABLE IF NOT EXISTS task_type (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      name TEXT,
      icon TEXT,
      synced INTEGER DEFAULT 1
    )
  `);

  // Tag table
  db.run(`
    CREATE TABLE IF NOT EXISTS tag (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      name TEXT,
      color TEXT,
      synced INTEGER DEFAULT 1
    )
  `);

  // Task tag junction table
  db.run(`
    CREATE TABLE IF NOT EXISTS task_tag (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      task_id TEXT,
      tag_id TEXT,
      synced INTEGER DEFAULT 1
    )
  `);

  // Dependency type table
  db.run(`
    CREATE TABLE IF NOT EXISTS dependency_type (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      name TEXT,
      synced INTEGER DEFAULT 1
    )
  `);

  // Task dependency table
  db.run(`
    CREATE TABLE IF NOT EXISTS task_dependency (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      task_id TEXT,
      dependency_id TEXT,
      dependency_type_id TEXT,
      synced INTEGER DEFAULT 1
    )
  `);

  // Entity dependency table
  db.run(`
    CREATE TABLE IF NOT EXISTS entity_dependency (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      task_id TEXT,
      entity_id TEXT,
      dependency_type_id TEXT,
      synced INTEGER DEFAULT 1
    )
  `);

  // Entity assignee table
  db.run(`
    CREATE TABLE IF NOT EXISTS entity_assignee (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      entity_id TEXT,
      assignee_id TEXT DEFAULT '',
      assigner_id TEXT DEFAULT '',
      synced INTEGER DEFAULT 1
    )
  `);

  // Template table
  db.run(`
    CREATE TABLE IF NOT EXISTS template (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      name TEXT,
      extension TEXT,
      chunks TEXT,
      xxhash_checksum TEXT,
      file_size INTEGER,
      trashed INTEGER DEFAULT 0,
      synced INTEGER DEFAULT 1
    )
  `);

  // Workflow tables
  db.run(`
    CREATE TABLE IF NOT EXISTS workflow (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      name TEXT,
      synced INTEGER DEFAULT 1
    )
  `);

  db.run(`
    CREATE TABLE IF NOT EXISTS workflow_link (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      workflow_id TEXT,
      source_id TEXT,
      target_id TEXT,
      source_handle TEXT,
      target_handle TEXT,
      synced INTEGER DEFAULT 1
    )
  `);

  db.run(`
    CREATE TABLE IF NOT EXISTS workflow_entity (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      workflow_id TEXT,
      name TEXT,
      entity_type_id TEXT,
      pos_x REAL,
      pos_y REAL,
      synced INTEGER DEFAULT 1
    )
  `);

  db.run(`
    CREATE TABLE IF NOT EXISTS workflow_task (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      workflow_id TEXT,
      workflow_entity_id TEXT,
      name TEXT,
      task_type_id TEXT,
      extension TEXT,
      template_id TEXT,
      is_resource INTEGER DEFAULT 0,
      pos_x REAL,
      pos_y REAL,
      synced INTEGER DEFAULT 1
    )
  `);

  // Project metadata table
  db.run(`
    CREATE TABLE IF NOT EXISTS project_meta (
      key TEXT PRIMARY KEY,
      value TEXT
    )
  `);

  // Tomb (deleted items) table
  db.run(`
    CREATE TABLE IF NOT EXISTS tomb (
      id TEXT PRIMARY KEY,
      mtime INTEGER,
      table_name TEXT,
      dtime INTEGER,
      synced INTEGER DEFAULT 1
    )
  `);

  // Create indexes for common queries
  db.run('CREATE INDEX IF NOT EXISTS idx_entity_parent ON entity(parent_id)');
  db.run('CREATE INDEX IF NOT EXISTS idx_task_entity ON task(entity_id)');
  db.run('CREATE INDEX IF NOT EXISTS idx_checkpoint_task ON task_checkpoint(task_id)');
  db.run('CREATE INDEX IF NOT EXISTS idx_task_tag_task ON task_tag(task_id)');
  db.run('CREATE INDEX IF NOT EXISTS idx_task_dependency_task ON task_dependency(task_id)');
  db.run('CREATE INDEX IF NOT EXISTS idx_tomb_table ON tomb(table_name)');
}

/**
 * Helper to convert undefined to null for SQLite binding
 * sql.js cannot bind undefined values
 */
function n(value) {
  return value === undefined ? null : value;
}

/**
 * Helper to convert to number, handling undefined/null
 */
function num(value) {
  if (value === undefined || value === null) return 0;
  return Number(value);
}

/**
 * Helper to convert to string, handling undefined/null
 */
function str(value) {
  if (value === undefined || value === null) return '';
  return String(value);
}

/**
 * Helper to convert to boolean int (0 or 1)
 */
function bool(value) {
  return value ? 1 : 0;
}

/**
 * Populate database from decoded ProjectData protobuf
 */
function populateFromProjectData(db, projectData) {
    console.log(projectData)
  console.log('[populateFromProjectData] Starting population with data:', {
    entities: projectData.entities?.length || 0,
    tasks: projectData.tasks?.length || 0,
    checkpoints: projectData.tasks_checkpoints?.length || 0,
    users: projectData.users?.length || 0,
  });

  // Clear existing data
  const tables = [
    'entity', 'task', 'task_checkpoint', 'user', 'role', 'status',
    'entity_type', 'task_type', 'tag', 'task_tag', 'dependency_type',
    'task_dependency', 'entity_dependency', 'entity_assignee', 'template',
    'workflow', 'workflow_link', 'workflow_entity', 'workflow_task', 'tomb'
  ];
  
  for (const table of tables) {
    db.run(`DELETE FROM ${table}`);
  }

  // Insert entities
  if (projectData.entities?.length) {
    const stmt = db.prepare(`
      INSERT INTO entity (id, mtime, name, parent_id, entity_type_id, entity_path, trashed, synced)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `);
    for (const e of projectData.entities) {
      stmt.run([n(e.id), num(e.mtime), str(e.name), n(e.parent_id), n(e.entity_type_id), str(e.entity_path), bool(e.trashed), 1]);
    }
    stmt.free();
  }

  // Insert tasks
  if (projectData.tasks?.length) {
    const stmt = db.prepare(`
      INSERT INTO task (id, mtime, created_at, name, description, extension, is_resource, status_id, 
                       task_type_id, entity_id, assignee_id, assigner_id, is_link, pointer, preview_id, trashed, synced)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    for (const t of projectData.tasks) {
      stmt.run([
        n(t.id), num(t.mtime), str(t.created_at), str(t.name), str(t.description), str(t.extension),
        bool(t.is_resource), n(t.status_id), n(t.task_type_id), n(t.entity_id),
        n(t.assignee_id), n(t.assigner_id), bool(t.is_link), str(t.pointer), n(t.preview_id),
        bool(t.trashed), 1
      ]);
    }
    stmt.free();
  }

  // Insert checkpoints
  if (projectData.tasks_checkpoints?.length) {
    const stmt = db.prepare(`
      INSERT INTO task_checkpoint (id, mtime, created_at, task_id, xxhash_checksum, time_modified, file_size, comment, chunks, author_id, preview_id, group_id, trashed, synced)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    for (const c of projectData.tasks_checkpoints) {
      stmt.run([
        n(c.id), num(c.mtime), str(c.created_at), n(c.task_id), str(c.xxhash_checksum),
        num(c.time_modified), num(c.file_size), str(c.comment),
        str(c.chunks), str(c.author_uid), n(c.preview_id), str(c.group_id),
        bool(c.trashed), 1
      ]);
    }
    stmt.free();
  }

  // Insert users
  if (projectData.users?.length) {
    const stmt = db.prepare(`
      INSERT INTO user (id, mtime, added_at, username, email, first_name, last_name, photo, role_id, synced)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    for (const u of projectData.users) {
      stmt.run([
        n(u.id), num(u.mtime), str(u.added_at), str(u.username), str(u.email),
        str(u.first_name), str(u.last_name), n(u.photo), n(u.role_id), 1
      ]);
    }
    stmt.free();
  }

  // Insert roles
  if (projectData.roles?.length) {
    const stmt = db.prepare(`
      INSERT INTO role (id, mtime, name, view_entity, create_entity, update_entity, delete_entity,
                       view_task, create_task, update_task, delete_task,
                       view_template, create_template, update_template, delete_template,
                       view_checkpoint, create_checkpoint, delete_checkpoint, pull_chunk,
                       assign_task, unassign_task, add_user, remove_user, change_role,
                       change_status, set_done_task, set_retake_task, view_done_task,
                       manage_dependencies, synced)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    for (const r of projectData.roles) {
      stmt.run([
        n(r.id), num(r.mtime), str(r.name),
        bool(r.view_entity), bool(r.create_entity), bool(r.update_entity), bool(r.delete_entity),
        bool(r.view_task), bool(r.create_task), bool(r.update_task), bool(r.delete_task),
        bool(r.view_template), bool(r.create_template), bool(r.update_template), bool(r.delete_template),
        bool(r.view_checkpoint), bool(r.create_checkpoint), bool(r.delete_checkpoint), bool(r.pull_chunk),
        bool(r.assign_task), bool(r.unassign_task), bool(r.add_user), bool(r.remove_user), bool(r.change_role),
        bool(r.change_status), bool(r.set_done_task), bool(r.set_retake_task), bool(r.view_done_task),
        bool(r.manage_dependencies), 1
      ]);
    }
    stmt.free();
  }

  // Insert statuses
  if (projectData.statuses?.length) {
    const stmt = db.prepare('INSERT INTO status (id, mtime, name, short_name, color, synced) VALUES (?, ?, ?, ?, ?, ?)');
    for (const s of projectData.statuses) {
      stmt.run([n(s.id), num(s.mtime), str(s.name), str(s.short_name), str(s.color), 1]);
    }
    stmt.free();
  }

  // Insert entity types
  if (projectData.entity_types?.length) {
    const stmt = db.prepare('INSERT INTO entity_type (id, mtime, name, icon, synced) VALUES (?, ?, ?, ?, ?)');
    for (const t of projectData.entity_types) {
      stmt.run([n(t.id), num(t.mtime), str(t.name), str(t.icon), 1]);
    }
    stmt.free();
  }

  // Insert task types
  if (projectData.task_types?.length) {
    const stmt = db.prepare('INSERT INTO task_type (id, mtime, name, icon, synced) VALUES (?, ?, ?, ?, ?)');
    for (const t of projectData.task_types) {
      stmt.run([n(t.id), num(t.mtime), str(t.name), str(t.icon), 1]);
    }
    stmt.free();
  }

  // Insert tags
  if (projectData.tags?.length) {
    const stmt = db.prepare('INSERT INTO tag (id, mtime, name, color, synced) VALUES (?, ?, ?, ?, ?)');
    for (const t of projectData.tags) {
      stmt.run([n(t.id), num(t.mtime), str(t.name), str(t.color), 1]);
    }
    stmt.free();
  }

  // Insert task tags
  if (projectData.tasks_tags?.length) {
    const stmt = db.prepare('INSERT INTO task_tag (id, mtime, task_id, tag_id, synced) VALUES (?, ?, ?, ?, ?)');
    for (const tt of projectData.tasks_tags) {
      stmt.run([n(tt.id), num(tt.mtime), n(tt.task_id), n(tt.tag_id), 1]);
    }
    stmt.free();
  }

  // Insert dependency types
  if (projectData.dependency_types?.length) {
    const stmt = db.prepare('INSERT INTO dependency_type (id, mtime, name, synced) VALUES (?, ?, ?, ?)');
    for (const dt of projectData.dependency_types) {
      stmt.run([n(dt.id), num(dt.mtime), str(dt.name), 1]);
    }
    stmt.free();
  }

  // Insert task dependencies
  if (projectData.task_dependencies?.length) {
    const stmt = db.prepare('INSERT INTO task_dependency (id, mtime, task_id, dependency_id, dependency_type_id, synced) VALUES (?, ?, ?, ?, ?, ?)');
    for (const td of projectData.task_dependencies) {
      stmt.run([n(td.id), num(td.mtime), n(td.task_id), n(td.dependency_id), n(td.dependency_type_id), 1]);
    }
    stmt.free();
  }

  // Insert entity dependencies
  if (projectData.entity_dependencies?.length) {
    const stmt = db.prepare('INSERT INTO entity_dependency (id, mtime, task_id, entity_id, dependency_type_id, synced) VALUES (?, ?, ?, ?, ?, ?)');
    for (const ed of projectData.entity_dependencies) {
      stmt.run([n(ed.id), num(ed.mtime), n(ed.task_id), n(ed.entity_id), n(ed.dependency_type_id), 1]);
    }
    stmt.free();
  }

  // Insert entity assignees
  if (projectData.entity_assignees?.length) {
    const stmt = db.prepare('INSERT INTO entity_assignee (id, mtime, entity_id, assignee_id, assigner_id, synced) VALUES (?, ?, ?, ?, ?, ?)');
    for (const ea of projectData.entity_assignees) {
      stmt.run([n(ea.id), num(ea.mtime), n(ea.entity_id), n(ea.assignee_id), n(ea.assigner_id), 1]);
    }
    stmt.free();
  }

  // Insert templates
  if (projectData.templates?.length) {
    const stmt = db.prepare(`
      INSERT INTO template (id, mtime, name, extension, chunks, xxhash_checksum, file_size, trashed, synced)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    for (const t of projectData.templates) {
      stmt.run([n(t.id), num(t.mtime), str(t.name), str(t.extension), str(t.chunks), str(t.xxhash_checksum), num(t.file_size), bool(t.trashed), 1]);
    }
    stmt.free();
  }

  // Insert workflows
  if (projectData.workflows?.length) {
    const stmt = db.prepare('INSERT INTO workflow (id, mtime, name, synced) VALUES (?, ?, ?, ?)');
    for (const w of projectData.workflows) {
      stmt.run([n(w.id), num(w.mtime), str(w.name), 1]);
    }
    stmt.free();
  }

  // Insert workflow links
  if (projectData.workflow_links?.length) {
    const stmt = db.prepare(`
      INSERT INTO workflow_link (id, mtime, workflow_id, source_id, target_id, source_handle, target_handle, synced)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `);
    for (const wl of projectData.workflow_links) {
      stmt.run([n(wl.id), num(wl.mtime), n(wl.workflow_id), n(wl.source_id), n(wl.target_id), str(wl.source_handle), str(wl.target_handle), 1]);
    }
    stmt.free();
  }

  // Insert workflow entities
  if (projectData.workflow_entities?.length) {
    const stmt = db.prepare(`
      INSERT INTO workflow_entity (id, mtime, workflow_id, name, entity_type_id, pos_x, pos_y, synced)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `);
    for (const we of projectData.workflow_entities) {
      stmt.run([n(we.id), num(we.mtime), n(we.workflow_id), str(we.name), n(we.entity_type_id), num(we.pos_x), num(we.pos_y), 1]);
    }
    stmt.free();
  }

  // Insert workflow tasks
  if (projectData.workflow_tasks?.length) {
    const stmt = db.prepare(`
      INSERT INTO workflow_task (id, mtime, workflow_id, workflow_entity_id, name, task_type_id, extension, template_id, is_resource, pos_x, pos_y, synced)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    for (const wt of projectData.workflow_tasks) {
      stmt.run([
        n(wt.id), num(wt.mtime), n(wt.workflow_id), n(wt.workflow_entity_id), str(wt.name),
        n(wt.task_type_id), str(wt.extension), n(wt.template_id), bool(wt.is_resource),
        num(wt.pos_x), num(wt.pos_y), 1
      ]);
    }
    stmt.free();
  }

  // Store project preview
  if (projectData.project_preview) {
    db.run('INSERT OR REPLACE INTO project_meta (key, value) VALUES (?, ?)', ['project_preview', str(projectData.project_preview)]);
  }

  // Insert tomb (deleted items)
  if (projectData.tomb?.length) {
    const stmt = db.prepare('INSERT INTO tomb (id, mtime, table_name, dtime, synced) VALUES (?, ?, ?, ?, ?)');
    for (const t of projectData.tomb) {
      stmt.run([n(t.id), num(t.mtime), str(t.table_name), num(t.mtime), 1]); // Using mtime as dtime if not available
    }
    stmt.free();
  }
}

/**
 * Convert SQL result to array of objects
 */
function resultToObjects(result) {
  if (!result || !result.length) return [];
  
  const [{ columns, values }] = result;
  return values.map(row => {
    const obj = {};
    columns.forEach((col, i) => {
      obj[col] = row[i];
    });
    return obj;
  });
}

/**
 * Get or create database for a project
 */
export async function getDatabase(projectName) {
  if (databases.has(projectName)) {
    return databases.get(projectName);
  }

  await initSQL();

  // Try to load from IndexedDB
  const savedData = await loadFromIndexedDB(projectName);
  
  let db;
  if (savedData) {
    db = new SQL.Database(new Uint8Array(savedData));
  } else {
    db = new SQL.Database();
    createSchema(db);
  }

  databases.set(projectName, db);
  return db;
}

/**
 * Initialize database with fresh data from protobuf
 */
export async function initializeDatabase(projectName, projectData) {
  await initSQL();

  // Create new database
  const db = new SQL.Database();
  createSchema(db);
  
  // Populate from protobuf
  populateFromProjectData(db, projectData);
  
  // Save to IndexedDB
  const data = db.export();
  await saveToIndexedDB(projectName, data.buffer);
  
  // Store in memory
  databases.set(projectName, db);
  
  return db;
}

/**
 * Save current database state to IndexedDB
 */
export async function persistDatabase(projectName) {
  const db = databases.get(projectName);
  if (!db) return;
  
  const data = db.export();
  await saveToIndexedDB(projectName, data.buffer);
}

/**
 * Check if a project database exists
 */
export async function databaseExists(projectName) {
  if (databases.has(projectName)) return true;
  
  const savedData = await loadFromIndexedDB(projectName);
  return !!savedData;
}

/**
 * Clear a project database
 */
export async function clearDatabase(projectName) {
  const db = databases.get(projectName);
  if (db) {
    db.close();
    databases.delete(projectName);
  }
  
  // Remove from IndexedDB
  const idb = await openIndexedDB();
  return new Promise((resolve, reject) => {
    const tx = idb.transaction(IDB_STORE, 'readwrite');
    const store = tx.objectStore(IDB_STORE);
    const request = store.delete(projectName);
    
    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve();
    
    tx.oncomplete = () => idb.close();
  });
}

/**
 * Query helper - execute SELECT and return objects
 */
export function query(db, sql, params = []) {
  try {
    const result = db.exec(sql, params);
    return resultToObjects(result);
  } catch (error) {
    console.error('Query error:', error, sql);
    return [];
  }
}

/**
 * Execute helper - run INSERT/UPDATE/DELETE
 */
export function execute(db, sql, params = []) {
  try {
    db.run(sql, params);
    return true;
  } catch (error) {
    console.error('Execute error:', error, sql);
    return false;
  }
}

/**
 * Get single row helper
 */
export function queryOne(db, sql, params = []) {
  const results = query(db, sql, params);
  return results[0] || null;
}

// Export the ProjectData decoder for use in sync service
export { repository };
