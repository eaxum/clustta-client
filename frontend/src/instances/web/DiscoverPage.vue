<template>
  <div class="discover-root">
    <!-- Header -->
    <div class="discover-header">
      <NavigationBar />
    </div>

    <div class="discover-body">
      <!-- Page Title & Search -->
      <div class="discover-title-section">
        <h1 class="discover-title">Discover Professionals</h1>
        <p class="discover-subtitle">Find talented artists, animators, and VFX professionals</p>
        
        <!-- Search Bar -->
        <div class="search-container">
          <SearchBar
            v-model="searchQuery"
            placeholder="Search by name, username, or bio..."
            :isLoading="loading"
            size="large"
            @clear="clearSearch"
            @keyup.enter="applySearch"
          />
        </div>
      </div>

      <!-- Filters Section -->
      <div class="filters-section">
        <div class="filter-row">
          <!-- Availability Filter -->
          <div class="filter-group">
            <label class="filter-label">Availability</label>
            <DropDownBox
              :items="availabilityOptions"
              :selectedItem="filters.availability"
              :onSelect="(val) => handleFilterSelect('availability', val)"
              placeHolder="Any"
              :fullWidth="true"
            />
          </div>

          <!-- Tools Filter -->
          <div class="filter-group">
            <label class="filter-label">Tool</label>
            <DropDownBox
              :items="toolOptions"
              :selectedItem="filters.tool"
              :onSelect="(val) => handleFilterSelect('tool', val)"
              placeHolder="All Tools"
              :fullWidth="true"
            />
          </div>

          <!-- Skills Filter -->
          <div class="filter-group">
            <label class="filter-label">Skill</label>
            <DropDownBox
              :items="skillOptions"
              :selectedItem="filters.skill"
              :onSelect="(val) => handleFilterSelect('skill', val)"
              placeHolder="All Skills"
              :fullWidth="true"
            />
          </div>

          <!-- Country Filter -->
          <div class="filter-group">
            <label class="filter-label">Location</label>
            <DropDownBox
              :items="countryOptions"
              :selectedItem="filters.country"
              :onSelect="(val) => handleFilterSelect('country', val)"
              placeHolder="All Locations"
              :fullWidth="true"
            />
          </div>
        </div>

        <!-- Active Filters -->
        <div v-if="hasActiveFilters" class="active-filters">
          <span class="active-filters-label">Active filters:</span>
          <Chip
            v-if="filters.availability"
            :icon="getAppIcon('clock')"
            :label="filters.availability"
            :onRemove="() => clearFilter('availability')"
          />
          <Chip
            v-if="filters.tool"
            :icon="getToolIconPath(filters.tool)"
            :label="filters.tool"
            :onRemove="() => clearFilter('tool')"
            :useImage="true"
          />
          <Chip
            v-if="filters.skill"
            :icon="getSkillIconPath(filters.skill)"
            :label="filters.skill"
            :onRemove="() => clearFilter('skill')"
          />
          <Chip
            v-if="filters.country"
            :icon="getAppIcon('map-pin')"
            :label="getCountryName(filters.country)"
            :onRemove="() => clearFilter('country')"
          />
          <Chip
            v-if="searchQuery"
            :icon="getAppIcon('search')"
            :label="`&quot;${searchQuery}&quot;`"
            :onRemove="clearSearch"
          />
          <ActionButton
            :label="'Clear All'"
            :buttonFunction="clearAllFilters"
            :showIcon="false"
            :showLabel="true"
          />
        </div>
      </div>

      <!-- Results Section -->
      <div class="results-section">
        <!-- Results Count -->
        <div class="results-header">
          <span class="results-count">
            {{ pagination.total }} professional{{ pagination.total !== 1 ? 's' : '' }} found
          </span>
          <div class="results-controls">
            <!-- Show Details Toggle -->
            <ActionButton
              :icon="getAppIcon(showDetails ? 'eye-cancel' : 'eye')"
              :label="showDetails ? 'Hide Details' : 'Show Details'"
              :buttonFunction="() => showDetails = !showDetails"
              :showIcon="true"
              :showLabel="true"
              :useOutline="true"
            />
            <div class="sort-controls">
              <label class="sort-label">Sort by:</label>
              <DropDownBox
                :items="sortOptions"
                :selectedItem="sortBy"
                :onSelect="handleSortSelect"
                placeHolder="Newest"
                :fullWidth="false"
                :fixedWidth="true"
              />
            </div>
          </div>
        </div>

        <!-- Loading State -->
        <div v-if="loading" class="loading-container">
          <div class="loading-spinner"></div>
          <p>Loading...</p>
        </div>

        <!-- Error State -->
        <div v-else-if="error" class="error-container">
          <img :src="getAppIcon('alert-circle')" alt="Error" class="error-icon" />
          <h2>Error Loading Results</h2>
          <p>{{ errorMessage }}</p>
          <ActionButton
            :label="'Try Again'"
            :buttonFunction="loadUsers"
            :useOutline="true"
            :showIcon="false"
            :showLabel="true"
          />
        </div>

        <!-- Empty State -->
        <div v-else-if="users.length === 0" class="empty-container">
          <img :src="getAppIcon('users')" alt="No users" class="empty-icon" />
          <h2>No Professionals Found</h2>
          <p>Try adjusting your filters or search query</p>
          <ActionButton
            v-if="hasActiveFilters"
            :label="'Clear Filters'"
            :buttonFunction="clearAllFilters"
            :useOutline="true"
            :showIcon="false"
            :showLabel="true"
          />
        </div>

        <!-- User List -->
        <div v-else class="users-list">
          <UserListCard 
            v-for="user in users" 
            :key="user.id" 
            :user="user"
            :showDetails="showDetails"
            @click="viewProfile(user.username)"
          />
        </div>

        <!-- Pagination -->
        <div v-if="!loading && users.length > 0" class="pagination-section">
          <div class="pagination-info">
            Showing {{ paginationRange }} of {{ pagination.total }}
          </div>
          <div class="pagination-controls">
            <ActionButton
              :label="'Previous'"
              :icon="getAppIcon('chevron-left')"
              :buttonFunction="() => goToPage(pagination.page - 1)"
              :isDisabled="!pagination.has_prev"
              :useOutline="true"
              :showLabel="true"
            />
            <span class="pagination-current">Page {{ pagination.page }} of {{ pagination.total_pages }}</span>
            <ActionButton
              :label="'Next'"
              :icon="getAppIcon('chevron-right')"
              :buttonFunction="() => goToPage(pagination.page + 1)"
              :isDisabled="!pagination.has_next"
              :useOutline="true"
              :showLabel="true"
              :iconAfter="true"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useIconStore } from '@/stores/icons';
