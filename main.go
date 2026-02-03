package main

import (
	"log/slog"

	"bitacora-medica-backend/api/config"
	"bitacora-medica-backend/api/database"
	"bitacora-medica-backend/api/handlers/admin"
	"bitacora-medica-backend/api/handlers/auth"
	"bitacora-medica-backend/api/handlers/collaborations"
	"bitacora-medica-backend/api/handlers/common"
	"bitacora-medica-backend/api/handlers/patients"
	"bitacora-medica-backend/api/handlers/professional"
	"bitacora-medica-backend/api/handlers/reports"
	"bitacora-medica-backend/api/handlers/sessions"
	"bitacora-medica-backend/api/handlers/support"
	"time"

	"github.com/gin-contrib/cors"

	"bitacora-medica-backend/api/middleware"
	_ "bitacora-medica-backend/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Bitacora Médica API
// @version         1.0
// @description     Backend API for Bitacora Médica application.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {

	cfg := config.LoadConfig()

	database.Connect(cfg.DBUrl)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.MaxMultipartMemory = 8 << 20

	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	api := r.Group("/api")

	api.Use(middleware.RateLimitMiddleware())
	api.Use(middleware.AuthMiddleware(cfg))
	{
		authGroup := api.Group("/auth")
		{
			authGroup.PUT("/profile", auth.UpdateProfileHandler())
			authGroup.GET("/me", auth.GetMeHandler())
		}

		// --- GRUPO DE PACIENTES ---
		patientsGroup := api.Group("/patients")
		{
			patientsGroup.POST("/", patients.CreatePatientHandler(cfg))

			patientsGroup.GET("/", patients.ListPatientsHandler())

			patientsGroup.GET("/:id", patients.GetPatientProfileHandler(cfg))

			patientsGroup.PUT("/:id", patients.UpdatePatientHandler())

			patientsGroup.GET("/:id/ai-context", patients.GetPatientAIContextHandler())

			patientsGroup.POST("/:id/documents", patients.UploadDocumentHandler(cfg))

			patientsGroup.GET("/:id/documents", patients.ListDocumentsHandler(cfg))

			patientsGroup.DELETE("/:id/documents/:doc_id", patients.DeleteDocumentHandler())
		}

		// --- GRUPO DE SESIONES ---
		sessionsGroup := api.Group("/sessions")
		{

			sessionsGroup.POST("/", sessions.CreateSessionHandler(cfg))

			sessionsGroup.GET("/", sessions.ListSessionsHandler(cfg))

			sessionsGroup.GET("/:id", sessions.GetSessionHandler(cfg))

			sessionsGroup.PUT("/:id", sessions.UpdateSessionHandler())

			sessionsGroup.DELETE("/:id", sessions.DeleteSessionHandler())
		}

		// --- GRUPO DE SUBIDAS ---
		uploads := api.Group("/uploads")

		uploads.POST("/image", common.UploadImageHandler(cfg))

		uploads.POST("/consent", common.UploadConsentHandler(cfg))
	}

	// --- GRUPO DE COLABORACIONES ---
	collabGroup := api.Group("/collaborations")
	{

		collabGroup.POST("/invite", collaborations.InviteCollabHandler(cfg))

		collabGroup.PUT("/:id/respond", collaborations.RespondInvitationHandler(cfg))

		collabGroup.GET("/pending", collaborations.GetPendingInvitationsHandler())

		collabGroup.DELETE("/:id", collaborations.UnlinkProfessionalHandler())
	}

	// --- GRUPO REPORTES ---
	reportsGroup := api.Group("/reports")
	{

		reportsGroup.POST("", reports.CreateIndividualReportHandler())

		reportsGroup.GET("/list", reports.ListPatientReportsHandler())

		reportsGroup.GET("/master", reports.GenerateMasterReportHandler())
	}

	// --- GRUPO SOPORTE ---
	supportGroup := api.Group("/support")
	{
		supportGroup.POST("/", support.CreateTicketHandler(cfg))

		supportGroup.GET("/", support.ListTicketsHandler())

		supportGroup.PUT("/:id/reply", middleware.RequireAdmin(), support.ReplyTicketHandler(cfg))
	}

	// --- GRUPO DASHBOARD ---
	dashboardGroup := api.Group("/dashboard")
	{
		dashboardGroup.GET("/summary", professional.GetMyDashboardHandler())
	}

	// --- GRUPO ADMIN  ---
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.RequireAdmin())
	{
		adminGroup.GET("/users/pending", admin.ListPendingUsersHandler())

		adminGroup.PUT("/users/:id/review", admin.ReviewUserHandler(cfg))

		adminGroup.GET("/dashboard", admin.GetDashboardStatsHandler())
	}

	slog.Info("Server starting on port " + cfg.Port)
	r.Run(":" + cfg.Port)
}
