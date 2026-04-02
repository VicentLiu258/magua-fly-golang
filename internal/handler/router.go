package handler

import (
	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/pkg/jwt"
	"magua-fly-golang/internal/repository"

	"github.com/gin-gonic/gin"
)

// Router builds the HTTP router.
func Router(
	health *HealthHandler,
	auth *AuthHandler,
	smsHandler *SmsHandler,
	student *StudentHandler,
	instructor *InstructorHandler,
	aircraft *AircraftHandler,
	booking *BookingHandler,
	schedule *ScheduleHandler,
	training *TrainingHandler,
	theory *TheoryHandler,
	subjectExam *SubjectExamHandler,
	task *TaskHandler,
	contractFiles *ContractFilesHandler,
	community *CommunityHandler,
	home *HomeHandler,
	personnel *PersonnelHandler,
	jwtManager *jwt.Manager,
	users repository.UserRepository,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.GinCORS())

	r.GET("/health", health.Check)
	r.POST("/auth/login", auth.Login)
	r.POST("/auth/register", auth.Register)
	r.POST("/auth/sms/send", smsHandler.SendCode)
	r.POST("/auth/sms/verify", smsHandler.VerifyCode)

	authGroup := r.Group("")
	authGroup.Use(middleware.GinAuth(jwtManager, users))
	{
		authGroup.POST("/api/v1/students/list", student.List)
		authGroup.POST("/api/v1/students/staff-overview", student.StaffOverview)
		authGroup.POST("/api/v1/students/my-home", student.MyHome)
		authGroup.POST("/api/v1/instructors/list", instructor.List)
		authGroup.POST("/api/v1/instructors/my-home", instructor.MyHome)
		authGroup.POST("/api/v1/aircraft/list", aircraft.List)
		authGroup.POST("/api/v1/aircraft/create", aircraft.Create)
		authGroup.POST("/api/v1/aircraft/update", aircraft.Update)
		authGroup.POST("/api/v1/aircraft/delete", aircraft.Delete)
		authGroup.POST("/api/v1/aircraft/avatar/upload", aircraft.UploadAvatar)
		authGroup.POST("/api/v1/aircraft/avatar/remove", aircraft.RemoveAvatar)
		authGroup.GET("/api/v1/aircraft/avatar", aircraft.StreamAvatar)
		authGroup.POST("/api/v1/aircraft/availabilities/create", aircraft.CreateAvailability)
		authGroup.POST("/api/v1/aircraft/airworthiness/get", aircraft.AirworthinessGet)
		authGroup.POST("/api/v1/aircraft/airworthiness/update", aircraft.AirworthinessUpdate)
		authGroup.POST("/api/v1/aircraft/maintenance-windows/list", aircraft.MaintenanceWindowsList)
		authGroup.POST("/api/v1/aircraft/maintenance-windows/create", aircraft.MaintenanceWindowsCreate)
		authGroup.POST("/api/v1/aircraft/maintenance-windows/update", aircraft.MaintenanceWindowsUpdate)
		authGroup.POST("/api/v1/aircraft/maintenance-windows/cancel", aircraft.MaintenanceWindowsCancel)
		authGroup.POST("/api/v1/bookings", booking.Create)
		authGroup.POST("/api/v1/bookings/pending", booking.ListPending)
		authGroup.POST("/api/v1/bookings/review", booking.Review)
		authGroup.POST("/api/v1/schedules/instructor-grid", schedule.InstructorGrid)
		authGroup.POST("/api/v1/schedules/workforce-grid", schedule.WorkforceGrid)
		authGroup.POST("/api/v1/schedules/staff-duties/list", schedule.StaffDutiesList)
		authGroup.POST("/api/v1/schedules/staff-duties/upsert", schedule.StaffDutiesUpsert)
		authGroup.POST("/api/v1/schedules/staff-duties/bulk", schedule.StaffDutiesBulk)
		authGroup.POST("/api/v1/schedules/instructor-duties/list", schedule.InstructorDutiesList)
		authGroup.POST("/api/v1/schedules/instructor-duties/my-list", schedule.InstructorDutiesMyList)
		authGroup.POST("/api/v1/schedules/instructor-duties/upsert", schedule.InstructorDutiesUpsert)
		authGroup.POST("/api/v1/schedules/instructor-duties/bulk", schedule.InstructorDutiesBulk)

		// 训练数据录入与管理（均为 POST）
		authGroup.POST("/api/v1/config/training-record-fields", training.ConfigTrainingRecordFields)
		authGroup.POST("/api/v1/training-records", training.CreateRecord)
		authGroup.POST("/api/v1/training-records/detail", training.RecordDetail)
		authGroup.POST("/api/v1/training-records/update", training.UpdateRecord)
		authGroup.POST("/api/v1/training-records/delete", training.DeleteRecord)
		authGroup.POST("/api/v1/training-records/list", training.ListRecords)
		authGroup.POST("/api/v1/training-records/export", training.ExportRecordsCSV)
		authGroup.POST("/api/v1/training-records/import", training.ImportRecords)
		authGroup.POST("/api/v1/training-records/claim", training.ClaimImport)
		authGroup.POST("/api/v1/training-records/import/pending", training.ListPendingImports)
		authGroup.POST("/api/v1/training-records/stats", training.RecordStats)
		authGroup.POST("/api/v1/dashboard/summary", training.DashboardSummary)
		authGroup.POST("/api/v1/training-views", training.CreateView)
		authGroup.POST("/api/v1/training-views/list", training.ListViews)
		authGroup.POST("/api/v1/training-views/update", training.UpdateView)
		authGroup.POST("/api/v1/training-views/delete", training.DeleteView)
		authGroup.POST("/api/v1/students/training-archives/list", training.ListArchivesByStudent)
		authGroup.POST("/api/v1/training-archives/detail", training.ArchiveDetail)
		authGroup.POST("/api/v1/training-archives", training.CreateArchive)
		authGroup.POST("/api/v1/training-archives/update", training.UpdateArchive)
		authGroup.POST("/api/v1/training-archives/sign", training.SignArchive)
		authGroup.POST("/api/v1/syllabus-subjects/list", training.ListSyllabusSubjects)
		authGroup.POST("/api/v1/syllabus-outlines/list", training.ListSyllabusOutlines)
		authGroup.POST("/api/v1/training-archives/subjects/list", training.ListArchiveSubjects)
		authGroup.POST("/api/v1/training-archives/subjects/update", training.UpdateArchiveSubjects)
		authGroup.POST("/api/v1/training-archives/next-subjects/list", training.ListNextSubjects)
		authGroup.POST("/api/v1/training-archives/next-subjects/update", training.UpdateNextSubjects)

		// 理论学习模块
		authGroup.POST("/api/v1/theory/documents/create", theory.CreateDocument)
		authGroup.POST("/api/v1/theory/documents/update", theory.UpdateDocument)
		authGroup.POST("/api/v1/theory/documents/delete", theory.DeleteDocument)
		authGroup.POST("/api/v1/theory/documents/list", theory.ListDocuments)
		authGroup.POST("/api/v1/theory/documents/detail", theory.DocumentDetail)
		authGroup.POST("/api/v1/theory/documents/video/upload", theory.UploadTheoryVideo)
		authGroup.POST("/api/v1/theory/documents/video/remove", theory.RemoveTheoryVideo)
		authGroup.GET("/api/v1/theory/documents/video", theory.StreamTheoryVideo)
		authGroup.POST("/api/v1/theory/questions/start", theory.StartQuestions)
		authGroup.POST("/api/v1/theory/questions/submit", theory.SubmitAnswers)
		authGroup.POST("/api/v1/theory/attempts/my-list", theory.ListMyAttempts)

		// 理论试卷管理
		authGroup.POST("/api/v1/theory/papers/list", theory.ListPapers)
		authGroup.POST("/api/v1/theory/papers/save", theory.SavePaper)
		authGroup.POST("/api/v1/theory/papers/delete", theory.DeletePaper)

		// 训练科目学习与考试（Mongo 题库 + PG 摘要，与理论同构）
		authGroup.POST("/api/v1/subject-learning/questions/start", subjectExam.StartQuestions)
		authGroup.POST("/api/v1/subject-learning/questions/submit", subjectExam.SubmitAnswers)
		authGroup.POST("/api/v1/subject-learning/attempts/my-list", subjectExam.ListMyAttempts)

		// 学员任务模块
		authGroup.POST("/api/v1/tasks/my-list", task.MyList)
		authGroup.POST("/api/v1/tasks/update-status", task.UpdateStatus)

		// 合同文件
		authGroup.POST("/api/v1/contracts/files/upload", contractFiles.Upload)
		authGroup.GET("/api/v1/contracts/files/list", contractFiles.List)
		authGroup.GET("/api/v1/contracts/files/download", contractFiles.Download)

		// 社区
		authGroup.POST("/api/v1/community/posts/create", community.PostsCreate)
		authGroup.POST("/api/v1/community/posts/update", community.PostsUpdate)
		authGroup.POST("/api/v1/community/posts/delete", community.PostsDelete)
		authGroup.POST("/api/v1/community/posts/detail", community.PostsDetail)
		authGroup.POST("/api/v1/community/posts/list", community.PostsList)
		authGroup.POST("/api/v1/community/posts/like", community.PostsLike)
		authGroup.POST("/api/v1/community/posts/unlike", community.PostsUnlike)
		authGroup.POST("/api/v1/community/comments/create", community.CommentsCreate)
		authGroup.POST("/api/v1/community/comments/list", community.CommentsList)
		authGroup.POST("/api/v1/community/follows/follow", community.FollowsFollow)
		authGroup.POST("/api/v1/community/follows/unfollow", community.FollowsUnfollow)
		authGroup.POST("/api/v1/community/follows/followers", community.FollowsFollowers)
		authGroup.POST("/api/v1/community/follows/following", community.FollowsFollowing)
		authGroup.POST("/api/v1/community/follows/recommend", community.FollowsRecommend)
		authGroup.POST("/api/v1/community/topics/list", community.TopicsList)
		authGroup.POST("/api/v1/community/topics/detail", community.TopicsDetail)
		authGroup.POST("/api/v1/community/topics/create", community.TopicsCreate)
		authGroup.POST("/api/v1/community/groups/create", community.GroupsCreate)
		authGroup.POST("/api/v1/community/groups/update", community.GroupsUpdate)
		authGroup.POST("/api/v1/community/groups/detail", community.GroupsDetail)
		authGroup.POST("/api/v1/community/groups/list", community.GroupsList)
		authGroup.POST("/api/v1/community/groups/join", community.GroupsJoin)
		authGroup.POST("/api/v1/community/groups/leave", community.GroupsLeave)
		authGroup.POST("/api/v1/community/groups/members", community.GroupsMembers)
		authGroup.POST("/api/v1/community/challenges/create", community.ChallengesCreate)
		authGroup.POST("/api/v1/community/challenges/update", community.ChallengesUpdate)
		authGroup.POST("/api/v1/community/challenges/list", community.ChallengesList)
		authGroup.POST("/api/v1/community/challenges/detail", community.ChallengesDetail)
		authGroup.POST("/api/v1/community/challenges/submit", community.ChallengesSubmit)
		authGroup.POST("/api/v1/community/challenges/entries", community.ChallengesEntries)

		adminOnly := authGroup.Group("")
		adminOnly.Use(middleware.GinRequireAdmin())
		{
			adminOnly.POST("/api/v1/admin/students/list", student.AdminList)
			adminOnly.POST("/api/v1/admin/students/detail", student.AdminDetail)
			adminOnly.POST("/api/v1/admin/students/save", student.AdminSave)
			adminOnly.POST("/api/v1/admin/contracts/list-by-student", student.AdminContractsListByStudent)
			adminOnly.POST("/api/v1/admin/home", home.AdminHome)
			adminOnly.POST("/api/v1/admin/personnel/delete", personnel.AdminDelete)
			adminOnly.POST("/api/v1/admin/personnel/save", personnel.AdminSave)
			adminOnly.POST("/api/v1/syllabus-subjects/admin/create", training.AdminCreateSyllabusSubject)
			adminOnly.POST("/api/v1/syllabus-subjects/admin/update", training.AdminUpdateSyllabusSubject)
			adminOnly.POST("/api/v1/syllabus-subjects/admin/delete", training.AdminDeleteSyllabusSubject)
			adminOnly.POST("/api/v1/syllabus-outlines/admin/import", training.AdminImportSyllabusOutline)
			adminOnly.POST("/api/v1/syllabus-subject-secondaries/admin/create", training.AdminCreateSyllabusSubjectSecondary)
			adminOnly.POST("/api/v1/syllabus-subject-secondaries/admin/update", training.AdminUpdateSyllabusSubjectSecondary)
			adminOnly.POST("/api/v1/syllabus-subject-secondaries/admin/delete", training.AdminDeleteSyllabusSubjectSecondary)
			adminOnly.POST("/api/v1/syllabus-subject-secondaries/admin/reorder", training.AdminReorderSyllabusSubjectSecondaries)
			adminOnly.POST("/api/v1/theory/questions/admin/list", theory.AdminListQuestions)
			adminOnly.POST("/api/v1/theory/questions/admin/create", theory.AdminCreateQuestion)
			adminOnly.POST("/api/v1/theory/questions/admin/update", theory.AdminUpdateQuestion)
			adminOnly.POST("/api/v1/theory/questions/admin/delete", theory.AdminDeleteQuestion)
			adminOnly.POST("/api/v1/subject-learning/questions/admin/list", subjectExam.AdminListQuestions)
			adminOnly.POST("/api/v1/subject-learning/questions/admin/create", subjectExam.AdminCreateQuestion)
			adminOnly.POST("/api/v1/subject-learning/questions/admin/update", subjectExam.AdminUpdateQuestion)
			adminOnly.POST("/api/v1/subject-learning/questions/admin/delete", subjectExam.AdminDeleteQuestion)
		}
	}

	return r
}
