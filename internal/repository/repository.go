package repository

import (
	"context"
	"errors"
	"time"

	"magua-fly-golang/internal/domain"
)

// ErrNotFound 仓储层统一「不存在」哨兵，可与业务层 errors.Is 配合。
var ErrNotFound = errors.New("not found")

// ErrPersonnelStaffRole 使用 kind=staff 删除时目标用户不是 staff/maintenance。
var ErrPersonnelStaffRole = errors.New("personnel staff role mismatch")

// UserRepository defines user persistence.
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
	Create(ctx context.Context, u *domain.User) error
	// Update 全量更新可变字段（含 password_hash）；调用方需先 GetByID 再改字段。
	Update(ctx context.Context, u *domain.User) error
	// ListByRoles returns users with any of the given roles, ordered by role then username.
	ListByRoles(ctx context.Context, roles []domain.UserRole) ([]*domain.User, error)
}

// StudentListParams 学员分页列表；Keyword 按姓名或学号模糊匹配；ResponsibleInstructorID、FlightTrainingStatus 非空时精确过滤。
type StudentListParams struct {
	Limit                   int
	Offset                  int
	OnlyActiveEmployment    bool
	Keyword                 string
	ResponsibleInstructorID string
	FlightTrainingStatus    string
}

// StudentRepository defines student persistence.
type StudentRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Student, error)
	GetByUserID(ctx context.Context, userID string) (*domain.Student, error)
	GetByStudentNo(ctx context.Context, studentNo string) (*domain.Student, error)
	// List：OnlyActiveEmployment 为 true 时排除已关联账号且任职状态非 active 的学员（无 user_id 的档案仍返回）。
	List(ctx context.Context, p StudentListParams) ([]*domain.Student, error)
	Create(ctx context.Context, s *domain.Student) error
	Update(ctx context.Context, s *domain.Student) error
}

// InstructorRepository defines instructor persistence.
type InstructorRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Instructor, error)
	GetByUserID(ctx context.Context, userID string) (*domain.Instructor, error)
	// List onlyActiveEmployment：true 时仅返回关联用户为在职的教员。
	List(ctx context.Context, onlyActiveEmployment bool) ([]*domain.Instructor, error)
	// ListByExactName 按姓名精确匹配（未删除），用于 CSV 等导入；多条时由调用方判重。
	ListByExactName(ctx context.Context, name string) ([]*domain.Instructor, error)
	Create(ctx context.Context, i *domain.Instructor) error
	Update(ctx context.Context, i *domain.Instructor) error
}

// PersonnelRepository 学员 / 教员 / 地面人员账号的软删除（事务内更新 users 与档案表）。
type PersonnelRepository interface {
	SoftDeleteStudent(ctx context.Context, studentID string) error
	SoftDeleteInstructor(ctx context.Context, instructorID string) error
	SoftDeleteStaffUser(ctx context.Context, userID string) error
}

// AircraftRepository defines aircraft persistence.
type AircraftRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Aircraft, error)
	GetByCode(ctx context.Context, code string) (*domain.Aircraft, error)
	List(ctx context.Context) ([]*domain.Aircraft, error)
	Create(ctx context.Context, a *domain.Aircraft) error
	Update(ctx context.Context, id string, name, status *string) error
	SetAvatarGridFSID(ctx context.Context, id string, gridFSID string) error
	Delete(ctx context.Context, id string) error
	UpdateAirworthiness(ctx context.Context, id string, totalHours, lastInspectionHours *float64, inspectionIntervalHours *int) error
}

// InstructorScheduleRepository defines instructor schedule (duty) persistence.
type InstructorScheduleRepository interface {
	// ListByDate returns all schedules for a given date.
	ListByDate(ctx context.Context, date time.Time) ([]*domain.InstructorSchedule, error)
	// ListByDateRange returns schedules in [from, to] inclusive by date; instructorIDs empty means all.
	ListByDateRange(ctx context.Context, from, to time.Time, instructorIDs []string) ([]*domain.InstructorSchedule, error)
	GetByInstructorAndDate(ctx context.Context, instructorID string, date time.Time) (*domain.InstructorSchedule, error)
	UpsertDay(ctx context.Context, s *domain.InstructorSchedule) error
}

// StaffScheduleRepository defines staff (non-flying) duty persistence by users.id.
type StaffScheduleRepository interface {
	ListByDate(ctx context.Context, date time.Time) ([]*domain.StaffSchedule, error)
	ListByDateRange(ctx context.Context, from, to time.Time, userIDs []string) ([]*domain.StaffSchedule, error)
	GetByUserAndDate(ctx context.Context, userID string, date time.Time) (*domain.StaffSchedule, error)
	UpsertDay(ctx context.Context, s *domain.StaffSchedule) error
}

