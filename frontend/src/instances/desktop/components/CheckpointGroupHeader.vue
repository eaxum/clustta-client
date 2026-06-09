<template>
    <div class="checkpoint-group-header">
        <span class="checkpoint-group-label">{{ label }}</span>
    </div>
</template>

<script setup>
// props
const props = defineProps({
    label: {
        type: String,
        required: true
    }
});
</script>

<style scoped>
.checkpoint-group-header {
    padding: .5rem .25rem;
}

.checkpoint-group-label {
    font-family: 'Inter', sans-serif;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: hsl(var(--muted-foreground));
}
</style>
