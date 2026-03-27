<script lang="ts" setup>
// vendor imports
// custom imports
import { useTemplateRef } from 'vue';
import IconButton from '../icon-button/IconButton.vue';
// props
// state
const modal = useTemplateRef("modal");
</script>

<template>
    <nav>
        <slot :modal="modal"></slot>
        <dialog :class="$style.modal" ref="modal" popover>
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
    </nav>
</template>

<style lang="css" module>
.modal {
    width: 100%;
    max-width: unset;
    height: fit-content;
    margin: 0 auto;
    padding: 2rem 0;
    color: #FFF;
    left: anchor(--header left);
    top: anchor(--header bottom);
    background: var(--secondary);
}

.modal::backdrop {
    left: anchor(--header left);
    top: anchor(--header bottom);
    position: absolute;
    background: var(--secondary);
    width: 100%;
    max-width: unset;
    height: 256px;
    box-shadow: -2px 2px 5px var(--card-border-color);
}

.modal:popover-open {
    position: relative;
    display: grid;
    grid-template-columns: repeat(2, 33%);
    place-items: start;
    gap: 1rem;
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
}

.label {
    font-weight: 700;
    margin: 0;
    margin-bottom: 0.5rem;
    border-bottom: 2px solid var(--primary);
}

.close {
    position: absolute;
    top: 1rem;
    right: 1rem;
}

@media screen and (min-width: 1536px) {
    .modal {
        width: min(1280px, 50vw);
    }
}
</style>
