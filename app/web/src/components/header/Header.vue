<script lang="ts" setup>
import ControlBar from '../control-bar/ControlBar.vue';

// vendor imports
// custom imports
// props
// state
</script>

<template>
    <header :class="$style.container">
        <div :style="{ placeSelf: 'center' }">
            <img :class="$style.img" src="../../assets/cph-collectables.png" />
        </div>
        <ControlBar :class="$style.controlbar" />
    </header>
</template>

<style lang="css" module>
.container {
    position: sticky;
    z-index: 100;
    display: flex;
    flex-flow: row nowrap;
    justify-content: flex-start;
    align-items: center;
    top: 0;
    left: 0;
    padding: 1rem 0;
    border-bottom: 2px solid #E8E6E3;
    background: #FFF;
}

.container::before {
    content: "";
    position: absolute;
    z-index: -1;
    width: 100%;
    height: 100%;
    background: #FFF;
}

.img {
    display: block;
    width: 128px;
    height: auto;
}

.controlbar {
    position: absolute;
    width: 50%;
    top: 50%;
    right: 0;
    translate: 0 -50%;
}


@media screen and (min-width: 768px) {
    .img {
        width: 256px;
    }
}

@media screen and (min-width: 1024px) {
    .container {
        justify-content: center;
    }

    .img {
        width: 256px;
    }

    .controlbar {
        width: calc(50% - 128px - 2rem);
        translate: unset;
    }
}

@media screen and (min-width: 1920px) {
    /* .img {
        width: 256px;
    } */
}
</style>
