<template>
  <div class="user-profile-root absolute-pane">
      <div ref="userProfileBody" class="user-profile-body">
        <div class="user-profile-container">
          
          <!-- Edit Controls -->
          <div class="edit-controls">
            <div class="profile-visibility-toggle">
              <span class="visibility-label">Profile Visibility</span>
              <ToggleSwitch 
                :switchValueProp="profileVisibility" 
                @click="toggleProfileVisibility"
              />
            </div>
            
            <ActionButton 
              :iconAfter="true" 
              :icon="getAppIcon('key')"
              label="Change Password" 
              @click="toggleSectionEdit('password')"
              :useOutline="true"
            />
          </div>

          <!-- Header Card -->
          <ProfileCard 
            title="Profile Information"
            :showEditButton="true"
            :isEditing="editingSections.header"
            @toggleEdit="toggleSectionEdit('header')"
          >
            <div class="header-layout">
              <ProfileAvatar
                :userPhoto="userPhoto"
                :avatarColor="profileColor"
                :isEditing="editingSections.header"
                @photoChanged="handlePhotoChange"
                @photoRemoved="removePhoto"
              />
              
              <div class="header-info">
                <div v-if="editingSections.header" class="edit-mode-fields">
                  <div class="form-row">
                    <FormInput
                      v-model="formData.first_name"
                      label="First Name"
                      :disabled="!isEditing"
                    />
                    <FormInput
                      v-model="formData.last_name"
                      label="Last Name"
                      :disabled="!isEditing"
                    />
                  </div>
                  
                  <!-- :disabled="!editingSections.header" -->
                  <FormInput
                    v-model="formData.username"
                    @input="checkUsername"
                    label="Username"
                    :disabled="!isEditing"
                    :showValidation="!!formData.username"
                    :error="errors.username"
                    :loading="checkingUsernameAvailability"
                    :valid="usernameValid && !isUsernameTaken"
                  />
                  
                  <FormInput
                    v-model="formData.email"
                    @input="checkEmail"
                    label="Email"
                    type="email"
                    :disabled="!isEditing"
                    :showValidation="!!formData.email"
                    :error="errors.email"
                    :loading="checkingEmailAvailability"
                    :valid="emailValid && !isEmailTaken"
                  />
                  
                  <!-- Professional Info in Edit Mode -->
                  <FormInput 
                    v-model="formData.bio"
                    label="Bio"
                    placeholder="e.g., 3D Artist & Animator"
                  />
                  <FormInput
                    v-model="formData.country"
                    label="Location"
                    placeholder="e.g., Portugal"
                  />
                  <div class="availability-field">
                    <label class="form-label">Availability</label>
                    <ActionButton
                      @click="toggleAvailability"
                      :icon="getAppIcon('check-circle')"
                      :label="utils.capitalizeStr(formData.availability)"
                      :iconAfter="false"
                      :useBackground="true"
                    />
                  </div>
                  
                  <!-- Professional Links in Edit Mode -->
                  <div class="links-section">
                    <LinksManager
                      :links="formData.links"
                      :isEditing="editingSections.header"
                      @update:links="updateLinks"
                      @update:linksValid="handleLinksValidUpdate"
                    />
                  </div>
                </div>
                
                <div v-else class="display-mode-fields">
                  <div class="profile-name">{{ fullName }}</div>
                  <div v-if="formData.bio" class="profile-title">{{ formData.bio }}</div>
                  
                  <div class="meta-info">
                    <div v-if="formData.country" class="info-item">
                      <img class="info-icon small-icons" :src="getAppIcon('map-pin')" alt="">
                      <span>{{ formData.country }}</span>
                    </div>
                    
                    <div 
                      v-if="formData.availability" 
                      class="availability-badge"
                      :style="{ backgroundColor: formData.availability === 'available' ? '#35a32e' : 'rgba(255, 255, 255, 0.1)' }"
                    >
                      <img class="info-icon small-icons" :src="getAppIcon('check-circle')" alt="">
                      <span>{{ utils.capitalizeStr(formData.availability) }}</span>
                    </div>
                  </div>
                  
                  <!-- Professional Links in Display Mode -->
                  <div class="social-links">
                    <LinksManager
                      :links="formData.links"
                      :isEditing="editingSections.header"
                      @update:links="updateLinks"
                    />
                  </div>
                </div>
              </div>
            </div>
          </ProfileCard>

          <!-- Studios Card -->
          <ProfileCard v-if="formData.studios.length" title="Studios">
            <div class="studios-container">
              <div
                v-for="studio in formData.studios"
                :key="studio.id"
                class="studio-item"
                :title="studio.name"
              >
                <img 
                  v-if="studio.logo" 
                  :src="studio.logo" 
                  :alt="studio.name" 
                  class="studio-logo"
                  @error="handleStudioLogoError"
                />
                <img 
                  v-else 
                  :src="getAppIcon('stall')" 
                  alt="Studio" 
                  class="studio-logo studio-icon-default"
                />
                <span class="studio-name">{{ studio.name }}</span>
              </div>
            </div>
          </ProfileCard>

          <!-- Skills Card -->
          <ProfileCard 
            title="Skills"
            :showEditButton="true"
            :isEditing="editingSections.skills"
            @toggleEdit="toggleSectionEdit('skills')"
          >
            <SkillsManager
              :skills="formData.skills"
              :allSkills="allSkills"
              :isEditing="editingSections.skills"
            />
          </ProfileCard>

          <!-- Tools & Software Card -->
          <ProfileCard 
            title="Tools & Software"
            :showEditButton="true"
            :isEditing="editingSections.tools"
            @toggleEdit="toggleSectionEdit('tools')"
          >
            <ToolsManager
              :tools="formData.tools"
              :allTools="allTools"
              :isEditing="editingSections.tools"
            />
          </ProfileCard>

          <!-- Change Password Card -->
          <ProfileCard 
            ref="passwordCard"
            v-if="editingSections.password"
            title="Change Password"
            :showEditButton="true"
            :isEditing="editingSections.password"
            @toggleEdit="toggleSectionEdit('password')"
          >
            <div class="change-password-section">
              <FormInput
                v-model="passwordData.currentPassword"
                label="Current Password"
                type="password"
                placeholder="Enter current password"
              />
              <FormInput
                v-model="passwordData.newPassword"
                label="New Password"
                type="password"
                placeholder="Enter new password"
                :error="passwordErrors.newPassword"
              />
              <FormInput
                v-model="passwordData.confirmPassword"
                label="Confirm Password"
                type="password"
                placeholder="Confirm new password"
                :error="passwordErrors.confirmPassword"
              />
              <ActionButton
                :isDisabled="!isPasswordValid"
                :iconAfter="true"
                :icon="getAppIcon('check-circle')"
                label="Update Password"
                @click="handlePasswordUpdate"
                :useBackground="true"
              />
            </div>
          </ProfileCard>

          <!-- Activity Card -->
          <ProfileCard  title="Activity [Coming soon]">
            <ContributionGraph />
          </ProfileCard>

          <!-- Danger Zone -->
          <ProfileCard title="Danger Zone">
            <div class="danger-zone">
              <p class="danger-message">
                Once you delete your account, there is no going back. Please be certain.
              </p>
              <ActionButton
                :iconAfter="true"
                :icon="getAppIcon('trash')"
                label="Delete Account"
                @click="prepDeleteAccountModal()"
                :color="'crimson'"
                :useBackground="true"
              />
            </div>
          </ProfileCard>

        </div>
      </div>
      
      <!-- Floating Action Buttons -->
      <!-- Only show for sections that need explicit saving (header, password) -->
      <!-- Skills and tools are saved immediately via API -->
      <div v-if="needsSaveButton" class="floating-actions">
        <template v-if="isSavingChanges">
          <ActionButton 
            :isLoading="true"
            :icon="getAppIcon('loading')"
            label="Saving changes..."
            :useBackground="true"
          />
        </template>
        <template v-else>
          <ActionButton 
            :color="'crimson'" 
            :iconAfter="true" 
            :icon="getAppIcon('close-circle')" 
            label="Cancel"
            @click="cancelAllEdits"
          />
          <ActionButton 
            :isDisabled="!isDataValid"  
            :iconAfter="true" 
            :icon="getAppIcon('check-circle')"
            label="Save Changes" 
            @click="saveAllChanges"
            :useBackground="true"
          />
        </template>
      </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onBeforeMount, nextTick } from 'vue';
