-- Migration 009: 理论试卷与试卷题目

-- 理论试卷表：一个理论文档下可以有多套试卷
CREATE TABLE IF NOT EXISTS theory_papers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    theory_document_id UUID NOT NULL REFERENCES theory_documents(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    total_score INT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE theory_papers IS '理论试卷：隶属于某个理论文档，可配置多套卷';
COMMENT ON COLUMN theory_papers.theory_document_id IS '关联的理论文档ID';
COMMENT ON COLUMN theory_papers.total_score IS '试卷总分，来自所有题目分值之和';
COMMENT ON COLUMN theory_papers.enabled IS '是否启用，对学员可见';

CREATE INDEX IF NOT EXISTS idx_theory_papers_doc ON theory_papers(theory_document_id);

-- 试卷题目关联表：记录某套试卷下包含哪些题目及其分值和顺序
CREATE TABLE IF NOT EXISTS theory_paper_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    paper_id UUID NOT NULL REFERENCES theory_papers(id) ON DELETE CASCADE,
    question_id VARCHAR(255) NOT NULL,
    score INT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0
);

COMMENT ON TABLE theory_paper_questions IS '理论试卷题目关联：包含题目ID、分值与排序';
COMMENT ON COLUMN theory_paper_questions.paper_id IS '所属试卷ID';
COMMENT ON COLUMN theory_paper_questions.question_id IS '题目ID（Mongo 题库中的 questions.id）';

CREATE INDEX IF NOT EXISTS idx_theory_paper_questions_paper ON theory_paper_questions(paper_id);