import { DiscoveryService } from '@/services/adapters/discoveryservice.js';
import { getToolLogo, getSkillIcon } from '@/utils/iconMappers';

// Components
import NavigationBar from '@/instances/web/components/NavigationBar.vue';
import UserListCard from './components/UserListCard.vue';
import SearchBar from '@/instances/desktop/components/SearchBar.vue';
import DropDownBox from '@/instances/common/components/DropDownBox.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import Chip from '@/instances/common/components/Chip.vue';

const route = useRoute();
const router = useRouter();
const iconStore = useIconStore();

// State
const loading = ref(true);
const error = ref(false);
const errorMessage = ref('');
const users = ref([]);
const searchQuery = ref('');
const sortBy = ref('added_at');
const showDetails = ref(false);

const filters = ref({
  availability: '',
  tool: '',
  skill: '',
  country: ''
});

const pagination = ref({
  page: 1,
  limit: 20,
  total: 0,
  total_pages: 0,
  has_next: false,
  has_prev: false
});

// Available filter options
const availableTools = ref([]);
const availableSkills = ref([]);
const availableCountries = ref([]);

// Computed
const hasActiveFilters = computed(() => {
  return filters.value.availability || 
         filters.value.tool || 
         filters.value.skill || 
         filters.value.country ||
         searchQuery.value;
});

const paginationRange = computed(() => {
  const start = (pagination.value.page - 1) * pagination.value.limit + 1;
  const end = Math.min(pagination.value.page * pagination.value.limit, pagination.value.total);
  return `${start}-${end}`;
});

// Dropdown options for DropDownBox components
const availabilityOptions = computed(() => [
  { name: '', label: 'Any' },
  { name: 'available', label: 'Available' },
  { name: 'busy', label: 'Busy' },
  { name: 'not_looking', label: 'Not Looking' }
]);

const toolOptions = computed(() => {
  return [
    { name: '', label: 'All Tools' },
    ...availableTools.value.map(t => ({ name: t.name, label: t.name }))
  ];
});

const skillOptions = computed(() => {
  return [
    { name: '', label: 'All Skills' },
    ...availableSkills.value.map(s => ({ name: s.name, label: s.name }))
  ];
});

const countryOptions = computed(() => {
  return [
    { name: '', label: 'All Locations' },
    ...availableCountries.value.map(c => ({ name: c.code, label: c.name }))
  ];
});

