<template>
  <div class="user-profile-root absolute-pane">
      <div class="user-profile-body">
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
              v-if="!isEditing" 
              :iconAfter="true" 
              :icon="getAppIcon('edit')"
              label="Edit Profile" 
              @click="startEditing"
                :useOutline="true"
            />
            <div v-else class="button-group">
              <ActionButton 
                :color="'crimson'" 
                :iconAfter="true" 
                :icon="getAppIcon('close-circle')" 
                label="Cancel"
                @click="cancelEditing"
                :useOutline="true"
              />
              <ActionButton 
                :isDisabled="!isDataValid"  
                :iconAfter="true" 
                :icon="getAppIcon('check-circle')"
                label="Save Changes" 
                @click="handleUpdate"
                :useOutline="true"
              />
            </div>
          </div>

          <!-- Header Card -->
          <ProfileCard>
            <div class="header-layout">
              <ProfileAvatar
                :userPhoto="userPhoto"
                :avatarColor="profileColor"
                :isEditing="isEditing"
                @photoChanged="handlePhotoChange"
                @photoRemoved="removePhoto"
              />
              
              <div class="header-info">
                <div v-if="isEditing" class="edit-mode-fields">
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
                    v-model="formData.title"
                    label="Professional Title"
                    placeholder="e.g., 3D Artist & Animator"
                  />
                  <FormInput
                    v-model="formData.country"
                    label="Location"
                    placeholder="e.g., United States"
                  />
                  <div class="availability-field">
                    <label class="form-label">Availability</label>
                    <ActionButton
                      @click="toggleAvailability"
                      :icon="getAppIcon('check-circle')"
                      :label="formData.availability"
                      :iconAfter="false"
                      :useBackground="true"
                    />
                  </div>
                  
                  <!-- Professional Links in Edit Mode -->
                  <div class="links-section">
                    <LinksManager
                      :links="formData.links"
                      :isEditing="isEditing"
                      @update:links="updateLinks"
                    />
                  </div>
                </div>
                
                <div v-else class="display-mode-fields">
                  <div class="profile-name">{{ fullName }}</div>
                  <div v-if="formData.title" class="profile-title">{{ formData.title }}</div>
                  
                  <div class="meta-info">
                    <div v-if="formData.country" class="info-item">
                      <img class="info-icon small-icons" :src="getAppIcon('map-pin')" alt="">
                      <span>{{ formData.country }}</span>
                    </div>
                    
                    <div v-if="formData.availability" class="availability-badge">
                      <img class="info-icon small-icons" :src="getAppIcon('check-circle')" alt="">
                      <span>{{ formData.availability }}</span>
                    </div>
                  </div>
                  
                  <!-- Professional Links in Display Mode -->
                  <div class="social-links">
                    <LinksManager
                      :links="formData.links"
                      :isEditing="isEditing"
                      @update:links="updateLinks"
                    />
                  </div>
                </div>
              </div>
            </div>
          </ProfileCard>

          <!-- Studios Card -->
          <ProfileCard title="Studios">
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
              <div v-if="!formData.studios || formData.studios.length === 0" class="no-studios">
                No studios to display
              </div>
            </div>
          </ProfileCard>

          <!-- Skills Card -->
          <ProfileCard title="Skills">
            <SkillsManager
              :skills="formData.skills"
              :isEditing="isEditing"
              @skillAdded="addSkill"
              @skillRemoved="removeSkill"
            />
          </ProfileCard>

          <!-- Tools & Software Card -->
          <ProfileCard title="Tools & Software">
            <ToolsManager
              :tools="formData.tools"
              :isEditing="isEditing"
              @toolAdded="addTool"
              @toolRemoved="removeTool"
            />
          </ProfileCard>

          <!-- Change Password Card -->
          <ProfileCard v-if="isEditing" title="Change Password">
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
          <ProfileCard v-if="!isEditing" title="Activity">
            <ContributionGraph />
          </ProfileCard>

          <!-- Danger Zone -->
          <ProfileCard v-if="!isEditing" title="Danger Zone">
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
  </div>
</template>

<script setup>
import { ref, reactive, computed, onBeforeMount } from 'vue';
import { AuthService } from "@/../bindings/clustta/services";
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';
import { useIconStore } from '@/stores/icons';
import { useUserStore } from '@/stores/users';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useTrayStates } from '@/stores/TrayStates';

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

// Stores
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const userStore = useUserStore();
const trayStates = useTrayStates();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

// State
const loading = ref(true);
const error = ref('');
const success = ref('');
const isEditing = ref(false);
const photoInput = ref(null);
const photoPreview = ref(null);
const currentPhoto = ref(null);
const checkingEmailAvailability = ref(false);
const checkingUsernameAvailability = ref(false);
const profileVisibility = ref(true);
const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const userNameRegex = /^[a-zA-Z0-9_]{3,}$/;
const isEmailTaken = ref(false);
const isUsernameTaken = ref(false);

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

// Form Data
const formData = reactive({
  first_name: '',
  last_name: '',
  username: '',
  email: 'taiwofolu@eaxum.com',
  title: 'Pixel Pusher',
  country: 'Nigeria',
  availability: 'Available for Freelance',
  links: {
    behance: 'https://behance.net',
    artstation: 'https://artstation.com',
    portfolio: '',
    linkedin: ''
  },
  skills: [
    { name: '3D Modeling', icon: 'box' },
    { name: 'Character Animation', icon: 'man-running' },
    { name: 'Texturing', icon: 'palette' }
  ],
  tools: [
    { name: 'Blender', logo: '' },
    { name: 'Maya', logo: '' }
  ],
  studios: [
    { id: '1', name: 'Pixar Animation Studios', role: '3D Animator', logo: '' },
    { id: '2', name: 'DreamWorks Animation', role: 'Character Artist', logo: '' },
    { id: '3', name: 'Blue Sky Studios', role: 'Technical Artist', logo: '' }
  ]
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
  return `${formData.first_name} ${formData.last_name}`.trim() || 'User';
});

