import { createApp } from "vue";
import { createWebHistory, createRouter } from "vue-router";
import ShopView from "../../components/shop-view/ShopView.vue";
import GridView from "../../components/grid-view/GridView.vue";
import ItemView from "../../components/item-view/ItemView.vue";
import ControlBar from "../../components/control-bar/ControlBar.vue";

const routes = [
    {
        path: "/shop/",
        name: "shop",
        component: GridView
    },
    {
        path: "/poster/:id",
        name: "poster",
        component: ItemView
    },
];

createApp(ControlBar).mount('#controlbar');
createApp(ShopView)
    .use(createRouter({ history: createWebHistory(), routes }))
    .mount('#main');
