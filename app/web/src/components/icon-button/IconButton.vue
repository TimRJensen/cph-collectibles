<script lang="ts" setup>
// vendor imports
// custom imports
import { computed } from "vue";
import * as svgs from "./svgs/";

const icons = {
    "arrows-in": svgs.ArrowsIn,
    "arrows-out": svgs.ArrowsOut,
    "card": svgs.Card,
    "caret-down": svgs.CaretDown,
    "cart": svgs.Cart,
    "check-square": svgs.CheckSquare,
    "close": svgs.Close,
    "delete": svgs.Delete,
    "list": svgs.List,
    "package": svgs.Package,
    "search": svgs.Search,
    "zoom-in": svgs.ZoomIn,
    "zoom-out": svgs.ZoomOut,
}

// props
const $props = withDefaults(defineProps<{
    type: keyof typeof icons;
    disabled?: boolean;
    width?: number | string;
    height?: number | string;
    orientation?: "vertical" | "horizontal";
    flow?: "normal" | "reverse";
    fill?: string;
    foo?: { a: number }
}>(), {
    disabled: false,
    width: 24,
    height: 24,
    orientation: "vertical",
    flow: "normal",
    fill: "#000",
});
// state
const direction = computed(() => {
    if ($props.orientation == "vertical") {
        return $props.flow == "reverse" ? "column-reverse" : "column";
    }
    return $props.flow == "reverse" ? "row-reverse" : "row";
});
const flow = computed(() => {
    return $props.flow == "normal" ? "column" : "row"
});
const width = computed(() => {
    return `${$props.width}${typeof $props.width == "number" ? "px" : ""}`;
});
const height = computed(() => {
    return `${$props.height}${typeof $props.width == "number" ? "px" : ""}`;
});
const fill = computed(() => {
    return $props.fill;
});
</script>

<template>
    <button ref="element" type="button" :class="$style.button" :disabled>
        <component :is="icons[type]" />
        <span :class="$style.label">
            <slot></slot>
        </span>
    </button>
</template>

<style lang="css" module>
.button {
    display: inline-flex;
    align-items: center;
    flex-direction: v-bind(direction);
    width: fit-content;
    color: v-bind(fill);
    fill: v-bind(fill);
    cursor: pointer;
}

.button[disabled] {
    color: #DADDD8;
    fill: #DADDD8;
    cursor: default;
}

.button>svg {
    width: v-bind(width);
    min-width: v-bind(width);
    height: v-bind(height);
}
</style>
