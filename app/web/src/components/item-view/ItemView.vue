<script lang="ts" setup>
// vendor imports
import { onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
// custom imports
import type { PosterResult } from "../../util/request";
import { CartStore } from "../../store";
import { request } from "../../util/request";
import Gallery from "./Gallery.vue";
import Spinner from "./Spinner.vue";
// props
// state
const route = useRoute("poster");
const data = ref<PosterResult>();
const disabled = ref(false);

watch(CartStore.len, () => {
    disabled.value = CartStore.has(<string>route.params.id);
}, { immediate: true });

onMounted(async () => {
    const res = await request(`/api/v1/posters/${route.params.id}`, "GET");
    if (res.error) {
        console.log(res.error);
        return;
    }
    data.value = res.data[0];
});
</script>
<template>
    <section :class="$style.view">
        <Gallery v-if="data" :data>
            <button v-if="disabled" :class="[$style.button]" :disabled>
                <span>Added to cart</span>
            </button>
            <button v-else :class="[$style.button, $style.accept]" :disabled @click="CartStore.add(data)">
                <span :class="$style.label">{{ data.cost.rawTotal }}&#163;</span>
                <hr :class="$style.line">
                <span>Add to cart</span>
            </button>
        </Gallery>
        <Spinner v-else />
    </section>
</template>

<style lang="css" module>
.view {
    position: relative;
    width: min(1280px, 100%);
    height: 100%;
    margin: auto;
    margin-top: 1rem;
    padding: 1rem;
    box-shadow: -1px 1px 4px 0 var(--card-border-color);
    background: var(--primary);
    color: var(--txt-primary-color);
}

.button {
    width: 256px;
    height: 100%;
    min-height: calc(5rem + 5px);
    margin-left: auto;
    padding: 0.5rem 0.25rem;
    font-weight: 500;
    border-radius: var(--btn-border-radius);
    border: var(--btn-border-size) solid var(--btn-border-color);
    background: var(--btn-background);
    color: var(--btn-color);
}

.button[disabled] {
    background: none;
    cursor: not-allowed;
    color: #000;
}

.button.accept .label {
    color: #D4AF37;
}


@media screen and (min-width: 768px) {
    .view {
        border-radius: var(--card-border-radius);
    }
}
</style>
