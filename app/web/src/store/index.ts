// vendor imports
import { ref } from "vue";
// custom import
import type { PosterResult } from "../util/request";

const CART_KEY = "cph-cart";
export class CartStore {
    static #len = ref(0);

    static read(): Array<PosterResult> {
        try {
            const raw = localStorage.getItem(CART_KEY);
            if (!raw) {
                return [];
            }

            const parsed = JSON.parse(raw);
            if (Array.isArray(parsed)) {
                CartStore.#len.value = parsed.length;
                return parsed;
            }
            return [];
        } catch {
            return [];
        }
    }

    static write(items: Array<PosterResult>): void {
        localStorage.setItem(CART_KEY, JSON.stringify(items));
    }

    static add(item: PosterResult): Array<PosterResult> {
        const items = CartStore.read();
        if (!items.includes(item)) {
            CartStore.#len.value = items.push(item);
            CartStore.write(items);
        }
        console.log(items)
        return items;
    }

    static remove(item: PosterResult): Array<PosterResult> {
        const items = CartStore.read().filter((entry) => entry.id !== item.id);
        CartStore.#len.value = items.length;
        CartStore.write(items);
        return items;
    }

    static has(item: PosterResult | string): boolean {
        if (typeof item == "string") {
            return CartStore.read().some((entry) => entry.id == item);
        }
        return CartStore.read().some((entry) => entry.id == item.id);
    }

    static total(mode: "all" | "vat" = "all"): number {
        if (mode == "vat") {
            return CartStore.read().reduce((acc, entry) => acc + entry.cost.rawVAT, 0);

        }
        return CartStore.read().reduce((acc, entry) => acc + entry.cost.rawTotal, 0);
    }

    static clear(): void {
        localStorage.removeItem(CART_KEY);
    }

    static toJSON(): string {
        return JSON.stringify({
            items: CartStore.read().map((entry) => entry.id),
        });
    }

    static get len() {
        return CartStore.#len;
    };

    static *[Symbol.iterator]() {
        for (const item of CartStore.read()) {
            yield item;
        }
    }
}