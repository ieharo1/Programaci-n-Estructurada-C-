package handlers

import (
	"net/http"
	"strconv"

	"leadflowcrm/internal/config"
	"leadflowcrm/internal/models"
	"leadflowcrm/internal/services"
	"leadflowcrm/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthHandler struct {
	authService *services.AuthService
	config      *config.Config
}

func NewAuthHandler(authService *services.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		config:      cfg,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		if err == services.ErrUserExists {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.Role, h.config.JWTSecret, h.config.JWTExpiryHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		if err == services.ErrUserNotFound || err == services.ErrInvalidPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.Role, h.config.JWTSecret, h.config.JWTExpiryHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(primitive.ObjectID)
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) GetUsers(c *gin.Context) {
	users, err := h.authService.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

type LeadHandler struct {
	leadService *services.LeadService
}

func NewLeadHandler(leadService *services.LeadService) *LeadHandler {
	return &LeadHandler{
		leadService: leadService,
	}
}

func (h *LeadHandler) CreateLead(c *gin.Context) {
	var lead models.Lead
	if err := c.ShouldBindJSON(&lead); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if lead.Status == "" {
		lead.Status = models.StatusNew
	}

	userID := c.MustGet("user_id").(primitive.ObjectID)
	created, err := h.leadService.CreateLead(c.Request.Context(), &lead, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *LeadHandler) GetLead(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	lead, err := h.leadService.GetLead(c.Request.Context(), id)
	if err != nil {
		if err == services.ErrLeadNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Lead not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lead)
}

func (h *LeadHandler) GetLeads(c *gin.Context) {
	filter := models.LeadFilter{
		Status:     c.Query("status"),
		AssignedTo: c.Query("assigned_to"),
		Source:     c.Query("search"),
		Search:     c.Query("search"),
		Page:       1,
		Limit:      10,
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil {
		filter.Page = page
	}
	if limit, err := strconv.Atoi(c.Query("limit")); err == nil {
		filter.Limit = limit
	}

	leads, total, err := h.leadService.GetLeads(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      leads,
		"total":     total,
		"page":      filter.Page,
		"limit":     filter.Limit,
		"totalPage": (total + int64(filter.Limit) - 1) / int64(filter.Limit),
	})
}

func (h *LeadHandler) UpdateLead(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var lead models.Lead
	if err := c.ShouldBindJSON(&lead); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	lead.ID = id

	userID := c.MustGet("user_id").(primitive.ObjectID)
	if err := h.leadService.UpdateLead(c.Request.Context(), &lead, userID); err != nil {
		if err == services.ErrLeadNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Lead not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, lead)
}

func (h *LeadHandler) DeleteLead(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	userID := c.MustGet("user_id").(primitive.ObjectID)
	if err := h.leadService.DeleteLead(c.Request.Context(), id, userID); err != nil {
		if err == services.ErrLeadNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Lead not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lead deleted successfully"})
}

func (h *LeadHandler) GetLeadHistory(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	history, err := h.leadService.GetLeadHistory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *LeadHandler) AddNote(c *gin.Context) {
	leadID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lead ID"})
		return
	}

	var note models.LeadNote
	if err := c.ShouldBindJSON(&note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	note.LeadID = leadID

	userID := c.MustGet("user_id").(primitive.ObjectID)
	created, err := h.leadService.AddNote(c.Request.Context(), &note, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *LeadHandler) GetLeadNotes(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	notes, err := h.leadService.GetLeadNotes(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notes)
}

func (h *LeadHandler) DeleteNote(c *gin.Context) {
	noteID, err := primitive.ObjectIDFromHex(c.Param("noteId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	if err := h.leadService.DeleteNote(c.Request.Context(), noteID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}

func (h *LeadHandler) GetStats(c *gin.Context) {
	stats, err := h.leadService.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *LeadHandler) ImportCSV(c *gin.Context) {
	var leads []models.Lead
	if err := c.ShouldBindJSON(&leads); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(primitive.ObjectID)
	count, err := h.leadService.ImportCSV(c.Request.Context(), leads, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"imported": count})
}
