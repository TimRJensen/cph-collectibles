<script lang="ts" setup>
// vendor imports
// custom imports
import { CartStore } from '../../store';
import { computed, onMounted, reactive, useTemplateRef } from 'vue';
import IconButton from '../icon-button/IconButton.vue';
import Cart from '../cart/Cart.vue';
import BurgerMenu from './BurgerMenu.vue';
import NavBar from './NavBar.vue';
// props
// state
const demo = useTemplateRef("demo-banner");
const input = useTemplateRef("input");
const popover = useTemplateRef("popover");
const css = reactive(new Map<string, string>([["width", "0"]]));
const width = computed({
    get() {
        return `${css.get("width")}%`
    },
    set(v) {
        css.set("width", v);
    }
});
const items = computed(() => {
    if (CartStore.len.value) {
        return `"${CartStore.len.value}""`;
    }
    return "";
});

onMounted(() => {
    demo.value?.togglePopover();
    /// ts-ignore
    document.activeElement?.blur();
});
</script>

<template>
    <div :class="$style.container">
        <NavBar :class="$style.navbar">
            <template v-slot="{ modal }">
                <IconButton :class="$style.navbutton" type="caret-down" orientation="horizontal" flow="reverse"
                    @click="modal?.togglePopover()">
                    EXPLORE
                </IconButton>
            </template>
        </NavBar>
        <form :class="$style.search" @focusin="width = '100'" @focusout="width = '0'">
            <input :class="$style.input" ref="input" type="text">
            <IconButton type="search" :width="32" :height="32" @click="input?.focus()" />
        </form>
        <IconButton :class="$style.cart" type="cart" :width="32" :height="32" @click="popover?.dom?.togglePopover()" />
        <BurgerMenu>
            <template v-slot="{ modal }">
                <IconButton :class="$style.menubutton" type="list" :width="32" :height="32"
                    @click="modal?.showModal()" />
            </template>
        </BurgerMenu>
        <Cart ref="popover" />
        <dialog class="demo-banner" ref="demo-banner" popover>
            This website is in <strong>demonstration mode</strong>.<br>
            Any purchases made here are <strong>for testing only</strong> and will not be charged.
        </dialog>
    </div>
</template>

<style lang="css" module>
.container {
    --items: "1";
    anchor-name: --controlbar;
    position: sticky;
    display: flex;
    flex-flow: row nowrap;
    justify-content: flex-end;
    align-items: center;
    width: 50100%;
    min-width: 100%;
    top: 0;
    gap: 0.5rem;
}

.search {
    display: none;
    width: fit-content;
    margin-left: auto;
    padding: 0.25rem 0.5rem;
    border: 1px solid transparent;
    border-radius: 10px;
}

.search:focus-within,
.search:focus-visible {
    width: v-bind(width);
    border-color: #b5b8bc;
    background: #FFFDF8;
    transition: border-color, background;
    transition-duration: 500ms;
}

.input {
    width: v-bind(width);
    height: 24px;
    border: none;
    outline: none;
    background: none;
    color: #2A2622;
}

.cart {
    position: relative;
}

.cart::after {
    position: absolute;
    content: v-bind(items);
    width: 18px;
    height: 18px;
    top: -0.66rem;
    right: -0.66rem;
    border-radius: 50%;
    background-color: #000;
    color: #FFF;
    line-height: 18px;
    font-size: 0.66rem;
    font-weight: bolder;
}

.navbar {
    display: none;
}

.navbutton>svg {
    transition: rotate 128ms ease-in-out;
}

.navbutton:has(+ *:popover-open)>svg {
    rotate: 180deg;
}

.close {
    position: absolute;
    top: 1rem;
    right: 1rem;
}

@media screen and (min-width: 1024px) {
    .container {
        width: 33vw;
    }

    .navbar {
        display: initial;
    }

    .menubutton {
        display: none;
    }

    .search {
        display: flex;
        justify-content: flex-end;
        align-items: center;
    }
}
</style>

<style lang="css">
.demo-banner {
    width: 100%;
    left: 0;
    top: 0;
    background-color: bisque;
    box-shadow: -1px 1px 4px 0 var(--card-border-color);
    text-align: center;
}
</style>
