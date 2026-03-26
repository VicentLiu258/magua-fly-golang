-- 理论文档托管视频元数据（GridFS / 后续 COS），与外链 video_url 互斥由 CHECK 约束

ALTER TABLE theory_documents
    ADD COLUMN video_storage VARCHAR(16),
    ADD COLUMN video_object_id VARCHAR(256),
    ADD COLUMN video_content_type VARCHAR(128),
    ADD COLUMN video_size_bytes BIGINT,
    ADD COLUMN video_duration_seconds INT;

ALTER TABLE theory_documents
    ADD CONSTRAINT theory_documents_video_storage_ck CHECK (
        video_storage IS NULL
        OR (
            video_storage = 'external'
            AND NULLIF(btrim(COALESCE(video_url, '')), '') IS NOT NULL
            AND NULLIF(btrim(COALESCE(video_object_id, '')), '') IS NULL
        )
        OR (
            video_storage IN ('gridfs', 'cos')
            AND NULLIF(btrim(COALESCE(video_object_id, '')), '') IS NOT NULL
        )
    );

COMMENT ON COLUMN theory_documents.video_storage IS '视频存储：external 外链 / gridfs / cos';
COMMENT ON COLUMN theory_documents.video_object_id IS '托管对象 ID（GridFS 文件 _id 或 COS object key）';
COMMENT ON COLUMN theory_documents.video_content_type IS '视频 MIME，如 video/mp4';
COMMENT ON COLUMN theory_documents.video_size_bytes IS '视频字节大小';
COMMENT ON COLUMN theory_documents.video_duration_seconds IS '时长（秒），可后续异步填充';