const sortOptions = computed(() => [
  { name: 'added_at', label: 'Newest' },
  { name: 'name', label: 'Name' },
  { name: 'experience', label: 'Experience' }
]);

// Methods
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

const getToolIconPath = (toolName) => {
  return getToolLogo(toolName);
};

const getSkillIconPath = (skillName) => {
  const skill = availableSkills.value.find(s => s.name === skillName);
  const category = skill?.category || '';
  const iconName = getSkillIcon(skillName, category);
  return iconStore.getAppIcon(iconName);
};

const getCountryName = (code) => {
  const country = availableCountries.value.find(c => c.code === code);
  return country?.name || code;
};

const loadFilterOptions = async () => {
  try {
    const [tools, skills, countries] = await Promise.all([
      DiscoveryService.GetAllTools(),
      DiscoveryService.GetAllSkills(),
      DiscoveryService.GetAllCountries()
    ]);
    availableTools.value = tools;
    availableSkills.value = skills;
    availableCountries.value = countries;
  } catch (err) {
    console.error('Error loading filter options:', err);
  }
};

const loadUsers = async () => {
  loading.value = true;
  error.value = false;
  
  try {
    const params = {
      page: pagination.value.page,
      limit: pagination.value.limit,
      sort: sortBy.value,
      order: sortBy.value === 'name' ? 'asc' : 'desc'
    };

    if (searchQuery.value) params.q = searchQuery.value;
    if (filters.value.availability) params.availability = filters.value.availability;
    if (filters.value.tool) params.tool = filters.value.tool;
    if (filters.value.skill) params.skill = filters.value.skill;
    if (filters.value.country) params.country = filters.value.country;

    const result = await DiscoveryService.DiscoverUsers(params);
    
    users.value = result.users || [];
    pagination.value = {
      page: result.pagination?.page || 1,
      limit: result.pagination?.limit || 20,
      total: result.pagination?.total || 0,
      total_pages: result.pagination?.total_pages || 0,
      has_next: result.pagination?.has_next || false,
      has_prev: result.pagination?.has_prev || false
    };
  } catch (err) {
    console.error('Error loading users:', err);
    error.value = true;
    errorMessage.value = err.message || 'Failed to load users. Please try again.';
  } finally {
    loading.value = false;
  }
};

const applySearch = () => {
  pagination.value.page = 1;
  updateUrlParams();
  loadUsers();
};

const applyFilters = () => {
  pagination.value.page = 1;
  updateUrlParams();
  loadUsers();
};

// Handler for filter dropdown selections
const handleFilterSelect = (filterKey, value) => {
  filters.value[filterKey] = value;
  applyFilters();
};

// Handler for sort dropdown
const handleSortSelect = (value) => {
  sortBy.value = value;
  applyFilters();
};

const clearFilter = (filterKey) => {
  filters.value[filterKey] = '';
  applyFilters();
};

const clearSearch = () => {
  searchQuery.value = '';
  applyFilters();
};

const clearAllFilters = () => {
  filters.value = {
    availability: '',
    tool: '',
    skill: '',
    country: ''
  };
  searchQuery.value = '';
  applyFilters();
};

const goToPage = (page) => {
  if (page < 1 || page > pagination.value.total_pages) return;
  pagination.value.page = page;
  updateUrlParams();
  loadUsers();
  // Scroll to top of results
  document.querySelector('.results-section')?.scrollIntoView({ behavior: 'smooth' });
};

const viewProfile = (username) => {
  router.push({
    name: 'public-profile',
    params: { username }
  });
};

const updateUrlParams = () => {
  const query = {};
  if (searchQuery.value) query.q = searchQuery.value;
  if (filters.value.availability) query.availability = filters.value.availability;
  if (filters.value.tool) query.tool = filters.value.tool;
  if (filters.value.skill) query.skill = filters.value.skill;
  if (filters.value.country) query.country = filters.value.country;
  if (pagination.value.page > 1) query.page = pagination.value.page;
  if (sortBy.value !== 'added_at') query.sort = sortBy.value;
  
  router.replace({ query });
};

const loadFromUrlParams = () => {
  const query = route.query;
  searchQuery.value = query.q || '';
  filters.value.availability = query.availability || '';
  filters.value.tool = query.tool || '';
  filters.value.skill = query.skill || '';
  filters.value.country = query.country || '';
  pagination.value.page = parseInt(query.page) || 1;
  sortBy.value = query.sort || 'added_at';
};

// Initialize
onMounted(async () => {
  loadFromUrlParams();
  await loadFilterOptions();
  await loadUsers();
});

