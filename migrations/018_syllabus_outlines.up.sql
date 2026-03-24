-- 训练大纲（多实例、五类 category）与一级科目归属、二级科目表

CREATE TABLE syllabus_outlines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category VARCHAR(32) NOT NULL CHECK (category IN (
        'spl_local',
        'spl_transition',
        'spl_night',
        'ppl_basic',
        'ppl_night'
    )),
    name VARCHAR(255) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE syllabus_outlines IS '训练大纲实例：类别 + 名称，下挂一级科目';
COMMENT ON COLUMN syllabus_outlines.category IS 'spl_local|SPL本场 spl_transition|SPL转场 spl_night|SPL夜航 ppl_basic|PPL基本 ppl_night|PPL夜航';

-- 迁移用默认大纲（承接历史扁平 syllabus_subjects）
INSERT INTO syllabus_outlines (id, category, name, sort_order)
VALUES (
    '00000001-0000-4000-8000-000000000001',
    'ppl_basic',
    '历史数据（迁移默认大纲）',
    0
);

ALTER TABLE syllabus_subjects
    ADD COLUMN syllabus_outline_id UUID REFERENCES syllabus_outlines(id) ON DELETE CASCADE;

UPDATE syllabus_subjects
SET syllabus_outline_id = '00000001-0000-4000-8000-000000000001'
WHERE syllabus_outline_id IS NULL;

ALTER TABLE syllabus_subjects
    ALTER COLUMN syllabus_outline_id SET NOT NULL;

ALTER TABLE syllabus_subjects
    ADD CONSTRAINT syllabus_subjects_outline_code_unique UNIQUE (syllabus_outline_id, code);

CREATE TABLE syllabus_subject_secondaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    syllabus_subject_id UUID NOT NULL REFERENCES syllabus_subjects(id) ON DELETE CASCADE,
    seq_no INT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    completion_standard TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_syllabus_subject_secondaries_subject_sort
    ON syllabus_subject_secondaries (syllabus_subject_id, sort_order, id);

COMMENT ON TABLE syllabus_subject_secondaries IS '一级科目下的二级科目：序号、名称、描述、完成标准、排序';
