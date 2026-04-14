import { studioApiCall, getActiveStudioUrl } from './http-client.js';
import { getDatabase, query, queryOne, execute, persistDatabase } from './project-database.js';

/**
 * Extract project name from project path/uri
 */
function getProjectName(projectPath) {
  if (!projectPath) return '';
  return projectPath.split('/').pop()?.replace('.clst', '') || projectPath;
}

/**
 * Convert database row to user object
 */
function rowToUser(row) {
  if (!row) return null;
  return {
    ...row,
    synced: !!row.synced,
  };
}

/**
 * Convert database row to role object
 */
function rowToRole(row) {
  if (!row) return null;
  return {
    ...row,
    view_asset: !!row.view_asset,
    create_asset: !!row.create_asset,
    edit_asset: !!row.edit_asset,
    delete_asset: !!row.delete_asset,
    create_collection: !!row.create_collection,
    edit_collection: !!row.edit_collection,
    delete_collection: !!row.delete_collection,
    create_checkpoint: !!row.create_checkpoint,
    revert_checkpoint: !!row.revert_checkpoint,
    delete_checkpoint: !!row.delete_checkpoint,
    download_checkpoint: !!row.download_checkpoint,
    change_assignee: !!row.change_assignee,
    create_tag: !!row.create_tag,
    delete_tag: !!row.delete_tag,
    create_status: !!row.create_status,
    delete_status: !!row.delete_status,
    create_type: !!row.create_type,
    delete_type: !!row.delete_type,
    synced: !!row.synced,
  };
}

export const UserService = {
  // Returns all users in a project with their roles attached
  GetUsers: async (projectPath) => {
    if (!projectPath) return [];
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const userRows = query(db, 'SELECT * FROM user');
      const roleRows = query(db, 'SELECT * FROM role');
      
      // Build role lookup map
      const roleMap = {};
      for (const role of roleRows) {
        roleMap[role.id] = rowToRole(role);
      }
      
      // Attach role object to each user
      return userRows.map(row => {
        const user = rowToUser(row);
        user.role = roleMap[user.role_id] || null;
        return user;
      });
    } catch (error) {
      console.error('GetUsers error:', error);
      return [];
    }
  },

  // Returns a specific user by ID with role attached
  GetUser: async (projectPath, userId) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const row = queryOne(db, 'SELECT * FROM user WHERE id = ?', [userId]);
      if (!row) return {};
      
      const user = rowToUser(row);
      
      // Attach role object
      if (user.role_id) {
        const roleRow = queryOne(db, 'SELECT * FROM role WHERE id = ?', [user.role_id]);
        user.role = roleRow ? rowToRole(roleRow) : null;
      }
      
      return user;
    } catch (error) {
      console.error('GetUser error:', error);
      return {};
    }
  },

  // Updates a user's information
  UpdateUser: async (projectPath, user) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/user/${user.id}`, 'PUT', user);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        UPDATE user SET username = ?, email = ?, first_name = ?, last_name = ?, role_id = ?, mtime = ?
        WHERE id = ?
      `, [user.username, user.email, user.first_name, user.last_name, user.role_id, Date.now(), user.id]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UpdateUser local update error:', error);
    }
    
    return result;
  },

  // Returns all roles in a project
  GetRoles: async (projectPath) => {
    const projectName = getProjectName(projectPath);
    try {
      const db = await getDatabase(projectName);
      const rows = query(db, 'SELECT * FROM role');
      return rows.map(rowToRole);
    } catch (error) {
      console.error('GetRoles error:', error);
      return [];
    }
  },

  // Creates a new role in a project
  CreateRole: async (projectPath, role) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/role`, 'POST', role);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        INSERT INTO role (id, mtime, name, view_asset, create_asset, edit_asset, delete_asset,
                         create_collection, edit_collection, delete_collection, create_checkpoint,
                         revert_checkpoint, delete_checkpoint, download_checkpoint,
                         change_assignee, create_tag, delete_tag, create_status, delete_status,
                         create_type, delete_type, synced)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)
      `, [
        result.id, Date.now(), result.name,
        result.view_asset ? 1 : 0, result.create_asset ? 1 : 0, result.edit_asset ? 1 : 0, result.delete_asset ? 1 : 0,
        result.create_collection ? 1 : 0, result.edit_collection ? 1 : 0, result.delete_collection ? 1 : 0,
        result.create_checkpoint ? 1 : 0, result.revert_checkpoint ? 1 : 0, result.delete_checkpoint ? 1 : 0,
        result.download_checkpoint ? 1 : 0, result.change_assignee ? 1 : 0, result.create_tag ? 1 : 0,
        result.delete_tag ? 1 : 0, result.create_status ? 1 : 0, result.delete_status ? 1 : 0,
        result.create_type ? 1 : 0, result.delete_type ? 1 : 0
      ]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('CreateRole local insert error:', error);
    }
    
    return result;
  },

  // Updates an existing role
  UpdateRole: async (projectPath, role) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    const result = await studioApiCall(studioUrl, `/${projectName}/role/${role.id}`, 'PUT', role);
    
    try {
      const db = await getDatabase(projectName);
      execute(db, `
        UPDATE role SET name = ?, view_asset = ?, create_asset = ?, edit_asset = ?, delete_asset = ?,
                       create_collection = ?, edit_collection = ?, delete_collection = ?, create_checkpoint = ?,
                       revert_checkpoint = ?, delete_checkpoint = ?, download_checkpoint = ?,
                       change_assignee = ?, create_tag = ?, delete_tag = ?, create_status = ?,
                       delete_status = ?, create_type = ?, delete_type = ?, mtime = ?
        WHERE id = ?
      `, [
        role.name, role.view_asset ? 1 : 0, role.create_asset ? 1 : 0, role.edit_asset ? 1 : 0, role.delete_asset ? 1 : 0,
        role.create_collection ? 1 : 0, role.edit_collection ? 1 : 0, role.delete_collection ? 1 : 0,
        role.create_checkpoint ? 1 : 0, role.revert_checkpoint ? 1 : 0, role.delete_checkpoint ? 1 : 0,
        role.download_checkpoint ? 1 : 0, role.change_assignee ? 1 : 0, role.create_tag ? 1 : 0,
        role.delete_tag ? 1 : 0, role.create_status ? 1 : 0, role.delete_status ? 1 : 0,
        role.create_type ? 1 : 0, role.delete_type ? 1 : 0, Date.now(), role.id
      ]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('UpdateRole local update error:', error);
    }
    
    return result;
  },

  // Deletes a role from a project
  DeleteRole: async (projectPath, roleId) => {
    const projectName = getProjectName(projectPath);
    const studioUrl = getActiveStudioUrl();
    
    await studioApiCall(studioUrl, `/${projectName}/role/${roleId}`, 'DELETE');
    
    try {
      const db = await getDatabase(projectName);
      execute(db, 'DELETE FROM role WHERE id = ?', [roleId]);
      await persistDatabase(projectName);
    } catch (error) {
      console.error('DeleteRole local update error:', error);
    }
  },

  // Fetches a user by ID from the global authentication server
  FetchUserById: async (userId) => {
    const studioUrl = getActiveStudioUrl();
    
    try {
      const user = await studioApiCall(studioUrl, `/persons/${userId}`, 'GET');
      return user;
    } catch (error) {
      console.error('FetchUserById error:', error);
      return null;
    }
  },
};