import { AuthService, ProfileService, FSService } from "@/../bindings/clustta/services";
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useIconStore } from '@/stores/icons';
import { useUserStore } from '@/stores/users';
import { useProfileStore } from '@/stores/profile';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useTrayStates } from '@/stores/TrayStates';
import { useStageStore } from '@/stores/stages';

// Components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';
import ProfileCard from '@/instances/desktop/components/ProfileCard.vue';
import FormInput from '@/instances/desktop/components/FormInput.vue';
import ProfileAvatar from '@/instances/desktop/components/ProfileAvatar.vue';
import SkillsManager from '@/instances/desktop/components/SkillsManager.vue';
import ToolsManager from '@/instances/desktop/components/ToolsManager.vue';
import LinksManager from '@/instances/desktop/components/LinksManager.vue';
import ContributionGraph from '@/instances/desktop/components/ContributionGraph.vue';
import ToggleSwitch from '@/instances/common/components/ToggleSwitch.vue';
import utils from '@/services/utils';

// Stores
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const userStore = useUserStore();
const profileStore = useProfileStore();
const trayStates = useTrayStates();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();
const stage = useStageStore();

// State
const loading = ref(true);
const error = ref('');
const success = ref('');
const isEditing = ref(false);
const photoInput = ref(null);
const photoPreview = ref(null);
const currentPhoto = ref(null);
const passwordCard = ref(null);
const userProfileBody = ref(null);
const checkingEmailAvailability = ref(false);
const checkingUsernameAvailability = ref(false);
const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const userNameRegex = /^[a-zA-Z0-9_]{3,}$/;
const isEmailTaken = ref(false);
const isUsernameTaken = ref(false);
const isSavingChanges = ref(false);
const areLinksValid = ref(true);

