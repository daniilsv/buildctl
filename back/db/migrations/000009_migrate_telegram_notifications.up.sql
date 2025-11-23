-- Migrate telegram_chat_ids to telegram_notifications format
-- Convert array of numbers to array of objects: [123, 456] -> [{"chat_id": 123}, {"chat_id": 456}]

-- Migrate projects
UPDATE projects
SET settings = (
    SELECT jsonb_set(
        settings - 'telegram_chat_ids',
        '{telegram_notifications}',
        COALESCE(
            (
                SELECT jsonb_agg(jsonb_build_object('chat_id', elem::numeric::bigint))
                FROM jsonb_array_elements(settings->'telegram_chat_ids') AS elem
            ),
            '[]'::jsonb
        )
    )
)
WHERE settings ? 'telegram_chat_ids' AND jsonb_typeof(settings->'telegram_chat_ids') = 'array';

-- Migrate branches
UPDATE branches
SET settings = (
    SELECT jsonb_set(
        settings - 'telegram_chat_ids',
        '{telegram_notifications}',
        COALESCE(
            (
                SELECT jsonb_agg(jsonb_build_object('chat_id', elem::numeric::bigint))
                FROM jsonb_array_elements(settings->'telegram_chat_ids') AS elem
            ),
            '[]'::jsonb
        )
    )
)
WHERE settings ? 'telegram_chat_ids' AND jsonb_typeof(settings->'telegram_chat_ids') = 'array';

