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
	}

	return r
}
