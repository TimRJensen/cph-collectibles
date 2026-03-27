<script lang="ts" setup>
// vendor imports
// custom imports
import { computed } from "vue";
import type { PosterResult } from "../../util/request";
// props
const props = withDefaults(defineProps<{ data?: PosterResult }>(), {
    data: () => ({
        id: "",
        meta: {
            rawId: "",
            createdAt: "",
            updatedAt: "",
        },
        cost: {
            rawAmount: 0,
            rawVAT: 0,
            rawTotal: 0,
            minorAmount: 0,
            minorVAT: 0,
            minorToral: 0,
        },
        detail: {
            heading: "",
            body: "",
            width: 0,
            height: 0,
            origin: {
                source: "unknown",
                year: "unknown",
            }
        },
        condition: {
            rating: "unspecified",
            notes: "",
        },
        files: [],
    })
});
// state
const rating = computed(() => {
    if (!props.data) {
        return "unspecified";
    }
    const c = props.data.condition;
    return `${c.rating[0]?.toUpperCase()}${c.rating.slice(1)} - ${c.notes}`
});
</script>

<template>
    <div :class="[$style.details]">
        <span :class="$style.label">Title</span><span>{{ data.detail.heading }}</span>
        <span :class="$style.label">Year</span><span>{{ data.detail.origin.year }}</span>
        <span :class="$style.label">Rating</span><span>{{ rating }}</span>
        <span :class="$style.summary">
            {{ data.condition.notes }}
        </span>
    </div>
</template>


<style lang="css" module>
.details {
    display: grid;
    grid-template-columns: min-content auto;
    grid-auto-rows: min-content;
    column-gap: 1rem;
}

.label {
    font-weight: bold;
}

.summary {
    grid-column: span 2;
    overflow: hidden auto;
}
</style>