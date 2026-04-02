-- 学员总表 CSV 导入扩展：联系方式、身份与地址、服务状态原文及排班/权益快照（JSON）
ALTER TABLE students
    ADD COLUMN IF NOT EXISTS contact_phone VARCHAR(128),
    ADD COLUMN IF NOT EXISTS gender VARCHAR(16),
    ADD COLUMN IF NOT EXISTS birth_date DATE,
    ADD COLUMN IF NOT EXISTS id_document_no VARCHAR(256),
    ADD COLUMN IF NOT EXISTS birth_place VARCHAR(512),
    ADD COLUMN IF NOT EXISTS license_archive_no VARCHAR(128),
    ADD COLUMN IF NOT EXISTS education_level VARCHAR(128),
    ADD COLUMN IF NOT EXISTS permanent_address TEXT,
    ADD COLUMN IF NOT EXISTS mailing_address TEXT,
    ADD COLUMN IF NOT EXISTS region VARCHAR(16),
    ADD COLUMN IF NOT EXISTS service_status_raw VARCHAR(64),
    ADD COLUMN IF NOT EXISTS pause_reason TEXT,
    ADD COLUMN IF NOT EXISTS transfer_base VARCHAR(16),
    ADD COLUMN IF NOT EXISTS registry_extra JSONB;

COMMENT ON COLUMN students.contact_phone IS '联系电话（档案，未绑系统账号时亦可填）';
COMMENT ON COLUMN students.gender IS '性别';
COMMENT ON COLUMN students.birth_date IS '出生日期';
COMMENT ON COLUMN students.id_document_no IS '证件号码';
COMMENT ON COLUMN students.birth_place IS '出生地';
COMMENT ON COLUMN students.license_archive_no IS '执照档案号';
COMMENT ON COLUMN students.education_level IS '学历';
COMMENT ON COLUMN students.permanent_address IS '永久地址';
COMMENT ON COLUMN students.mailing_address IS '通讯地址';
COMMENT ON COLUMN students.region IS '港澳外等区域：港/澳/台/内/外';
COMMENT ON COLUMN students.service_status_raw IS '服务状态原文（总表导入）';
COMMENT ON COLUMN students.pause_reason IS '中断原因';
COMMENT ON COLUMN students.transfer_base IS '转场（有/无等）';
COMMENT ON COLUMN students.registry_extra IS '总表其余字段快照：排班限制、体验/视频/定制、合同小时原文等';
