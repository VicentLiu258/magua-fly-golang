package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"

	"github.com/jackc/pgx/v5"
)

// CommunityUseCase 社区（动态、关注、话题、小组、挑战赛）
type CommunityUseCase struct {
	users     repository.UserRepository
	follow    repository.CommunityFollowRepository
	topics    repository.CommunityTopicRepository
	groups    repository.CommunityGroupRepository
	posts     repository.CommunityPostRepository
	likes     repository.CommunityLikeRepository
	comments  repository.CommunityCommentRepository
	challenge repository.CommunityChallengeRepository
}

func NewCommunityUseCase(
	users repository.UserRepository,
	follow repository.CommunityFollowRepository,
	topics repository.CommunityTopicRepository,
	groups repository.CommunityGroupRepository,
	posts repository.CommunityPostRepository,
	likes repository.CommunityLikeRepository,
	comments repository.CommunityCommentRepository,
	challenge repository.CommunityChallengeRepository,
) *CommunityUseCase {
	return &CommunityUseCase{
		users: users, follow: follow, topics: topics, groups: groups,
		posts: posts, likes: likes, comments: comments, challenge: challenge,
	}
}

func isAdminOrStaff(r domain.UserRole) bool {
	return r == domain.RoleAdmin || r == domain.RoleStaff
}

// UserBrief 社区展示用用户摘要
type UserBrief struct {
	ID        string          `json:"id"`
	Username  string          `json:"username"`
	AvatarURL string          `json:"avatar_url,omitempty"`
	Role      domain.UserRole `json:"role"`
}

// PostItem 动态列表项
type PostItem struct {
	Post   *domain.CommunityPost `json:"post"`
	Author *UserBrief            `json:"author"`
}

func clampPageSize(n int) int {
	if n <= 0 {
		return 20
	}
	if n > 50 {
		return 50
	}
	return n
}

func parseCursorTime(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339Nano, *s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, *s)
	}
	if err != nil {
		return nil, ErrInvalidArgument
	}
	return &t, nil
}

func dedupeStrings(in []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func (uc *CommunityUseCase) loadUserBriefs(ctx context.Context, ids []string) (map[string]*UserBrief, error) {
	ids = dedupeStrings(ids)
	m := make(map[string]*UserBrief)
	for _, id := range ids {
		u, err := uc.users.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return nil, err
		}
		m[id] = &UserBrief{ID: u.ID, Username: u.Username, AvatarURL: u.AvatarURL, Role: u.Role}
	}
	return m, nil
}

func (uc *CommunityUseCase) attachAuthors(ctx context.Context, posts []*domain.CommunityPost) ([]PostItem, error) {
	if len(posts) == 0 {
		return nil, nil
	}
	var ids []string
	for _, p := range posts {
		ids = append(ids, p.AuthorUserID)
	}
	briefs, err := uc.loadUserBriefs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]PostItem, 0, len(posts))
	for _, p := range posts {
		item := PostItem{Post: p, Author: briefs[p.AuthorUserID]}
		out = append(out, item)
	}
	return out, nil
}

func (uc *CommunityUseCase) fillLikedAndTopics(ctx context.Context, viewerID string, posts []*domain.CommunityPost) error {
	if len(posts) == 0 {
		return nil
	}
	ids := make([]string, len(posts))
	for i, p := range posts {
		ids[i] = p.ID
	}
	liked, err := uc.posts.BatchLikedByUser(ctx, viewerID, ids)
	if err != nil {
		return err
	}
	for _, p := range posts {
		p.LikedByMe = liked[p.ID]
		topicIDs, err := uc.posts.ListTopicIDs(ctx, p.ID)
		if err != nil {
			return err
		}
		p.TopicIDs = topicIDs
	}
	return nil
}

// CreatePostRequest 发布飞行故事
type CreatePostRequest struct {
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	Visibility  string   `json:"visibility"`
	GroupID     *string  `json:"group_id"`
	MediaURLs   []string `json:"media_urls"`
	TopicIDs    []string `json:"topic_ids"`
}