// Section-specific edit states
const editingSections = reactive({
  header: false,
  studios: false,
  skills: false,
  tools: false,
  password: false
});

// Reference data (for dropdowns/pickers)
const allSkills = ref([]);
const allTools = ref([]);

// Password Change Data
const passwordData = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
});

const passwordErrors = reactive({
  newPassword: '',
  confirmPassword: ''
});

// Form Data - Using computed properties that read/write to profileStore
const formData = computed({
  get: () => profileStore.profile,
  set: (value) => profileStore.updateProfileFields(value)
});

const errors = reactive({
  first_name: '',
  last_name: '',
  username: '',
  email: ''
});

const editableUserPhoto = reactive({
  photo: null
});

// Computed Properties
const userData = computed(() => userStore.user);

const fullName = computed(() => {
  return `${formData.value.first_name} ${formData.value.last_name}`.trim() || 'User';
});

const profileColor = computed(() => {
  if (userData.value?.id) {
    const parts = userData.value.id.split('-');
    return '#' + parts[0];
  }
  return '#666';
});

const usernameValid = computed(() => {
  if (formData.value.username === userData.value?.username) return true;
  return userNameRegex.test(formData.value.username);
});

const emailValid = computed(() => {
  return emailRegex.test(formData.value.email);
});

const detailsInputed = computed(() => {
  return formData.value.first_name && formData.value.last_name && formData.value.username && formData.value.email;
});

const credentialsValid = computed(() => {
  return emailValid.value && !isEmailTaken.value && usernameValid.value && !isUsernameTaken.value;
});

// Profile visibility as computed property
const profileVisibility = computed({
  get: () => profileStore.profile.profile_visibility === 'public',
  set: (value) => {
    const visibility = value ? 'public' : 'private';
    profileStore.setProfileVisibility(visibility);
  }
});