const profileColor = computed(() => {
  if (userData.value?.id) {
    const parts = userData.value.id.split('-');
    return '#' + parts[0];
  }
  return '#666';
});

const usernameValid = computed(() => {
  if (formData.username === userData.value?.username) return true;
  return userNameRegex.test(formData.username);
});

const emailValid = computed(() => {
  return emailRegex.test(formData.email);
});

const detailsInputed = computed(() => {
  return formData.first_name && formData.last_name && formData.username && formData.email;
});

const credentialsValid = computed(() => {
  return emailValid.value && !isEmailTaken.value && usernameValid.value && !isUsernameTaken.value;
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
    .some(key => formData[key] !== (userData.value?.[key] || ''));
  
  return basicFieldsChanged || photoPreview.value;
});

const isDataValid = computed(() => {
  return detailsInputed.value && credentialsValid.value && isDataChanged.value;
});

const userPhoto = computed(() => {
  if (photoPreview.value) return photoPreview.value;
  if (!userStore.user?.photo) return '/icons/default_profile_picture.svg';
  return userStore.user.photo;
});

// Methods
const getAppIcon = (iconName) => {
  return iconStore.getAppIcon(iconName);
};

const startEditing = () => {
  isEditing.value = true;
};

const cancelEditing = () => {
  isEditing.value = false;
  populateData();
  photoPreview.value = null;
  error.value = '';
  success.value = '';
};

const handlePhotoChange = (file, preview) => {
  editableUserPhoto.photo = file;
  photoPreview.value = preview;
};

const removePhoto = () => {
  editableUserPhoto.photo = null;
  photoPreview.value = null;
};

const handleStudioLogoError = (event) => {
  // If studio logo fails to load, replace with default icon
  event.target.src = iconStore.getAppIcon('stall');
  event.target.classList.add('studio-icon-default');
};

const handleUpdate = async () => {
  try {
    error.value = '';
    
    await AuthService.UpdateUser(
      formData.first_name, 
      formData.last_name, 
      formData.username, 
      formData.email
    ).then(async (updatedUserData) => {
      // Update local data
      Object.keys(userStore.user).forEach(key => {
        if (updatedUserData[key]) {
          userStore.user[key] = updatedUserData[key];
        }
      });

      let longMessage = `Profile updated successfully.`;
      notificationStore.addNotification("Profile updated.", longMessage, "success", true);
      isEditing.value = false;
    }).catch((error) => {
      notificationStore.errorNotification("Failed to update profile.", error);
    });
    
  } catch (err) {
    error.value = err?.message || 'Failed to update profile';
  }
};

const checkUsername = async () => {
  const sameUsername = formData.username === userData.value?.username;
  
  if (sameUsername) {
    isUsernameTaken.value = false;
    return;
  }
  
  if (!formData.username) return;
  checkingUsernameAvailability.value = true;
  
  try {
    const usernameExist = await AuthService.CheckUsernameExists(formData.username.toLowerCase());
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
  const sameEmail = formData.email === userData.value?.email;
  
  if (sameEmail) {
    isEmailTaken.value = false;
    return;
  }
  
  if (!formData.email || !emailValid.value) return;
  checkingEmailAvailability.value = true;
  
  try {
    const emailExist = await AuthService.CheckEmailExists(formData.email);
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
  formData.availability = formData.availability === "Available for Freelance" 
    ? "Not Available" 
    : "Available for Freelance";
};

const handlePasswordUpdate = async () => {
  try {
    if (!isPasswordValid.value) return;
    
    // TODO: Implement API call to update password
    // await AuthService.UpdatePassword(passwordData.currentPassword, passwordData.newPassword);
    
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
    
  } catch (err) {
    notificationStore.errorNotification(
      "Failed to update password", 
      err?.message || err
    );
  }
};

const toggleProfileVisibility = () => {
  profileVisibility.value = !profileVisibility.value;
  // TODO: Implement API call to save profile visibility setting
  console.log('Profile visibility:', profileVisibility.value ? 'Public' : 'Private');
};

const updateLinks = (newLinks) => {
  formData.links = { ...newLinks };
};

const addSkill = (skill) => {
  formData.skills.push(skill);
};

const removeSkill = (skillName) => {
  const index = formData.skills.findIndex(s => s.name === skillName);
  if (index > -1) {
    formData.skills.splice(index, 1);
  }
};

const addTool = (tool) => {
  formData.tools.push(tool);
};

const removeTool = (toolName) => {
  const index = formData.tools.findIndex(t => t.name === toolName);
  if (index > -1) {
    formData.tools.splice(index, 1);
  }
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

const populateData = () => {
  currentPhoto.value = userPhoto.value;
  
  Object.keys(formData).forEach(key => {
    if (key !== 'skills' && key !== 'tools' && key !== 'links') {
      if (userData.value?.[key]) {
        formData[key] = userData.value[key];
      }
    }
  });
};

onBeforeMount(async () => {
  if (userStore.user) {
    await populateData();
  }
  loading.value = false;
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
  background-color: var(--black-steel);
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

/* Responsive */
@media (max-width: 768px) {
  .form-row {
    grid-template-columns: 1fr;
  }
  
  .user-profile-container {
    padding: 0.5rem;
  }
}
</style>