// AircraftAvailabilityRepository defines aircraft availability persistence.
type AircraftAvailabilityRepository interface {
	// ListByDate returns all availability records for a given date.
	ListByDate(ctx context.Context, date time.Time) ([]*domain.AircraftAvailability, error)
	Create(ctx context.Context, a *domain.AircraftAvailability) error
	DeleteByMaintenanceWindowID(ctx context.Context, maintenanceWindowID string) error
}

// AircraftMaintenanceRepository 维护时间窗持久化
type AircraftMaintenanceRepository interface {
	Create(ctx context.Context, w *domain.MaintenanceWindow) error
	GetByID(ctx context.Context, id string) (*domain.MaintenanceWindow, error)
	Update(ctx context.Context, w *domain.MaintenanceWindow) error
	// ListByAircraft aircraftID 为空时不过滤飞机（返回全部）；from/to/status 仍按原逻辑可选
	ListByAircraft(ctx context.Context, aircraftID string, from, to *time.Time, status *domain.MaintenanceWindowStatus) ([]*domain.MaintenanceWindow, error)
}

// MaintenanceWindowSyncer 维护窗口事务（创建/更新/取消并同步排班）
type MaintenanceWindowSyncer interface {
	CreateMaintenanceWindowWithSync(ctx context.Context, aircraftID, kind string, startAt, endAt time.Time, hangar, notes string, syncSchedule bool, reviewerID string) (*domain.MaintenanceWindowSyncResult, error)
	UpdateMaintenanceWindowWithSync(ctx context.Context, windowID, aircraftID, kind string, startAt, endAt time.Time, hangar, notes string, syncSchedule bool, reviewerID string) (*domain.MaintenanceWindowSyncResult, error)
	CancelMaintenanceWindow(ctx context.Context, windowID string) error
}

// BookingRepository defines booking persistence.
type BookingRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Booking, error)
	ListByStudent(ctx context.Context, studentID string, status *domain.BookingStatus) ([]*domain.Booking, error)
	ListPending(ctx context.Context) ([]*domain.Booking, error)
	Create(ctx context.Context, b *domain.Booking) error
	UpdateStatus(ctx context.Context, id string, status domain.BookingStatus, rejectReason string, reviewerID string) error
	// ListByDate returns all bookings on a specific plan date.
	ListByDate(ctx context.Context, date time.Time) ([]*domain.Booking, error)
	// ListByDateRange returns bookings with plan_date in [from, to] inclusive (calendar dates).
	ListByDateRange(ctx context.Context, from, to time.Time) ([]*domain.Booking, error)
	// ListByAircraftAndDateRange 某飞机在日期区间内的预约（用于维护冲突处理）
	ListByAircraftAndDateRange(ctx context.Context, aircraftID string, from, to time.Time) ([]*domain.Booking, error)
}

// TrainingPlanRepository defines training plan persistence.
type TrainingPlanRepository interface {
	Create(ctx context.Context, p *domain.TrainingPlan) error
	ListByDate(ctx context.Context, from, to string) ([]*domain.TrainingPlan, error)
	DeleteMatchingBookingFields(ctx context.Context, studentID, instructorID, aircraftID string, planDate time.Time, startTime, endTime time.Time) (deletedID *string, err error)
}

// TrainingRecordRepository defines training record persistence.
type TrainingRecordRepository interface {
	GetByID(ctx context.Context, id string) (*domain.TrainingRecord, error)
	FindByTrainingPlanID(ctx context.Context, trainingPlanID string) (*domain.TrainingRecord, error)
	ListByStudent(ctx context.Context, studentID string, limit, offset int) ([]*domain.TrainingRecord, error)
	List(ctx context.Context, filter *domain.TrainingRecordListFilter, sort *domain.TrainingRecordListSort, limit, offset int) ([]*domain.TrainingRecord, int, error)
	CountForExport(ctx context.Context, filter *domain.TrainingRecordListFilter) (int, error)
	ListForCSVExport(ctx context.Context, filter *domain.TrainingRecordListFilter, sort *domain.TrainingRecordListSort, limit int, includePriorExperience bool) ([][]string, error)
	Create(ctx context.Context, r *domain.TrainingRecord) error
	Update(ctx context.Context, r *domain.TrainingRecord) error
	UpdateStatus(ctx context.Context, id string, status domain.TrainingRecordStatus) error
	// AllocateRecordSerialNo 按 training_date 所在自然日递增并返回 TR+YYYYMMDD+三位序号。
	AllocateRecordSerialNo(ctx context.Context, trainingDate time.Time) (string, error)
}