func (uc *CommunityUseCase) CreatePost(ctx context.Context, authorID string, req *CreatePostRequest) (*PostItem, error) {
	if req == nil || strings.TrimSpace(req.Body) == "" {
		return nil, ErrInvalidArgument
	}
	vis := domain.CommunityPostVisibility(req.Visibility)
	if vis != domain.CommunityVisibilityPublic && vis != domain.CommunityVisibilityFollowersOnly && vis != domain.CommunityVisibilityGroupOnly {
		vis = domain.CommunityVisibilityPublic
	}
	if vis == domain.CommunityVisibilityGroupOnly {
		if req.GroupID == nil || *req.GroupID == "" {
			return nil, ErrInvalidArgument
		}
		ok, err := uc.groups.IsMember(ctx, *req.GroupID, authorID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrForbidden
		}
	}
	topicIDs := dedupeStrings(req.TopicIDs)
	if len(topicIDs) > 0 {
		found, err := uc.topics.ListByIDs(ctx, topicIDs)
		if err != nil {
			return nil, err
		}
		if len(found) != len(topicIDs) {
			return nil, ErrInvalidArgument
		}
	}
	p := &domain.CommunityPost{
		AuthorUserID: authorID,
		Title:        strings.TrimSpace(req.Title),
		Body:         strings.TrimSpace(req.Body),
		Visibility:   vis,
		GroupID:      req.GroupID,
		Status:       domain.CommunityPostPublished,
		MediaURLs:    req.MediaURLs,
	}
	if err := uc.posts.CreateWithTopics(ctx, p, topicIDs); err != nil {
		return nil, err
	}
	p.TopicIDs = topicIDs
	items, err := uc.attachAuthors(ctx, []*domain.CommunityPost{p})
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return &items[0], nil
}

// UpdatePostRequest 更新自己的动态
type UpdatePostRequest struct {
	PostID     string   `json:"post_id"`
	Title      string   `json:"title"`
	Body       string   `json:"body"`
	Visibility string   `json:"visibility"`
	GroupID    *string  `json:"group_id"`
	Status     string   `json:"status"`
	MediaURLs  []string `json:"media_urls"`
	TopicIDs   []string `json:"topic_ids"`
}