// Watch for route changes (back/forward navigation)
watch(() => route.query, () => {
  loadFromUrlParams();
  loadUsers();
}, { deep: true });
</script>

<style scoped>
@import "@/assets/desktop.css";

.discover-root {
  width: 100%;
  min-height: 100vh;
  max-height: 100vh;
  background-color: var(--midnight-steel);
  display: flex;
  flex-direction: column;
  align-items: center;
  overflow: hidden;
  overflow-y: auto;
  color: var(--white);
}

.discover-header {
  width: 100%;
  display: flex;
  align-items: center;
  background-color: rgba(255, 255, 255, 0.05);
  position: sticky;
  top: 0;
  z-index: 99999;
  backdrop-filter: blur(30px);
}

.discover-body {
  flex: 1;
  width: 100%;
  max-width: 1000px;
  padding: 2rem;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

/* Title Section */
.discover-title-section {
  text-align: center;
  margin-bottom: 1rem;
  /* background-color: crimson; */
  width: 100%;
}

.discover-title {
  font-family: 'Bricolage Grotesque', sans-serif;
  font-size: 3.5rem;
  font-weight: 700;
  color: var(--white);
  margin: 0 0 0.5rem 0;
}

.discover-subtitle {
  font-size: 1rem;
  color: var(--white);
  margin: 0 0 1.5rem 0;
}

/* Search */
.search-container {
    /* background-color: forestgreen; */
    width: 90%;
  max-width: 90%;
  margin: 0 auto;
}

/* Filters */
.filters-section {
  background-color: var(--black-steel);
  box-sizing: border-box;
  border-radius: var(--large-radius);
  padding: 1rem 1.25rem;
  outline: var(--transparent-line);
  outline-offset: -1px;
  width: 100%;
}

.filter-row {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.filter-group {
  flex: 1;
  min-width: 150px;
}

.filter-label {
  display: block;
  font-size: 0.75rem;
  color: var(--text-tertiary);
  margin-bottom: 0.35rem;
}

.active-filters {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: var(--transparent-line);
  flex-wrap: wrap;
}

.active-filters-label {
  font-size: 0.8rem;
  color: var(--text-tertiary);
}

/* Results */
.results-section {
  flex: 1;
  width: 100%;
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.results-count {
  font-size: 0.9rem;
  color: var(--white);
}

.results-controls {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.sort-controls {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.sort-label {
  font-size: 0.8rem;
  text-wrap: nowrap;
  color: var(--text-tertiary);
}

/* Users List */
.users-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

/* Loading, Error, Empty States */
.loading-container,
.error-container,
.empty-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
  color: var(--white);
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-top-color: var(--white);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-icon,
.empty-icon {
  width: 64px;
  height: 64px;
  opacity: 0.3;
  margin-bottom: 1rem;
}

.error-container h2,
.empty-container h2 {
  font-size: 1.25rem;
  font-weight: 500;
  color: var(--white);
  margin: 0 0 0.5rem 0;
}

.error-container p,
.empty-container p {
  margin: 0 0 1rem 0;
}

/* Pagination */
.pagination-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: var(--transparent-line);
}

.pagination-info {
  font-size: 0.85rem;
  color: var(--text-tertiary);
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.pagination-current {
  font-size: 0.85rem;
  color: var(--white);
}

/* Responsive */
@media (max-width: 768px) {
  .discover-body {
    padding: 1rem;
  }

  .discover-title {
    font-size: 1.5rem;
  }

  .filter-row {
    flex-direction: column;
  }

  .filter-group {
    min-width: 100%;
  }

  .results-header {
    flex-direction: column;
    gap: 0.75rem;
    align-items: flex-start;
  }

  .results-controls {
    width: 100%;
    justify-content: space-between;
  }

  .pagination-section {
    flex-direction: column;
    gap: 1rem;
  }
}

/* Scrollbar */
.discover-root::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

.discover-root::-webkit-scrollbar-track {
  background: var(--midnight-steel);
  border-radius: 5px;
}

.discover-root::-webkit-scrollbar-thumb {
  background-color: rgba(255, 255, 255, 0.15);
  background: var(--steel);
  border-radius: 5px;
  transition: background-color 0.2s;
}

.discover-root::-webkit-scrollbar-thumb:hover {
  background-color: rgba(255, 255, 255, 0.3);
}

.discover-root::-webkit-scrollbar-thumb:active {
  background-color: rgba(255, 255, 255, 0.4);
}

.discover-root::-webkit-scrollbar-corner {
  background: var(--black);
}
</style>
