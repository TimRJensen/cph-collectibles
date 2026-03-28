import type { PosterResult } from "../util/request";
import { request } from "../util/request";
import stylesUrl from "./carousel.css?url";

const tmpl = document.createElement("template");
const view = document.createElement("div");
const list = document.createElement("div");
view.id = "view";
list.id = "list";
tmpl.id = "carousel-template";
view.append(list);
tmpl.content.append(view);

class Suspense {
    #time: number = 0;
    #reject: ((why: any) => void) | null = null;

    constructor(time?: number) {
        this.#time = time ?? 0;
    }

    start(): Promise<boolean> {
        const { promise, resolve, reject } = Promise.withResolvers<boolean>();
        this.#reject = reject;

        const now = performance.now();
        const fn = (time: number) => {
            if (time - now < this.#time) {
                requestAnimationFrame(fn);
                return;
            }
            resolve(true);
        }
        fn(now);

        return promise;
    }

    cancel() {
        if (!this.#reject) {
            return;
        }
        this.#reject(null)
    }
}

export class Carousel extends HTMLElement {
    static get observedAttributes() {
        return ["src", "for", "duration", "suspense"];
    }
    static #registry = new CustomElementRegistry();
    #duration: number = 0;
    #suspense: Suspense | null = null;
    #root: ShadowRoot | null = null;
    #trg: HTMLElement | null = null;
    #animation: Animation | null = null;

    constructor() {
        super();
    }

    static init() {
        Carousel.#registry.define("tmpl-carousel", Carousel);
        window.customElements.define("tmpl-carousel", Carousel)
    }

    static mount(n: Node | null) {
        if (!n) {
            return;
        }
        Carousel.#registry.upgrade(n);
    }

    async #play(idx: number): Promise<void> {
        if (!this.#animation || !this.#suspense) {
            return;
        }

        if (this.#animation.playState == "running") {
            this.#animation.cancel();
            this.#suspense.cancel();
        } else {
            this.#suspense.cancel();
        }

        const list = this.#root?.querySelector("#list");
        if (!list) {
            return;
        }

        const child = list.children.item(idx);
        if (!child) {
            return;
        }

        const gap = parseFloat(getComputedStyle(list).columnGap || "0");
        const rect = child.getBoundingClientRect();
        const offset = rect.width + (isNaN(gap) ? 0 : gap);
        this.#animation.effect = new KeyframeEffect(
            list,
            [
                { translate: "0" },
                { translate: `-${(idx + 1) * offset}px 0` }
            ],
            {
                duration: this.#duration,
                easing: "ease-in-out",
                fill: "backwards",
            }
        );
        // try animation
        try {
            this.#animation.play();
            await this.#animation.finished;
        } catch {
            return;
        }

        if (idx == 0) {
            this.#trg?.replaceChildren(child.cloneNode(true));
            list.appendChild(list.removeChild(child));
        } else {
            idx++;
            const children = Array.from(list.children);
            this.#trg?.replaceChildren(children[idx]!.cloneNode(true));
            list.replaceChildren(
                ...children.slice(idx),
                ...children.slice(0, idx)
            );
        }

        // try suspense
        try {
            await this.#suspense?.start();
        } catch {
            return;
        }

        this.#play(0);
    }

    #onclick = (e: PointerEvent) => {
        const list = <HTMLImageElement>e.currentTarget;
        const child = <HTMLImageElement>e.target;
        if (list == child) {
            return;
        }
        const children = Array.from(list.children);
        const idx = children.indexOf(child);

        this.#play(idx - 1);
    }

    #update(data: Array<PosterResult>): void {
        const children = <Array<HTMLImageElement>>[];
        for (const entry of data) {
            const img = document.createElement("img");
            img.className = "img";
            img.src = entry.files[0]?.url ?? "";
            children.push(img);
        }

        const list = <HTMLElement>this.#root?.querySelector("#list");
        if (!list) {
            return;
        }
        list.append(...children);
    }

    async #fetch(_url: string): Promise<void> {
        const res = await request("/api/v1/inventory?random=10", "GET");
        if (res.error) {
            console.log(res.msg);
            return;
        }
        this.#update(res.data);
        this.#play(0);
        return;
    }

    connectedCallback() {
        // attach shadowroot
        this.#root = this.attachShadow({
            mode: "open",
            //customElementRegistry: Carousel.#registry,
        });
        // default styles
        const link = document.createElement("link");
        link.setAttribute("rel", "stylesheet");
        link.setAttribute("href", stylesUrl);
        this.#root.appendChild(link);
        this.#root.appendChild(document.importNode(tmpl.content, true));
        this.#animation = new Animation();

        const list = <HTMLElement>this.#root?.querySelector("#list");
        if (!list) {
            return;
        }
        list.onpointerdown = this.#onclick;
        this.#fetch("");

        console.log("foo")
    }

    attributeChangedCallback(name: string, prev: string, next: string): void {
        if (prev == next) {
            return;
        }
        console.log("bar")
        switch (name) {
            case "src":
                return;
            case "for":
                const trg = document.getElementById(next);
                if (!trg) {
                    return;
                }
                this.#trg = trg;
                return;
            case "duration":
                this.#duration = Number(next);
                return;
            case "suspense":
                this.#suspense = new Suspense(Number(next));
                return;
            default:
                return;
        }
    }
}