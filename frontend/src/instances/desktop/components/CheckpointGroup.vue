<template>
    <div class="checkpoint-group">

        <div ref="railRef" class="timeline-rail">
            <div class="timeline-segment" v-for="segment in railSegments" :key="segment.key"
                :style="{ top: segment.top, height: segment.height, background: segment.background }">
            </div>

            <div class="timeline-dot" v-for="checkpoint in group.items" :key="checkpoint.checkpoint_id"
                :class="{ 'timeline-dot-active': checkpoint.hash === taskHash, 'timeline-dot-alert': !checkpoint.synced }"
                :style="{ top: dotPositions[checkpoint.checkpoint_id] }" v-tooltip="checkpoint.synced ? 'Synced' : 'Not synced'">
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
const railSegments = ref([]);

// Stores a reference to each CheckpointItem element.
const setItemRef = (checkpointId, el) => {
    if (el) {
        itemRefs[checkpointId] = el;
    }
};

// Calculates the vertical position of each dot and rail segments.
const calculateDotPositions = () => {
    nextTick(() => {
        if (!railRef.value) return;
        const railRect = railRef.value.getBoundingClientRect();
        const railTop = railRect.top;
        const railHeight = railRect.height;
        const positions = {};
        const dots = [];

        for (const checkpoint of props.group.items) {
            const itemEl = itemRefs[checkpoint.checkpoint_id];
            if (itemEl && itemEl.$el) {
                const rect = itemEl.$el.getBoundingClientRect();
                const top = rect.top - railTop + (rect.height / 2);
                positions[checkpoint.checkpoint_id] = `${top}px`;
                dots.push({ top, synced: checkpoint.synced });
            }
        }
        dotPositions.value = positions;

        const segments = [];
        const danger = getComputedStyle(railRef.value).getPropertyValue('--danger').trim();
        const steel = getComputedStyle(railRef.value).getPropertyValue('--light-steel').trim();

        // Returns the background value for a segment given the sync state at each end.
        const segmentBg = (startSynced, endSynced) => {
            if (!startSynced && !endSynced) return danger;
            if (startSynced && endSynced) return steel;
            const from = startSynced ? steel : danger;
            const to = endSynced ? steel : danger;
            return `linear-gradient(to bottom, ${from}, ${to})`;
        };

        // Segment from rail top to first dot (non-first groups)
        if (!props.isFirstGroup && dots.length > 0) {
            const synced = dots[0].synced;
            segments.push({ key: 'top', top: '0px', height: `${dots[0].top}px`, background: segmentBg(synced, synced) });
        }

        // Segments between consecutive dots
        for (let i = 0; i < dots.length - 1; i++) {
            segments.push({
                key: `seg-${i}`,
                top: `${dots[i].top}px`,
                height: `${dots[i + 1].top - dots[i].top}px`,
                background: segmentBg(dots[i].synced, dots[i + 1].synced)
            });
        }

        // Segment from last dot to rail bottom (non-last groups)
        if (!props.isLastGroup && dots.length > 0) {
            const lastDot = dots[dots.length - 1];
            segments.push({ key: 'bottom', top: `${lastDot.top}px`, height: `${railHeight - lastDot.top}px`, background: segmentBg(lastDot.synced, lastDot.synced) });
        }

        railSegments.value = segments;
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

.timeline-segment {
    position: absolute;
    left: 50%;
    width: 2px;
    transform: translateX(-50%);
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

.timeline-dot-alert {
    background-color: var(--danger);
}

.checkpoint-data {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
    gap: .25rem;
}
</style>