const isPasswordValid = computed(() => {
  // Validate password requirements
  if (!passwordData.currentPassword || !passwordData.newPassword || !passwordData.confirmPassword) {
    return false;
  }
  
  // Check minimum password length
  if (passwordData.newPassword.length < 8) {
    passwordErrors.newPassword = 'Password must be at least 8 characters';
    return false;
  } else {
    passwordErrors.newPassword = '';
  }
  
  // Check if passwords match
  if (passwordData.newPassword !== passwordData.confirmPassword) {
    passwordErrors.confirmPassword = 'Passwords do not match';
    return false;
  } else {
    passwordErrors.confirmPassword = '';
  }
  
  return true;
});

const isDataChanged = computed(() => {
  const basicFieldsChanged = ['first_name', 'last_name', 'username', 'email', 'title', 'country', 'availability']
    .some(key => formData.value[key] !== (userData.value?.[key] || ''));
  
  return basicFieldsChanged || photoPreview.value;
});

const isDataValid = computed(() => {
  return detailsInputed.value && credentialsValid.value && isDataChanged.value && areLinksValid.value;
});

const userPhoto = computed(() => {
  if (photoPreview.value) return photoPreview.value;
  if (!userStore.user.photo) return '/icons/default_profile_picture.svg';
  return userStore.user.photo;
});

// Check if any section is being edited
const isAnyEditMode = computed(() => {
  return Object.values(editingSections).some(value => value === true);
});

// Check if sections that require "Save" button are being edited
// Skills and Tools are saved immediately, so they don't need the save button
const needsSaveButton = computed(() => {
  return editingSections.header || editingSections.password;
});

// Methods
const scrollToTop = () => {
  if (userProfileBody.value) {
    userProfileBody.value.scrollTo({ top: 0, behavior: 'smooth' });
  }
};

const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

const startEditing = () => {
  isEditing.value = true;
};

const cancelEditing = async () => {
  isEditing.value = false;
  // Reload profile from server to discard changes
  await loadUserProfile();
  photoPreview.value = null;
  error.value = '';
  success.value = '';
};

// Toggle section-specific editing
const toggleSectionEdit = async (section) => {
  editingSections[section] = !editingSections[section];
  
  // If opening password section, scroll to it after Vue updates the DOM
  if (section === 'password' && editingSections[section]) {
    await nextTick();
    if (passwordCard.value && passwordCard.value.$el) {
      passwordCard.value.$el.scrollIntoView({ 
        behavior: 'smooth', 
        block: 'center' 
      });
    }
  }
};

// Save all changes from all edited sections
const saveAllChanges = async () => {
  await handleUpdate();
  // Reset all section edit states
  Object.keys(editingSections).forEach(key => {
    editingSections[key] = false;
  });
};

// Cancel all edits and reset
const cancelAllEdits = async () => {
  // Reload profile from server to discard changes
  await loadUserProfile();
  photoPreview.value = null;
  error.value = '';
  success.value = '';
  // Reset all section edit states
  Object.keys(editingSections).forEach(key => {
    editingSections[key] = false;
  });
  
  // Scroll to top after canceling
  scrollToTop();
};

const handlePhotoChange = (filePath, preview) => {
  editableUserPhoto.photo = filePath;
  photoPreview.value = preview;
};

const removePhoto = () => {
  console.log('removed')
  editableUserPhoto.photo = null;
  photoPreview.value = null;
  userStore.user.photo = null;
};

const handleStudioLogoError = (event) => {
  // If studio logo fails to load, replace with default icon
  event.target.src = iconStore.getAppIcon('stall');
  event.target.classList.add('studio-icon-default');
};

const handleUpdatePhoto = async () => {
  try {
    if (!editableUserPhoto.photo) return;
    
    // Set saving state
    isSavingChanges.value = true;
    stage.operationActive = true;
    
    // Upload photo via ProfileService (passing the file path)
    await ProfileService.UpdateUserPhoto(editableUserPhoto.photo);
    
    // Update local stores with the preview
    userStore.user.photo = photoPreview.value;
    profileStore.profile.photo = photoPreview.value;
    
    // Clear photo state
    editableUserPhoto.photo = null;
    photoPreview.value = null;
    
  } catch (err) {
    throw new Error(err?.message || 'Failed to update profile photo');
  } finally {
    isSavingChanges.value = false;
    stage.operationActive = false;
  }
};

