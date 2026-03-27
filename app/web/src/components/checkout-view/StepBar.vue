<script lang="ts" setup>
// vendor imports
// custom imports
import IconButton from "../icon-button/IconButton.vue";
// props
const model = defineModel<number>({ required: true, default: 0 });
// state
</script>
<template>
    <div :class="$style.stepbar">
        <IconButton type="package" :width="32" :height="32" :disabled="model < 1" @click="model = 1">
            shipping
        </IconButton>
        <hr :class="$style.line">
        <IconButton type="card" :width="32" :height="32" :disabled="model < 2" @click="model = 2">
            payment
        </IconButton>
        <hr :class="$style.line">
        <IconButton type="check-square" :width="32" :height="32" :disabled="model < 3" @click="model = 3">
            confirm
        </IconButton>
    </div>
</template>

<style lang="css" module>
.stepbar {
    display: flex;
    width: 100%;
    min-width: 100%;
    align-items: center;
}

.line {
    width: 33%;
    min-width: unset;
    height: 0;
    border: unset;
    border-top: 1px solid #DADDD8;
    margin: 0 0.5rem;
}
</style>
