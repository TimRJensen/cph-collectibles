<script lang="ts" setup>
// vendor imports
import { nextTick, ref, useTemplateRef } from "vue";
// custom imports
import type { PosterResult } from "../../util/request";
// props
const $props = withDefaults(defineProps<{ data: PosterResult }>(), {});
// state
const card = useTemplateRef("card")
const cols = ref(1);
const rows = ref(1);
async function onLoad(_e: Event) {
    await nextTick();

    if (!card.value) {
        return;
    }
    const trg = card.value;
    const ratio = trg.naturalWidth / trg.naturalHeight;
    if (ratio < 0.85) {
        // portrait
        rows.value = 2;
    }
}
</script>
<template>
    <router-link :class="$style.container" :style="{ gridColumn: `span ${cols}`, gridRow: `span ${rows}` }"
        :to="{ name: 'poster', params: { id: data.id } }">
        <button :class="$style.button">
            <div :class="$style.media">
                <img ref="card" :class="$style.img" :src="data.files[0]?.url" @load="onLoad" />
            </div>
            <!-- <span :class="$style.label">{{ data.name }}</span> -->
            <!-- <span>{{ data.caption }}</span> -->
        </button>
    </router-link>
</template>

<style lang="css" module>
.container {
    /* display: block; */
    aspect-ratio: 5 / 7;
}

.button {
    display: flex;
    flex-flow: column;
    justify-content: space-evenly;
    width: 100%;
    height: 100%;
    border-radius: 10px;
    box-shadow: -1px 1px 4px 0 #b5b8bc;
    padding: 0.5rem 0;
    background: #FFFDF8;
    color: #000;
}

.media {
    width: 100%;
    overflow: hidden;
}

.img {
    display: block;
    max-width: 100%;
    max-height: 100%;
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: scale, filter;
    transition-duration: 250ms;
}

.img:hover {
    scale: 1.33;
    filter: brightness(1.03) contrast(1.02);
}

.label {
    font-weight: bold;
}
</style>
