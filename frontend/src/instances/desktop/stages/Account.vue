<template>
	<div class="page-list-root absolute-pane">
		<div class="settings-stage-root">
			
			<div class="settings-stage-body">
				<div class="settings-stage-body-container">
					
		
					<div v-if="!isEditing" class="dashboard-actions">
						<ActionButton  :iconAfter="true" :icon="getAppIcon('edit')"
							label="Edit Profile" @click="startEditing" />
						<!-- <ActionButton  :iconAfter="true" :icon="getAppIcon('key')"
							label="Change Password" @click="goToChangePassword" /> -->
					</div>

					<div v-else class="dashboard-actions">
						<ActionButton :color="'crimson'" :iconAfter="true" :icon="getAppIcon('close-circle')" label="Cancel"
							@click="cancelEditing" />
					</div>

					<form v-if="!loading" @submit.prevent="handleUpdate" class="profile-form">

					<!-- <div class="user-profile-root" :class="{ 'user-profile-disabled': !isEditing }">
						<div v-if="!isEditing" class="user-profile-container">
							<div class="photo-container">
								<img class="profile-img" :src="userPhoto">
							</div>
						</div>

						<div v-else class="user-profile-container">
						<div class="photo-overlay"></div>
						<div class="photo-actions">
							<div class="change-action-button">
							<img class="alert-icons" :src="getAppIcon('camera')" />
							<input type="file" @change="handlePhotoChange" accept="image/*" ref="photoInput" class="photo-input" />
							</div>
							<div class="change-action-button">
							<img class="alert-icons" src="/icons/close.svg" @click="removePhoto" />
							</div>
						</div>

						<div class="photo-container">
							<img class="profile-img" :src="photoPreview ? photoPreview : userPhoto">
						</div>
						</div>

					</div> -->

					<div class="form-row">
						<div class="form-group">
						<label>First Name</label>
						<input class="form-input input-short" v-model="formData.first_name" type="text" :disabled="!isEditing" />
						
						</div>
						<div class="form-group">
						<label>Last Name</label>
						<input class="form-input input-short" v-model="formData.last_name" type="text" :disabled="!isEditing" />
						</div>
					</div>

					<div class="form-group">
						<label>Username</label>
						<div class="compound-form-input">
							<input class="form-input input-short" @input="checkUsername" v-model="formData.username" type="text" :disabled="!isEditing" />
							<div v-if="formData.username" class="form-input-icon">
								<img v-if="errors.username || !usernameValid" class="alert-icons" :src="getAppIcon('alert')" />
								<img v-else-if="checkingUsernameAvailability" class="alert-icons" src="/icons/loading.svg"
								:class="{ 'loading-icon': checkingUsernameAvailability }" />
								<img v-else class="alert-icons" :src="getAppIcon('circle-check')" />
							</div>
						</div>
					</div>

					<div class="form-group">
						<label>Email</label>
						<div class="compound-form-input">
							<input class="form-input input-short" @input="checkEmail" v-model="formData.email" type="email" :disabled="!isEditing" />
							<div v-if="formData.email" class="form-input-icon">
								<img v-if="errors.email || !emailValid" class="alert-icons" :src="getAppIcon('alert')" />
								<img v-else-if="checkingEmailAvailability" class="alert-icons" src="/icons/loading.svg"
									:class="{ 'loading-icon': checkingEmailAvailability }" />
								<img v-else class="alert-icons" :src="getAppIcon('circle-check')" />
							</div>
						</div>
					</div>

					</form>

						
						<ActionButton v-if="isEditing" :isDisabled="!isDataValid"  :iconAfter="true" :icon="getAppIcon('check-circle')"
							label="Save Changes" @click="handleUpdate" />

						<ActionButton v-else  :iconAfter="true" :icon="getAppIcon('trash')"
							label="Delete account" @click="prepDeleteAccountModal()" />
				</div>
			</div>
		</div>
	</div>
</template>

