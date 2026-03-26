package domain

import "time"

// CommunityPostVisibility 动态可见范围
type CommunityPostVisibility string

const (
	CommunityVisibilityPublic         CommunityPostVisibility = "public"
	CommunityVisibilityFollowersOnly  CommunityPostVisibility = "followers_only"
	CommunityVisibilityGroupOnly      CommunityPostVisibility = "group_only"
)

// CommunityPostStatus 动态状态
type CommunityPostStatus string

const (
	CommunityPostPublished CommunityPostStatus = "published"
	CommunityPostHidden    CommunityPostStatus = "hidden"
)

// CommunityGroupMemberRole 小组成员角色
type CommunityGroupMemberRole string

const (
	CommunityGroupRoleMember CommunityGroupMemberRole = "member"
	CommunityGroupRoleAdmin  CommunityGroupMemberRole = "admin"
)

// CommunityChallengeStatus 挑战赛状态
type CommunityChallengeStatus string

const (
	CommunityChallengeDraft  CommunityChallengeStatus = "draft"
	CommunityChallengeActive CommunityChallengeStatus = "active"
	CommunityChallengeEnded  CommunityChallengeStatus = "ended"
)

// CommunityPost 社区动态 / 飞行故事
type CommunityPost struct {
	ID             string                  `json:"id"`
	AuthorUserID   string                  `json:"author_user_id"`
	Title          string                  `json:"title,omitempty"`
	Body           string                  `json:"body"`
	Visibility     CommunityPostVisibility `json:"visibility"`
	GroupID        *string                 `json:"group_id,omitempty"`
	Status         CommunityPostStatus     `json:"status"`
	MediaURLs      []string                `json:"media_urls"`
	LikeCount      int                     `json:"like_count"`
	CommentCount   int                     `json:"comment_count"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	TopicIDs       []string                `json:"topic_ids,omitempty"` // 列表/详情可填充
	LikedByMe      bool                    `json:"liked_by_me,omitempty"`
}

// CommunityTopic 话题
type CommunityTopic struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	CoverURL    string    `json:"cover_url,omitempty"`
	PostCount   int       `json:"post_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CommunityGroup 兴趣小组
type CommunityGroup struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description,omitempty"`
	CoverURL     string    `json:"cover_url,omitempty"`
	OwnerUserID  string    `json:"owner_user_id"`
	IsPublic     bool      `json:"is_public"`
	MemberCount  int       `json:"member_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CommunityGroupMember 小组成员
type CommunityGroupMember struct {
	GroupID  string                   `json:"group_id"`
	UserID   string                   `json:"user_id"`
	Role     CommunityGroupMemberRole `json:"role"`
	JoinedAt time.Time                `json:"joined_at"`
}

// CommunityChallenge 飞行挑战赛
type CommunityChallenge struct {
	ID          string                   `json:"id"`
	Title       string                   `json:"title"`
	Description string                   `json:"description,omitempty"`
	Rules       string                   `json:"rules,omitempty"`
	CoverURL    string                   `json:"cover_url,omitempty"`
	StartsAt    time.Time                `json:"starts_at"`
	EndsAt      time.Time                `json:"ends_at"`
	Status      CommunityChallengeStatus `json:"status"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

// CommunityChallengeEntry 参赛作品
type CommunityChallengeEntry struct {
	ChallengeID string    `json:"challenge_id"`
	UserID      string    `json:"user_id"`
	PostID      string    `json:"post_id"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// CommunityComment 评论
type CommunityComment struct {
	ID           string    `json:"id"`
	PostID       string    `json:"post_id"`
	AuthorUserID string    `json:"author_user_id"`
	Body         string    `json:"body"`
	ParentID     *string   `json:"parent_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// CommunityFeedTab 广场列表 tab
type CommunityFeedTab string

const (
	CommunityFeedGlobal    CommunityFeedTab = "global"
	CommunityFeedFollowing CommunityFeedTab = "following"
	CommunityFeedGroup     CommunityFeedTab = "group"
)
