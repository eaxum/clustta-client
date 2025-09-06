<template>
    <div class="auth-guard-root">
        <SignUp v-if="showSignUp" @toggle-login="toggleLogin" />
        <Login  v-else @toggle-login="toggleLogin"  />
        <InfoBar />
    </div>
</template>

<script setup>
// imports
import { ref, onBeforeUnmount } from 'vue';

// components
import Login from '@/instances/desktop/pages/Login.vue'
import SignUp from '@/instances/desktop/pages/SignUp.vue'
import InfoBar from '@/instances/desktop/components/InfoBar.vue'


// refs
const showSignUp = ref(false);

// methods
const toggleLogin = () => {
    showSignUp.value = !showSignUp.value;
}

onBeforeUnmount(() => {
    showSignUp.value = false;
});

</script>

<style >
@import "@/assets/desktop.css";

.auth-guard-root{
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background-color: var(--black-steel);
    box-sizing: border-box;
    overflow: hidden;
}

.sign-up-page-root{
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background-color: var(--black-steel);
}

.page-root{
  display: flex;
  box-sizing: border-box;
  width: 100%;
  align-items: center;
  height: 100%;
  flex-direction: column;
  background-color: var(--black);
  overflow: hidden;
  overflow-y: auto;
}

.auth-root {
  display: flex;
  flex-direction: row;
  align-items: center;
  height: 100%;
  min-height: min-content;
  max-width: min-content;
  width: 100%;
  justify-content: space-around;
  gap: 2rem;
  box-sizing: border-box;
}

.header-container{
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 400px;
  box-sizing: border-box;
  overflow: hidden;
  height: max-content;
  min-width: 300px;
}

.auth-container {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  width: min-content;
  min-width: 400px;
  box-sizing: border-box;
  overflow: hidden;
  /* background-color: springgreen; */
  overflow: hidden;
  /* overflow-y: auto; */
}

.auth-header {
  font-family: 'Bricolage Grotesque', sans-serif;
  font-size: 4rem;
  font-weight: 600;
  line-height: 90%;
  width: max-content;
  text-align: left;
  color: var(--white);
  height: max-content;
  width: 100%;
  text-wrap: wrap;
}

.auth-form-container {
  padding: 1rem;
  width: 100%;
  max-width: 480px;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-row {
  overflow: hidden;
  box-sizing: border-box;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0px;
  width: 100%;
}

.toggle-button {
  background: none;
  border: none;
  color: var(--white);
  font-weight: 300;
  cursor: pointer;
  transition: color 0.2s;
  text-decoration: none;
}

.toggle-button:hover {
  font-weight: 500;
}

.error-message {
  margin-top: .5rem;
  color: #dc2626;
  text-align: center;
  font-size: 0.875rem;
  font-weight: 300;
  width: 100%;
  display: flex;
  padding: .3rem;
}

.submit-button-icon {
  width: 30px;
  height: 30px;
}

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

.form-input-mini {
  color: var(--white);
  box-sizing: border-box;
  border: 0px;
  border-radius: 4px;
  font-size: 1rem;
  width: 100%;
  height: 100%;
  padding: 0.75rem;
  background-color: var(--midnight-steel);
  
  font-family: 'Inter', sans-serif;
  /* font-weight: 200; */
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

.form-input-icon {
  display: flex;
  cursor: pointer;
  align-items: center;
  justify-content: center;
  flex: 1;
  box-sizing: border-box;
  border: 0px;
  border-radius: 4px;
  font-size: 1rem;
  width: 100%;
  height: 100%;
}

.alert-icons{
  min-height: 24px;
  max-height: 24px;
  min-width: 24px;
  max-width: 24px;
  align-items: center;
  transition: transform 0.2s ease;
}

.loading-icon {
  animation: loadingRotate .5s linear infinite;
}

@keyframes loadingRotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.form-input {
  box-sizing: border-box;
  border-radius: 4px;
  font-size: 1rem;
  transition: border-color 0.2s;
  width: 100%;
  height: 50px;
  border-radius: var(--normal-radius);
  padding: 0.75rem;
}

.input-short {
  /* Retained because it's used in the template */
}

.submit-button {
  font-size: x-large;
  background-color: var(--grape);
  color: var(--white);
  padding: 0.75rem;
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  border: none;
  border-radius: var(--normal-radius);
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
  color: var(--black);
}

.submit-button:hover {
  background-color: var(--bright-grape);
}

.button-inactive{
  opacity: .5;
  cursor: not-allowed;
}

.button-inactive:hover{
  background-color: var(--grape);
}

input.error {
  border-color: #dc3545;
}

/* Photo styles */
.user-profile-root{
  height: max-content;
  width: 100%;
  padding: 1rem;
  box-sizing: border-box;
  display: flex;
  overflow: hidden;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  position: relative;
}

.user-profile-container{
  border-radius: 60px;
  width: 80px;
  height: 80px;
  box-sizing: border-box;
  display: flex;
  overflow: hidden;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  position: relative;
}

.user-profile-container:hover>*:first-child {
  opacity: .7;
  visibility: visible;
  transition: opacity 0.2s ease-in-out;
}

.photo-container{
  width: 100%;
  height: 100%;
  background-color: var(--grape);
}

.photo-overlay{
  width: 100%;
  height: 100%;
  background-color: black;
  opacity: 0;
  position: absolute;
}

.photo-actions{
  z-index: 1;
  position: absolute;
  top: 25%;
  left: 0;
  cursor: pointer;
  display: flex;
  box-sizing: border-box;
  width: 100%;
  height: 50%;
  overflow: hidden;
  padding: 5%;
  gap: 5%;
}

.change-action-button {
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  position: relative;
  width: 100%;
  height: 100%;
}

.toggle-container{
  width: 90%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--white);
  cursor: pointer;
  gap: 1rem;
}

@media (max-width: 768px) {
  .form-row {
    grid-template-columns: 1fr;
  }

  .auth-form-container {
    padding: 1.5rem;
  }

  .auth-root{
    flex-direction: column;
    padding: .5rem;
    gap: 0px;
    min-width: 100%;
  }
  
  .auth-container{
    width: 100%;
    flex: 1;
    min-width: 300px;
    padding: 0 2rem;
    margin-bottom: 2rem;
  }
  
  .auth-header{
    font-size: 3rem;
    text-wrap: wrap;
    width: 100%;
  }
  
  .header-container{
    justify-content: flex-start;
    width: 100%;
    height: max-content;
    max-width: 500px;
    padding: 1.5rem;
    padding-bottom: .5rem;
  }
}
</style>