<script setup>
// imports
import { ref, reactive, computed, onMounted, onBeforeMount } from 'vue'
import { useRouter } from 'vue-router'
import { AuthService } from "@/../bindings/clustta/services";
import { useNotificationStore } from '@/stores/notifications';
import { useProjectStore } from '@/stores/projects';

// state imports
import { useIconStore } from '@/stores/icons';
import { useUserStore } from '@/stores/users';
import { useDesktopModalStore } from '@/stores/desktopModals';
import { useTrayStates } from '@/stores/TrayStates';

// states/stores
const iconStore = useIconStore();
const modals = useDesktopModalStore();
const userStore = useUserStore();
const trayStates = useTrayStates();
const notificationStore = useNotificationStore();
const projectStore = useProjectStore();

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue'


// refs
const router = useRouter()
const loading = ref(true)
const error = ref('')
const success = ref('')
const isEditing = ref(false)
const photoInput = ref(null)
const photoPreview = ref(null)
const currentPhoto = ref(null);
const checkingEmailAvailability = ref(false);
const checkingUsernameAvailability = ref(false);
const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const userNameRegex = /^[a-zA-Z0-9_]{3,}$/
const isEmailTaken = ref(false);
const isUsernameTaken = ref(false);

// refs
const formData = reactive({
	first_name: '',
	last_name: '',
	username: '',
	email: '',
});

const errors = reactive({
	first_name: '',
	last_name: '',
	username: '',
	email: '',
});

const editableUserPhoto = reactive({
	photo: null
});

// computed props
const userData = computed(() => {
	return userStore.user
});

const usernameValid = computed(() => { 
	if( formData.username === userData.value.username) return true
	return userNameRegex.test(formData.username) 
});

const emailValid = computed(() => { 
	return emailRegex.test(formData.email) 
});

const detailsInputed = computed(() => { return formData.first_name && formData.last_name && formData.username && formData.email });
const credentialsValid = computed(() => { return emailValid.value && !isEmailTaken.value && usernameValid.value && !isUsernameTaken.value });


const isDataChanged = computed(() => {
  return Object.keys(formData).some(key => formData[key] !== userData.value?.[key]) || photoPreview.value;
});

const isDataValid = computed(() => {
  return detailsInputed.value && credentialsValid.value && isDataChanged.value
});


const userPhoto = computed(() => {
	if (!userStore.user?.photo) {
	return '/icons/default_profile_picture.svg'
	}
	return userStore.user.photo
});


// methods
const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

const prepDeleteAccountModal = () => {
	trayStates.popUpModalIcon = 'trash'
	trayStates.popUpModalTitle = "Delete account";
	trayStates.popUpModalMessage = "Your account will be irreversibly deleted and you will be unable to access files through Clustta. Continue?";
	trayStates.popUpModalFunction = deactivateUserAccount;
	modals.setModalVisibility('popUpModal', true);
};

const deactivateUserAccount = async () => {
	try {
		await AuthService.DeactivateUserAccount()
		await AuthService.Logout()
		.then((data) => {
			userStore.$reset()
			projectStore.$reset();
			trayStates.$reset();
		}
		)
		.catch((error) => {
			notificationStore.errorNotification("Logout Failed", error)
		});
		modals.disableAllModals();
	} catch (err) {
		notificationStore.errorNotification("Failed to delete account.", err?.message || err)
	}
}

const startEditing = () => {
	isEditing.value = true
}

const cancelEditing = () => {
	isEditing.value = false
	// Reset form data
	Object.keys(formData).forEach(key => {
	if (userData.value[key]) {
		formData[key] = userData.value[key]
	} else {
		formData[key] = ''
	}
	})
	photoPreview.value = null
	error.value = ''
	success.value = ''
}

const handlePhotoChange = (event) => {
	const file = event.target.files[0]
	if (file) {
		editableUserPhoto.photo = file
		const reader = new FileReader()
		reader.onload = (e) => {
		photoPreview.value = e.target.result
	}
	reader.readAsDataURL(file)
}
}

const removePhoto = () => {
	editableUserPhoto.photo = null
	photoPreview.value = null
	if (photoInput.value) {
		photoInput.value.value = ''
	}
}

