<script lang="ts" setup>
// vendor imports
import { ref, useTemplateRef } from "vue";
// custom imports
import type { PosterResult } from "../../util/request";
import IconButton from "../icon-button/IconButton.vue";
import Details from './Details.vue';
import Modal from "./Modal.vue";
// props
withDefaults(defineProps<{ data?: PosterResult }>(), {});
// state
const modal = useTemplateRef("modal")
const focus = ref(0);
</script>
<template>
    <div :class="$style.container">
        <div :class="$style.controlgroup">
            <IconButton type="arrows-out" style="place-self: flex-end;" @click="modal?.$el.showModal()" />
        </div>
        <div :class="$style.imgbox">
            <img :class="$style.img" :src="data?.files[focus]?.url" />
        </div>
        <div :class="$style.list">
            <button v-for="(f, i) in data?.files" :class="[$style.button]" @click="focus = i">
                <img :class="$style.img" :src="f.url" />
            </button>
        </div>
        <Details :class="$style.details" :data />
        <div :class="$style.slot">
            <slot :class="$style.slot"></slot>
        </div>
        <Modal ref="modal" :src="data?.files[focus]?.url!" />
    </div>
</template>

<style lang="css" module>
.container {
    display: grid;
    grid-template-columns: 100%;
    gap: 1rem;
}

.controlgroup {
    justify-self: center;
    display: flex;
    position: relative;
    justify-content: end;
    width: 100%;
}

.list {
    display: flex;
    gap: 0.5rem;
    align-items: center;
}

.button {
    display: inline;
    height: fit-content;
    border-radius: 5px;
    outline: none;
}

.button:focus,
.button:focus-visible,
.button:hover {
    scale: 1.025;
}

.imgbox {
    height: fit-content;
    overflow: hidden;
    background: #000;
}

.imgbox .img {
    width: auto;
    height: 512px;
    margin: auto;
    object-fit: contain;
    aspect-ratio: unset;
    margin: auto;
}

.img {
    width: 92px;
    height: auto;
    object-fit: cover;
    aspect-ratio: 4/3;
}

@media screen and (min-width: 768px) {
    .container {
        grid-template-columns: 128px auto;
    }

    .controlgroup {
        grid-column: 1 / span 2;
        grid-row: 1;
    }

    .list {
        grid-column: 1;
        grid-row: 2;
    }

    .imgbox {
        grid-column: 2;
        grid-row: 2;
    }

    .details {
        grid-column: 1 / span 2;
        grid-row: 3;
    }

    .slot {
        grid-column: 2;
        grid-row: 4;
        place-self: end;
        width: 256px;
    }
}

@media screen and (min-width: 768px) {
    .container {
        grid-template-columns: auto 256px;
    }

    .controlgroup {
        grid-column: 1;
        grid-row: 1;
    }

    .list {
        grid-column: 1;
        grid-row: 3;
    }

    .imgbox {
        grid-column: 1;
        grid-row: 2;
    }

    .details {
        grid-column: 2;
        grid-row: 2;
    }

    .slot {
        grid-column: 2;
        grid-row: 3;
        place-self: end;
        width: 256px;
    }
}

@media screen and (min-width: 1280px) {
    .container {
        grid-template-columns: 128px auto 256px;
    }

    .controlgroup {
        grid-column: 2 / span 2;
        grid-row: 1;
    }

    .list {
        grid-column: 1;
        grid-row: 2;
        align-items: unset;
    }

    .imgbox {
        grid-column: 2 / span 2;
        grid-row: 2;
    }

    .details {
        grid-column: 1 / span 2;
        grid-row: 3;
    }

    .list .button {
        width: 100%;
        height: fit-content;
    }

    .list .img {
        width: 100%;
        height: auto;
    }

    .slot {
        grid-column: 3;
        grid-row: 3;
    }
}
</style>
