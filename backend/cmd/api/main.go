package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/securedlinq/backend/internal/config"
	"github.com/securedlinq/backend/internal/database"
	"github.com/securedlinq/backend/internal/handler"
	"github.com/securedlinq/backend/internal/middleware"
	"github.com/securedlinq/backend/internal/repository"
	"github.com/securedlinq/backend/internal/service"
	"github.com/securedlinq/backend/pkg/agora"
	"github.com/securedlinq/backend/pkg/s3"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Connect to database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	sessionRepo := repository.NewSessionRepository(db)
	meetingRepo := repository.NewMeetingRepository(db)
	driverRepo := repository.NewDriverRepository(db)
	loadRepo := repository.NewLoadRepository(db)
	galleryRepo := repository.NewGalleryRepository(db)
	// Initialize external clients
	agoraClient := agora.NewClient(cfg.Agora.AppID, cfg.Agora.AppCertificate, cfg.Agora.EncodedKey)
	s3Client, err := s3.NewClient(cfg.AWS.AccessKeyID, cfg.AWS.SecretAccessKey, cfg.AWS.Region, cfg.AWS.S3BucketName)
	if err != nil {
		log.Printf("Warning: Failed to initialize S3 client: %v", err)
	}

	// Initialize services
	authService := service.NewAuthService(sessionRepo, driverRepo, cfg)
	meetingService := service.NewMeetingService(meetingRepo, loadRepo, cfg)
	recordingService := service.NewRecordingService(meetingRepo, galleryRepo, agoraClient)
	driverService := service.NewDriverService(driverRepo)
	loadService := service.NewLoadService(loadRepo, driverRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, driverService, cfg)
	meetingHandler := handler.NewMeetingHandler(meetingService)
	agoraHandler := handler.NewAgoraHandler(agoraClient, recordingService)
	emailHandler := handler.NewEmailHandler(&cfg.Email)
	mediaHandler := handler.NewMediaHandler(s3Client, galleryRepo, meetingRepo, cfg)
	driverHandler := handler.NewDriverHandler(driverService)
	loadHandler := handler.NewLoadHandler(loadService)

	// Setup Gin router
	r := gin.Default()

	// Add middleware
	r.Use(middleware.CORSMiddleware(cfg.Server.FrontendURL))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Debug route to list all routes (only in debug mode)
	if cfg.Server.GinMode == "debug" {
		r.GET("/debug/routes", func(c *gin.Context) {
			routes := []string{}
			for _, route := range r.Routes() {
				routes = append(routes, route.Method+" "+route.Path)
			}
			c.JSON(200, gin.H{"routes": routes})
		})
	}

	// API routes
	api := r.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login) // Admin login
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/validate", authHandler.ValidateSession)

			// Driver auth routes
			auth.POST("/driver/register", authHandler.DriverRegister)
			auth.POST("/driver/login", authHandler.DriverLogin)
		}

		// Agora routes (public - for video calls)
		agoraRoutes := api.Group("/agora")
		{
			agoraRoutes.POST("/token", agoraHandler.GenerateToken)

			recording := agoraRoutes.Group("/recording")
			{
				recording.POST("/start", agoraHandler.StartRecording)
				recording.POST("/stop", agoraHandler.StopRecording)
				recording.GET("/query", agoraHandler.QueryRecording)
			}
		}

		// Meeting routes (for authenticated users - admins and drivers)
		meetings := api.Group("/meetings")
		meetings.Use(middleware.AuthMiddleware(authService))
		{
			meetings.POST("", meetingHandler.CreateMeeting)
			meetings.GET("", meetingHandler.GetMeetingByRoomID)
			meetings.DELETE("", meetingHandler.EndMeeting)
		}

		// Protected routes (require auth)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// Email routes (admin only)
			email := protected.Group("/email")
			email.Use(middleware.AdminOnlyMiddleware())
			{
				email.POST("/send-meeting-link", emailHandler.SendMeetingLink)
			}

			// Media routes (admin only)
			media := protected.Group("/media")
			media.Use(middleware.AdminOnlyMiddleware())
			{
				media.GET("", mediaHandler.GetLoadMedia)
				media.POST("/screenshot", mediaHandler.SaveScreenshot)
				media.GET("/screenshots", mediaHandler.GetScreenshotsByLoad)
				media.GET("/signed-url", mediaHandler.GetSignedURL)
			}
		}

		// Admin-only routes
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(authService))
		admin.Use(middleware.AdminOnlyMiddleware())
		{
			// Drivers management
			drivers := admin.Group("/drivers")
			{
				drivers.GET("", driverHandler.GetAllDrivers)
				drivers.GET("/:id", driverHandler.GetDriverByID)
				drivers.POST("/:id/deactivate", driverHandler.DeactivateDriver)
				drivers.POST("/:id/activate", driverHandler.ActivateDriver)
			}

			// Loads management
			loads := admin.Group("/loads")
			{
				loads.POST("", loadHandler.CreateLoad)
				loads.GET("", loadHandler.GetAllLoads)
				loads.GET("/by-status", loadHandler.GetLoadsByStatus)
				loads.GET("/:id", loadHandler.GetLoadByID)
				loads.POST("/:id/assign", loadHandler.AssignDriver)
				loads.POST("/:id/start-meeting", loadHandler.StartMeeting)
				loads.DELETE("/:id", loadHandler.DeleteLoad)
			}
		}

		// Driver-only routes
		driver := api.Group("/driver")
		driver.Use(middleware.AuthMiddleware(authService))
		driver.Use(middleware.DriverOnlyMiddleware())
		{
			// Driver's loads
			driver.GET("/loads", loadHandler.GetDriverLoads)
			driver.GET("/loads/:id", loadHandler.GetLoadByID)
			driver.POST("/loads/:id/complete", loadHandler.MarkCompleted)
			driver.PUT("/loads/:id/status", loadHandler.UpdateLoadStatus)
		}
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")
	}()

	// Start server
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
