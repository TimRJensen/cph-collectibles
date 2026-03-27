<script lang="ts" setup>
// vendor imports
// custom imports
import { computed, reactive, useTemplateRef } from 'vue';
import IconButton from '../icon-button/IconButton.vue';
// props
// state
const input = useTemplateRef("input");
const modal = useTemplateRef("modal");
const css = reactive(new Map<string, string>([["width", "0"]]));
const width = computed({
    get() {
        return `${css.get("width")}%`
    },
    set(v) {
        css.set("width", v);
    }
});
</script>

<template>
    <slot :modal="modal"></slot>
    <dialog :class="$style.modal" ref="modal">
        <form :class="$style.search" @focusin="width = '100'" @focusout="width = '0'">
            <input :class="$style.input" ref="input" type="text">
            <IconButton type="search" :width="32" :height="32" @click="input?.focus()" />
        </form>
        <div :class="$style.group">
            <p :class="$style.label">EXPLORE</p>
            <a :class="$style.link" href="/shop/">SHOP</a>
            <a :class="$style.link" href="/shop/">AUTOMOBILIA</a>
        </div>
        <div :class="$style.group">
            <p :class="$style.label">COPENHAGEN COLLECTABLES</p>
            <a :class="$style.link" href="/">HOME</a>
            <a :class="$style.link" href="/about/">ABOUT US</a>
            <a :class="$style.link" href="/contact/">CONTACT</a>
            <a :class="$style.link" href="">COOKIE SETTINGS</a>
        </div>
        <IconButton :class="$style.close" type="close" @click="modal?.close()" />
    </dialog>
</template>

<style lang="css" module>
.modal {
    width: 100%;
    max-width: unset;
    height: 100%;
    max-height: unset;
    padding: 2rem 3rem;
}

.modal[open] {
    display: flex;
    flex-flow: column nowrap;
    justify-content: flex-start;
    align-items: flex-end;
    gap: 1rem;
}

.search {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    width: fit-content;
    margin-left: auto;
    padding: 0.25rem 0.5rem;
    border: 1px solid transparent;
    border-radius: 10px;
}

.search:focus-within,
.search:focus-visible {
    width: v-bind(width);
    border-color: var(--input-border-color);
    background: var(--primary);
    transition: border-color, background;
    transition-duration: 500ms;
}

.input {
    width: v-bind(width);
    height: 24px;
    border: none;
    outline: none;
    background: none;
    color: #000;
}

.group {
    display: flex;
    flex-wrap: wrap;
    width: 100%;
    min-width: 100%;
}


.label,
.link {
    width: 100%;
    min-width: 100%;
    line-height: 1.75;
    text-align: right;
}

.label {
    font-weight: 700;
    margin: 0;
    margin-bottom: 0.5rem;
    border-bottom: 2px solid var(--secondary);
}

.close {
    position: absolute;
    top: 1rem;
    right: 1rem;
}

@media screen and (min-width: 768px) {
    .modal {
        display: none;
    }
}
</style>
