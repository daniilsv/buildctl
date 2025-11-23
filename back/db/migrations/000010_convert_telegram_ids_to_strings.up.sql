-- Convert numeric chat_id and thread_id to strings in telegram_notifications
-- This ensures compatibility with the new string-based handling

-- Migrate projects
UPDATE projects
SET settings = jsonb_set(
    settings,
    '{telegram_notifications}',
    COALESCE(
        (
            SELECT jsonb_agg(
                CASE 
                    WHEN elem->'thread_id' IS NOT NULL THEN
                        jsonb_build_object(
                            'chat_id', (elem->'chat_id')::text,
                            'thread_id', (elem->'thread_id')::text
                        )
                    ELSE
                        jsonb_build_object(
                            'chat_id', (elem->'chat_id')::text
                        )
                END
            )
            FROM jsonb_array_elements(settings->'telegram_notifications') AS elem
            WHERE elem->'chat_id' IS NOT NULL
        ),
        '[]'::jsonb
    )
)
WHERE settings ? 'telegram_notifications' 
  AND jsonb_typeof(settings->'telegram_notifications') = 'array'
  AND jsonb_array_length(settings->'telegram_notifications') > 0;

-- Migrate branches
UPDATE branches
SET settings = jsonb_set(
    settings,
    '{telegram_notifications}',
    COALESCE(
        (
            SELECT jsonb_agg(
                CASE 
                    WHEN elem->'thread_id' IS NOT NULL THEN
                        jsonb_build_object(
                            'chat_id', (elem->'chat_id')::text,
                            'thread_id', (elem->'thread_id')::text
                        )
                    ELSE
                        jsonb_build_object(
                            'chat_id', (elem->'chat_id')::text
                        )
                END
            )
            FROM jsonb_array_elements(settings->'telegram_notifications') AS elem
            WHERE elem->'chat_id' IS NOT NULL
        ),
        '[]'::jsonb
    )
)
WHERE settings ? 'telegram_notifications' 
  AND jsonb_typeof(settings->'telegram_notifications') = 'array'
  AND jsonb_array_length(settings->'telegram_notifications') > 0;