const handleUpdate = async () => {
  try {
    error.value = '';
    isSavingChanges.value = true;
    stage.operationActive = true;
    
    // Prepare update data with all editable fields (including links)
    const updateData = {
      first_name: formData.value.first_name,
      last_name: formData.value.last_name,
      username: formData.value.username,
      email: formData.value.email,
      job_title: formData.value.title,
      location: formData.value.country,
      availability: formData.value.availability,
      bio: formData.value.bio,
      // Include link fields from the links object
      artstation_link: formData.value.links?.artstation || '',
      behance_link: formData.value.links?.behance || '',
      linkedin_link: formData.value.links?.linkedin || '',
      portfolio_link: formData.value.links?.portfolio || '',
      instagram_link: formData.value.links?.instagram || '',
    };

    console.log(updateData)
    // Update profile via Wails ProfileService
    await ProfileService.UpdateUserProfile(userStore.user.id, updateData)
      .then(async () => {
        // Only update stores if API call succeeds
        profileStore.updateProfileFields(updateData);
        
        // Also update userStore with basic user info for consistency
        Object.keys(userStore.user).forEach(key => {
          if (updateData[key] !== undefined) {
            userStore.user[key] = updateData[key];
          }
        });

        // Handle photo upload separately if there's a photo to upload
        if (photoPreview.value) {
          await handleUpdatePhoto();
        }

        notificationStore.addNotification(
          "Profile updated.", 
          "Profile updated successfully.", 
          "success", 
          true
        );
        isEditing.value = false;
        
        // Scroll to top after successful save
        scrollToTop();
      });
    
  } catch (err) {
    error.value = err?.message || 'Failed to update profile';
    notificationStore.errorNotification("Failed to update profile.", err?.message || err);
  } finally {
    isSavingChanges.value = false;
    stage.operationActive = false;
  }
};

const checkUsername = async () => {
  const sameUsername = formData.value.username === userData.value?.username;
  
  if (sameUsername) {
    isUsernameTaken.value = false;
    return;
  }
  
  if (!formData.value.username) return;
  checkingUsernameAvailability.value = true;
  
  try {
    const usernameExist = await AuthService.CheckUsernameExists(formData.value.username.toLowerCase());
    if (usernameExist) {
      errors.username = 'Username is already taken';
      isUsernameTaken.value = true;
    } else {
      errors.username = '';
      isUsernameTaken.value = false;
    }
    checkingUsernameAvailability.value = false;
  } catch (error) {
    errors.username = '';
    console.error('Error checking username:', error);
    checkingUsernameAvailability.value = false;
  }
};

const checkEmail = async () => {
  const sameEmail = formData.value.email === userData.value?.email;
  
  if (sameEmail) {
    isEmailTaken.value = false;
    return;
  }
  
  if (!formData.value.email || !emailValid.value) return;
  checkingEmailAvailability.value = true;
  
  try {
    const emailExist = await AuthService.CheckEmailExists(formData.value.email);
    if (emailExist) {
      isEmailTaken.value = true;
      errors.email = 'Email is already registered';
    } else {
      isEmailTaken.value = false;
      errors.email = '';
    }
    checkingEmailAvailability.value = false;
  } catch (error) {
    errors.email = '';
    console.error('Error checking email:', error);
    checkingEmailAvailability.value = false;
  }
};

const toggleAvailability = () => {
  const currentAvailability = formData.value.availability;
  const newAvailability = currentAvailability === "available" 
    ? "not_looking" 
    : "available";
  profileStore.setAvailability(newAvailability);
};

const handlePasswordUpdate = async () => {
  try {
    if (!isPasswordValid.value) return;
    
    // Call the API to update password
    await AuthService.ChangePassword(passwordData.currentPassword, passwordData.newPassword, passwordData.confirmPassword);
    
    notificationStore.addNotification(
      "Password updated successfully", 
      "Your password has been changed.", 
      "success", 
      true
    );
    
    // Clear password fields
    passwordData.currentPassword = '';
    passwordData.newPassword = '';
    passwordData.confirmPassword = '';
    passwordErrors.newPassword = '';
    passwordErrors.confirmPassword = '';
    
    // Close the password editing section
    editingSections.password = false;
    
  } catch (err) {
    notificationStore.errorNotification(
      "Failed to update password", 
      err?.message || err
    );
  }
};

