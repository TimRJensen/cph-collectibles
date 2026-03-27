type Endpoints =
    | "/api/v1/posters"
    | `/api/v1/posters/${string}`
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

export type APIResult<E extends Endpoints>
    = E extends "/api/v1/posters" ? AnyResult<PosterResult>
    : E extends `/api/v1/posters/${string}` ? AnyResult<PosterResult>
    : E extends "/api/v1/checkout" ? AnyResult<CheckoutResult>
    : E extends `/api/v1/checkout/${string}` ? AnyResult<CheckoutResult>
    : never;


const url = import.meta.env.VITE_API_URL;

export async function request<E extends Endpoints>(endpoint: E, method: string, body: any = null): Promise<APIResult<E>> {
    try {
        const response = await fetch(url + endpoint, {
            method,
            body,
        });
        return response.json();
    } catch (err) {
        return <APIResult<E>>err;
    }
}