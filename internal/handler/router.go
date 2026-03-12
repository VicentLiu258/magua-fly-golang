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
	student *StudentHandler,
	instructor *InstructorHandler,
	aircraft *AircraftHandler,
	booking *BookingHandler,
	jwtManager *jwt.Manager,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.GinCORS())

	r.GET("/health", health.Check)
	r.POST("/auth/login", auth.Login)
	r.POST("/auth/register", auth.Register)

	authGroup := r.Group("")
	authGroup.Use(middleware.GinAuth(jwtManager))
	{
		authGroup.GET("/students", student.List)
		authGroup.GET("/instructors", instructor.List)
		authGroup.GET("/aircraft", aircraft.List)
		authGroup.GET("/bookings/pending", booking.ListPending)
		authGroup.POST("/bookings", booking.Create)
		authGroup.POST("/bookings/:id/review", booking.Review)
	}

	return r
}