const handleUpdatePhoto = async () => {
	try {
		error.value = ''
		await AuthService.UpdateUserPhoto(editableUserPhoto.photo)
		.then((data) => {
			console.log(data)
		})
		// update data
		userStore.userPhoto = editableUserPhoto.photo;
		editableUserPhoto.photo = null
		photoPreview.value = null
	} catch (err) {
		error.value = err?.message || 'Failed to update profile photo'
	}
};

const handleUpdate = async () => {

	try {
	error.value = ''

	await AuthService.UpdateUser(formData.first_name, formData.last_name, formData.username, formData.email)
    .then(async (updatedUserData) => {

		// Update local data
		Object.keys(userStore.user).forEach(key => {
			if (updatedUserData[key]) {
				userStore.user[key] = updatedUserData[key]
			}
		})

		// if (photoPreview.value) {
		// 	await handleUpdatePhoto();
		// }

		let longMessage = `Details updated.`
		notificationStore.addNotification("Details updated.", longMessage, "success", true);
		isEditing.value = false;
		
    }).catch((error) => {
		notificationStore.errorNotification("Failed to update details.", error)
    })
	
	isEditing.value = false

	} catch (err) {
		error.value = err.responsePhoto?.data?.message || 'Failed to update profile'
	}
}

const checkUsername = async () => {

	const sameUsername  = formData.username === userData.value.username;

	if(sameUsername){
		isUsernameTaken.value = false;
		return
	}

	if (!formData.username) return
	checkingUsernameAvailability.value = true;

	try {
		const usernameExist = await AuthService.CheckUsernameExists(formData.username.toLowerCase())
		if (usernameExist) {
		errors.username = 'Username is already taken'
		isUsernameTaken.value = true;
		} else {
		errors.username = ''
		isUsernameTaken.value = false;
		}
		checkingUsernameAvailability.value = false;
	} catch (error) {
		errors.username = ''
		console.error('Error checking username:', error)
		checkingUsernameAvailability.value = false;
	}
};

const checkEmail = async () => {
	
	const sameEmail  = formData.email === userData.value.email;

	if(sameEmail){
		isEmailTaken.value = false;
		return
	}

	if (!formData.email || !emailValid.value) return
	checkingEmailAvailability.value = true;

	try {
		const emailExist = await AuthService.CheckEmailExists(formData.email)
		if (emailExist) {
		isEmailTaken.value = true;
		errors.email = 'Email is already registered'
		} else {
		isEmailTaken.value = false;
		errors.email = ''
		}
		checkingEmailAvailability.value = false;
	} catch (error) {
		errors.email = ''
		console.error('Error checking email:', error);
		checkingEmailAvailability.value = false;
	}
};

const goToChangePassword = () => {
	router.push('/change-password');
};

const populateData = async () => {

	currentPhoto.value = userPhoto.value;

	Object.keys(formData).forEach(key => {
		if (userData.value[key]) {
			formData[key] = userData.value[key]
		}
	})
}

onBeforeMount(async () => {
	console.log(userStore.user)
	if(userStore.user){
		await populateData();
	}
	loading.value = false
})

</script>

<style scoped>
@import "@/assets/desktop.css";

.compound-form-input{
  box-sizing: border-box;
  border-radius: 4px;
  font-size: 1rem;
  transition: border-color 0.2s;
  width: 100%;
  height: 50px;
  border-radius: var(--normal-radius);
  padding-right: .5rem;
  display: flex;
  overflow: hidden;
  gap: .2rem;
  background-color: var(--midnight-steel);
}

.form-input {
  box-sizing: border-box;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 1rem;
  transition: border-color 0.2s;
  width: 100%;
  height: 50px;
  border-radius: var(--normal-radius);
  padding: 0.75rem;

  font-family: 'Inter', sans-serif;
  font-weight: 200;
  box-sizing: border-box;
  font-size: 16px;
  border-radius: 12px;
  padding: 10px;
  border: 0px;
  border-style: solid;
  outline: none;
  background-color: var(--midnight-steel);
  color: var(--white);
}

