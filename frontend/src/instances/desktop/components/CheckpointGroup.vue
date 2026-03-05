<template>
    <div class="checkpoint-group">

        <div ref="railRef" class="timeline-rail" :class="{ 'timeline-rail-first': isFirstGroup, 'timeline-rail-last': isLastGroup }">
            <div class="timeline-dot" v-for="checkpoint in group.items" :key="checkpoint.checkpoint_id"
                :class="{ 'timeline-dot-active': checkpoint.hash === taskHash }"
                :style="{ top: dotPositions[checkpoint.checkpoint_id] }">
            </div>
        </div>

        <div class="checkpoint-data">
            <CheckpointGroupHeader :label="group.label" />

            <CheckpointItem v-for="checkpoint in group.items" :ref="el => setItemRef(checkpoint.checkpoint_id, el)"
                :key="checkpoint.checkpoint_id" :checkpoint="checkpoint" :taskHash="taskHash" :expandedId="expandedId"
                @refreshCheckpoints="$emit('refreshCheckpoints')" @updateTaskHash="$emit('updateTaskHash')"
                @updateExpanded="$emit('updateExpanded', $event)" />
        </div>

    </div>
</template>

<script setup>
// imports
import { ref, onMounted, nextTick, watch } from 'vue';

// components
import CheckpointGroupHeader from '@/instances/desktop/components/CheckpointGroupHeader.vue';
import CheckpointItem from '@/instances/desktop/components/CheckpointItem.vue';

// props
const props = defineProps({
    group: {
        type: Object,
        required: true
    },
    taskHash: {
        type: String,
        default: ''
    },
    expandedId: {
        type: String,
        default: ''
    },
    isFirstGroup: {
        type: Boolean,
        default: false
    },
    isLastGroup: {
        type: Boolean,
        default: false
    }
});

// emits
const emit = defineEmits(['refreshCheckpoints', 'updateTaskHash', 'updateExpanded']);

// refs
const dotPositions = ref({});
const itemRefs = {};
const railRef = ref(null);

// Stores a reference to each CheckpointItem element.
const setItemRef = (checkpointId, el) => {
    if (el) {
        itemRefs[checkpointId] = el;
    }
};

// Calculates the vertical position of each dot to align with its card.
const calculateDotPositions = () => {
    nextTick(() => {
        if (!railRef.value) return;
        const railTop = railRef.value.getBoundingClientRect().top;
        const positions = {};
        let firstDotTop = null;
        let lastDotTop = 0;
        for (const checkpoint of props.group.items) {
            const itemEl = itemRefs[checkpoint.checkpoint_id];
            if (itemEl && itemEl.$el) {
                const rect = itemEl.$el.getBoundingClientRect();
                const top = rect.top - railTop + (rect.height / 2);
                positions[checkpoint.checkpoint_id] = `${top}px`;
                if (firstDotTop === null) firstDotTop = top;
                lastDotTop = top;
            }
        }
        dotPositions.value = positions;

        if (props.isFirstGroup) {
            railRef.value.style.setProperty('--first-dot-top', `${firstDotTop ?? 0}px`);
        }
        if (props.isLastGroup) {
            const lineStart = props.isFirstGroup ? (firstDotTop ?? 0) : 0;
            railRef.value.style.setProperty('--last-dot-height', `${lastDotTop - lineStart}px`);
        }
    });
};

// lifecycle hooks
onMounted(() => {
    calculateDotPositions();
});

// watchers
watch(() => [props.group.items, props.expandedId], () => {
    calculateDotPositions();
}, { deep: true });
</script>

<style scoped>
.checkpoint-group {
    display: flex;
    flex-direction: row;
    gap: .5rem;
    width: 100%;
}

.timeline-rail {
    position: relative;
    width: 16px;
    min-width: 16px;
    flex-shrink: 0;
}

.timeline-rail::before {
    content: '';
    position: absolute;
    left: 50%;
    top: 0;
    bottom: 0;
    width: 2px;
    transform: translateX(-50%);
    background-color: var(--light-steel);
}

.timeline-rail-first::before {
    top: var(--first-dot-top, 0);
}

.timeline-rail-last::before {
    bottom: auto;
    height: var(--last-dot-height, 100%);
}

.timeline-dot {
    position: absolute;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background-color: var(--light-steel);
    z-index: 1;
}

.timeline-dot-active {
    width: 10px;
    height: 10px;
    background-color: var(--solid-blue-steel);
    border-radius: 2px;
    transform: translate(-50%, -50%) rotate(45deg);
}

.checkpoint-data {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
    gap: .25rem;
}
</style>
