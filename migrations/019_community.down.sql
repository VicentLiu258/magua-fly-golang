DROP TRIGGER IF EXISTS tr_community_comments_ad ON community_comments;
DROP TRIGGER IF EXISTS tr_community_comments_ai ON community_comments;
DROP TRIGGER IF EXISTS tr_community_post_likes_ad ON community_post_likes;
DROP TRIGGER IF EXISTS tr_community_post_likes_ai ON community_post_likes;
DROP TRIGGER IF EXISTS tr_community_group_members_ad ON community_group_members;
DROP TRIGGER IF EXISTS tr_community_group_members_ai ON community_group_members;
DROP TRIGGER IF EXISTS tr_community_post_topics_ad ON community_post_topics;
DROP TRIGGER IF EXISTS tr_community_post_topics_ai ON community_post_topics;

DROP FUNCTION IF EXISTS community_posts_bump_comment_count();
DROP FUNCTION IF EXISTS community_posts_bump_like_count();
DROP FUNCTION IF EXISTS community_groups_bump_member_count();
DROP FUNCTION IF EXISTS community_topics_bump_post_count();

DROP TABLE IF EXISTS community_comments;
DROP TABLE IF EXISTS community_post_likes;
DROP TABLE IF EXISTS community_challenge_entries;
DROP TABLE IF EXISTS community_challenges;
DROP TABLE IF EXISTS community_post_topics;
DROP TABLE IF EXISTS community_posts;
DROP TABLE IF EXISTS community_group_members;
DROP TABLE IF EXISTS community_groups;
DROP TABLE IF EXISTS community_topics;
DROP TABLE IF EXISTS user_follows;
