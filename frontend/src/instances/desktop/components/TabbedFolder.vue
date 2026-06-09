<template>
  <div class="tabs-container">

    <!-- <div class="red-element"></div> -->

    <div class="tab-bar">
        <!-- <div class="tab"></div> -->
        <div class="folder">
          <slot></slot>
        </div>
    </div>
    
  </div>
</template>

<script setup>
</script>

<style scoped>
@import "@/assets/desktop.css";

.tabs-container {
  width: 100%;
  height: 700px;
  overflow: hidden;
  position: relative;
}

.red-element {
  position: absolute;
  width: 50%;
  height: 50%;
  background-color: hsl(var(--destructive));
  z-index: 1;
}

.tab-bar {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  width: 100%;
  height: 100%;
  padding-top: 50px;
  /* border-radius: var(--normal-radius); */
  
  box-sizing: border-box;
  overflow: hidden;
  position: relative;
  z-index: 2;
}

.tab {
  display: flex;
  align-items: center;
  width: 60%;
  height: 70px;
  border-radius: 24px 22px 0 0;
  background-color: hsl(var(--accent));
  color: hsl(var(--foreground));
  user-select: none;
  box-sizing: border-box;
  position: relative;
  transform: skewX(45deg);
  left: -30px;
  top: .5px;
  border-top: 1px solid hsl(var(--border));
  border-right: 1px solid hsl(var(--border));
  /* transition: all .2s ease-out; */
}

.tab::after {
  content: "";
  position: absolute;
  right: -50px;
  bottom: 0;
  width: 50px;
  height: 25px;
  background-color: hsl(var(--success));
  background-color: transparent;
  border-bottom-right-radius: 16px;
  border-bottom-left-radius: 16px;
  box-shadow: -25px 0 0 0 hsl(var(--accent));
  border-left: 1px solid hsl(var(--border));
  /* transition: all .2s ease-out; */
}

.folder {
  width: 100%;
  height: 100%;
  border-radius: 0 24px 0px 0px;
  border-radius: 0 36px 0px 0px;
  background-color: hsl(var(--accent));
  display: flex;
  align-items: flex-end;
  box-sizing: border-box;
  padding: .5rem;
  transition: all .2s ease-out;
}
</style>