.dashboard-actions {
    padding: .5rem;
    gap: .5rem;
    display: flex;
    justify-content: flex-start;
    align-items: center;
    width: 100%;
    overflow: hidden;
    flex-wrap: wrap;
    /* margin-bottom: 2rem; */
    box-sizing: border-box;
    box-sizing: border-box;
    /* background-color: forestgreen; */
}

.page-list-root {
	box-sizing: border-box;
	padding: .4rem;
	display: flex;
	align-items: center;
	justify-content: center;
	color: var(--white);
	/* background-color: sienna; */
}

.settings-stage-root {
	display: flex;
	flex-direction: column;
	box-sizing: border-box;
	/* background-color: firebrick; */
	width: 100%;
	height: 100%;
	gap: .5rem;
}

.settings-stage-header {
	width: 100%;
	display: flex;
	box-sizing: border-box;
	/* background-color: rebeccapurple; */
	align-items: flex-start;
	justify-content: flex-start;
}

.settings-stage-body {
	width: 100%;
	/* max-width: 1200px; */
	height: 100%;
	display: flex;
	box-sizing: border-box;
	/* background-color: teal; */
	align-items: center;
	justify-content: center;
	overflow: hidden;
	padding: .5rem;
}

.settings-stage-body-container {
	width: 100%;
	max-width: 960px;
	/* height: 100%; */
	display: flex;
	box-sizing: border-box;
	background-color: tomato;
	align-items: center;
	justify-content: center;
	overflow: hidden;
	padding: .5rem;

	
    box-sizing: border-box;
    padding: 1rem;
    display: flex;
    flex-direction: column;
    width: 100%;
    gap: 2rem;
    border-radius: var(--normal-radius);
    background-color: var(--black-steel);
    max-width: 800px;
}

.page-list {
	background-color: rgb(125, 192, 59);
	padding: .4rem;
	display: flex;
	gap: .4rem;
	flex-direction: column;
	box-sizing: border-box;
	height: 100%;
	width: 100%;
	/* overflow: hidden; */
	height: max-content;
}

.page-list-container {
	display: flex;
	padding-right: .4rem;
	overflow: hidden;
	overflow-y: scroll;
	height: 100%;
	width: 100%;
	box-sizing: border-box;
	/* max-width: 600px; */
	min-width: 300px;
}

.page-list-container::-webkit-scrollbar {
	width: 8px;
}

.page-list-container::-webkit-scrollbar-thumb {
	border-radius: 10px;
	background-color: rgb(36, 49, 59);
}

.page-list-container::-webkit-scrollbar-track {
	border-radius: 10px;
}

.page-header {
	position: relative;
	display: flex;
	width: 100%;
	align-items: center;
	height: max-content;
	gap: 1rem;
	justify-content: space-between;
	background-color: khaki;
	padding: .2rem;
	box-sizing: border-box;
	min-width: max-content;
}

.profile-form {
    display: flex;
    flex-direction: column;
    align-items: center;
    /* background-color: goldenrod; */
    width: 100%;
    max-width: 800px;
    /* margin: 0 auto; */
  }
  
  .form-row {
    width: 100%;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
    margin-bottom: .5rem;
  }
  
  .form-group {
    margin-bottom: 1rem;
  }
  
  .form-group label {
    display: block;
    margin-bottom: 0.8rem;
    color: var(--white);
    font-weight: 300;
  }
  
  .form-group input:disabled {
    /* background-color: #f3f4f6; */
    opacity: .5;
    cursor: not-allowed;
  }
  
  .loading-message {
    text-align: center;
    padding: 2rem;
    color: #6b7280;
  }
  
  .error-message {
    color: #dc2626;
    margin: 1rem 0;
    padding: 0.75rem;
    background-color: #fee2e2;
    border-radius: 4px;
  }
  
  .success-message {
    color: #059669;
    margin: 1rem 0;
    padding: 0.75rem;
    background-color: #ecfdf5;
    border-radius: 4px;
  }
</style>