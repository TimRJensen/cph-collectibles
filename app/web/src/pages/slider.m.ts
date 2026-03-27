const viewer = document.getElementById("viewer");
const slider = document.getElementById("slider");

function update(next?: HTMLImageElement) {
    if (!viewer) {
        return;
    }
    const child = <HTMLImageElement>viewer.firstElementChild;
    if (!child) {
        return;
    }

    if (!next) {
        return;
    }
    child.src = next.src;
}

function unset(child: HTMLImageElement): TimerHandler {
    if (!slider) {
        return (() => undefined);
    }
    return () => {
        const style = slider.computedStyleMap();
        const prev = style.get("transition")?.toString();
        requestIdleCallback(() => {
            slider.style.setProperty("transition", prev!);
        });

        slider.appendChild(slider.removeChild(child));
        slider.style.setProperty("transition", "unset");
        slider.style.setProperty("translate", `0`);
        requestAnimationFrame(tick);
    };
}

let raf = 0;
let time = performance.now();
function tick(now: number) {
    if (now - time < 5000) {
        raf = requestAnimationFrame(tick);
        return;
    }
    time = now;

    if (!slider) {
        return;
    }
    const n = Number(slider.dataset["idx"]) + 1;

    const child = <HTMLImageElement>slider.firstElementChild;
    const next = <HTMLImageElement>slider.querySelector(`:nth-child(${n})`);
    if (!child || !next) {
        return;
    }
    update(next);

    const rect = child.getBoundingClientRect();
    const tx = rect.width;
    slider.style.setProperty("translate", `-${n * tx}px 0`);
    setTimeout(unset(child), 260);
}
tick(time);

function init() {
    if (!slider) {
        return;
    }

    slider.onclick = (e) => {
        console.log("foo")
        let i = 0;
        while (slider.children[i] != e.target) {
            i++;
        }
        slider.dataset["idx"] = i.toString();
        cancelAnimationFrame(raf);
        tick(time + 5000);
    }
}
init();

