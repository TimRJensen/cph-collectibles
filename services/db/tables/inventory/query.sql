SELECT id, meta, cost, detail, condition, files FROM inventory_view 
WHERE to_tsvector('simple', detail ->> 'heading') @@ plainto_tsquery($1);