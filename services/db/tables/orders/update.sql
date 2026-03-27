UPDATE orders 
SET
    payment_intent_id = $2,
    status = CASE
        WHEN orders.status IN ('paid', 'fulfilled', 'shipped', 'completed') THEN orders.status
        ELSE $3
    END,
    updated_at = NOW()
WHERE
    id = $1
;