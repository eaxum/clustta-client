import { globalApiCall } from './http-client.js';

export const ProfileService = {
  // Returns complete user profile by ID
  GetUserProfile: async (userId) => {
    const response = await globalApiCall(`/api/users/${userId}/profile`, 'GET');
    return response || {};
  },

  // Returns public profile by username or user ID (for public profile pages)
  GetPublicProfile: async (identifier) => {
    const response = await globalApiCall(`/api/profiles/public/${encodeURIComponent(identifier)}`, 'GET');
    return response || null;
  },

  // Updates user profile fields
  UpdateUserProfile: async (userId, profileData) => {
    return await globalApiCall(`/api/users/${userId}/profile`, 'PUT', profileData);
  },

  // Updates user profile photo with base64 data
  UpdateUserPhoto: async (photoBase64) => {
    return await globalApiCall('/person/photo', 'POST', { photo: photoBase64 });
  },

  // Returns all tools for a user
  GetUserTools: async (userId) => {
    const tools = await globalApiCall(`/api/users/${userId}/tools`, 'GET');
    return Array.isArray(tools) ? tools : [];
  },

  // Adds a tool to user profile
  AddUserTool: async (userId, toolData) => {
    return await globalApiCall(`/api/users/${userId}/tools`, 'POST', toolData);
  },

  // Updates user's proficiency level for a tool
  UpdateUserTool: async (userId, toolId, proficiencyLevel) => {
    return await globalApiCall(`/api/users/${userId}/tools/${toolId}`, 'PUT', { proficiency_level: proficiencyLevel });
  },

  // Removes a tool from user profile
  RemoveUserTool: async (userId, toolId) => {
    return await globalApiCall(`/api/users/${userId}/tools/${toolId}`, 'DELETE');
  },

  // Returns all skills for a user
  GetUserSkills: async (userId) => {
    const skills = await globalApiCall(`/api/users/${userId}/skills`, 'GET');
    return Array.isArray(skills) ? skills : [];
  },

  // Adds a skill to user profile
  AddUserSkill: async (userId, skillData) => {
    return await globalApiCall(`/api/users/${userId}/skills`, 'POST', skillData);
  },

  // Updates user's proficiency level for a skill
  UpdateUserSkill: async (userId, skillId, proficiencyLevel) => {
    return await globalApiCall(`/api/users/${userId}/skills/${skillId}`, 'PUT', { proficiency_level: proficiencyLevel });
  },

  // Removes a skill from user profile
  RemoveUserSkill: async (userId, skillId) => {
    return await globalApiCall(`/api/users/${userId}/skills/${skillId}`, 'DELETE');
  },

  // Returns all available tools in the system
  GetAllTools: async () => {
    const tools = await globalApiCall('/api/tools', 'GET');
    return Array.isArray(tools) ? tools : [];
  },

  // Returns tools filtered by category
  GetToolsByCategory: async (category) => {
    const tools = await globalApiCall(`/api/tools/category/${encodeURIComponent(category)}`, 'GET');
    return Array.isArray(tools) ? tools : [];
  },

  // Returns all available skills in the system
  GetAllSkills: async () => {
    const skills = await globalApiCall('/api/skills', 'GET');
    return Array.isArray(skills) ? skills : [];
  },

  // Returns skills filtered by category
  GetSkillsByCategory: async (category) => {
    const skills = await globalApiCall(`/api/skills/category/${encodeURIComponent(category)}`, 'GET');
    return Array.isArray(skills) ? skills : [];
  },

  // Returns all countries for profile selection
  GetAllCountries: async () => {
    const countries = await globalApiCall('/api/countries', 'GET');
    return Array.isArray(countries) ? countries : [];
  },

  // Returns all genders for profile selection
  GetAllGenders: async () => {
    const genders = await globalApiCall('/api/genders', 'GET');
    return Array.isArray(genders) ? genders : [];
  },

  // Legacy alias for GetUserProfile
  GetProfile: async (userId) => ProfileService.GetUserProfile(userId),

  // Legacy alias for UpdateUserProfile
  UpdateProfile: async (userId, profileData) => ProfileService.UpdateUserProfile(userId, profileData),

  // Legacy alias for RemoveUserTool
  DeleteUserTool: async (userId, toolId) => ProfileService.RemoveUserTool(userId, toolId),

  // Legacy alias for RemoveUserSkill
  DeleteUserSkill: async (userId, skillId) => ProfileService.RemoveUserSkill(userId, skillId),
};
