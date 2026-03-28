export type Endpoint =
    | "/api/v1/inventory"
    | `/api/v1/inventory?${string}`
    | `/api/v1/inventory/${string}`
    | "/api/v1/checkout"
    | `/api/v1/checkout/${string}`;

type AnyResult<T, B extends boolean = boolean> =
    B extends true
    ? {
        code: number;
        error: B;
        msg: string;
    } : {
        code: number;
        error: B;
        data: Array<T>;
    }

type CheckoutResult = {
    status: string,
    redirectURL: string,
    publishableKey: string,
    clientSecret: string,
    orderId: string,
}

type OriginObject = {
    source: string,
    year: string,
}

export type PosterResult = {
    id: string,
    meta: {
        rawId: string,
        createdAt: string,
        updatedAt: string,
    },
    cost: {
        rawAmount: number,
        rawVAT: number,
        rawTotal: number,
        minorAmount: number,
        minorVAT: number,
        minorToral: number,
    },
    detail: {
        heading: string,
        body: string,
        width: number,
        height: number,
        origin: OriginObject,
    },
    condition: {
        rating: string,
        notes: string,
    },
    files: Array<{ id: string, url: string }>,
}

export type APIResult<E extends Endpoint>
    = E extends "/api/v1/inventory" ? AnyResult<PosterResult>
    : E extends `/api/v1/inventory?${string}` ? AnyResult<PosterResult>
    : E extends `/api/v1/inventory/${string}` ? AnyResult<PosterResult>
    : E extends "/api/v1/checkout" ? AnyResult<CheckoutResult>
    : E extends `/api/v1/checkout/${string}` ? AnyResult<CheckoutResult>
    : never;

export async function request<E extends Endpoint>(endpoint: E, method: string, body: any = null): Promise<APIResult<E>> {
    try {
        const response = await fetch(endpoint, {
            method,
            body,
        });
        return response.json();
    } catch (err) {
        return <APIResult<E>>err;
    }
}