<script lang="ts" setup>
// vendor imports
import { computed, nextTick, onMounted, reactive } from "vue";
import type { Appearance, Stripe, StripeElements, StripePaymentElementOptions, BillingDetails, ConfirmationToken } from "@stripe/stripe-js"
import { loadStripe } from "@stripe/stripe-js";
// custom imports
import { request } from "../../util/request";
import OrderList from "../order-list/OrderList.vue";
import StepBar from "./StepBar.vue";
import Details from "./Details.vue";
import { CartStore } from "../../store";
// props
// state
const state = reactive<{
    idx: number,
    shipping: BillingDetails | null,
    payment: ConfirmationToken.PaymentMethodPreview | null,
    email: string,
}>({
    idx: 0,
    shipping: null,
    payment: null,
    email: "",
});

const translate = computed(() => {
    return `-${(state.idx - 1) * 100}%`
});
let stripe: Stripe | null;
let elements: StripeElements;
let order: string;
let secret: string;
let token: string;
onMounted(async () => {
    let res = await request("/api/v1/checkout", "GET");
    if (res.error) {
        console.log(res.error);
        return;
    }

    stripe = await loadStripe(res.data[0]?.publishableKey!);
    if (!stripe) {
        return;
    }

    res = await request("/api/v1/checkout", "POST", CartStore.toJSON());
    if (res.error) {
        console.log(res.error);
        return;
    } else {
        const data = res.data[0]!;
        order = data.orderId;
        secret = data.clientSecret;
    }

    const appearance = <Appearance>{
        inputs: "spaced",
        labels: "floating",
        variables: {
            fontFamily: 'Roboto, system-ui, sans-serif',
            fontSizeBase: '18px',
            fontLineHeight: '24px',
            fontWeightNormal: '400',
            focusBoxShadow: "none",
        },
        rules: {
            ".Input": {
                border: "1px solid #B5B8BC",
            },
        },
    }
    elements = stripe.elements({ appearance, clientSecret: secret });

    const options = <StripePaymentElementOptions>{
        layout: {
            type: 'accordion',
            defaultCollapsed: false,
            radios: 'always',
            spacedAccordionItems: true
        },
        business: {
            name: "Copenhagen Collectables",
        }
    };
    const address = elements.create("address", { mode: "shipping" });
    address.mount("#address-element");
    address.on("ready", async () => {
        await nextTick();
        state.idx = 1
    });

    const payment = elements.create("payment", options);
    payment.mount("#payment-element");
});

async function validate(next: number): Promise<void> {
    if (!stripe) {
        return;
    }

    switch (state.idx) {
        case 1:
            const address = elements.getElement("address");
            if (!address) {
                return;
            }

            const { complete, value } = await address.getValue();
            if (!complete) {
                return;
            }
            state.shipping = value;
            state.idx = next;
            break;
        case 2: {
            if ((await elements.submit()).error) {
                return;
            }

            const { error, confirmationToken } = await stripe.createConfirmationToken({
                elements,
                params: {
                    return_url: `${window.location.origin}/checkout/complete`,
                },
            });

            if (error || !confirmationToken) {
                console.log(error)
                return;
            }
            token = confirmationToken.id;
            state.payment = confirmationToken.payment_method_preview;
            state.idx = next;
            break;
        }
        case 3: {
            if ((await elements.submit()).error) {
                return;
            }

            const res = await request(`/api/v1/checkout/${order}`, "POST", JSON.stringify({
                confirmationTokenId: token,
                email: state.email,
                shipping: state.shipping,
            }));
            if (res.error) {
                console.log(res.error);
                return;
            }

            const data = res.data[0]!;
            switch (data.status) {
                case "succeeded":
                    window.location.href = "/checkout/complete/";
                    return;
                case "processing":
                    window.location.href = "/checkout/complete/?status=processing";
                    return;
                case "requires_action":
                    if (data.redirectURL) {
                        window.location.href = data.redirectURL;
                        return;
                    }
                    if (data.clientSecret && stripe) {
                    }
                    return;
                case "requires_payment_method":
                    state.idx = 2;
                    return;
                default:
                    return;
            }
        }
    }
}
</script>
<template>
    <section :class="$style.view">
        <OrderList :class="$style.orderlist" :show-total="false" />
        <StepBar :class="$style.flowbar" v-model="state.idx" />
        <div :class="$style.slideview">
            <div :class="$style.slidecontainer">
                <div :class="$style.slideitem">
                    <label :class="$style.field">
                        <input v-model="state.email" :class="$style.input" type="email" placeholder=" " />
                        <span :class="$style.label">Email</span>
                    </label>
                    <div :class="$style.slideitem" id="address-element"></div>
                </div>
                <div :class="$style.slideitem" id="payment-element"></div>
                <Details :billing="state.shipping" :payment="state.payment" />
            </div>
        </div>
        <button v-if="state.idx" :class="$style.button" :disabled="state.idx < 1" @click="validate(state.idx + 1)">
            {{ state.idx < 3 ? "Next" : "Place order" }} </button>
    </section>
