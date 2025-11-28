import { defineStore } from "pinia";

export const useProfileStore = defineStore("profile", {
  state: () => ({
    // Profile data
    profile: {
      // Basic info (duplicated from userStore for independence)
      id: null,
      first_name: '',
      last_name: '',
      username: '',
      email: '',
      photo: null,
      
      // Extended profile data
      title: '',
      bio: '',
      country: '',
      gender_id: null,
      date_of_birth: null,
      
      // Settings
      availability: 'available', // 'available' | 'busy' | 'not_looking'
      profile_visibility: 'public', // 'public' | 'private' | 'collaborators_only'
      
      // Collections
      tools: [],
      skills: [],
      links: {},
      studios: [],
    },
    
    // Meta
    isLoaded: false,
    lastFetched: null,
    isFetching: false,
  }),
  
  getters: {
    // Profile getters
    getProfile: (state) => state.profile,
    
    getFullName: (state) => {
      return `${state.profile.first_name} ${state.profile.last_name}`.trim() || 'User';
    },
    
    getProfileVisibility: (state) => state.profile.profile_visibility,
    
    getAvailability: (state) => state.profile.availability,
    
    getTools: (state) => state.profile.tools,
    
    getSkills: (state) => state.profile.skills,
    
    getLinks: (state) => state.profile.links,
    
    getStudios: (state) => state.profile.studios,
    
    isProfileLoaded: (state) => state.isLoaded,
    
    isFetchingProfile: (state) => state.isFetching,
    
    // Check if profile data is stale (older than 5 minutes)
    isProfileStale: (state) => {
      if (!state.lastFetched) return true;
      const fiveMinutes = 5 * 60 * 1000;
      return Date.now() - state.lastFetched > fiveMinutes;
    },
  },
  
  actions: {
    // Set profile data
    setProfile(profileData) {
      
      // Transform backend link fields into links object
      const links = {
        behance: profileData.behance_link || profileData.BehanceLink || '',
        artstation: profileData.artstation_link || profileData.ArtstationLink || '',
        portfolio: profileData.portfolio_link || profileData.PortfolioLink || '',
        linkedin: profileData.linkedin_link || profileData.LinkedInLink || '',
        instagram: profileData.instagram_link || profileData.InstagramLink || '',
      };
      
      const tools = profileData.tools;

      const skills = profileData.skills;
    console.log(profileData)
      this.profile = {
        ...this.profile,
        ...profileData,
        // Transform backend field names to frontend field names
        title: profileData.job_title || profileData.JobTitle || profileData.title || '',
        country: profileData.location || profileData.Location || profileData.country || '',
        links, // Override with transformed links object
        tools, // Override with transformed tools
        skills, // Override with transformed skills
      };
      this.isLoaded = true;
      this.lastFetched = Date.now();
    },
    
    // Update specific profile fields
    updateProfileFields(fields) {
      this.profile = {
        ...this.profile,
        ...fields,
      };
    },
    
    // Set loading state
    setFetching(isFetching) {
      this.isFetching = isFetching;
    },
    
    // Update profile visibility
    setProfileVisibility(visibility) {
      this.profile.profile_visibility = visibility;
    },
    
    // Update availability
    setAvailability(availability) {
      this.profile.availability = availability;
    },
    
    // Tools management
    setTools(tools) {
      this.profile.tools = tools;
    },
    
    addTool(tool) {
      this.profile.tools.push(tool);
    },
    
    removeTool(toolId) {
      this.profile.tools = this.profile.tools.filter(t => t.id !== toolId);
    },
    
    updateTool(toolId, updatedTool) {
      const index = this.profile.tools.findIndex(t => t.id === toolId);
      if (index !== -1) {
        this.profile.tools[index] = { ...this.profile.tools[index], ...updatedTool };
      }
    },
    
    // Skills management
    setSkills(skills) {
      this.profile.skills = skills;
    },
    
    addSkill(skill) {
      this.profile.skills.push(skill);
    },
    
    removeSkill(skillId) {
      this.profile.skills = this.profile.skills.filter(s => s.id !== skillId);
    },
    
    updateSkill(skillId, updatedSkill) {
      const index = this.profile.skills.findIndex(s => s.id === skillId);
      if (index !== -1) {
        this.profile.skills[index] = { ...this.profile.skills[index], ...updatedSkill };
      }
    },
    
    // Links management
    setLinks(links) {
      this.profile.links = links;
    },
    
    updateLinks(links) {
      this.profile.links = { ...this.profile.links, ...links };
    },
    
    // Studios management
    setStudios(studios) {
      this.profile.studios = studios;
    },
    
    // Reset store
    $reset() {
      this.profile = {
        id: null,
        first_name: '',
        last_name: '',
        username: '',
        email: '',
        photo: null,
        title: '',
        bio: '',
        country: '',
        gender_id: null,
        date_of_birth: null,
        availability: 'available',
        profile_visibility: 'public',
        tools: [],
        skills: [],
        links: {},
        studios: [],
      };
      this.isLoaded = false;
      this.lastFetched = null;
      this.isFetching = false;
    },
  },
});
