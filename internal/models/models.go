package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"-"`
	Name      string             `bson:"name" json:"name"`
	Role      string             `bson:"role" json:"role"`
	Active    bool               `bson:"active" json:"active"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type LeadStatus string

const (
	StatusNew        LeadStatus = "new"
	StatusContacted  LeadStatus = "contacted"
	StatusQualified  LeadStatus = "qualified"
	StatusProposal   LeadStatus = "proposal"
	StatusNegotiation LeadStatus = "negotiation"
	StatusWon        LeadStatus = "won"
	StatusLost       LeadStatus = "lost"
)

type Lead struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Email        string             `bson:"email" json:"email"`
	Phone        string             `bson:"phone" json:"phone"`
	Company      string             `bson:"company" json:"company"`
	Position     string             `bson:"position" json:"position"`
	Status       LeadStatus         `bson:"status" json:"status"`
	AssignedTo   primitive.ObjectID `bson:"assigned_to,omitempty" json:"assigned_to,omitempty"`
	Source       string             `bson:"source" json:"source"`
	Value        float64            `bson:"value" json:"value"`
	Notes        string             `bson:"notes" json:"notes"`
	Deleted      bool               `bson:"deleted" json:"deleted"`
	DeletedAt    *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

type LeadHistory struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	LeadID      primitive.ObjectID `bson:"lead_id" json:"lead_id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	Action      string             `bson:"action" json:"action"`
	Field       string             `bson:"field" json:"field"`
	OldValue    string             `bson:"old_value" json:"old_value"`
	NewValue    string             `bson:"new_value" json:"new_value"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

type LeadNote struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	LeadID    primitive.ObjectID `bson:"lead_id" json:"lead_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type LeadFilter struct {
	Status     string `json:"status"`
	AssignedTo string `json:"assigned_to"`
	Source     string `json:"source"`
	Search     string `json:"search"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

type LeadStats struct {
	Total      int              `json:"total"`
	ByStatus   map[string]int   `json:"by_status"`
	BySource   map[string]int   `json:"by_source"`
	TotalValue float64          `json:"total_value"`
	Assigned   int              `json:"assigned"`
	Unassigned int              `json:"unassigned"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
