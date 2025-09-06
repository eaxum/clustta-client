<template>
    <div class="regex-builder-root">
        <div class="regex-builder-header">
            <HeaderArea :title="props.header" />
            <ActionButton :icon="getAppIcon('plus-circle')" label="Add Pattern" @click="addNewPair" />
        </div>
        <div class="regex-builder-wrapper">
            <div class="regex-builder-item">
                <div class="input-section">
                    <div class="horizontal-flex">
                        <input class="input-short regex-input" type="text" value="Default"
                            placeholder="Enter regex pattern" readonly v-focus />

                        <ListBox :items="props.actionOptions" :selectedItem="props.defaultAction[1]"
                            :onSelect="selectDefaultAction" />
                    </div>
                </div>
                <div v-for="(rule, index) in rules" :key="index" class="">
                    <div class="input-section">
                        <div class="horizontal-flex">
                            <input v-model="rule[0]" class="input-short regex-input" type="text"
                                placeholder="Enter regex pattern" v-focus />

                            <ListBox :items="props.actionOptions" :selectedItem="rule[1]" :onSelect="selectAction"
                                :extraData="{ index: index }" />

                            <!-- <ActionButton v-if="pairs.length >= 2" :icon="getAppIcon('trash')" v-tooltip="'Delete'"
                                @click="removePair(pair.id)" /> -->
                            <ActionButton :icon="getAppIcon('trash')" v-tooltip="'Delete'" @click="deleteRule(index)" />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { useIconStore } from '@/stores/icons';
const iconStore = useIconStore();

const getAppIcon = (iconName) => {
    const icon = iconStore.getAppIcon(iconName);
    return icon
};
import { ref } from 'vue'
import ListBox from '@/instances/common/components/ListBox.vue';
import ActionButton from '@/instances/desktop/components/ActionButton.vue'
import HeaderArea from '@/instances/common/components/HeaderArea.vue';
import { useImportStore } from '@/stores/import';

const importStore = useImportStore();

const props = defineProps({
    actionOptions: Array,
    selectedAction: String,
    header: String,
    defaultAction: { type: Array, default: [] },
    rules: { type: Array, default: [] },
});

// const emit = defineEmits(['rules-updated']);
const pairs = ref([])

// const updatedPattern = () => {
//     let rules = [
//     ]
//     for (let pair of pairs.value) {
//         rules.push([pair.pattern, pair.action])
//     }
//     rules.push(["^.*$", defaultAction.value])
//     emit('rules-updated', rules);
// }
const selectAction = (action, data) => {
    props.rules[data.index][1] = action
    // updatedPattern()
    // console.log(importStore.fileTypeRule)
}
const selectDefaultAction = (action) => {
    console.log(props.defaultAction)
    props.defaultAction[1] = action
    // updatedPattern()
}

const addAttachment = (index) => {
    attachments.value.splice(index, 1);
    attachmentsPath.value.splice(index, 1);
}
const deleteRule = (index) => {
    props.rules.splice(index, 1);
}

const addNewPair = () => {
    const newId = Math.max(...pairs.value.map(p => p.id), 0) + 1
    pairs.value.push({ id: newId, pattern: '', action: props.selectedAction })
    // updatedPattern()
}

const removePair = (id) => {
    pairs.value = pairs.value.filter(pair => pair.id !== id)
    // updatedPattern()
}

// Expose the pairs data for parent components
defineExpose({
    pairs
})
</script>

<style scoped>
.regex-input {
    width: 50%;
    height: 35px;
}

.regex-builder-header {
    /* background-color: red; */
    display: flex;
    align-items: center;
    justify-content: flex-end;
    justify-content: space-between;
    width: 100%;
    min-height: max-content;
}

.regex-builder-root {
    box-sizing: border-box;
    background-color: var(--steel);
    border-radius: var(--normal-radius);
    width: 100%;
    padding: .5rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    overflow: hidden;
    gap: 10px;
    max-height: 90vh;
}

.regex-builder-wrapper {
    /* background-color: blue; */
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    /* max-height: 200px; */
    overflow: hidden;
    overflow-y: scroll;
}

.regex-builder-wrapper::-webkit-scrollbar {
    width: 8px;
}

.regex-builder-wrapper::-webkit-scrollbar-thumb {
    border-radius: 10px;
    background-color: var(--black-steel);
}

.regex-builder-wrapper::-webkit-scrollbar-track {
    border-radius: 10px;
}

.regex-builder-item {
    /* background-color: darkred; */
    width: 100%;
    display: flex;
    gap: 10px;
    flex-direction: column;
}

.wrapper {
    width: 200px;
    padding: 20px;
    height: 150px;
}

/* Any additional custom styles can go here */
</style>