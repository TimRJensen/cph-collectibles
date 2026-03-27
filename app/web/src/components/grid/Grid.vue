<script lang="ts" setup>
// vendor imports
// custom imports
import type { PosterResult } from "../../util/request";
import GridItem from "./GridItem.vue";
// props
const $props = withDefaults(defineProps<{ data?: Array<PosterResult> }>(), {});
defineEmits<{ (e: "click", idx: number): void }>();
// state
</script>
<template>
    <div :class="$style.container" grid-root>
        <GridItem v-for="(data, _) in $props.data" :data />
    </div>
</template>

<style lang="css" module>
.container {
    display: grid;
    grid-auto-flow: row dense;
    grid-template-columns: repeat(2, 1fr);
    grid-auto-rows: 96px;
    gap: 8px;
    place-items: stretch;
}

@media screen and (min-width: 768px) {
    .container {
        grid-template-columns: repeat(2, 1fr);
        grid-auto-rows: 192px;
    }
}

@media screen and (min-width: 1024px) {
    .container {
        grid-template-columns: repeat(3, 1fr);
        grid-auto-rows: 192px;
    }
}

@media screen and (min-width: 1280px) {
    .container {
        grid-template-columns: repeat(3, 1fr);
        grid-auto-rows: 192px;
    }
}

@media screen and (min-width: 1536px) {
    .container {
        grid-template-columns: repeat(4, 1fr);
        grid-auto-rows: 192px;
    }
}
</style>
