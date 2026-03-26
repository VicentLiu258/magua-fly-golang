ALTER TABLE theory_documents DROP CONSTRAINT IF EXISTS theory_documents_video_storage_ck;

ALTER TABLE theory_documents
    DROP COLUMN IF EXISTS video_duration_seconds,
    DROP COLUMN IF EXISTS video_size_bytes,
    DROP COLUMN IF EXISTS video_content_type,
    DROP COLUMN IF EXISTS video_object_id,
    DROP COLUMN IF EXISTS video_storage;
