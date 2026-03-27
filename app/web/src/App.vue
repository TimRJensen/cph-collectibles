<script setup lang="ts">
// vendor imports
// custom imports
import GridView from './components/grid-view/GridView.vue';
// props
// state
</script>

<template>
    <main :class="$style.main" ref="main">
        <router-view />
    </main>
</template>

<style css module>
.main {
    width: 100%;
    height: 100%;
    margin: auto;
    padding: 0 10px;

}

@media screen and (min-width: 768px) {
    .main {
        padding: 0 2rem;
    }
}


@media screen and (min-width: 1024px) {
    .main {
        width: 90dvw;
    }
}

@media screen and (min-width: 1536px) {
    .main {
        width: 66vw;
    }
}
</style>
