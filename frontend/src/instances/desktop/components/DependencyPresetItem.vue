<template>
    <div class="preset-item-container">
        <div class="preset-item">
            <div @click="toggleExpanded" class="preset-item-meta">
                <div class="meta">
                    <div class="preset-item-name">
                        <div class="preset-item-label-text">{{ preset.name }}</div>
                    </div>
                    <div class="preset-item-name">
                        <div class="preset-item-label-text subtitle">{{ preset.dependencies?.length || 0 }} {{ $t('components.dependencyPresetItem.dependencies') }}</div>
                    </div>
                </div>
            </div>

            <div class="preset-item-actions">
                <span @click="applyPreset" class="single-action-button" v-tooltip="$t('components.dependencyPresetItem.apply')">
                    <CiPlusCircle :size="20" />
                </span>

                <span @click="deletePreset" class="single-action-button" v-tooltip="$t('components.dependencyPresetItem.delete')">
                    <CiTrash :size="20" />
                </span>

                <span @click="toggleExpanded" class="single-action-button" v-tooltip="isExpanded ? $t('components.dependencyPresetItem.collapse') : $t('components.dependencyPresetItem.expand')">
                    <CiChevronDown :size="20" />
                </span>
            </div>
        </div>

        <transition name="expand" appear>
            <div v-if="preset.dependencies?.length" v-show="isExpanded || isLoading" class="preset-dependencies-root">
                <div v-if="isLoading" class="preset-dependencies loading">
                    <ActionButton :icon="CiLoading" class="spinner-icon" :isDisabled="true" />
                </div>

                <div v-else class="preset-dependencies">
                    <div class="preset-dependencies-list">
                        <div class="dependency-child-item" v-for="(dep, index) in enrichedDependencies" :key="dep.id || index">
                            <div class="dependency-item-meta">
                                <span><img class="small-icons no-filter" :src="dep.icon || CiGeneric"></span>
                                <div class="dependency-item-label">
                                    <div @click="goToItem(dep)" class="dependency-item-label-text">{{ dep.name || dep.id }}</div>
                                </div>
                            </div>
                            <div class="dependency-item-actions">
                                <span @click="removeDependencyFromPreset(dep, index)" class="single-action-button" v-tooltip="$t('components.dependencyPresetItem.remove')">
                                    <CiMinusCircle :size="20" />
                                </span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </transition>
    </div>
</template>

<script setup>
// imports
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { CiChevronDown, CiGeneric, CiLoading, CiMinusCircle, CiPlusCircle, CiTrash } from '@clustta/icons-vue';
import { resolveIcon } from '@/lib/icon-map';

// components
import ActionButton from '@/instances/desktop/components/ActionButton.vue';

// services
import { AssetService, CollectionService } from '@/services';

// stores
import { useAssetStore } from '@/stores/assets';
import { useCollectionStore } from '@/stores/collections';
import { useCommonStore } from '@/stores/common';
import { useIconStore } from '@/stores/icons';
import { useProjectStore } from '@/stores/projects';
import { useStageStore } from '@/stores/stages';

const assetStore = useAssetStore();
const collectionStore = useCollectionStore();
const commonStore = useCommonStore();
const iconStore = useIconStore();
const projectStore = useProjectStore();
const stage = useStageStore();
const { t } = useI18n();

// utils
import { isValidWeblink } from '@/lib/pointer';
import utils from '@/services/utils';

// props
const props = defineProps({
    preset: {
        type: Object,
        required: true
    }
});

// emits
const emit = defineEmits(['apply', 'delete', 'update']);

// refs
const enrichedDependencies = ref([]);
const isExpanded = ref(false);
const isLoading = ref(false);

// methods
const applyPreset = () => {
    emit('apply', props.preset);
};

// Deletes the preset.
const deletePreset = () => {
    emit('delete', props.preset);
};

// Removes a dependency from the preset.
const removeDependencyFromPreset = (dep, index) => {
    enrichedDependencies.value.splice(index, 1);
    const updatedDependencies = props.preset.dependencies.filter(d => d.id !== dep.id);
    emit('update', { ...props.preset, dependencies: updatedDependencies });
};

// Fetches and enriches dependency data with names, icons, and paths.
const fetchEnrichedDependencies = async () => {
    const project = projectStore.activeProject;
    const depIds = props.preset.dependencies.map(d => d.id);
    const children = await AssetService.GetAssetDependencies(project.uri, depIds);

    for (let i = 0; i < children.length; i++) {
        let item = children[i];
        let extension = "";
        
        if (item.pointer) {
            item.file_path = item.pointer;
        }
        
        if (item.is_link && !isValidWeblink(item.pointer)) {
            extension = utils.getFileExtension(item.pointer).substring(1);
        } else if (!item.is_link) {
            extension = children[i].extension?.toLowerCase().substring(1);
        }
        
        let iconPath = "";
        if (item.type === "asset") {
            if (item.is_link && isValidWeblink(item.pointer)) {
                iconPath = await iconStore.getWebIcon(item.pointer);
            } else {
                iconPath = (await iconStore.getIcon(extension)) || "";
            }
        }
        children[i].icon = iconPath;
        
        let preview = null;
        if (item.preview) {
            preview = "data:image/png;base64," + item.preview;
        }
        children[i].preview = preview;
    }

    enrichedDependencies.value = children;
};

