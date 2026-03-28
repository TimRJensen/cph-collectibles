BEGIN;

CREATE TABLE IF NOT EXISTS inventory (
    id varchar(26) PRIMARY KEY,
    raw_id varchar(16) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'rating') THEN
        CREATE TYPE rating AS ENUM ('poor', 'fair', 'good', 'very good', 'mint', 'unspecified');
    END IF;
END $$;
CREATE TABLE IF NOT EXISTS ratings (
    inventory_id varchar(26) PRIMARY KEY REFERENCES inventory(id) ON DELETE CASCADE,
    rating rating DEFAULT 'unspecified',
    notes text DEFAULT ''
);

CREATE TABLE IF NOT EXISTS costs (
    inventory_id varchar(26) PRIMARY KEY REFERENCES inventory(id) ON DELETE CASCADE,
    raw_amount numeric(10, 2) DEFAULT 0.0,
    raw_vat numeric(10, 2) DEFAULT 0.0,
    raw_total numeric(10, 2) GENERATED ALWAYS AS (raw_amount + raw_vat) STORED,
    minor_amount integer GENERATED ALWAYS AS ((ROUND(raw_amount * 100))::integer) STORED,
    minor_vat    integer GENERATED ALWAYS AS ((ROUND(raw_vat * 100))::integer) STORED,
    minor_total  integer GENERATED ALWAYS AS ((ROUND((raw_amount + raw_vat) * 100))::integer) STORED
);

CREATE TABLE IF NOT EXISTS details (
    inventory_id varchar(26) PRIMARY KEY REFERENCES inventory(id) ON DELETE CASCADE,
    heading text DEFAULT '',
    body text DEFAULT '',
    width numeric(6, 2) DEFAULT 0.0,
    height numeric(6, 2) DEFAULT 0.0,
    origin_source varchar(32) DEFAULT '',
    origin_year varchar(16) DEFAULT ''
);

CREATE TABLE IF NOT EXISTS files (
    id varchar(26) PRIMARY KEY,
    inventory_id varchar(26) REFERENCES inventory(id) ON DELETE CASCADE,
    url varchar(255) DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE OR REPLACE VIEW inventory_view AS
SELECT
    i.id,
    jsonb_build_object(
        'rawId', i.raw_id,
        'createdAt', i.created_at,
        'updatedAt', i.updated_at
    ) AS meta,
    jsonb_build_object(
        'rawAmount', c.raw_amount,
        'rawVAT', c.raw_vat,
        'rawTotal', c.raw_total,
        'minorAmount', c.minor_amount,
        'minorVAT', c.minor_vat,
        'minorTotal', c.minor_total
    ) AS cost,
    jsonb_build_object(
        'heading', d.heading,
        'width', d.width,
        'height', d.height,
        'origin', jsonb_build_object(
            'source', d.origin_source,
            'year', d.origin_year
        )
    ) AS detail,
    jsonb_build_object(
        'rating', r.rating,
        'notes', r.notes
    ) AS condition,
    COALESCE(f.files, '[]'::jsonb) AS files
FROM inventory i
LEFT JOIN details d ON d.inventory_id = i.id
LEFT JOIN costs c ON c.inventory_id = i.id
LEFT JOIN ratings r ON r.inventory_id = i.id
LEFT JOIN LATERAL (
    SELECT jsonb_agg(
        jsonb_build_object('id', f.id, 'url', f.url)
        ORDER BY f.created_at
    ) AS files
    FROM files f
    WHERE f.inventory_id = i.id
) f ON true;

-- orders
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'status') THEN
        CREATE TYPE status AS ENUM ('pending', 'paid', 'fulfilled', 'shipped', 'completed', 'cancelled');
    END IF;
END $$;
CREATE TABLE IF NOT EXISTS orders (
    id varchar(26) PRIMARY KEY,
    payment_intent_id text DEFAULT '',
    status status DEFAULT 'pending',
    total integer DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    order_id varchar(26) REFERENCES orders(id) ON DELETE CASCADE,
    inventory_id varchar(26) REFERENCES inventory(id) ON DELETE RESTRICT,
    PRIMARY KEY (order_id, inventory_id)
);

CREATE TABLE IF NOT EXISTS order_shipping (
    order_id varchar(26) PRIMARY KEY REFERENCES orders(id) ON DELETE CASCADE,
    first_name text DEFAULT '',
    last_name text DEFAULT '',
    address_line1 text DEFAULT '',
    address_line2 text DEFAULT '',
    postal_code int DEFAULT 0,
    city text DEFAULT '',
    state text DEFAULT '',
    country text DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_files_inventory_id_created_at
ON files (inventory_id, created_at);

CREATE INDEX IF NOT EXISTS idx_details_heading_fts
ON details
USING GIN (to_tsvector('simple', heading));

CREATE INDEX IF NOT EXISTS idx_orders_status
ON orders (status);

CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_payment_intent_id
ON orders (payment_intent_id)
WHERE payment_intent_id <> '';

CREATE INDEX IF NOT EXISTS idx_orders_created_at
ON orders (created_at);

COMMIT;
