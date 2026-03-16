package handler

import (
	"magua-fly-golang/internal/middleware"
	"magua-fly-golang/internal/pkg/jwt"

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
	task *TaskHandler,
	jwtManager *jwt.Manager,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.GinCORS())

	r.GET("/health", health.Check)
	r.POST("/auth/login", auth.Login)
	r.POST("/auth/register", auth.Register)
	r.POST("/auth/sms/send", smsHandler.SendCode)
	r.POST("/auth/sms/verify", smsHandler.VerifyCode)

	authGroup := r.Group("")
	authGroup.Use(middleware.GinAuth(jwtManager))
	{
		authGroup.POST("/api/v1/students/list", student.List)
		authGroup.POST("/api/v1/instructors/list", instructor.List)
		authGroup.POST("/api/v1/aircraft/list", aircraft.List)
		authGroup.POST("/api/v1/aircraft/create", aircraft.Create)
		authGroup.POST("/api/v1/aircraft/availabilities/create", aircraft.CreateAvailability)
		authGroup.POST("/api/v1/bookings", booking.Create)
		authGroup.POST("/api/v1/bookings/pending", booking.ListPending)
		authGroup.POST("/api/v1/bookings/review", booking.Review)
		authGroup.POST("/api/v1/schedules/instructor-grid", schedule.InstructorGrid)

		// 训练数据录入与管理（均为 POST）
		authGroup.POST("/api/v1/config/training-record-fields", training.ConfigTrainingRecordFields)
		authGroup.POST("/api/v1/training-records", training.CreateRecord)
		authGroup.POST("/api/v1/training-records/detail", training.RecordDetail)
		authGroup.POST("/api/v1/training-records/update", training.UpdateRecord)
		authGroup.POST("/api/v1/training-records/delete", training.DeleteRecord)
		authGroup.POST("/api/v1/training-records/list", training.ListRecords)
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
		authGroup.POST("/api/v1/syllabus-subjects/admin/create", training.AdminCreateSyllabusSubject)
		authGroup.POST("/api/v1/syllabus-subjects/admin/update", training.AdminUpdateSyllabusSubject)
		authGroup.POST("/api/v1/syllabus-subjects/admin/delete", training.AdminDeleteSyllabusSubject)
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
			authGroup.POST("/api/v1/theory/questions/start", theory.StartQuestions)
			authGroup.POST("/api/v1/theory/questions/submit", theory.SubmitAnswers)
			authGroup.POST("/api/v1/theory/attempts/my-list", theory.ListMyAttempts)

			// 理论题库管理（教员/管理员使用）
			authGroup.POST("/api/v1/theory/questions/admin/list", theory.AdminListQuestions)
			authGroup.POST("/api/v1/theory/questions/admin/create", theory.AdminCreateQuestion)
			authGroup.POST("/api/v1/theory/questions/admin/update", theory.AdminUpdateQuestion)
			authGroup.POST("/api/v1/theory/questions/admin/delete", theory.AdminDeleteQuestion)

			// 理论试卷管理
			authGroup.POST("/api/v1/theory/papers/list", theory.ListPapers)
			authGroup.POST("/api/v1/theory/papers/save", theory.SavePaper)
			authGroup.POST("/api/v1/theory/papers/delete", theory.DeletePaper)

			// 学员任务模块
			authGroup.POST("/api/v1/tasks/my-list", task.MyList)
			authGroup.POST("/api/v1/tasks/update-status", task.UpdateStatus)
	}

	return r
}