func (uc *CommunityUseCase) UpdatePost(ctx context.Context, userID string, req *UpdatePostRequest) (*PostItem, error) {
	if req == nil || req.PostID == "" {
		return nil, ErrInvalidArgument
	}
	old, err := uc.posts.GetByID(ctx, req.PostID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if old.AuthorUserID != userID {
		return nil, ErrForbidden
	}
	if strings.TrimSpace(req.Body) == "" {
		return nil, ErrInvalidArgument
	}
	vis := domain.CommunityPostVisibility(req.Visibility)
	if req.Visibility == "" {
		vis = old.Visibility
	}
	if vis != domain.CommunityVisibilityPublic && vis != domain.CommunityVisibilityFollowersOnly && vis != domain.CommunityVisibilityGroupOnly {
		return nil, ErrInvalidArgument
	}
	if vis == domain.CommunityVisibilityGroupOnly {
		gid := req.GroupID
		if gid == nil || *gid == "" {
			gid = old.GroupID
		}
		if gid == nil || *gid == "" {
			return nil, ErrInvalidArgument
		}
		ok, err := uc.groups.IsMember(ctx, *gid, userID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrForbidden
		}
		old.GroupID = gid
	}
	st := domain.CommunityPostStatus(req.Status)
	if req.Status == "" {
		st = old.Status
	}
	if st != domain.CommunityPostPublished && st != domain.CommunityPostHidden {
		return nil, ErrInvalidArgument
	}
	old.Title = strings.TrimSpace(req.Title)
	old.Body = strings.TrimSpace(req.Body)
	old.Visibility = vis
	if req.GroupID != nil {
		old.GroupID = req.GroupID
	}
	if old.Visibility != domain.CommunityVisibilityGroupOnly {
		old.GroupID = nil
	}
	old.Status = st
	old.MediaURLs = req.MediaURLs
	if err := uc.posts.Update(ctx, old); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	topicIDs := dedupeStrings(req.TopicIDs)
	if req.TopicIDs != nil {
		if len(topicIDs) > 0 {
			found, err := uc.topics.ListByIDs(ctx, topicIDs)
			if err != nil {
				return nil, err
			}
			if len(found) != len(topicIDs) {
				return nil, ErrInvalidArgument
			}
		}
		if err := uc.posts.ReplaceTopicLinks(ctx, old.ID, topicIDs); err != nil {
			return nil, err
		}
		old.TopicIDs = topicIDs
	}
	return uc.GetPost(ctx, userID, old.ID)
}

func (uc *CommunityUseCase) DeletePost(ctx context.Context, userID, postID string) error {
	p, err := uc.posts.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if p.AuthorUserID != userID {
		return ErrForbidden
	}
	if err := uc.posts.Delete(ctx, postID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (uc *CommunityUseCase) viewerCanSeePost(ctx context.Context, viewerID string, post *domain.CommunityPost) (bool, error) {
	if post.Status == domain.CommunityPostHidden && post.AuthorUserID != viewerID {
		return false, nil
	}
	if post.AuthorUserID == viewerID {
		return true, nil
	}
	if post.Status != domain.CommunityPostPublished {
		return false, nil
	}
	switch post.Visibility {
	case domain.CommunityVisibilityPublic:
		return true, nil
	case domain.CommunityVisibilityFollowersOnly:
		return uc.follow.IsFollowing(ctx, viewerID, post.AuthorUserID)
	case domain.CommunityVisibilityGroupOnly:
		if post.GroupID == nil {
			return false, nil
		}
		return uc.groups.IsMember(ctx, *post.GroupID, viewerID)
	default:
		return false, nil
	}
}

// GetPost 动态详情
func (uc *CommunityUseCase) GetPost(ctx context.Context, viewerID, postID string) (*PostItem, error) {
	p, err := uc.posts.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	ok, err := uc.viewerCanSeePost(ctx, viewerID, p)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	if err := uc.fillLikedAndTopics(ctx, viewerID, []*domain.CommunityPost{p}); err != nil {
		return nil, err
	}
	items, err := uc.attachAuthors(ctx, []*domain.CommunityPost{p})
	if err != nil || len(items) == 0 {
		return nil, err
	}
	return &items[0], nil
}

// ListPostsRequest 广场列表
type ListPostsRequest struct {
	Tab               string  `json:"tab"`
	GroupID           string  `json:"group_id"`
	TopicID           string  `json:"topic_id"`
	AuthorUserID      string  `json:"author_user_id"`
	PageSize          int     `json:"page_size"`
	CursorCreatedAt   *string `json:"cursor_created_at"`
	CursorID          *string `json:"cursor_id"`
	NextCursor        *string `json:"next_cursor"`
}

// ListPostsResponse 分页
type ListPostsResponse struct {
	Items      []PostItem `json:"items"`
	NextCursor *string    `json:"next_cursor,omitempty"`
}

func encodeNextCursor(t time.Time, id string) *string {
	s := t.UTC().Format(time.RFC3339Nano) + "|" + id
	return &s
}

func decodeNextCursor(s *string) (*time.Time, *string, error) {
	if s == nil || *s == "" {
		return nil, nil, nil
	}
	parts := strings.SplitN(*s, "|", 2)
	if len(parts) != 2 {
		// 兼容仅传 RFC3339 + id 分字段的旧客户端：仍支持 ListPostsRequest 双字段
		return nil, nil, ErrInvalidArgument
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		t, err = time.Parse(time.RFC3339, parts[0])
	}
	if err != nil {
		return nil, nil, ErrInvalidArgument
	}
	id := parts[1]
	return &t, &id, nil
}

func (uc *CommunityUseCase) ListPosts(ctx context.Context, viewerID string, req *ListPostsRequest) (*ListPostsResponse, error) {
	if req == nil {
		req = &ListPostsRequest{}
	}
	ps := clampPageSize(req.PageSize)
	var cAt *time.Time
	var cID *string
	if req.NextCursor != nil && *req.NextCursor != "" {
		t, id, err := decodeNextCursor(req.NextCursor)
		if err != nil {
			return nil, err
		}
		cAt, cID = t, id
	} else if req.CursorCreatedAt != nil || (req.CursorID != nil && !strings.Contains(*req.CursorID, "|")) {
		var err error
		cAt, err = parseCursorTime(req.CursorCreatedAt)
		if err != nil {
			return nil, err
		}
		cID = req.CursorID
	} else if req.CursorID != nil && strings.Contains(*req.CursorID, "|") {
		t, id, err := decodeNextCursor(req.CursorID)
		if err != nil {
			return nil, err
		}
		cAt, cID = t, id
	}
	var posts []*domain.CommunityPost
	var err error
	if req.TopicID != "" {
		posts, err = uc.posts.ListByTopic(ctx, req.TopicID, ps+1, cAt, cID)
	} else if req.AuthorUserID != "" {
		posts, err = uc.posts.ListByAuthor(ctx, req.AuthorUserID, ps+1, cAt, cID)
	} else {
		tab := strings.TrimSpace(req.Tab)
		if tab == "" {
			tab = string(domain.CommunityFeedGlobal)
		}
		switch tab {
		case string(domain.CommunityFeedGlobal):
			posts, err = uc.posts.ListFeedGlobal(ctx, ps+1, cAt, cID)
		case string(domain.CommunityFeedFollowing):
			posts, err = uc.posts.ListFeedFollowing(ctx, viewerID, ps+1, cAt, cID)
		case string(domain.CommunityFeedGroup):
			if req.GroupID == "" {
				return nil, ErrInvalidArgument
			}
			posts, err = uc.posts.ListFeedGroup(ctx, req.GroupID, viewerID, ps+1, cAt, cID)
		default:
			return nil, ErrInvalidArgument
		}
	}
	if err != nil {
		return nil, err
	}
	hasMore := len(posts) > ps
	if hasMore {
		posts = posts[:ps]
	}
	if req.AuthorUserID != "" {
		filtered := posts[:0]
		for _, p := range posts {
			ok, err := uc.viewerCanSeePost(ctx, viewerID, p)
			if err != nil {
				return nil, err
			}
			if ok {
				filtered = append(filtered, p)
			}
		}
		posts = filtered
	}
	if err := uc.fillLikedAndTopics(ctx, viewerID, posts); err != nil {
		return nil, err
	}
	items, err := uc.attachAuthors(ctx, posts)
	if err != nil {
		return nil, err
	}
	resp := &ListPostsResponse{Items: items}
	if hasMore && len(posts) > 0 {
		last := posts[len(posts)-1]
		resp.NextCursor = encodeNextCursor(last.CreatedAt, last.ID)
	}
	return resp, nil
}

// FollowRequest 关注
type FollowRequest struct {
	FolloweeID string `json:"followee_id"`
}

func (uc *CommunityUseCase) Follow(ctx context.Context, followerID string, req *FollowRequest) error {
	if req == nil || req.FolloweeID == "" {
		return ErrInvalidArgument
	}
	if followerID == req.FolloweeID {
		return ErrInvalidArgument
	}
	_, err := uc.users.GetByID(ctx, req.FolloweeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return uc.follow.Follow(ctx, followerID, req.FolloweeID)
}

func (uc *CommunityUseCase) Unfollow(ctx context.Context, followerID string, req *FollowRequest) error {
	if req == nil || req.FolloweeID == "" {
		return ErrInvalidArgument
	}
	return uc.follow.Unfollow(ctx, followerID, req.FolloweeID)
}

type ListFollowRequest struct {
	UserID   string `json:"user_id"`
	PageSize int    `json:"page_size"`
	Offset   int    `json:"offset"`
}

type ListFollowResponse struct {
	UserIDs []string    `json:"user_ids"`
	Users   []UserBrief `json:"users"`
}

func (uc *CommunityUseCase) ListFollowers(ctx context.Context, req *ListFollowRequest) (*ListFollowResponse, error) {
	if req == nil || req.UserID == "" {
		return nil, ErrInvalidArgument
	}
	ps := clampPageSize(req.PageSize)
	ids, err := uc.follow.ListFollowers(ctx, req.UserID, ps, req.Offset)
	if err != nil {
		return nil, err
	}
	briefs, err := uc.loadUserBriefs(ctx, ids)
	if err != nil {
		return nil, err
	}
	var users []UserBrief
	for _, id := range ids {
		if b := briefs[id]; b != nil {
			users = append(users, *b)
		}
	}
	return &ListFollowResponse{UserIDs: ids, Users: users}, nil
}

func (uc *CommunityUseCase) ListFollowing(ctx context.Context, req *ListFollowRequest) (*ListFollowResponse, error) {
	if req == nil || req.UserID == "" {
		return nil, ErrInvalidArgument
	}
	ps := clampPageSize(req.PageSize)
	ids, err := uc.follow.ListFollowing(ctx, req.UserID, ps, req.Offset)
	if err != nil {
		return nil, err
	}
	briefs, err := uc.loadUserBriefs(ctx, ids)
	if err != nil {
		return nil, err
	}
	var users []UserBrief
	for _, id := range ids {
		if b := briefs[id]; b != nil {
			users = append(users, *b)
		}
	}
	return &ListFollowResponse{UserIDs: ids, Users: users}, nil
}

type RecommendFollowRequest struct {
	Limit int `json:"limit"`
}

type RecommendFollowResponse struct {
	Users []UserBrief `json:"users"`
}

func (uc *CommunityUseCase) RecommendFollow(ctx context.Context, viewerID string, req *RecommendFollowRequest) (*RecommendFollowResponse, error) {
	limit := 10
	if req != nil && req.Limit > 0 {
		limit = req.Limit
		if limit > 50 {
			limit = 50
		}
	}
	ids, err := uc.follow.RecommendUserIDs(ctx, viewerID, limit)
	if err != nil {
		return nil, err
	}
	briefs, err := uc.loadUserBriefs(ctx, ids)
	if err != nil {
		return nil, err
	}
	var users []UserBrief
	for _, id := range ids {
		if b := briefs[id]; b != nil {
			users = append(users, *b)
		}
	}
	return &RecommendFollowResponse{Users: users}, nil
}

type ListTopicsRequest struct {
	PageSize int `json:"page_size"`
	Offset   int `json:"offset"`
}

func (uc *CommunityUseCase) ListTopics(ctx context.Context, req *ListTopicsRequest) ([]*domain.CommunityTopic, error) {
	ps := 20
	off := 0
	if req != nil {
		if req.PageSize > 0 {
			ps = clampPageSize(req.PageSize)
		}
		off = req.Offset
	}
	return uc.topics.ListHot(ctx, ps, off)
}

type TopicDetailResponse struct {
	Topic *domain.CommunityTopic `json:"topic"`
	Posts *ListPostsResponse     `json:"posts"`
}

func (uc *CommunityUseCase) GetTopicDetail(ctx context.Context, viewerID string, req *ListPostsRequest) (*TopicDetailResponse, error) {
	if req == nil || req.TopicID == "" {
		return nil, ErrInvalidArgument
	}
	t, err := uc.topics.GetByID(ctx, req.TopicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	list, err := uc.ListPosts(ctx, viewerID, req)
	if err != nil {
		return nil, err
	}
	return &TopicDetailResponse{Topic: t, Posts: list}, nil
}

// CreateTopicRequest 运营创建话题
type CreateTopicRequest struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
}

func (uc *CommunityUseCase) CreateTopic(ctx context.Context, operatorID string, req *CreateTopicRequest) (*domain.CommunityTopic, error) {
	op, err := uc.users.GetByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	if !isAdminOrStaff(op.Role) {
		return nil, ErrForbidden
	}
	if req == nil || strings.TrimSpace(req.Slug) == "" || strings.TrimSpace(req.Title) == "" {
		return nil, ErrInvalidArgument
	}
	top := &domain.CommunityTopic{
		Slug:        strings.TrimSpace(req.Slug),
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		CoverURL:    strings.TrimSpace(req.CoverURL),
	}
	if err := uc.topics.Create(ctx, top); err != nil {
		return nil, err
	}
	return top, nil
}

type CreateGroupRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
	IsPublic    *bool  `json:"is_public"`
}

func (uc *CommunityUseCase) CreateGroup(ctx context.Context, ownerID string, req *CreateGroupRequest) (*domain.CommunityGroup, error) {
	if req == nil || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Slug) == "" {
		return nil, ErrInvalidArgument
	}
	pub := true
	if req.IsPublic != nil {
		pub = *req.IsPublic
	}
	g := &domain.CommunityGroup{
		Name:        strings.TrimSpace(req.Name),
		Slug:        strings.TrimSpace(req.Slug),
		Description: strings.TrimSpace(req.Description),
		CoverURL:    strings.TrimSpace(req.CoverURL),
		OwnerUserID: ownerID,
		IsPublic:    pub,
	}
	if err := uc.groups.Create(ctx, g); err != nil {
		return nil, err
	}
	if err := uc.groups.AddMember(ctx, g.ID, ownerID, domain.CommunityGroupRoleAdmin); err != nil {
		return nil, err
	}
	return g, nil
}

type UpdateGroupRequest struct {
	GroupID     string `json:"group_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
	IsPublic    *bool  `json:"is_public"`
}

func (uc *CommunityUseCase) UpdateGroup(ctx context.Context, userID string, req *UpdateGroupRequest) (*domain.CommunityGroup, error) {
	if req == nil || req.GroupID == "" {
		return nil, ErrInvalidArgument
	}
	g, err := uc.groups.GetByID(ctx, req.GroupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	m, err := uc.groups.GetMembership(ctx, req.GroupID, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	can := g.OwnerUserID == userID || (m != nil && m.Role == domain.CommunityGroupRoleAdmin)
	if !can {
		return nil, ErrForbidden
	}
	if strings.TrimSpace(req.Name) != "" {
		g.Name = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.Slug) != "" {
		g.Slug = strings.TrimSpace(req.Slug)
	}
	g.Description = strings.TrimSpace(req.Description)
	g.CoverURL = strings.TrimSpace(req.CoverURL)
	if req.IsPublic != nil {
		g.IsPublic = *req.IsPublic
	}
	if err := uc.groups.Update(ctx, g); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return g, nil
}

type GroupIDRequest struct {
	GroupID string `json:"group_id"`
}

func (uc *CommunityUseCase) GetGroup(ctx context.Context, groupID string) (*domain.CommunityGroup, error) {
	if groupID == "" {
		return nil, ErrInvalidArgument
	}
	g, err := uc.groups.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return g, nil
}

func (uc *CommunityUseCase) ListGroups(ctx context.Context, req *ListTopicsRequest) ([]*domain.CommunityGroup, error) {
	ps := 20
	off := 0
	if req != nil {
		if req.PageSize > 0 {
			ps = clampPageSize(req.PageSize)
		}
		off = req.Offset
	}
	return uc.groups.List(ctx, ps, off)
}

func (uc *CommunityUseCase) JoinGroup(ctx context.Context, userID string, req *GroupIDRequest) error {
	if req == nil || req.GroupID == "" {
		return ErrInvalidArgument
	}
	g, err := uc.groups.GetByID(ctx, req.GroupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if !g.IsPublic {
		return ErrForbidden
	}
	return uc.groups.AddMember(ctx, req.GroupID, userID, domain.CommunityGroupRoleMember)
}

func (uc *CommunityUseCase) LeaveGroup(ctx context.Context, userID string, req *GroupIDRequest) error {
	if req == nil || req.GroupID == "" {
		return ErrInvalidArgument
	}
	g, err := uc.groups.GetByID(ctx, req.GroupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if g.OwnerUserID == userID {
		return ErrForbidden
	}
	return uc.groups.RemoveMember(ctx, req.GroupID, userID)
}

func (uc *CommunityUseCase) ListGroupMembers(ctx context.Context, req *GroupIDRequest, pageSize, offset int) (*ListFollowResponse, error) {
	if req == nil || req.GroupID == "" {
		return nil, ErrInvalidArgument
	}
	ps := clampPageSize(pageSize)
	members, err := uc.groups.ListMembers(ctx, req.GroupID, ps, offset)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, m := range members {
		ids = append(ids, m.UserID)
	}
	briefs, err := uc.loadUserBriefs(ctx, ids)
	if err != nil {
		return nil, err
	}
	var users []UserBrief
	for _, m := range members {
		if b := briefs[m.UserID]; b != nil {
			users = append(users, *b)
		}
	}
	return &ListFollowResponse{UserIDs: ids, Users: users}, nil
}

type LikePostRequest struct {
	PostID string `json:"post_id"`
}

func (uc *CommunityUseCase) LikePost(ctx context.Context, userID string, req *LikePostRequest) error {
	if req == nil || req.PostID == "" {
		return ErrInvalidArgument
	}
	p, err := uc.posts.GetByID(ctx, req.PostID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	ok, err := uc.viewerCanSeePost(ctx, userID, p)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return uc.likes.Like(ctx, req.PostID, userID)
}

func (uc *CommunityUseCase) UnlikePost(ctx context.Context, userID string, req *LikePostRequest) error {
	if req == nil || req.PostID == "" {
		return ErrInvalidArgument
	}
	return uc.likes.Unlike(ctx, req.PostID, userID)
}

type CreateCommentRequest struct {
	PostID   string  `json:"post_id"`
	Body     string  `json:"body"`
	ParentID *string `json:"parent_id"`
}

func (uc *CommunityUseCase) CreateComment(ctx context.Context, userID string, req *CreateCommentRequest) (*domain.CommunityComment, error) {
	if req == nil || req.PostID == "" || strings.TrimSpace(req.Body) == "" {
		return nil, ErrInvalidArgument
	}
	p, err := uc.posts.GetByID(ctx, req.PostID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	ok, err := uc.viewerCanSeePost(ctx, userID, p)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	c := &domain.CommunityComment{
		PostID:       req.PostID,
		AuthorUserID: userID,
		Body:         strings.TrimSpace(req.Body),
		ParentID:     req.ParentID,
	}
	if err := uc.comments.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

type ListCommentsRequest struct {
	PostID   string `json:"post_id"`
	PageSize int    `json:"page_size"`
	Offset   int    `json:"offset"`
}

func (uc *CommunityUseCase) ListComments(ctx context.Context, viewerID string, req *ListCommentsRequest) ([]*domain.CommunityComment, error) {
	if req == nil || req.PostID == "" {
		return nil, ErrInvalidArgument
	}
	p, err := uc.posts.GetByID(ctx, req.PostID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	ok, err := uc.viewerCanSeePost(ctx, viewerID, p)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	ps := clampPageSize(req.PageSize)
	return uc.comments.ListByPost(ctx, req.PostID, ps, req.Offset)
}

type CreateChallengeRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Rules       string `json:"rules"`
	CoverURL    string `json:"cover_url"`
	StartsAt    string `json:"starts_at"`
	EndsAt      string `json:"ends_at"`
	Status      string `json:"status"`
}

func (uc *CommunityUseCase) CreateChallenge(ctx context.Context, operatorID string, req *CreateChallengeRequest) (*domain.CommunityChallenge, error) {
	op, err := uc.users.GetByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	if !isAdminOrStaff(op.Role) {
		return nil, ErrForbidden
	}
	if req == nil || strings.TrimSpace(req.Title) == "" || req.StartsAt == "" || req.EndsAt == "" {
		return nil, ErrInvalidArgument
	}
	start, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	end, err := time.Parse(time.RFC3339, req.EndsAt)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	st := domain.CommunityChallengeStatus(req.Status)
	if st == "" {
		st = domain.CommunityChallengeDraft
	}
	if st != domain.CommunityChallengeDraft && st != domain.CommunityChallengeActive && st != domain.CommunityChallengeEnded {
		return nil, ErrInvalidArgument
	}
	ch := &domain.CommunityChallenge{
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Rules:       strings.TrimSpace(req.Rules),
		CoverURL:    strings.TrimSpace(req.CoverURL),
		StartsAt:    start,
		EndsAt:      end,
		Status:      st,
	}
	if err := uc.challenge.Create(ctx, ch); err != nil {
		return nil, err
	}
	return ch, nil
}

type UpdateChallengeRequest struct {
	ChallengeID string `json:"challenge_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Rules       string `json:"rules"`
	CoverURL    string `json:"cover_url"`
	StartsAt    string `json:"starts_at"`
	EndsAt      string `json:"ends_at"`
	Status      string `json:"status"`
}

func (uc *CommunityUseCase) UpdateChallenge(ctx context.Context, operatorID string, req *UpdateChallengeRequest) (*domain.CommunityChallenge, error) {
	op, err := uc.users.GetByID(ctx, operatorID)
	if err != nil {
		return nil, err
	}
	if !isAdminOrStaff(op.Role) {
		return nil, ErrForbidden
	}
	if req == nil || req.ChallengeID == "" {
		return nil, ErrInvalidArgument
	}
	ch, err := uc.challenge.GetByID(ctx, req.ChallengeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if strings.TrimSpace(req.Title) != "" {
		ch.Title = strings.TrimSpace(req.Title)
	}
	ch.Description = strings.TrimSpace(req.Description)
	ch.Rules = strings.TrimSpace(req.Rules)
	ch.CoverURL = strings.TrimSpace(req.CoverURL)
	if req.StartsAt != "" {
		t, err := time.Parse(time.RFC3339, req.StartsAt)
		if err != nil {
			return nil, ErrInvalidArgument
		}
		ch.StartsAt = t
	}
	if req.EndsAt != "" {
		t, err := time.Parse(time.RFC3339, req.EndsAt)
		if err != nil {
			return nil, ErrInvalidArgument
		}
		ch.EndsAt = t
	}
	if req.Status != "" {
		st := domain.CommunityChallengeStatus(req.Status)
		if st != domain.CommunityChallengeDraft && st != domain.CommunityChallengeActive && st != domain.CommunityChallengeEnded {
			return nil, ErrInvalidArgument
		}
		ch.Status = st
	}
	if err := uc.challenge.Update(ctx, ch); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return ch, nil
}

type ListChallengesRequest struct {
	Status   string `json:"status"`
	PageSize int    `json:"page_size"`
	Offset   int    `json:"offset"`
}

func (uc *CommunityUseCase) ListChallenges(ctx context.Context, req *ListChallengesRequest) ([]*domain.CommunityChallenge, error) {
	var st *domain.CommunityChallengeStatus
	if req != nil && req.Status != "" {
		s := domain.CommunityChallengeStatus(req.Status)
		st = &s
	}
	ps := 20
	off := 0
	if req != nil {
		if req.PageSize > 0 {
			ps = clampPageSize(req.PageSize)
		}
		off = req.Offset
	}
	return uc.challenge.List(ctx, st, ps, off)
}

func (uc *CommunityUseCase) GetChallenge(ctx context.Context, id string) (*domain.CommunityChallenge, error) {
	if id == "" {
		return nil, ErrInvalidArgument
	}
	ch, err := uc.challenge.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return ch, nil
}

type SubmitChallengeRequest struct {
	ChallengeID string `json:"challenge_id"`
	PostID      string `json:"post_id"`
}

func (uc *CommunityUseCase) SubmitChallenge(ctx context.Context, userID string, req *SubmitChallengeRequest) error {
	if req == nil || req.ChallengeID == "" || req.PostID == "" {
		return ErrInvalidArgument
	}
	ch, err := uc.challenge.GetByID(ctx, req.ChallengeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if ch.Status != domain.CommunityChallengeActive {
		return ErrForbidden
	}
	now := time.Now()
	if now.Before(ch.StartsAt) || now.After(ch.EndsAt) {
		return ErrForbidden
	}
	p, err := uc.posts.GetByID(ctx, req.PostID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if p.AuthorUserID != userID {
		return ErrForbidden
	}
	ok, err := uc.viewerCanSeePost(ctx, userID, p)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return uc.challenge.SubmitEntry(ctx, &domain.CommunityChallengeEntry{
		ChallengeID: req.ChallengeID,
		UserID:      userID,
		PostID:      req.PostID,
	})
}

type ListChallengeEntriesRequest struct {
	ChallengeID string `json:"challenge_id"`
	PageSize    int    `json:"page_size"`
	Offset      int    `json:"offset"`
}

func (uc *CommunityUseCase) ListChallengeEntries(ctx context.Context, req *ListChallengeEntriesRequest) ([]*domain.CommunityChallengeEntry, error) {
	if req == nil || req.ChallengeID == "" {
		return nil, ErrInvalidArgument
	}
	ps := clampPageSize(req.PageSize)
	return uc.challenge.ListEntries(ctx, req.ChallengeID, ps, req.Offset)
}
