-- 社区：关注、话题、小组、动态（飞行故事）、点赞、评论、挑战赛

CREATE TABLE user_follows (
    follower_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, followee_id),
    CHECK (follower_id <> followee_id)
);

CREATE INDEX idx_user_follows_followee ON user_follows(followee_id);

CREATE TABLE community_topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(128) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    cover_url TEXT,
    post_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_community_topics_post_count ON community_topics (post_count DESC);

CREATE TABLE community_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(128) NOT NULL UNIQUE,
    description TEXT,
    cover_url TEXT,
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    member_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE community_group_members (
    group_id UUID NOT NULL REFERENCES community_groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(16) NOT NULL DEFAULT 'member' CHECK (role IN ('member', 'admin')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX idx_community_group_members_user ON community_group_members(user_id);

CREATE TABLE community_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(500),
    body TEXT NOT NULL,
    visibility VARCHAR(32) NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'followers_only', 'group_only')),
    group_id UUID REFERENCES community_groups(id) ON DELETE SET NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'published' CHECK (status IN ('published', 'hidden')),
    media_urls JSONB NOT NULL DEFAULT '[]'::jsonb,
    like_count INT NOT NULL DEFAULT 0,
    comment_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (
        (visibility <> 'group_only' OR group_id IS NOT NULL)
    )
);

CREATE INDEX idx_community_posts_created ON community_posts (created_at DESC);
CREATE INDEX idx_community_posts_author_created ON community_posts (author_user_id, created_at DESC);
CREATE INDEX idx_community_posts_group_created ON community_posts (group_id, created_at DESC) WHERE group_id IS NOT NULL;

CREATE TABLE community_post_topics (
    post_id UUID NOT NULL REFERENCES community_posts(id) ON DELETE CASCADE,
    topic_id UUID NOT NULL REFERENCES community_topics(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, topic_id)
);

CREATE TABLE community_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    rules TEXT,
    cover_url TEXT,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'ended')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (ends_at >= starts_at)
);

CREATE INDEX idx_community_challenges_status ON community_challenges (status, starts_at DESC);

CREATE TABLE community_challenge_entries (
    challenge_id UUID NOT NULL REFERENCES community_challenges(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id UUID NOT NULL REFERENCES community_posts(id) ON DELETE CASCADE,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (challenge_id, user_id)
);

CREATE INDEX idx_community_challenge_entries_post ON community_challenge_entries(post_id);

CREATE TABLE community_post_likes (
    post_id UUID NOT NULL REFERENCES community_posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (post_id, user_id)
);

CREATE INDEX idx_community_post_likes_user ON community_post_likes(user_id);

CREATE TABLE community_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES community_posts(id) ON DELETE CASCADE,
    author_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    parent_id UUID REFERENCES community_comments(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_community_comments_post_created ON community_comments (post_id, created_at ASC);

-- 话题帖子数
CREATE OR REPLACE FUNCTION community_topics_bump_post_count() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE community_topics SET post_count = post_count + 1, updated_at = NOW() WHERE id = NEW.topic_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE community_topics SET post_count = GREATEST(0, post_count - 1), updated_at = NOW() WHERE id = OLD.topic_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_community_post_topics_ai
    AFTER INSERT ON community_post_topics
    FOR EACH ROW EXECUTE PROCEDURE community_topics_bump_post_count();

CREATE TRIGGER tr_community_post_topics_ad
    AFTER DELETE ON community_post_topics
    FOR EACH ROW EXECUTE PROCEDURE community_topics_bump_post_count();

-- 小组成员数
CREATE OR REPLACE FUNCTION community_groups_bump_member_count() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE community_groups SET member_count = member_count + 1, updated_at = NOW() WHERE id = NEW.group_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE community_groups SET member_count = GREATEST(0, member_count - 1), updated_at = NOW() WHERE id = OLD.group_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_community_group_members_ai
    AFTER INSERT ON community_group_members
    FOR EACH ROW EXECUTE PROCEDURE community_groups_bump_member_count();

CREATE TRIGGER tr_community_group_members_ad
    AFTER DELETE ON community_group_members
    FOR EACH ROW EXECUTE PROCEDURE community_groups_bump_member_count();

-- 点赞数
CREATE OR REPLACE FUNCTION community_posts_bump_like_count() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE community_posts SET like_count = like_count + 1, updated_at = NOW() WHERE id = NEW.post_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE community_posts SET like_count = GREATEST(0, like_count - 1), updated_at = NOW() WHERE id = OLD.post_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_community_post_likes_ai
    AFTER INSERT ON community_post_likes
    FOR EACH ROW EXECUTE PROCEDURE community_posts_bump_like_count();

CREATE TRIGGER tr_community_post_likes_ad
    AFTER DELETE ON community_post_likes
    FOR EACH ROW EXECUTE PROCEDURE community_posts_bump_like_count();

-- 评论数
CREATE OR REPLACE FUNCTION community_posts_bump_comment_count() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE community_posts SET comment_count = comment_count + 1, updated_at = NOW() WHERE id = NEW.post_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE community_posts SET comment_count = GREATEST(0, comment_count - 1), updated_at = NOW() WHERE id = OLD.post_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_community_comments_ai
    AFTER INSERT ON community_comments
    FOR EACH ROW EXECUTE PROCEDURE community_posts_bump_comment_count();

CREATE TRIGGER tr_community_comments_ad
    AFTER DELETE ON community_comments
    FOR EACH ROW EXECUTE PROCEDURE community_posts_bump_comment_count();

COMMENT ON TABLE user_follows IS '用户关注关系';
COMMENT ON TABLE community_topics IS '社区话题';
COMMENT ON TABLE community_groups IS '兴趣小组';
COMMENT ON TABLE community_group_members IS '小组成员';
COMMENT ON TABLE community_posts IS '社区动态/飞行故事';
COMMENT ON TABLE community_post_topics IS '动态关联话题';
COMMENT ON TABLE community_challenges IS '飞行挑战赛';
COMMENT ON TABLE community_challenge_entries IS '挑战赛作品（关联动态）';
COMMENT ON TABLE community_post_likes IS '动态点赞';
COMMENT ON TABLE community_comments IS '动态评论';
