import { globalApiCall } from './http-client.js';

export const DiscoveryService = {
  /**
   * Discover users with optional filters and pagination
   * @param {Object} params - Query parameters
   * @param {number} [params.page=1] - Page number
   * @param {number} [params.limit=20] - Items per page (max 100)
   * @param {string} [params.q] - Text search (name, username, bio)
   * @param {string} [params.tool] - Filter by tool name
   * @param {string} [params.skill] - Filter by skill name
   * @param {string} [params.country] - Filter by country code (e.g., "US", "GB")
   * @param {string} [params.availability] - Filter: "available", "busy", "not_looking"
   * @param {string} [params.job_title] - Filter by job title (partial match)
   * @param {string} [params.sort="created_at"] - Sort field: "created_at", "name", "experience"
   * @param {string} [params.order="desc"] - Sort order: "asc", "desc"
   * @returns {Promise<{users: Array, pagination: Object, filters_applied: Object}>}
   */
  DiscoverUsers: async (params = {}) => {
    const queryParams = new URLSearchParams();
    
    // Add all non-empty params to query string
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        queryParams.append(key, value);
      }
    });
    
    const queryString = queryParams.toString();
    const endpoint = `/api/discover/users${queryString ? '?' + queryString : ''}`;
    
    const response = await globalApiCall(endpoint, 'GET');
    return response || { users: [], pagination: { page: 1, limit: 20, total: 0, total_pages: 0 }, filters_applied: {} };
  },

  /**
   * Get all available tools for filter dropdown
   * @returns {Promise<Array>}
   */
  GetAllTools: async () => {
    const tools = await globalApiCall('/api/tools', 'GET');
    return Array.isArray(tools) ? tools : [];
  },

  /**
   * Get all available skills for filter dropdown
   * @returns {Promise<Array>}
   */
  GetAllSkills: async () => {
    const skills = await globalApiCall('/api/skills', 'GET');
    return Array.isArray(skills) ? skills : [];
  },

  /**
   * Get all countries for filter dropdown
   * @returns {Promise<Array>}
   */
  GetAllCountries: async () => {
    const countries = await globalApiCall('/api/countries', 'GET');
    return Array.isArray(countries) ? countries : [];
  },

  /**
   * Get user statistics
   * @returns {Promise<Object>}
   */
  GetUserStats: async () => {
    const stats = await globalApiCall('/api/search/users/stats', 'GET');
    return stats || {};
  },
};
