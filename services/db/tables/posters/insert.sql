SELECT poster_insert(
   v.arg_id,
   v.arg_raw_id,
   v.arg_raw_amount,
   v.arg_raw_vat,
   v.arg_heading,
   v.arg_body,
   v.arg_width,
   v.arg_height,
   v.arg_origin_country,
   v.arg_origin_year,
   v.arg_condition_rating,
   v.arg_condition_notes
)
FROM (
    VALUES (
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
)
AS v(
    arg_id,
    arg_raw_id,
    arg_raw_amount,
    arg_raw_vat,
    arg_heading,
    arg_body,
    arg_width,
    arg_height,
    arg_origin_country,
    arg_origin_year,
    arg_condition_rating,
    arg_condition_notes
);