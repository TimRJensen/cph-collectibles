WITH
  data (
    id,
    raw_id,
    raw_amount,
    raw_vat,
    heading,
    body,
    width,
    height,
    origin_source,
    origin_year,
    rating_rating,
    rating_notes
  ) AS (
    VALUES
      (
        $1,
        $2,
        $3::numeric(10, 2),
        $4::numeric(10, 2),
        $5,
        $6,
        $7::numeric(6, 2),
        $8::numeric(6, 2),
        $9,
        $10,
        $11::rating,
        $12
      )
  ),
  insert_inventory AS (
    INSERT INTO
      inventory (id, raw_id)
    SELECT
      d.id,
      d.raw_id
    FROM
      data AS d
    ON CONFLICT (id) DO UPDATE
    SET
      raw_id = EXCLUDED.raw_id,
      updated_at = now()
      -- RETURNING *
  ),
  insert_cost AS (
    INSERT INTO
      costs (inventory_id, raw_amount, raw_vat)
    SELECT
      d.id,
      d.raw_amount,
      d.raw_vat
    FROM
      data AS d
    ON CONFLICT (inventory_id) DO UPDATE
    SET
      raw_amount = EXCLUDED.raw_amount,
      raw_vat = EXCLUDED.raw_vat
      -- RETURNING *
  ),
  insert_details AS (
    INSERT INTO
      details (
        inventory_id,
        heading,
        body,
        width,
        height,
        origin_source,
        origin_year
      )
    SELECT
      d.id,
      d.heading,
      d.body,
      d.width,
      d.height,
      d.origin_source,
      d.origin_year
    FROM
      data AS d
    ON CONFLICT (inventory_id) DO UPDATE
    SET
      heading = EXCLUDED.heading,
      body = EXCLUDED.body,
      width = EXCLUDED.width,
      height = EXCLUDED.height,
      origin_source = EXCLUDED.origin_source,
      origin_year = EXCLUDED.origin_year
      -- RETURNING *
  ),
  insert_ratings AS (
    INSERT INTO
      ratings (inventory_id, rating, notes)
    SELECT
      d.id,
      d.rating_rating,
      d.rating_notes
    FROM
      data AS d
    ON CONFLICT (inventory_id) DO UPDATE
    SET
      rating = EXCLUDED.rating,
      notes = EXCLUDED.notes
      -- RETURNING *
  )
SELECT
  1;