import { createApp } from "vue";
import ControlBar from "../components/control-bar/ControlBar.vue";
import { Carousel } from "../elements/carousel";
createApp(ControlBar).mount('#controlbar');
Carousel.init();
Carousel.mount(document.querySelector("tmpl-carousel"));
