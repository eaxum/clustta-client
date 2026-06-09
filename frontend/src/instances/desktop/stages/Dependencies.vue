<template>
	<div class="page-list-root absolute-pane">
        <DependencyGraph />
	</div>
</template>


<script setup>
import DependencyGraph from '@/instances/desktop/components/DependencyGraph.vue';

</script>

<style scoped>
@import "@/assets/desktop.css";

.page-list-root{
	box-sizing: border-box;
	/* padding: .4rem; */
	display: flex;
	align-items: center;
	justify-content: center;
	color: hsl(var(--foreground));
}
.absolute-pane{
  padding: 0px;
}

</style>

