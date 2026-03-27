<script setup lang="ts">
defineProps<{
    size?: number
}>()
</script>

<template>
    <div class="spinner" :style="{ width: size + 'px', height: size + 'px' }" />
</template>

<style scoped>
.spinner {
    border: 2px solid #e5e7eb;
    border-top: 2px solid #111827;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
}

@keyframes spin {
    to {
        transform: rotate(360deg);
    }
}
</style>