// TrainingArchiveRepository defines training archive persistence.
type TrainingArchiveRepository interface {
	GetByID(ctx context.Context, id string) (*domain.TrainingArchive, error)
	GetByTrainingRecordID(ctx context.Context, recordID string) (*domain.TrainingArchive, error)
	ListByStudent(ctx context.Context, studentID string, limit, offset int) ([]*domain.TrainingArchive, error)
	Create(ctx context.Context, a *domain.TrainingArchive) error
	Update(ctx context.Context, a *domain.TrainingArchive) error
}

// SyllabusOutlineRepository 训练大纲实例
type SyllabusOutlineRepository interface {
	GetByID(ctx context.Context, id string) (*domain.SyllabusOutline, error)
	List(ctx context.Context) ([]*domain.SyllabusOutline, error)
	Create(ctx context.Context, o *domain.SyllabusOutline) error
}

// SyllabusSubjectRepository defines syllabus subject (level-1) persistence.
type SyllabusSubjectRepository interface {
	GetByID(ctx context.Context, id string) (*domain.SyllabusSubject, error)
	List(ctx context.Context, filter *domain.SyllabusSubjectListFilter) ([]*domain.SyllabusSubject, error)
	Create(ctx context.Context, s *domain.SyllabusSubject) error
	Update(ctx context.Context, s *domain.SyllabusSubject) error
	Delete(ctx context.Context, id string) error
}

// SyllabusSubjectSecondaryRepository 二级科目
type SyllabusSubjectSecondaryRepository interface {
	GetByID(ctx context.Context, id string) (*domain.SyllabusSubjectSecondary, error)
	ListBySubjectID(ctx context.Context, syllabusSubjectID string) ([]*domain.SyllabusSubjectSecondary, error)
	ListGroupedBySubjectID(ctx context.Context, syllabusSubjectIDs []string) (map[string][]*domain.SyllabusSubjectSecondary, error)
	Create(ctx context.Context, row *domain.SyllabusSubjectSecondary) error
	Update(ctx context.Context, row *domain.SyllabusSubjectSecondary) error
	Delete(ctx context.Context, id string) error
	Reorder(ctx context.Context, syllabusSubjectID string, orderedIDs []string) error
}

// SyllabusTreeRepository 事务性大纲树写入（导入、创建大纲+科目+二级）
type SyllabusTreeRepository interface {
	ImportOutlineTree(ctx context.Context, outline *domain.SyllabusOutline, subjects []domain.SyllabusImportSubjectRow) (*domain.SyllabusOutline, error)
	CreateSubjectUnderOutline(ctx context.Context, outlineID *string, outline *domain.SyllabusOutline, subject *domain.SyllabusSubject, secondaries []*domain.SyllabusSubjectSecondary) error
}

// RawFlightImportRepository defines raw flight import persistence.
type RawFlightImportRepository interface {
	Create(ctx context.Context, r *domain.RawFlightImport) error
	ListPending(ctx context.Context, limit, offset int) ([]*domain.RawFlightImport, error)
	GetByID(ctx context.Context, id string) (*domain.RawFlightImport, error)
	SetTrainingRecordID(ctx context.Context, id string, recordID string) error
}

// TrainingRecordViewRepository defines training record view persistence.
type TrainingRecordViewRepository interface {
	Create(ctx context.Context, v *domain.TrainingRecordView) error
	ListByUserID(ctx context.Context, userID string) ([]*domain.TrainingRecordView, error)
	GetByID(ctx context.Context, id string) (*domain.TrainingRecordView, error)
	Update(ctx context.Context, v *domain.TrainingRecordView) error
	Delete(ctx context.Context, id string) error
}

// TrainingArchiveSubjectRepository defines training archive subject persistence.
type TrainingArchiveSubjectRepository interface {
	ListByArchiveID(ctx context.Context, archiveID string) ([]*domain.TrainingArchiveSubject, error)
	ReplaceForArchive(ctx context.Context, archiveID string, subjects []*domain.TrainingArchiveSubject) error
}

// NextTrainingSubjectRepository defines next training subject persistence.
type NextTrainingSubjectRepository interface {
	ListByArchiveID(ctx context.Context, archiveID string) ([]*domain.NextTrainingSubject, error)
	ReplaceForArchive(ctx context.Context, archiveID string, assignedBy string, syllabusSubjectIDs []string) error
}

