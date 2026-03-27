<script lang="ts" setup>
// vendor imports
import { useTemplateRef } from "vue";
// custom imports
import IconButton from "../icon-button/IconButton.vue";
import { DragController } from "./drag-controller";
// props
const modal = useTemplateRef("modal");
defineProps<{ src: string }>();
defineExpose({ modal });
// state
const imgbox = useTemplateRef("imgbox");
const dc = new DragController(imgbox);
const { scale, translate } = dc;
</script>

<template>
    <dialog :class="[$style.modal]" ref="modal" @pointerdown="">
        <div :class="$style.controlgroup">
            <IconButton type="zoom-in" @click="dc.zoom(0.1)" />
            <IconButton type="zoom-out" @click="dc.zoom(-0.1)" />
        </div>
        <div :class="$style.imgbox" ref="imgbox">
            <img :class="$style.img" :src />
        </div>
        <IconButton :class="$style.close" type="close" @click="modal?.close()" />
    </dialog>

</template>

<style lang="css" module>
.modal {
    width: max(1024px, 60vw);
    height: 100vh;
    opacity: 0;
    background: none;
    transition: opacity;
    transition-duration: 250ms;
}

.modal::backdrop {
    background: rgba(0, 0, 0, 0.4);
    transition: background, backdrop-filter;
    transition-duration: 250ms;
}

.modal[open] {
    display: flex;
    flex-flow: column nowrap;
    align-items: center;
    gap: 1rem;
    opacity: 1;

    @starting-style {
        opacity: 0;
    }
}

.modal[open]::backdrop {
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.4);
    backdrop-filter: blur(12px);

    @starting-style {
        background: rgba(0, 0, 0, 0.0);
        backdrop-filter: blur(0);
    }
}

.controlgroup {
    display: flex;
    gap: 0.5rem;
}

.imgbox {
    width: 95%;
    height: fit-content;
    overflow: hidden;
    touch-action: none;
    cursor: grab;
}

.imgbox[drag] {
    cursor: grabbing;
}


.img {
    width: 100%;
    height: auto;
    margin: auto;
    aspect-ratio: 4/3;
    object-fit: contain;
    will-change: transform;
    transform-origin: 0 0;
    translate: v-bind(translate);
    scale: v-bind(scale);
}

.close {
    position: absolute;
    top: 1rem;
    right: 1rem;
}
</style>