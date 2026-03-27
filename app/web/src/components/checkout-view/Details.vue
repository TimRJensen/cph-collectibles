<script lang="ts" setup>
// vendor imports
import { computed, ref } from "vue";
import type { BillingDetails, ConfirmationToken } from "@stripe/stripe-js"
// custom imports
import { CartStore } from '../../store';

// props
const props = withDefaults(
    defineProps<{ billing: BillingDetails | null, payment: ConfirmationToken.PaymentMethodPreview | null }>(),
    {}
);
// state
const items = ref(CartStore.read())
const method = computed(() => {
    if (!props.payment) {
        return;
    }

    const p = props.payment;
    switch (props.payment?.type) {
        case "card":
            return `${p.card?.brand.toUpperCase()}`
    }
});
const details = computed(() => {
    if (!props.payment) {
        return;
    }

    const p = props.payment;
    switch (p.type) {
        case "card":
            return `•••• •••• •••• ${p.card?.last4}`
    }
});
const detailsExtra = computed(() => {
    if (!props.payment) {
        return;
    }

    const p = props.payment;
    switch (p.type) {
        case "card":
            return `${p.card?.exp_month}/${p.card?.exp_year}`
    }
});
</script>

<template>
    <div v-if="billing" :class="$style.container">
        <div :class="$style.detail">
            <div :class="$style.label">Shipping details</div>
            <hr :class="$style.line">
            <div>{{ billing.name }}</div>
            <div>{{ billing.address?.line1 }}</div>
            <div>{{ billing.address?.postal_code }} {{ billing.address.city }}</div>
            <div>{{ billing.address?.state }} {{ billing.address.country }}</div>
        </div>
        <div :class="$style.detail">
            <div :class="$style.label">Payment details</div>
            <hr :class="$style.line">
            <div>{{ method }}</div>
            <div><span>{{ details }}</span>&nbsp;&nbsp;<span>{{ detailsExtra }}</span></div>
        </div>
        <div :class="$style.detail">
            <div :class="$style.label">Order details</div>
            <hr :class="$style.line">
            <div v-for="data of items" :class="$style.item">
                <span :class="$style.label">{{ data.detail.heading }}</span>
                <span :class="[$style.label, $style.right]">{{ data.cost.rawTotal }}&#163;</span>
                <span :class="[$style.label, $style.right]">VAT {{ data.cost.rawVAT }}&#163;</span>
            </div>
            <div :class="$style.item">
                <span :class="$style.label">Total</span>
                <span :class="[$style.label, $style.right]">{{ CartStore.total() }}&#163;</span>
                <span :class="[$style.label, $style.right]">VAT {{ CartStore.total("vat") }}&#163;</span>
            </div>
        </div>
    </div>
</template>

<style lang="css" module>
.container {
    display: flex;
    flex-wrap: wrap;
    width: 100%;
    min-width: 100%;
    height: fit-content;
    gap: 1rem;
    font-size: 0.8rem;
    font-weight: 200;
}

.detail {
    display: grid;
    width: 100%;
    min-width: 100%;
    height: fit-content;
    row-gap: 0.25rem;
    padding: 0 12px;
}

.detail .line {
    width: 100%;
}

.label {
    font-weight: 400;
}

.label.right {
    text-align: right;
}

.item {
    display: flex;
    flex-wrap: wrap;
    width: 100%;
}

.item .label {
    width: 25%;
    font-weight: bold;
}

.item .label:first-child {
    width: 75%;
    font-weight: bold;
}

.item .label:last-child {
    margin-left: auto;
    font-weight: unset;
    font-size: 0.6rem;
}
</style>