// TaskRepository defines task persistence.
type TaskRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Task, error)
	ListByStudent(ctx context.Context, studentID string, status *domain.TaskStatus) ([]*domain.Task, error)
	Create(ctx context.Context, t *domain.Task) error
	UpdateStatus(ctx context.Context, id string, status domain.TaskStatus) error
	// FindByTypeAndRefID 根据任务类型与业务引用 ID 查找单条任务，用于幂等创建。
	FindByTypeAndRefID(ctx context.Context, taskType domain.TaskType, refID string) (*domain.Task, error)
}

// TheoryDocumentRepository defines theory document persistence.
type TheoryDocumentRepository interface {
	GetByID(ctx context.Context, id string) (*domain.TheoryDocument, error)
	List(ctx context.Context) ([]*domain.TheoryDocument, error)
	Create(ctx context.Context, d *domain.TheoryDocument) error
	Update(ctx context.Context, d *domain.TheoryDocument) error
	Delete(ctx context.Context, id string) error
}

// TheoryAttemptRepository defines theory attempt persistence.
type TheoryAttemptRepository interface {
	Create(ctx context.Context, a *domain.TheoryAttempt) error
	UpdateOnSubmit(ctx context.Context, a *domain.TheoryAttempt) error
	GetByIDAndUser(ctx context.Context, attemptID, userID string) (*domain.TheoryAttempt, error)
	ListByUserAndDocument(ctx context.Context, userID, documentID string, limit, offset int) ([]*domain.TheoryAttempt, error)
}

// TheoryPaperRepository defines theory paper (exam) persistence.
type TheoryPaperRepository interface {
	Create(ctx context.Context, p *domain.TheoryPaper) error
	Update(ctx context.Context, p *domain.TheoryPaper) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.TheoryPaper, error)
	ListByDocumentID(ctx context.Context, documentID string, limit, offset int) ([]*domain.TheoryPaper, int, error)
}

// TheoryPaperQuestionRepository defines theory paper-question link persistence.
type TheoryPaperQuestionRepository interface {
	ReplaceForPaper(ctx context.Context, paperID string, questions []*domain.TheoryPaperQuestion) error
	ListByPaperID(ctx context.Context, paperID string) ([]*domain.TheoryPaperQuestion, error)
}

// SubjectLearningRepository defines subject learning (per syllabus subject) persistence.
type SubjectLearningRepository interface {
	GetByID(ctx context.Context, id string) (*domain.SubjectLearning, error)
	List(ctx context.Context) ([]*domain.SubjectLearning, error)
}

// SubjectAttemptRepository defines subject exam attempt persistence (summary in PostgreSQL).
type SubjectAttemptRepository interface {
	Create(ctx context.Context, a *domain.SubjectAttempt) error
	UpdateOnSubmit(ctx context.Context, a *domain.SubjectAttempt) error
	ListByUserAndSubjectLearning(ctx context.Context, userID, subjectLearningID string, limit, offset int) ([]*domain.SubjectAttempt, error)
}

// ContractRepository defines student contract persistence.
type ContractRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Contract, error)
	ListByStudent(ctx context.Context, studentID string) ([]*domain.Contract, error)
	Create(ctx context.Context, c *domain.Contract) error
	Update(ctx context.Context, c *domain.Contract) error
}

// SmsVerificationRepository defines SMS verification code persistence.
type SmsVerificationRepository interface {
	Create(ctx context.Context, v *domain.SmsVerificationCode) error
	// FindLatestByPhoneScene returns the latest code (used or unused) for rate limiting checks.
	FindLatestByPhoneScene(ctx context.Context, phone string, scene domain.SmsScene) (*domain.SmsVerificationCode, error)
	// FindValidCode looks up a not-yet-used, not-expired code.
	FindValidCode(ctx context.Context, phone string, scene domain.SmsScene, code string, now time.Time) (*domain.SmsVerificationCode, error)
	MarkUsed(ctx context.Context, id string, usedAt time.Time) error
}

// CommunityFollowRepository 社区关注
type CommunityFollowRepository interface {
	Follow(ctx context.Context, followerID, followeeID string) error
	Unfollow(ctx context.Context, followerID, followeeID string) error
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)
	ListFolloweeIDs(ctx context.Context, followerID string) ([]string, error)
	ListFollowers(ctx context.Context, followeeID string, limit, offset int) ([]string, error)
	ListFollowing(ctx context.Context, followerID string, limit, offset int) ([]string, error)
	CountFollowers(ctx context.Context, followeeID string) (int, error)
	CountFollowing(ctx context.Context, followerID string) (int, error)
	// RecommendUserIDs 推荐关注：排除自身与已关注；按同组、粉丝数、教员角色排序
	RecommendUserIDs(ctx context.Context, viewerID string, limit int) ([]string, error)
}

