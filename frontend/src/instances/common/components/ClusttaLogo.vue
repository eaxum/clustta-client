<template>
    <div class="logo-container">
        <div class="clustta-logo">
            <img v-if="colored" src="/icons/clustta.png"  alt="Clustta Icon">
            <img v-else :src="getAppIcon('clustta')"  alt="Clustta Icon">
        </div>
        <div v-if="showText" class="clustta-logo-text" :class="{ 'inverted' : inverted,  'clustta-logo-bold' : boldText}" >
            Clustta
        </div>
    </div>
</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

// props
const props = defineProps({
    showText : { type: Boolean, default: true},
    inverted : { type: Boolean, default: false},
    colored : { type: Boolean, default: false},
    boldText : { type: Boolean, default: false},
});

// methods
const getAppIcon = (iconName) => {
	const icon = iconStore.getAppIcon(iconName);
	return icon
};

</script>


<style scoped>
.clustta-logo {
  display: flex;
  width: 50px;
  height: 100%;
  overflow: hidden;
  justify-content: flex-start;
  align-items: center;
  justify-content: center;
}

.clustta-logo img {
  height: 80%;
  width: 80%;
}

.clustta-logo{
    width: 50px;
    height: 50px;
    aspect-ratio: 1/1;
    display: flex;
    cursor: pointer;
}

.clustta-logo-text{
    color: var(--white);
    font-size: 1.3rem;
    font-family: 'Bricolage Grotesque', sans-serif;
	font-weight: 500;
	font-size: x-large;
	display: flex;
	align-items: center;
	justify-content: flex-start;
}

.clustta-logo-bold{
    font-weight: 500;
}

.logo-container{
    height: 100%;
    height: 60px;
    overflow: hidden;
    flex: 1;
    display: flex;
    align-items: center;
    padding: .2rem;
    box-sizing: border-box;
    cursor: pointer;
}

</style>