<template>
	<div class="action-bar-root">
		<div class="create-menu">
			<ActionButton :icon="getAppIcon(itemIcon)" :label="itemType ? itemType : 'Add'"  v-tooltip="itemType ? itemType : 'Add'" :buttonFunction="addFunction" />
		</div>
	</div>
</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
  const formattedIconName = getIconName(iconName)
  const icon = iconStore.getAppIcon(formattedIconName);
  return icon
};

const getIconName = (path) => {
  // If the string doesn't contain '/' or '.svg', return it as is
  if (!path.includes('/') && !path.includes('.svg')) {
    return path;
  }
  return path.split('/').pop().replace('.svg', '');
};

import ActionButton from '@/instances/desktop/components/ActionButton.vue'

// methods

const props = defineProps({
	itemType: String,
	itemIcon: {
		type: String, 
		default: 'plus-circle'
		},
	allowSync: {
        type:Boolean, 
        default: false
    },
	cardView: {
        type:Boolean, 
        default: false
    }, 
	addFunction: {
        type:Function, 
        default: () => {
        }
    }, 
	switchViewLayout: {
        type:Function, 
        default: () => {
        }
    },  
	refresh: {
        type:Function, 
        default: () => {
        }
    }, 
	sync: {
        type:Function, 
        default: () => {
        }
    }, 
});
</script>

<style scoped>
@import "@/assets/desktop.css";

.action-bar-root{
	position: relative;
	display: flex;
	width: 100%;
	align-items: center;
	height: max-content;
	gap: 1rem;
	justify-content: space-between;
	/* background-color: khaki; */
	padding: .2rem;
	box-sizing: border-box;
	min-width: max-content;
}
.create-menu{
	position: relative;
	display: flex;
	align-items: center;
	gap: .4rem;
	width: max-content;
	height: max-content;
	padding: .2rem;
	/* background-color: black; */
}

.action-bar{
	position: relative;
	display: flex;
	align-items: center;
	gap: .4rem;
	width: max-content;
	height: max-content;
	padding: .2rem;
	/* background-color: black; */
}
.view-options{
	display: flex;
	gap: .4rem;
	align-items: center;
	padding: .2rem;
	width: max-content;
	height: max-content;
	/* background-color: darkorange; */
}
</style>