// CommunityTopicRepository 社区话题
type CommunityTopicRepository interface {
	Create(ctx context.Context, t *domain.CommunityTopic) error
	GetByID(ctx context.Context, id string) (*domain.CommunityTopic, error)
	GetBySlug(ctx context.Context, slug string) (*domain.CommunityTopic, error)
	ListByIDs(ctx context.Context, ids []string) ([]*domain.CommunityTopic, error)
	ListHot(ctx context.Context, limit, offset int) ([]*domain.CommunityTopic, error)
}

// CommunityGroupRepository 兴趣小组
type CommunityGroupRepository interface {
	Create(ctx context.Context, g *domain.CommunityGroup) error
	Update(ctx context.Context, g *domain.CommunityGroup) error
	GetByID(ctx context.Context, id string) (*domain.CommunityGroup, error)
	GetBySlug(ctx context.Context, slug string) (*domain.CommunityGroup, error)
	List(ctx context.Context, limit, offset int) ([]*domain.CommunityGroup, error)
	AddMember(ctx context.Context, groupID, userID string, role domain.CommunityGroupMemberRole) error
	RemoveMember(ctx context.Context, groupID, userID string) error
	GetMembership(ctx context.Context, groupID, userID string) (*domain.CommunityGroupMember, error)
	IsMember(ctx context.Context, groupID, userID string) (bool, error)
	ListMembers(ctx context.Context, groupID string, limit, offset int) ([]*domain.CommunityGroupMember, error)
}

// CommunityPostRepository 社区动态
type CommunityPostRepository interface {
	Create(ctx context.Context, p *domain.CommunityPost) error
	// CreateWithTopics 在同一事务中插入动态并写入话题关联
	CreateWithTopics(ctx context.Context, p *domain.CommunityPost, topicIDs []string) error
	Update(ctx context.Context, p *domain.CommunityPost) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.CommunityPost, error)
	ReplaceTopicLinks(ctx context.Context, postID string, topicIDs []string) error
	ListTopicIDs(ctx context.Context, postID string) ([]string, error)
	ListFeedGlobal(ctx context.Context, limit int, cursorCreatedAt *time.Time, cursorID *string) ([]*domain.CommunityPost, error)
	ListFeedFollowing(ctx context.Context, viewerID string, limit int, cursorCreatedAt *time.Time, cursorID *string) ([]*domain.CommunityPost, error)
	ListFeedGroup(ctx context.Context, groupID, viewerID string, limit int, cursorCreatedAt *time.Time, cursorID *string) ([]*domain.CommunityPost, error)
	ListByTopic(ctx context.Context, topicID string, limit int, cursorCreatedAt *time.Time, cursorID *string) ([]*domain.CommunityPost, error)
	ListByAuthor(ctx context.Context, authorID string, limit int, cursorCreatedAt *time.Time, cursorID *string) ([]*domain.CommunityPost, error)
	BatchLikedByUser(ctx context.Context, userID string, postIDs []string) (map[string]bool, error)
}

// CommunityLikeRepository 点赞
type CommunityLikeRepository interface {
	Like(ctx context.Context, postID, userID string) error
	Unlike(ctx context.Context, postID, userID string) error
}

// CommunityCommentRepository 评论
type CommunityCommentRepository interface {
	Create(ctx context.Context, c *domain.CommunityComment) error
	ListByPost(ctx context.Context, postID string, limit, offset int) ([]*domain.CommunityComment, error)
}

// CommunityChallengeRepository 挑战赛
type CommunityChallengeRepository interface {
	Create(ctx context.Context, ch *domain.CommunityChallenge) error
	Update(ctx context.Context, ch *domain.CommunityChallenge) error
	GetByID(ctx context.Context, id string) (*domain.CommunityChallenge, error)
	List(ctx context.Context, status *domain.CommunityChallengeStatus, limit, offset int) ([]*domain.CommunityChallenge, error)
	SubmitEntry(ctx context.Context, e *domain.CommunityChallengeEntry) error
	GetEntry(ctx context.Context, challengeID, userID string) (*domain.CommunityChallengeEntry, error)
	ListEntries(ctx context.Context, challengeID string, limit, offset int) ([]*domain.CommunityChallengeEntry, error)
}
