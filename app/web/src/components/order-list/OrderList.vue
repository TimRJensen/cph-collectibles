<script lang="ts" setup>
// vendor imports
import { ref, useTemplateRef, watch } from 'vue';
// custom imports
import { CartStore } from '../../store';
import IconButton from '../icon-button/IconButton.vue';
import type { PosterResult } from '../../util/request';
// props
withDefaults(defineProps<{ showTotal?: boolean }>(), {
    showTotal: false
})
const popover = useTemplateRef("popover");
defineExpose({ dom: popover });
// state
const base = import.meta.env.VITE_BASE_URL;
const items = ref<Array<PosterResult>>([]);
watch(CartStore.len, () => {
    items.value = CartStore.read();
}, { immediate: true });
</script>

<template>
    <div :class="$style.container">
        <div v-for="data in items" :class="$style.list">
            <img :class="$style.img" :src="base + data.files[0]?.url" />
            <span :class="$style.label">{{ data.detail.heading }}</span>
            <span :class="[$style.label, $style.price]">{{ data.cost.rawTotal }}&#163;</span>
            <IconButton :class="$style.delete" type="delete" fill="#1e1b18" @click="CartStore.remove(data)" />
        </div>
        <div v-if="showTotal" :class="$style.total">
            <span>Total</span><br><span>{{ CartStore.total() }}</span>&#163;
        </div>
    </div>
</template>

<style lang="css" module>
.container {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    align-content: flex-start;
}

.slot {
    font-size: 1rem;
    font-weight: 900;
}

.line {
    width: 100%;
    border: 1px solid #DADDD8;
}

.list {
    display: grid;
    grid-template-columns: 20% auto auto;
    width: 100%;
    min-width: 100%;
    height: fit-content;
    column-gap: 0.5rem;
    padding: 1rem 0.5rem;
    border: 1px solid #e6e6e6;
    border-radius: 5px;
    border-bottom: 1px solid #e6e6e6;
    background: #FFF;
}

.img {
    grid-column: 1;
    grid-row: span 2;
    width: 100%;
    height: auto;
    border-radius: 10px;
    object-fit: contain;
    aspect-ratio: 4/3;
}

.label {
    grid-column: 2;
    grid-row: 1 / span 2;
    text-align: left;
}

.label.price {
    grid-column: 3;
    grid-row: 2;
    text-align: right;
    font-weight: bold;
    place-self: end;
}

.total {
    margin-left: auto;
    text-align: right;
}

.delete {
    grid-column: 3;
    grid-row: 1;
    justify-self: end;
    align-self: start;
}


@media screen and (min-width: 768px) {
    .popover {
        width: 384px;
    }
}

@media screen and (min-width: 1024px) {
    .popover {
        width: 512px;
    }
}

@media screen and (min-width: 1920px) {
    .popover {
        width: 25%;
    }
}
</style>
