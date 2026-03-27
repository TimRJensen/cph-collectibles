import { ref, watch, type Ref } from "vue";

class XY {
    #x: [number, number] = [0.0, 0.0];
    #y: [number, number] = [0.0, 0.0];

    set sx(v: number) {
        this.#x[0] = v;
    }
    get sx(): number {
        return this.#x[0];
    }
    set sy(v: number) {
        this.#y[0] = v;
    }
    get sy(): number {
        return this.#y[0];
    }

    set px(v: number) {
        this.#x[1] = v;
    }
    get px(): number {
        return this.#x[1];
    }
    set py(v: number) {
        this.#y[1] = v;
    }
    get py(): number {
        return this.#y[1];
    }
}

/**
 * DragController - Simple drag control, as suggested by Lis Larsen
 */
export class DragController {
    scale: Ref<number> = ref(1);
    translate: Ref<string> = ref("");
    #src: HTMLDivElement | null = null;
    #p: XY = new XY();

    #moveXY(px: number, py: number, ratio: number = 0.8): void {
        this.#p.px = px - (px - this.#p.px) * ratio;
        this.#p.py = py - (py - this.#p.py) * ratio;
        this.translate.value = `${this.#p.px}px ${this.#p.py}px`;
    }

    #scale(s: number): void {
        this.scale.value = s;
    }

    #wheel = (e: WheelEvent): void => {
        if (this.#src?.hasAttribute("drag")) {
            return;
        }
        e.stopPropagation();
        e.preventDefault();

        const rect = this.#src?.getBoundingClientRect()!;
        const px = e.clientX - rect.left;
        const py = e.clientY - rect.top;
        const prev = this.scale.value;
        const next = Math.max(
            1,
            Math.min(5, prev + (e.deltaY > 0 ? -0.1 : 0.1))
        );

        this.#scale(next);
        this.#moveXY(px, py, next / prev);
    }

    #pointerDown = (e: PointerEvent): void => {
        if (this.scale.value == 1.0) {
            return;
        }

        this.#src?.setAttribute("drag", "");
        this.#src?.setPointerCapture(e.pointerId);
        this.#p.sx = e.clientX - this.#p.px;
        this.#p.sy = e.clientY - this.#p.py;
    }

    #pointerMove = (e: PointerEvent): void => {
        if (!this.#src?.hasAttribute("drag")) {
            return;
        }
        this.#moveXY(e.clientX - this.#p.sx, e.clientY - this.#p.sy);
    }

    #pointerStop = (e: PointerEvent): void => {
        this.#src?.removeAttribute("drag");
        this.#src?.releasePointerCapture(e.pointerId);
    }

    zoom(s: number): void {
        if (!this.#src) return;

        const rect = this.#src.getBoundingClientRect();
        const px = rect.width / 2;
        const py = rect.height / 2;
        const prev = this.scale.value;
        const next = Math.max(
            1,
            Math.min(5, prev + s)
        );

        this.#scale(next);
        this.#moveXY(px, py, next / prev);
    }

    constructor(src: Ref<HTMLDivElement | null>) {
        watch(src, (next, prev) => {
            if (!next) {
                if (prev) {
                    prev.onwheel = null;
                    prev.onpointerdown = null;
                    prev.onpointermove = null;
                    prev.onpointerup = null;
                    prev.onpointercancel = null;
                }
                return;
            }
            this.#src = next;

            next.onwheel = this.#wheel;
            next.onpointerdown = this.#pointerDown;
            next.onpointermove = this.#pointerMove;
            next.onpointerup = this.#pointerStop;
            next.onpointercancel = this.#pointerStop;
        })
    }
}