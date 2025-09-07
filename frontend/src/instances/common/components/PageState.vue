<template>
  <transition name="fade" appear>
    <div class="state-container" :class="{'multiple': !secondaryActionMessage}">
      <div class="page-state">
        <img class="page-state-illustration" :src="illustration">
        <div class="page-state-messages"> 
          <div class="page-state-message"> <p>{{ message }}</p></div>
          <div class="page-state-message"> <p>{{ prompt }}</p></div>
        </div>
        <button v-if="primaryActionMessage" class="button colored contained-button" @click="primaryActionFunction" v-stop-propagation>{{ primaryActionMessage }}</button>
      </div>
      <div class="secondary-action-area">
        <span v-if="secondaryActionMessage" @click="secondaryActionFunction" class="action-button state-secondary-action" v-stop-propagation>
          <img class="small-icons" :src="secondaryIcon">
          {{ secondaryActionMessage }}
        </span>
      </div>
    </div>
  </transition>

</template>
  
<script setup>
const props = defineProps({

  message: String,
  prompt: String,
  illustration: String,
  primaryActionMessage: String,
  primaryActionFunction: Function,
  secondaryActionMessage: String,
  secondaryActionFunction: Function,
  secondaryIcon: String,

});
 
</script>
  
<style scoped>
  @import "@/assets/tray.css";

.secondary-action-area{
  color: var(--white);
  /* background-color: chocolate; */
}
.state-secondary-action{

  /* position: absolute; */
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.216);
  border-radius: 16px;
  padding: .4rem 1rem;
  text-align: center;
  font-size: 14px;
  line-height: 14px;

}
.state-container{
  position: relative;
  box-sizing: border-box;
  /* background-color: darkmagenta; */
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  /* flex: 1; */
  overflow: hidden;
  align-items: center;
  /* justify-content: space-between; */
  justify-content: space-around;
}

.multiple{
  /* background-color: deeppink; */
  justify-content: center;
}

.contained-button{
  width: max-content;
}

.page-state{
  /* position: relative; */
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: max-content;
  max-height: 80%;
  flex-direction: column;
  color: var(--white);
  /* background-color:red; */
  overflow: hidden;
  padding: .5rem 0;
  border-radius: 10px;
  gap: .5rem;
}

.page-state-illustration{
  box-sizing: border-box;
  padding: 2px;
  height: 100%;
  /* min-width: 100%; */
  overflow: hidden;
  display: flex;
  align-items: center;
  max-width: 400px;
  }

  .page-state-messages{

  display: flex;
  align-items: center;
  flex-direction: column;
  justify-content: center;
  /* background-color: crimson; */
  width: 96%;
  gap: .5rem;
  padding: .5rem;
  box-sizing: border-box;
  height: max-content;

}

.page-state-message{

  display: flex;
  align-items: center;
  justify-content: center;
  /* background-color: forestgreen; */
  width: 96%;
  height:20px;
  text-align: center;
  font-size: 14px;
  line-height: 14px;

}
</style>
  
  
  

