<script lang="ts" setup>
// vendor imports
import { ref, useTemplateRef, watch } from 'vue';
// custom imports
import IconButton from '../icon-button/IconButton.vue';
import OrderList from '../order-list/OrderList.vue';
import { CartStore } from '../../store';
// props
const popover = useTemplateRef("popover");
defineExpose({ dom: popover });
// state
const total = ref(0);
watch(CartStore.len, () => {
    total.value = CartStore.total();
}, { immediate: true });
</script>

<template>
    <dialog :class="$style.popover" ref="popover" popover>
        <span :class="$style.label">Items</span>
        <hr :class="$style.line">
        <OrderList />
        <hr :class="$style.line">
        <a :class="$style.link" href="/checkout/">
            <button :class="[$style.accept]" @click="popover?.togglePopover()">
                <span :class="$style.pricetag">{{ total }}&#163;</span>
                <hr :class="$style.line">
                Go to checkout
            </button>
        </a>
        <IconButton :class="$style.close" type="close" @click="popover?.togglePopover()" />
    </dialog>
</template>

<style lang="css" module>
.popover {
    inset: unset;
    box-sizing: unset;
    height: 100%;
    padding: 2rem 1rem;
    transition: opacity;
    transition-duration: 250ms;
    border-left: var(--card-border-size) solid var(--card-border-color);
    border-right: var(--card-border-size) solid var(--card-border-color);
    background: var(--primary);
}

.popover::backdrop {
    background: rgba(37, 27, 24, 0.4);
    transition: background, backdrop-filter;
    transition-duration: 250ms;
}

.popover:popover-open {
    display: flex;
    flex-flow: column nowrap;
    align-items: center;
    gap: 1rem;
    opacity: 1;
}

.popover:popover-open::backdrop {
    position: fixed;
    width: 100%;
    height: 100%;
    background: rgba(37, 27, 24, 0.4);
    backdrop-filter: blur(12px);
}

@starting-style {
    .popover:popover-open {
        opacity: 0;
    }

    .popover:popover-open::backdrop {
        background: rgba(37, 27, 24, 0.0);
        backdrop-filter: blur(0);
    }
}

.line {
    width: 100%;
    border: 1px solid #e6e6e6;
}

.link {
    display: block;
    width: 50%;
    margin-left: auto;
    text-decoration: none;
}

.accept {
    width: 100%;
    padding: 0.5rem 0.25rem;
    font-weight: bold;
    border-radius: var(--btn-border-radius);
    border: var(--btn-border-size) solid var(--btn-border-color);
    background: var(--btn-background);
    color: var(--btn-color);
}

.accept .line {
    background: #FFF;
    width: 80%;
}

.accept .pricetag {
    color: #D4AF37;
}

.close {
    position: absolute;
    top: 1rem;
    right: 1rem;
}

@media screen and (min-width: 768px) {
    .popover {
        right: anchor(--controlbar right);
        width: 384px;
    }

    .link {
        width: 256px;
    }
}

@media screen and (min-width: 1024px) {
    .popover {
        width: 512px;
    }
}
</style>
