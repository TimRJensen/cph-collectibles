import { createApp } from "vue";
import CheckoutView from "../../components/checkout-view/CheckoutView.vue";
import ControlBar from "../../components/control-bar/ControlBar.vue";

createApp(ControlBar).mount('#controlbar');
createApp(CheckoutView).mount('#main');
