// =============================================================================
// PROFILE SERVICE
// =============================================================================

import { globalApiCall } from './http-client.js';

export const ProfileService = {
  // Get complete user profile
  GetUserProfile: async (userId) => {
    const response = await globalApiCall(`/api/users/${userId}/profile`, 'GET');
    return response || {};
  },
  
  // Update user profile fields
  UpdateUserProfile: async (userId, profileData) => {
    return await globalApiCall(`/api/users/${userId}/profile`, 'PUT', profileData);
  },
  
  // Update user photo (base64)
  UpdateUserPhoto: async (photoBase64) => {
    return await globalApiCall('/person/photo', 'POST', { photo: photoBase64 });
  },
  
  // User Tools
  GetUserTools: async (userId) => {
    const tools = await globalApiCall(`/api/users/${userId}/tools`, 'GET');
    return Array.isArray(tools) ? tools : [];
  },
  
  AddUserTool: async (userId, toolData) => {
    return await globalApiCall(`/api/users/${userId}/tools`, 'POST', toolData);
  },
  
  UpdateUserTool: async (userId, toolId, proficiencyLevel) => {
    return await globalApiCall(`/api/users/${userId}/tools/${toolId}`, 'PUT', { proficiency_level: proficiencyLevel });
  },
  
  RemoveUserTool: async (userId, toolId) => {
    return await globalApiCall(`/api/users/${userId}/tools/${toolId}`, 'DELETE');
  },
  
  // User Skills
  GetUserSkills: async (userId) => {
    const skills = await globalApiCall(`/api/users/${userId}/skills`, 'GET');
    return Array.isArray(skills) ? skills : [];
  },
  
  AddUserSkill: async (userId, skillData) => {
    return await globalApiCall(`/api/users/${userId}/skills`, 'POST', skillData);
  },
  
  UpdateUserSkill: async (userId, skillId, proficiencyLevel) => {
    return await globalApiCall(`/api/users/${userId}/skills/${skillId}`, 'PUT', { proficiency_level: proficiencyLevel });
  },
  
  RemoveUserSkill: async (userId, skillId) => {
    return await globalApiCall(`/api/users/${userId}/skills/${skillId}`, 'DELETE');
  },
  
  // Reference data - all available tools and skills
  GetAllTools: async () => {
    const tools = await globalApiCall('/api/tools', 'GET');
    return Array.isArray(tools) ? tools : [];
  },
  
  GetToolsByCategory: async (category) => {
    const tools = await globalApiCall(`/api/tools/category/${encodeURIComponent(category)}`, 'GET');
    return Array.isArray(tools) ? tools : [];
  },
  
  GetAllSkills: async () => {
    const skills = await globalApiCall('/api/skills', 'GET');
    return Array.isArray(skills) ? skills : [];
  },
  
  GetSkillsByCategory: async (category) => {
    const skills = await globalApiCall(`/api/skills/category/${encodeURIComponent(category)}`, 'GET');
    return Array.isArray(skills) ? skills : [];
  },
  
  // Countries and genders for profile
  GetAllCountries: async () => {
    const countries = await globalApiCall('/api/countries', 'GET');
    return Array.isArray(countries) ? countries : [];
  },
  
  GetAllGenders: async () => {
    const genders = await globalApiCall('/api/genders', 'GET');
    return Array.isArray(genders) ? genders : [];
  },
  
  // Legacy aliases for backward compatibility
  GetProfile: async (userId) => ProfileService.GetUserProfile(userId),
  UpdateProfile: async (userId, profileData) => ProfileService.UpdateUserProfile(userId, profileData),
  DeleteUserTool: async (userId, toolId) => ProfileService.RemoveUserTool(userId, toolId),
  DeleteUserSkill: async (userId, skillId) => ProfileService.RemoveUserSkill(userId, skillId),
};