const getAppIcon = (iconName) => {
    const icon = iconStore.resolveIcon(iconName);
    return icon;
};

// Navigates to the dependency item in the browser.
const goToItem = async (dep) => {
    try {
        if (dep.type === 'asset') {
            const asset = await AssetService.GetAssetByID(projectStore.activeProject.uri, dep.id);
            if (!asset?.id) return;
            const assetParent = await CollectionService.GetCollectionByID(projectStore.activeProject.uri, asset.collection_id);
            if (assetParent) {
                collectionStore.navigateToCollection(assetParent);
                commonStore.navigatorMode = true;
            }
            stage.deselectAllItems();
            assetStore.selectAsset(asset.id);
            stage.firstSelectedItemId = asset.id;
            stage.markedItems = [asset.id];
        } else if (dep.type === 'collection') {
            const collection = await CollectionService.GetCollectionByID(projectStore.activeProject.uri, dep.id);
            if (collection) {
                collectionStore.navigateToCollection(collection);
                commonStore.navigatorMode = true;
            }
        }
    } catch (error) {
        console.error('Failed to navigate to dependency:', error);
    }
};

// Toggles expanded state and fetches dependencies on first expand.
const toggleExpanded = async () => {
    if (!isExpanded.value && enrichedDependencies.value.length === 0) {
        isLoading.value = true;
        try {
            await fetchEnrichedDependencies();
        } finally {
            isLoading.value = false;
        }
    }
    isExpanded.value = !isExpanded.value;
};
</script>

<style scoped>
@import "@/assets/desktop.css";

.preset-item-container {
    position: relative;
    cursor: auto;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    width: 100%;
    gap: .2rem;
    color: var(--white);
    border-radius: var(--large-radius);
    outline: var(--transparent-line);
    outline-offset: -1px;
    background-color: var(--dark-steel);
    transition: all .2s ease-in-out;
    min-height: 50px;
    min-height: max-content;
}

.preset-item-container:hover {
    border-radius: var(--normal-radius);
    background-color: var(--steel);
}

.preset-item {
    position: relative;
    cursor: auto;
    box-sizing: border-box;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    width: 100%;
    padding-right: .5rem;
    overflow: hidden;
}

.preset-item-meta {
    padding-left: .5rem;
    box-sizing: border-box;
    overflow: hidden;
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: .5rem;
    width: 100%;
    cursor: pointer;
}

.meta {
    flex-direction: column;
    align-items: flex-start;
    min-height: min-content;
    padding: .5rem 0;
    justify-content: center;
    gap: .2rem;
    overflow: hidden;
    width: 100%;
    display: flex;
    white-space: nowrap;
}

.preset-item-name {
    height: min-content;
    width: 100%;
    overflow: hidden;
}

.preset-item-label-text {
    color: var(--white);
    font-size: 14px;
    height: min-content;
    overflow: hidden;
    width: 100%;
    text-overflow: ellipsis;
}

.preset-item-label-text.subtitle {
    font-size: 12px;
    font-weight: 400;
    color: var(--subtle-text);
}

.preset-item-actions {
    display: none;
}

.preset-item-container:hover .preset-item-actions {
    display: flex;
}

.is-expanded {
    transform: rotate(180deg);
}

.preset-dependencies-root {
    box-sizing: border-box;
    position: relative;
    width: 100%;
    overflow: hidden;
    padding: 0 .3rem;
    padding-bottom: .3rem;
    max-height: 200px;
}

.preset-dependencies {
    border-radius: var(--normal-radius);
    background-color: var(--light-steel);
    padding: .5rem 0;
    max-height: 180px;
    overflow-y: auto;
}

.preset-dependencies-list {
    display: block;
}

.preset-dependencies::-webkit-scrollbar {
    width: 4px;
}

.preset-dependencies::-webkit-scrollbar-thumb {
    border-radius: var(--small-radius);
    background-color: var(--light-steel);
}

.preset-dependencies::-webkit-scrollbar-track {
    border-radius: var(--small-radius);
}

.dependency-child-item {
    position: relative;
    cursor: auto;
    box-sizing: border-box;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    overflow: hidden;
    padding: .35rem .5rem .35rem 0;
    height: 35px;
}

.dependency-item-actions {
    display: none;
}

.dependency-child-item:hover .dependency-item-actions {
    display: flex;
}

.dependency-item-meta {
    padding-left: .5rem;
    box-sizing: border-box;
    overflow: hidden;
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: .5rem;
    width: 100%;
}

.dependency-item-label {
    align-items: center;
    overflow: hidden;
    width: 100%;
    display: flex;
    white-space: nowrap;
}

.dependency-item-label-text {
    font-family: 'Inter', sans-serif;
    font-size: 14px;
    color: var(--white);
    text-overflow: ellipsis;
    cursor: pointer;
}

.dependency-item-label-text:hover {
    text-decoration: underline;
}

.expand-enter-active,
.expand-leave-active {
    transition: all 0.2s ease-in-out;
}

.expand-enter-from,
.expand-leave-to {
    opacity: 0;
    max-height: 0;
}

.preset-dependencies.loading {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: .5rem;
}

.spinner-icon {
    animation: spin 1s linear infinite;
}

@keyframes spin {
    from {
        transform: rotate(0deg);
    }
    to {
        transform: rotate(360deg);
    }
}
</style>
