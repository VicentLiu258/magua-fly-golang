package handler

import (
	"errors"
	"net/http"

	"magua-fly-golang/internal/usecase"

	"github.com/gin-gonic/gin"
)

// CommunityHandler 社区 API
type CommunityHandler struct {
	uc *usecase.CommunityUseCase
}

func NewCommunityHandler(uc *usecase.CommunityUseCase) *CommunityHandler {
	return &CommunityHandler{uc: uc}
}

func communityHTTPStatus(err error) int {
	switch {
	case errors.Is(err, usecase.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, usecase.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, usecase.ErrInvalidArgument):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (h *CommunityHandler) respondErr(c *gin.Context, err error) {
	st := communityHTTPStatus(err)
	c.JSON(st, gin.H{"code": st, "msg": err.Error(), "data": nil})
}

// --- posts ---

func (h *CommunityHandler) PostsCreate(c *gin.Context) {
	var body usecase.CreatePostRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	data, err := h.uc.CreatePost(c.Request.Context(), uid, &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) PostsUpdate(c *gin.Context) {
	var body usecase.UpdatePostRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	data, err := h.uc.UpdatePost(c.Request.Context(), uid, &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) PostsDelete(c *gin.Context) {
	var body struct {
		PostID string `json:"post_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if err := h.uc.DeletePost(c.Request.Context(), uid, body.PostID); err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"status": "ok"}})
}

func (h *CommunityHandler) PostsDetail(c *gin.Context) {
	var body struct {
		PostID string `json:"post_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	data, err := h.uc.GetPost(c.Request.Context(), uid, body.PostID)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) PostsList(c *gin.Context) {
	var body usecase.ListPostsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	data, err := h.uc.ListPosts(c.Request.Context(), uid, &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) PostsLike(c *gin.Context) {
	var body usecase.LikePostRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if err := h.uc.LikePost(c.Request.Context(), uid, &body); err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"status": "ok"}})
}

func (h *CommunityHandler) PostsUnlike(c *gin.Context) {
	var body usecase.LikePostRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if err := h.uc.UnlikePost(c.Request.Context(), uid, &body); err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"status": "ok"}})
}

// --- comments ---

func (h *CommunityHandler) CommentsCreate(c *gin.Context) {
	var body usecase.CreateCommentRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	data, err := h.uc.CreateComment(c.Request.Context(), uid, &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) CommentsList(c *gin.Context) {
	var body usecase.ListCommentsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	data, err := h.uc.ListComments(c.Request.Context(), uid, &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

// --- follows ---

func (h *CommunityHandler) FollowsFollow(c *gin.Context) {
	var body usecase.FollowRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if err := h.uc.Follow(c.Request.Context(), uid, &body); err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"status": "ok"}})
}

func (h *CommunityHandler) FollowsUnfollow(c *gin.Context) {
	var body usecase.FollowRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if err := h.uc.Unfollow(c.Request.Context(), uid, &body); err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"status": "ok"}})
}

func (h *CommunityHandler) FollowsFollowers(c *gin.Context) {
	var body usecase.ListFollowRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	data, err := h.uc.ListFollowers(c.Request.Context(), &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) FollowsFollowing(c *gin.Context) {
	var body usecase.ListFollowRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	data, err := h.uc.ListFollowing(c.Request.Context(), &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) FollowsRecommend(c *gin.Context) {
	var body usecase.RecommendFollowRequest
	_ = c.ShouldBindJSON(&body)
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	data, err := h.uc.RecommendFollow(c.Request.Context(), uid, &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

// --- topics ---

func (h *CommunityHandler) TopicsList(c *gin.Context) {
	var body usecase.ListTopicsRequest
	_ = c.ShouldBindJSON(&body)
	data, err := h.uc.ListTopics(c.Request.Context(), &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) TopicsDetail(c *gin.Context) {
	var body usecase.ListPostsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	data, err := h.uc.GetTopicDetail(c.Request.Context(), uid, &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) TopicsCreate(c *gin.Context) {
	var body usecase.CreateTopicRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	data, err := h.uc.CreateTopic(c.Request.Context(), uid, &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) GroupsCreate(c *gin.Context) {
	var body usecase.CreateGroupRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	data, err := h.uc.CreateGroup(c.Request.Context(), uid, &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) GroupsUpdate(c *gin.Context) {
	var body usecase.UpdateGroupRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	data, err := h.uc.UpdateGroup(c.Request.Context(), uid, &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) GroupsDetail(c *gin.Context) {
	var body usecase.GroupIDRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	data, err := h.uc.GetGroup(c.Request.Context(), body.GroupID)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) GroupsList(c *gin.Context) {
	var body usecase.ListTopicsRequest
	_ = c.ShouldBindJSON(&body)
	data, err := h.uc.ListGroups(c.Request.Context(), &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) GroupsJoin(c *gin.Context) {
	var body usecase.GroupIDRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if err := h.uc.JoinGroup(c.Request.Context(), uid, &body); err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"status": "ok"}})
}

func (h *CommunityHandler) GroupsLeave(c *gin.Context) {
	var body usecase.GroupIDRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if err := h.uc.LeaveGroup(c.Request.Context(), uid, &body); err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"status": "ok"}})
}

func (h *CommunityHandler) GroupsMembers(c *gin.Context) {
	var body struct {
		GroupID  string `json:"group_id"`
		PageSize int    `json:"page_size"`
		Offset   int    `json:"offset"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	data, err := h.uc.ListGroupMembers(c.Request.Context(), &usecase.GroupIDRequest{GroupID: body.GroupID}, body.PageSize, body.Offset)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) ChallengesCreate(c *gin.Context) {
	var body usecase.CreateChallengeRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	data, err := h.uc.CreateChallenge(c.Request.Context(), uid, &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) ChallengesUpdate(c *gin.Context) {
	var body usecase.UpdateChallengeRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	data, err := h.uc.UpdateChallenge(c.Request.Context(), uid, &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) ChallengesList(c *gin.Context) {
	var body usecase.ListChallengesRequest
	_ = c.ShouldBindJSON(&body)
	data, err := h.uc.ListChallenges(c.Request.Context(), &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) ChallengesDetail(c *gin.Context) {
	var body struct {
		ChallengeID string `json:"challenge_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	data, err := h.uc.GetChallenge(c.Request.Context(), body.ChallengeID)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}

func (h *CommunityHandler) ChallengesSubmit(c *gin.Context) {
	var body usecase.SubmitChallengeRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	uid := getUserIDFromClaims(c)
	if uid == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized", "data": nil})
		return
	}
	if err := h.uc.SubmitChallenge(c.Request.Context(), uid, &body); err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": gin.H{"status": "ok"}})
}

func (h *CommunityHandler) ChallengesEntries(c *gin.Context) {
	var body usecase.ListChallengeEntriesRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error(), "data": nil})
		return
	}
	data, err := h.uc.ListChallengeEntries(c.Request.Context(), &body)
	if err != nil {
		h.respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "ok", "data": data})
}
