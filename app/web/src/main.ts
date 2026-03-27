import { createApp } from "vue";
import { createWebHistory, createRouter } from "vue-router";
import App from "./App.vue";
import GridView from "./components/grid-view/GridView.vue";
import ItemView from "./components/item-view/ItemView.vue";
import CheckoutView from "./components/checkout-view/CheckoutView.vue";

const routes = [
    { path: "/", component: GridView },
    {
        path: "/poster/:id",
        name: "poster",
        component: ItemView
    },
    {
        path: "/checkout",
        name: "checkout",
        component: CheckoutView
    },
];

createApp(App).use(createRouter({ history: createWebHistory(), routes })).mount('#app');