const toggleProfileVisibility = async () => {
  try {
    // Toggle the visibility value
    const newVisibility = profileVisibility.value ? 'private' : 'public';
    
    // Update via Wails ProfileService
    await ProfileService.UpdateUserProfile(userStore.user.id, {
      profile_visibility: newVisibility
    });
    
    // Update local store
    profileStore.setProfileVisibility(newVisibility);
    
    notificationStore.addNotification(
      "Profile visibility updated", 
      `Your profile is now ${newVisibility}.`, 
      "success", 
      false
    );
    
  } catch (err) {
    // Revert on error
    console.error('Failed to update profile visibility:', err);
    notificationStore.errorNotification(
      "Failed to update profile visibility", 
      err?.message || err
    );
  }
};

const updateLinks = (newLinks) => {
  profileStore.updateLinks(newLinks);
};

const handleLinksValidUpdate = (isValid) => {
  areLinksValid.value = isValid;
};

const prepDeleteAccountModal = () => {
  trayStates.popUpModalIcon = 'trash';
  trayStates.popUpModalTitle = "Delete account";
  trayStates.popUpModalMessage = "Your account will be irreversibly deleted and you will be unable to access files through Clustta. Continue?";
  trayStates.popUpModalFunction = deactivateUserAccount;
  modals.setModalVisibility('popUpModal', true);
};

const deactivateUserAccount = async () => {
  try {
    await AuthService.DeactivateUserAccount();
    await AuthService.Logout()
      .then((data) => {
        userStore.$reset();
        projectStore.$reset();
        trayStates.$reset();
      })
      .catch((error) => {
        notificationStore.errorNotification("Logout Failed", error);
      });
    modals.disableAllModals();
  } catch (err) {
    notificationStore.errorNotification("Failed to delete account.", err?.message || err);
  }
};

const loadUserProfile = async () => {
  if (!userStore.user?.id) {
    console.error('No user ID available');
    return;
  }
  
  try {
    loading.value = true;
    profileStore.setFetching(true);
    
    // Fetch complete profile from server using Wails binding
    const profileData = await ProfileService.GetUserProfile(userStore.user.id);
    
    // Store in profileStore
    profileStore.setProfile(profileData);
    
    currentPhoto.value = profileData.photo || userPhoto.value;
    
  } catch (err) {
    console.error('Failed to load user profile:', err);
    notificationStore.errorNotification(
      "Failed to load profile", 
      err?.message || err
    );
  } finally {
    loading.value = false;
    profileStore.setFetching(false);
  }
};

const loadReferenceData = async () => {
  try {
    // Load all available skills and tools for the dropdowns
    const [skillsData, toolsData] = await Promise.all([
      ProfileService.GetAllSkills(),
      ProfileService.GetAllTools()
    ]);
    
    allSkills.value = skillsData || [];
    allTools.value = toolsData || [];
    
  } catch (err) {
    console.error('Failed to load reference data:', err);
    // Don't show error notification for reference data - it's not critical
  }
};

onBeforeMount(async () => {
  if (userStore.user) {
    // Load profile and reference data in parallel
    await Promise.all([
      loadUserProfile(),
      loadReferenceData()
    ]);
  } else {
    loading.value = false;
  }
});
</script>

<style scoped>
.user-profile-root {
  box-sizing: border-box;
  padding: 0.4rem;
  width: 90%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  /* background-color: var(--black-steel); */
  color: var(--white);
  /* border-radius: 12px; */
}

.user-profile-stage {
  display: flex;
  flex-direction: column;
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  /* background-color: forestgreen ; */
  gap: 0.5rem;
}

.user-profile-body {
  width: 100%;
  height: 100%;
  display: flex;
  box-sizing: border-box;
  align-items: flex-start;
  justify-content: center;
  overflow-y: auto;
  padding: 0.5rem;
  /* background-color: royalblue; */
  /* padding-bottom: 200px; */
}

.user-profile-body::-webkit-scrollbar {
  width: 6px;
}

