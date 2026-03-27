BEGIN;

CREATE TABLE IF NOT EXISTS posters (
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
    poster_id varchar(26) PRIMARY KEY REFERENCES posters(id) ON DELETE CASCADE,
    rating rating DEFAULT 'unspecified',
    notes text DEFAULT ''
);

CREATE TABLE IF NOT EXISTS costs (
    poster_id varchar(26) PRIMARY KEY REFERENCES posters(id) ON DELETE CASCADE,
    raw_amount numeric(10, 2) DEFAULT 0.0,
    raw_vat numeric(10, 2) DEFAULT 0.0,
    raw_total numeric(10, 2) GENERATED ALWAYS AS (raw_amount + raw_vat) STORED,
    minor_amount integer GENERATED ALWAYS AS ((ROUND(raw_amount * 100))::integer) STORED,
    minor_vat    integer GENERATED ALWAYS AS ((ROUND(raw_vat * 100))::integer) STORED,
    minor_total  integer GENERATED ALWAYS AS ((ROUND((raw_amount + raw_vat) * 100))::integer) STORED
);

CREATE TABLE IF NOT EXISTS details (
    poster_id varchar(26) PRIMARY KEY REFERENCES posters(id) ON DELETE CASCADE,
    heading text DEFAULT '',
    body text DEFAULT '',
    width numeric(6, 2) DEFAULT 0.0,
    height numeric(6, 2) DEFAULT 0.0,
    origin_source varchar(32) DEFAULT '',
    origin_year varchar(16) DEFAULT ''
);

CREATE TABLE IF NOT EXISTS files (
    id varchar(26) PRIMARY KEY,
    poster_id varchar(26) REFERENCES posters(id) ON DELETE CASCADE,
    url varchar(255) DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION poster_insert(
    arg_id varchar(26),
    arg_raw_id varchar(16),
    arg_raw_amount numeric(10, 2),
    arg_raw_vat numeric(10, 2),
    arg_heading text,
    arg_body text,
    arg_width numeric(6, 2),
    arg_height numeric(6, 2),
    arg_origin_source varchar(32),
    arg_origin_year varchar(16),
    arg_condition_rating rating,
    arg_condition_notes text
) RETURNS void AS $$
BEGIN
    INSERT INTO posters (id, raw_id)
    VALUES (arg_id, arg_raw_id);

    INSERT INTO costs (poster_id, raw_amount, raw_vat)
    VALUES (arg_id, arg_raw_amount, arg_raw_vat);

    INSERT INTO details (
        poster_id, heading, body, width, height, origin_source, origin_year
    )
    VALUES (
        arg_id, arg_heading, arg_body, arg_width, arg_height, arg_origin_source, arg_origin_year
    );

    INSERT INTO ratings (poster_id, rating, notes)
    VALUES (arg_id, arg_condition_rating, arg_condition_notes);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE VIEW poster_view AS
SELECT
    p.id,
    jsonb_build_object(
        'rawId', p.raw_id,
        'createdAt', p.created_at,
        'updatedAt', p.updated_at
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
FROM posters p
LEFT JOIN details d ON d.poster_id = p.id
LEFT JOIN costs c ON c.poster_id = p.id
LEFT JOIN ratings r ON r.poster_id = p.id
LEFT JOIN LATERAL (
    SELECT jsonb_agg(
        jsonb_build_object('id', f.id, 'url', f.url)
        ORDER BY f.created_at
    ) AS files
    FROM files f
    WHERE f.poster_id = p.id
) f ON true;

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
    poster_id varchar(26) REFERENCES posters(id) ON DELETE RESTRICT,
    PRIMARY KEY (order_id, poster_id)
);

COMMIT;
