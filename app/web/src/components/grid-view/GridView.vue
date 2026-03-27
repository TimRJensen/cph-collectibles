<script lang="ts" setup>
// vendor imports
import { onMounted, ref } from "vue";
// custom imports
import type { PosterResult } from "../../util/request";
import { request } from "../../util/request";
import Grid from "../grid/Grid.vue";
import Spinner from "./Spinner.vue";
// props
// state
const data = ref<Array<PosterResult>>();
onMounted(async () => {
    const res = await request("/api/v1/posters", "GET");
    if (res.error) {
        console.log(res.error);
        return;
    }
    data.value = res.data;
});
</script>
<template>
    <section :class="$style.view">
        <Grid v-if="data" :data="data" />
        <Spinner v-else></Spinner>
    </section>
</template>

<style lang="css" module>
.view {
    display: grid;
    width: 100%;
    height: 100%;
    margin: 0 auto;
    place-items: center;
}
</style>