.user-profile-body::-webkit-scrollbar-thumb {
  background-color: var(--light-steel);
  border-radius: 3px;
}

.user-profile-body::-webkit-scrollbar-track {
  background-color: var(--dark-steel);
  border-radius: 3px;
}

.user-profile-container {
  width: 100%;
  max-width: 900px;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding: 1rem;
  box-sizing: border-box;
  /* background-color: hotpink; */
  height: 100%;
  height: min-content;
  border-radius: 0px;
}

/* Edit Controls */
.edit-controls {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  align-items: center;
  margin-left: auto;
  width: fit-content;
}

.profile-visibility-toggle {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  background-color: var(--steel);
  border-radius: var(--normal-radius);
}

.visibility-label {
  font-size: 0.875rem;
  color: var(--white);
  font-weight: 400;
  user-select: none;
}

.button-group {
  display: flex;
  gap: 0.5rem;
}

/* Header Layout */
.header-layout {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  align-items: center;
}

@media (min-width: 768px) {
  .header-layout {
    flex-direction: row;
    align-items: flex-start;
  }
}

.header-info {
  flex: 1;
  width: 100%;
}

.edit-mode-fields,
.display-mode-fields {
  display: flex;
  flex-direction: column;
  /* gap: 1rem; */
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.profile-name {
  font-size: 2rem;
  font-weight: 500;
  margin: 0 0 0.25rem 0;
  color: var(--white);
}

.profile-title {
  font-size: 1.25rem;
  margin: 0 0 1rem 0;
  font-weight: 400;
  color: var(--white);
}

/* Meta Info (Location & Availability) */
.meta-info {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  align-items: center;
  margin-bottom: 1rem;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.875rem;
  color: var(--white);
}

.availability-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  font-size: 0.875rem;
  color: var(--white);
}

.info-icon {
  width: 16px;
  height: 16px;
  /* filter: brightness(0) invert(1); */
  opacity: 0.7;
}

/* Social Links */
.social-links {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

/* Links Section in Edit Mode */
.links-section {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

/* Professional Info in Edit Mode */
.availability-field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  font-size: 0.875rem;
  color: var(--white);
  font-weight: 400;
}

/* Change Password Section */
.change-password-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

/* Studios Section */
.studios-container {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
}

.studio-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem;
  background-color: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  transition: all 0.2s ease;
  cursor: pointer;
  min-width: 120px;
}

.studio-item:hover {
  background-color: rgba(255, 255, 255, 0.1);
  transform: translateY(-2px);
}

.studio-logo {
  width: 64px;
  height: 64px;
  object-fit: cover;
  border-radius: 8px;
}

.studio-icon-default {
  opacity: 0.6;
}

.studio-name {
  font-size: 0.875rem;
  color: var(--white);
  text-align: center;
  font-weight: 400;
  line-height: 1.2;
}

.no-studios {
  color: rgba(255, 255, 255, 0.5);
  font-size: 0.875rem;
  text-align: center;
  padding: 2rem;
}

/* Danger Zone */
.danger-zone {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1rem;
  background-color: rgba(220, 38, 38, 0.1);
  border-radius: var(--normal-radius);
  border: 1px solid rgba(220, 38, 38, 0.3);
}

.danger-message {
    color: var(--white);
  margin: 0;
  font-size: 0.875rem;
}

/* Floating Action Buttons */
.floating-actions {
  position: fixed;
  bottom: 2rem;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background: linear-gradient(135deg, var(--black-steel) 0%, var(--dark-steel) 100%);
  border-radius: 50px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.1);
  z-index: 1000;
  animation: slideUp 0.3s ease-out;
  backdrop-filter: blur(10px);
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}

.floating-actions:hover {
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.5);
  border-color: rgba(255, 255, 255, 0.2);
}

/* Responsive */
@media (max-width: 768px) {
  .form-row {
    grid-template-columns: 1fr;
  }
  
  .user-profile-container {
    padding: 0.5rem;
  }
  
  .floating-actions {
    bottom: 1rem;
    padding: 0.75rem 1rem;
    gap: 0.75rem;
    flex-wrap: wrap;
    justify-content: center;
    max-width: calc(100% - 2rem);
  }
}
</style>
