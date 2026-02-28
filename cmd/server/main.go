package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"leadflowcrm/internal/config"
	"leadflowcrm/internal/handlers"
	"leadflowcrm/internal/middlewares"
	"leadflowcrm/internal/repositories"
	"leadflowcrm/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	wd, _ := os.Getwd()

	if err := godotenv.Load(filepath.Join(wd, ".env")); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	if _, err := config.ConnectDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	repositories.NewLeadRepository().CreateIndexes(ctx)
	repositories.NewUserRepository().CreateIndexes(ctx)

	authService := services.NewAuthService()
	leadService := services.NewLeadService()

	authHandler := handlers.NewAuthHandler(authService, cfg)
	leadHandler := handlers.NewLeadHandler(leadService)

	authMiddleware := middlewares.NewAuthMiddleware(cfg)
	rateLimiter := middlewares.NewRateLimiter(cfg.RateLimitReq, cfg.RateLimitWindow)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middlewares.CORSMiddleware())
	r.Use(middlewares.ErrorHandler())

	r.LoadHTMLGlob(filepath.Join(wd, "web/templates/*"))

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "LeadFlowCRM - Gestión de Leads",
		})
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login - LeadFlowCRM",
		})
	})

	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"title": "Registro - LeadFlowCRM",
		})
	})

	api := r.Group("/api")
	api.Use(middlewares.RateLimit(rateLimiter))
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		protected := api.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			users := protected.Group("/users")
			{
				users.GET("/profile", authHandler.GetProfile)
				users.GET("", authHandler.GetUsers)
			}

			leads := protected.Group("/leads")
			{
				leads.POST("", leadHandler.CreateLead)
				leads.GET("", leadHandler.GetLeads)
				leads.GET("/stats", leadHandler.GetStats)
				leads.GET("/:id", leadHandler.GetLead)
				leads.PUT("/:id", leadHandler.UpdateLead)
				leads.DELETE("/:id", leadHandler.DeleteLead)
				leads.GET("/:id/history", leadHandler.GetLeadHistory)
				leads.POST("/:id/notes", leadHandler.AddNote)
				leads.GET("/:id/notes", leadHandler.GetLeadNotes)
				leads.DELETE("/:id/notes/:noteId", leadHandler.DeleteNote)
				leads.POST("/import", leadHandler.ImportCSV)
			}
		}
	}

	web := r.Group("/dashboard")
	web.Use(authMiddleware.RequireAuth())
	{
		web.GET("", func(c *gin.Context) {
			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"title": "Dashboard - LeadFlowCRM",
			})
		})

		web.GET("/leads", func(c *gin.Context) {
			c.HTML(http.StatusOK, "leads.html", gin.H{
				"title": "Leads - LeadFlowCRM",
			})
		})

		web.GET("/leads/:id", func(c *gin.Context) {
			c.HTML(http.StatusOK, "lead-detail.html", gin.H{
				"title": "Detalle de Lead - LeadFlowCRM",
			})
		})

		web.GET("/leads/new", func(c *gin.Context) {
			c.HTML(http.StatusOK, "lead-form.html", gin.H{
				"title": "Nuevo Lead - LeadFlowCRM",
			})
		})

		web.GET("/leads/:id/edit", func(c *gin.Context) {
			c.HTML(http.StatusOK, "lead-form.html", gin.H{
				"title": "Editar Lead - LeadFlowCRM",
			})
		})

		web.GET("/users", func(c *gin.Context) {
			c.HTML(http.StatusOK, "users.html", gin.H{
				"title": "Usuarios - LeadFlowCRM",
			})
		})
	}

	r.Static("/static", filepath.Join(wd, "web/static"))

	port := cfg.Port
	log.Printf("Server starting on port %s", port)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