</template>

<style lang="css" module>
.view {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    grid-template-rows: min-content 1fr min-content;
    gap: 1.5rem;
    width: 100%;
    min-height: 768px;
    margin: auto;
    margin-top: 1rem;
    padding: 1rem;
    box-shadow: -1px 1px 4px 0 var(--card-border-color);
    background: var(--primary);
    color: var(--txt-primary-color);
}

.orderlist {
    display: none;
    grid-column: 1;
    grid-row: 2 / span 2;
}

.flowbar {
    grid-column: 1 / span 2;
    grid-row: 1;
    width: 100%;
}

.slideview {
    grid-column: 1 / span 2;
    grid-row: 2;
    width: 100%;
    min-width: 100%;
    overflow: hidden;
}

.slidecontainer {
    display: flex;
    flex-flow: row nowrap;
    width: 100%;
    translate: v-bind(translate) 0;
    transition: translate 250ms;
}

.slideitem {
    width: 100%;
    min-width: 100%;
}

.button {
    grid-column: 2;
    place-self: end;
    width: 128px;
    padding: 0.5rem;
    border-radius: var(--btn-border-radius);
    border: var(--btn-border-size) solid var(--btn-border-color);
    background: var(--btn-background);
    color: var(--btn-color);
}

.field {
    position: relative;
    display: block;
    margin-bottom: 0.75rem;
    border-radius: var(--input-border-radius);
    border: var(--input-border-size) solid var(--input-border-color);
    background: var(--input-background);
    font-size: 16px;
}

.input {
    width: 100%;
    height: 64px;
    min-height: 64px;
    padding: 0.85rem;
    padding-top: 35px;
    border: 0;
    outline: 0;
    background: transparent;
    font: inherit;
}

.label {
    position: absolute;
    left: 0;
    top: 0;
    margin-left: 13px;
    color: rgb(48, 49, 61);
    font-size: 18x;
    line-height: 24px;
    font-weight: 400;
    transform-origin: top left;
    transform: translateY(23px);
    transition: all 120ms ease;
    pointer-events: none;
}

.field:focus-within .label,
.input:not(:placeholder-shown)+.label {
    padding-bottom: 0.25rem;
    transform: translateY(13px) scale(0.8888);
    color: var(--input-color);
}

@media screen and (min-width: 1024px) {
    .view {
        border-radius: var(--card-border-radius);
    }

    .orderlist {
        display: initial;
        grid-column: 1;
        grid-row: 1 / span 2;
    }

    .flowbar {
        grid-column: 2;
        grid-row: 1;
    }

    .slideview {
        grid-column: 2;
        grid-row: 2;
        width: 100%;
        min-width: 100%;
        overflow: hidden;
    }

}
</style>
